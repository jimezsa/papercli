package cmd

import "testing"

func TestParseGlobalArgs_GlobalFlagAfterCommand(t *testing.T) {
	globals, showVersion, showHelp, command, args, err := ParseGlobalArgs([]string{
		"search", "transformers", "--json",
	})
	if err != nil {
		t.Fatalf("ParseGlobalArgs returned error: %v", err)
	}
	if showVersion {
		t.Fatalf("expected showVersion=false")
	}
	if showHelp {
		t.Fatalf("expected showHelp=false")
	}
	if command != "search" {
		t.Fatalf("expected command search, got %q", command)
	}
	if len(args) != 1 || args[0] != "transformers" {
		t.Fatalf("unexpected args: %#v", args)
	}
	if !globals.JSON {
		t.Fatalf("expected JSON global flag to be true")
	}
}

func TestParseGlobalArgs_HelpAfterCommand(t *testing.T) {
	_, _, showHelp, command, args, err := ParseGlobalArgs([]string{
		"seen", "diff", "--help",
	})
	if err != nil {
		t.Fatalf("ParseGlobalArgs returned error: %v", err)
	}
	if !showHelp {
		t.Fatalf("expected showHelp=true")
	}
	if command != "seen" {
		t.Fatalf("expected command seen, got %q", command)
	}
	if len(args) != 1 || args[0] != "diff" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestParseGlobalArgs_OutputModeConflict(t *testing.T) {
	_, _, _, _, _, err := ParseGlobalArgs([]string{
		"search", "transformers", "--json", "--plain",
	})
	if err == nil {
		t.Fatalf("expected conflict error for --json and --plain")
	}
}

func TestParseGlobalArgs_InvalidColor(t *testing.T) {
	_, _, _, _, _, err := ParseGlobalArgs([]string{
		"--color", "not-a-color", "version",
	})
	if err == nil {
		t.Fatalf("expected invalid color error")
	}
}

func TestValidateSeenFlags(t *testing.T) {
	if err := validateSeenFlags("", true, ""); err == nil {
		t.Fatalf("expected error when --new-only is set without --seen")
	}
	if err := validateSeenFlags("", false, "new.json"); err == nil {
		t.Fatalf("expected error when --new-out is set without --seen")
	}
	if err := validateSeenFlags("seen.json", true, "new.json"); err != nil {
		t.Fatalf("did not expect error with valid seen flags: %v", err)
	}
}
