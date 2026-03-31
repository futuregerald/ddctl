// cmd/monitor/rollback.go
package monitor

import (
	"fmt"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var rollbackFlagToVersion int

var rollbackCmd = &cobra.Command{
	Use:   "rollback [id]",
	Short: "Rollback to a previous version (local only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if rollbackFlagToVersion <= 0 {
			return fmt.Errorf("--to-version is required")
		}

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		return cmdutil.RollbackResource(deps, args[0], "monitor", rollbackFlagToVersion)
	},
}

func init() {
	rollbackCmd.Flags().IntVar(&rollbackFlagToVersion, "to-version", 0, "Version number to rollback to")
}
