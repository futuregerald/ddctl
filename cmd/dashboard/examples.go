// cmd/dashboard/examples.go
package dashboard

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"github.com/futuregerald/ddctl/internal/output"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

//go:embed examples/*.yaml
var exampleFS embed.FS

var (
	examplesFlagPreview bool
	examplesFlagImport  bool
)

var examplesCmd = &cobra.Command{
	Use:   "examples [name]",
	Short: "List, preview, or import bundled example dashboards",
	Long: `Without arguments, lists all bundled example dashboards.
With a name argument, use --preview to see the YAML or --import to import into local store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := exampleFS.ReadDir("examples")
		if err != nil {
			return fmt.Errorf("reading examples: %w", err)
		}

		if len(args) == 0 {
			// List examples
			type exampleInfo struct {
				Name string `json:"name" yaml:"name"`
				File string `json:"file" yaml:"file"`
			}
			var items []exampleInfo
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
					continue
				}
				name := strings.TrimSuffix(e.Name(), ".yaml")
				items = append(items, exampleInfo{Name: name, File: e.Name()})
			}

			deps, err := cmdutil.InitDeps(cmd, false)
			if err != nil {
				// For listing examples, we can work without a connection
				headers := []string{"NAME", "FILE"}
				var rows [][]string
				for _, item := range items {
					rows = append(rows, []string{item.Name, item.File})
				}
				output.Table(os.Stdout, headers, rows)
				return nil
			}
			defer deps.Close()

			switch deps.Format {
			case "json":
				output.JSON(os.Stdout, items)
			case "yaml":
				output.YAML(os.Stdout, items)
			default:
				headers := []string{"NAME", "FILE"}
				var rows [][]string
				for _, item := range items {
					rows = append(rows, []string{item.Name, item.File})
				}
				output.Table(os.Stdout, headers, rows)
			}
			return nil
		}

		// Named example
		name := args[0]
		filename := name + ".yaml"
		data, err := exampleFS.ReadFile("examples/" + filename)
		if err != nil {
			return fmt.Errorf("example %q not found", name)
		}

		if examplesFlagPreview {
			fmt.Print(string(data))
			return nil
		}

		if examplesFlagImport {
			deps, err := cmdutil.InitDeps(cmd, false)
			if err != nil {
				return err
			}
			defer deps.Close()

			// Validate YAML
			var content interface{}
			if err := yaml.Unmarshal(data, &content); err != nil {
				return fmt.Errorf("invalid example YAML: %w", err)
			}

			var meta struct {
				Title string `yaml:"title"`
			}
			if err := yaml.Unmarshal(data, &meta); err != nil {
				return fmt.Errorf("parsing example metadata: %w", err)
			}
			title := meta.Title
			if title == "" {
				title = name
			}

			dashID := fmt.Sprintf("example-%s", name)
			if err := deps.Store.TrackResource(dashID, "dashboard", deps.ConnName, title); err != nil {
				return fmt.Errorf("tracking resource: %w", err)
			}
			if err := deps.Store.SaveVersion(dashID, "dashboard", deps.ConnName, string(data), "", "", "imported from example"); err != nil {
				return fmt.Errorf("saving version: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Imported example %q as dashboard %s\n", name, dashID)
			return nil
		}

		// Default: show the YAML
		fmt.Print(string(data))
		return nil
	},
}

func init() {
	examplesCmd.Flags().BoolVar(&examplesFlagPreview, "preview", false, "Preview the example YAML")
	examplesCmd.Flags().BoolVar(&examplesFlagImport, "import", false, "Import the example into local store")
}
