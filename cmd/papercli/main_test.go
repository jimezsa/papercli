package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/jimezsa/papercli/internal/config"
)

func TestRunAutoInitializesConfigOnBareInvocation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run(nil, &stdout, &stderr); err != nil {
		t.Fatalf("run(nil) returned error: %v", err)
	}

	path, err := config.Path()
	if err != nil {
		t.Fatalf("config.Path() returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist at %q: %v", path, err)
	}
	if !strings.Contains(stdout.String(), "papercli - search and aggregate academic papers") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunConfigInitSkipsAutoInitAndStillSucceeds(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run([]string{"config", "init"}, &stdout, &stderr); err != nil {
		t.Fatalf("run(config init) returned error: %v", err)
	}

	path, err := config.Path()
	if err != nil {
		t.Fatalf("config.Path() returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist at %q: %v", path, err)
	}
	if got := strings.TrimSpace(stdout.String()); got != path {
		t.Fatalf("expected stdout path %q, got %q", path, got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestShouldAutoInitConfig(t *testing.T) {
	tests := []struct {
		name        string
		showVersion bool
		showHelp    bool
		command     string
		args        []string
		want        bool
	}{
		{name: "bare invocation", want: true},
		{name: "explicit help", showHelp: true, want: false},
		{name: "version flag", showVersion: true, want: false},
		{name: "version command", command: "version", want: false},
		{name: "config init", command: "config", args: []string{"init"}, want: false},
		{name: "config path", command: "config", args: []string{"path"}, want: true},
		{name: "search command", command: "search", args: []string{"transformers"}, want: true},
	}

	for _, tt := range tests {
		if got := shouldAutoInitConfig(tt.showVersion, tt.showHelp, tt.command, tt.args); got != tt.want {
			t.Fatalf("%s: expected %t, got %t", tt.name, tt.want, got)
		}
	}
}
