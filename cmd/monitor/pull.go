// cmd/monitor/pull.go
package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [id]",
	Short: "Pull a monitor from Datadog into local store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		monID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid monitor ID: %w", err)
		}

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		jsonBytes, err := deps.Client.GetMonitor(monID)
		if err != nil {
			return err
		}

		var raw interface{}
		if err := json.Unmarshal(jsonBytes, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
		yamlBytes, _ := yaml.Marshal(raw)

		var meta struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(jsonBytes, &meta); err != nil {
			return fmt.Errorf("parsing metadata: %w", err)
		}

		resourceID := args[0]
		if err := deps.Store.TrackResource(resourceID, "monitor", deps.ConnName, meta.Name); err != nil {
			return err
		}
		if err := deps.Store.SaveVersion(resourceID, "monitor", deps.ConnName, string(yamlBytes), "", "", "pulled from remote"); err != nil {
			return err
		}
		deps.Store.UpdateResourceSync(resourceID, "monitor", deps.ConnName, nil, "", "")

		fmt.Fprintf(os.Stderr, "Pulled monitor %s (%s)\n", resourceID, meta.Name)
		return nil
	},
}
