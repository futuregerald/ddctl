// cmd/slo/pull.go
package slo

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
	Short: "Pull an SLO from Datadog into local store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sloID := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		jsonBytes, err := deps.Client.GetSLO(sloID)
		if err != nil {
			return err
		}

		var raw interface{}
		json.Unmarshal(jsonBytes, &raw)
		yamlBytes, _ := yaml.Marshal(raw)

		var meta struct {
			Name string `json:"name"`
		}
		json.Unmarshal(jsonBytes, &meta)

		if err := deps.Store.TrackResource(sloID, "slo", deps.ConnName, meta.Name); err != nil {
			return err
		}
		if err := deps.Store.SaveVersion(sloID, "slo", deps.ConnName, string(yamlBytes), "", "", "pulled from remote"); err != nil {
			return err
		}
		deps.Store.UpdateResourceSync(sloID, "slo", deps.ConnName, nil, "", "")

		fmt.Fprintf(os.Stderr, "Pulled SLO %s (%s)\n", sloID, meta.Name)
		return nil
	},
}
