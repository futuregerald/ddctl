// cmd/dashboard/export.go
package dashboard

import (
	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var exportFlagOutput string

var exportCmd = &cobra.Command{
	Use:   "export [id]",
	Short: "Export a dashboard from local store to a YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		return cmdutil.ExportResource(deps, args[0], "dashboard", exportFlagOutput)
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFlagOutput, "file", "o", "", "Output file path (defaults to stdout)")
}
