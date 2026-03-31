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
