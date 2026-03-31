# ddctl -- Codebase Guide

This is a Go CLI tool for managing Datadog resources. This document is for LLMs working on the codebase.

## Architecture

ddctl is a Cobra-based CLI with SQLite for local state and the Datadog SDK for API access. Resources (dashboards, monitors, SLOs) follow a push/pull/version pattern similar to git.

```
ddctl/
├── cmd/                     # Cobra command tree (one package per command group)
│   ├── root.go              # Root command, global flags, command registration
│   ├── version.go           # Version command with ldflags
│   ├── helpers.go           # Shared command helpers
│   ├── cmdutil/cmdutil.go   # Dependency injection (store, client, config)
│   ├── dashboard/           # dashboard subcommands (list, pull, push, edit, etc.)
│   ├── monitor/             # monitor subcommands
│   ├── slo/                 # slo subcommands
│   ├── metrics/             # metrics search, query
│   ├── logs/                # logs search
│   ├── api/                 # API pass-through (list, call, raw)
│   ├── connection/          # connection add, list, default, remove, test
│   ├── auth/                # auth login, status, logout
│   ├── mcp/                 # mcp serve (MCP server mode)
│   └── db/                  # db prune, stats
├── internal/
│   ├── client/              # Datadog API client wrapper (rate-limited, connection-aware)
│   ├── store/               # SQLite operations (connections, resources, versions, migrations)
│   ├── mcp/                 # MCP server implementation (server.go + tools.go)
│   ├── config/              # YAML config parsing with defaults
│   ├── output/              # JSON, YAML, table output formatters
│   ├── keyring/             # Credential resolution (env vars -> keychain -> file)
│   └── theme/               # Terminal styling, ASCII art banner, color palette
├── examples/                # Example YAML files for browsing in repo
│   ├── dashboards/          # Dashboard examples
│   ├── monitors/            # Monitor examples
│   └── slos/                # SLO examples
├── cmd/dashboard/examples/  # Embedded examples (go:embed for CLI)
├── main.go                  # Entry point
├── go.mod / go.sum
├── Makefile
└── docs/                    # Documentation
```

## Key Packages

### cmd/cmdutil
`InitDeps()` is the central dependency injection point. It loads config, opens the SQLite store, resolves the connection profile, and optionally creates an API client. Every command that needs state calls `InitDeps(cmd, requireClient)`.

### internal/store
SQLite store with WAL mode, foreign keys, and auto-migrations. Core tables: `connections`, `resources`, `resource_versions`, `schema_migrations`.

### internal/client
Thin wrapper around the Datadog Go SDK. Provides typed methods for dashboards, monitors, SLOs, metrics, and logs. Includes rate limiting via a simple throttle mechanism.

### internal/mcp
MCP server using mark3labs/mcp-go. Exposes CLI operations as MCP tools with configurable safety levels. Each tool accepts parameters matching CLI flags and returns JSON.

### internal/keyring
Credential resolution chain: connection-specific env vars -> generic env vars -> system keychain. Uses zalando/go-keyring.

## Build, Test, Lint

```bash
# Build
make build                    # or: go build -o ddctl .

# Test
make test                     # or: go test ./...

# Lint
make lint                     # or: go vet ./...

# Build with version info
make build                    # uses ldflags for Version, CommitSHA, BuildDate

# Install
make install                  # go install with ldflags
```

## Key Conventions

- **All resource types follow the same pattern**: list, pull, push, edit, diff, history, rollback, import, export, delete, create
- **YAML is the canonical format**: Resources are stored as YAML in SQLite. Converted to JSON for API calls.
- **Version variables** (`Version`, `CommitSHA`, `BuildDate`) are in `cmd/version.go`, set via ldflags at build time
- **No CGO**: SQLite via modernc.org/sqlite (pure Go), keyring via zalando/go-keyring
- **Output format**: Commands check `deps.Format` and branch on "json", "yaml", "table"
- **Error handling**: Commands return errors from `RunE`; the `output.Error()` function formats structured errors
- **Global flags**: `--connection`, `--output`, `--yes`, `--debug` defined in `cmd/root.go`

## Testing Patterns

- Store tests use `t.TempDir()` for isolated SQLite databases
- Config tests create temp YAML files
- Keyring tests use the real keychain (macOS Keychain on macOS)
- No mocks for the Datadog API client yet -- integration tests are planned

## Adding a New Resource Type

1. Create `internal/client/<type>.go` with CRUD methods
2. Create `cmd/<type>/` with the standard command set (copy from dashboard/)
3. Register in `cmd/root.go`
4. Add MCP tools in `internal/mcp/tools.go`
5. Use `resource_type = "<type>"` in store operations -- no schema changes needed

## MCP Server

The MCP server reuses the store and client directly (no shelling out to CLI). Safety levels control which tools are registered:
- `read-only`: list, search, query, diff, history
- `read-write`: adds create, push, pull, edit
- `unrestricted`: adds delete, rollback (still require `confirm: true`)
