package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/config"
)

// ParseINI parses an INI file into a config Tree.
func ParseINI(path string) (*config.Tree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return ParseINIBytes(data)
}

// ParseINIBytes parses INI bytes into a config Tree.
func ParseINIBytes(data []byte) (*config.Tree, error) {
	tree := config.NewTree()
	tree.Format = "ini"

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	currentSection := ""
	sectionMap := tree.Root

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || line[0] == ';' || line[0] == '#' {
			continue
		}

		// Section header
		if line[0] == '[' && line[len(line)-1] == ']' {
			currentSection = line[1 : len(line)-1]
			sectionMap = config.NewMapValue()
			tree.Root.Set(currentSection, sectionMap)
			continue
		}

		// Key = Value
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			// Strip quotes
			value = strings.Trim(value, "\"'")
			sectionMap.SetKey(key, config.NewScalarValue(value))
		}
	}

	return tree, scanner.Err()
}

// WriteINI writes a config Tree as INI.
func WriteINI(tree *config.Tree, path string) error {
	data, err := MarshalINI(tree)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// MarshalINI marshals a config Tree to INI bytes.
func MarshalINI(tree *config.Tree) ([]byte, error) {
	var sb strings.Builder

	// Global keys (entries directly in root before sections)
	for _, child := range tree.Root.Children {
		if child.Value.Kind == config.KindScalar {
			sb.WriteString(fmt.Sprintf("%s = %v\n", child.Key, child.Value.Scalar))
		}
	}

	// Sections
	for _, child := range tree.Root.Children {
		if child.Value.Kind == config.KindMap {
			sb.WriteString(fmt.Sprintf("\n[%s]\n", child.Key))
			for _, entry := range child.Value.Children {
				if entry.Value.Kind == config.KindScalar {
					sb.WriteString(fmt.Sprintf("%s = %v\n", entry.Key, entry.Value.Scalar))
				}
			}
		}
	}

	return []byte(sb.String()), nil
}
