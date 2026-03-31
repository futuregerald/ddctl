// cmd/auth/auth.go
package auth

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for authentication.
var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
	Long:  `Login, check status, and logout from Datadog connections.`,
}

func init() {
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(logoutCmd)
}
