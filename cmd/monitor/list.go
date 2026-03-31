// cmd/monitor/list.go
package monitor

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/spf13/cobra"
)

var listFlagRemote bool

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List monitors (local or remote)",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, listFlagRemote)
		if err != nil {
			return err
		}
		defer deps.Close()

		if listFlagRemote {
			monitors, err := deps.Client.ListMonitors()
			if err != nil {
				return err
			}

			type remoteMon struct {
				ID   int64  `json:"id" yaml:"id"`
				Name string `json:"name" yaml:"name"`
				Type string `json:"type" yaml:"type"`
			}

			var items []remoteMon
			for _, m := range monitors {
				items = append(items, remoteMon{
					ID:   m.GetId(),
					Name: m.GetName(),
					Type: string(m.GetType()),
				})
			}

			switch deps.Format {
			case "json":
				output.JSON(os.Stdout, items)
			case "yaml":
				output.YAML(os.Stdout, items)
			default:
				headers := []string{"ID", "NAME", "TYPE"}
				var rows [][]string
				for _, m := range items {
					rows = append(rows, []string{fmt.Sprintf("%d", m.ID), m.Name, m.Type})
				}
				output.Table(os.Stdout, headers, rows)
			}
			return nil
		}

		resources, err := deps.Store.ListResources("monitor", deps.ConnName)
		if err != nil {
			return err
		}

		switch deps.Format {
		case "json":
			output.JSON(os.Stdout, resources)
		case "yaml":
			output.YAML(os.Stdout, resources)
		default:
			headers := []string{"ID", "TITLE", "STATUS", "LAST SYNCED"}
			var rows [][]string
			for _, r := range resources {
				synced := ""
				if r.LastSyncedAt != nil {
					synced = r.LastSyncedAt.Format("2006-01-02 15:04")
				}
				rows = append(rows, []string{r.ResourceID, r.Title, r.Status, synced})
			}
			output.Table(os.Stdout, headers, rows)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listFlagRemote, "remote", false, "List all remote monitors from Datadog")
}
