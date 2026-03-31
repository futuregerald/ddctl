// cmd/connection/test_cmd.go
package connection

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/internal/client"
	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Test connectivity for a connection profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		s, err := store.New(config.DBPath())
		if err != nil {
			return fmt.Errorf("opening store: %w", err)
		}
		defer s.Close()

		conn, err := s.GetConnection(name)
		if err != nil {
			return err
		}

		creds, source, err := keyring.ResolveCredentials(name)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Testing connection %q (site: %s, auth: %s)... ", name, conn.Site, source)

		c, err := client.New(conn.Site, creds)
		if err != nil {
			fmt.Fprintln(os.Stderr, "FAILED")
			return fmt.Errorf("creating client: %w", err)
		}

		if err := c.Validate(); err != nil {
			fmt.Fprintln(os.Stderr, "FAILED")
			return fmt.Errorf("authentication failed: %w", err)
		}

		fmt.Fprintln(os.Stderr, "OK")
		return nil
	},
}
