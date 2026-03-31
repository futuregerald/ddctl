// cmd/dashboard/delete.go
package dashboard

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a dashboard from Datadog",
	Long:  `Deletes a dashboard from Datadog. Requires confirmation.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dashID := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		if !cmdutil.ConfirmOrSkip(cmd, fmt.Sprintf("Delete dashboard %s from Datadog?", dashID)) {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}

		if err := deps.Client.DeleteDashboard(dashID); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Deleted dashboard %s from Datadog.\n", dashID)
		return nil
	},
}
