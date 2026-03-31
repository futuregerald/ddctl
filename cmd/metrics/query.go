// cmd/metrics/query.go
package metrics

import (
	"fmt"
	"os"
	"time"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/spf13/cobra"
)

var (
	queryFlagFrom string
	queryFlagTo   string
)

var queryCmd = &cobra.Command{
	Use:   "query [query]",
	Short: "Query timeseries metrics",
	Long:  `Query Datadog timeseries data. Use --from and --to to specify the time range (defaults to last 1 hour).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		to := time.Now()
		from := to.Add(-1 * time.Hour)

		if queryFlagFrom != "" {
			parsed, err := time.Parse(time.RFC3339, queryFlagFrom)
			if err != nil {
				// Try relative duration
				dur, durErr := time.ParseDuration(queryFlagFrom)
				if durErr != nil {
					return fmt.Errorf("invalid --from: %w", err)
				}
				from = to.Add(-dur)
			} else {
				from = parsed
			}
		}

		if queryFlagTo != "" {
			parsed, err := time.Parse(time.RFC3339, queryFlagTo)
			if err != nil {
				return fmt.Errorf("invalid --to: %w", err)
			}
			to = parsed
		}

		result, err := deps.Client.QueryMetrics(query, from, to)
		if err != nil {
			return err
		}

		switch deps.Format {
		case "yaml":
			output.YAML(os.Stdout, result)
		default:
			// JSON is default for timeseries data
			os.Stdout.Write(result)
			fmt.Fprintln(os.Stdout)
		}
		return nil
	},
}

func init() {
	queryCmd.Flags().StringVar(&queryFlagFrom, "from", "", "Start time (RFC3339 or duration like 2h, 30m)")
	queryCmd.Flags().StringVar(&queryFlagTo, "to", "", "End time (RFC3339, defaults to now)")
}
