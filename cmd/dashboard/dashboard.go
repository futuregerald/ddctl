// cmd/dashboard/dashboard.go
package dashboard

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for dashboard management.
var Cmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Manage Datadog dashboards",
	Long:  `Create, edit, push, pull, and version-control Datadog dashboards.`,
	Aliases: []string{"dash"},
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(pullCmd)
	Cmd.AddCommand(pushCmd)
	Cmd.AddCommand(editCmd)
	Cmd.AddCommand(diffCmd)
	Cmd.AddCommand(historyCmd)
	Cmd.AddCommand(rollbackCmd)
	Cmd.AddCommand(importCmd)
	Cmd.AddCommand(exportCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(syncCmd)
	Cmd.AddCommand(examplesCmd)
}
