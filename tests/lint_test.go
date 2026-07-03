package configforge_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/lint"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
)

func TestLintClean(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"server": {"host": "localhost", "port": 8080}}`))
	result := lint.Lint(tree)
	if len(result.Issues) > 0 {
		t.Errorf("expected no issues for clean config, got %d", len(result.Issues))
	}
}

func TestLintEmptyValue(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"name": ""}`))
	result := lint.Lint(tree)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "empty-value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected empty-value issue")
	}
}

func TestLintSecretDetection(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"password": "mysecret123"}`))
	result := lint.Lint(tree)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "secret-detection" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected secret-detection issue")
	}
}

func TestLintHardcodedURL(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"api_url": "https://api.example.com/v1"}`))
	result := lint.Lint(tree)
	found := false
	for _, issue := range result.Issues {
		if issue.Rule == "hardcoded-url" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected hardcoded-url issue")
	}
}
