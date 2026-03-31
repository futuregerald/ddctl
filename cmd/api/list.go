// cmd/api/list.go
package api

import (
	"fmt"
	"os"
	"sort"

	"github.com/futuregerald/ddctl/internal/output"
	"github.com/spf13/cobra"
)

// apiGroups defines available API groups and their operations.
var apiGroups = map[string][]apiOperation{
	"dashboards": {
		{Name: "list", Method: "GET", Path: "/v1/dashboard", Description: "List all dashboards"},
		{Name: "get", Method: "GET", Path: "/v1/dashboard/{id}", Description: "Get a dashboard by ID"},
		{Name: "create", Method: "POST", Path: "/v1/dashboard", Description: "Create a dashboard"},
		{Name: "update", Method: "PUT", Path: "/v1/dashboard/{id}", Description: "Update a dashboard"},
		{Name: "delete", Method: "DELETE", Path: "/v1/dashboard/{id}", Description: "Delete a dashboard"},
	},
	"monitors": {
		{Name: "list", Method: "GET", Path: "/v1/monitor", Description: "List all monitors"},
		{Name: "get", Method: "GET", Path: "/v1/monitor/{id}", Description: "Get a monitor by ID"},
		{Name: "create", Method: "POST", Path: "/v1/monitor", Description: "Create a monitor"},
		{Name: "update", Method: "PUT", Path: "/v1/monitor/{id}", Description: "Update a monitor"},
		{Name: "delete", Method: "DELETE", Path: "/v1/monitor/{id}", Description: "Delete a monitor"},
	},
	"slos": {
		{Name: "list", Method: "GET", Path: "/v1/slo", Description: "List all SLOs"},
		{Name: "get", Method: "GET", Path: "/v1/slo/{id}", Description: "Get an SLO by ID"},
		{Name: "create", Method: "POST", Path: "/v1/slo", Description: "Create an SLO"},
		{Name: "update", Method: "PUT", Path: "/v1/slo/{id}", Description: "Update an SLO"},
		{Name: "delete", Method: "DELETE", Path: "/v1/slo/{id}", Description: "Delete an SLO"},
	},
	"metrics": {
		{Name: "search", Method: "GET", Path: "/v1/search", Description: "Search metrics"},
		{Name: "query", Method: "GET", Path: "/v1/query", Description: "Query timeseries data"},
		{Name: "metadata", Method: "GET", Path: "/v1/metrics/{metric}", Description: "Get metric metadata"},
	},
	"logs": {
		{Name: "search", Method: "POST", Path: "/v2/logs/events/search", Description: "Search logs"},
	},
	"events": {
		{Name: "list", Method: "GET", Path: "/v1/events", Description: "List events"},
		{Name: "get", Method: "GET", Path: "/v1/events/{id}", Description: "Get an event"},
		{Name: "create", Method: "POST", Path: "/v1/events", Description: "Post an event"},
	},
	"hosts": {
		{Name: "list", Method: "GET", Path: "/v1/hosts", Description: "List hosts"},
		{Name: "search", Method: "GET", Path: "/v1/search", Description: "Search hosts"},
	},
	"downtimes": {
		{Name: "list", Method: "GET", Path: "/v1/downtime", Description: "List downtimes"},
		{Name: "get", Method: "GET", Path: "/v1/downtime/{id}", Description: "Get a downtime"},
		{Name: "create", Method: "POST", Path: "/v1/downtime", Description: "Create a downtime"},
		{Name: "update", Method: "PUT", Path: "/v1/downtime/{id}", Description: "Update a downtime"},
		{Name: "delete", Method: "DELETE", Path: "/v1/downtime/{id}", Description: "Cancel a downtime"},
	},
}

type apiOperation struct {
	Name        string `json:"name" yaml:"name"`
	Method      string `json:"method" yaml:"method"`
	Path        string `json:"path" yaml:"path"`
	Description string `json:"description" yaml:"description"`
}

var listCmd = &cobra.Command{
	Use:   "list [group]",
	Short: "List API groups or operations within a group",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List all groups
			type groupInfo struct {
				Name       string `json:"name" yaml:"name"`
				Operations int    `json:"operations" yaml:"operations"`
			}
			var groups []groupInfo
			names := make([]string, 0, len(apiGroups))
			for name := range apiGroups {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				groups = append(groups, groupInfo{Name: name, Operations: len(apiGroups[name])})
			}

			headers := []string{"GROUP", "OPERATIONS"}
			var rows [][]string
			for _, g := range groups {
				rows = append(rows, []string{g.Name, fmt.Sprintf("%d", g.Operations)})
			}
			output.Table(os.Stdout, headers, rows)
			return nil
		}

		// List operations in a group
		groupName := args[0]
		ops, ok := apiGroups[groupName]
		if !ok {
			return fmt.Errorf("unknown API group %q. Run 'ddctl api list' to see available groups", groupName)
		}

		headers := []string{"OPERATION", "METHOD", "PATH", "DESCRIPTION"}
		var rows [][]string
		for _, op := range ops {
			rows = append(rows, []string{
				fmt.Sprintf("%s.%s", groupName, op.Name),
				op.Method,
				op.Path,
				op.Description,
			})
		}
		output.Table(os.Stdout, headers, rows)
		return nil
	},
}
