// cmd/connection/add.go
package connection

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

var (
	addFlagName string
	addFlagSite string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new connection profile",
	Long:  `Add a new Datadog connection profile. Prompts for API and App keys, validates them, and stores in the system keychain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		name := addFlagName
		if name == "" {
			fmt.Fprint(os.Stderr, "Connection name: ")
			n, _ := reader.ReadString('\n')
			name = strings.TrimSpace(n)
		}
		if name == "" {
			return fmt.Errorf("connection name is required")
		}

		site := addFlagSite
		if site == "" {
			fmt.Fprint(os.Stderr, "Datadog site (e.g. datadoghq.com): ")
			s, _ := reader.ReadString('\n')
			site = strings.TrimSpace(s)
		}
		if site == "" {
			site = "datadoghq.com"
		}

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

		// Validate credentials
		fmt.Fprint(os.Stderr, "Validating credentials... ")
		creds := &keyring.Credentials{APIKey: apiKey, AppKey: appKey}
		c, err := client.New(site, creds)
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}
		if err := c.Validate(); err != nil {
			fmt.Fprintln(os.Stderr, "FAILED")
			return fmt.Errorf("credential validation failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, "OK")

		// Store in keychain
		if err := keyring.StoreCredentials(name, creds); err != nil {
			return fmt.Errorf("storing credentials: %w", err)
		}

		// Store in SQLite
		s, err := store.New(config.DBPath())
		if err != nil {
			return fmt.Errorf("opening store: %w", err)
		}
		defer s.Close()

		if err := s.AddConnection(name, site); err != nil {
			return fmt.Errorf("saving connection: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Connection %q added successfully.\n", name)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addFlagName, "name", "", "Connection name")
	addCmd.Flags().StringVar(&addFlagSite, "site", "", "Datadog site (e.g. datadoghq.com)")
}
