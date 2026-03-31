// cmd/connection/connection.go
package connection

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for connection management.
var Cmd = &cobra.Command{
	Use:   "connection",
	Short: "Manage Datadog connection profiles",
	Long:  `Add, list, remove, and test Datadog connection profiles.`,
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(defaultCmd)
	Cmd.AddCommand(removeCmd)
	Cmd.AddCommand(testCmd)
}
