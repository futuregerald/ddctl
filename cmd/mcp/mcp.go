// cmd/mcp/mcp.go
// MCP server command — starts ddctl as an MCP server on stdio.
package mcp

import (
	"fmt"

	mcpserver "github.com/futuregerald/ddctl/internal/mcp"
	"github.com/spf13/cobra"
)

// Version is set by the cmd package at init time to avoid import cycles.
var Version = "dev"

var (
	flagSafety     string
	flagConnection string
)

// Cmd is the parent command for MCP operations.
var Cmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server mode for LLM integration",
	Long:  `Run ddctl as an MCP (Model Context Protocol) server for LLM agent integration.`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server on stdio",
	Long: `Start ddctl as an MCP server using stdio transport.

The server exposes Datadog operations as MCP tools for LLM agents.
Configure the safety level to control which operations are available:

  read-only (default)  — list, search, query, diff, history
  read-write           — reads + create, push, pull, edit
  unrestricted         — all operations (delete/rollback require confirm=true)

Example MCP configuration:

  {
    "mcpServers": {
      "ddctl": {
        "command": "ddctl",
        "args": ["mcp", "serve", "--safety", "read-write"]
      }
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		safety := mcpserver.SafetyLevel(flagSafety)
		switch safety {
		case mcpserver.SafetyReadOnly, mcpserver.SafetyReadWrite, mcpserver.SafetyUnrestricted:
			// valid
		default:
			return fmt.Errorf("invalid safety level %q: must be read-only, read-write, or unrestricted", flagSafety)
		}

		// Use the --connection flag from either local or root command
		connName := flagConnection
		if connName == "" {
			if f := cmd.Root().PersistentFlags().Lookup("connection"); f != nil && f.Changed {
				connName = f.Value.String()
			}
		}

		srv, err := mcpserver.NewServer(safety, connName, Version)
		if err != nil {
			return fmt.Errorf("starting MCP server: %w", err)
		}
		defer srv.Close()

		return srv.Serve()
	},
}

func init() {
	serveCmd.Flags().StringVar(&flagSafety, "safety", "read-only", "Safety level: read-only, read-write, or unrestricted")
	serveCmd.Flags().StringVar(&flagConnection, "connection", "", "Connection profile to use (overrides default)")
	Cmd.AddCommand(serveCmd)
}
