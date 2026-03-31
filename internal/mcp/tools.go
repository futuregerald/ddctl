// internal/mcp/tools.go
// MCP tool definitions mapping CLI commands to MCP tools.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// registerTools adds all tools to the MCP server based on the safety level.
func (s *Server) registerTools() {
	// Read tools — always available
	s.mcpServer.AddTool(dashboardListTool(), s.handleDashboardList)
	s.mcpServer.AddTool(dashboardListRemoteTool(), s.handleDashboardListRemote)
	s.mcpServer.AddTool(dashboardDiffTool(), s.handleDashboardDiff)
	s.mcpServer.AddTool(dashboardHistoryTool(), s.handleDashboardHistory)
	s.mcpServer.AddTool(monitorListTool(), s.handleMonitorList)
	s.mcpServer.AddTool(monitorListRemoteTool(), s.handleMonitorListRemote)
	s.mcpServer.AddTool(sloListTool(), s.handleSLOList)
	s.mcpServer.AddTool(sloListRemoteTool(), s.handleSLOListRemote)
	s.mcpServer.AddTool(metricsQueryTool(), s.handleMetricsQuery)
	s.mcpServer.AddTool(metricsSearchTool(), s.handleMetricsSearch)
	s.mcpServer.AddTool(logsSearchTool(), s.handleLogsSearch)

	// Write tools — read-write and unrestricted
	if s.isAllowed("write") {
		s.mcpServer.AddTool(dashboardPullTool(), s.handleDashboardPull)
		s.mcpServer.AddTool(dashboardPushTool(), s.handleDashboardPush)
		s.mcpServer.AddTool(dashboardCreateTool(), s.handleDashboardCreate)
		s.mcpServer.AddTool(dashboardEditTool(), s.handleDashboardEdit)
		s.mcpServer.AddTool(monitorPullTool(), s.handleMonitorPull)
		s.mcpServer.AddTool(monitorPushTool(), s.handleMonitorPush)
		s.mcpServer.AddTool(monitorCreateTool(), s.handleMonitorCreate)
		s.mcpServer.AddTool(sloPullTool(), s.handleSLOPull)
		s.mcpServer.AddTool(sloPushTool(), s.handleSLOPush)
		s.mcpServer.AddTool(sloCreateTool(), s.handleSLOCreate)
		s.mcpServer.AddTool(apiCallTool(), s.handleAPICall)
	}

	// Destructive tools — unrestricted only
	if s.isAllowed("destructive") {
		s.mcpServer.AddTool(dashboardDeleteTool(), s.handleDashboardDelete)
		s.mcpServer.AddTool(dashboardRollbackTool(), s.handleDashboardRollback)
		s.mcpServer.AddTool(monitorDeleteTool(), s.handleMonitorDelete)
		s.mcpServer.AddTool(sloDeleteTool(), s.handleSLODelete)
	}
}

// ─── Tool Definitions ───────────────────────────────────────────────────────

func dashboardListTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_list",
		mcplib.WithDescription("List locally tracked dashboards. Returns dashboards that have been pulled or imported into the local store."),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
	)
}

func dashboardListRemoteTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_list_remote",
		mcplib.WithDescription("List all dashboards from Datadog. Returns all dashboards visible to the configured API key, including ID, title, and URL."),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func dashboardPullTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_pull",
		mcplib.WithDescription("Pull a dashboard from Datadog into the local store. Creates a new local version with the remote state."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog dashboard ID (e.g., 'abc-def-123')")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func dashboardPushTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_push",
		mcplib.WithDescription("Push a locally tracked dashboard to Datadog. Updates the remote dashboard with the latest local version. Requires confirm=true."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog dashboard ID")),
		mcplib.WithBoolean("confirm", mcplib.Description("Must be true to confirm the push operation")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func dashboardCreateTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_create",
		mcplib.WithDescription("Create a new dashboard in Datadog from a YAML definition. The YAML should match the Datadog API dashboard format."),
		mcplib.WithString("yaml_content", mcplib.Required(), mcplib.Description("The dashboard definition in YAML format (Datadog API format)")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func dashboardEditTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_edit",
		mcplib.WithDescription("Update a locally tracked dashboard's content. Saves the new YAML as a new local version. Use dashboard_push to apply remotely."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog dashboard ID")),
		mcplib.WithString("yaml_content", mcplib.Required(), mcplib.Description("The updated dashboard YAML content")),
		mcplib.WithDestructiveHintAnnotation(false),
	)
}

func dashboardDiffTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_diff",
		mcplib.WithDescription("Show differences between the local version and remote version of a dashboard."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog dashboard ID")),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
	)
}

func dashboardHistoryTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_history",
		mcplib.WithDescription("Show version history of a locally tracked dashboard."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog dashboard ID")),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
	)
}

func dashboardRollbackTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_rollback",
		mcplib.WithDescription("Rollback a dashboard to a previous local version. Creates a new version with the content from the specified version. Requires confirm=true."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog dashboard ID")),
		mcplib.WithNumber("to_version", mcplib.Required(), mcplib.Description("The version number to rollback to")),
		mcplib.WithBoolean("confirm", mcplib.Description("Must be true to confirm the rollback")),
		mcplib.WithDestructiveHintAnnotation(true),
	)
}

func dashboardDeleteTool() mcplib.Tool {
	return mcplib.NewTool("dashboard_delete",
		mcplib.WithDescription("Delete a dashboard from Datadog. This is irreversible. Requires confirm=true."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog dashboard ID")),
		mcplib.WithBoolean("confirm", mcplib.Description("Must be true to confirm the deletion")),
		mcplib.WithDestructiveHintAnnotation(true),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func monitorListTool() mcplib.Tool {
	return mcplib.NewTool("monitor_list",
		mcplib.WithDescription("List locally tracked monitors."),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
	)
}

func monitorListRemoteTool() mcplib.Tool {
	return mcplib.NewTool("monitor_list_remote",
		mcplib.WithDescription("List all monitors from Datadog."),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func monitorPullTool() mcplib.Tool {
	return mcplib.NewTool("monitor_pull",
		mcplib.WithDescription("Pull a monitor from Datadog into the local store."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog monitor ID (numeric)")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func monitorPushTool() mcplib.Tool {
	return mcplib.NewTool("monitor_push",
		mcplib.WithDescription("Push a locally tracked monitor to Datadog. Requires confirm=true."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog monitor ID (numeric)")),
		mcplib.WithBoolean("confirm", mcplib.Description("Must be true to confirm the push")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func monitorCreateTool() mcplib.Tool {
	return mcplib.NewTool("monitor_create",
		mcplib.WithDescription("Create a new monitor in Datadog from a YAML definition."),
		mcplib.WithString("yaml_content", mcplib.Required(), mcplib.Description("The monitor definition in YAML format (Datadog API format)")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func monitorDeleteTool() mcplib.Tool {
	return mcplib.NewTool("monitor_delete",
		mcplib.WithDescription("Delete a monitor from Datadog. Irreversible. Requires confirm=true."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog monitor ID (numeric)")),
		mcplib.WithBoolean("confirm", mcplib.Description("Must be true to confirm the deletion")),
		mcplib.WithDestructiveHintAnnotation(true),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func sloListTool() mcplib.Tool {
	return mcplib.NewTool("slo_list",
		mcplib.WithDescription("List locally tracked SLOs."),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
	)
}

func sloListRemoteTool() mcplib.Tool {
	return mcplib.NewTool("slo_list_remote",
		mcplib.WithDescription("List all SLOs from Datadog."),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func sloPullTool() mcplib.Tool {
	return mcplib.NewTool("slo_pull",
		mcplib.WithDescription("Pull an SLO from Datadog into the local store."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog SLO ID")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func sloPushTool() mcplib.Tool {
	return mcplib.NewTool("slo_push",
		mcplib.WithDescription("Push a locally tracked SLO to Datadog. Requires confirm=true."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog SLO ID")),
		mcplib.WithBoolean("confirm", mcplib.Description("Must be true to confirm the push")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func sloCreateTool() mcplib.Tool {
	return mcplib.NewTool("slo_create",
		mcplib.WithDescription("Create a new SLO in Datadog from a YAML definition."),
		mcplib.WithString("yaml_content", mcplib.Required(), mcplib.Description("The SLO definition in YAML format (Datadog API format)")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func sloDeleteTool() mcplib.Tool {
	return mcplib.NewTool("slo_delete",
		mcplib.WithDescription("Delete an SLO from Datadog. Irreversible. Requires confirm=true."),
		mcplib.WithString("id", mcplib.Required(), mcplib.Description("The Datadog SLO ID")),
		mcplib.WithBoolean("confirm", mcplib.Description("Must be true to confirm the deletion")),
		mcplib.WithDestructiveHintAnnotation(true),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func metricsQueryTool() mcplib.Tool {
	return mcplib.NewTool("metrics_query",
		mcplib.WithDescription("Query Datadog timeseries metrics. Returns metric data points for the given query and time range."),
		mcplib.WithString("query", mcplib.Required(), mcplib.Description("The metrics query string (e.g., 'avg:system.cpu.user{*}')")),
		mcplib.WithString("from", mcplib.Description("Start time as duration ago (e.g., '1h', '30m', '7d'). Default: 1h")),
		mcplib.WithString("to", mcplib.Description("End time as duration ago or 'now'. Default: now")),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func metricsSearchTool() mcplib.Tool {
	return mcplib.NewTool("metrics_search",
		mcplib.WithDescription("Search for available Datadog metrics by name pattern."),
		mcplib.WithString("query", mcplib.Required(), mcplib.Description("Search query to filter metric names (e.g., 'system.cpu')")),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func logsSearchTool() mcplib.Tool {
	return mcplib.NewTool("logs_search",
		mcplib.WithDescription("Search Datadog logs with a query. Returns matching log entries."),
		mcplib.WithString("query", mcplib.Required(), mcplib.Description("Log search query (e.g., 'service:my-app status:error')")),
		mcplib.WithString("from", mcplib.Description("Start time (ISO 8601 or relative like '1h'). Default: 15m ago")),
		mcplib.WithString("to", mcplib.Description("End time (ISO 8601 or 'now'). Default: now")),
		mcplib.WithNumber("limit", mcplib.Description("Maximum number of log entries to return. Default: 50")),
		mcplib.WithReadOnlyHintAnnotation(true),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

func apiCallTool() mcplib.Tool {
	return mcplib.NewTool("api_call",
		mcplib.WithDescription("Make an arbitrary Datadog API call. Use for any operation not covered by the specific tools. Supports GET, POST, PUT, DELETE methods."),
		mcplib.WithString("method", mcplib.Required(), mcplib.Description("HTTP method (GET, POST, PUT, DELETE)")),
		mcplib.WithString("path", mcplib.Required(), mcplib.Description("API path (e.g., '/api/v1/dashboard')")),
		mcplib.WithString("body", mcplib.Description("Request body as JSON string (for POST/PUT)")),
		mcplib.WithDestructiveHintAnnotation(false),
		mcplib.WithOpenWorldHintAnnotation(true),
	)
}

// ─── Tool Handlers ──────────────────────────────────────────────────────────

func (s *Server) handleDashboardList(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	resources, err := s.store.ListResources("dashboard", conn)
	if err != nil {
		return errorResult(fmt.Sprintf("listing dashboards: %v", err)), nil
	}

	data, _ := json.MarshalIndent(resources, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleDashboardListRemote(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	dashboards, err := c.ListDashboards()
	if err != nil {
		return errorResult(fmt.Sprintf("listing remote dashboards: %v", err)), nil
	}

	type dashInfo struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		URL   string `json:"url,omitempty"`
	}
	var items []dashInfo
	for _, d := range dashboards {
		items = append(items, dashInfo{
			ID:    d.GetId(),
			Title: d.GetTitle(),
			URL:   d.GetUrl(),
		})
	}

	data, _ := json.MarshalIndent(items, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleDashboardPull(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	content, err := c.GetDashboard(id)
	if err != nil {
		return errorResult(fmt.Sprintf("pulling dashboard: %v", err)), nil
	}

	// Convert JSON to YAML for storage
	var raw interface{}
	json.Unmarshal(content, &raw)
	yamlContent, _ := yaml.Marshal(raw)

	// Extract title
	var meta struct {
		Title string `json:"title"`
	}
	json.Unmarshal(content, &meta)

	if err := s.store.TrackResource(id, "dashboard", conn, meta.Title); err != nil {
		return errorResult(fmt.Sprintf("tracking resource: %v", err)), nil
	}
	if err := s.store.SaveVersion(id, "dashboard", conn, string(yamlContent), "", "", "pulled from remote"); err != nil {
		return errorResult(fmt.Sprintf("saving version: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Pulled dashboard %s (%s) successfully", id, meta.Title)), nil
}

func (s *Server) handleDashboardPush(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	if !request.GetBool("confirm", false) {
		return errorResult("Push requires confirm=true to proceed"), nil
	}

	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	version, err := s.store.GetLatestVersion(id, "dashboard", conn)
	if err != nil {
		return errorResult(fmt.Sprintf("getting latest version: %v", err)), nil
	}

	// Convert YAML to JSON for API
	var raw interface{}
	if err := yaml.Unmarshal([]byte(version.Content), &raw); err != nil {
		return errorResult(fmt.Sprintf("parsing YAML: %v", err)), nil
	}
	body, _ := json.Marshal(raw)

	if err := c.UpdateDashboard(id, body); err != nil {
		return errorResult(fmt.Sprintf("pushing dashboard: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Pushed dashboard %s (version %d) to Datadog", id, version.Version)), nil
}

func (s *Server) handleDashboardCreate(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	yamlContent, err := request.RequireString("yaml_content")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	// Convert YAML to JSON for API
	var raw interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		return errorResult(fmt.Sprintf("invalid YAML: %v", err)), nil
	}
	body, _ := json.Marshal(raw)

	id, err := c.CreateDashboard(body)
	if err != nil {
		return errorResult(fmt.Sprintf("creating dashboard: %v", err)), nil
	}

	// Track locally
	var meta struct {
		Title string `yaml:"title"`
	}
	yaml.Unmarshal([]byte(yamlContent), &meta)

	if err := s.store.TrackResource(id, "dashboard", conn, meta.Title); err != nil {
		return errorResult(fmt.Sprintf("tracking resource: %v", err)), nil
	}
	if err := s.store.SaveVersion(id, "dashboard", conn, yamlContent, "", "", "created via MCP"); err != nil {
		return errorResult(fmt.Sprintf("saving version: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Created dashboard %s", id)), nil
}

func (s *Server) handleDashboardEdit(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	yamlContent, err := request.RequireString("yaml_content")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	// Validate YAML
	var raw interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		return errorResult(fmt.Sprintf("invalid YAML: %v", err)), nil
	}

	if err := s.store.SaveVersion(id, "dashboard", conn, yamlContent, "", "", "edited via MCP"); err != nil {
		return errorResult(fmt.Sprintf("saving version: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Updated dashboard %s locally. Use dashboard_push to apply to Datadog.", id)), nil
}

func (s *Server) handleDashboardDiff(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	localVersion, err := s.store.GetLatestVersion(id, "dashboard", conn)
	if err != nil {
		return errorResult(fmt.Sprintf("getting local version: %v", err)), nil
	}

	// If we have a client, also get remote
	if s.client != nil {
		remoteContent, err := s.client.GetDashboard(id)
		if err != nil {
			return errorResult(fmt.Sprintf("getting remote dashboard: %v", err)), nil
		}

		// Convert remote to YAML for comparison
		var raw interface{}
		json.Unmarshal(remoteContent, &raw)
		remoteYAML, _ := yaml.Marshal(raw)

		result := map[string]interface{}{
			"dashboard_id":  id,
			"local_version": localVersion.Version,
			"local_content": localVersion.Content,
			"remote_content": string(remoteYAML),
			"match":         localVersion.Content == string(remoteYAML),
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		return textResult(string(data)), nil
	}

	// Local-only mode: show content
	result := map[string]interface{}{
		"dashboard_id":  id,
		"local_version": localVersion.Version,
		"local_content": localVersion.Content,
		"note":          "No API client configured, showing local content only",
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleDashboardHistory(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	versions, err := s.store.ListVersions(id, "dashboard", conn)
	if err != nil {
		return errorResult(fmt.Sprintf("listing versions: %v", err)), nil
	}

	type versionInfo struct {
		Version   int    `json:"version"`
		AppliedAt string `json:"applied_at"`
		AppliedBy string `json:"applied_by,omitempty"`
		Message   string `json:"message,omitempty"`
	}
	var items []versionInfo
	for _, v := range versions {
		items = append(items, versionInfo{
			Version:   v.Version,
			AppliedAt: v.AppliedAt.Format(time.RFC3339),
			AppliedBy: v.AppliedBy,
			Message:   v.Message,
		})
	}

	data, _ := json.MarshalIndent(items, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleDashboardRollback(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	if !request.GetBool("confirm", false) {
		return errorResult("Rollback requires confirm=true to proceed"), nil
	}

	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	toVersion := request.GetInt("to_version", 0)
	if toVersion <= 0 {
		return errorResult("to_version must be a positive integer"), nil
	}

	version, err := s.store.GetVersion(id, "dashboard", conn, toVersion)
	if err != nil {
		return errorResult(fmt.Sprintf("getting version %d: %v", toVersion, err)), nil
	}

	if err := s.store.SaveVersion(id, "dashboard", conn, version.Content, "", "", fmt.Sprintf("rollback to version %d", toVersion)); err != nil {
		return errorResult(fmt.Sprintf("saving rollback version: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Rolled back dashboard %s to version %d. Use dashboard_push to apply to Datadog.", id, toVersion)), nil
}

func (s *Server) handleDashboardDelete(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	if !request.GetBool("confirm", false) {
		return errorResult("Delete requires confirm=true to proceed. This action is irreversible."), nil
	}

	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	if err := c.DeleteDashboard(id); err != nil {
		return errorResult(fmt.Sprintf("deleting dashboard: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Deleted dashboard %s from Datadog", id)), nil
}

// ─── Monitor Handlers ───────────────────────────────────────────────────────

func (s *Server) handleMonitorList(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	resources, err := s.store.ListResources("monitor", conn)
	if err != nil {
		return errorResult(fmt.Sprintf("listing monitors: %v", err)), nil
	}

	data, _ := json.MarshalIndent(resources, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleMonitorListRemote(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	monitors, err := c.ListMonitors()
	if err != nil {
		return errorResult(fmt.Sprintf("listing remote monitors: %v", err)), nil
	}

	type monInfo struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	var items []monInfo
	for _, m := range monitors {
		items = append(items, monInfo{
			ID:   m.GetId(),
			Name: m.GetName(),
			Type: string(m.GetType()),
		})
	}

	data, _ := json.MarshalIndent(items, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleMonitorPull(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	idStr, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return errorResult(fmt.Sprintf("invalid monitor ID %q: must be numeric", idStr)), nil
	}

	content, err := c.GetMonitor(id)
	if err != nil {
		return errorResult(fmt.Sprintf("pulling monitor: %v", err)), nil
	}

	var raw interface{}
	json.Unmarshal(content, &raw)
	yamlContent, _ := yaml.Marshal(raw)

	var meta struct {
		Name string `json:"name"`
	}
	json.Unmarshal(content, &meta)

	if err := s.store.TrackResource(idStr, "monitor", conn, meta.Name); err != nil {
		return errorResult(fmt.Sprintf("tracking resource: %v", err)), nil
	}
	if err := s.store.SaveVersion(idStr, "monitor", conn, string(yamlContent), "", "", "pulled from remote"); err != nil {
		return errorResult(fmt.Sprintf("saving version: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Pulled monitor %s (%s) successfully", idStr, meta.Name)), nil
}

func (s *Server) handleMonitorPush(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	if !request.GetBool("confirm", false) {
		return errorResult("Push requires confirm=true to proceed"), nil
	}

	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	idStr, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return errorResult(fmt.Sprintf("invalid monitor ID: %v", err)), nil
	}

	version, err := s.store.GetLatestVersion(idStr, "monitor", conn)
	if err != nil {
		return errorResult(fmt.Sprintf("getting latest version: %v", err)), nil
	}

	var raw interface{}
	if err := yaml.Unmarshal([]byte(version.Content), &raw); err != nil {
		return errorResult(fmt.Sprintf("parsing YAML: %v", err)), nil
	}
	body, _ := json.Marshal(raw)

	if err := c.UpdateMonitor(id, body); err != nil {
		return errorResult(fmt.Sprintf("pushing monitor: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Pushed monitor %s (version %d) to Datadog", idStr, version.Version)), nil
}

func (s *Server) handleMonitorCreate(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	yamlContent, err := request.RequireString("yaml_content")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	var raw interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		return errorResult(fmt.Sprintf("invalid YAML: %v", err)), nil
	}
	body, _ := json.Marshal(raw)

	id, err := c.CreateMonitor(body)
	if err != nil {
		return errorResult(fmt.Sprintf("creating monitor: %v", err)), nil
	}

	idStr := strconv.FormatInt(id, 10)
	var meta struct {
		Name string `yaml:"name"`
	}
	yaml.Unmarshal([]byte(yamlContent), &meta)

	if err := s.store.TrackResource(idStr, "monitor", conn, meta.Name); err != nil {
		return errorResult(fmt.Sprintf("tracking resource: %v", err)), nil
	}
	if err := s.store.SaveVersion(idStr, "monitor", conn, yamlContent, "", "", "created via MCP"); err != nil {
		return errorResult(fmt.Sprintf("saving version: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Created monitor %s", idStr)), nil
}

func (s *Server) handleMonitorDelete(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	if !request.GetBool("confirm", false) {
		return errorResult("Delete requires confirm=true to proceed. This action is irreversible."), nil
	}

	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	idStr, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return errorResult(fmt.Sprintf("invalid monitor ID: %v", err)), nil
	}

	if err := c.DeleteMonitor(id); err != nil {
		return errorResult(fmt.Sprintf("deleting monitor: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Deleted monitor %s from Datadog", idStr)), nil
}

// ─── SLO Handlers ───────────────────────────────────────────────────────────

func (s *Server) handleSLOList(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	resources, err := s.store.ListResources("slo", conn)
	if err != nil {
		return errorResult(fmt.Sprintf("listing SLOs: %v", err)), nil
	}

	data, _ := json.MarshalIndent(resources, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleSLOListRemote(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	slos, err := c.ListSLOs()
	if err != nil {
		return errorResult(fmt.Sprintf("listing remote SLOs: %v", err)), nil
	}

	type sloInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	var items []sloInfo
	for _, slo := range slos {
		items = append(items, sloInfo{
			ID:   slo.GetId(),
			Name: slo.GetName(),
			Type: string(slo.GetType()),
		})
	}

	data, _ := json.MarshalIndent(items, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleSLOPull(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	content, err := c.GetSLO(id)
	if err != nil {
		return errorResult(fmt.Sprintf("pulling SLO: %v", err)), nil
	}

	var raw interface{}
	json.Unmarshal(content, &raw)
	yamlContent, _ := yaml.Marshal(raw)

	var meta struct {
		Name string `json:"name"`
	}
	json.Unmarshal(content, &meta)

	if err := s.store.TrackResource(id, "slo", conn, meta.Name); err != nil {
		return errorResult(fmt.Sprintf("tracking resource: %v", err)), nil
	}
	if err := s.store.SaveVersion(id, "slo", conn, string(yamlContent), "", "", "pulled from remote"); err != nil {
		return errorResult(fmt.Sprintf("saving version: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Pulled SLO %s (%s) successfully", id, meta.Name)), nil
}

func (s *Server) handleSLOPush(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	if !request.GetBool("confirm", false) {
		return errorResult("Push requires confirm=true to proceed"), nil
	}

	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	version, err := s.store.GetLatestVersion(id, "slo", conn)
	if err != nil {
		return errorResult(fmt.Sprintf("getting latest version: %v", err)), nil
	}

	var raw interface{}
	if err := yaml.Unmarshal([]byte(version.Content), &raw); err != nil {
		return errorResult(fmt.Sprintf("parsing YAML: %v", err)), nil
	}
	body, _ := json.Marshal(raw)

	if err := c.UpdateSLO(id, body); err != nil {
		return errorResult(fmt.Sprintf("pushing SLO: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Pushed SLO %s (version %d) to Datadog", id, version.Version)), nil
}

func (s *Server) handleSLOCreate(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}
	conn, err := s.requireConnection()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	yamlContent, err := request.RequireString("yaml_content")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	var raw interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &raw); err != nil {
		return errorResult(fmt.Sprintf("invalid YAML: %v", err)), nil
	}
	body, _ := json.Marshal(raw)

	id, err := c.CreateSLO(body)
	if err != nil {
		return errorResult(fmt.Sprintf("creating SLO: %v", err)), nil
	}

	var meta struct {
		Name string `yaml:"name"`
	}
	yaml.Unmarshal([]byte(yamlContent), &meta)

	if err := s.store.TrackResource(id, "slo", conn, meta.Name); err != nil {
		return errorResult(fmt.Sprintf("tracking resource: %v", err)), nil
	}
	if err := s.store.SaveVersion(id, "slo", conn, yamlContent, "", "", "created via MCP"); err != nil {
		return errorResult(fmt.Sprintf("saving version: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Created SLO %s", id)), nil
}

func (s *Server) handleSLODelete(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	if !request.GetBool("confirm", false) {
		return errorResult("Delete requires confirm=true to proceed. This action is irreversible."), nil
	}

	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	id, err := request.RequireString("id")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	if err := c.DeleteSLO(id); err != nil {
		return errorResult(fmt.Sprintf("deleting SLO: %v", err)), nil
	}

	return textResult(fmt.Sprintf("Deleted SLO %s from Datadog", id)), nil
}

// ─── Metrics & Logs Handlers ────────────────────────────────────────────────

func (s *Server) handleMetricsQuery(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	query, err := request.RequireString("query")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	fromStr := request.GetString("from", "1h")
	from := parseDuration(fromStr)
	to := time.Now()

	result, err := c.QueryMetrics(query, from, to)
	if err != nil {
		return errorResult(fmt.Sprintf("querying metrics: %v", err)), nil
	}

	return textResult(string(result)), nil
}

func (s *Server) handleMetricsSearch(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	query, err := request.RequireString("query")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	result, err := c.SearchMetrics(query)
	if err != nil {
		return errorResult(fmt.Sprintf("searching metrics: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), nil
}

func (s *Server) handleLogsSearch(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	c, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	query, err := request.RequireString("query")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	from := request.GetString("from", "now-15m")
	to := request.GetString("to", "now")
	limit := request.GetInt("limit", 50)

	result, err := c.SearchLogs(query, from, to, limit)
	if err != nil {
		return errorResult(fmt.Sprintf("searching logs: %v", err)), nil
	}

	return textResult(string(result)), nil
}

// ─── API Pass-through Handler ───────────────────────────────────────────────

func (s *Server) handleAPICall(_ context.Context, request mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	_, err := s.requireClient()
	if err != nil {
		return errorResult(err.Error()), nil
	}

	method, err := request.RequireString("method")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	path, err := request.RequireString("path")
	if err != nil {
		return errorResult(err.Error()), nil
	}

	_ = request.GetString("body", "")

	// For now, return info about the request — full HTTP pass-through
	// requires the raw API client which we'll implement with the api_call command
	result := map[string]string{
		"method":  method,
		"path":    path,
		"status":  "API pass-through is available via 'ddctl api raw' command. Direct HTTP calls from MCP are planned for a future release.",
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

// parseDuration converts a human-friendly duration string to a time.Time in the past.
func parseDuration(s string) time.Time {
	now := time.Now()
	if s == "" || s == "now" {
		return now
	}

	// Try standard Go duration first
	if d, err := time.ParseDuration(s); err == nil {
		return now.Add(-d)
	}

	// Try custom formats like "7d", "30d"
	if len(s) > 1 && s[len(s)-1] == 'd' {
		if days, err := strconv.Atoi(s[:len(s)-1]); err == nil {
			return now.Add(-time.Duration(days) * 24 * time.Hour)
		}
	}

	// Default: 1 hour ago
	return now.Add(-time.Hour)
}
