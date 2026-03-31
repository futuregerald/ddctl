// cmd/dashboard/sync.go
package dashboard

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Pull all tracked dashboards from Datadog",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		resources, err := deps.Store.ListResources("dashboard", deps.ConnName)
		if err != nil {
			return fmt.Errorf("listing tracked dashboards: %w", err)
		}

		if len(resources) == 0 {
			fmt.Fprintln(os.Stderr, "No tracked dashboards to sync.")
			return nil
		}

		var synced, failed int
		for _, r := range resources {
			jsonBytes, err := deps.Client.GetDashboard(r.ResourceID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to sync %s: %v\n", r.ResourceID, err)
				failed++
				continue
			}

			var raw interface{}
			json.Unmarshal(jsonBytes, &raw)
			yamlBytes, _ := yaml.Marshal(raw)

			var meta struct {
				Title string `json:"title"`
			}
			json.Unmarshal(jsonBytes, &meta)
			if meta.Title != "" {
				deps.Store.TrackResource(r.ResourceID, "dashboard", deps.ConnName, meta.Title)
			}

			deps.Store.SaveVersion(r.ResourceID, "dashboard", deps.ConnName, string(yamlBytes), "", "", "synced from remote")
			deps.Store.UpdateResourceSync(r.ResourceID, "dashboard", deps.ConnName, nil, "", "")
			synced++
		}

		fmt.Fprintf(os.Stderr, "Synced %d dashboards", synced)
		if failed > 0 {
			fmt.Fprintf(os.Stderr, " (%d failed)", failed)
		}
		fmt.Fprintln(os.Stderr)
		return nil
	},
}
