// cmd/slo/push.go
package slo

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [id]",
	Short: "Push local SLO state to Datadog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sloID := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		localVersion, err := deps.Store.GetLatestVersion(sloID, "slo", deps.ConnName)
		if err != nil {
			return err
		}

		snapshotA, err := deps.Client.GetSLO(sloID)
		if err != nil {
			return err
		}

		var meta struct {
			Name string `json:"name"`
		}
		json.Unmarshal(snapshotA, &meta)

		fmt.Fprintf(os.Stderr, "SLO: %s (%s)\n", sloID, meta.Name)

		if !cmdutil.ConfirmOrSkip(cmd, "Push local changes to Datadog?") {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}

		snapshotB, err := deps.Client.GetSLO(sloID)
		if err != nil {
			return err
		}
		if string(snapshotA) != string(snapshotB) {
			return fmt.Errorf("remote state changed during confirmation. Pull first")
		}

		var localData interface{}
		yaml.Unmarshal([]byte(localVersion.Content), &localData)
		jsonBytes, _ := json.Marshal(localData)

		if err := deps.Client.UpdateSLO(sloID, jsonBytes); err != nil {
			return err
		}

		var rawRemote interface{}
		json.Unmarshal(snapshotB, &rawRemote)
		snapshotYAML, _ := yaml.Marshal(rawRemote)
		deps.Store.SaveVersion(sloID, "slo", deps.ConnName, localVersion.Content, string(snapshotYAML), "", "pushed to remote")

		fmt.Fprintf(os.Stderr, "Pushed SLO %s successfully.\n", sloID)
		return nil
	},
}
