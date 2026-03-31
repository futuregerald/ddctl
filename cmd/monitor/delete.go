// cmd/monitor/delete.go
package monitor

import (
	"fmt"
	"os"
	"strconv"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a monitor from Datadog",
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

		if !cmdutil.ConfirmOrSkip(cmd, fmt.Sprintf("Delete monitor %d from Datadog?", monID)) {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}

		if err := deps.Client.DeleteMonitor(monID); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Deleted monitor %d from Datadog.\n", monID)
		return nil
	},
}
