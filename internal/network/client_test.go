package network

import "testing"

func TestNewCarriesVerboseOption(t *testing.T) {
	disabled := New(Options{Verbose: false})
	if disabled.verbose {
		t.Fatalf("expected verbose=false")
	}

	enabled := New(Options{Verbose: true})
	if !enabled.verbose {
		t.Fatalf("expected verbose=true")
	}
}

func TestShouldRetryStatus(t *testing.T) {
	retryable := []int{429, 500, 502, 503, 504}
	for _, code := range retryable {
		if !shouldRetryStatus(code) {
			t.Fatalf("expected status %d to be retryable", code)
		}
	}

	nonRetryable := []int{200, 201, 204, 400, 401, 403, 404}
	for _, code := range nonRetryable {
		if shouldRetryStatus(code) {
			t.Fatalf("expected status %d to be non-retryable", code)
		}
	}
}
