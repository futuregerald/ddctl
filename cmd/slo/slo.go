// cmd/slo/slo.go
package slo

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for SLO management.
var Cmd = &cobra.Command{
	Use:   "slo",
	Short: "Manage Datadog SLOs",
	Long:  `Create, edit, push, pull, and version-control Datadog service level objectives.`,
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
