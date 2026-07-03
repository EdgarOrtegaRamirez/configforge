package configforge_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/diff"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
)

func TestDiffIdentical(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1, "b": "hello"}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 1, "b": "hello"}`))

	result := diff.Diff(a, b)
	if len(result.Changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(result.Changes))
	}
}

func TestDiffAdded(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 1, "b": 2}`))

	result := diff.Diff(a, b)
	if result.Stats.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Stats.Added)
	}
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	if result.Changes[0].Path != "b" {
		t.Errorf("expected path 'b', got %q", result.Changes[0].Path)
	}
}

func TestDiffRemoved(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1, "b": 2}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 1}`))

	result := diff.Diff(a, b)
	if result.Stats.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Stats.Removed)
	}
}

func TestDiffModified(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 2}`))

	result := diff.Diff(a, b)
	if result.Stats.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", result.Stats.Modified)
	}
}

func TestDiffNested(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"server": {"host": "localhost", "port": 8080}}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"server": {"host": "localhost", "port": 9090}}`))

	result := diff.Diff(a, b)
	if result.Stats.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", result.Stats.Modified)
	}
	if result.Changes[0].Path != "server.port" {
		t.Errorf("expected path 'server.port', got %q", result.Changes[0].Path)
	}
}

func TestDiffLists(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"items": ["a", "b"]}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"items": ["a", "b", "c"]}`))

	result := diff.Diff(a, b)
	if result.Stats.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Stats.Added)
	}
}

func TestDiffFormatText(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 2}`))

	result := diff.Diff(a, b)
	text := diff.FormatText(result)
	if text == "" {
		t.Error("expected non-empty text output")
	}
}

func TestDiffTypeChange(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": "hello"}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 42}`))

	result := diff.Diff(a, b)
	if result.Stats.Modified != 1 {
		t.Errorf("expected 1 modified for type change, got %d", result.Stats.Modified)
	}
}
