package configforge_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/check"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
)

func TestCheckClean(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"server": {"host": "localhost", "port": 8080}}`))
	result := check.Check(tree)
	if result.Score != 100 {
		t.Errorf("expected score 100 for clean config, got %d", result.Score)
	}
}

func TestCheckSecret(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"password": "mysecret123", "api_key": "sk-abcdefghijklmnop"}`))
	result := check.Check(tree)
	if result.Score >= 100 {
		t.Error("expected score < 100 for config with secrets")
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Type == check.SecretFound {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected secret issue")
	}
}

func TestCheckWeakPassword(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"password": "password"}`))
	result := check.Check(tree)
	found := false
	for _, issue := range result.Issues {
		if issue.Type == check.WeakPassword {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected weak password issue")
	}
}

func TestCheckEmptySensitive(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"api_key": ""}`))
	result := check.Check(tree)
	found := false
	for _, issue := range result.Issues {
		if issue.Type == check.EmptySensitive {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected empty-sensitive issue")
	}
}

func TestCheckScoreDeduction(t *testing.T) {
	tree, _ := parser.ParseJSONBytes([]byte(`{"password": "password123", "debug": true, "api_key": ""}`))
	result := check.Check(tree)
	if result.Score >= 100 {
		t.Errorf("expected score < 100, got %d", result.Score)
	}
}
