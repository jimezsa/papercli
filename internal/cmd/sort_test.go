package cmd

import (
	"reflect"
	"testing"
)

func TestValidateSort(t *testing.T) {
	for _, value := range []string{"", "relevance", "date", "citations"} {
		if err := validateSort(value); err != nil {
			t.Fatalf("expected sort %q to be valid: %v", value, err)
		}
	}

	if err := validateSort("invalid"); err == nil {
		t.Fatalf("expected invalid sort to fail")
	}
}

func TestUnsupportedSortProviders(t *testing.T) {
	if got := unsupportedSortProviders("arxiv", "date"); len(got) != 0 {
		t.Fatalf("expected arxiv date to be supported, got %v", got)
	}

	wantDateAll := []string{"semantic", "scholar"}
	if got := unsupportedSortProviders("all", "date"); !reflect.DeepEqual(got, wantDateAll) {
		t.Fatalf("unexpected unsupported providers for date/all: got=%v want=%v", got, wantDateAll)
	}

	wantCitationsAll := []string{"arxiv", "semantic", "scholar"}
	if got := unsupportedSortProviders("all", "citations"); !reflect.DeepEqual(got, wantCitationsAll) {
		t.Fatalf("unexpected unsupported providers for citations/all: got=%v want=%v", got, wantCitationsAll)
	}
}
