package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/configforge/pkg/check"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/convert"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/diff"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/lint"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/merge"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/parser"
	"github.com/EdgarOrtegaRamirez/configforge/pkg/query"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "configforge",
		Short: "Universal config file toolkit",
		Long:  `ConfigForge - Parse, lint, diff, merge, convert, query, and check config files across formats (YAML, TOML, JSON, INI, .env).`,
	}

	rootCmd.AddCommand(
		newLintCmd(),
		newDiffCmd(),
		newMergeCmd(),
		newConvertCmd(),
		newQueryCmd(),
		newCheckCmd(),
		newInfoCmd(),
		newVersionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newLintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lint [file]",
		Short: "Lint a config file for issues",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tree, err := parser.ParseFile(args[0])
			if err != nil {
				return err
			}
			result := lint.Lint(tree)
			if len(result.Issues) == 0 {
				fmt.Println("✓ No issues found")
				return nil
			}
			for _, issue := range result.Issues {
				fmt.Printf("[%s] %s: %s (%s)\n", issue.Severity, issue.Path, issue.Message, issue.Rule)
			}
			fmt.Printf("\n%d issue(s) found\n", len(result.Issues))
			return nil
		},
	}
}

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <file1> <file2>",
		Short: "Semantic diff between two config files",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := parser.ParseFile(args[0])
			if err != nil {
				return fmt.Errorf("parsing %s: %w", args[0], err)
			}
			b, err := parser.ParseFile(args[1])
			if err != nil {
				return fmt.Errorf("parsing %s: %w", args[1], err)
			}
			format, _ := cmd.Flags().GetString("format")
			result := diff.Diff(a, b)
			switch format {
			case "json":
				fmt.Print(diff.FormatJSON(result))
			default:
				fmt.Print(diff.FormatText(result))
			}
			return nil
		},
	}
	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")
	return cmd
}

func newMergeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge <file1> <file2> -o <output>",
		Short: "Deep merge two config files",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := parser.ParseFile(args[0])
			if err != nil {
				return fmt.Errorf("parsing %s: %w", args[0], err)
			}
			b, err := parser.ParseFile(args[1])
			if err != nil {
				return fmt.Errorf("parsing %s: %w", args[1], err)
			}
			strategy, _ := cmd.Flags().GetString("strategy")
			output, _ := cmd.Flags().GetString("output")

			var strat merge.Strategy
			switch strategy {
			case "last-wins":
				strat = merge.LastWins
			case "first-wins":
				strat = merge.FirstWins
			case "union":
				strat = merge.Union
			default:
				strat = merge.LastWins
			}

			result, err := merge.Merge(a, b, strat)
			if err != nil {
				return err
			}

			if len(result.Conflicts) > 0 {
				fmt.Printf("⚠ %d conflict(s) found:\n", len(result.Conflicts))
				for _, c := range result.Conflicts {
					fmt.Printf("  %s: %s vs %s\n", c.Path, c.A, c.B)
				}
			}

			if output != "" {
				format := parser.DetectFormat(output)
				switch format {
				case "json":
					return parser.WriteJSON(result.Tree, output)
				case "yaml":
					return parser.WriteYAML(result.Tree, output)
				case "toml":
					return parser.WriteTOML(result.Tree, output)
				case "ini":
					return parser.WriteINI(result.Tree, output)
				default:
					return parser.WriteJSON(result.Tree, output)
				}
			}

			data, err := parser.MarshalJSON(result.Tree)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
			return nil
		},
	}
	cmd.Flags().StringP("strategy", "s", "last-wins", "Merge strategy (last-wins, first-wins, union)")
	cmd.Flags().StringP("output", "o", "", "Output file")
	return cmd
}

func newConvertCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "convert <input> <output>",
		Short: "Convert between config formats",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return convert.Convert(args[0], args[1])
		},
	}
}

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query <file> <expression>",
		Short: "Query a config file with path expressions",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tree, err := parser.ParseFile(args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			results, err := query.Query(tree, args[1])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Println("No results")
				return nil
			}
			fmt.Print(query.FormatResults(results, format))
			return nil
		},
	}
	cmd.Flags().StringP("format", "f", "both", "Output format (paths, values, both)")
	return cmd
}

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [file]",
		Short: "Security and quality check for config files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tree, err := parser.ParseFile(args[0])
			if err != nil {
				return err
			}
			result := check.Check(tree)
			if len(result.Issues) == 0 {
				fmt.Printf("✓ No issues found (score: %d/100)\n", result.Score)
				return nil
			}
			for _, issue := range result.Issues {
				fmt.Printf("[%s] %s: %s\n", issue.Type, issue.Path, issue.Message)
			}
			fmt.Printf("\nScore: %d/100 (%d issue(s) found)\n", result.Score, len(result.Issues))
			return nil
		},
	}
}

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [file]",
		Short: "Show config file information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tree, err := parser.ParseFile(args[0])
			if err != nil {
				return err
			}
			flat := tree.Flatten()
			fmt.Printf("Format:   %s\n", tree.Format)
			fmt.Printf("Source:   %s\n", args[0])
			fmt.Printf("Entries:  %d\n", len(flat))
			fmt.Printf("Keys:\n")
			keys := make([]string, 0, len(flat))
			for k := range flat {
				keys = append(keys, k)
			}
			for _, k := range keys {
				fmt.Printf("  %s = %s\n", k, truncate(flat[k], 60))
			}
			return nil
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("configforge v%s\n", version)
		},
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func init() {
	// Suppress cobra's auto-generated completions
	_ = strings.TrimSpace
}
