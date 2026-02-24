package seen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"papercli/internal/models"
)

type Store struct {
	IDs []string `json:"ids"`
}

func Load(path string) (Store, error) {
	var s Store
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Store{}, nil
		}
		return s, fmt.Errorf("read seen file %q: %w", path, err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return Store{}, nil
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("decode seen file %q: %w", path, err)
	}
	return s, nil
}

func Save(path string, store Store) error {
	store.IDs = dedupeAndSort(store.IDs)
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("encode seen file %q: %w", path, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write seen file %q: %w", path, err)
	}
	return nil
}

func Normalize(provider models.Provider, id, rawURL string) string {
	base := strings.TrimSpace(id)
	if base == "" {
		base = strings.TrimSpace(rawURL)
	}
	base = strings.ToLower(base)
	return fmt.Sprintf("%s:%s", strings.ToLower(string(provider)), base)
}

func ToSet(store Store) map[string]struct{} {
	out := make(map[string]struct{}, len(store.IDs))
	for _, id := range store.IDs {
		out[strings.ToLower(strings.TrimSpace(id))] = struct{}{}
	}
	return out
}

func Diff(papers []models.Paper, set map[string]struct{}) []models.Paper {
	var unseen []models.Paper
	for _, paper := range papers {
		norm := Normalize(paper.Provider, paper.ID, paper.URL)
		if _, ok := set[norm]; !ok {
			unseen = append(unseen, paper)
		}
	}
	return unseen
}

func AddPapers(store Store, papers []models.Paper) Store {
	set := ToSet(store)
	for _, paper := range papers {
		set[Normalize(paper.Provider, paper.ID, paper.URL)] = struct{}{}
	}

	store.IDs = store.IDs[:0]
	for id := range set {
		store.IDs = append(store.IDs, id)
	}
	store.IDs = dedupeAndSort(store.IDs)
	return store
}

func LoadPapers(path string) ([]models.Paper, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read papers file %q: %w", path, err)
	}
	var papers []models.Paper
	if err := json.Unmarshal(data, &papers); err != nil {
		return nil, fmt.Errorf("decode papers file %q: %w", path, err)
	}
	return papers, nil
}

func SavePapers(path string, papers []models.Paper) error {
	data, err := json.MarshalIndent(papers, "", "  ")
	if err != nil {
		return fmt.Errorf("encode papers file %q: %w", path, err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func dedupeAndSort(ids []string) []string {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.ToLower(strings.TrimSpace(id))
		if id == "" {
			continue
		}
		set[id] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
