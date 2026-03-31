# ddctl — Datadog Control CLI

**Date:** 2026-03-30
**Author:** Gerald Onyango
**Status:** Draft

## Overview

`ddctl` is a Go CLI for managing Datadog resources (dashboards, monitors, SLOs) with built-in versioning, multi-org support, and an MCP server mode for LLM integration. Designed for engineers who want to manage Datadog infrastructure from the terminal, and for LLM agents that need programmatic Datadog access beyond the read-only official MCP server.

### Design Principles

- **Generic** — not tied to any specific organization; works for any Datadog customer
- **LLM-first** — structured output by default, MCP server mode, declarative YAML definitions
- **Self-contained** — SQLite for local state and versioning; no external dependencies beyond Datadog API keys
- **Safe by default** — never overwrite remote resources without confirmation, diff, and version snapshot
- **Fun** — retro cyberpunk terminal aesthetic inspired by Deus Ex, 90s hacker culture, and early BBS systems

### Audience

All engineers (and any Datadog user). The CLI and MCP server are designed for both human operators and LLM agents as first-class consumers.

## Architecture

### Project Structure

```
ddctl/
├── cmd/                    # Cobra command tree
│   ├── root.go             # Global flags (--connection, --output, --yes, --debug)
│   ├── dashboard/          # dashboard create|import|push|pull|sync|list|edit|delete|diff|history|rollback|export|examples
│   ├── monitor/            # monitor create|import|push|pull|list|edit|delete|history|rollback|export
│   ├── slo/                # slo create|import|push|pull|list|edit|delete|history|rollback|export
│   ├── metrics/            # metrics search|query|context
│   ├── logs/               # logs search|analyze
│   ├── api/                # api list|<group>.<operation>|raw <method> <path>
│   ├── connection/         # connection add|list|default|remove|test
│   ├── auth/               # auth login|status|logout
│   ├── mcp/                # mcp serve
│   └── db/                 # db prune|stats|migrate
├── internal/
│   ├── client/             # Datadog API client wrapper (connection-aware, rate-limited)
│   ├── store/              # SQLite operations (versions, connections, metadata, migrations)
│   ├── dsl/                # Dashboard/monitor/SLO DSL parser + translator to/from raw API format
│   ├── mcp/                # MCP server implementation
│   ├── output/             # JSON / YAML / table formatters
│   ├── theme/              # Terminal styling, ASCII art, color palette, flavor text
│   ├── keyring/            # Cross-platform keychain abstraction
│   └── apigen/             # Generated SDK operation bindings
├── gen/                    # Code generation tooling
│   └── apigen.go           # Parses Datadog OpenAPI spec, generates operation bindings
├── examples/               # Embedded example dashboards, monitors, SLOs
│   ├── dashboards/
│   │   ├── service-overview.yaml
│   │   ├── log-analytics.yaml
│   │   ├── slo-tracking.yaml
│   │   └── deployment.yaml
│   ├── monitors/
│   │   └── error-rate-alert.yaml
│   └── slos/
│       └── availability-slo.yaml
├── docs/
│   ├── getting-started.md
│   ├── configuration.md
│   ├── authentication.md
│   ├── dashboards.md
│   ├── monitors.md
│   ├── slos.md
│   ├── api-passthrough.md
│   ├── mcp-server.md
│   ├── versioning.md
│   └── dsl-reference.md
├── README.md
├── CLAUDE.md
├── go.mod
└── go.sum
```

### Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/DataDog/datadog-api-client-go/v2` | Datadog SDK |
| `github.com/zalando/go-keyring` | Cross-platform keychain (macOS Keychain, Linux libsecret, Windows Credential Manager) |
| `modernc.org/sqlite` | Pure Go SQLite (no CGO required) |
| `github.com/mark3labs/mcp-go` | MCP server SDK |
| `github.com/charmbracelet/lipgloss` | Terminal styling and colors |
| `gopkg.in/yaml.v3` | Config and DSL YAML parsing |

Note: Viper was considered for config management but rejected in favor of a simple YAML config struct. The config needs are minimal (5 keys) and don't justify Viper's dependency weight.

## Configuration

Config lives at `~/.config/ddctl/config.yaml` (respects `XDG_CONFIG_HOME`). SQLite database at `~/.config/ddctl/ddctl.db`.

```yaml
# ~/.config/ddctl/config.yaml
output: json                # json | table | yaml
default_connection: production
editor: cursor              # cursor | code | vim | nvim | nano
theme: retro                # retro | minimal | none
versions_to_keep: 50        # max versions per resource (0 = unlimited)
```

### Config Options

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `output` | `json`, `table`, `yaml` | `json` | Default output format |
| `default_connection` | string | (first added) | Connection profile to use when `--connection` not specified |
| `editor` | `cursor`, `code`, `vim`, `nvim`, `nano`, or any command | `$EDITOR` | Editor for `edit` commands |
| `theme` | `retro`, `minimal`, `none` | `retro` | Terminal decoration level |
| `versions_to_keep` | integer | `50` | Max versions per resource before auto-pruning |

### Editor Support

| Editor | Command Used |
|--------|----------|
| `vim` / `nvim` / `nano` | Opens in terminal, blocks naturally |
| `cursor` | `cursor --wait <file>` |
| `code` | `code --wait <file>` |

JSON/YAML output modes (`--output json`, `--output yaml`) automatically suppress all theme decoration regardless of theme setting.

### Global Flags

| Flag | Description |
|------|-------------|
| `-c, --connection` | Connection profile to use |
| `-o, --output` | Output format (json, table, yaml) |
| `-y, --yes` | Skip confirmation prompts |
| `--debug` | Enable debug output (HTTP requests/responses with key redaction, SQLite queries, auth resolution steps) to stderr |
| `-h, --help` | Help for any command |

## Authentication

### Resolution Order (highest priority first)

1. **Environment variables** — `DD_API_KEY` / `DD_APP_KEY` (or `DD_<CONNECTION>_API_KEY` / `DD_<CONNECTION>_APP_KEY` for named profiles)
2. **System keychain** — stored per connection name, written by `ddctl auth login`
3. **Credential file** — `~/.config/ddctl/credentials` (mode `0600`), for headless/CI/Docker environments where keychain is unavailable
4. **Error with setup instructions** — if none found

API keys are **never** stored in the config file or SQLite database.

> **Security note:** Environment variables are visible in process listings (`ps auxe`). Keychain is the recommended method for developer workstations. Use env vars or credential files only for CI/headless environments where keychain is unavailable. `ddctl auth status` will warn when credentials are sourced from env vars.

### Auth Backends

| Backend | Best For | How |
|---------|----------|-----|
| Keychain | Developer workstations | `ddctl auth login` (default) |
| Environment variables | CI/CD, MCP server | `DD_API_KEY` / `DD_APP_KEY` |
| Credential file | Headless servers, Docker | `ddctl auth login --backend file` |

The keychain backend requires a desktop session (macOS Keychain, Linux libsecret with D-Bus, Windows Credential Manager). In headless environments (SSH, Docker, CI), the credential file backend is used automatically as fallback when keychain is unavailable.

### Auth Commands

```bash
# Store keys in keychain (prompts, validates, stores)
ddctl auth login

# Store keys in credential file (for headless environments)
ddctl auth login --backend file

# Show where keys are being read from (warns if using env vars)
ddctl auth status

# Remove keys from keychain/credential file
ddctl auth logout
```

`ddctl auth login` prompts for API key + app key, validates them with a test API call, and stores in the selected backend under `ddctl/<connection-name>/api_key` and `ddctl/<connection-name>/app_key`.

### Connection Management

```bash
# Add a connection (prompts for keys, stores in keychain)
ddctl connection add --name production --site datadoghq.com

# Test connectivity
ddctl connection test production

# Set default
ddctl connection default production

# List connections
ddctl connection list

# Override per-command
ddctl dashboard list --connection staging

# Remove a connection (removes keys from keychain too)
ddctl connection remove staging
```

### SQLite Schema — Connections

```sql
CREATE TABLE connections (
    name        TEXT PRIMARY KEY,
    site        TEXT NOT NULL,
    is_default  BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Resource Management (Dashboards, Monitors, SLOs)

All three resource types follow the same push/pull/edit/history/rollback pattern. SQLite is local state, Datadog is remote state. `push`/`pull` syncs between them.

### Command Pattern

Every managed resource type supports:

| Command | Description | Requires Network |
|---------|-------------|-----------------|
| `create` | Create a new resource interactively or from flags | Yes |
| `import -f <file>` | Import a YAML file into local SQLite (`--push` to also push to remote) | Only with `--push` |
| `push <id>` | Push local state to Datadog (diffs, confirms, versions) | Yes |
| `pull <id>` | Pull remote state into SQLite (creates new local version) | Yes |
| `sync` | Pull all tracked resources from remote | Yes |
| `list` | List locally tracked resources (`--remote` for all remote resources) | Only with `--remote` |
| `edit <id>` | Open in editor from SQLite, save back to SQLite | No |
| `delete <id>` | Delete from remote (with confirmation) | Yes |
| `diff <id>` | Diff local vs remote, or against a specific version | Against remote: Yes |
| `history <id>` | Show version history | No |
| `rollback <id> --to-version N` | Rollback to a previous version (creates a new version) | No (local only; `push` separately) |
| `export <id> -o <file>` | Export from SQLite to a YAML file | No |

### Dashboard-Specific Commands

```bash
# List bundled examples
ddctl dashboard examples

# Preview or import an example
ddctl dashboard examples service-overview --preview
ddctl dashboard examples service-overview --import
```

### Push Safety

Every `push` operation follows this exact sequence:

1. Fetch current remote state (snapshot A)
2. Diff local state against snapshot A
3. Display who last modified the remote resource and when
4. Ask for explicit confirmation (`--yes` to skip)
5. Re-fetch current remote state (snapshot B)
6. Compare snapshot A and snapshot B — if they differ, **abort** and inform the user that the remote changed during confirmation (TOCTOU protection)
7. Store snapshot B in SQLite as `remote_snapshot`
8. Apply the local state to remote
9. Record a new version in SQLite

### Conflict Resolution

**`sync` (pull all)** — remote always wins. Pulling creates a new local version, so nothing is ever lost. If a resource was modified both locally (in SQLite) and remotely since the last sync:

- The remote state becomes the latest local version
- The previous local state is preserved in version history
- A warning is shown: "Dashboard abc-123 had local changes that were not pushed. Previous local state preserved as version N."

**`push`** — refuses if remote has changed since the last pull (optimistic locking). The user must `pull` first to incorporate remote changes, then `push` again:

```
◄ PUSH REJECTED ► Remote modified since last pull.
  Remote modified: 2026-03-30 18:45:00 UTC by molly.finn@cobalt.io
  Last pulled:     2026-03-30 14:00:00 UTC

  Run 'ddctl dashboard pull abc-123' first, then retry push.
```

**Remote deletion** — if a tracked resource was deleted from Datadog, `sync` marks it as `deleted_remotely` in the `resources` table and warns the user. The version history is preserved.

### SQLite Schema — Resources & Versions

```sql
-- Required pragmas (set at connection time)
-- PRAGMA journal_mode=WAL;
-- PRAGMA foreign_keys=ON;
-- PRAGMA busy_timeout=5000;

CREATE TABLE schema_migrations (
    version     INTEGER PRIMARY KEY,
    applied_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE resources (
    resource_id         TEXT NOT NULL,
    resource_type       TEXT NOT NULL,       -- dashboard | monitor | slo
    connection          TEXT NOT NULL,
    title               TEXT,
    remote_modified_at  TIMESTAMP,
    remote_modified_by  TEXT,
    remote_etag         TEXT,                -- for optimistic locking if available
    last_synced_at      TIMESTAMP,
    status              TEXT DEFAULT 'active', -- active | deleted_remotely
    PRIMARY KEY (resource_id, resource_type, connection),
    FOREIGN KEY (connection) REFERENCES connections(name) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE resource_versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    resource_id     TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    connection      TEXT NOT NULL,
    version         INTEGER NOT NULL,
    content         TEXT NOT NULL,           -- full YAML at time of apply
    remote_snapshot TEXT,                    -- remote state before overwrite
    applied_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    applied_by      TEXT,                   -- system username
    message         TEXT,                   -- optional commit-style message
    UNIQUE(resource_id, resource_type, connection, version),
    FOREIGN KEY (resource_id, resource_type, connection)
        REFERENCES resources(resource_id, resource_type, connection)
        ON DELETE CASCADE
);

-- Indexes for common query patterns
CREATE INDEX idx_versions_lookup
    ON resource_versions(resource_id, resource_type, connection, version DESC);
CREATE INDEX idx_versions_applied_at
    ON resource_versions(applied_at);
CREATE INDEX idx_resources_connection
    ON resources(connection);
```

Adding new resource types later (notebooks, synthetics) requires only a new `resource_type` value and a Cobra command tree — no schema changes.

### Version Retention

Versions are auto-pruned per resource based on `versions_to_keep` config (default: 50). When a new version is created, if the count exceeds the limit, the oldest versions beyond the limit are deleted. `ddctl db prune` can also be run manually. `ddctl db stats` shows database size and version counts.

### SQLite Security

The database file is created with mode `0600` (owner-only read/write). The `content` and `remote_snapshot` columns may contain sensitive organizational information (service names, infrastructure details, query patterns). The database path should be added to `.gitignore` if it ever lives near a repository. Optional encryption at rest may be added in a future version.

### Schema Migrations

The `schema_migrations` table tracks applied migrations. On every `ddctl` invocation, pending migrations are run automatically before any database operation. Migrations are embedded in the binary via Go's `embed` package.

## Dashboard DSL

The simplified DSL covers common widget types. Anything the DSL doesn't model uses a `raw:` escape hatch.

### Example

```yaml
name: TPM Assignment Monitoring
description: Round-robin assignment tracking
layout: ordered
tags:
  - team:delivery
  - managed_by:ddctl

widgets:
  - type: timeseries
    title: Assignments per TPM (24h)
    query: >
      count:logs{@metadata.event:tpm_auto_assigned}
      by {@metadata.selected_tpm}.as_count()
    display: bars
    interval: 3600

  - type: query_value
    title: Total Assignments Today
    query: >
      count:logs{@metadata.event:tpm_auto_assigned}.as_count()
    timeframe: 1d

  - type: toplist
    title: TPM Utilization
    query: >
      avg:logs{@metadata.event:tpm_auto_assigned}
      by {@metadata.selected_tpm}

  - type: group
    title: Capacity Details
    layout: ordered
    widgets:
      - type: timeseries
        title: Over-Capacity Events
        query: >
          count:logs{@metadata.event:tpm_auto_assigned
          @metadata.all_over_capacity:true}.as_count()

  - raw:  # Escape hatch for unsupported widget types
      definition:
        type: scatterplot
        requests: { ... }
```

### Supported Widget Types (v1)

`timeseries`, `query_value`, `toplist`, `table`, `heatmap`, `note`, `group`, `alert_graph`, `slo`

All other widget types: use the `raw:` escape hatch block.

### Monitor DSL Example

```yaml
name: TPM All Over Capacity
type: logs_alert
query: >
  logs("@metadata.event:tpm_auto_assigned
  @metadata.all_over_capacity:true").index("*").rollup("count").last("15m") > 3
message: |
  All TPMs are over capacity. Last 15min had {{value}} over-capacity assignments.
  @slack-delivery-alerts
tags:
  - team:delivery
  - managed_by:ddctl
priority: 2
thresholds:
  critical: 3
  warning: 1
```

### SLO DSL Example

```yaml
name: TPM Assignment Success Rate
type: monitor
description: Percentage of assignments going to under-capacity TPMs
monitor_ids:
  - 12345678
target: 99.5
timeframe: 30d
warning: 99.9
tags:
  - team:delivery
```

## API Pass-through

### Named SDK Operations

Operations are generated from the Datadog OpenAPI spec via a `go generate` step. The generator (`gen/apigen.go`) parses the spec and produces typed operation bindings in `internal/apigen/`. When bumping the Datadog SDK version, re-running `go generate ./gen/...` regenerates the bindings to include any new operations.

```bash
# List all available API groups
ddctl api list

# List operations in a group
ddctl api list dashboards

# Execute an operation
ddctl api dashboards.list
ddctl api dashboards.get --id abc-123
ddctl api metrics.search --filter "tpm"
ddctl api monitors.create --body @monitor.json
```

### Raw HTTP Escape Hatch

For anything the generated bindings don't cover yet:

```bash
ddctl api raw GET /v2/some/new/endpoint
ddctl api raw POST /v1/dashboard --body @payload.json
```

## MCP Server Mode

`ddctl mcp serve` starts a stdio-transport MCP server. Every CLI command maps to an MCP tool automatically.

### Setup

**With keychain auth (recommended — no secrets in config files):**

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "ddctl",
      "args": ["mcp", "serve"]
    }
  }
}
```

This works when `ddctl auth login` was previously run. The MCP server reads keys from the keychain using the default connection profile.

**With a specific connection:**

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "ddctl",
      "args": ["mcp", "serve", "--connection", "production"]
    }
  }
}
```

**With environment variables (CI or when keychain is unavailable):**

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "ddctl",
      "args": ["mcp", "serve"],
      "env": {
        "DD_API_KEY": "...",
        "DD_APP_KEY": "..."
      }
    }
  }
}
```

> **Warning:** Putting secrets in `.mcp.json` env blocks means they may be committed to version control. Prefer keychain auth for developer workstations.

### MCP Safety Levels

The MCP server supports configurable safety levels to prevent destructive operations by LLM agents:

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "ddctl",
      "args": ["mcp", "serve", "--safety", "read-write"]
    }
  }
}
```

| Level | Behavior |
|-------|----------|
| `read-only` | Only read operations (list, search, query, diff, history). **Default.** |
| `read-write` | Read + non-destructive writes (create, push, pull, edit). Delete and rollback require `confirm: true` parameter. |
| `unrestricted` | All operations. Destructive operations still require `confirm: true` parameter. |

Even at `unrestricted`, destructive operations (delete, rollback, push to a resource modified by someone else) require the LLM to explicitly pass `confirm: true` in the tool call — an extra friction point that forces the LLM to make a deliberate choice rather than accidentally triggering a destructive action.

### Tool Mapping

| CLI Command | MCP Tool | Description |
|---|---|---|
| `ddctl dashboard list` | `dashboard_list` | List tracked dashboards |
| `ddctl dashboard list --remote` | `dashboard_list_remote` | List all remote dashboards |
| `ddctl dashboard push` | `dashboard_push` | Push local to remote |
| `ddctl dashboard pull` | `dashboard_pull` | Pull remote to local |
| `ddctl dashboard edit` | `dashboard_edit` | Update dashboard content (takes YAML string, no editor) |
| `ddctl dashboard create` | `dashboard_create` | Create new dashboard |
| `ddctl dashboard diff` | `dashboard_diff` | Diff local vs remote |
| `ddctl dashboard history` | `dashboard_history` | Version history |
| `ddctl dashboard rollback` | `dashboard_rollback` | Rollback to previous version |
| `ddctl monitor *` | `monitor_*` | Same pattern as dashboards |
| `ddctl slo *` | `slo_*` | Same pattern as dashboards |
| `ddctl metrics query` | `metrics_query` | Query timeseries |
| `ddctl metrics search` | `metrics_search` | Search available metrics |
| `ddctl logs search` | `logs_search` | Search logs |
| `ddctl logs analyze` | `logs_analyze` | Analyze logs with SQL |
| `ddctl api` | `api_call` | Pass-through SDK operation |

### MCP Behavior Differences from CLI

- Output is always JSON (no theme decoration)
- `dashboard_edit` accepts YAML content as a parameter instead of opening an editor
- Destructive operations require `confirm: true` parameter (see Safety Levels above)
- Tools include rich descriptions and parameter schemas so LLMs know what's available

## Error Handling

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication error (missing/invalid keys) |
| 3 | API error (rate limit, server error, not found) |
| 4 | Conflict error (remote modified, push rejected) |
| 5 | Validation error (invalid YAML, DSL parse error) |

### Error Output

- In `table` mode: human-readable error message to stderr
- In `json`/`yaml` mode: structured error object to stdout (`{"error": {"code": 3, "message": "...", "details": {...}}}`)
- MCP mode: standard MCP error response with code and message

### Retry Policy

Transient API errors are retried with exponential backoff and jitter:

- **429 (Rate Limited):** Respect `X-RateLimit-Remaining` and `X-RateLimit-Reset` headers. Wait until reset time plus jitter. Max 3 retries.
- **500/502/503:** Backoff starting at 1s, doubling each retry. Max 3 retries.
- **Network errors (timeout, DNS):** Same backoff as 5xx. Max 3 retries.
- **4xx (other):** No retry (client error).

### Rate Limiting

The `internal/client` package implements a token-bucket rate limiter that respects Datadog's `X-RateLimit-*` response headers. Default: 30 requests/minute (conservative, below most Datadog endpoint limits). Configurable via `--rate-limit` flag.

In MCP mode, the rate limiter is shared across all tool invocations to prevent LLM agents from overwhelming the API.

### Graceful Degradation

- **Keychain unavailable:** Fall back to credential file, then env vars, with a warning.
- **Network unavailable:** Local-only commands work normally. Network-requiring commands fail immediately with a clear error rather than hanging on timeout.
- **SQLite locked:** WAL mode + `busy_timeout=5000ms` handles most concurrent access. If still locked after timeout, error with "Database locked — is another ddctl process running?"

## Rate Limiting

Datadog enforces rate limits on all API endpoints (typically 60 requests/minute for most endpoints, some as low as 5/minute). `ddctl` respects these limits:

- The `internal/client` reads `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset` from every API response
- When `Remaining` is low, the client proactively slows down
- `sync` operations with many tracked resources are automatically paced
- `--rate-limit N` flag overrides the automatic rate limiting (requests per minute)

## Edit Workflow

The `edit` command follows this exact sequence:

1. Read resource YAML from SQLite
2. Write to a temp file in `$TMPDIR` with mode `0600`
3. Record file checksum (SHA-256) before opening editor
4. Open editor with appropriate flags (`--wait` for GUI editors)
5. On editor exit, compute new checksum
6. If unchanged, report "No changes" and clean up
7. If changed, validate YAML syntax
8. If valid, save to SQLite as a new local version
9. Clean up temp file
10. Report success and suggest `ddctl dashboard push <id>` to apply remotely

If `$EDITOR` is not set and no editor is configured, error with setup instructions.

## Terminal UX & Theme

### Accessibility

- Respects `NO_COLOR` environment variable (de facto standard) — disables all colors
- Respects `TERM=dumb` — disables colors and ASCII art
- All text meets WCAG AA contrast ratios (4.5:1 minimum) against dark terminal backgrounds
- ASCII art and flavor text only shown on `ddctl` with no subcommand (not on every invocation)
- `theme: none` provides plain unformatted output for screen readers and piping

### Color Palette (retro theme)

| Element | Color | Contrast (on #000) |
|---------|-------|-------------------|
| Primary text | Bright green (#00ff41) | 8.2:1 |
| Headers/borders | Medium green (#00cc33) | 6.1:1 |
| Success | Bright cyan (#00ffff) | 10.5:1 |
| Warnings | Amber (#ffbf00) | 12.4:1 |
| Errors | Red (#ff4444) | 5.3:1 |
| IDs/links | Cyan (#00cccc) | 7.1:1 |
| Flavor text | Medium green (#00cc33) | 6.1:1 |
| ASCII art | Bright green (#00ff41) | 8.2:1 |

### Theme Options

| Theme | Behavior |
|-------|----------|
| `retro` | Full ASCII art, flavor text, CRT colors, rotating quotes |
| `minimal` | Colors and clean tables, no ASCII art or flavor text |
| `none` | Plain unformatted output (for piping, CI/CD, screen readers) |

### Mascot & Banner

```
    ╔══════════════════════════════════════╗
    ║                                      ║
    ║       ▄▄▄                            ║
    ║      █▀█▀█  ┏━╸╺┳╸╻                 ║
    ║      █▄█▄█  ┃   ┃ ┃                 ║
    ║      ╰█▀█╯  ┗━╸ ╹ ┗━╸              ║
    ║       ╰─╯   d a t a d o g           ║
    ║              c o n t r o l           ║
    ║                                      ║
    ║  [v0.1.0]    ◄◄ JACK IN ►►          ║
    ║                                      ║
    ╚══════════════════════════════════════╝
```

### Rotating Startup Quotes

Shown only on bare `ddctl` invocation (no subcommand), not on every command.

```
"The net is vast and infinite." — Ghost in the Shell
"I never asked for this." — Adam Jensen
"Shall we play a game?" — WOPR
"In a world of locked doors, the man with the key is king."
"I'm in." — Every 90s hacker movie
"The more things change, the more they stay the same."
```

## Bundled Examples

Embedded in the binary via Go's `embed` package.

| Example | Description |
|---------|-------------|
| `service-overview` | Golden signals (latency, throughput, errors, saturation) for any service |
| `log-analytics` | Log volume, patterns, and error rates |
| `slo-tracking` | SLO burn rate and error budget |
| `deployment` | Deploy frequency, rollback rate, change failure rate |

```bash
ddctl dashboard examples                             # list all
ddctl dashboard examples service-overview --preview   # show YAML
ddctl dashboard examples service-overview --import    # import to SQLite
```

## Testing Strategy

### Unit Tests

- DSL translation (YAML to/from Datadog API JSON) — golden file tests comparing expected output
- SQLite store operations (CRUD, versioning, pruning, migrations)
- Auth resolution (keychain → credential file → env var fallback)
- Rate limiter behavior
- Output formatters (JSON, YAML, table)
- Config parsing

### Integration Tests

- Recorded HTTP fixtures using `go-vcr` or `httpmock` — capture real Datadog API responses, replay in tests
- SQLite migration tests (apply all migrations to empty database, verify schema)
- MCP server tests using `mcp-go` test utilities

### End-to-End Tests

- Test connection profile pointing to a Datadog sandbox org (optional, for CI)
- CLI golden tests — run commands, compare stdout/stderr against expected output

### Running Tests

```bash
go test ./...                    # all tests
go test ./internal/store/...     # store tests only
go test -run TestDSL ./internal/dsl/...  # specific test
```

## Offline Behavior

Commands that only touch SQLite work without network access:

**Offline:** `list` (local), `edit`, `history`, `rollback`, `export`, `import` (without `--push`), `db prune`, `db stats`

**Requires network:** `push`, `pull`, `sync`, `list --remote`, `delete`, `create`, `diff` (against remote), `connection test`, `auth login`, `api *`, `metrics *`, `logs *`

Network-requiring commands fail immediately with a clear error when the API is unreachable, rather than hanging on a long timeout.

## v1 Scope

### In Scope

- Dashboards — full CRUD, sync, versioning, diff, rollback, DSL, examples
- Monitors — full CRUD, sync, versioning, DSL
- SLOs — full CRUD, sync, versioning, DSL
- Metrics — read (search, query timeseries, get metadata)
- Logs — read (search, analyze via SQL)
- Connections — multi-org profile management
- Auth — keychain + credential file + env var credential management
- API pass-through — generated SDK operations + raw HTTP
- MCP server — exposes all commands as tools with configurable safety levels
- Documentation — full docs, CLAUDE.md, examples
- Testing — unit, integration, golden file tests

### Deferred (v2+)

- Notebooks
- Synthetic tests
- Downtimes
- `brew install` distribution
- Dashboard templates with variable substitution
- Team-shared dashboard libraries
- SQLite encryption at rest
- `DD_API_KEY_FILE` / `DD_APP_KEY_FILE` credential file reference pattern
