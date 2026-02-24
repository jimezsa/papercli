package cmd

import (
	"fmt"

	"papercli/internal/config"
)

type ConfigCmd struct {
	Init InitConfigCmd `cmd:"" help:"Initialize default config file." name:"init"`
	Path PathConfigCmd `cmd:"" help:"Print config file path." name:"path"`
}

type InitConfigCmd struct{}

func (c *InitConfigCmd) Run(app *App) error {
	path, err := config.InitFile()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(app.Stdout, path)
	return err
}

type PathConfigCmd struct{}

func (c *PathConfigCmd) Run(app *App) error {
	path, err := config.Path()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(app.Stdout, path)
	return err
}
