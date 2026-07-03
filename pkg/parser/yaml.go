package parser

import (
	"fmt"
	"os"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
	"gopkg.in/yaml.v3"
)

// ParseYAML parses a YAML file into a config Tree.
func ParseYAML(path string) (*config.Tree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return ParseYAMLBytes(data)
}

// ParseYAMLBytes parses YAML bytes into a config Tree.
func ParseYAMLBytes(data []byte) (*config.Tree, error) {
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	tree := config.NewTree()
	tree.Format = "yaml"
	tree.Root = convertFromYAML(raw)
	return tree, nil
}

func convertFromYAML(v interface{}) *config.Value {
	if v == nil {
		return &config.Value{Kind: config.KindNil}
	}
	switch val := v.(type) {
	case map[string]interface{}:
		m := config.NewMapValue()
		for k, child := range val {
			m.Children = append(m.Children, &config.Entry{
				Key:   k,
				Value: convertFromYAML(child),
			})
		}
		return m
	case []interface{}:
		l := config.NewListValue()
		for _, child := range val {
			l.Children = append(l.Children, &config.Entry{
				Key:   "",
				Value: convertFromYAML(child),
			})
		}
		return l
	case string:
		return config.NewScalarValue(val)
	case int:
		return config.NewScalarValue(float64(val))
	case int64:
		return config.NewScalarValue(float64(val))
	case float64:
		return config.NewScalarValue(val)
	case bool:
		return config.NewScalarValue(val)
	default:
		return config.NewScalarValue(fmt.Sprintf("%v", val))
	}
}

// WriteYAML writes a config Tree as YAML.
func WriteYAML(tree *config.Tree, path string) error {
	data, err := MarshalYAML(tree)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// MarshalYAML marshals a config Tree to YAML bytes.
func MarshalYAML(tree *config.Tree) ([]byte, error) {
	goVal := convertToGoYAML(tree.Root)
	return yaml.Marshal(goVal)
}

func convertToGoYAML(v *config.Value) interface{} {
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
			m[child.Key] = convertToGoYAML(child.Value)
		}
		return m
	case config.KindList:
		l := make([]interface{}, 0, len(v.Children))
		for _, child := range v.Children {
			l = append(l, convertToGoYAML(child.Value))
		}
		return l
	}
	return nil
}
