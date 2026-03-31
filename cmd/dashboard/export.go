// cmd/dashboard/export.go
package dashboard

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var exportFlagOutput string

var exportCmd = &cobra.Command{
	Use:   "export [id]",
	Short: "Export a dashboard from local store to a YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dashID := args[0]

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		version, err := deps.Store.GetLatestVersion(dashID, "dashboard", deps.ConnName)
		if err != nil {
			return err
		}

		if exportFlagOutput == "" {
			// Write to stdout
			fmt.Print(version.Content)
			return nil
		}

		if err := os.WriteFile(exportFlagOutput, []byte(version.Content), 0644); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Exported dashboard %s to %s\n", dashID, exportFlagOutput)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFlagOutput, "file", "o", "", "Output file path (defaults to stdout)")
}
