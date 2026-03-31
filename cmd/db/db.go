// cmd/db/db.go
package db

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for database management.
var Cmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the local SQLite database",
	Long:  `Prune old versions and view database statistics.`,
}

func init() {
	Cmd.AddCommand(pruneCmd)
	Cmd.AddCommand(statsCmd)
}
