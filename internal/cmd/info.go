package cmd

import "papercli/internal/models"

type InfoCmd struct {
	ID       string `arg:"" required:"" help:"Provider paper identifier."`
	Provider string `help:"Provider to query." enum:"arxiv,semantic,scholar,all" default:"all"`
	Format   string `help:"Output format." enum:"csv,json,md" default:"json"`
	Links    string `help:"Link rendering mode." enum:"short,full" default:"full"`
	Out      string `name:"out" aliases:"output" help:"Output file path."`
}

func (c *InfoCmd) Run(app *App) error {
	paper, err := app.Manager.Info(app.Context(), app.EffectiveProvider(c.Provider), c.ID)
	if err != nil {
		return err
	}
	flags := QueryFlags{
		Format: c.Format,
		Links:  c.Links,
		Out:    c.Out,
	}
	return app.RenderPapers([]models.Paper{*paper}, flags)
}
