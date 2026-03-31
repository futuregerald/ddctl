// cmd/monitor/import_cmd.go
package monitor

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
	RunE: func(cmd *cobra.Command, args []string) error {
		if importFlagFile == "" {
			return fmt.Errorf("--file is required")
		}

		deps, err := cmdutil.InitDeps(cmd, importFlagPush)
		if err != nil {
			return err
		}
		defer deps.Close()

		data, err := os.ReadFile(importFlagFile)
		if err != nil {
			return err
		}

		var content interface{}
		if err := yaml.Unmarshal(data, &content); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}

		var meta struct {
			ID   int64  `yaml:"id"`
			Name string `yaml:"name"`
		}
		yaml.Unmarshal(data, &meta)

		resourceID := fmt.Sprintf("%d", meta.ID)
		title := meta.Name
		if title == "" {
			title = cmdutil.StripExtension(importFlagFile)
		}

		if importFlagPush {
			jsonBytes, _ := json.Marshal(content)
			if meta.ID == 0 {
				newID, err := deps.Client.CreateMonitor(jsonBytes)
				if err != nil {
					return err
				}
				resourceID = fmt.Sprintf("%d", newID)
				fmt.Fprintf(os.Stderr, "Created monitor %s on Datadog.\n", resourceID)
			} else {
				if err := deps.Client.UpdateMonitor(meta.ID, jsonBytes); err != nil {
					return err
				}
			}
		}

		if resourceID == "0" {
			resourceID = fmt.Sprintf("local-%s", cmdutil.SHA256(data)[:12])
		}

		deps.Store.TrackResource(resourceID, "monitor", deps.ConnName, title)
		deps.Store.SaveVersion(resourceID, "monitor", deps.ConnName, string(data), "", "", "imported from file")

		fmt.Fprintf(os.Stderr, "Imported monitor %s (%s)\n", resourceID, title)
		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importFlagFile, "file", "f", "", "YAML file to import")
	importCmd.Flags().BoolVar(&importFlagPush, "push", false, "Also push to Datadog after importing")
}
