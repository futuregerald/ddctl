// cmd/dashboard/list.go
package dashboard

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/spf13/cobra"
)

var listFlagRemote bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List dashboards (local or remote)",
	Long:  `List locally tracked dashboards, or all remote dashboards with --remote.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, listFlagRemote)
		if err != nil {
			return err
		}
		defer deps.Close()

		if listFlagRemote {
			dashboards, err := deps.Client.ListDashboards()
			if err != nil {
				return err
			}

			type remoteDash struct {
				ID          string `json:"id" yaml:"id"`
				Title       string `json:"title" yaml:"title"`
				Description string `json:"description,omitempty" yaml:"description,omitempty"`
				URL         string `json:"url,omitempty" yaml:"url,omitempty"`
			}

			var items []remoteDash
			for _, d := range dashboards {
				items = append(items, remoteDash{
					ID:          d.GetId(),
					Title:       d.GetTitle(),
					Description: d.GetDescription(),
					URL:         d.GetUrl(),
				})
			}

			switch deps.Format {
			case "json":
				output.JSON(os.Stdout, items)
			case "yaml":
				output.YAML(os.Stdout, items)
			default:
				headers := []string{"ID", "TITLE"}
				var rows [][]string
				for _, d := range items {
					rows = append(rows, []string{d.ID, d.Title})
				}
				output.Table(os.Stdout, headers, rows)
			}
			return nil
		}

		// Local list
		resources, err := deps.Store.ListResources("dashboard", deps.ConnName)
		if err != nil {
			return fmt.Errorf("listing local dashboards: %w", err)
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
	listCmd.Flags().BoolVar(&listFlagRemote, "remote", false, "List all remote dashboards from Datadog")
}
