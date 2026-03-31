// cmd/slo/import_cmd.go
package slo

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
			ID   string `yaml:"id"`
			Name string `yaml:"name"`
		}
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return fmt.Errorf("parsing metadata: %w", err)
		}

		sloID := meta.ID
		title := meta.Name
		if title == "" {
			title = cmdutil.StripExtension(importFlagFile)
		}

		if importFlagPush {
			jsonBytes, _ := json.Marshal(content)
			if sloID == "" {
				newID, err := deps.Client.CreateSLO(jsonBytes)
				if err != nil {
					return err
				}
				sloID = newID
				fmt.Fprintf(os.Stderr, "Created SLO %s on Datadog.\n", sloID)
			} else {
				if err := deps.Client.UpdateSLO(sloID, jsonBytes); err != nil {
					return err
				}
			}
		}

		if sloID == "" {
			sloID = fmt.Sprintf("local-%s", cmdutil.SHA256(data)[:12])
		}

		deps.Store.TrackResource(sloID, "slo", deps.ConnName, title)
		deps.Store.SaveVersion(sloID, "slo", deps.ConnName, string(data), "", "", "imported from file")

		fmt.Fprintf(os.Stderr, "Imported SLO %s (%s)\n", sloID, title)
		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importFlagFile, "file", "f", "", "YAML file to import")
	importCmd.Flags().BoolVar(&importFlagPush, "push", false, "Also push to Datadog after importing")
}
