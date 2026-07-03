// Package convert provides format conversion between config formats.
package convert

import (
	"fmt"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
)

// Convert reads a config file and writes it in another format.
func Convert(inputPath, outputPath string) error {
	tree, err := parser.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("parsing input: %w", err)
	}

	format := parser.DetectFormat(outputPath)
	switch format {
	case "json":
		return parser.WriteJSON(tree, outputPath)
	case "yaml":
		return parser.WriteYAML(tree, outputPath)
	case "toml":
		return parser.WriteTOML(tree, outputPath)
	case "ini":
		return parser.WriteINI(tree, outputPath)
	case "dotenv":
		return parser.WriteDotenv(tree, outputPath)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// ConvertBytes converts config data from one format to another.
func ConvertBytes(data []byte, fromFormat, toFormat string) ([]byte, error) {
	var tree *config.Tree
	var err error

	switch fromFormat {
	case "json":
		tree, err = parser.ParseJSONBytes(data)
	case "yaml":
		tree, err = parser.ParseYAMLBytes(data)
	case "toml":
		tree, err = parser.ParseTOMLBytes(data)
	case "ini":
		tree, err = parser.ParseINIBytes(data)
	case "dotenv":
		tree, err = parser.ParseDotenvBytes(data)
	default:
		return nil, fmt.Errorf("unsupported input format: %s", fromFormat)
	}

	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}

	switch toFormat {
	case "json":
		return parser.MarshalJSON(tree)
	case "yaml":
		return parser.MarshalYAML(tree)
	case "toml":
		return parser.MarshalTOML(tree)
	case "ini":
		return parser.MarshalINI(tree)
	case "dotenv":
		return parser.MarshalDotenv(tree)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", toFormat)
	}
}
