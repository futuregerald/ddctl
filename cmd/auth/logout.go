// cmd/auth/logout.go
package auth

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove credentials from the keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(config.Path())
		connName := cmd.Root().PersistentFlags().Lookup("connection").Value.String()

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
			return fmt.Errorf("no connection specified")
		}

		if err := keyring.DeleteCredentials(connName); err != nil {
			return fmt.Errorf("removing credentials: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Credentials removed for connection %q.\n", connName)
		return nil
	},
}
