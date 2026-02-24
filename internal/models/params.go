package models

const (
	DefaultProvider = "all"
	DefaultLimit    = 20
)

type SearchParams struct {
	Query    string
	Sort     string
	YearFrom int
	YearTo   int
	Limit    int
	Offset   int
}

type AuthorParams struct {
	Name     string
	Sort     string
	YearFrom int
	YearTo   int
	Limit    int
	Offset   int
}

func (p SearchParams) EffectiveLimit() int {
	if p.Limit <= 0 {
		return DefaultLimit
	}
	return p.Limit
}

func (p AuthorParams) EffectiveLimit() int {
	if p.Limit <= 0 {
		return DefaultLimit
	}
	return p.Limit
}
