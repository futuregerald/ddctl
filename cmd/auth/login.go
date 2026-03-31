// cmd/auth/login.go
package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/futuregerald/ddctl/internal/client"
	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store API credentials in the system keychain",
	Long:  `Prompts for Datadog API key and App key, validates them, and stores in the system keychain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(config.Path())
		connName, _ := cmd.Flags().GetString("connection")
		if connName == "" {
			connName = cmd.Root().PersistentFlags().Lookup("connection").Value.String()
		}

		// Resolve connection name
		if connName == "" {
			if cfg.DefaultConnection != "" {
				connName = cfg.DefaultConnection
			} else {
				s, err := store.New(config.DBPath())
				if err == nil {
					defer s.Close()
					if dc, err := s.GetDefaultConnection(); err == nil {
						connName = dc.Name
					}
				}
			}
		}

		if connName == "" {
			return fmt.Errorf("no connection specified. Use --connection or set a default with 'ddctl connection default'")
		}

		// Get connection from store to know the site
		s, err := store.New(config.DBPath())
		if err != nil {
			return fmt.Errorf("opening store: %w", err)
		}
		defer s.Close()

		conn, err := s.GetConnection(connName)
		if err != nil {
			return fmt.Errorf("connection %q not found. Run 'ddctl connection add' first", connName)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Fprint(os.Stderr, "API Key: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return fmt.Errorf("API key is required")
		}

		fmt.Fprint(os.Stderr, "App Key: ")
		appKey, _ := reader.ReadString('\n')
		appKey = strings.TrimSpace(appKey)
		if appKey == "" {
			return fmt.Errorf("App key is required")
		}

		// Validate
		fmt.Fprint(os.Stderr, "Validating... ")
		creds := &keyring.Credentials{APIKey: apiKey, AppKey: appKey}
		c, err := client.New(conn.Site, creds)
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		if err := c.Validate(); err != nil {
			fmt.Fprintln(os.Stderr, "FAILED")
			return err
		}
		fmt.Fprintln(os.Stderr, "OK")

		// Store
		if err := keyring.StoreCredentials(connName, creds); err != nil {
			return fmt.Errorf("storing credentials: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Credentials stored for connection %q.\n", connName)
		return nil
	},
}
