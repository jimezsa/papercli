package config

import (
	"os"
	"testing"
)

func TestEnsureFileCreatesDefaultConfigOnce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path, err := EnsureFile()
	if err != nil {
		t.Fatalf("EnsureFile() first run failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist at %q: %v", path, err)
	}

	pathAgain, err := EnsureFile()
	if err != nil {
		t.Fatalf("EnsureFile() second run failed: %v", err)
	}
	if pathAgain != path {
		t.Fatalf("expected EnsureFile() to return %q, got %q", path, pathAgain)
	}
}

func TestInitFileRequiresForceToOverwrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path, err := InitFile(false)
	if err != nil {
		t.Fatalf("InitFile(false) first run failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist at %q: %v", path, err)
	}

	if _, err := InitFile(false); err == nil {
		t.Fatalf("expected second InitFile(false) to fail without --force")
	}

	overwritePath, err := InitFile(true)
	if err != nil {
		t.Fatalf("InitFile(true) failed: %v", err)
	}
	if overwritePath != path {
		t.Fatalf("expected overwrite path %q, got %q", path, overwritePath)
	}
}
