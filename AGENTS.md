# ConfigForge - AI Agent Guide

## Overview

ConfigForge is a universal config file toolkit in Go. It parses, lints, diffs, merges, converts, and queries config files across multiple formats (YAML, TOML, JSON, INI, .env).

## Building

```bash
go build -o configforge ./cmd/configforge/
```

## Running Tests

```bash
go test ./...
```

## Project Structure

- `pkg/config/` - Universal tree data structure (Value, Entry, Tree)
- `pkg/parser/` - Format-specific parsers (JSON, YAML, TOML, INI, dotenv)
- `pkg/diff/` - Semantic diff engine (tree comparison)
- `pkg/merge/` - Deep merge with conflict resolution strategies
- `pkg/lint/` - Config linting rules (empty values, secrets, URLs)
- `pkg/query/` - Path-based querying (dot notation, wildcards)
- `pkg/check/` - Security checks (secret detection, weak passwords)
- `pkg/convert/` - Format conversion between all supported formats
- `cmd/configforge/` - Cobra CLI with 8 commands

## Key Patterns

- All parsers convert to/from `config.Tree` (universal representation)
- Tree uses `Value` type with `Kind` enum (Scalar, Map, List, Nil)
- Path notation uses dots: `server.port`, `items[0].name`
- Merge uses Strategy pattern: LastWins, FirstWins, Union
- Lint/Check use Rule interface for extensibility

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/BurntSushi/toml` - TOML parsing
