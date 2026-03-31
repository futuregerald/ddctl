# ddctl Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI (`ddctl`) for managing Datadog dashboards, monitors, and SLOs with built-in versioning, multi-org support, and MCP server mode.

**Architecture:** Cobra CLI with SQLite for local state (`modernc.org/sqlite`), Datadog Go SDK for API access, `go-keyring` for credential storage, `mcp-go` for MCP server mode. Resources follow a push/pull/edit/history/rollback pattern with SQLite as local state and Datadog as remote. Raw API YAML is the default format; simplified DSL is opt-in.

**Tech Stack:** Go 1.25, Cobra, SQLite (modernc.org/sqlite), Datadog API Client Go v2, go-keyring, mcp-go, lipgloss, gopkg.in/yaml.v3

**Design Doc:** `docs/plans/2026-03-30-ddctl-design.md`

---

## Phase 1: Project Scaffolding & Core Infrastructure

### Task 1: Initialize Go Module and Dependencies

**Files:**

- Create: `go.mod`
- Create: `go.sum`
- Create: `main.go`

**Step 1: Initialize the Go module**

Run:
```bash
cd ~/Documents/dev/ddctl
go mod init github.com/futuregerald/ddctl
```

**Step 2: Add core dependencies**

Run:
```bash
go get github.com/spf13/cobra@latest
go get modernc.org/sqlite@latest
go get github.com/zalando/go-keyring@latest
go get github.com/charmbracelet/lipgloss@latest
go get gopkg.in/yaml.v3@latest
go get github.com/DataDog/datadog-api-client-go/v2@latest
go get github.com/mark3labs/mcp-go@latest
```

**Step 3: Create main.go entry point**

```go
// main.go
package main

import (
	"os"

	"github.com/futuregerald/ddctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Step 4: Verify it compiles**

Run: `go build ./...`
Expected: No errors (won't run yet — cmd package doesn't exist)

**Step 5: Commit**

```bash
git add go.mod go.sum main.go
git commit -m "feat: initialize Go module with core dependencies"
```

---

### Task 2: Root Cobra Command with Global Flags

**Files:**

- Create: `cmd/root.go`
- Create: `cmd/version.go`

**Step 1: Create the root command**

```go
// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagConnection string
	flagOutput     string
	flagYes        bool
	flagDebug      bool
)

var rootCmd = &cobra.Command{
	Use:   "ddctl",
	Short: "Datadog control CLI",
	Long:  `ddctl — manage Datadog dashboards, monitors & SLOs from your terminal.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Will initialize config, client, etc. in later tasks
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagConnection, "connection", "c", "", "Connection profile to use")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format (json|table|yaml)")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug output to stderr")
}
```

**Step 2: Create the version command**

```go
// cmd/version.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print ddctl version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ddctl %s (commit: %s, built: %s)\n", Version, CommitSHA, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
```

**Step 3: Verify it builds and runs**

Run: `cd ~/Documents/dev/ddctl && go build -o ddctl . && ./ddctl version`
Expected: `ddctl dev (commit: unknown, built: unknown)`

Run: `./ddctl --help`
Expected: Help text showing global flags

**Step 4: Commit**

```bash
git add cmd/root.go cmd/version.go
git commit -m "feat: add root Cobra command with global flags and version"
```

---

### Task 3: Config System

**Files:**

- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the config test**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Output != "json" {
		t.Errorf("expected default output 'json', got %q", cfg.Output)
	}
	if cfg.Theme != "retro" {
		t.Errorf("expected default theme 'retro', got %q", cfg.Theme)
	}
	if cfg.VersionsToKeep != 50 {
		t.Errorf("expected default versions_to_keep 50, got %d", cfg.VersionsToKeep)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return defaults when file not found
	if cfg.Output != "json" {
		t.Errorf("expected default output, got %q", cfg.Output)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("output: table\ntheme: minimal\neditor: vim\nversions_to_keep: 10\n")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Output != "table" {
		t.Errorf("expected output 'table', got %q", cfg.Output)
	}
	if cfg.Theme != "minimal" {
		t.Errorf("expected theme 'minimal', got %q", cfg.Theme)
	}
	if cfg.Editor != "vim" {
		t.Errorf("expected editor 'vim', got %q", cfg.Editor)
	}
	if cfg.VersionsToKeep != 10 {
		t.Errorf("expected versions_to_keep 10, got %d", cfg.VersionsToKeep)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/config/...`
Expected: FAIL — package doesn't exist yet

**Step 3: Implement config**

```go
// internal/config/config.go
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Output            string `yaml:"output"`
	DefaultConnection string `yaml:"default_connection"`
	Editor            string `yaml:"editor"`
	Theme             string `yaml:"theme"`
	VersionsToKeep    int    `yaml:"versions_to_keep"`
}

func DefaultConfig() *Config {
	return &Config{
		Output:         "json",
		Theme:          "retro",
		VersionsToKeep: 50,
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "ddctl")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ddctl")
}

func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

func DBPath() string {
	return filepath.Join(Dir(), "ddctl.db")
}
```

**Step 4: Run tests**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/config/... -v`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config system with YAML parsing and defaults"
```

---

### Task 4: SQLite Store — Schema and Migrations

**Files:**

- Create: `internal/store/store.go`
- Create: `internal/store/migrations.go`
- Create: `internal/store/store_test.go`

**Step 1: Write the store tests**

```go
// internal/store/store_test.go
package store

import (
	"path/filepath"
	"testing"
)

func TestNewStore_CreatesDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Verify tables exist by querying them
	var count int
	err = s.db.QueryRow("SELECT count(*) FROM connections").Scan(&count)
	if err != nil {
		t.Fatalf("connections table not created: %v", err)
	}

	err = s.db.QueryRow("SELECT count(*) FROM resources").Scan(&count)
	if err != nil {
		t.Fatalf("resources table not created: %v", err)
	}

	err = s.db.QueryRow("SELECT count(*) FROM resource_versions").Scan(&count)
	if err != nil {
		t.Fatalf("resource_versions table not created: %v", err)
	}
}

func TestNewStore_WALMode(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	var journalMode string
	err = s.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatal(err)
	}
	if journalMode != "wal" {
		t.Errorf("expected WAL journal mode, got %q", journalMode)
	}
}

func TestNewStore_ForeignKeysEnabled(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	var fk int
	err = s.db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Error("expected foreign keys to be enabled")
	}
}

func TestNewStore_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	s.Close()

	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}
}
```

**Step 2: Run tests to verify failure**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/store/... -v`
Expected: FAIL

**Step 3: Implement store with migrations**

```go
// internal/store/migrations.go
package store

var migrations = []string{
	// Migration 1: Initial schema
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     INTEGER PRIMARY KEY,
		applied_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS connections (
		name        TEXT PRIMARY KEY,
		site        TEXT NOT NULL,
		is_default  INTEGER DEFAULT 0,
		created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS resources (
		resource_id         TEXT NOT NULL,
		resource_type       TEXT NOT NULL,
		connection          TEXT NOT NULL,
		title               TEXT,
		remote_modified_at  TIMESTAMP,
		remote_modified_by  TEXT,
		remote_etag         TEXT,
		last_synced_at      TIMESTAMP,
		status              TEXT DEFAULT 'active',
		PRIMARY KEY (resource_id, resource_type, connection),
		FOREIGN KEY (connection) REFERENCES connections(name) ON UPDATE CASCADE ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS resource_versions (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		resource_id     TEXT NOT NULL,
		resource_type   TEXT NOT NULL,
		connection       TEXT NOT NULL,
		version         INTEGER NOT NULL,
		content         TEXT NOT NULL,
		remote_snapshot TEXT,
		applied_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		applied_by      TEXT,
		message         TEXT,
		UNIQUE(resource_id, resource_type, connection, version),
		FOREIGN KEY (resource_id, resource_type, connection)
			REFERENCES resources(resource_id, resource_type, connection)
			ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_versions_lookup
		ON resource_versions(resource_id, resource_type, connection, version DESC);
	CREATE INDEX IF NOT EXISTS idx_versions_applied_at
		ON resource_versions(applied_at);
	CREATE INDEX IF NOT EXISTS idx_resources_connection
		ON resources(connection);`,
}
```

```go
// internal/store/store.go
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Set file permissions to 0600
	if err := os.Chmod(dbPath, 0600); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting db permissions: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	// Ensure schema_migrations table exists
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	for i, m := range migrations {
		version := i + 1
		var exists int
		err := s.db.QueryRow("SELECT count(*) FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", version, err)
		}
		if _, err := s.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("recording migration %d: %w", version, err)
		}
	}

	return nil
}
```

**Step 4: Run tests**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/store/... -v`
Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add internal/store/
git commit -m "feat: add SQLite store with migrations, WAL mode, and foreign keys"
```

---

### Task 5: Connection CRUD Operations

**Files:**

- Create: `internal/store/connections.go`
- Create: `internal/store/connections_test.go`

**Step 1: Write connection tests**

```go
// internal/store/connections_test.go
package store

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestAddConnection(t *testing.T) {
	s := newTestStore(t)

	err := s.AddConnection("prod", "datadoghq.com")
	if err != nil {
		t.Fatalf("failed to add connection: %v", err)
	}

	conns, err := s.ListConnections()
	if err != nil {
		t.Fatal(err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if conns[0].Name != "prod" {
		t.Errorf("expected name 'prod', got %q", conns[0].Name)
	}
	if conns[0].Site != "datadoghq.com" {
		t.Errorf("expected site 'datadoghq.com', got %q", conns[0].Site)
	}
}

func TestAddConnection_FirstIsDefault(t *testing.T) {
	s := newTestStore(t)

	_ = s.AddConnection("prod", "datadoghq.com")

	conn, err := s.GetDefaultConnection()
	if err != nil {
		t.Fatal(err)
	}
	if conn.Name != "prod" {
		t.Errorf("expected first connection to be default, got %q", conn.Name)
	}
}

func TestSetDefaultConnection(t *testing.T) {
	s := newTestStore(t)

	_ = s.AddConnection("prod", "datadoghq.com")
	_ = s.AddConnection("staging", "datadoghq.com")

	err := s.SetDefaultConnection("staging")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := s.GetDefaultConnection()
	if err != nil {
		t.Fatal(err)
	}
	if conn.Name != "staging" {
		t.Errorf("expected default 'staging', got %q", conn.Name)
	}
}

func TestRemoveConnection(t *testing.T) {
	s := newTestStore(t)

	_ = s.AddConnection("prod", "datadoghq.com")
	err := s.RemoveConnection("prod")
	if err != nil {
		t.Fatal(err)
	}

	conns, err := s.ListConnections()
	if err != nil {
		t.Fatal(err)
	}
	if len(conns) != 0 {
		t.Errorf("expected 0 connections, got %d", len(conns))
	}
}

func TestGetConnection_NotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.GetConnection("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
}
```

**Step 2: Run tests to verify failure**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/store/... -v -run TestAdd`
Expected: FAIL

**Step 3: Implement connection operations**

```go
// internal/store/connections.go
package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Connection struct {
	Name      string
	Site      string
	IsDefault bool
	CreatedAt time.Time
}

func (s *Store) AddConnection(name, site string) error {
	// Check if this is the first connection
	var count int
	if err := s.db.QueryRow("SELECT count(*) FROM connections").Scan(&count); err != nil {
		return err
	}
	isDefault := count == 0

	_, err := s.db.Exec(
		"INSERT INTO connections (name, site, is_default) VALUES (?, ?, ?)",
		name, site, isDefault,
	)
	if err != nil {
		return fmt.Errorf("adding connection %q: %w", name, err)
	}
	return nil
}

func (s *Store) GetConnection(name string) (*Connection, error) {
	conn := &Connection{}
	err := s.db.QueryRow(
		"SELECT name, site, is_default, created_at FROM connections WHERE name = ?",
		name,
	).Scan(&conn.Name, &conn.Site, &conn.IsDefault, &conn.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("connection %q not found", name)
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Store) GetDefaultConnection() (*Connection, error) {
	conn := &Connection{}
	err := s.db.QueryRow(
		"SELECT name, site, is_default, created_at FROM connections WHERE is_default = 1",
	).Scan(&conn.Name, &conn.Site, &conn.IsDefault, &conn.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no default connection configured")
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Store) ListConnections() ([]Connection, error) {
	rows, err := s.db.Query("SELECT name, site, is_default, created_at FROM connections ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conns []Connection
	for rows.Next() {
		var c Connection
		if err := rows.Scan(&c.Name, &c.Site, &c.IsDefault, &c.CreatedAt); err != nil {
			return nil, err
		}
		conns = append(conns, c)
	}
	return conns, rows.Err()
}

func (s *Store) SetDefaultConnection(name string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verify connection exists
	var exists int
	if err := tx.QueryRow("SELECT count(*) FROM connections WHERE name = ?", name).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("connection %q not found", name)
	}

	// Clear existing default
	if _, err := tx.Exec("UPDATE connections SET is_default = 0"); err != nil {
		return err
	}
	// Set new default
	if _, err := tx.Exec("UPDATE connections SET is_default = 1 WHERE name = ?", name); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) RemoveConnection(name string) error {
	result, err := s.db.Exec("DELETE FROM connections WHERE name = ?", name)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("connection %q not found", name)
	}
	return nil
}
```

**Step 4: Run tests**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/store/... -v`
Expected: PASS (all tests including previous store tests)

**Step 5: Commit**

```bash
git add internal/store/connections.go internal/store/connections_test.go
git commit -m "feat: add connection CRUD operations in SQLite store"
```

---

### Task 6: Resource and Version CRUD Operations

**Files:**

- Create: `internal/store/resources.go`
- Create: `internal/store/resources_test.go`

**Step 1: Write resource tests**

```go
// internal/store/resources_test.go
package store

import (
	"testing"
)

func seedConnection(t *testing.T, s *Store) {
	t.Helper()
	if err := s.AddConnection("prod", "datadoghq.com"); err != nil {
		t.Fatal(err)
	}
}

func TestTrackResource(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)

	err := s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")
	if err != nil {
		t.Fatal(err)
	}

	resources, err := s.ListResources("dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].ResourceID != "abc-123" {
		t.Errorf("expected id 'abc-123', got %q", resources[0].ResourceID)
	}
}

func TestSaveVersion(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	err := s.SaveVersion("abc-123", "dashboard", "prod", "content here", "", "gerald", "initial version")
	if err != nil {
		t.Fatal(err)
	}

	versions, err := s.ListVersions("abc-123", "dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].Version != 1 {
		t.Errorf("expected version 1, got %d", versions[0].Version)
	}
	if versions[0].Message != "initial version" {
		t.Errorf("expected message 'initial version', got %q", versions[0].Message)
	}
}

func TestSaveVersion_AutoIncrements(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	_ = s.SaveVersion("abc-123", "dashboard", "prod", "v1", "", "gerald", "first")
	_ = s.SaveVersion("abc-123", "dashboard", "prod", "v2", "", "gerald", "second")

	versions, err := s.ListVersions("abc-123", "dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	// Versions returned newest first
	if versions[0].Version != 2 {
		t.Errorf("expected latest version 2, got %d", versions[0].Version)
	}
}

func TestGetLatestVersion(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	_ = s.SaveVersion("abc-123", "dashboard", "prod", "old content", "", "gerald", "first")
	_ = s.SaveVersion("abc-123", "dashboard", "prod", "new content", "", "gerald", "second")

	v, err := s.GetLatestVersion("abc-123", "dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if v.Content != "new content" {
		t.Errorf("expected 'new content', got %q", v.Content)
	}
}

func TestGetVersion(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	_ = s.SaveVersion("abc-123", "dashboard", "prod", "v1 content", "", "gerald", "first")
	_ = s.SaveVersion("abc-123", "dashboard", "prod", "v2 content", "", "gerald", "second")

	v, err := s.GetVersion("abc-123", "dashboard", "prod", 1)
	if err != nil {
		t.Fatal(err)
	}
	if v.Content != "v1 content" {
		t.Errorf("expected 'v1 content', got %q", v.Content)
	}
}

func TestPruneVersions(t *testing.T) {
	s := newTestStore(t)
	seedConnection(t, s)
	_ = s.TrackResource("abc-123", "dashboard", "prod", "My Dashboard")

	for i := 0; i < 5; i++ {
		_ = s.SaveVersion("abc-123", "dashboard", "prod", "content", "", "gerald", "")
	}

	pruned, err := s.PruneVersions("abc-123", "dashboard", "prod", 3)
	if err != nil {
		t.Fatal(err)
	}
	if pruned != 2 {
		t.Errorf("expected 2 pruned, got %d", pruned)
	}

	versions, err := s.ListVersions("abc-123", "dashboard", "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 {
		t.Errorf("expected 3 remaining, got %d", len(versions))
	}
}
```

**Step 2: Run tests to verify failure**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/store/... -v -run TestTrack`
Expected: FAIL

**Step 3: Implement resource operations**

```go
// internal/store/resources.go
package store

import (
	"database/sql"
	"fmt"
	"os/user"
	"time"
)

type Resource struct {
	ResourceID       string
	ResourceType     string
	Connection       string
	Title            string
	RemoteModifiedAt *time.Time
	RemoteModifiedBy string
	RemoteEtag       string
	LastSyncedAt     *time.Time
	Status           string
}

type ResourceVersion struct {
	ID             int
	ResourceID     string
	ResourceType   string
	Connection     string
	Version        int
	Content        string
	RemoteSnapshot string
	AppliedAt      time.Time
	AppliedBy      string
	Message        string
}

func (s *Store) TrackResource(resourceID, resourceType, connection, title string) error {
	_, err := s.db.Exec(
		`INSERT INTO resources (resource_id, resource_type, connection, title)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(resource_id, resource_type, connection)
		 DO UPDATE SET title = excluded.title`,
		resourceID, resourceType, connection, title,
	)
	return err
}

func (s *Store) ListResources(resourceType, connection string) ([]Resource, error) {
	rows, err := s.db.Query(
		`SELECT resource_id, resource_type, connection, title, remote_modified_at,
		        remote_modified_by, last_synced_at, status
		 FROM resources WHERE resource_type = ? AND connection = ?
		 ORDER BY title`,
		resourceType, connection,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		var r Resource
		if err := rows.Scan(&r.ResourceID, &r.ResourceType, &r.Connection,
			&r.Title, &r.RemoteModifiedAt, &r.RemoteModifiedBy,
			&r.LastSyncedAt, &r.Status); err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, rows.Err()
}

func (s *Store) SaveVersion(resourceID, resourceType, connection, content, remoteSnapshot, appliedBy, message string) error {
	if appliedBy == "" {
		if u, err := user.Current(); err == nil {
			appliedBy = u.Username
		}
	}

	// Get next version number
	var maxVersion int
	err := s.db.QueryRow(
		`SELECT COALESCE(MAX(version), 0) FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?`,
		resourceID, resourceType, connection,
	).Scan(&maxVersion)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO resource_versions (resource_id, resource_type, connection, version, content, remote_snapshot, applied_by, message)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		resourceID, resourceType, connection, maxVersion+1, content, remoteSnapshot, appliedBy, message,
	)
	return err
}

func (s *Store) ListVersions(resourceID, resourceType, connection string) ([]ResourceVersion, error) {
	rows, err := s.db.Query(
		`SELECT id, resource_id, resource_type, connection, version, content, remote_snapshot, applied_at, applied_by, message
		 FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?
		 ORDER BY version DESC`,
		resourceID, resourceType, connection,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []ResourceVersion
	for rows.Next() {
		var v ResourceVersion
		var remoteSnapshot sql.NullString
		var appliedBy sql.NullString
		var message sql.NullString
		if err := rows.Scan(&v.ID, &v.ResourceID, &v.ResourceType, &v.Connection,
			&v.Version, &v.Content, &remoteSnapshot, &v.AppliedAt, &appliedBy, &message); err != nil {
			return nil, err
		}
		v.RemoteSnapshot = remoteSnapshot.String
		v.AppliedBy = appliedBy.String
		v.Message = message.String
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (s *Store) GetLatestVersion(resourceID, resourceType, connection string) (*ResourceVersion, error) {
	v := &ResourceVersion{}
	var remoteSnapshot, appliedBy, message sql.NullString
	err := s.db.QueryRow(
		`SELECT id, resource_id, resource_type, connection, version, content, remote_snapshot, applied_at, applied_by, message
		 FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?
		 ORDER BY version DESC LIMIT 1`,
		resourceID, resourceType, connection,
	).Scan(&v.ID, &v.ResourceID, &v.ResourceType, &v.Connection,
		&v.Version, &v.Content, &remoteSnapshot, &v.AppliedAt, &appliedBy, &message)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no versions found for %s/%s", resourceType, resourceID)
	}
	if err != nil {
		return nil, err
	}
	v.RemoteSnapshot = remoteSnapshot.String
	v.AppliedBy = appliedBy.String
	v.Message = message.String
	return v, nil
}

func (s *Store) GetVersion(resourceID, resourceType, connection string, version int) (*ResourceVersion, error) {
	v := &ResourceVersion{}
	var remoteSnapshot, appliedBy, msg sql.NullString
	err := s.db.QueryRow(
		`SELECT id, resource_id, resource_type, connection, version, content, remote_snapshot, applied_at, applied_by, message
		 FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ? AND version = ?`,
		resourceID, resourceType, connection, version,
	).Scan(&v.ID, &v.ResourceID, &v.ResourceType, &v.Connection,
		&v.Version, &v.Content, &remoteSnapshot, &v.AppliedAt, &appliedBy, &msg)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("version %d not found for %s/%s", version, resourceType, resourceID)
	}
	if err != nil {
		return nil, err
	}
	v.RemoteSnapshot = remoteSnapshot.String
	v.AppliedBy = appliedBy.String
	v.Message = msg.String
	return v, nil
}

func (s *Store) PruneVersions(resourceID, resourceType, connection string, keep int) (int, error) {
	result, err := s.db.Exec(
		`DELETE FROM resource_versions
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?
		   AND version NOT IN (
		     SELECT version FROM resource_versions
		     WHERE resource_id = ? AND resource_type = ? AND connection = ?
		     ORDER BY version DESC LIMIT ?
		   )`,
		resourceID, resourceType, connection,
		resourceID, resourceType, connection, keep,
	)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

func (s *Store) UpdateResourceSync(resourceID, resourceType, connection string, modifiedAt *time.Time, modifiedBy, etag string) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE resources SET remote_modified_at = ?, remote_modified_by = ?, remote_etag = ?, last_synced_at = ?
		 WHERE resource_id = ? AND resource_type = ? AND connection = ?`,
		modifiedAt, modifiedBy, etag, now, resourceID, resourceType, connection,
	)
	return err
}
```

**Step 4: Run tests**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/store/... -v`
Expected: PASS (all tests)

**Step 5: Commit**

```bash
git add internal/store/resources.go internal/store/resources_test.go
git commit -m "feat: add resource and version CRUD with auto-increment and pruning"
```

---

### Task 7: Keyring Authentication

**Files:**

- Create: `internal/keyring/keyring.go`
- Create: `internal/keyring/keyring_test.go`

**Step 1: Write keyring tests**

```go
// internal/keyring/keyring_test.go
package keyring

import (
	"os"
	"testing"
)

func TestResolveCredentials_EnvVars(t *testing.T) {
	t.Setenv("DD_API_KEY", "test-api-key")
	t.Setenv("DD_APP_KEY", "test-app-key")

	creds, source, err := ResolveCredentials("prod")
	if err != nil {
		t.Fatal(err)
	}
	if creds.APIKey != "test-api-key" {
		t.Errorf("expected api key 'test-api-key', got %q", creds.APIKey)
	}
	if creds.AppKey != "test-app-key" {
		t.Errorf("expected app key 'test-app-key', got %q", creds.AppKey)
	}
	if source != SourceEnvVar {
		t.Errorf("expected source %q, got %q", SourceEnvVar, source)
	}
}

func TestResolveCredentials_ConnectionSpecificEnvVars(t *testing.T) {
	t.Setenv("DD_STAGING_API_KEY", "staging-api")
	t.Setenv("DD_STAGING_APP_KEY", "staging-app")

	creds, source, err := ResolveCredentials("staging")
	if err != nil {
		t.Fatal(err)
	}
	if creds.APIKey != "staging-api" {
		t.Errorf("expected 'staging-api', got %q", creds.APIKey)
	}
	if source != SourceEnvVar {
		t.Errorf("expected source %q, got %q", SourceEnvVar, source)
	}
}

func TestResolveCredentials_NoCredentials(t *testing.T) {
	// Clear any DD env vars
	os.Unsetenv("DD_API_KEY")
	os.Unsetenv("DD_APP_KEY")
	os.Unsetenv("DD_PROD_API_KEY")
	os.Unsetenv("DD_PROD_APP_KEY")

	_, _, err := ResolveCredentials("prod")
	if err == nil {
		t.Error("expected error when no credentials found")
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abcdef1234567890", "••••••••••••7890"},
		{"short", "••••••••••••hort"},
		{"ab", "••••••••••••••ab"},
	}

	for _, tt := range tests {
		got := MaskKey(tt.input)
		if got != tt.expected {
			t.Errorf("MaskKey(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
```

**Step 2: Run tests to verify failure**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/keyring/... -v`
Expected: FAIL

**Step 3: Implement keyring**

```go
// internal/keyring/keyring.go
package keyring

import (
	"fmt"
	"os"
	"strings"

	gokeyring "github.com/zalando/go-keyring"
)

const serviceName = "ddctl"

type Credentials struct {
	APIKey string
	AppKey string
}

type Source string

const (
	SourceEnvVar  Source = "environment"
	SourceKeyring Source = "keychain"
	SourceFile    Source = "credential_file"
)

func ResolveCredentials(connection string) (*Credentials, Source, error) {
	// 1. Connection-specific env vars
	upper := strings.ToUpper(connection)
	if apiKey := os.Getenv("DD_" + upper + "_API_KEY"); apiKey != "" {
		if appKey := os.Getenv("DD_" + upper + "_APP_KEY"); appKey != "" {
			return &Credentials{APIKey: apiKey, AppKey: appKey}, SourceEnvVar, nil
		}
	}

	// 2. Generic env vars
	if apiKey := os.Getenv("DD_API_KEY"); apiKey != "" {
		if appKey := os.Getenv("DD_APP_KEY"); appKey != "" {
			return &Credentials{APIKey: apiKey, AppKey: appKey}, SourceEnvVar, nil
		}
	}

	// 3. System keychain
	apiKey, err := gokeyring.Get(serviceName, connection+"/api_key")
	if err == nil {
		appKey, err := gokeyring.Get(serviceName, connection+"/app_key")
		if err == nil {
			return &Credentials{APIKey: apiKey, AppKey: appKey}, SourceKeyring, nil
		}
	}

	return nil, "", fmt.Errorf("no credentials found for connection %q.\n\nSetup options:\n  ddctl auth login                  # store in keychain\n  export DD_API_KEY=<key>           # environment variable\n  export DD_APP_KEY=<key>", connection)
}

func StoreCredentials(connection string, creds *Credentials) error {
	if err := gokeyring.Set(serviceName, connection+"/api_key", creds.APIKey); err != nil {
		return fmt.Errorf("storing API key in keychain: %w", err)
	}
	if err := gokeyring.Set(serviceName, connection+"/app_key", creds.AppKey); err != nil {
		return fmt.Errorf("storing App key in keychain: %w", err)
	}
	return nil
}

func DeleteCredentials(connection string) error {
	_ = gokeyring.Delete(serviceName, connection+"/api_key")
	_ = gokeyring.Delete(serviceName, connection+"/app_key")
	return nil
}

func MaskKey(key string) string {
	if len(key) <= 4 {
		return "••••••••••••••" + key
	}
	return "••••••••••••" + key[len(key)-4:]
}
```

**Step 4: Run tests**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/keyring/... -v`
Expected: PASS (env var tests pass; keyring tests may require mock — the env var path is the critical one)

Note: The keyring system calls (`gokeyring.Get/Set`) require a desktop session. Tests that exercise the env var path will pass in any environment. Tests for the keyring path are best done manually or in CI with a mock keyring setup.

**Step 5: Commit**

```bash
git add internal/keyring/
git commit -m "feat: add credential resolution (env vars, keychain) with masking"
```

---

### Task 8: Output Formatters

**Files:**

- Create: `internal/output/output.go`
- Create: `internal/output/output_test.go`

**Step 1: Write output tests**

```go
// internal/output/output_test.go
package output

import (
	"bytes"
	"testing"
)

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test", "id": "123"}

	err := JSON(&buf, data)
	if err != nil {
		t.Fatal(err)
	}

	expected := "{\n  \"id\": \"123\",\n  \"name\": \"test\"\n}\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestYAML(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test"}

	err := YAML(&buf, data)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("name: test")) {
		t.Errorf("expected YAML with 'name: test', got %q", buf.String())
	}
}
```

**Step 2: Run tests to verify failure**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/output/... -v`
Expected: FAIL

**Step 3: Implement output formatters**

```go
// internal/output/output.go
package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func JSON(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func YAML(w io.Writer, data any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(data)
}

func Table(w io.Writer, headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Fprintln(w, "  No results.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Fprintf(w, "  %-*s", widths[i]+2, h)
	}
	fmt.Fprintln(w)

	// Print separator
	for i := range headers {
		fmt.Fprintf(w, "  ")
		for j := 0; j < widths[i]; j++ {
			fmt.Fprintf(w, "─")
		}
	}
	fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(w, "  %-*s", widths[i]+2, cell)
			}
		}
		fmt.Fprintln(w)
	}
}

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code" yaml:"code"`
		Message string `json:"message" yaml:"message"`
	} `json:"error" yaml:"error"`
}

func Error(w io.Writer, format string, code int, message string) {
	switch format {
	case "json":
		resp := ErrorResponse{}
		resp.Error.Code = code
		resp.Error.Message = message
		JSON(w, resp)
	case "yaml":
		resp := ErrorResponse{}
		resp.Error.Code = code
		resp.Error.Message = message
		YAML(w, resp)
	default:
		fmt.Fprintf(w, "Error: %s\n", message)
	}
}
```

**Step 4: Run tests**

Run: `cd ~/Documents/dev/ddctl && go test ./internal/output/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/output/
git commit -m "feat: add JSON, YAML, and table output formatters"
```

---

### Task 9: Theme System (Retro Cyberpunk)

**Files:**

- Create: `internal/theme/theme.go`
- Create: `internal/theme/banner.go`

**Step 1: Implement theme**

```go
// internal/theme/theme.go
package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name    string
	Enabled bool
}

var (
	Green     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff41"))
	DimGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00cc33"))
	Cyan      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00cccc"))
	BrightCyan = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff"))
	Yellow    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffbf00"))
	Red       = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444"))
	Dim       = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
)

func New(name string) *Theme {
	enabled := name != "none" && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
	return &Theme{Name: name, Enabled: enabled}
}

func (t *Theme) IsRetro() bool {
	return t.Enabled && t.Name == "retro"
}

func (t *Theme) Success(s string) string {
	if !t.Enabled {
		return s
	}
	return BrightCyan.Render(s)
}

func (t *Theme) Warning(s string) string {
	if !t.Enabled {
		return s
	}
	return Yellow.Render(s)
}

func (t *Theme) Err(s string) string {
	if !t.Enabled {
		return s
	}
	return Red.Render(s)
}

func (t *Theme) Highlight(s string) string {
	if !t.Enabled {
		return s
	}
	return Green.Render(s)
}

func (t *Theme) ID(s string) string {
	if !t.Enabled {
		return s
	}
	return Cyan.Render(s)
}

func (t *Theme) Label(s string) string {
	if !t.Enabled {
		return s
	}
	return DimGreen.Render(s)
}

func (t *Theme) Flavor(s string) string {
	if !t.Enabled || t.Name != "retro" {
		return ""
	}
	return DimGreen.Italic(true).Render(s)
}
```

```go
// internal/theme/banner.go
package theme

import (
	"fmt"
	"math/rand"
)

var quotes = []string{
	`"The net is vast and infinite." — Ghost in the Shell`,
	`"I never asked for this." — Adam Jensen`,
	`"Shall we play a game?" — WOPR`,
	`"In a world of locked doors, the man with the key is king."`,
	`"I'm in." — Every 90s hacker movie`,
	`"The more things change, the more they stay the same."`,
}

func (t *Theme) Banner(version string) string {
	if !t.IsRetro() {
		return ""
	}

	quote := quotes[rand.Intn(len(quotes))]

	banner := fmt.Sprintf(`
    %s
    %s
    %s
    %s  %s
    %s  %s
    %s  %s
    %s   %s
    %s   %s
    %s
    %s  %s
    %s
    %s
`,
		Green.Render("╔══════════════════════════════════════╗"),
		Green.Render("║                                      ║"),
		Green.Render("║       ▄▄▄                            ║"),
		Green.Render("║      █▀█▀█"), Green.Render("┏━╸╺┳╸╻                 ║"),
		Green.Render("║      █▄█▄█"), Green.Render("┃   ┃ ┃                 ║"),
		Green.Render("║      ╰█▀█╯"), Green.Render("┗━╸ ╹ ┗━╸              ║"),
		Green.Render("║       ╰─╯"), Green.Render("d a t a d o g           ║"),
		Green.Render("║          "), Green.Render("c o n t r o l           ║"),
		Green.Render("║                                      ║"),
		Green.Render("║  ["+version+"]"), Green.Render("   ◄◄ JACK IN ►►          ║"),
		Green.Render("║                                      ║"),
		Green.Render("╚══════════════════════════════════════╝"),
	)

	banner += "\n  " + DimGreen.Italic(true).Render(quote) + "\n"

	return banner
}
```

**Step 2: Verify it compiles**

Run: `cd ~/Documents/dev/ddctl && go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/theme/
git commit -m "feat: add retro cyberpunk theme with banner, colors, and quotes"
```

---

### Task 10: Datadog API Client Wrapper

**Files:**

- Create: `internal/client/client.go`
- Create: `internal/client/dashboards.go`

**Step 1: Implement the connection-aware client**

```go
// internal/client/client.go
package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/futuregerald/ddctl/internal/keyring"
)

type Client struct {
	apiClient *datadog.APIClient
	ctx       context.Context
	mu        sync.Mutex
	lastCall  time.Time
	rateLimit time.Duration
}

func New(site string, creds *keyring.Credentials) (*Client, error) {
	ctx := datadog.NewDefaultContext(context.Background())
	ctx = context.WithValue(ctx, datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: creds.APIKey},
		"appKeyAuth": {Key: creds.AppKey},
	})
	ctx = context.WithValue(ctx, datadog.ContextServerVariables, map[string]string{
		"site": site,
	})

	config := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(config)

	return &Client{
		apiClient: apiClient,
		ctx:       ctx,
		rateLimit: 2 * time.Second, // 30 req/min default
	}, nil
}

func (c *Client) throttle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	since := time.Since(c.lastCall)
	if since < c.rateLimit {
		time.Sleep(c.rateLimit - since)
	}
	c.lastCall = time.Now()
}

func (c *Client) APIClient() *datadog.APIClient {
	return c.apiClient
}

func (c *Client) Context() context.Context {
	return c.ctx
}

func (c *Client) Validate() error {
	c.throttle()
	_, _, err := c.apiClient.AuthenticationApi.Validate(c.ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}
```

```go
// internal/client/dashboards.go
package client

import (
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func (c *Client) ListDashboards() ([]datadogV1.DashboardSummaryDefinition, error) {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)
	resp, _, err := api.ListDashboards(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing dashboards: %w", err)
	}
	return resp.GetDashboards(), nil
}

func (c *Client) GetDashboard(id string) ([]byte, error) {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)
	resp, _, err := api.GetDashboard(c.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting dashboard %s: %w", id, err)
	}
	return json.MarshalIndent(resp, "", "  ")
}

func (c *Client) CreateDashboard(body []byte) (string, error) {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)

	var dashboard datadogV1.Dashboard
	if err := json.Unmarshal(body, &dashboard); err != nil {
		return "", fmt.Errorf("parsing dashboard JSON: %w", err)
	}

	resp, _, err := api.CreateDashboard(c.ctx, dashboard)
	if err != nil {
		return "", fmt.Errorf("creating dashboard: %w", err)
	}
	return resp.GetId(), nil
}

func (c *Client) UpdateDashboard(id string, body []byte) error {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)

	var dashboard datadogV1.Dashboard
	if err := json.Unmarshal(body, &dashboard); err != nil {
		return fmt.Errorf("parsing dashboard JSON: %w", err)
	}

	_, _, err := api.UpdateDashboard(c.ctx, id, dashboard)
	if err != nil {
		return fmt.Errorf("updating dashboard %s: %w", id, err)
	}
	return nil
}

func (c *Client) DeleteDashboard(id string) error {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)
	_, _, err := api.DeleteDashboard(c.ctx, id)
	if err != nil {
		return fmt.Errorf("deleting dashboard %s: %w", id, err)
	}
	return nil
}
```

**Step 2: Verify it compiles**

Run: `cd ~/Documents/dev/ddctl && go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/client/
git commit -m "feat: add Datadog API client with rate limiting and dashboard operations"
```

---

## Phase 2: CLI Commands

### Task 11: Connection Commands

**Files:**

- Create: `cmd/connection/connection.go`
- Create: `cmd/connection/add.go`
- Create: `cmd/connection/list.go`
- Create: `cmd/connection/default_cmd.go`
- Create: `cmd/connection/remove.go`
- Create: `cmd/connection/test_cmd.go`

These commands wire together the store and keyring packages. Each is a thin Cobra command that:
1. Opens the store
2. Calls the appropriate store/keyring method
3. Formats output

**Implementation:** Create each file following the Cobra subcommand pattern. The `connection.go` file creates the parent command and adds subcommands. Each subcommand file registers itself in `init()`.

**Step 1: Implement all connection commands** (see design doc for full command signatures)

**Step 2: Verify**

Run: `cd ~/Documents/dev/ddctl && go build -o ddctl . && ./ddctl connection --help`
Expected: Help text showing add, list, default, remove, test subcommands

**Step 3: Commit**

```bash
git add cmd/connection/
git commit -m "feat: add connection management commands (add, list, default, remove, test)"
```

---

### Task 12: Auth Commands

**Files:**

- Create: `cmd/auth/auth.go`
- Create: `cmd/auth/login.go`
- Create: `cmd/auth/status.go`
- Create: `cmd/auth/logout.go`

**Implementation:** `login` prompts for API key + app key, validates via `client.Validate()`, stores in keychain via `keyring.StoreCredentials()`. `status` shows credential source and masked keys. `logout` removes from keychain.

**Step 1: Implement auth commands**

**Step 2: Verify**

Run: `cd ~/Documents/dev/ddctl && go build -o ddctl . && ./ddctl auth --help`
Expected: Help text showing login, status, logout subcommands

**Step 3: Commit**

```bash
git add cmd/auth/
git commit -m "feat: add auth commands (login, status, logout)"
```

---

### Task 13: Dashboard Commands

**Files:**

- Create: `cmd/dashboard/dashboard.go`
- Create: `cmd/dashboard/list.go`
- Create: `cmd/dashboard/pull.go`
- Create: `cmd/dashboard/push.go`
- Create: `cmd/dashboard/edit.go`
- Create: `cmd/dashboard/diff.go`
- Create: `cmd/dashboard/history.go`
- Create: `cmd/dashboard/rollback.go`
- Create: `cmd/dashboard/import_cmd.go`
- Create: `cmd/dashboard/export.go`
- Create: `cmd/dashboard/delete.go`
- Create: `cmd/dashboard/create.go`
- Create: `cmd/dashboard/sync.go`
- Create: `cmd/dashboard/examples.go`

This is the largest command set. Each command follows the workflow defined in the design doc:

- `list` — queries SQLite (local) or Datadog API (`--remote`)
- `pull` — fetches from Datadog, stores in SQLite as new version
- `push` — implements full push safety workflow (fetch, diff, confirm, re-fetch, apply)
- `edit` — temp file + editor + save back to SQLite
- `diff` — compares local latest vs remote (or local vs specific version)
- `history` — lists versions from SQLite
- `rollback` — copies an old version as a new version in SQLite
- `import` — reads YAML file, stores in SQLite, optionally pushes
- `export` — writes SQLite content to YAML file
- `delete` — deletes from Datadog with confirmation
- `create` — creates on Datadog, pulls back to SQLite
- `sync` — pulls all tracked dashboards
- `examples` — lists/previews/imports bundled examples

**Step 1: Implement each command file**

**Step 2: Verify**

Run: `cd ~/Documents/dev/ddctl && go build -o ddctl . && ./ddctl dashboard --help`
Expected: Help text showing all dashboard subcommands

**Step 3: Commit**

```bash
git add cmd/dashboard/
git commit -m "feat: add dashboard commands (list, pull, push, edit, diff, history, rollback, import, export, delete, create, sync, examples)"
```

---

### Task 14: Monitor Commands

**Files:**

- Create: `cmd/monitor/monitor.go`
- Create: `cmd/monitor/list.go`
- Create: `cmd/monitor/pull.go`
- Create: `cmd/monitor/push.go`
- Create: `cmd/monitor/edit.go`
- Create: `cmd/monitor/history.go`
- Create: `cmd/monitor/rollback.go`
- Create: `cmd/monitor/import_cmd.go`
- Create: `cmd/monitor/export.go`
- Create: `cmd/monitor/delete.go`
- Create: `cmd/monitor/create.go`
- Create: `internal/client/monitors.go`

Same pattern as dashboards. Add monitor API methods to the client, then wire up Cobra commands.

**Step 1: Add monitor methods to client (`internal/client/monitors.go`)**
**Step 2: Implement monitor commands**
**Step 3: Verify and commit**

```bash
git add internal/client/monitors.go cmd/monitor/
git commit -m "feat: add monitor commands and API client methods"
```

---

### Task 15: SLO Commands

**Files:**

- Create: `cmd/slo/slo.go`
- Create: `cmd/slo/list.go`
- Create: `cmd/slo/pull.go`
- Create: `cmd/slo/push.go`
- Create: `cmd/slo/edit.go`
- Create: `cmd/slo/history.go`
- Create: `cmd/slo/rollback.go`
- Create: `cmd/slo/import_cmd.go`
- Create: `cmd/slo/export.go`
- Create: `cmd/slo/delete.go`
- Create: `cmd/slo/create.go`
- Create: `internal/client/slos.go`

Same pattern as dashboards and monitors.

**Step 1: Add SLO methods to client**
**Step 2: Implement SLO commands**
**Step 3: Verify and commit**

```bash
git add internal/client/slos.go cmd/slo/
git commit -m "feat: add SLO commands and API client methods"
```

---

### Task 16: Metrics and Logs Read Commands

**Files:**

- Create: `cmd/metrics/metrics.go`
- Create: `cmd/metrics/search.go`
- Create: `cmd/metrics/query.go`
- Create: `cmd/logs/logs.go`
- Create: `cmd/logs/search.go`
- Create: `internal/client/metrics.go`
- Create: `internal/client/logs.go`

Read-only commands for gathering context when creating dashboards.

**Step 1: Add metrics/logs API methods to client**
**Step 2: Implement commands**
**Step 3: Verify and commit**

```bash
git add internal/client/metrics.go internal/client/logs.go cmd/metrics/ cmd/logs/
git commit -m "feat: add metrics and logs read commands"
```

---

### Task 17: API Pass-through Command

**Files:**

- Create: `cmd/api/api.go`
- Create: `cmd/api/list.go`
- Create: `cmd/api/call.go`
- Create: `cmd/api/raw.go`

`api list` shows all available API groups. `api list <group>` shows operations. `api <group>.<operation>` executes. `api raw <method> <path>` is the HTTP escape hatch.

For v1, the `api list` and `api <group>.<operation>` can use a hardcoded mapping of the most common operations. The `go generate` step from the OpenAPI spec can be added as a fast follow.

**Step 1: Implement API pass-through**
**Step 2: Verify and commit**

```bash
git add cmd/api/
git commit -m "feat: add API pass-through commands (list, call, raw)"
```

---

### Task 18: DB Management Commands

**Files:**

- Create: `cmd/db/db.go`
- Create: `cmd/db/prune.go`
- Create: `cmd/db/stats.go`

`db prune` runs version pruning across all resources. `db stats` shows database size, version counts, etc.

**Step 1: Implement db commands**
**Step 2: Verify and commit**

```bash
git add cmd/db/
git commit -m "feat: add db management commands (prune, stats)"
```

---

## Phase 3: MCP Server

### Task 19: MCP Server Implementation

**Files:**

- Create: `internal/mcp/server.go`
- Create: `internal/mcp/tools.go`
- Create: `cmd/mcp/mcp.go`

The MCP server uses `mcp-go` to expose CLI commands as tools via stdio transport. Each tool maps to a CLI command. Output is always JSON. Safety levels control which tools are available.

**Step 1: Implement MCP server with tool registration**
**Step 2: Implement `ddctl mcp serve` command with `--safety` and `--connection` flags**
**Step 3: Verify and commit**

```bash
git add internal/mcp/ cmd/mcp/
git commit -m "feat: add MCP server with safety levels and tool mapping"
```

---

## Phase 4: Examples and Documentation

### Task 20: Bundled Examples

**Files:**

- Create: `examples/dashboards/service-overview.yaml`
- Create: `examples/dashboards/log-analytics.yaml`
- Create: `examples/dashboards/slo-tracking.yaml`
- Create: `examples/dashboards/deployment.yaml`
- Create: `examples/monitors/error-rate-alert.yaml`
- Create: `examples/slos/availability-slo.yaml`

Use Go's `embed` package to bundle these in the binary.

**Step 1: Create example YAML files with realistic Datadog API format**
**Step 2: Embed them in the dashboard examples command**
**Step 3: Commit**

```bash
git add examples/
git commit -m "feat: add bundled example dashboards, monitors, and SLOs"
```

---

### Task 21: Documentation

**Files:**

- Create: `README.md` (rewrite with full content)
- Create: `CLAUDE.md`
- Create: `docs/getting-started.md`
- Create: `docs/configuration.md`
- Create: `docs/authentication.md`
- Create: `docs/dashboards.md`
- Create: `docs/monitors.md`
- Create: `docs/slos.md`
- Create: `docs/api-passthrough.md`
- Create: `docs/mcp-server.md`
- Create: `docs/versioning.md`
- Create: `docs/dsl-reference.md`

**Step 1: Write all documentation per design doc spec**
**Step 2: Commit**

```bash
git add README.md CLAUDE.md docs/
git commit -m "docs: add comprehensive documentation and CLAUDE.md"
```

---

## Phase 5: Polish and Release

### Task 22: Build Configuration

**Files:**

- Create: `Makefile`
- Create: `.goreleaser.yaml` (optional, for future releases)

**Step 1: Create Makefile**

```makefile
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/futuregerald/ddctl/cmd.Version=$(VERSION) -X github.com/futuregerald/ddctl/cmd.CommitSHA=$(COMMIT) -X github.com/futuregerald/ddctl/cmd.BuildDate=$(DATE)"

.PHONY: build test lint clean

build:
	go build $(LDFLAGS) -o ddctl .

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -f ddctl
```

**Step 2: Verify**

Run: `make build && ./ddctl version`
Expected: Version with real commit SHA and date

**Step 3: Commit**

```bash
git add Makefile
git commit -m "feat: add Makefile with build, test, and lint targets"
```

---

### Task 23: Final Integration Test

Run the full test suite:

```bash
cd ~/Documents/dev/ddctl
make test
make build
./ddctl --help
./ddctl version
./ddctl dashboard --help
./ddctl connection --help
./ddctl auth --help
./ddctl mcp --help
./ddctl api --help
```

Fix any issues, then final commit:

```bash
git commit -m "chore: final integration verification"
```
