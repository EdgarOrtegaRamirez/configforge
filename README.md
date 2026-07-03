# ConfigForge 🔧

Universal config file toolkit — parse, lint, diff, merge, convert, query, and check config files across formats.

## Features

- **Multi-format support**: YAML, TOML, JSON, INI, .env
- **Lint**: Detect secrets, empty values, insecure defaults, hardcoded URLs
- **Diff**: Semantic diff between any two config files (same or different formats)
- **Merge**: Deep merge with conflict resolution (last-wins, first-wins, union)
- **Convert**: Convert between any supported formats
- **Query**: Dot-notation path queries with wildcards
- **Check**: Security scoring, secret detection, weak password detection
- **Info**: Show config file structure and metadata

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/configforge/cmd/configforge@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/configforge
cd configforge
go build -o configforge ./cmd/configforge/
```

## Quick Start

```bash
# Show config info
configforge info config.yaml

# Lint for issues
configforge lint config.yaml

# Security check with score
configforge check config.yaml

# Diff two configs
configforge diff config.yaml config.production.yaml

# Convert formats
configforge convert config.json config.yaml
configforge convert config.toml config.json

# Query values
configforge query config.yaml "server.port"
configforge query config.yaml "servers.*.host"
configforge query config.yaml "database.**"

# Merge configs (later values win conflicts)
configforge merge base.yaml override.yaml -o merged.yaml
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `info <file>` | Show config file structure and metadata |
| `lint <file>` | Lint config for issues (secrets, empty values, etc.) |
| `check <file>` | Security and quality check with scoring |
| `diff <file1> <file2>` | Semantic diff between two config files |
| `merge <file1> <file2>` | Deep merge with conflict resolution |
| `convert <input> <output>` | Convert between formats (JSON, YAML, TOML, INI, .env) |
| `query <file> <expr>` | Query with dot-notation path expressions |
| `version` | Print version |

## Supported Formats

| Format | Extensions | Read | Write |
|--------|-----------|------|-------|
| JSON | `.json` | ✅ | ✅ |
| YAML | `.yaml`, `.yml` | ✅ | ✅ |
| TOML | `.toml` | ✅ | ✅ |
| INI | `.ini`, `.cfg`, `.conf` | ✅ | ✅ |
| dotenv | `.env`, `.env.*` | ✅ | ✅ |

## Query Syntax

```bash
# Simple path
configforge query config.yaml "server.host"

# Wildcard (all children)
configforge query config.yaml "servers.*"

# Deep wildcard
configforge query config.yaml "servers.*.port"

# Recursive descent
configforge query config.yaml "**.host"

# Array indexing
configforge query config.yaml "items[0].name"
```

## Lint Rules

| Rule | Severity | Description |
|------|----------|-------------|
| `empty-value` | Warning | Empty strings or null values |
| `secret-detection` | Error | Potential hardcoded secrets (passwords, API keys, tokens) |
| `hardcoded-url` | Info | URLs that should be configurable |
| `deprecated-key` | Warning | Deprecated configuration keys |
| `type-consistency` | Warning | Mixed types in arrays |

## Security Checks

| Check | Score Impact | Description |
|-------|-------------|-------------|
| Secret detection | -30 | Hardcoded passwords, API keys, tokens |
| Weak password | -20 | Common or short passwords |
| Insecure default | -10 | Debug mode enabled, SSL verification disabled |
| Unsafe setting | -15 | eval, shell execution, auth disabled |
| Empty sensitive | -5 | Empty password or API key fields |

## Merge Strategies

| Strategy | Behavior |
|----------|----------|
| `last-wins` | Later value overwrites earlier (default) |
| `first-wins` | Earlier value takes precedence |
| `union` | Combine both values (for lists) |

## Architecture

```
configforge/
├── cmd/configforge/     # CLI entry point
├── pkg/
│   ├── config/          # Universal config tree data structure
│   ├── parser/          # Format-specific parsers (JSON, YAML, TOML, INI, .env)
│   ├── diff/            # Semantic diff engine
│   ├── merge/           # Deep merge with conflict resolution
│   ├── lint/            # Config linting rules
│   ├── query/           # Path-based querying
│   ├── check/           # Security and quality checks
│   └── convert/         # Format conversion
└── tests/               # Comprehensive test suite
```

## License

MIT
