package configforge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/convert"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
)

func TestConvertJSONToYAML(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.json")
	output := filepath.Join(dir, "output.yaml")

	os.WriteFile(input, []byte(`{"server": {"host": "localhost", "port": 8080}}`), 0644)

	err := convert.Convert(input, output)
	if err != nil {
		t.Fatal(err)
	}

	tree, err := parser.ParseFile(output)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := tree.Get("server.host")
	if !ok || val.Scalar != "localhost" {
		t.Errorf("expected server.host=localhost, got %v", val)
	}
}

func TestConvertJSONToTOML(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.json")
	output := filepath.Join(dir, "output.toml")

	os.WriteFile(input, []byte(`{"key": "value", "number": 42}`), 0644)

	err := convert.Convert(input, output)
	if err != nil {
		t.Fatal(err)
	}

	tree, err := parser.ParseFile(output)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := tree.Get("key")
	if !ok || val.Scalar != "value" {
		t.Errorf("expected key=value, got %v", val)
	}
}

func TestConvertBytes(t *testing.T) {
	input := []byte(`{"a": 1, "b": "hello"}`)
	output, err := convert.ConvertBytes(input, "json", "yaml")
	if err != nil {
		t.Fatal(err)
	}

	tree, err := parser.ParseYAMLBytes(output)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := tree.Get("a")
	if !ok || val.Scalar != 1.0 {
		t.Errorf("expected a=1, got %v", val)
	}
}
