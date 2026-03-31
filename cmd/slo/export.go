// cmd/slo/export.go
package slo

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var exportFlagOutput string

var exportCmd = &cobra.Command{
	Use:   "export [id]",
	Short: "Export an SLO from local store to a YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sloID := args[0]

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		version, err := deps.Store.GetLatestVersion(sloID, "slo", deps.ConnName)
		if err != nil {
			return err
		}

		if exportFlagOutput == "" {
			fmt.Print(version.Content)
			return nil
		}

		if err := os.WriteFile(exportFlagOutput, []byte(version.Content), 0644); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Exported SLO %s to %s\n", sloID, exportFlagOutput)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFlagOutput, "file", "o", "", "Output file path")
}
