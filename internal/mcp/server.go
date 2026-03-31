// internal/mcp/server.go
// MCP server for ddctl — exposes Datadog operations as MCP tools for LLM agents.
package mcp

import (
	"context"
	"fmt"
	"log"
	"os"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/futuregerald/ddctl/internal/client"
	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/futuregerald/ddctl/internal/store"
)

// SafetyLevel controls which tools are exposed in MCP mode.
type SafetyLevel string

const (
	SafetyReadOnly    SafetyLevel = "read-only"
	SafetyReadWrite   SafetyLevel = "read-write"
	SafetyUnrestricted SafetyLevel = "unrestricted"
)

// Server wraps the MCP server with ddctl context.
type Server struct {
	mcpServer *server.MCPServer
	store     *store.Store
	client    *client.Client
	connName  string
	safety    SafetyLevel
}

// NewServer creates a new MCP server with the given safety level, connection, and version.
func NewServer(safety SafetyLevel, connName string, version string) (*Server, error) {
	cfg, err := config.Load(config.Path())
	if err != nil {
		cfg = config.DefaultConfig()
	}

	s, err := store.New(config.DBPath())
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}

	// Resolve connection name
	if connName == "" && cfg.DefaultConnection != "" {
		connName = cfg.DefaultConnection
	}
	if connName == "" {
		if dc, err := s.GetDefaultConnection(); err == nil {
			connName = dc.Name
		}
	}

	var c *client.Client
	if connName != "" {
		conn, err := s.GetConnection(connName)
		if err == nil {
			creds, _, err := keyring.ResolveCredentials(connName)
			if err == nil {
				c, err = client.New(conn.Site, creds)
				if err != nil {
					log.Printf("warning: could not create API client: %v", err)
				}
			} else {
				log.Printf("warning: could not resolve credentials: %v", err)
			}
		} else {
			log.Printf("warning: connection %q not found: %v", connName, err)
		}
	}

	mcpSrv := server.NewMCPServer(
		"ddctl",
		version,
		server.WithInstructions("ddctl is a Datadog control CLI. Use these tools to manage Datadog dashboards, monitors, SLOs, query metrics, search logs, and make arbitrary API calls. All operations respect the configured safety level."),
	)

	srv := &Server{
		mcpServer: mcpSrv,
		store:     s,
		client:    c,
		connName:  connName,
		safety:    safety,
	}

	srv.registerTools()

	return srv, nil
}

// Serve starts the MCP server on stdio.
func (s *Server) Serve() error {
	stdio := server.NewStdioServer(s.mcpServer)
	stdio.SetErrorLogger(log.New(os.Stderr, "ddctl-mcp: ", log.LstdFlags))
	return stdio.Listen(context.Background(), os.Stdin, os.Stdout)
}

// Close cleans up resources.
func (s *Server) Close() {
	if s.store != nil {
		s.store.Close()
	}
}

// isAllowed checks whether an operation category is allowed under the current safety level.
func (s *Server) isAllowed(category string) bool {
	switch s.safety {
	case SafetyReadOnly:
		return category == "read"
	case SafetyReadWrite:
		return category == "read" || category == "write"
	case SafetyUnrestricted:
		return true
	default:
		return category == "read"
	}
}

// requireClient returns the client or an error result if not configured.
func (s *Server) requireClient() (*client.Client, error) {
	if s.client == nil {
		return nil, fmt.Errorf("no API client configured. Run 'ddctl auth login' or provide DD_API_KEY/DD_APP_KEY environment variables")
	}
	return s.client, nil
}

// requireConnection returns the connection name or an error.
func (s *Server) requireConnection() (string, error) {
	if s.connName == "" {
		return "", fmt.Errorf("no connection configured. Run 'ddctl connection add' first")
	}
	return s.connName, nil
}

// textResult creates a successful text result.
func textResult(text string) *mcplib.CallToolResult {
	return &mcplib.CallToolResult{
		Content: []mcplib.Content{
			mcplib.NewTextContent(text),
		},
	}
}

// errorResult creates an error result.
func errorResult(msg string) *mcplib.CallToolResult {
	return &mcplib.CallToolResult{
		Content: []mcplib.Content{
			mcplib.NewTextContent(msg),
		},
		IsError: true,
	}
}
