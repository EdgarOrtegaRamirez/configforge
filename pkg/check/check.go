// Package check provides security and quality checks for config files.
package check

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
)

// IssueType categorizes check findings.
type IssueType int

const (
	SecretFound IssueType = iota
	WeakPassword
	InsecureDefault
	UnsafeSetting
	EmptySensitive
)

func (t IssueType) String() string {
	switch t {
	case SecretFound:
		return "secret"
	case WeakPassword:
		return "weak-password"
	case InsecureDefault:
		return "insecure-default"
	case UnsafeSetting:
		return "unsafe-setting"
	case EmptySensitive:
		return "empty-sensitive"
	}
	return "unknown"
}

// Issue represents a security or quality finding.
type Issue struct {
	Path    string
	Type    IssueType
	Message string
	Value   string
}

// Result holds all check findings.
type Result struct {
	Issues []Issue
	Score  int // 0-100, higher is better
}

// Check runs all security and quality checks on a config tree.
func Check(tree *config.Tree) *Result {
	result := &Result{Score: 100}

	checkSecrets(tree, result)
	checkWeakPasswords(tree, result)
	checkInsecureDefaults(tree, result)
	checkUnsafeSettings(tree, result)
	checkEmptySensitive(tree, result)

	// Deduct points for issues
	for _, issue := range result.Issues {
		switch issue.Type {
		case SecretFound:
			result.Score -= 30
		case WeakPassword:
			result.Score -= 20
		case InsecureDefault:
			result.Score -= 10
		case UnsafeSetting:
			result.Score -= 15
		case EmptySensitive:
			result.Score -= 5
		}
	}

	if result.Score < 0 {
		result.Score = 0
	}

	return result
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(secret|token)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(private[_-]?key)\s*[=:]\s*\S+`),
	regexp.MustCompile(`-----BEGIN (RSA |EC )?PRIVATE KEY-----`),
	regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`),
	regexp.MustCompile(`sk-[A-Za-z0-9]{32,}`),
	regexp.MustCompile(`(?i)(aws[_-]?secret[_-]?access[_-]?key)\s*[=:]\s*\S+`),
}

func checkSecrets(tree *config.Tree, result *Result) {
	flat := tree.Flatten()
	for path, value := range flat {
		for _, pat := range secretPatterns {
			if pat.MatchString(path + "=" + value) {
				result.Issues = append(result.Issues, Issue{
					Path:    path,
					Type:    SecretFound,
					Message: "potential hardcoded secret detected",
					Value:   maskValue(value),
				})
				break
			}
		}
	}
}

func maskValue(v string) string {
	if len(v) <= 4 {
		return "****"
	}
	return v[:2] + strings.Repeat("*", len(v)-4) + v[len(v)-2:]
}

func checkWeakPasswords(tree *config.Tree, result *Result) {
	sensitiveKeys := []string{"password", "passwd", "pwd", "secret"}
	flat := tree.Flatten()

	for path, value := range flat {
		key := strings.ToLower(path)
		if idx := strings.LastIndex(key, "."); idx >= 0 {
			key = key[idx+1:]
		}

		for _, sensitive := range sensitiveKeys {
			if strings.Contains(key, sensitive) {
				if isWeakPassword(value) {
					result.Issues = append(result.Issues, Issue{
						Path:    path,
						Type:    WeakPassword,
						Message: fmt.Sprintf("weak password: %s", weakPasswordReason(value)),
						Value:   maskValue(value),
					})
				}
				break
			}
		}
	}
}

func isWeakPassword(pw string) bool {
	if len(pw) < 8 {
		return true
	}
	// Check for common patterns
	common := []string{
		"password", "123456", "qwerty", "admin", "letmein",
		"welcome", "monkey", "dragon", "master", "login",
		"abc123", "password1", "12345678", "123456789",
	}
	lower := strings.ToLower(pw)
	for _, c := range common {
		if lower == c {
			return true
		}
	}
	return false
}

func weakPasswordReason(pw string) string {
	if len(pw) < 8 {
		return "too short"
	}
	lower := strings.ToLower(pw)
	common := []string{"password", "123456", "qwerty", "admin", "letmein"}
	for _, c := range common {
		if lower == c {
			return "common password"
		}
	}
	return "weak"
}

func checkInsecureDefaults(tree *config.Tree, result *Result) {
	insecureDefaults := map[string][]string{
		"ssl_verify":    {"false", "0", "no"},
		"tls_verify":    {"false", "0", "no"},
		"debug":         {"true", "1", "yes"},
		"allow_all_origins": {"true", "1", "yes"},
		"cors_allow_all": {"true", "1", "yes"},
		"trust_proxy":   {"true", "1", "yes"},
	}

	flat := tree.Flatten()
	for path, value := range flat {
		key := strings.ToLower(path)
		if idx := strings.LastIndex(key, "."); idx >= 0 {
			key = key[idx+1:]
		}

		if insecureValues, ok := insecureDefaults[key]; ok {
			lowerValue := strings.ToLower(value)
			for _, insecureVal := range insecureValues {
				if lowerValue == insecureVal {
					result.Issues = append(result.Issues, Issue{
						Path:    path,
						Type:    InsecureDefault,
						Message: fmt.Sprintf("insecure default: %s=%s", key, value),
						Value:   value,
					})
					break
				}
			}
		}
	}
}

func checkUnsafeSettings(tree *config.Tree, result *Result) {
	unsafeSettings := map[string]string{
		"eval":            "use of eval is dangerous",
		"shell":           "shell execution detected",
		"exec":            "command execution detected",
		"allow_http":      "HTTP should be disabled in production",
		"disable_auth":    "authentication is disabled",
		"no_auth":         "authentication is disabled",
		"anonymous":       "anonymous access enabled",
	}

	flat := tree.Flatten()
	for path, value := range flat {
		key := strings.ToLower(path)
		if idx := strings.LastIndex(key, "."); idx >= 0 {
			key = key[idx+1:]
		}

		for unsafeKey, msg := range unsafeSettings {
			if strings.Contains(key, unsafeKey) {
				lowerValue := strings.ToLower(value)
				if lowerValue == "true" || lowerValue == "1" || lowerValue == "yes" {
					result.Issues = append(result.Issues, Issue{
						Path:    path,
						Type:    UnsafeSetting,
						Message: msg,
						Value:   value,
					})
				}
			}
		}
	}
}

func checkEmptySensitive(tree *config.Tree, result *Result) {
	sensitiveKeys := []string{
		"password", "secret", "token", "api_key", "apikey",
		"access_key", "private_key", "auth_token",
	}

	flat := tree.Flatten()
	for path, value := range flat {
		key := strings.ToLower(path)
		if idx := strings.LastIndex(key, "."); idx >= 0 {
			key = key[idx+1:]
		}

		for _, sensitive := range sensitiveKeys {
			if strings.Contains(key, sensitive) && value == "" {
				result.Issues = append(result.Issues, Issue{
					Path:    path,
					Type:    EmptySensitive,
					Message: fmt.Sprintf("empty sensitive field: %s", key),
					Value:   "",
				})
				break
			}
		}
	}
}

// Shannon entropy calculation for detecting high-entropy strings (likely secrets).
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]float64)
	for _, ch := range s {
		freq[ch]++
	}
	length := float64(len(s))
	entropy := 0.0
	for _, count := range freq {
		p := count / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}
