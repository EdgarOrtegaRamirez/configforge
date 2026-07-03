package parser

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
	"github.com/BurntSushi/toml"
)

// ParseTOML parses a TOML file into a config Tree.
func ParseTOML(path string) (*config.Tree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return ParseTOMLBytes(data)
}

// ParseTOMLBytes parses TOML bytes into a config Tree.
func ParseTOMLBytes(data []byte) (*config.Tree, error) {
	var raw map[string]interface{}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing TOML: %w", err)
	}
	tree := config.NewTree()
	tree.Format = "toml"
	tree.Root = convertFromTOML(raw)
	return tree, nil
}

func convertFromTOML(v interface{}) *config.Value {
	if v == nil {
		return &config.Value{Kind: config.KindNil}
	}
	switch val := v.(type) {
	case map[string]interface{}:
		m := config.NewMapValue()
		for k, child := range val {
			m.Children = append(m.Children, &config.Entry{
				Key:   k,
				Value: convertFromTOML(child),
			})
		}
		return m
	case []interface{}:
		l := config.NewListValue()
		for _, child := range val {
			l.Children = append(l.Children, &config.Entry{
				Key:   "",
				Value: convertFromTOML(child),
			})
		}
		return l
	case string:
		return config.NewScalarValue(val)
	case int64:
		return config.NewScalarValue(float64(val))
	case float64:
		return config.NewScalarValue(val)
	case bool:
		return config.NewScalarValue(val)
	case time.Time:
		return config.NewScalarValue(val.Format(time.RFC3339))
	default:
		return config.NewScalarValue(fmt.Sprintf("%v", val))
	}
}

// WriteTOML writes a config Tree as TOML.
func WriteTOML(tree *config.Tree, path string) error {
	data, err := MarshalTOML(tree)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// MarshalTOML marshals a config Tree to TOML bytes.
func MarshalTOML(tree *config.Tree) ([]byte, error) {
	goVal := convertToGoTOML(tree.Root)
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(goVal); err != nil {
		return nil, fmt.Errorf("encoding TOML: %w", err)
	}
	return buf.Bytes(), nil
}

func convertToGoTOML(v *config.Value) interface{} {
	if v == nil {
		return nil
	}
	switch v.Kind {
	case config.KindNil:
		return nil
	case config.KindScalar:
		return v.Scalar
	case config.KindMap:
		m := make(map[string]interface{})
		for _, child := range v.Children {
			m[child.Key] = convertToGoTOML(child.Value)
		}
		return m
	case config.KindList:
		l := make([]interface{}, 0, len(v.Children))
		for _, child := range v.Children {
			l = append(l, convertToGoTOML(child.Value))
		}
		return l
	}
	return nil
}
