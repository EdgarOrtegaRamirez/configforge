package configforge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
)

func TestParseJSON(t *testing.T) {
	data := []byte(`{"server": {"host": "localhost", "port": 8080}, "debug": true}`)
	tree, err := parser.ParseJSONBytes(data)
	if err != nil {
		t.Fatalf("ParseJSONBytes failed: %v", err)
	}
	if tree.Format != "json" {
		t.Errorf("expected format 'json', got %q", tree.Format)
	}

	val, ok := tree.Get("server.host")
	if !ok || val.Scalar != "localhost" {
		t.Errorf("expected server.host=localhost, got %v", val)
	}

	val, ok = tree.Get("server.port")
	if !ok || val.Scalar != 8080.0 {
		t.Errorf("expected server.port=8080, got %v", val)
	}
}

func TestParseYAML(t *testing.T) {
	data := []byte(`server:
  host: localhost
  port: 8080
debug: true`)
	tree, err := parser.ParseYAMLBytes(data)
	if err != nil {
		t.Fatalf("ParseYAMLBytes failed: %v", err)
	}
	if tree.Format != "yaml" {
		t.Errorf("expected format 'yaml', got %q", tree.Format)
	}

	val, ok := tree.Get("server.host")
	if !ok || val.Scalar != "localhost" {
		t.Errorf("expected server.host=localhost, got %v", val)
	}
}

func TestParseTOML(t *testing.T) {
	data := []byte(`[server]
host = "localhost"
port = 8080
debug = true`)
	tree, err := parser.ParseTOMLBytes(data)
	if err != nil {
		t.Fatalf("ParseTOMLBytes failed: %v", err)
	}
	if tree.Format != "toml" {
		t.Errorf("expected format 'toml', got %q", tree.Format)
	}

	val, ok := tree.Get("server.host")
	if !ok || val.Scalar != "localhost" {
		t.Errorf("expected server.host=localhost, got %v", val)
	}
}

func TestParseINI(t *testing.T) {
	data := []byte(`[server]
host = localhost
port = 8080

[database]
driver = mysql`)
	tree, err := parser.ParseINIBytes(data)
	if err != nil {
		t.Fatalf("ParseINIBytes failed: %v", err)
	}
	if tree.Format != "ini" {
		t.Errorf("expected format 'ini', got %q", tree.Format)
	}

	val, ok := tree.Get("server.host")
	if !ok || val.Scalar != "localhost" {
		t.Errorf("expected server.host=localhost, got %v", val)
	}

	val, ok = tree.Get("database.driver")
	if !ok || val.Scalar != "mysql" {
		t.Errorf("expected database.driver=mysql, got %v", val)
	}
}

func TestParseDotenv(t *testing.T) {
	data := []byte(`HOST=localhost
PORT=8080
DATABASE_URL=postgres://localhost/mydb`)
	tree, err := parser.ParseDotenvBytes(data)
	if err != nil {
		t.Fatalf("ParseDotenvBytes failed: %v", err)
	}
	if tree.Format != "dotenv" {
		t.Errorf("expected format 'dotenv', got %q", tree.Format)
	}

	val, ok := tree.Get("HOST")
	if !ok || val.Scalar != "localhost" {
		t.Errorf("expected HOST=localhost, got %v", val)
	}
}

func TestMarshalJSON(t *testing.T) {
	tree, err := parser.ParseJSONBytes([]byte(`{"a": 1, "b": "hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	data, err := parser.MarshalJSON(tree)
	if err != nil {
		t.Fatal(err)
	}
	// Re-parse to verify round-trip
	tree2, err := parser.ParseJSONBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if !tree.Root.Equals(tree2.Root) {
		t.Error("JSON round-trip failed")
	}
}

func TestMarshalYAML(t *testing.T) {
	tree, err := parser.ParseYAMLBytes([]byte(`a: 1
b: hello`))
	if err != nil {
		t.Fatal(err)
	}
	data, err := parser.MarshalYAML(tree)
	if err != nil {
		t.Fatal(err)
	}
	tree2, err := parser.ParseYAMLBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if !tree.Root.Equals(tree2.Root) {
		t.Error("YAML round-trip failed")
	}
}

func TestMarshalTOML(t *testing.T) {
	tree, err := parser.ParseTOMLBytes([]byte(`a = 1
b = "hello"`))
	if err != nil {
		t.Fatal(err)
	}
	data, err := parser.MarshalTOML(tree)
	if err != nil {
		t.Fatal(err)
	}
	tree2, err := parser.ParseTOMLBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if !tree.Root.Equals(tree2.Root) {
		t.Error("TOML round-trip failed")
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		path   string
		format string
	}{
		{"config.json", "json"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"config.toml", "toml"},
		{"config.ini", "ini"},
		{"config.cfg", "ini"},
		{".env", "dotenv"},
		{".env.local", "dotenv"},
	}
	for _, tt := range tests {
		got := parser.DetectFormat(tt.path)
		if got != tt.format {
			t.Errorf("DetectFormat(%q) = %q, want %q", tt.path, got, tt.format)
		}
	}
}

func TestParseFileJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	os.WriteFile(path, []byte(`{"key": "value"}`), 0644)

	tree, err := parser.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	val, ok := tree.Get("key")
	if !ok || val.Scalar != "value" {
		t.Errorf("expected key=value, got %v", val)
	}
}

func TestParseFileYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	os.WriteFile(path, []byte("key: value"), 0644)

	tree, err := parser.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	val, ok := tree.Get("key")
	if !ok || val.Scalar != "value" {
		t.Errorf("expected key=value, got %v", val)
	}
}
