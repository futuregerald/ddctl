// cmd/monitor/history.go
package monitor

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [id]",
	Short: "Show version history for a monitor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		monID := args[0]

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		versions, err := deps.Store.ListVersions(monID, "monitor", deps.ConnName)
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			fmt.Fprintln(os.Stderr, "No versions found.")
			return nil
		}

		switch deps.Format {
		case "json":
			output.JSON(os.Stdout, versions)
		case "yaml":
			output.YAML(os.Stdout, versions)
		default:
			headers := []string{"VERSION", "APPLIED AT", "BY", "MESSAGE"}
			var rows [][]string
			for _, v := range versions {
				rows = append(rows, []string{
					fmt.Sprintf("%d", v.Version),
					v.AppliedAt.Format("2006-01-02 15:04:05"),
					v.AppliedBy,
					v.Message,
				})
			}
			output.Table(os.Stdout, headers, rows)
		}
		return nil
	},
}
