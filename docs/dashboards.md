# Dashboards

ddctl provides a full push/pull workflow for Datadog dashboards with local versioning.

## Workflow

```
Remote (Datadog)  <--pull-->  Local (SQLite)  <--edit-->  Your Editor
                  <--push-->
```

1. **Pull** a dashboard from Datadog to create a local copy
2. **Edit** the local copy in your editor
3. **Diff** to see changes between local and remote
4. **Push** the local copy back to Datadog
5. **History** and **rollback** if something goes wrong

## Commands

### List

```bash
# List locally tracked dashboards
ddctl dashboard list

# List all remote dashboards from Datadog
ddctl dashboard list --remote
```

### Pull

Pull a dashboard from Datadog into the local store:

```bash
ddctl dashboard pull abc-def-123
```

This fetches the current remote state and saves it as a new local version.

### Edit

Open a dashboard in your editor:

```bash
ddctl dashboard edit abc-def-123
```

The edit workflow:
1. Reads the latest local version from SQLite
2. Writes to a temp file
3. Opens your configured editor
4. On save, validates YAML and stores as a new local version
5. Reports success and suggests `dashboard push` to apply

### Diff

Compare local state against remote:

```bash
ddctl dashboard diff abc-def-123
```

### Push

Push local changes to Datadog:

```bash
ddctl dashboard push abc-def-123
```

Push safety checks:
1. Fetches current remote state
2. Shows who last modified and when
3. Asks for confirmation (`--yes` to skip)
4. Applies the local state

### Create

Create a new dashboard from a YAML file:

```bash
ddctl dashboard create -f dashboard.yaml
```

Or interactively with flags.

### Import / Export

```bash
# Import a YAML file into local store
ddctl dashboard import -f dashboard.yaml

# Import and push to Datadog
ddctl dashboard import -f dashboard.yaml --push

# Export from local store to file
ddctl dashboard export abc-def-123 -o dashboard.yaml
```

### History

View all local versions:

```bash
ddctl dashboard history abc-def-123
```

### Rollback

Restore a previous version:

```bash
ddctl dashboard rollback abc-def-123 --to-version 3
```

This creates a new version with the content from version 3. Push separately to apply to Datadog.

### Delete

Delete from Datadog (with confirmation):

```bash
ddctl dashboard delete abc-def-123
```

### Sync

Pull all tracked dashboards from remote:

```bash
ddctl dashboard sync
```

## YAML Format

Dashboards use the Datadog API format in YAML. This is a 1:1 mapping of the JSON API schema.

```yaml
title: "My Dashboard"
description: "Service overview"
layout_type: ordered
widgets:
  - definition:
      title: "Request Latency"
      type: timeseries
      requests:
        - q: "avg:trace.http.request.duration{service:my-service}"
          display_type: line
    layout:
      x: 0
      y: 0
      width: 6
      height: 3
tags:
  - "managed_by:ddctl"
```

## Bundled Examples

```bash
# List bundled examples
ddctl dashboard examples

# Preview an example
ddctl dashboard examples service-overview --preview

# Import an example to local store
ddctl dashboard examples service-overview --import
```

Available examples:
- `service-overview` -- Golden signals (latency, throughput, errors, saturation)
- `log-analytics` -- Log volume, error rates, top sources

## Version Retention

Versions are auto-pruned per resource based on `versions_to_keep` config (default: 50). Manage manually:

```bash
ddctl db prune     # Remove old versions
ddctl db stats     # Show version counts
```
