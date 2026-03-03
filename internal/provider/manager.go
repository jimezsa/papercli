package provider

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/jimezsa/papercli/internal/models"
)

type Manager struct {
	providers map[string]Provider
}

type Result struct {
	Papers []models.Paper
	Errors map[string]error
}

func NewManager(providers ...Provider) *Manager {
	m := &Manager{
		providers: make(map[string]Provider, len(providers)),
	}
	for _, p := range providers {
		if p == nil {
			continue
		}
		m.providers[strings.ToLower(p.Name())] = p
	}
	return m
}

func (m *Manager) Search(ctx context.Context, providerName string, params models.SearchParams) (Result, error) {
	selected, err := m.selectProviders(providerName)
	if err != nil {
		return Result{}, err
	}
	return m.fanOut(ctx, selected, func(ctx context.Context, p Provider) ([]models.Paper, error) {
		return p.Search(ctx, params)
	})
}

func (m *Manager) Author(ctx context.Context, providerName string, params models.AuthorParams) (Result, error) {
	selected, err := m.selectProviders(providerName)
	if err != nil {
		return Result{}, err
	}
	return m.fanOut(ctx, selected, func(ctx context.Context, p Provider) ([]models.Paper, error) {
		return p.Author(ctx, params)
	})
}

func (m *Manager) Info(ctx context.Context, providerName, id string) (*models.Paper, error) {
	selected, err := m.selectProviders(providerName)
	if err != nil {
		return nil, err
	}
	var errs []error
	for _, p := range selected {
		paper, err := p.Info(ctx, id)
		if err != nil {
			errs = append(errs, ProviderError{Provider: p.Name(), Op: "info", Err: err})
			continue
		}
		if paper != nil {
			return paper, nil
		}
		errs = append(errs, ProviderError{Provider: p.Name(), Op: "info", Err: errors.New("paper not found")})
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return nil, errors.New("paper not found")
}

func (m *Manager) fanOut(
	ctx context.Context,
	selected []Provider,
	fn func(context.Context, Provider) ([]models.Paper, error),
) (Result, error) {
	type providerResult struct {
		papers []models.Paper
		err    error
		name   string
	}

	out := make(chan providerResult, len(selected))
	var wg sync.WaitGroup
	for _, p := range selected {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			papers, err := fn(ctx, p)
			if err != nil {
				out <- providerResult{name: p.Name(), err: err}
				return
			}
			out <- providerResult{name: p.Name(), papers: papers}
		}()
	}

	wg.Wait()
	close(out)

	result := Result{Errors: map[string]error{}}
	var successful bool
	for r := range out {
		if r.err != nil {
			result.Errors[r.name] = r.err
			continue
		}
		successful = true
		result.Papers = append(result.Papers, r.papers...)
	}

	sort.SliceStable(result.Papers, func(i, j int) bool {
		return result.Papers[i].Year > result.Papers[j].Year
	})

	if successful {
		return result, nil
	}

	var errs []error
	for name, err := range result.Errors {
		errs = append(errs, fmt.Errorf("%s: %w", name, err))
	}
	if len(errs) == 0 {
		return result, errors.New("no providers returned results")
	}
	return result, errors.Join(errs...)
}

func (m *Manager) selectProviders(providerName string) ([]Provider, error) {
	name := strings.ToLower(strings.TrimSpace(providerName))
	if name == "" || name == "all" {
		order := []string{"arxiv", "semantic", "scholar"}
		selected := make([]Provider, 0, len(order))
		seen := make(map[string]struct{}, len(m.providers))
		for _, provider := range order {
			p, ok := m.providers[provider]
			if !ok {
				continue
			}
			selected = append(selected, p)
			seen[provider] = struct{}{}
		}

		extraNames := make([]string, 0, len(m.providers))
		for provider := range m.providers {
			if _, ok := seen[provider]; ok {
				continue
			}
			extraNames = append(extraNames, provider)
		}
		sort.Strings(extraNames)
		for _, provider := range extraNames {
			selected = append(selected, m.providers[provider])
		}
		return selected, nil
	}

	p, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", providerName)
	}
	return []Provider{p}, nil
}
