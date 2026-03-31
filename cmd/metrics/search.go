// cmd/metrics/search.go
package metrics

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search available metrics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		result, err := deps.Client.SearchMetrics(query)
		if err != nil {
			return err
		}

		switch deps.Format {
		case "json":
			output.JSON(os.Stdout, result)
		case "yaml":
			output.YAML(os.Stdout, result)
		default:
			headers := []string{"METRIC"}
			var rows [][]string
			for _, m := range result.Metrics {
				rows = append(rows, []string{m})
			}
			output.Table(os.Stdout, headers, rows)
		}

		fmt.Fprintf(os.Stderr, "Found %d metrics.\n", len(result.Metrics))
		return nil
	},
}
