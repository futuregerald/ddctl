// cmd/connection/default_cmd.go
package connection

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var defaultCmd = &cobra.Command{
	Use:   "default [name]",
	Short: "Set the default connection profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		s, err := store.New(config.DBPath())
		if err != nil {
			return fmt.Errorf("opening store: %w", err)
		}
		defer s.Close()

		if err := s.SetDefaultConnection(name); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Default connection set to %q.\n", name)
		return nil
	},
}
