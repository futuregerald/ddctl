// cmd/logs/search.go
package logs

import (
	"fmt"
	"os"
	"time"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var (
	searchFlagFrom  string
	searchFlagTo    string
	searchFlagLimit int
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search logs",
	Long:  `Search Datadog logs with a query string. Defaults to last 15 minutes.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		to := time.Now()
		from := to.Add(-15 * time.Minute)

		fromStr := searchFlagFrom
		if fromStr == "" {
			fromStr = from.Format(time.RFC3339)
		}
		toStr := searchFlagTo
		if toStr == "" {
			toStr = to.Format(time.RFC3339)
		}

		result, err := deps.Client.SearchLogs(query, fromStr, toStr, searchFlagLimit)
		if err != nil {
			return err
		}

		os.Stdout.Write(result)
		fmt.Fprintln(os.Stdout)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchFlagFrom, "from", "", "Start time (RFC3339)")
	searchCmd.Flags().StringVar(&searchFlagTo, "to", "", "End time (RFC3339)")
	searchCmd.Flags().IntVar(&searchFlagLimit, "limit", 50, "Maximum number of log entries")
}
