// internal/store/store_test.go
package store

import (
	"os"
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
