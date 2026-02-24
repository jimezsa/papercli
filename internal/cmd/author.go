package cmd

import "papercli/internal/models"

type AuthorCmd struct {
	Name string `arg:"" required:"" help:"Author name."`
	QueryFlags
}

func (c *AuthorCmd) Run(app *App) error {
	params := models.AuthorParams{
		Name:     c.Name,
		Sort:     c.Sort,
		YearFrom: c.YearFrom,
		YearTo:   c.YearTo,
		Limit:    app.EffectiveLimit(c.Limit),
		Offset:   c.Offset,
	}

	providerName := app.EffectiveProvider(c.Provider)
	result, err := app.Manager.Author(app.Context(), providerName, params)
	if err != nil && len(result.Papers) == 0 {
		return err
	}
	app.LogProviderErrors(providerName, result.Errors)
	return app.RenderPapers(result.Papers, c.QueryFlags)
}
