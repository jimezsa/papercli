package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/jimezsa/papercli/internal/config"
	"github.com/jimezsa/papercli/internal/models"
	"github.com/jimezsa/papercli/internal/network"
	"github.com/jimezsa/papercli/internal/provider"
)

type App struct {
	Version string
	Stdout  io.Writer
	Stderr  io.Writer
	Globals Globals
	Config  config.Config
	Logger  *log.Logger
	HTTP    *network.Client
	Manager *provider.Manager
}

func NewApp(version string, globals Globals, stdout, stderr io.Writer) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := log.New(stderr, "papercli: ", log.LstdFlags)

	httpClient := network.New(network.Options{
		Timeout:     15 * time.Second,
		MaxRetries:  3,
		BaseBackoff: 350 * time.Millisecond,
		Logger:      logger,
		Verbose:     globals.Verbose,
	})

	mgr := provider.NewManager(
		provider.NewArxiv(httpClient),
		provider.NewSemanticScholar(httpClient, cfg.SemanticAPIKey),
		provider.NewSerpAPI(httpClient, cfg.SerpAPIKey),
	)

	return &App{
		Version: version,
		Stdout:  stdout,
		Stderr:  stderr,
		Globals: globals,
		Config:  cfg,
		Logger:  logger,
		HTTP:    httpClient,
		Manager: mgr,
	}, nil
}

func (a *App) Context() context.Context {
	return context.Background()
}

func (a *App) EffectiveProvider(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v != "" {
		return v
	}
	if a.Config.DefaultProvider != "" {
		return strings.ToLower(a.Config.DefaultProvider)
	}
	return models.DefaultProvider
}

func (a *App) EffectiveLimit(limit int) int {
	if limit > 0 {
		return limit
	}
	if a.Config.DefaultLimit > 0 {
		return a.Config.DefaultLimit
	}
	return models.DefaultLimit
}

func (a *App) LogProviderErrors(providerName string, errs map[string]error) {
	if len(errs) == 0 {
		return
	}
	for name, err := range errs {
		if errors.Is(err, provider.ErrNotConfigured) && providerName == "all" {
			if a.Globals.Verbose {
				a.Logger.Printf("provider %s not configured, skipping", name)
			}
			continue
		}
		fmt.Fprintf(a.Stderr, "provider %s: %v\n", name, err)
	}
}
