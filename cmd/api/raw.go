// cmd/api/raw.go
package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/spf13/cobra"
)

var rawFlagBody string

var rawCmd = &cobra.Command{
	Use:   "raw [METHOD] [path]",
	Short: "Execute a raw HTTP request to the Datadog API",
	Long: `Send a raw HTTP request to any Datadog API endpoint.

  ddctl api raw GET /v1/dashboard
  ddctl api raw POST /v1/dashboard --body @payload.json
  ddctl api raw GET /v2/some/new/endpoint`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		method := strings.ToUpper(args[0])
		path := args[1]

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		return executeRawRequest(deps, method, path, rawFlagBody)
	},
}

func executeRawRequest(deps *cmdutil.Deps, method, path, body string) error {
	conn, err := deps.Store.GetConnection(deps.ConnName)
	if err != nil {
		return err
	}

	creds, _, err := keyring.ResolveCredentials(deps.ConnName)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := fmt.Sprintf("https://api.%s%s", conn.Site, path)

	var bodyReader io.Reader
	if body != "" {
		if strings.HasPrefix(body, "@") {
			data, err := os.ReadFile(body[1:])
			if err != nil {
				return fmt.Errorf("reading body file: %w", err)
			}
			bodyReader = strings.NewReader(string(data))
		} else {
			bodyReader = strings.NewReader(body)
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", creds.APIKey)
	req.Header.Set("DD-APPLICATION-KEY", creds.AppKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20))
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "HTTP %d %s\n", resp.StatusCode, resp.Status)
	}

	os.Stdout.Write(respBody)
	fmt.Fprintln(os.Stdout)
	return nil
}

func init() {
	rawCmd.Flags().StringVar(&rawFlagBody, "body", "", "Request body (JSON string or @filename)")
}
