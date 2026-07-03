package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
)

// ParseJSON parses a JSON file into a config Tree.
func ParseJSON(path string) (*config.Tree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return ParseJSONBytes(data)
}

// ParseJSONBytes parses JSON bytes into a config Tree.
func ParseJSONBytes(data []byte) (*config.Tree, error) {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	tree := config.NewTree()
	tree.Format = "json"
	tree.Root = convertFromGo(raw)
	return tree, nil
}

func convertFromGo(v interface{}) *config.Value {
	if v == nil {
		return &config.Value{Kind: config.KindNil}
	}
	switch val := v.(type) {
	case map[string]interface{}:
		m := config.NewMapValue()
		for k, child := range val {
			m.Children = append(m.Children, &config.Entry{
				Key:   k,
				Value: convertFromGo(child),
			})
		}
		return m
	case []interface{}:
		l := config.NewListValue()
		for _, child := range val {
			l.Children = append(l.Children, &config.Entry{
				Key:   "",
				Value: convertFromGo(child),
			})
		}
		return l
	case string:
		return config.NewScalarValue(val)
	case float64:
		return config.NewScalarValue(val)
	case bool:
		return config.NewScalarValue(val)
	default:
		return config.NewScalarValue(fmt.Sprintf("%v", val))
	}
}

// WriteJSON writes a config Tree as formatted JSON.
func WriteJSON(tree *config.Tree, path string) error {
	data, err := MarshalJSON(tree)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// MarshalJSON marshals a config Tree to JSON bytes.
func MarshalJSON(tree *config.Tree) ([]byte, error) {
	goVal := convertToGo(tree.Root)
	return json.MarshalIndent(goVal, "", "  ")
}

func convertToGo(v *config.Value) interface{} {
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
			m[child.Key] = convertToGo(child.Value)
		}
		return m
	case config.KindList:
		l := make([]interface{}, 0, len(v.Children))
		for _, child := range v.Children {
			l = append(l, convertToGo(child.Value))
		}
		return l
	}
	return nil
}

// DetectFormat detects the config format from file extension.
func DetectFormat(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".json"):
		return "json"
	case strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml"):
		return "yaml"
	case strings.HasSuffix(lower, ".toml"):
		return "toml"
	case strings.HasSuffix(lower, ".ini") || strings.HasSuffix(lower, ".cfg") || strings.HasSuffix(lower, ".conf"):
		return "ini"
	case strings.HasSuffix(lower, ".env") || strings.HasSuffix(lower, ".env.*") || strings.HasPrefix(lower, ".env"):
		return "dotenv"
	default:
		return "json"
	}
}

// ParseFile auto-detects format and parses a config file.
func ParseFile(path string) (*config.Tree, error) {
	format := DetectFormat(path)
	switch format {
	case "json":
		return ParseJSON(path)
	case "yaml":
		return ParseYAML(path)
	case "toml":
		return ParseTOML(path)
	case "ini":
		return ParseINI(path)
	case "dotenv":
		return ParseDotenv(path)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
