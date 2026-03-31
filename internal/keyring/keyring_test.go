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
