package cmd

import "fmt"

type VersionCmd struct{}

func (v *VersionCmd) Run(app *App) error {
	_, err := fmt.Fprintln(app.Stdout, app.Version)
	return err
}
