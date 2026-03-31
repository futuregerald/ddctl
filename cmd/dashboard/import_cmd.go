// cmd/dashboard/import_cmd.go
package dashboard

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var (
	importFlagFile string
	importFlagPush bool
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a YAML file into local store",
	Long:  `Reads a dashboard YAML file and stores it in SQLite. Use --push to also push to Datadog.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if importFlagFile == "" {
			return fmt.Errorf("--file is required")
		}

		deps, err := cmdutil.InitDeps(cmd, importFlagPush)
		if err != nil {
			return err
		}
		defer deps.Close()

		// Read file
		data, err := os.ReadFile(importFlagFile)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		// Validate YAML
		var content interface{}
		if err := yaml.Unmarshal(data, &content); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}

		// Try to extract ID and title from the YAML content
		var meta struct {
			ID    string `yaml:"id"`
			Title string `yaml:"title"`
		}
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return fmt.Errorf("parsing metadata: %w", err)
		}

		dashID := meta.ID
		title := meta.Title
		if title == "" {
			title = cmdutil.StripExtension(importFlagFile)
		}

		if importFlagPush {
			// Create on Datadog if no ID
			jsonBytes, err := json.Marshal(content)
			if err != nil {
				return fmt.Errorf("converting to JSON: %w", err)
			}

			if dashID == "" {
				newID, err := deps.Client.CreateDashboard(jsonBytes)
				if err != nil {
					return fmt.Errorf("creating dashboard: %w", err)
				}
				dashID = newID
				fmt.Fprintf(os.Stderr, "Created dashboard %s on Datadog.\n", dashID)
			} else {
				if err := deps.Client.UpdateDashboard(dashID, jsonBytes); err != nil {
					return fmt.Errorf("updating dashboard: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Updated dashboard %s on Datadog.\n", dashID)
			}
		}

		if dashID == "" {
			dashID = fmt.Sprintf("local-%s", cmdutil.SHA256(data)[:12])
		}

		// Track and save
		if err := deps.Store.TrackResource(dashID, "dashboard", deps.ConnName, title); err != nil {
			return fmt.Errorf("tracking resource: %w", err)
		}
		if err := deps.Store.SaveVersion(dashID, "dashboard", deps.ConnName, string(data), "", "", "imported from file"); err != nil {
			return fmt.Errorf("saving version: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Imported dashboard %s (%s)\n", dashID, title)
		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importFlagFile, "file", "f", "", "YAML file to import")
	importCmd.Flags().BoolVar(&importFlagPush, "push", false, "Also push to Datadog after importing")
}
