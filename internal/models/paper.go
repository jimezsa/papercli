package models

type Provider string

const (
	ProviderArXiv    Provider = "arxiv"
	ProviderSemantic Provider = "semantic"
	ProviderScholar  Provider = "scholar"
)

type Paper struct {
	ID       string   `json:"id"`
	Provider Provider `json:"provider"`
	Title    string   `json:"title"`
	Authors  []string `json:"authors,omitempty"`
	Abstract string   `json:"abstract,omitempty"`
	Year     int      `json:"year,omitempty"`
	URL      string   `json:"url,omitempty"`
	PDFURL   string   `json:"pdf_url,omitempty"`
}
