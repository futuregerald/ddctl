# Getting Started with ddctl

This guide walks you through installing ddctl, connecting to Datadog, and managing your first dashboard.

## Prerequisites

- Go 1.21+ (for `go install`)
- A Datadog account with an API key and Application key
- macOS, Linux, or Windows

## Installation

```bash
go install github.com/futuregerald/ddctl@latest
```

Or clone and build:

```bash
git clone https://github.com/futuregerald/ddctl.git
cd ddctl
make build
./ddctl version
```

## Step 1: Add a Connection

A connection is a named Datadog organization profile. You can have multiple connections (production, staging, etc.).

```bash
ddctl connection add --name production --site datadoghq.com
```

The `--site` flag matches your Datadog site:
- `datadoghq.com` (US1, default)
- `us3.datadoghq.com` (US3)
- `us5.datadoghq.com` (US5)
- `datadoghq.eu` (EU)
- `ap1.datadoghq.com` (AP1)

## Step 2: Authenticate

Store your API key and Application key in the system keychain:

```bash
ddctl auth login
```

This prompts for your keys, validates them against the Datadog API, and stores them securely. Check the status:

```bash
ddctl auth status
```

## Step 3: List Your Dashboards

```bash
# List all dashboards in Datadog
ddctl dashboard list --remote

# Output as JSON
ddctl dashboard list --remote -o json
```

## Step 4: Pull a Dashboard

Pull a dashboard from Datadog into your local store. This creates a versioned snapshot:

```bash
ddctl dashboard pull abc-def-123
```

Now it's tracked locally:

```bash
ddctl dashboard list
```

## Step 5: Edit and Push

Open the dashboard in your editor:

```bash
ddctl dashboard edit abc-def-123
```

This opens the YAML in your configured editor. Save and close. Then review the changes:

```bash
ddctl dashboard diff abc-def-123
```

Push the changes to Datadog:

```bash
ddctl dashboard push abc-def-123
```

## Step 6: View History

Every pull, push, and edit creates a version:

```bash
ddctl dashboard history abc-def-123
```

Rollback to any previous version:

```bash
ddctl dashboard rollback abc-def-123 --to-version 2
ddctl dashboard push abc-def-123
```

## Next Steps

- [Configuration](configuration.md) -- Customize editor, theme, output format
- [Authentication](authentication.md) -- Multi-org setup, CI/CD credentials
- [Dashboards](dashboards.md) -- Full dashboard workflow reference
- [MCP Server](mcp-server.md) -- LLM integration
