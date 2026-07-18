// Package merge provides deep merging of config trees with conflict resolution.
package merge

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
)

// Strategy defines how to resolve merge conflicts.
type Strategy int

const (
	LastWins  Strategy = iota // Later value overwrites earlier
	FirstWins                 // Earlier value takes precedence
	Union                     // Combine both values (for lists)
	Error                     // Return error on conflict
)

// Conflict represents a merge conflict.
type Conflict struct {
	Path string
	A    *config.Value
	B    *config.Value
}

// Result holds the merge output.
type Result struct {
	Tree      *config.Tree
	Conflicts []Conflict
}

// Merge combines two config trees with the given conflict strategy.
func Merge(a, b *config.Tree, strategy Strategy) (*Result, error) {
	result := &Result{
		Tree: config.NewTree(),
	}
	result.Tree.Format = a.Format
	if b.Format != "" {
		result.Tree.Format = b.Format
	}

	merged, conflicts := mergeValues(a.Root, b.Root, "", strategy)
	result.Tree.Root = merged
	result.Conflicts = conflicts

	if strategy == Error && len(conflicts) > 0 {
		var msgs []string
		for _, c := range conflicts {
			msgs = append(msgs, fmt.Sprintf("conflict at %s: %s vs %s", c.Path, c.A, c.B))
		}
		return result, fmt.Errorf("merge conflicts:\n%s", strings.Join(msgs, "\n"))
	}

	return result, nil
}

func mergeValues(a, b *config.Value, path string, strategy Strategy) (*config.Value, []Conflict) {
	var conflicts []Conflict

	if a == nil && b == nil {
		return nil, nil
	}
	if a == nil {
		return deepCopy(b), nil
	}
	if b == nil {
		return deepCopy(a), nil
	}

	// Different kinds
	if a.Kind != b.Kind {
		conflicts = append(conflicts, Conflict{Path: path, A: a, B: b})
		switch strategy {
		case LastWins:
			return deepCopy(b), conflicts
		case FirstWins:
			return deepCopy(a), conflicts
		default:
			return deepCopy(a), conflicts
		}
	}

	switch a.Kind {
	case config.KindScalar:
		if !a.Equals(b) {
			conflicts = append(conflicts, Conflict{Path: path, A: a, B: b})
			switch strategy {
			case LastWins:
				return deepCopy(b), conflicts
			case FirstWins:
				return deepCopy(a), conflicts
			default:
				return deepCopy(a), conflicts
			}
		}
		return deepCopy(a), nil

	case config.KindNil:
		return &config.Value{Kind: config.KindNil}, nil

	case config.KindMap:
		return mergeMaps(a, b, path, strategy, &conflicts)

	case config.KindList:
		return mergeLists(a, b, path, strategy, &conflicts)
	}

	return deepCopy(a), nil
}

func mergeMaps(a, b *config.Value, path string, strategy Strategy, conflicts *[]Conflict) (*config.Value, []Conflict) {
	result := config.NewMapValue()
	aKeys := a.Keys()
	bKeys := b.Keys()
	allKeys := mergeKeySets(aKeys, bKeys)

	for _, key := range allKeys {
		childPath := key
		if path != "" {
			childPath = path + "." + key
		}

		aEntry := findEntry(a, key)
		bEntry := findEntry(b, key)

		var merged *config.Value
		var childConflicts []Conflict

		if aEntry == nil {
			merged = deepCopy(bEntry.Value)
		} else if bEntry == nil {
			merged = deepCopy(aEntry.Value)
		} else {
			merged, childConflicts = mergeValues(aEntry.Value, bEntry.Value, childPath, strategy)
		}

		*conflicts = append(*conflicts, childConflicts...)
		result.SetKey(key, merged)
	}

	return result, *conflicts
}

func mergeLists(a, b *config.Value, path string, strategy Strategy, conflicts *[]Conflict) (*config.Value, []Conflict) {
	result := config.NewListValue()

	switch strategy {
	case Union:
		// Append all from both
		for _, item := range a.Children {
			result.Children = append(result.Children, &config.Entry{
				Key:   "",
				Value: deepCopy(item.Value),
			})
		}
		for _, item := range b.Children {
			result.Children = append(result.Children, &config.Entry{
				Key:   "",
				Value: deepCopy(item.Value),
			})
		}
	case LastWins:
		// Use b's list entirely
		for _, item := range b.Children {
			result.Children = append(result.Children, &config.Entry{
				Key:   "",
				Value: deepCopy(item.Value),
			})
		}
	default:
		// FirstWins: use a's list entirely
		for _, item := range a.Children {
			result.Children = append(result.Children, &config.Entry{
				Key:   "",
				Value: deepCopy(item.Value),
			})
		}
	}

	return result, nil
}

func deepCopy(v *config.Value) *config.Value {
	if v == nil {
		return nil
	}
	cp := &config.Value{
		Kind:   v.Kind,
		Scalar: v.Scalar,
	}
	if v.Kind == config.KindMap || v.Kind == config.KindList {
		cp.Children = make([]*config.Entry, len(v.Children))
		for i, child := range v.Children {
			cp.Children[i] = &config.Entry{
				Key:   child.Key,
				Value: deepCopy(child.Value),
			}
		}
	}
	return cp
}

func findEntry(v *config.Value, key string) *config.Entry {
	for _, e := range v.Children {
		if e.Key == key {
			return e
		}
	}
	return nil
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
