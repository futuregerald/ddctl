// cmd/monitor/create.go
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
	createFlagName  string
	createFlagType  string
	createFlagQuery string
	createFlagFile  string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new monitor on Datadog",
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
			if createFlagName == "" || createFlagType == "" || createFlagQuery == "" {
				return fmt.Errorf("--name, --type, and --query are required (or use --file)")
			}
			mon := map[string]interface{}{
				"name":    createFlagName,
				"type":    createFlagType,
				"query":   createFlagQuery,
				"message": "",
			}
			body, _ = json.Marshal(mon)
		}

		newID, err := deps.Client.CreateMonitor(body)
		if err != nil {
			return err
		}

		resourceID := fmt.Sprintf("%d", newID)

		// Pull back
		jsonBytes, err := deps.Client.GetMonitor(newID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Created monitor %s but failed to pull: %v\n", resourceID, err)
			return nil
		}

		var raw interface{}
		if err := json.Unmarshal(jsonBytes, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
		yamlBytes, _ := yaml.Marshal(raw)

		name := createFlagName
		var meta struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(jsonBytes, &meta); err != nil {
			return fmt.Errorf("parsing metadata: %w", err)
		}
		if meta.Name != "" {
			name = meta.Name
		}

		deps.Store.TrackResource(resourceID, "monitor", deps.ConnName, name)
		deps.Store.SaveVersion(resourceID, "monitor", deps.ConnName, string(yamlBytes), "", "", "created")

		fmt.Fprintf(os.Stderr, "Created monitor %s (%s)\n", resourceID, name)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createFlagName, "name", "", "Monitor name")
	createCmd.Flags().StringVar(&createFlagType, "type", "", "Monitor type")
	createCmd.Flags().StringVar(&createFlagQuery, "query", "", "Monitor query")
	createCmd.Flags().StringVarP(&createFlagFile, "file", "f", "", "Create from YAML file")
}
