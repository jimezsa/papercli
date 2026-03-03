package cmd

import "github.com/jimezsa/papercli/internal/models"

type SearchCmd struct {
	Query string `arg:"" required:"" help:"Search query."`
	QueryFlags
}

func (c *SearchCmd) Run(app *App) error {
	params := models.SearchParams{
		Query:    c.Query,
		Sort:     c.Sort,
		YearFrom: c.YearFrom,
		YearTo:   c.YearTo,
		Limit:    app.EffectiveLimit(c.Limit),
		Offset:   c.Offset,
	}

	providerName := app.EffectiveProvider(c.Provider)
	warnUnsupportedSort(app, providerName, c.Sort)
	result, err := app.Manager.Search(app.Context(), providerName, params)
	if err != nil && len(result.Papers) == 0 {
		return err
	}
	app.LogProviderErrors(providerName, result.Errors)
	return app.RenderPapers(result.Papers, c.QueryFlags)
}
