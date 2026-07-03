// Package diff provides semantic diffing for config trees.
package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
)

// ChangeKind represents the type of change.
type ChangeKind int

const (
	Added ChangeKind = iota
	Removed
	Modified
	Unchanged
)

func (c ChangeKind) String() string {
	switch c {
	case Added:
		return "added"
	case Removed:
		return "removed"
	case Modified:
		return "modified"
	case Unchanged:
		return "unchanged"
	}
	return "unknown"
}

// Change represents a single difference between two config trees.
type Change struct {
	Path     string
	Kind     ChangeKind
	OldValue *config.Value
	NewValue *config.Value
}

// DiffResult holds all differences between two trees.
type DiffResult struct {
	Changes []Change
	Stats   DiffStats
}

// DiffStats summarizes the diff.
type DiffStats struct {
	Added    int
	Removed  int
	Modified int
	Unchanged int
}

// Diff computes semantic differences between two config trees.
func Diff(a, b *config.Tree) *DiffResult {
	result := &DiffResult{}
	diffValues(a.Root, b.Root, "", result)
	sort.Slice(result.Changes, func(i, j int) bool {
		return result.Changes[i].Path < result.Changes[j].Path
	})
	return result
}

func diffValues(a, b *config.Value, path string, result *DiffResult) {
	if a == nil && b == nil {
		return
	}
	if a == nil {
		result.Changes = append(result.Changes, Change{
			Path:     path,
			Kind:     Added,
			NewValue: b,
		})
		result.Stats.Added++
		return
	}
	if b == nil {
		result.Changes = append(result.Changes, Change{
			Path:     path,
			Kind:     Removed,
			OldValue: a,
		})
		result.Stats.Removed++
		return
	}

	// Different kinds = modified
	if a.Kind != b.Kind {
		result.Changes = append(result.Changes, Change{
			Path:     path,
			Kind:     Modified,
			OldValue: a,
			NewValue: b,
		})
		result.Stats.Modified++
		return
	}

	switch a.Kind {
	case config.KindScalar:
		if !a.Equals(b) {
			result.Changes = append(result.Changes, Change{
				Path:     path,
				Kind:     Modified,
				OldValue: a,
				NewValue: b,
			})
			result.Stats.Modified++
		} else {
			result.Stats.Unchanged++
		}
	case config.KindNil:
		result.Stats.Unchanged++
	case config.KindMap:
		diffMaps(a, b, path, result)
	case config.KindList:
		diffLists(a, b, path, result)
	}
}

func diffMaps(a, b *config.Value, path string, result *DiffResult) {
	aKeys := a.Keys()
	bKeys := b.Keys()
	allKeys := mergeKeySets(aKeys, bKeys)

	for _, key := range allKeys {
		childPath := key
		if path != "" {
			childPath = path + "." + key
		}
		aChild := findEntry(a, key)
		bChild := findEntry(b, key)

		if aChild == nil {
			result.Changes = append(result.Changes, Change{
				Path:     childPath,
				Kind:     Added,
				NewValue: bChild.Value,
			})
			result.Stats.Added++
		} else if bChild == nil {
			result.Changes = append(result.Changes, Change{
				Path:     childPath,
				Kind:     Removed,
				OldValue: aChild.Value,
			})
			result.Stats.Removed++
		} else {
			diffValues(aChild.Value, bChild.Value, childPath, result)
		}
	}
}

func diffLists(a, b *config.Value, path string, result *DiffResult) {
	maxLen := len(a.Children)
	if len(b.Children) > maxLen {
		maxLen = len(b.Children)
	}

	for i := 0; i < maxLen; i++ {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		if i >= len(a.Children) {
			result.Changes = append(result.Changes, Change{
				Path:     itemPath,
				Kind:     Added,
				NewValue: b.Children[i].Value,
			})
			result.Stats.Added++
		} else if i >= len(b.Children) {
			result.Changes = append(result.Changes, Change{
				Path:     itemPath,
				Kind:     Removed,
				OldValue: a.Children[i].Value,
			})
			result.Stats.Removed++
		} else {
			diffValues(a.Children[i].Value, b.Children[i].Value, itemPath, result)
		}
	}
}

func mergeKeySets(a, b []string) []string {
	keyMap := make(map[string]bool)
	for _, k := range a {
		keyMap[k] = true
	}
	for _, k := range b {
		keyMap[k] = true
	}
	keys := make([]string, 0, len(keyMap))
	for k := range keyMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// FormatText outputs the diff as human-readable text.
func FormatText(result *DiffResult) string {
	var sb strings.Builder

	for _, change := range result.Changes {
		switch change.Kind {
		case Added:
			sb.WriteString(fmt.Sprintf("+ %s = %s\n", change.Path, change.NewValue))
		case Removed:
			sb.WriteString(fmt.Sprintf("- %s = %s\n", change.Path, change.OldValue))
		case Modified:
			sb.WriteString(fmt.Sprintf("~ %s: %s → %s\n", change.Path, change.OldValue, change.NewValue))
		}
	}

	sb.WriteString(fmt.Sprintf("\n--- Summary ---\n"))
	sb.WriteString(fmt.Sprintf("Added:    %d\n", result.Stats.Added))
	sb.WriteString(fmt.Sprintf("Removed:  %d\n", result.Stats.Removed))
	sb.WriteString(fmt.Sprintf("Modified: %d\n", result.Stats.Modified))

	return sb.String()
}

// FormatJSON outputs the diff as JSON.
func FormatJSON(result *DiffResult) string {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, change := range result.Changes {
		sb.WriteString(fmt.Sprintf(`  {"path": %q, "kind": %q`, change.Path, change.Kind.String()))
		if change.OldValue != nil {
			sb.WriteString(fmt.Sprintf(`, "old": %q`, change.OldValue))
		}
		if change.NewValue != nil {
			sb.WriteString(fmt.Sprintf(`, "new": %q`, change.NewValue))
		}
		sb.WriteString("}")
		if i < len(result.Changes)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("]\n")
	return sb.String()
}

func findEntry(v *config.Value, key string) *config.Entry {
	for _, e := range v.Children {
		if e.Key == key {
			return e
		}
	}
	return nil
}
