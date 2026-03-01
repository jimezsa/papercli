package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	httpClient  *http.Client
	maxRetries  int
	baseBackoff time.Duration
	logger      *log.Logger
}

type Options struct {
	Timeout     time.Duration
	MaxRetries  int
	BaseBackoff time.Duration
	Logger      *log.Logger
}

func New(opts Options) *Client {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 12 * time.Second
	}
	maxRetries := opts.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	backoff := opts.BaseBackoff
	if backoff <= 0 {
		backoff = 300 * time.Millisecond
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries:  maxRetries,
		baseBackoff: backoff,
		logger:      opts.Logger,
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request, timeout time.Duration, userAgent string) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	if timeout <= 0 {
		timeout = 12 * time.Second
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, timeout)
		attemptReq := req.Clone(attemptCtx)
		if userAgent != "" && attemptReq.Header.Get("User-Agent") == "" {
			attemptReq.Header.Set("User-Agent", userAgent)
		}

		resp, err := c.httpClient.Do(attemptReq)
		if err != nil {
			cancel()
			lastErr = err
			if !retryableError(err) || attempt == c.maxRetries {
				return nil, fmt.Errorf("http request failed: %w", err)
			}
			if sleepErr := sleepWithContext(ctx, c.jitteredBackoff(attempt)); sleepErr != nil {
				return nil, sleepErr
			}
			continue
		}

		if !shouldRetryStatus(resp.StatusCode) || attempt == c.maxRetries {
			resp.Body = &cancelReadCloser{ReadCloser: resp.Body, cancel: cancel}
			return resp, nil
		}

		cancel()
		_ = drainAndClose(resp.Body)
		delay := retryAfterDelay(resp.Header.Get("Retry-After"))
		if delay <= 0 {
			delay = c.jitteredBackoff(attempt)
		}
		if c.logger != nil {
			c.logger.Printf("retrying request status=%d attempt=%d delay=%s", resp.StatusCode, attempt+1, delay)
		}
		if err := sleepWithContext(ctx, delay); err != nil {
			return nil, err
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("http request failed after retries: %w", lastErr)
	}
	return nil, errors.New("http request failed after retries")
}

func (c *Client) jitteredBackoff(attempt int) time.Duration {
	mult := 1 << attempt
	backoff := time.Duration(mult) * c.baseBackoff
	jitter := time.Duration(rand.Intn(200)) * time.Millisecond
	return backoff + jitter
}

func retryableError(err error) bool {
	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return true
}

func shouldRetryStatus(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusBadGateway || code == http.StatusServiceUnavailable || code == http.StatusGatewayTimeout || code >= 500
}

func retryAfterDelay(v string) time.Duration {
	if v == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(v); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func drainAndClose(rc io.ReadCloser) error {
	if rc == nil {
		return nil
	}
	_, _ = io.Copy(io.Discard, rc)
	return rc.Close()
}

type cancelReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
	once   sync.Once
}

func (c *cancelReadCloser) Close() error {
	err := c.ReadCloser.Close()
	c.once.Do(c.cancel)
	return err
}
