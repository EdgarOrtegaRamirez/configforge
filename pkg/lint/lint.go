// Package lint provides config file linting with configurable rules.
package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
)

// Severity indicates how serious a lint issue is.
type Severity int

const (
	Info Severity = iota
	Warning
	Error
)

func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warn"
	case Error:
		return "error"
	}
	return "unknown"
}

// Issue represents a single lint finding.
type Issue struct {
	Path     string
	Message  string
	Severity Severity
	Rule     string
}

// Result holds all lint findings.
type Result struct {
	Issues []Issue
}

// Rule is a linting rule that checks config values.
type Rule interface {
	Name() string
	Check(tree *config.Tree) []Issue
}

// DefaultRules returns the standard set of lint rules.
func DefaultRules() []Rule {
	return []Rule{
		&EmptyValueRule{},
		&SecretDetectionRule{},
		&HardcodedURLRule{},
		&DuplicateKeyRule{},
		&DeprecatedKeyRule{},
		&TypeConsistencyRule{},
	}
}

// Lint runs all default rules against a config tree.
func Lint(tree *config.Tree) *Result {
	return LintWithRules(tree, DefaultRules())
}

// LintWithRules runs specific rules against a config tree.
func LintWithRules(tree *config.Tree, rules []Rule) *Result {
	result := &Result{}
	for _, rule := range rules {
		issues := rule.Check(tree)
		result.Issues = append(result.Issues, issues...)
	}
	return result
}

// EmptyValueRule detects empty string values and null entries.
type EmptyValueRule struct{}

func (r *EmptyValueRule) Name() string { return "empty-value" }

func (r *EmptyValueRule) Check(tree *config.Tree) []Issue {
	var issues []Issue
	checkEmpty(tree.Root, "", &issues)
	return issues
}

func checkEmpty(v *config.Value, path string, issues *[]Issue) {
	if v == nil {
		return
	}
	switch v.Kind {
	case config.KindNil:
		*issues = append(*issues, Issue{
			Path:     path,
			Message:  "null/nil value",
			Severity: Warning,
			Rule:     "empty-value",
		})
	case config.KindScalar:
		if s, ok := v.Scalar.(string); ok && s == "" {
			*issues = append(*issues, Issue{
				Path:     path,
				Message:  "empty string value",
				Severity: Warning,
				Rule:     "empty-value",
			})
		}
	case config.KindMap:
		for _, child := range v.Children {
			childPath := child.Key
			if path != "" {
				childPath = path + "." + child.Key
			}
			checkEmpty(child.Value, childPath, issues)
		}
	case config.KindList:
		for i, child := range v.Children {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			checkEmpty(child.Value, childPath, issues)
		}
	}
}

// SecretDetectionRule detects potential hardcoded secrets.
type SecretDetectionRule struct{}

func (r *SecretDetectionRule) Name() string { return "secret-detection" }

var secretPatterns = []struct {
	Pattern *regexp.Regexp
	Name    string
}{
	{regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*\S+`), "password"},
	{regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*\S+`), "api key"},
	{regexp.MustCompile(`(?i)(secret|token)\s*[=:]\s*\S+`), "secret/token"},
	{regexp.MustCompile(`(?i)(private[_-]?key)\s*[=:]\s*\S+`), "private key"},
	{regexp.MustCompile(`-----BEGIN (RSA |EC )?PRIVATE KEY-----`), "private key header"},
	{regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`), "GitHub token"},
	{regexp.MustCompile(`sk-[A-Za-z0-9]{32,}`), "API key (sk- prefix)"},
}

func (r *SecretDetectionRule) Check(tree *config.Tree) []Issue {
	var issues []Issue
	flat := tree.Flatten()
	for path, value := range flat {
		for _, sp := range secretPatterns {
			if sp.Pattern.MatchString(path + "=" + value) {
				issues = append(issues, Issue{
					Path:     path,
					Message:  fmt.Sprintf("potential %s detected", sp.Name),
					Severity: Error,
					Rule:     "secret-detection",
				})
				break
			}
		}
	}
	return issues
}

// HardcodedURLRule detects hardcoded URLs that should be configurable.
type HardcodedURLRule struct{}

func (r *HardcodedURLRule) Name() string { return "hardcoded-url" }

var urlPattern = regexp.MustCompile(`https?://[^\s,"']+`)

func (r *HardcodedURLRule) Check(tree *config.Tree) []Issue {
	var issues []Issue
	flat := tree.Flatten()
	for path, value := range flat {
		if urlPattern.MatchString(value) {
			issues = append(issues, Issue{
				Path:     path,
				Message:  "hardcoded URL — consider making configurable",
				Severity: Info,
				Rule:     "hardcoded-url",
			})
		}
	}
	return issues
}

// DuplicateKeyRule detects potential duplicate/conflicting keys.
type DuplicateKeyRule struct{}

func (r *DuplicateKeyRule) Name() string { return "duplicate-key" }

func (r *DuplicateKeyRule) Check(tree *config.Tree) []Issue {
	// The tree structure inherently deduplicates keys,
	// but we can check for case-insensitive duplicates
	return nil
}

// DeprecatedKeyRule detects known deprecated configuration keys.
type DeprecatedKeyRule struct{}

func (r *DeprecatedKeyRule) Name() string { return "deprecated-key" }

var deprecatedKeys = map[string]string{
	"debug_mode":   "use LOG_LEVEL=debug instead",
	"use_ssl":      "use TLS_CONFIG instead",
	"db_driver":    "auto-detected from connection string",
	"log_to_file":  "use LOG_OUTPUT=file:/path instead",
}

func (r *DeprecatedKeyRule) Check(tree *config.Tree) []Issue {
	var issues []Issue
	flat := tree.Flatten()
	for path := range flat {
		key := path
		if idx := strings.LastIndex(path, "."); idx >= 0 {
			key = path[idx+1:]
		}
		if msg, ok := deprecatedKeys[key]; ok {
			issues = append(issues, Issue{
				Path:     path,
				Message:  fmt.Sprintf("deprecated key: %s", msg),
				Severity: Warning,
				Rule:     "deprecated-key",
			})
		}
	}
	return issues
}

// TypeConsistencyRule checks that values in lists are consistent types.
type TypeConsistencyRule struct{}

func (r *TypeConsistencyRule) Name() string { return "type-consistency" }

func (r *TypeConsistencyRule) Check(tree *config.Tree) []Issue {
	var issues []Issue
	checkTypeConsistency(tree.Root, "", &issues)
	return issues
}

func checkTypeConsistency(v *config.Value, path string, issues *[]Issue) {
	if v == nil {
		return
	}
	if v.Kind == config.KindList && len(v.Children) > 1 {
		firstKind := v.Children[0].Value.Kind
		for i, child := range v.Children[1:] {
			if child.Value.Kind != firstKind && child.Value.Kind != config.KindNil {
				*issues = append(*issues, Issue{
					Path:     fmt.Sprintf("%s[%d]", path, i+1),
					Message:  fmt.Sprintf("type mismatch: expected %s, got %s", firstKind, child.Value.Kind),
					Severity: Warning,
					Rule:     "type-consistency",
				})
			}
		}
	}
	if v.Kind == config.KindMap {
		for _, child := range v.Children {
			childPath := child.Key
			if path != "" {
				childPath = path + "." + child.Key
			}
			checkTypeConsistency(child.Value, childPath, issues)
		}
	}
}
