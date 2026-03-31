// cmd/slo/export.go
package slo

import (
	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var exportFlagOutput string

var exportCmd = &cobra.Command{
	Use:   "export [id]",
	Short: "Export an SLO from local store to a YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		return cmdutil.ExportResource(deps, args[0], "slo", exportFlagOutput)
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFlagOutput, "file", "o", "", "Output file path")
}
