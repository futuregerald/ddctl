// cmd/metrics/metrics.go
package metrics

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for metrics operations.
var Cmd = &cobra.Command{
	Use:   "metrics",
	Short: "Search and query Datadog metrics",
	Long:  `Read-only commands for searching available metrics and querying timeseries data.`,
}

func init() {
	Cmd.AddCommand(searchCmd)
	Cmd.AddCommand(queryCmd)
}
