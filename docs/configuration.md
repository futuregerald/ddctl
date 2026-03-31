# Configuration

ddctl configuration lives at `~/.config/ddctl/config.yaml`. The path respects `XDG_CONFIG_HOME`.

## Config File

```yaml
output: json                # json | table | yaml
default_connection: production
editor: cursor              # cursor | code | vim | nvim | nano
theme: retro                # retro | minimal | none
versions_to_keep: 50        # max versions per resource (0 = unlimited)
```

## Options

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `output` | `json`, `table`, `yaml` | `json` | Default output format |
| `default_connection` | string | (first added) | Connection profile to use when `--connection` not specified |
| `editor` | `cursor`, `code`, `vim`, `nvim`, `nano`, or any command | `$EDITOR` | Editor for `edit` commands |
| `theme` | `retro`, `minimal`, `none` | `retro` | Terminal decoration level |
| `versions_to_keep` | integer | `50` | Max versions per resource before auto-pruning |

## Editor Support

| Editor | How it's invoked |
|--------|------------------|
| `vim` / `nvim` / `nano` | Opens in terminal, blocks naturally |
| `cursor` | `cursor --wait <file>` |
| `code` | `code --wait <file>` |
| Other | `<editor> <file>` |

If `editor` is not set in config, ddctl falls back to the `$EDITOR` environment variable.

## Themes

| Theme | Behavior |
|-------|----------|
| `retro` | Full ASCII art, flavor text, CRT-green colors, rotating quotes |
| `minimal` | Colors and clean tables, no ASCII art or flavor text |
| `none` | Plain unformatted output (for piping, CI/CD, screen readers) |

Theme decoration is automatically suppressed when using `--output json` or `--output yaml`, regardless of the theme setting.

### Accessibility

- Respects `NO_COLOR` environment variable -- disables all colors
- Respects `TERM=dumb` -- disables colors and ASCII art
- All text meets WCAG AA contrast ratios against dark terminal backgrounds

## Global Flags

These flags override config values on any command:

| Flag | Description |
|------|-------------|
| `-c, --connection` | Connection profile to use |
| `-o, --output` | Output format (json, table, yaml) |
| `-y, --yes` | Skip confirmation prompts |
| `--debug` | Enable debug output to stderr |

## Database Location

SQLite database: `~/.config/ddctl/ddctl.db`

Manage the database with:

```bash
ddctl db stats     # Show database size and resource counts
ddctl db prune     # Remove old versions beyond the retention limit
```
