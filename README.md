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

# ddctl

Manage Datadog dashboards, monitors, and SLOs from your terminal. Version-controlled, multi-org, LLM-ready.

`ddctl` gives you a push/pull workflow for Datadog resources with local versioning (SQLite), multi-org profiles, and an MCP server for LLM integration. Think `git` for your Datadog infrastructure.

---

## Install

```bash
go install github.com/futuregerald/ddctl@latest
```

Or build from source:

```bash
git clone https://github.com/futuregerald/ddctl.git
cd ddctl
make build
```

## Quickstart

```bash
# 1. Add a connection and authenticate
ddctl connection add --name production --site datadoghq.com
ddctl auth login

# 2. List remote dashboards
ddctl dashboard list --remote

# 3. Pull a dashboard to local store
ddctl dashboard pull abc-def-123

# 4. Edit locally (opens in your editor)
ddctl dashboard edit abc-def-123

# 5. See what changed
ddctl dashboard diff abc-def-123

# 6. Push changes to Datadog
ddctl dashboard push abc-def-123

# 7. View version history
ddctl dashboard history abc-def-123
```

## Features

**Resource Management**
- Push/pull dashboards, monitors, and SLOs
- Local version history with rollback
- Diff local vs. remote state
- Import/export YAML files
- Bundled example dashboards

**Multi-Org Support**
- Named connection profiles
- Per-command `--connection` override
- Credentials stored in system keychain

**Metrics & Logs**
- Query timeseries metrics
- Search available metrics
- Search and filter logs

**API Pass-through**
- Named SDK operations for any Datadog endpoint
- Raw HTTP escape hatch for new/undocumented APIs

**MCP Server**
- Full LLM integration via Model Context Protocol
- Configurable safety levels (read-only, read-write, unrestricted)
- Works with Claude Desktop, Cursor, and any MCP client

**Terminal UX**
- Retro cyberpunk aesthetic (or `--output json` for machines)
- Respects `NO_COLOR` and `TERM=dumb`
- Three themes: retro, minimal, none

## MCP Server Setup

`ddctl` includes a built-in MCP server for LLM agents. Add to your MCP config:

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

For write operations, set the safety level:

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

| Safety Level | Behavior |
|---|---|
| `read-only` (default) | List, search, query, diff, history |
| `read-write` | Reads + create, push, pull, edit |
| `unrestricted` | All operations (delete/rollback require `confirm: true`) |

Authentication uses the same credential resolution as the CLI: keychain first, then environment variables. Run `ddctl auth login` before starting the MCP server, or pass `DD_API_KEY`/`DD_APP_KEY` in the env block.

## Output Formats

```bash
ddctl dashboard list                    # default format (from config)
ddctl dashboard list -o json            # JSON
ddctl dashboard list -o yaml            # YAML
ddctl dashboard list -o table           # table
```

JSON and YAML output suppresses all theme decoration, making it safe for piping and scripting.

## Configuration

Config file: `~/.config/ddctl/config.yaml` (respects `XDG_CONFIG_HOME`)

```yaml
output: json                # json | table | yaml
default_connection: production
editor: cursor              # cursor | code | vim | nvim | nano
theme: retro                # retro | minimal | none
versions_to_keep: 50        # max versions per resource
```

## Examples

Browse bundled example dashboards:

```bash
ddctl dashboard examples                             # list all
ddctl dashboard examples service-overview --preview   # show YAML
ddctl dashboard examples service-overview --import    # import to local store
```

Example files are also available in the `examples/` directory.

## Documentation

- [Getting Started](docs/getting-started.md) -- First-run walkthrough
- [Configuration](docs/configuration.md) -- Config file, themes, editor
- [Authentication](docs/authentication.md) -- Keychain, env vars, profiles
- [Dashboards](docs/dashboards.md) -- Push/pull workflow, versioning
- [MCP Server](docs/mcp-server.md) -- MCP setup, tool reference, safety

## Commands

```
ddctl connection add|list|default|remove|test    Connection profiles
ddctl auth login|status|logout                    Credential management
ddctl dashboard list|pull|push|edit|diff|...      Dashboard management
ddctl monitor list|pull|push|edit|...             Monitor management
ddctl slo list|pull|push|edit|...                 SLO management
ddctl metrics search|query                        Metrics operations
ddctl logs search                                 Log search
ddctl api list|<operation>|raw                    API pass-through
ddctl db prune|stats                              Database maintenance
ddctl mcp serve                                   MCP server mode
ddctl version                                     Version info
```

## License

MIT
