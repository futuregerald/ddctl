// cmd/monitor/rollback.go
package monitor

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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		monID := args[0]

		if rollbackFlagToVersion <= 0 {
			return fmt.Errorf("--to-version is required")
		}

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		targetVersion, err := deps.Store.GetVersion(monID, "monitor", deps.ConnName, rollbackFlagToVersion)
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("rollback to version %d", rollbackFlagToVersion)
		if err := deps.Store.SaveVersion(monID, "monitor", deps.ConnName, targetVersion.Content, "", "", msg); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Rolled back to version %d. Run 'ddctl monitor push %s' to apply remotely.\n", rollbackFlagToVersion, monID)
		return nil
	},
}

func init() {
	rollbackCmd.Flags().IntVar(&rollbackFlagToVersion, "to-version", 0, "Version number to rollback to")
}
