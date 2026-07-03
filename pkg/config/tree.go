// Package config provides the universal config tree representation.
// All format-specific parsers convert to/from this tree structure.
package config

import (
	"fmt"
	"sort"
	"strings"
)

// Value represents a config value - can be scalar, map, or list.
type Value struct {
	Kind     ValueKind
	Scalar   interface{} // string, int, float64, bool, nil
	Children []*Entry     // for maps and lists
	Raw      string       // original raw representation
	Line     int          // source line number (0 if unknown)
}

// ValueKind enumerates the types of config values.
type ValueKind int

const (
	KindScalar ValueKind = iota
	KindMap
	KindList
	KindNil
)

func (k ValueKind) String() string {
	switch k {
	case KindScalar:
		return "scalar"
	case KindMap:
		return "map"
	case KindList:
		return "list"
	case KindNil:
		return "nil"
	default:
		return "unknown"
	}
}

// Entry is a key-value pair in a config tree.
type Entry struct {
	Key   string
	Value *Value
}

// Tree is a universal config representation.
type Tree struct {
	Root    *Value
	Format  string // yaml, toml, json, ini, dotenv
	Source  string // file path
	Entries []*Entry
}

// NewTree creates a new empty tree.
func NewTree() *Tree {
	return &Tree{
		Root: &Value{
			Kind: KindMap,
		},
	}
}

// NewScalarValue creates a scalar value.
func NewScalarValue(v interface{}) *Value {
	return &Value{
		Kind:   KindScalar,
		Scalar: v,
	}
}

// NewMapValue creates a map value.
func NewMapValue() *Value {
	return &Value{
		Kind:     KindMap,
		Children: make([]*Entry, 0),
	}
}

// NewListValue creates a list value.
func NewListValue() *Value {
	return &Value{
		Kind:     KindList,
		Children: make([]*Entry, 0),
	}
}

// Get retrieves a value by dot-separated path.
// Example: "server.port" returns the port value.
func (t *Tree) Get(path string) (*Value, bool) {
	if path == "" {
		return t.Root, true
	}
	return t.Root.Get(path)
}

// Get retrieves a value by dot-separated path from a value.
func (v *Value) Get(path string) (*Value, bool) {
	if v == nil {
		return nil, false
	}
	parts := strings.Split(path, ".")
	current := v
	for _, part := range parts {
		if current == nil || current.Kind != KindMap {
			return nil, false
		}
		found := false
		for _, child := range current.Children {
			if child.Key == part {
				current = child.Value
				found = true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	return current, true
}

// Set sets a value at the given dot-separated path.
// Creates intermediate maps as needed.
func (t *Tree) Set(path string, val *Value) {
	t.Root.Set(path, val)
}

// Set sets a value at the given dot-separated path.
func (v *Value) Set(path string, val *Value) {
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		v.SetKey(parts[0], val)
		return
	}

	// Find or create intermediate map
	key := parts[0]
	remaining := strings.Join(parts[1:], ".")
	child := v.findOrCreateMap(key)
	child.Set(remaining, val)
}

// SetKey sets a key at the current map level.
func (v *Value) SetKey(key string, val *Value) {
	if v.Kind != KindMap {
		return
	}
	for _, child := range v.Children {
		if child.Key == key {
			child.Value = val
			return
		}
	}
	v.Children = append(v.Children, &Entry{Key: key, Value: val})
}

func (v *Value) findOrCreateMap(key string) *Value {
	if v.Kind != KindMap {
		v.Kind = KindMap
		v.Children = make([]*Entry, 0)
	}
	for _, child := range v.Children {
		if child.Key == key {
			if child.Value.Kind != KindMap {
				child.Value = NewMapValue()
			}
			return child.Value
		}
	}
	newMap := NewMapValue()
	v.Children = append(v.Children, &Entry{Key: key, Value: newMap})
	return newMap
}

// Keys returns sorted keys of a map value.
func (v *Value) Keys() []string {
	if v == nil || v.Kind != KindMap {
		return nil
	}
	keys := make([]string, 0, len(v.Children))
	for _, child := range v.Children {
		keys = append(keys, child.Key)
	}
	sort.Strings(keys)
	return keys
}

// Len returns the number of children for maps and lists.
func (v *Value) Len() int {
	if v == nil {
		return 0
	}
	return len(v.Children)
}

// String returns a human-readable representation.
func (v *Value) String() string {
	if v == nil {
		return "<nil>"
	}
	switch v.Kind {
	case KindScalar:
		return fmt.Sprintf("%v", v.Scalar)
	case KindMap:
		return fmt.Sprintf("{map with %d entries}", len(v.Children))
	case KindList:
		return fmt.Sprintf("[list with %d items]", len(v.Children))
	case KindNil:
		return "<null>"
	default:
		return "<unknown>"
	}
}

// Equals deep-compares two values.
func (v *Value) Equals(other *Value) bool {
	if v == nil && other == nil {
		return true
	}
	if v == nil || other == nil {
		return false
	}
	if v.Kind != other.Kind {
		return false
	}
	switch v.Kind {
	case KindScalar:
		return fmt.Sprintf("%v", v.Scalar) == fmt.Sprintf("%v", other.Scalar)
	case KindNil:
		return true
	case KindMap:
		if len(v.Children) != len(other.Children) {
			return false
		}
		vKeys := v.Keys()
		oKeys := other.Keys()
		for i, k := range vKeys {
			if k != oKeys[i] {
				return false
			}
		}
		for _, e := range v.Children {
			oe := other.findEntry(e.Key)
			if oe == nil || !e.Value.Equals(oe.Value) {
				return false
			}
		}
		return true
	case KindList:
		if len(v.Children) != len(other.Children) {
			return false
		}
		for i, e := range v.Children {
			if !e.Value.Equals(other.Children[i].Value) {
				return false
			}
		}
		return true
	}
	return false
}

func (v *Value) findEntry(key string) *Entry {
	for _, e := range v.Children {
		if e.Key == key {
			return e
		}
	}
	return nil
}

// Flatten returns all key-value pairs with dot-separated paths.
func (t *Tree) Flatten() map[string]string {
	result := make(map[string]string)
	t.flatten("", t.Root, result)
	return result
}

func (t *Tree) flatten(prefix string, v *Value, result map[string]string) {
	if v == nil {
		return
	}
	switch v.Kind {
	case KindScalar:
		result[prefix] = fmt.Sprintf("%v", v.Scalar)
	case KindNil:
		result[prefix] = ""
	case KindMap:
		for _, child := range v.Children {
			key := child.Key
			if prefix != "" {
				key = prefix + "." + child.Key
			}
			t.flatten(key, child.Value, result)
		}
	case KindList:
		for i, child := range v.Children {
			key := fmt.Sprintf("%s[%d]", prefix, i)
			t.flatten(key, child.Value, result)
		}
	}
}

// KeysAtPath returns the direct children keys at a given path.
func (t *Tree) KeysAtPath(path string) []string {
	val, ok := t.Get(path)
	if !ok {
		return nil
	}
	return val.Keys()
}
