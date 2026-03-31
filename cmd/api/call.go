// cmd/api/call.go
package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/spf13/cobra"
)

var (
	callFlagID   string
	callFlagBody string
)

var callCmd = &cobra.Command{
	Use:   "call [group.operation]",
	Short: "Execute a named API operation",
	Long: `Execute a Datadog API operation by its group and name.

  ddctl api call dashboards.list
  ddctl api call dashboards.get --id abc-123
  ddctl api call monitors.create --body @monitor.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[0], ".", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected format: group.operation (e.g. dashboards.list)")
		}
		groupName := parts[0]
		opName := parts[1]

		ops, ok := apiGroups[groupName]
		if !ok {
			return fmt.Errorf("unknown API group %q", groupName)
		}

		var op *apiOperation
		for _, o := range ops {
			if o.Name == opName {
				op = &o
				break
			}
		}
		if op == nil {
			return fmt.Errorf("unknown operation %q in group %q", opName, groupName)
		}

		deps, err := cmdutil.InitDeps(cmd, true)
		if err != nil {
			return err
		}
		defer deps.Close()

		// Build the path
		path := op.Path
		if callFlagID != "" {
			path = strings.Replace(path, "{id}", callFlagID, 1)
			path = strings.Replace(path, "{metric}", callFlagID, 1)
		}

		// Get the site from the connection
		conn, err := deps.Store.GetConnection(deps.ConnName)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("https://api.%s%s", conn.Site, path)

		var bodyReader io.Reader
		if callFlagBody != "" {
			if strings.HasPrefix(callFlagBody, "@") {
				data, err := os.ReadFile(callFlagBody[1:])
				if err != nil {
					return fmt.Errorf("reading body file: %w", err)
				}
				bodyReader = strings.NewReader(string(data))
			} else {
				bodyReader = strings.NewReader(callFlagBody)
			}
		}

		req, err := http.NewRequest(op.Method, url, bodyReader)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		// Use the client's credentials
		ctx := deps.Client.Context()
		apiKeys, ok := ctx.Value(
			apiKeysContextKey{},
		).(map[string]interface{})
		_ = apiKeys
		_ = ok

		// For simplicity, execute through the raw HTTP path
		return executeRawRequest(deps, op.Method, path, callFlagBody)
	},
}

type apiKeysContextKey struct{}

func init() {
	callCmd.Flags().StringVar(&callFlagID, "id", "", "Resource ID for operations that require one")
	callCmd.Flags().StringVar(&callFlagBody, "body", "", "Request body (JSON string or @filename)")
}
