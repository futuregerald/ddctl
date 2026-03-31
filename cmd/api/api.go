// cmd/api/api.go
package api

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for API pass-through operations.
var Cmd = &cobra.Command{
	Use:   "api",
	Short: "Datadog API pass-through",
	Long: `Execute Datadog API operations directly.

  ddctl api list                  # List all API groups
  ddctl api list dashboards       # List operations in a group
  ddctl api dashboards.list       # Execute an operation
  ddctl api raw GET /v1/dashboard # Raw HTTP request`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(callCmd)
	Cmd.AddCommand(rawCmd)
}
