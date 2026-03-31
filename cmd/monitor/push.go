// cmd/monitor/push.go
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

var pushCmd = &cobra.Command{
	Use:   "push [id]",
	Short: "Push local monitor state to Datadog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		monID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid monitor ID: %w", err)
		}
		resourceID := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		localVersion, err := deps.Store.GetLatestVersion(resourceID, "monitor", deps.ConnName)
		if err != nil {
			return err
		}

		// Fetch remote snapshot A
		snapshotA, err := deps.Client.GetMonitor(monID)
		if err != nil {
			return err
		}

		var meta struct {
			Name       string `json:"name"`
			ModifiedAt string `json:"modified"`
		}
		if err := json.Unmarshal(snapshotA, &meta); err != nil {
			return fmt.Errorf("parsing remote metadata: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Monitor: %s (%s)\n", resourceID, meta.Name)

		if !cmdutil.ConfirmOrSkip(cmd, "Push local changes to Datadog?") {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}

		// TOCTOU check
		snapshotB, err := deps.Client.GetMonitor(monID)
		if err != nil {
			return err
		}
		if string(snapshotA) != string(snapshotB) {
			return fmt.Errorf("remote state changed during confirmation. Pull first")
		}

		var localData interface{}
		if err := yaml.Unmarshal([]byte(localVersion.Content), &localData); err != nil {
			return fmt.Errorf("parsing local YAML: %w", err)
		}
		jsonBytes, _ := json.Marshal(localData)

		if err := deps.Client.UpdateMonitor(monID, jsonBytes); err != nil {
			return err
		}

		var rawRemote interface{}
		if err := json.Unmarshal(snapshotB, &rawRemote); err != nil {
			return fmt.Errorf("parsing remote snapshot: %w", err)
		}
		snapshotYAML, _ := yaml.Marshal(rawRemote)
		deps.Store.SaveVersion(resourceID, "monitor", deps.ConnName, localVersion.Content, string(snapshotYAML), "", "pushed to remote")

		fmt.Fprintf(os.Stderr, "Pushed monitor %s successfully.\n", resourceID)
		return nil
	},
}
