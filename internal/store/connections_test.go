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
