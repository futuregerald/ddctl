// cmd/monitor/monitor.go
package monitor

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for monitor management.
var Cmd = &cobra.Command{
	Use:   "monitor",
	Short: "Manage Datadog monitors",
	Long:  `Create, edit, push, pull, and version-control Datadog monitors.`,
	Aliases: []string{"mon"},
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(pullCmd)
	Cmd.AddCommand(pushCmd)
	Cmd.AddCommand(editCmd)
	Cmd.AddCommand(historyCmd)
	Cmd.AddCommand(rollbackCmd)
	Cmd.AddCommand(importCmd)
	Cmd.AddCommand(exportCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(createCmd)
}
