// cmd/dashboard/create.go
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
	createFlagTitle       string
	createFlagDescription string
	createFlagLayout      string
	createFlagFile        string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new dashboard on Datadog",
	Long:  `Create a new dashboard from flags or a YAML file, then pull it back to local store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		var body []byte

		if createFlagFile != "" {
			// Read from file
			data, err := os.ReadFile(createFlagFile)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			var yamlContent interface{}
			if err := yaml.Unmarshal(data, &yamlContent); err != nil {
				return fmt.Errorf("invalid YAML: %w", err)
			}
			body, err = json.Marshal(yamlContent)
			if err != nil {
				return fmt.Errorf("converting to JSON: %w", err)
			}
		} else {
			if createFlagTitle == "" {
				return fmt.Errorf("--title is required (or use --file)")
			}
			layout := createFlagLayout
			if layout == "" {
				layout = "ordered"
			}

			dash := map[string]interface{}{
				"title":       createFlagTitle,
				"description": createFlagDescription,
				"layout_type": layout,
				"widgets":     []interface{}{},
			}
			body, _ = json.Marshal(dash)
		}

		newID, err := deps.Client.CreateDashboard(body)
		if err != nil {
			return err
		}

		// Pull back to get full state
		jsonBytes, err := deps.Client.GetDashboard(newID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Created dashboard %s but failed to pull: %v\n", newID, err)
			return nil
		}

		var raw interface{}
		json.Unmarshal(jsonBytes, &raw)
		yamlBytes, _ := yaml.Marshal(raw)

		title := createFlagTitle
		var meta struct {
			Title string `json:"title"`
		}
		json.Unmarshal(jsonBytes, &meta)
		if meta.Title != "" {
			title = meta.Title
		}

		deps.Store.TrackResource(newID, "dashboard", deps.ConnName, title)
		deps.Store.SaveVersion(newID, "dashboard", deps.ConnName, string(yamlBytes), "", "", "created")

		fmt.Fprintf(os.Stderr, "Created dashboard %s (%s)\n", newID, title)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createFlagTitle, "title", "", "Dashboard title")
	createCmd.Flags().StringVar(&createFlagDescription, "description", "", "Dashboard description")
	createCmd.Flags().StringVar(&createFlagLayout, "layout", "ordered", "Layout type (ordered|free)")
	createCmd.Flags().StringVarP(&createFlagFile, "file", "f", "", "Create from YAML file")
}
