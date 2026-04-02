# ddctl

Manage Datadog dashboards, monitors, and SLOs from your terminal. Version-controlled, multi-org, LLM-ready.

## Install

```bash
npm install ddctl
```

Or run directly:

```bash
npx ddctl version
```

## MCP Server Setup

Add to your MCP client config (Claude Desktop, Cursor, etc.):

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

For write operations:

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

## Prerequisites

Before using the MCP server, authenticate with Datadog:

```bash
npx ddctl connection add --name production --site datadoghq.com
npx ddctl auth login
```

## Features

- Push/pull dashboards, monitors, and SLOs with local version history
- Multi-org support with named connection profiles
- Metrics and logs querying
- API pass-through for any Datadog endpoint
- MCP server for LLM integration
- Configurable safety levels (read-only, read-write, unrestricted)

## Documentation

Full docs: https://github.com/futuregerald/ddctl

## License

MIT
