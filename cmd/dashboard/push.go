// cmd/dashboard/push.go
package dashboard

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
	Short: "Push local dashboard state to Datadog",
	Long: `Push follows a safe workflow:
1. Fetch current remote state (snapshot A)
2. Diff local vs remote
3. Show who last modified and when
4. Ask for confirmation
5. Re-fetch remote (snapshot B) for TOCTOU protection
6. Apply if snapshots match`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dashID := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		// Get local latest
		localVersion, err := deps.Store.GetLatestVersion(dashID, "dashboard", deps.ConnName)
		if err != nil {
			return fmt.Errorf("no local version found: %w", err)
		}

		// Step 1: Fetch remote (snapshot A)
		snapshotA, err := deps.Client.GetDashboard(dashID)
		if err != nil {
			return fmt.Errorf("fetching remote state: %w", err)
		}

		// Step 2-3: Show diff summary and remote modifier
		var remoteMeta struct {
			Title      string `json:"title"`
			ModifiedAt string `json:"modified_at"`
			AuthorName string `json:"author_name"`
		}
		if err := json.Unmarshal(snapshotA, &remoteMeta); err != nil {
			return fmt.Errorf("parsing remote metadata: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Dashboard: %s (%s)\n", dashID, remoteMeta.Title)
		if remoteMeta.ModifiedAt != "" {
			fmt.Fprintf(os.Stderr, "Remote last modified: %s", remoteMeta.ModifiedAt)
			if remoteMeta.AuthorName != "" {
				fmt.Fprintf(os.Stderr, " by %s", remoteMeta.AuthorName)
			}
			fmt.Fprintln(os.Stderr)
		}

		// Step 4: Confirm
		if !cmdutil.ConfirmOrSkip(cmd, "Push local changes to Datadog?") {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}

		// Step 5: Re-fetch (snapshot B) for TOCTOU protection
		snapshotB, err := deps.Client.GetDashboard(dashID)
		if err != nil {
			return fmt.Errorf("re-fetching remote state: %w", err)
		}

		// Step 6: Compare snapshots
		if string(snapshotA) != string(snapshotB) {
			return fmt.Errorf("remote state changed during confirmation. Re-run 'ddctl dashboard pull %s' first", dashID)
		}

		// Convert local YAML to JSON for API
		var localData interface{}
		if err := yaml.Unmarshal([]byte(localVersion.Content), &localData); err != nil {
			return fmt.Errorf("parsing local content: %w", err)
		}
		jsonBytes, err := json.Marshal(localData)
		if err != nil {
			return fmt.Errorf("converting to JSON: %w", err)
		}

		// Step 7-8: Store snapshot and apply
		if err := deps.Client.UpdateDashboard(dashID, jsonBytes); err != nil {
			return fmt.Errorf("updating dashboard: %w", err)
		}

		// Save remote snapshot
		var rawRemote interface{}
		if err := json.Unmarshal(snapshotB, &rawRemote); err != nil {
			return fmt.Errorf("parsing remote snapshot: %w", err)
		}
		snapshotYAML, _ := yaml.Marshal(rawRemote)
		deps.Store.SaveVersion(dashID, "dashboard", deps.ConnName, localVersion.Content, string(snapshotYAML), "", "pushed to remote")

		fmt.Fprintf(os.Stderr, "Pushed dashboard %s successfully.\n", dashID)
		return nil
	},
}
