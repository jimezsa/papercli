package cmd

import (
	"fmt"

	"github.com/jimezsa/papercli/internal/seen"
)

type SeenCmd struct {
	Diff   SeenDiffCmd   `cmd:"" help:"Diff new papers against a seen set." name:"diff"`
	Update SeenUpdateCmd `cmd:"" help:"Update seen set with input papers." name:"update"`
}

type SeenDiffCmd struct {
	New   string `name:"new" help:"Input papers JSON path." required:""`
	Seen  string `name:"seen" help:"Seen JSON path." required:""`
	Out   string `name:"out" aliases:"output,file" help:"Output JSON path." required:""`
	Stats bool   `name:"stats" help:"Print diff stats to stderr."`
}

func (c *SeenDiffCmd) Run(app *App) error {
	papers, err := seen.LoadPapers(c.New)
	if err != nil {
		return err
	}
	store, err := seen.Load(c.Seen)
	if err != nil {
		return err
	}
	out := seen.Diff(papers, seen.ToSet(store))
	if err := seen.SavePapers(c.Out, out); err != nil {
		return err
	}

	if c.Stats {
		_, _ = fmt.Fprintf(app.Stderr, "total=%d unseen=%d seen=%d\n", len(papers), len(out), len(papers)-len(out))
	}
	return nil
}

type SeenUpdateCmd struct {
	Seen  string `name:"seen" help:"Current seen JSON path." required:""`
	Input string `name:"input" help:"Input papers JSON path." required:""`
	Out   string `name:"out" aliases:"output,file" help:"Updated seen JSON path." required:""`
	Stats bool   `name:"stats" help:"Print update stats to stderr."`
}

func (c *SeenUpdateCmd) Run(app *App) error {
	store, err := seen.Load(c.Seen)
	if err != nil {
		return err
	}
	papers, err := seen.LoadPapers(c.Input)
	if err != nil {
		return err
	}

	before := len(store.IDs)
	updated := seen.AddPapers(store, papers)
	if err := seen.Save(c.Out, updated); err != nil {
		return err
	}

	if c.Stats {
		_, _ = fmt.Fprintf(app.Stderr, "before=%d after=%d added=%d\n", before, len(updated.IDs), len(updated.IDs)-before)
	}
	return nil
}
