package configforge_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/merge"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
)

func TestMergeIdentical(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1, "b": "hello"}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 1, "b": "hello"}`))

	result, err := merge.Merge(a, b, merge.LastWins)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(result.Conflicts))
	}
}

func TestMergeNonOverlapping(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"b": 2}`))

	result, err := merge.Merge(a, b, merge.LastWins)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(result.Conflicts))
	}

	val, ok := result.Tree.Get("a")
	if !ok || val.Scalar != 1.0 {
		t.Errorf("expected a=1, got %v", val)
	}
	val, ok = result.Tree.Get("b")
	if !ok || val.Scalar != 2.0 {
		t.Errorf("expected b=2, got %v", val)
	}
}

func TestMergeConflictLastWins(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 2}`))

	result, err := merge.Merge(a, b, merge.LastWins)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}

	val, ok := result.Tree.Get("a")
	if !ok || val.Scalar != 2.0 {
		t.Errorf("expected a=2 (last wins), got %v", val)
	}
}

func TestMergeConflictFirstWins(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"a": 1}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"a": 2}`))

	result, err := merge.Merge(a, b, merge.FirstWins)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := result.Tree.Get("a")
	if !ok || val.Scalar != 1.0 {
		t.Errorf("expected a=1 (first wins), got %v", val)
	}
}

func TestMergeUnion(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"items": ["a", "b"]}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"items": ["c", "d"]}`))

	result, err := merge.Merge(a, b, merge.Union)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := result.Tree.Get("items")
	if !ok || val.Len() != 4 {
		t.Errorf("expected 4 items after union, got %d", val.Len())
	}
}

func TestMergeNested(t *testing.T) {
	a, _ := parser.ParseJSONBytes([]byte(`{"server": {"host": "localhost"}}`))
	b, _ := parser.ParseJSONBytes([]byte(`{"server": {"port": 8080}}`))

	result, err := merge.Merge(a, b, merge.LastWins)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := result.Tree.Get("server.host")
	if !ok || val.Scalar != "localhost" {
		t.Errorf("expected server.host=localhost, got %v", val)
	}
	val, ok = result.Tree.Get("server.port")
	if !ok || val.Scalar != 8080.0 {
		t.Errorf("expected server.port=8080, got %v", val)
	}
}
