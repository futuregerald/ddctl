// cmd/logs/logs.go
package logs

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for logs operations.
var Cmd = &cobra.Command{
	Use:   "logs",
	Short: "Search Datadog logs",
	Long:  `Read-only commands for searching and analyzing Datadog logs.`,
}

func init() {
	Cmd.AddCommand(searchCmd)
}
