package configforge_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/query"
)

func TestQuerySimple(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"server": {"host": "localhost", "port": 8080}}`))
	results, err := query.Query(tree, "server.host")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value.Scalar != "localhost" {
		t.Errorf("expected 'localhost', got %v", results[0].Value)
	}
}

func TestQueryWildcard(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"a": 1, "b": 2, "c": 3}`))
	results, err := query.Query(tree, "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestQueryNonexistent(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"a": 1}`))
	results, err := query.Query(tree, "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent key, got %d", len(results))
	}
}

func TestQueryArray(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"items": ["a", "b", "c"]}`))
	results, err := query.Query(tree, "items[1]")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value.Scalar != "b" {
		t.Errorf("expected 'b', got %v", results[0].Value)
	}
}

func TestQueryDeepNested(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"a": {"b": {"c": {"d": "deep"}}}}`))
	results, err := query.Query(tree, "a.b.c.d")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value.Scalar != "deep" {
		t.Errorf("expected 'deep', got %v", results[0].Value)
	}
}

func TestQueryWildcardDeep(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"servers": {"web": {"port": 80}, "db": {"port": 5432}}}`))
	results, err := query.Query(tree, "servers.*.port")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}
