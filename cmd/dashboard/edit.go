// cmd/dashboard/edit.go
package dashboard

import (
	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit a dashboard in your editor",
	Long: `Opens the dashboard YAML in your configured editor.
On save, validates the YAML and stores as a new local version.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		return cmdutil.EditResource(deps, args[0], "dashboard")
	},
}
