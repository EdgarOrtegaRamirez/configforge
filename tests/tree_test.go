package configforge_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
)

func TestNewTree(t *testing.T) {
	tree := config.NewTree()
	if tree == nil {
		t.Fatal("NewTree returned nil")
	}
	if tree.Root == nil {
		t.Fatal("tree.Root is nil")
	}
	if tree.Root.Kind != config.KindMap {
		t.Errorf("expected KindMap, got %v", tree.Root.Kind)
	}
}

func TestScalarValue(t *testing.T) {
	v := config.NewScalarValue("hello")
	if v.Kind != config.KindScalar {
		t.Errorf("expected KindScalar, got %v", v.Kind)
	}
	if v.Scalar != "hello" {
		t.Errorf("expected 'hello', got %v", v.Scalar)
	}
}

func TestMapValue(t *testing.T) {
	m := config.NewMapValue()
	m.SetKey("name", config.NewScalarValue("test"))
	m.SetKey("count", config.NewScalarValue(42.0))

	if m.Len() != 2 {
		t.Errorf("expected 2 children, got %d", m.Len())
	}
	keys := m.Keys()
	if len(keys) != 2 || keys[0] != "count" || keys[1] != "name" {
		t.Errorf("unexpected keys: %v", keys)
	}
}

func TestTreeGetSet(t *testing.T) {
	tree := config.NewTree()
	tree.Set("server.host", config.NewScalarValue("localhost"))
	tree.Set("server.port", config.NewScalarValue(8080.0))

	val, ok := tree.Get("server.host")
	if !ok || val.Scalar != "localhost" {
		t.Errorf("expected 'localhost', got %v", val)
	}

	val, ok = tree.Get("server.port")
	if !ok || val.Scalar != 8080.0 {
		t.Errorf("expected 8080, got %v", val)
	}

	_, ok = tree.Get("server.nonexistent")
	if ok {
		t.Error("expected not found for nonexistent key")
	}
}

func TestTreeFlatten(t *testing.T) {
	tree := config.NewTree()
	tree.Set("a.b.c", config.NewScalarValue("deep"))
	tree.Set("a.x", config.NewScalarValue(1.0))

	flat := tree.Flatten()
	if flat["a.b.c"] != "deep" {
		t.Errorf("expected 'deep', got %v", flat["a.b.c"])
	}
	if flat["a.x"] != "1" {
		t.Errorf("expected '1', got %v", flat["a.x"])
	}
}

func TestValueEquals(t *testing.T) {
	a := config.NewScalarValue("hello")
	b := config.NewScalarValue("hello")
	c := config.NewScalarValue("world")

	if !a.Equals(b) {
		t.Error("expected equal values")
	}
	if a.Equals(c) {
		t.Error("expected different values")
	}
}

func TestListValue(t *testing.T) {
	l := config.NewListValue()
	l.Children = append(l.Children, &config.Entry{Value: config.NewScalarValue("a")})
	l.Children = append(l.Children, &config.Entry{Value: config.NewScalarValue("b")})

	if l.Len() != 2 {
		t.Errorf("expected 2 items, got %d", l.Len())
	}
}

func TestTreeGetNestedMap(t *testing.T) {
	tree := config.NewTree()
	tree.Set("a.b.c.d", config.NewScalarValue("deep"))

	val, ok := tree.Get("a.b.c.d")
	if !ok || val.Scalar != "deep" {
		t.Errorf("expected 'deep' at nested path, got %v", val)
	}

	// Get intermediate map
	val, ok = tree.Get("a.b")
	if !ok || val.Kind != config.KindMap {
		t.Errorf("expected map at a.b, got %v", val)
	}
}
