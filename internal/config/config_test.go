package config

import (
	"os"
	"testing"
)

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
