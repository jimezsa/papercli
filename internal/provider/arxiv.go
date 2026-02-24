package provider

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"papercli/internal/models"
	"papercli/internal/network"
)

const (
	arxivEndpoint = "https://export.arxiv.org/api/query"
	arxivUA       = "papercli/0.1 (+https://github.com/jesusjimenez/papercli)"
)

type Arxiv struct {
	client   *network.Client
	endpoint string
}

func NewArxiv(client *network.Client) *Arxiv {
	return &Arxiv{
		client:   client,
		endpoint: arxivEndpoint,
	}
}

func (a *Arxiv) Name() string { return "arxiv" }

func (a *Arxiv) Search(ctx context.Context, params models.SearchParams) ([]models.Paper, error) {
	v := url.Values{}
	v.Set("search_query", "all:"+params.Query)
	v.Set("start", strconv.Itoa(max(0, params.Offset)))
	v.Set("max_results", strconv.Itoa(max(1, params.EffectiveLimit())))
	v.Set("sortBy", arxivSortBy(params.Sort))
	v.Set("sortOrder", "descending")
	return a.query(ctx, v, params.YearFrom, params.YearTo)
}

func (a *Arxiv) Author(ctx context.Context, params models.AuthorParams) ([]models.Paper, error) {
	v := url.Values{}
	v.Set("search_query", "au:"+params.Name)
	v.Set("start", strconv.Itoa(max(0, params.Offset)))
	v.Set("max_results", strconv.Itoa(max(1, params.EffectiveLimit())))
	v.Set("sortBy", arxivSortBy(params.Sort))
	v.Set("sortOrder", "descending")
	return a.query(ctx, v, params.YearFrom, params.YearTo)
}

func (a *Arxiv) Info(ctx context.Context, id string) (*models.Paper, error) {
	v := url.Values{}
	v.Set("id_list", id)
	v.Set("max_results", "1")
	papers, err := a.query(ctx, v, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(papers) == 0 {
		return nil, fmt.Errorf("paper %q not found", id)
	}
	return &papers[0], nil
}

func (a *Arxiv) query(ctx context.Context, params url.Values, yearFrom, yearTo int) ([]models.Paper, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(ctx, req, 15*time.Second, arxivUA)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("arxiv status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	feed, err := parseArxivFeed(resp.Body)
	if err != nil {
		return nil, err
	}
	out := make([]models.Paper, 0, len(feed.Entries))
	for _, e := range feed.Entries {
		p, err := arxivEntryToPaper(e)
		if err != nil {
			continue
		}
		if !withinYearRange(p.Year, yearFrom, yearTo) {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func arxivSortBy(sort string) string {
	switch strings.ToLower(sort) {
	case "date":
		return "submittedDate"
	default:
		return "relevance"
	}
}

func withinYearRange(year, from, to int) bool {
	if year <= 0 {
		return true
	}
	if from > 0 && year < from {
		return false
	}
	if to > 0 && year > to {
		return false
	}
	return true
}

type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID        string       `xml:"id"`
	Title     string       `xml:"title"`
	Summary   string       `xml:"summary"`
	Published string       `xml:"published"`
	Authors   []atomAuthor `xml:"author"`
	Links     []atomLink   `xml:"link"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomLink struct {
	Href  string `xml:"href,attr"`
	Rel   string `xml:"rel,attr"`
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr"`
}

func parseArxivFeed(r io.Reader) (atomFeed, error) {
	var feed atomFeed
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&feed); err != nil {
		return atomFeed{}, fmt.Errorf("decode arxiv atom feed: %w", err)
	}
	return feed, nil
}

func arxivEntryToPaper(e atomEntry) (models.Paper, error) {
	id := strings.TrimSpace(e.ID)
	if id == "" {
		return models.Paper{}, fmt.Errorf("missing arxiv id")
	}
	parsedID := extractArxivID(id)
	published := parseTime(e.Published)
	url := id
	pdfURL := ""
	for _, link := range e.Links {
		if link.Rel == "alternate" && link.Href != "" {
			url = link.Href
		}
		if strings.Contains(link.Type, "pdf") || strings.EqualFold(link.Title, "pdf") {
			pdfURL = link.Href
		}
	}
	authors := make([]string, 0, len(e.Authors))
	for _, a := range e.Authors {
		name := strings.TrimSpace(a.Name)
		if name != "" {
			authors = append(authors, name)
		}
	}
	title := oneLine(e.Title)
	abstract := strings.TrimSpace(e.Summary)
	year := 0
	if !published.IsZero() {
		year = published.Year()
	}

	return models.Paper{
		ID:       parsedID,
		Provider: models.ProviderArXiv,
		Title:    title,
		Authors:  authors,
		Abstract: abstract,
		Year:     year,
		URL:      url,
		PDFURL:   pdfURL,
	}, nil
}

func extractArxivID(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	last := path.Base(u.Path)
	if last == "." || last == "/" || last == "" {
		return raw
	}
	return last
}

func parseTime(v string) time.Time {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}
	}
	return t
}

func oneLine(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "\n", " ")
	return strings.Join(strings.Fields(v), " ")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
