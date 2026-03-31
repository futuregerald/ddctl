// cmd/slo/delete.go
package slo

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an SLO from Datadog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sloID := args[0]

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		if !cmdutil.ConfirmOrSkip(cmd, fmt.Sprintf("Delete SLO %s from Datadog?", sloID)) {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}

		if err := deps.Client.DeleteSLO(sloID); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Deleted SLO %s from Datadog.\n", sloID)
		return nil
	},
}
