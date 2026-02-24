package provider

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"papercli/internal/models"
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
	if len(selected) == 1 {
		p, err := selected[0].Info(ctx, id)
		if err != nil {
			return nil, ProviderError{Provider: selected[0].Name(), Op: "info", Err: err}
		}
		return p, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type infoResult struct {
		paper *models.Paper
		err   error
		name  string
	}
	results := make(chan infoResult, len(selected))

	var wg sync.WaitGroup
	for _, p := range selected {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			paper, err := p.Info(ctx, id)
			if err != nil {
				results <- infoResult{name: p.Name(), err: ProviderError{Provider: p.Name(), Op: "info", Err: err}}
				return
			}
			if paper != nil {
				cancel()
				results <- infoResult{name: p.Name(), paper: paper}
				return
			}
			results <- infoResult{name: p.Name(), err: ProviderError{Provider: p.Name(), Op: "info", Err: errors.New("paper not found")}}
		}()
	}
	wg.Wait()
	close(results)

	var errs []error
	for res := range results {
		if res.paper != nil {
			return res.paper, nil
		}
		if res.err != nil {
			errs = append(errs, res.err)
		}
	}
	if len(errs) == 0 {
		return nil, errors.New("paper not found")
	}
	return nil, errors.Join(errs...)
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
		selected := make([]Provider, 0, len(m.providers))
		for _, p := range m.providers {
			selected = append(selected, p)
		}
		sort.Slice(selected, func(i, j int) bool {
			return selected[i].Name() < selected[j].Name()
		})
		return selected, nil
	}

	p, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", providerName)
	}
	return []Provider{p}, nil
}
