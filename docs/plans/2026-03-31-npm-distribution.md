# npm Distribution via Platform Packages — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Distribute ddctl as an npm package so MCP clients can use `npx ddctl mcp serve` without requiring Go or PATH configuration.

**Architecture:** A root `ddctl` npm package declares 6 platform-specific `@ddctl/<os>-<arch>` packages as `optionalDependencies`. Each platform package contains only the Go binary. The root package has a thin JS bin script that resolves the correct platform binary via `require.resolve` and execs it. A GitHub Actions workflow triggers on release publication, downloads the GoReleaser assets, injects them into the platform packages, and publishes all 7 packages to npm.

**Tech Stack:** npm (registry + CLI), Node.js (bin script), GitHub Actions (CI publishing)

---

## Platform Mapping

GoReleaser uses Go's GOOS/GOARCH naming. npm uses Node.js `process.platform`/`process.arch` naming. The mapping:

| npm package | Node platform | Node arch | GoReleaser asset |
|---|---|---|---|
| `@ddctl/darwin-arm64` | `darwin` | `arm64` | `ddctl_VERSION_darwin_arm64.tar.gz` |
| `@ddctl/darwin-x64` | `darwin` | `x64` | `ddctl_VERSION_darwin_amd64.tar.gz` |
| `@ddctl/linux-arm64` | `linux` | `arm64` | `ddctl_VERSION_linux_arm64.tar.gz` |
| `@ddctl/linux-x64` | `linux` | `x64` | `ddctl_VERSION_linux_amd64.tar.gz` |
| `@ddctl/win32-arm64` | `win32` | `arm64` | `ddctl_VERSION_windows_arm64.zip` |
| `@ddctl/win32-x64` | `win32` | `x64` | `ddctl_VERSION_windows_amd64.zip` |

Note: Go uses `amd64`, Node uses `x64`. Go uses `windows`, Node uses `win32`.

---

### Task 1: Create platform package scaffolds

**Files:**

- Create: `npm/@ddctl/darwin-arm64/package.json`
- Create: `npm/@ddctl/darwin-x64/package.json`
- Create: `npm/@ddctl/linux-arm64/package.json`
- Create: `npm/@ddctl/linux-x64/package.json`
- Create: `npm/@ddctl/win32-arm64/package.json`
- Create: `npm/@ddctl/win32-x64/package.json`

Each platform package has the same structure, differing only in name, description, and `os`/`cpu` fields. The `os` and `cpu` fields tell npm which platform this package is for — npm will skip installing packages that don't match.

**Step 1: Create all 6 platform package.json files**

Each follows this pattern (example for darwin-arm64):

```json
{
  "name": "@ddctl/darwin-arm64",
  "version": "0.0.0",
  "description": "ddctl binary for macOS arm64",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/futuregerald/ddctl"
  },
  "os": ["darwin"],
  "cpu": ["arm64"],
  "files": ["bin/ddctl"]
}
```

For Windows packages, the binary name is `bin/ddctl.exe` and the `files` field reflects that.

**Important:** Platform packages do NOT declare a `"bin"` field. Only the root `ddctl` package registers the bin entry. This avoids conflicting bin registrations (follows the esbuild/SWC pattern).

Platform package.json field values:

| Package | `os` | `cpu` | binary |
|---|---|---|---|
| `@ddctl/darwin-arm64` | `["darwin"]` | `["arm64"]` | `bin/ddctl` |
| `@ddctl/darwin-x64` | `["darwin"]` | `["x64"]` | `bin/ddctl` |
| `@ddctl/linux-arm64` | `["linux"]` | `["arm64"]` | `bin/ddctl` |
| `@ddctl/linux-x64` | `["linux"]` | `["x64"]` | `bin/ddctl` |
| `@ddctl/win32-arm64` | `["win32"]` | `["arm64"]` | `bin/ddctl.exe` |
| `@ddctl/win32-x64` | `["win32"]` | `["x64"]` | `bin/ddctl.exe` |

**Step 2: Create placeholder bin directories**

Each platform package needs a `bin/` directory. Create `.gitkeep` files since the actual binaries are injected by CI:

```
npm/@ddctl/darwin-arm64/bin/.gitkeep
npm/@ddctl/darwin-x64/bin/.gitkeep
npm/@ddctl/linux-arm64/bin/.gitkeep
npm/@ddctl/linux-x64/bin/.gitkeep
npm/@ddctl/win32-arm64/bin/.gitkeep
npm/@ddctl/win32-x64/bin/.gitkeep
```

**Step 3: Commit**

```bash
git add npm/@ddctl/
git commit -m "feat(npm): add platform package scaffolds for 6 targets"
```

---

### Task 2: Create root ddctl npm package

**Files:**

- Create: `npm/ddctl/package.json`
- Create: `npm/ddctl/bin/ddctl.js`

**Step 1: Create `npm/ddctl/package.json`**

```json
{
  "name": "ddctl",
  "version": "0.0.0",
  "description": "Manage Datadog dashboards, monitors, and SLOs from your terminal. MCP server included.",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/futuregerald/ddctl"
  },
  "bin": {
    "ddctl": "bin/ddctl.js"
  },
  "files": ["bin/ddctl.js"],
  "optionalDependencies": {
    "@ddctl/darwin-arm64": "0.0.0",
    "@ddctl/darwin-x64": "0.0.0",
    "@ddctl/linux-arm64": "0.0.0",
    "@ddctl/linux-x64": "0.0.0",
    "@ddctl/win32-arm64": "0.0.0",
    "@ddctl/win32-x64": "0.0.0"
  }
}
```

The `optionalDependencies` versions are `0.0.0` placeholders — CI stamps the real version from the git tag before publishing.

**Step 2: Create `npm/ddctl/bin/ddctl.js`**

This is the bin entrypoint. It resolves the platform binary and execs it:

```javascript
#!/usr/bin/env node

const { execFileSync } = require("child_process");
const path = require("path");

const PLATFORMS = {
  "darwin-arm64": "@ddctl/darwin-arm64",
  "darwin-x64": "@ddctl/darwin-x64",
  "linux-arm64": "@ddctl/linux-arm64",
  "linux-x64": "@ddctl/linux-x64",
  "win32-arm64": "@ddctl/win32-arm64",
  "win32-x64": "@ddctl/win32-x64",
};

const key = `${process.platform}-${process.arch}`;
const pkg = PLATFORMS[key];

if (!pkg) {
  console.error(`ddctl: unsupported platform ${key}`);
  process.exit(1);
}

const bin = process.platform === "win32" ? "ddctl.exe" : "ddctl";

let binPath;
try {
  binPath = path.join(path.dirname(require.resolve(`${pkg}/package.json`)), "bin", bin);
} catch {
  console.error(
    `ddctl: could not find package ${pkg}.\n` +
    `Make sure your package manager installed the platform-specific dependency.\n` +
    `Try: npm install ddctl`
  );
  process.exit(1);
}

try {
  execFileSync(binPath, process.argv.slice(2), { stdio: "inherit" });
} catch (e) {
  if (e.status !== undefined) {
    process.exit(e.status);
  }
  throw e;
}
```

**Step 3: Commit**

```bash
git add npm/ddctl/
git commit -m "feat(npm): add root ddctl package with platform binary resolver"
```

---

### Task 3: Create npm-publish GitHub Actions workflow

**Files:**

- Create: `.github/workflows/npm-publish.yaml`

**Step 1: Create the workflow file**

```yaml
name: Publish npm packages

on:
  release:
    types: [published]

permissions:
  contents: read

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          registry-url: 'https://registry.npmjs.org'

      - name: Get version from tag
        id: version
        run: echo "VERSION=${GITHUB_REF_NAME#v}" >> "$GITHUB_OUTPUT"

      - name: Download release assets
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          mkdir -p dist
          gh release download "$GITHUB_REF_NAME" --dir dist --repo "$GITHUB_REPOSITORY"

      - name: Verify checksums
        run: cd dist && sha256sum -c checksums.txt

      - name: Extract and place binaries
        run: |
          VERSION="${{ steps.version.outputs.VERSION }}"

          # Map: npm-dir goreleaser-archive binary-name
          declare -A ARCHIVES=(
            ["darwin-arm64"]="ddctl_${VERSION}_darwin_arm64.tar.gz"
            ["darwin-x64"]="ddctl_${VERSION}_darwin_amd64.tar.gz"
            ["linux-arm64"]="ddctl_${VERSION}_linux_arm64.tar.gz"
            ["linux-x64"]="ddctl_${VERSION}_linux_amd64.tar.gz"
            ["win32-arm64"]="ddctl_${VERSION}_windows_arm64.zip"
            ["win32-x64"]="ddctl_${VERSION}_windows_amd64.zip"
          )

          for platform in "${!ARCHIVES[@]}"; do
            archive="${ARCHIVES[$platform]}"
            dest="npm/@ddctl/${platform}/bin"
            mkdir -p "$dest"

            # GoReleaser produces flat archives (no wrapping directory)
            mkdir -p /tmp/extract-${platform}

            if [[ "$archive" == *.zip ]]; then
              unzip -o "dist/$archive" -d /tmp/extract-${platform}
              cp /tmp/extract-${platform}/ddctl.exe "$dest/"
            else
              tar -xzf "dist/$archive" -C /tmp/extract-${platform}
              cp /tmp/extract-${platform}/ddctl "$dest/"
            fi

            chmod +x "$dest"/*
          done

      - name: Stamp versions in all package.json files
        run: |
          VERSION="${{ steps.version.outputs.VERSION }}"
          for pkg in npm/@ddctl/*/package.json npm/ddctl/package.json; do
            sed -i "s/\"version\": \"0.0.0\"/\"version\": \"${VERSION}\"/" "$pkg"
            # Also update optionalDependencies versions in root package
            sed -i "s/\"@ddctl\/\([^\"]*\)\": \"0.0.0\"/\"@ddctl\/\1\": \"${VERSION}\"/g" "$pkg"
          done

      - name: Publish platform packages
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        run: |
          for dir in npm/@ddctl/*/; do
            echo "Publishing $(basename "$dir")..."
            npm publish "$dir" --access public
          done

      - name: Publish root package
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        run: npm publish npm/ddctl --access public
```

Key details:
- `sha256sum -c checksums.txt` validates all downloaded archives against GoReleaser's checksums
- Platform packages are published first (root depends on them)
- `--access public` is required for scoped packages
- GoReleaser archives are flat (no wrapping directory) — binary and README.md at root
- The `sed` commands stamp the version from the git tag into all package.json files, including the `optionalDependencies` block in the root package

**Publish failure recovery:** If the workflow fails partway through publishing (e.g., network error on package 4 of 6), the npm registry will be in an inconsistent state. To recover: `npm unpublish @ddctl/<pkg>@<version>` for each partially published package (within 72 hours), then re-trigger the workflow by re-running it from the Actions tab.

**Step 2: Commit**

```bash
git add .github/workflows/npm-publish.yaml
git commit -m "ci: add npm publish workflow triggered by GitHub releases"
```

---

### Task 4: Update README and MCP docs

**Files:**

- Modify: `README.md:25-37` (Install section)
- Modify: `README.md:97-123` (MCP Server Setup section)
- Modify: `docs/mcp-server.md:1-16` (Quick Setup section)

**Step 1: Update README Install section**

Replace lines 25-37 with:

```markdown
## Install

```bash
# npm (recommended for MCP server usage)
npx ddctl version

# Go
go install github.com/futuregerald/ddctl@latest

# Download binary
# See https://github.com/futuregerald/ddctl/releases

# Build from source
git clone https://github.com/futuregerald/ddctl.git
cd ddctl
make build
```
```

**Step 2: Update README MCP Server Setup section**

Replace the MCP config JSON at lines 99-110 with:

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

And the read-write example at lines 112-121 with:

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

**Step 3: Update `docs/mcp-server.md` Quick Setup**

Replace the Quick Setup JSON block at lines 7-16 with the same `npx` pattern:

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

Also update the Authentication section examples (lines 45-88) to use `npx`:

- Keychain example: `"command": "npx"`, `"args": ["-y", "ddctl", "mcp", "serve"]`
- Connection example: `"command": "npx"`, `"args": ["-y", "ddctl", "mcp", "serve", "--connection", "production"]`
- Env vars example: `"command": "npx"`, `"args": ["-y", "ddctl", "mcp", "serve"]`

**Step 4: Commit**

```bash
git add README.md docs/mcp-server.md
git commit -m "docs: update MCP setup to use npx for installation"
```

---

### Task 5: Add .gitignore for platform binaries

**Files:**

- Create or modify: `npm/.gitignore`

The platform package `bin/` directories will have `.gitkeep` files checked in, but actual binaries should never be committed (they're injected by CI).

**Step 1: Create `npm/.gitignore`**

```
# Platform binaries are injected by CI, never committed
@ddctl/*/bin/ddctl
@ddctl/*/bin/ddctl.exe
```

**Step 2: Commit**

```bash
git add npm/.gitignore
git commit -m "chore: add gitignore for npm platform binaries"
```

---

## Pre-Publish Checklist (manual, one-time)

Before the first release that triggers npm publishing:

1. **Create npm org:** Go to https://www.npmjs.com and create the `@ddctl` organization
2. **Create npm automation token:** In npm account settings, generate an automation token with publish access
3. **Add GitHub secret:** In `futuregerald/ddctl` repo settings > Secrets > Actions, add `NPM_TOKEN` with the automation token value
4. **Verify:** Run `npm whoami --registry https://registry.npmjs.org` to confirm auth works

---

## Verification

After the first release with npm publishing:

```bash
# Should download and run the binary
npx ddctl version

# Should show the correct version
npx ddctl version
# Expected: ddctl 0.X.0 (commit: ..., built: ...)

# MCP config should work
# Add to Claude Desktop config:
# { "mcpServers": { "ddctl": { "command": "npx", "args": ["-y", "ddctl", "mcp", "serve"] } } }
```
