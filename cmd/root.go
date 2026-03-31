// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	flagConnection string
	flagOutput     string
	flagYes        bool
	flagDebug      bool
)

var rootCmd = &cobra.Command{
	Use:   "ddctl",
	Short: "Datadog control CLI",
	Long:  `ddctl — manage Datadog dashboards, monitors & SLOs from your terminal.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Will initialize config, client, etc. in later tasks
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagConnection, "connection", "c", "", "Connection profile to use")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format (json|table|yaml)")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug output to stderr")
}
