// cmd/slo/create.go
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
	createFlagName   string
	createFlagType   string
	createFlagTarget float64
	createFlagFile   string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new SLO on Datadog",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		var body []byte

		if createFlagFile != "" {
			data, err := os.ReadFile(createFlagFile)
			if err != nil {
				return err
			}
			var yamlContent interface{}
			if err := yaml.Unmarshal(data, &yamlContent); err != nil {
				return fmt.Errorf("invalid YAML: %w", err)
			}
			body, _ = json.Marshal(yamlContent)
		} else {
			if createFlagName == "" || createFlagType == "" {
				return fmt.Errorf("--name and --type are required (or use --file)")
			}
			slo := map[string]interface{}{
				"name":      createFlagName,
				"type":      createFlagType,
				"thresholds": []map[string]interface{}{
					{"timeframe": "30d", "target": createFlagTarget},
				},
			}
			body, _ = json.Marshal(slo)
		}

		newID, err := deps.Client.CreateSLO(body)
		if err != nil {
			return err
		}

		// Pull back
		jsonBytes, err := deps.Client.GetSLO(newID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Created SLO %s but failed to pull: %v\n", newID, err)
			return nil
		}

		var raw interface{}
		json.Unmarshal(jsonBytes, &raw)
		yamlBytes, _ := yaml.Marshal(raw)

		name := createFlagName
		var meta struct {
			Name string `json:"name"`
		}
		json.Unmarshal(jsonBytes, &meta)
		if meta.Name != "" {
			name = meta.Name
		}

		deps.Store.TrackResource(newID, "slo", deps.ConnName, name)
		deps.Store.SaveVersion(newID, "slo", deps.ConnName, string(yamlBytes), "", "", "created")

		fmt.Fprintf(os.Stderr, "Created SLO %s (%s)\n", newID, name)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createFlagName, "name", "", "SLO name")
	createCmd.Flags().StringVar(&createFlagType, "type", "", "SLO type (metric|monitor)")
	createCmd.Flags().Float64Var(&createFlagTarget, "target", 99.9, "SLO target percentage")
	createCmd.Flags().StringVarP(&createFlagFile, "file", "f", "", "Create from YAML file")
}
