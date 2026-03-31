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
