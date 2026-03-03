package cmd

import (
	"errors"
	"fmt"
	"strings"
)

func validateSort(sort string) error {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "", "relevance", "date", "citations":
		return nil
	default:
		return fmt.Errorf("unsupported sort %q (expected relevance|date|citations)", sort)
	}
}

func warnUnsupportedSort(app *App, providerName, sort string) {
	sort = strings.ToLower(strings.TrimSpace(sort))
	if sort == "" || sort == "relevance" {
		return
	}

	unsupported := unsupportedSortProviders(providerName, sort)
	if len(unsupported) == 0 {
		return
	}
	fmt.Fprintf(app.Stderr, "warning: --sort %q not supported for provider(s): %s; using provider default\n", sort, strings.Join(unsupported, ", "))
}

func unsupportedSortProviders(providerName, sort string) []string {
	sort = strings.ToLower(strings.TrimSpace(sort))
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if sort == "" || sort == "relevance" {
		return nil
	}

	switch providerName {
	case "arxiv":
		if sort == "citations" {
			return []string{"arxiv"}
		}
		return nil
	case "semantic":
		return []string{"semantic"}
	case "scholar":
		return []string{"scholar"}
	case "all":
		switch sort {
		case "date":
			return []string{"semantic", "scholar"}
		case "citations":
			return []string{"arxiv", "semantic", "scholar"}
		default:
			return nil
		}
	default:
		return nil
	}
}

func validateOutputMode(globals Globals) error {
	if globals.JSON && globals.Plain {
		return errors.New("cannot use --json and --plain together")
	}
	return nil
}
