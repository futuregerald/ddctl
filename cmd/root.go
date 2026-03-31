// cmd/root.go
package cmd

import (
	"github.com/futuregerald/ddctl/cmd/api"
	"github.com/futuregerald/ddctl/cmd/auth"
	"github.com/futuregerald/ddctl/cmd/connection"
	"github.com/futuregerald/ddctl/cmd/dashboard"
	"github.com/futuregerald/ddctl/cmd/db"
	"github.com/futuregerald/ddctl/cmd/logs"
	"github.com/futuregerald/ddctl/cmd/metrics"
	"github.com/futuregerald/ddctl/cmd/monitor"
	"github.com/futuregerald/ddctl/cmd/slo"
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
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagConnection, "connection", "c", "", "Connection profile to use")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format (json|table|yaml)")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug output to stderr")

	rootCmd.AddCommand(connection.Cmd)
	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(dashboard.Cmd)
	rootCmd.AddCommand(monitor.Cmd)
	rootCmd.AddCommand(slo.Cmd)
	rootCmd.AddCommand(metrics.Cmd)
	rootCmd.AddCommand(logs.Cmd)
	rootCmd.AddCommand(api.Cmd)
	rootCmd.AddCommand(db.Cmd)
}
