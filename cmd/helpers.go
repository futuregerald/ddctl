// cmd/helpers.go
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/futuregerald/ddctl/internal/client"
	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/futuregerald/ddctl/internal/theme"
)

// GetConfig loads the config and resolves output format from flags.
func GetConfig() *config.Config {
	cfg, err := config.Load(config.Path())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}
	return cfg
}

// GetOutputFormat resolves the output format from flag or config.
func GetOutputFormat(cfg *config.Config) string {
	if flagOutput != "" {
		return flagOutput
	}
	return cfg.Output
}

// GetTheme returns the theme based on config.
func GetTheme(cfg *config.Config) *theme.Theme {
	return theme.New(cfg.Theme)
}

// OpenStore opens the SQLite store.
func OpenStore() (*store.Store, error) {
	return store.New(config.DBPath())
}

// ResolveConnection resolves the connection name from flag, config, or store default.
func ResolveConnection(cfg *config.Config, s *store.Store) (string, error) {
	if flagConnection != "" {
		return flagConnection, nil
	}
	if cfg.DefaultConnection != "" {
		return cfg.DefaultConnection, nil
	}
	conn, err := s.GetDefaultConnection()
	if err != nil {
		return "", fmt.Errorf("no connection specified and no default configured.\n\nRun 'ddctl connection add' to create one")
	}
	return conn.Name, nil
}

// NewClient creates a Datadog API client for the given connection.
func NewClient(connName string, s *store.Store) (*client.Client, error) {
	conn, err := s.GetConnection(connName)
	if err != nil {
		return nil, err
	}

	creds, _, err := keyring.ResolveCredentials(connName)
	if err != nil {
		return nil, err
	}

	return client.New(conn.Site, creds)
}

// PrintOutput writes data in the configured format.
func PrintOutput(w io.Writer, format string, data any, headers []string, toRows func(any) [][]string) {
	switch format {
	case "json":
		output.JSON(w, data)
	case "yaml":
		output.YAML(w, data)
	default:
		if toRows != nil {
			output.Table(w, headers, toRows(data))
		} else {
			output.JSON(w, data)
		}
	}
}

// ExitError prints an error message and exits.
func ExitError(format string, msg string, code int) {
	output.Error(os.Stderr, format, code, msg)
	os.Exit(code)
}

// ConfirmOrSkip asks for confirmation unless --yes is set.
func ConfirmOrSkip(prompt string) bool {
	if flagYes {
		return true
	}
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes"
}
