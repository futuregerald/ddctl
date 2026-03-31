// cmd/dashboard/rollback.go
package dashboard

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var rollbackFlagToVersion int

var rollbackCmd = &cobra.Command{
	Use:   "rollback [id]",
	Short: "Rollback to a previous version (local only)",
	Long:  `Copies the content of a previous version as a new version. Use 'push' to apply to Datadog.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dashID := args[0]

		if rollbackFlagToVersion <= 0 {
			return fmt.Errorf("--to-version is required")
		}

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		// Get the target version
		targetVersion, err := deps.Store.GetVersion(dashID, "dashboard", deps.ConnName, rollbackFlagToVersion)
		if err != nil {
			return err
		}

		// Save as new version
		msg := fmt.Sprintf("rollback to version %d", rollbackFlagToVersion)
		if err := deps.Store.SaveVersion(dashID, "dashboard", deps.ConnName, targetVersion.Content, "", "", msg); err != nil {
			return fmt.Errorf("saving rollback version: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Rolled back to version %d. Run 'ddctl dashboard push %s' to apply remotely.\n", rollbackFlagToVersion, dashID)
		return nil
	},
}

func init() {
	rollbackCmd.Flags().IntVar(&rollbackFlagToVersion, "to-version", 0, "Version number to rollback to")
}
