package cmd

type Globals struct {
	Color   string `help:"Color output mode." enum:"auto,always,never" default:"auto" env:"PAPERCLI_COLOR"`
	JSON    bool   `help:"Output JSON to stdout." env:"PAPERCLI_JSON"`
	Plain   bool   `help:"Output TSV to stdout."`
	Verbose bool   `help:"Enable debug logging to stderr." env:"PAPERCLI_VERBOSE"`
}
