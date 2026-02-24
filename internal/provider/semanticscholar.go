package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jimezsa/papercli/internal/models"
	"github.com/jimezsa/papercli/internal/network"
)

const (
	semanticBaseURL = "https://api.semanticscholar.org/graph/v1"
	semanticUA      = "papercli/0.1 (+https://github.com/jesusjimenez/papercli)"
)

type SemanticScholar struct {
	client  *network.Client
	baseURL string
	apiKey  string
}

func NewSemanticScholar(client *network.Client, apiKey string) *SemanticScholar {
	return &SemanticScholar{
		client:  client,
		baseURL: semanticBaseURL,
		apiKey:  strings.TrimSpace(apiKey),
	}
}

func (s *SemanticScholar) Name() string { return "semantic" }

func (s *SemanticScholar) Search(ctx context.Context, params models.SearchParams) ([]models.Paper, error) {
	v := url.Values{}
	v.Set("query", params.Query)
	v.Set("limit", strconv.Itoa(max(1, params.EffectiveLimit())))
	v.Set("offset", strconv.Itoa(max(0, params.Offset)))
	v.Set("fields", semanticPaperFields())

	endpoint := s.baseURL + "/paper/search?" + v.Encode()
	var payload semanticPaperSearchResponse
	if err := s.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	papers := make([]models.Paper, 0, len(payload.Data))
	for _, p := range payload.Data {
		model := p.toModel()
		if !withinYearRange(model.Year, params.YearFrom, params.YearTo) {
			continue
		}
		papers = append(papers, model)
	}
	return papers, nil
}

func (s *SemanticScholar) Author(ctx context.Context, params models.AuthorParams) ([]models.Paper, error) {
	authorID, err := s.lookupAuthorID(ctx, params.Name)
	if err != nil {
		return nil, err
	}
	if authorID == "" {
		return nil, nil
	}

	v := url.Values{}
	v.Set("limit", strconv.Itoa(max(1, params.EffectiveLimit())))
	v.Set("offset", strconv.Itoa(max(0, params.Offset)))
	v.Set("fields", semanticPaperFields())
	endpoint := s.baseURL + "/author/" + url.PathEscape(authorID) + "/papers?" + v.Encode()

	var payload semanticPaperSearchResponse
	if err := s.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	papers := make([]models.Paper, 0, len(payload.Data))
	for _, p := range payload.Data {
		model := p.toModel()
		if !withinYearRange(model.Year, params.YearFrom, params.YearTo) {
			continue
		}
		papers = append(papers, model)
	}
	return papers, nil
}

func (s *SemanticScholar) Info(ctx context.Context, id string) (*models.Paper, error) {
	endpoint := s.baseURL + "/paper/" + url.PathEscape(id) + "?fields=" + url.QueryEscape(semanticPaperFields())
	var paper semanticPaper
	if err := s.getJSON(ctx, endpoint, &paper); err != nil {
		return nil, err
	}
	model := paper.toModel()
	return &model, nil
}

func (s *SemanticScholar) lookupAuthorID(ctx context.Context, query string) (string, error) {
	v := url.Values{}
	v.Set("query", query)
	v.Set("limit", "1")
	v.Set("offset", "0")
	v.Set("fields", "authorId,name,url")
	endpoint := s.baseURL + "/author/search?" + v.Encode()

	var payload struct {
		Data []struct {
			AuthorID string `json:"authorId"`
		} `json:"data"`
	}
	if err := s.getJSON(ctx, endpoint, &payload); err != nil {
		return "", err
	}
	if len(payload.Data) == 0 {
		return "", nil
	}
	return payload.Data[0].AuthorID, nil
}

func (s *SemanticScholar) getJSON(ctx context.Context, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	if s.apiKey != "" {
		req.Header.Set("x-api-key", s.apiKey)
	}

	resp, err := s.client.Do(ctx, req, 15*time.Second, semanticUA)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("semantic scholar status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode semantic scholar response: %w", err)
	}
	return nil
}

func semanticPaperFields() string {
	return "paperId,title,authors,year,url,abstract,openAccessPdf"
}

type semanticPaperSearchResponse struct {
	Data []semanticPaper `json:"data"`
}

type semanticPaper struct {
	PaperID       string           `json:"paperId"`
	Title         string           `json:"title"`
	Authors       []semanticAuthor `json:"authors"`
	Year          int              `json:"year"`
	URL           string           `json:"url"`
	Abstract      string           `json:"abstract"`
	OpenAccessPDF *semanticPDF     `json:"openAccessPdf"`
}

type semanticAuthor struct {
	Name string `json:"name"`
}

type semanticPDF struct {
	URL string `json:"url"`
}

func (p semanticPaper) toModel() models.Paper {
	authors := make([]string, 0, len(p.Authors))
	for _, a := range p.Authors {
		name := strings.TrimSpace(a.Name)
		if name != "" {
			authors = append(authors, name)
		}
	}
	pdfURL := ""
	if p.OpenAccessPDF != nil {
		pdfURL = p.OpenAccessPDF.URL
	}
	link := strings.TrimSpace(p.URL)
	if link == "" && p.PaperID != "" {
		link = "https://www.semanticscholar.org/paper/" + path.Clean(p.PaperID)
	}
	return models.Paper{
		ID:       p.PaperID,
		Provider: models.ProviderSemantic,
		Title:    oneLine(p.Title),
		Authors:  authors,
		Abstract: strings.TrimSpace(p.Abstract),
		Year:     p.Year,
		URL:      link,
		PDFURL:   pdfURL,
	}
}
