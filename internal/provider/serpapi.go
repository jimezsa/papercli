package provider

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"papercli/internal/models"
	"papercli/internal/network"
)

const (
	serpapiEndpoint = "https://serpapi.com/search.json"
	serpapiUA       = "papercli/0.1 (+https://github.com/jesusjimenez/papercli)"
)

var yearRe = regexp.MustCompile(`\b(19|20)\d{2}\b`)

type SerpAPI struct {
	client   *network.Client
	endpoint string
	apiKey   string
}

func NewSerpAPI(client *network.Client, apiKey string) *SerpAPI {
	return &SerpAPI{
		client:   client,
		endpoint: serpapiEndpoint,
		apiKey:   strings.TrimSpace(apiKey),
	}
}

func (s *SerpAPI) Name() string { return "scholar" }

func (s *SerpAPI) Search(ctx context.Context, params models.SearchParams) ([]models.Paper, error) {
	if s.apiKey == "" {
		return nil, ErrNotConfigured
	}
	v := url.Values{}
	v.Set("engine", "google_scholar")
	v.Set("q", params.Query)
	v.Set("api_key", s.apiKey)
	v.Set("num", strconv.Itoa(max(1, params.EffectiveLimit())))
	if params.Offset > 0 {
		v.Set("start", strconv.Itoa(params.Offset))
	}

	endpoint := s.endpoint + "?" + v.Encode()
	var payload serpScholarResponse
	if err := s.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, err
	}

	out := make([]models.Paper, 0, len(payload.OrganicResults))
	for _, result := range payload.OrganicResults {
		paper := result.toModel()
		if !withinYearRange(paper.Year, params.YearFrom, params.YearTo) {
			continue
		}
		out = append(out, paper)
	}
	return out, nil
}

func (s *SerpAPI) Author(ctx context.Context, params models.AuthorParams) ([]models.Paper, error) {
	if s.apiKey == "" {
		return nil, ErrNotConfigured
	}
	v := url.Values{}
	v.Set("engine", "google_scholar_profiles")
	v.Set("mauthors", params.Name)
	v.Set("api_key", s.apiKey)

	endpoint := s.endpoint + "?" + v.Encode()
	var payload serpProfilesResponse
	if err := s.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, err
	}

	limit := params.EffectiveLimit()
	if len(payload.Profiles) < limit {
		limit = len(payload.Profiles)
	}

	out := make([]models.Paper, 0, limit)
	for i := 0; i < limit; i++ {
		profile := payload.Profiles[i]
		out = append(out, models.Paper{
			ID:       strings.TrimSpace(profile.AuthorID),
			Provider: models.ProviderScholar,
			Title:    strings.TrimSpace(profile.Name),
			Authors:  []string{strings.TrimSpace(profile.Name)},
			URL:      strings.TrimSpace(profile.Link),
		})
	}
	return out, nil
}

func (s *SerpAPI) Info(ctx context.Context, id string) (*models.Paper, error) {
	if s.apiKey == "" {
		return nil, ErrNotConfigured
	}
	if strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://") {
		return &models.Paper{
			ID:       hashID(id),
			Provider: models.ProviderScholar,
			Title:    id,
			URL:      id,
		}, nil
	}
	return nil, fmt.Errorf("scholar info lookup by id is not supported, pass URL")
}

func (s *SerpAPI) getJSON(ctx context.Context, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(ctx, req, 15*time.Second, serpapiUA)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("serpapi status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode serpapi response: %w", err)
	}
	return nil
}

type serpScholarResponse struct {
	OrganicResults []serpOrganicResult `json:"organic_results"`
}

type serpOrganicResult struct {
	ResultID        string `json:"result_id"`
	Title           string `json:"title"`
	Link            string `json:"link"`
	PublicationInfo struct {
		Summary string `json:"summary"`
	} `json:"publication_info"`
}

func (r serpOrganicResult) toModel() models.Paper {
	id := strings.TrimSpace(r.ResultID)
	if id == "" {
		id = hashID(r.Link + "::" + r.Title)
	}
	summary := strings.TrimSpace(r.PublicationInfo.Summary)
	authors := parseAuthors(summary)
	year := parseYear(summary)
	return models.Paper{
		ID:       id,
		Provider: models.ProviderScholar,
		Title:    oneLine(r.Title),
		Authors:  authors,
		Year:     year,
		URL:      strings.TrimSpace(r.Link),
	}
}

type serpProfilesResponse struct {
	Profiles []serpProfile `json:"profiles"`
}

type serpProfile struct {
	Name     string `json:"name"`
	Link     string `json:"link"`
	AuthorID string `json:"author_id"`
}

func parseYear(v string) int {
	m := yearRe.FindString(v)
	if m == "" {
		return 0
	}
	year, err := strconv.Atoi(m)
	if err != nil {
		return 0
	}
	return year
}

func parseAuthors(summary string) []string {
	parts := strings.Split(summary, "-")
	if len(parts) == 0 {
		return nil
	}
	head := strings.TrimSpace(parts[0])
	if head == "" {
		return nil
	}
	raw := strings.Split(head, ",")
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		out = append(out, name)
	}
	return out
}

func hashID(v string) string {
	sum := sha1.Sum([]byte(v))
	return hex.EncodeToString(sum[:8])
}
