// cmd/dashboard/history.go
package dashboard

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [id]",
	Short: "Show version history for a dashboard",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dashID := args[0]

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		versions, err := deps.Store.ListVersions(dashID, "dashboard", deps.ConnName)
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			fmt.Fprintln(os.Stderr, "No versions found.")
			return nil
		}

		type versionInfo struct {
			Version   int    `json:"version" yaml:"version"`
			AppliedAt string `json:"applied_at" yaml:"applied_at"`
			AppliedBy string `json:"applied_by,omitempty" yaml:"applied_by,omitempty"`
			Message   string `json:"message,omitempty" yaml:"message,omitempty"`
		}

		var items []versionInfo
		for _, v := range versions {
			items = append(items, versionInfo{
				Version:   v.Version,
				AppliedAt: v.AppliedAt.Format("2006-01-02 15:04:05"),
				AppliedBy: v.AppliedBy,
				Message:   v.Message,
			})
		}

		switch deps.Format {
		case "json":
			output.JSON(os.Stdout, items)
		case "yaml":
			output.YAML(os.Stdout, items)
		default:
			headers := []string{"VERSION", "APPLIED AT", "BY", "MESSAGE"}
			var rows [][]string
			for _, v := range items {
				rows = append(rows, []string{
					fmt.Sprintf("%d", v.Version),
					v.AppliedAt,
					v.AppliedBy,
					v.Message,
				})
			}
			output.Table(os.Stdout, headers, rows)
		}
		return nil
	},
}
