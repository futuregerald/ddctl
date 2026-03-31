// cmd/connection/remove.go
package connection

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a connection profile",
	Long:  `Remove a connection profile and its stored credentials from the keychain.`,
	Aliases: []string{"rm"},
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		s, err := store.New(config.DBPath())
		if err != nil {
			return fmt.Errorf("opening store: %w", err)
		}
		defer s.Close()

		// Check if this connection is the default before removing
		conn, err := s.GetConnection(name)
		if err != nil {
			return err
		}
		wasDefault := conn.IsDefault

		if err := s.RemoveConnection(name); err != nil {
			return err
		}

		// Remove credentials from keychain (best effort)
		keyring.DeleteCredentials(name)

		fmt.Fprintf(os.Stderr, "Connection %q removed.\n", name)

		// If the removed connection was the default, reassign
		if wasDefault {
			remaining, err := s.ListConnections()
			if err == nil && len(remaining) > 0 {
				if err := s.SetDefaultConnection(remaining[0].Name); err == nil {
					fmt.Fprintf(os.Stderr, "Default connection reassigned to %q.\n", remaining[0].Name)
				}
			} else if err == nil {
				fmt.Fprintln(os.Stderr, "Warning: no remaining connections. Run 'ddctl connection add' to create one.")
			}
		}

		return nil
	},
}
