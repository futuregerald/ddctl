# MCP Server

ddctl includes a built-in MCP (Model Context Protocol) server that exposes Datadog operations as tools for LLM agents. This enables Claude Desktop, Cursor, and any MCP-compatible client to manage Datadog resources programmatically.

## Quick Setup

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "npx",
      "args": ["-y", "ddctl", "mcp", "serve"]
    }
  }
}
```

**Prerequisite:** Run `ddctl auth login` first to store credentials in the keychain.

## Safety Levels

The `--safety` flag controls which operations LLM agents can perform:

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "npx",
      "args": ["-y", "ddctl", "mcp", "serve", "--safety", "read-write"]
    }
  }
}
```

| Level | Operations | Use Case |
|-------|-----------|----------|
| `read-only` (default) | list, search, query, diff, history | Safe exploration and monitoring |
| `read-write` | reads + create, push, pull, edit | Active management with guardrails |
| `unrestricted` | all operations | Full control (delete/rollback require `confirm: true`) |

The safety level is set at server startup and cannot be overridden per tool call. This ensures the human who configures the MCP server makes a deliberate choice.

## Authentication

### Keychain (recommended)

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "npx",
      "args": ["-y", "ddctl", "mcp", "serve"]
    }
  }
}
```

Works when `ddctl auth login` was previously run.

### Specific connection

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "npx",
      "args": ["-y", "ddctl", "mcp", "serve", "--connection", "production"]
    }
  }
}
```

### Environment variables

```json
{
  "mcpServers": {
    "ddctl": {
      "command": "npx",
      "args": ["-y", "ddctl", "mcp", "serve"],
      "env": {
        "DD_API_KEY": "your-api-key",
        "DD_APP_KEY": "your-app-key"
      }
    }
  }
}
```

**Warning:** Secrets in MCP config files may be committed to version control.

## Tool Reference

### Dashboard Tools

| Tool | Safety | Description |
|------|--------|-------------|
| `dashboard_list` | read-only | List locally tracked dashboards |
| `dashboard_list_remote` | read-only | List all remote dashboards |
| `dashboard_diff` | read-only | Diff local vs remote |
| `dashboard_history` | read-only | Version history |
| `dashboard_pull` | read-write | Pull remote to local |
| `dashboard_push` | read-write | Push local to remote (requires `confirm: true`) |
| `dashboard_create` | read-write | Create new dashboard from YAML |
| `dashboard_edit` | read-write | Update local content |
| `dashboard_rollback` | unrestricted | Rollback to previous version (requires `confirm: true`) |
| `dashboard_delete` | unrestricted | Delete from Datadog (requires `confirm: true`) |

### Monitor Tools

| Tool | Safety | Description |
|------|--------|-------------|
| `monitor_list` | read-only | List locally tracked monitors |
| `monitor_list_remote` | read-only | List all remote monitors |
| `monitor_pull` | read-write | Pull remote to local |
| `monitor_push` | read-write | Push local to remote |
| `monitor_create` | read-write | Create new monitor from YAML |
| `monitor_delete` | unrestricted | Delete from Datadog |

### SLO Tools

| Tool | Safety | Description |
|------|--------|-------------|
| `slo_list` | read-only | List locally tracked SLOs |
| `slo_list_remote` | read-only | List all remote SLOs |
| `slo_pull` | read-write | Pull remote to local |
| `slo_push` | read-write | Push local to remote |
| `slo_create` | read-write | Create new SLO from YAML |
| `slo_delete` | unrestricted | Delete from Datadog |

### Metrics & Logs Tools

| Tool | Safety | Description |
|------|--------|-------------|
| `metrics_query` | read-only | Query timeseries metrics |
| `metrics_search` | read-only | Search available metrics |
| `logs_search` | read-only | Search logs |

### API Pass-through

| Tool | Safety | Description |
|------|--------|-------------|
| `api_call` | read-write | Arbitrary API call (method, path, body) |

## Behavior Differences from CLI

- **Output is always JSON** -- no theme decoration
- **`dashboard_edit` accepts YAML as a parameter** instead of opening an editor
- **Destructive operations require `confirm: true`** -- the LLM must explicitly opt in
- **Tools include rich descriptions** so LLMs understand capabilities and constraints

## Confirm Parameter

At `read-write` and `unrestricted` safety levels, operations that modify remote state require a `confirm: true` parameter:

- `dashboard_push`, `monitor_push`, `slo_push`
- `dashboard_delete`, `monitor_delete`, `slo_delete`
- `dashboard_rollback`

This is an intentional friction point. The LLM must make a deliberate choice rather than accidentally triggering a destructive action.

Example tool call:
```json
{
  "name": "dashboard_push",
  "arguments": {
    "id": "abc-def-123",
    "confirm": true
  }
}
```

## Transport

The MCP server uses **stdio transport**. It reads JSON-RPC messages from stdin and writes responses to stdout. Diagnostic output goes to stderr.
