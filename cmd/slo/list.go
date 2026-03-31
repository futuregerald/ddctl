// cmd/slo/list.go
package slo

import (
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/spf13/cobra"
)

var listFlagRemote bool

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List SLOs (local or remote)",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, listFlagRemote)
		if err != nil {
			return err
		}
		defer deps.Close()

		if listFlagRemote {
			slos, err := deps.Client.ListSLOs()
			if err != nil {
				return err
			}

			type remoteSLO struct {
				ID   string `json:"id" yaml:"id"`
				Name string `json:"name" yaml:"name"`
				Type string `json:"type" yaml:"type"`
			}

			var items []remoteSLO
			for _, s := range slos {
				items = append(items, remoteSLO{
					ID:   s.GetId(),
					Name: s.GetName(),
					Type: string(s.GetType()),
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
				for _, s := range items {
					rows = append(rows, []string{s.ID, s.Name, s.Type})
				}
				output.Table(os.Stdout, headers, rows)
			}
			return nil
		}

		resources, err := deps.Store.ListResources("slo", deps.ConnName)
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
	listCmd.Flags().BoolVar(&listFlagRemote, "remote", false, "List all remote SLOs from Datadog")
}
