// Package query provides path-based querying of config trees.
package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
)

// QueryResult holds the result of a query operation.
type QueryResult struct {
	Path  string
	Value *config.Value
}

// Query executes a dot-notation path query on a config tree.
// Supports:
//   - dot notation: server.port
//   - array indexing: items[0]
//   - wildcard: *.port
//   - recursive: **.key
func Query(tree *config.Tree, expr string) ([]QueryResult, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty query expression")
	}

	// Wildcard at root level
	if expr == "*" {
		return queryWildcard(tree.Root, ""), nil
	}

	// Recursive descent
	if strings.HasPrefix(expr, "**.") {
		key := expr[3:]
		return queryRecursive(tree.Root, "", key), nil
	}

	// Simple dot path
	parts := splitPath(expr)
	return queryParts(tree.Root, parts, "")
}

func splitPath(expr string) []string {
	var parts []string
	var current strings.Builder
	bracketDepth := 0

	for _, ch := range expr {
		switch ch {
		case '.':
			if bracketDepth == 0 {
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
				continue
			}
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		}
		current.WriteRune(ch)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

var indexPattern = regexp.MustCompile(`^(.+)\[(\d+)\]$`)
var bareIndexPattern = regexp.MustCompile(`^\[(\d+)\]$`)

func queryParts(v *config.Value, parts []string, basePath string) ([]QueryResult, error) {
	if v == nil || len(parts) == 0 {
		return []QueryResult{{Path: basePath, Value: v}}, nil
	}

	head := parts[0]
	tail := parts[1:]

	// Check for array indexing: key[index] or just [index]
	key, index, hasIndex := parseIndex(head)

	// Wildcard
	if key == "*" || head == "*" {
		if v.Kind != config.KindMap {
			return nil, nil
		}
		var results []QueryResult
		for _, child := range v.Children {
			childPath := child.Key
			if basePath != "" {
				childPath = basePath + "." + child.Key
			}
			r, err := queryParts(child.Value, tail, childPath)
			if err != nil {
				return nil, err
			}
			results = append(results, r...)
		}
		return results, nil
	}

	// Navigate to key
	child := findEntry(v, key)
	if child == nil {
		return nil, nil
	}

	childPath := key
	if basePath != "" {
		childPath = basePath + "." + key
	}

	// Apply array index if present
	val := child.Value
	if hasIndex && val.Kind == config.KindList {
		if index >= 0 && index < len(val.Children) {
			val = val.Children[index].Value
			childPath = fmt.Sprintf("%s[%d]", childPath, index)
		} else {
			return nil, nil
		}
	}

	return queryParts(val, tail, childPath)
}

func parseIndex(s string) (string, int, bool) {
	matches := indexPattern.FindStringSubmatch(s)
	if matches != nil {
		idx, _ := strconv.Atoi(matches[2])
		return matches[1], idx, true
	}
	// Just [index] without key
	matches2 := bareIndexPattern.FindStringSubmatch(s)
	if matches2 != nil {
		idx, _ := strconv.Atoi(matches2[1])
		return "", idx, true
	}
	return s, 0, false
}

func queryWildcard(v *config.Value, basePath string) []QueryResult {
	if v == nil || v.Kind != config.KindMap {
		return nil
	}
	var results []QueryResult
	for _, child := range v.Children {
		childPath := child.Key
		if basePath != "" {
			childPath = basePath + "." + child.Key
		}
		results = append(results, QueryResult{Path: childPath, Value: child.Value})
	}
	return results
}

func queryRecursive(v *config.Value, basePath string, key string) []QueryResult {
	if v == nil {
		return nil
	}
	var results []QueryResult

	if v.Kind == config.KindMap {
		for _, child := range v.Children {
			childPath := child.Key
			if basePath != "" {
				childPath = basePath + "." + child.Key
			}
			if child.Key == key {
				results = append(results, QueryResult{Path: childPath, Value: child.Value})
			}
			results = append(results, queryRecursive(child.Value, childPath, key)...)
		}
	}
	return results
}

func findEntry(v *config.Value, key string) *config.Entry {
	if v == nil || v.Kind != config.KindMap {
		return nil
	}
	for _, e := range v.Children {
		if e.Key == key {
			return e
		}
	}
	return nil
}

// FormatResults formats query results for display.
func FormatResults(results []QueryResult, format string) string {
	var sb strings.Builder
	for _, r := range results {
		switch format {
		case "paths":
			sb.WriteString(r.Path + "\n")
		case "values":
			sb.WriteString(fmt.Sprintf("%s = %s\n", r.Path, r.Value))
		default:
			sb.WriteString(fmt.Sprintf("%-40s %s\n", r.Path, r.Value))
		}
	}
	return sb.String()
}
