// cmd/dashboard/pull.go
package dashboard

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [id]",
	Short: "Pull a dashboard from Datadog into local store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dashID := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		// Fetch from Datadog
		jsonBytes, err := deps.Client.GetDashboard(dashID)
		if err != nil {
			return err
		}

		// Convert JSON to YAML for storage
		var raw interface{}
		if err := json.Unmarshal(jsonBytes, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
		yamlBytes, err := yaml.Marshal(raw)
		if err != nil {
			return fmt.Errorf("converting to YAML: %w", err)
		}

		// Extract title
		var meta struct {
			Title      string  `json:"title"`
			ModifiedAt *string `json:"modified_at"`
			AuthorName *string `json:"author_name"`
		}
		if err := json.Unmarshal(jsonBytes, &meta); err != nil {
			return fmt.Errorf("parsing metadata: %w", err)
		}

		// Track resource
		if err := deps.Store.TrackResource(dashID, "dashboard", deps.ConnName, meta.Title); err != nil {
			return fmt.Errorf("tracking resource: %w", err)
		}

		// Save version
		if err := deps.Store.SaveVersion(dashID, "dashboard", deps.ConnName, string(yamlBytes), "", "", "pulled from remote"); err != nil {
			return fmt.Errorf("saving version: %w", err)
		}

		// Update sync metadata
		deps.Store.UpdateResourceSync(dashID, "dashboard", deps.ConnName, nil, "", "")

		fmt.Fprintf(os.Stderr, "Pulled dashboard %s (%s)\n", dashID, meta.Title)
		return nil
	},
}
