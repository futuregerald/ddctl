# Authentication

ddctl resolves credentials in priority order. No credentials are ever stored in config files or the SQLite database.

## Resolution Order

1. **Connection-specific environment variables**: `DD_<CONNECTION>_API_KEY` / `DD_<CONNECTION>_APP_KEY`
2. **Generic environment variables**: `DD_API_KEY` / `DD_APP_KEY`
3. **System keychain**: stored per connection by `ddctl auth login`
4. **Error with setup instructions**

## Auth Commands

### Login (Interactive)

```bash
# Store keys in system keychain (recommended for workstations)
ddctl auth login

# Store keys in credential file (for headless environments)
ddctl auth login --backend file
```

`auth login` prompts for your API key and Application key, validates them with a test API call, and stores them securely.

### Check Status

```bash
ddctl auth status
```

Shows where credentials are being sourced from for the current connection. Warns if using environment variables (visible in process listings).

### Logout

```bash
ddctl auth logout
```

Removes stored credentials from the keychain or credential file.

## Multi-Org Setup

Use named connections for multiple Datadog organizations:

```bash
# Add connections
ddctl connection add --name production --site datadoghq.com
ddctl connection add --name staging --site datadoghq.com
ddctl connection add --name eu-org --site datadoghq.eu

# Authenticate each
ddctl auth login --connection production
ddctl auth login --connection staging

# Set default
ddctl connection default production

# Override per command
ddctl dashboard list --remote --connection staging
```

## Environment Variables

For CI/CD and headless environments where the keychain is unavailable:

```bash
# Generic (used for all connections)
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"

# Connection-specific (takes priority)
export DD_PRODUCTION_API_KEY="prod-api-key"
export DD_PRODUCTION_APP_KEY="prod-app-key"
export DD_STAGING_API_KEY="staging-api-key"
export DD_STAGING_APP_KEY="staging-app-key"
```

**Security note:** Environment variables are visible in process listings (`ps auxe`). Use the keychain for developer workstations.

## Credential Backends

| Backend | Best For | Command |
|---------|----------|---------|
| System keychain | Developer workstations | `ddctl auth login` (default) |
| Environment variables | CI/CD, MCP server | `DD_API_KEY` / `DD_APP_KEY` |
| Credential file | Headless servers, Docker | `ddctl auth login --backend file` |

The credential file is stored at `~/.config/ddctl/credentials` with mode `0600`.

## MCP Server Authentication

The MCP server uses the same credential resolution. For keychain auth:

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

For environment variable auth:

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

**Warning:** Secrets in MCP config files may be committed to version control. Prefer keychain auth when possible.
