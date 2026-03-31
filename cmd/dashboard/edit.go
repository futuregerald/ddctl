// cmd/dashboard/edit.go
package dashboard

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit a dashboard in your editor",
	Long: `Opens the dashboard YAML in your configured editor.
On save, validates the YAML and stores as a new local version.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dashID := args[0]

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		// Get latest version
		version, err := deps.Store.GetLatestVersion(dashID, "dashboard", deps.ConnName)
		if err != nil {
			return err
		}

		// Write to temp file
		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, fmt.Sprintf("ddctl-dashboard-%s.yaml", dashID))
		if err := os.WriteFile(tmpFile, []byte(version.Content), 0600); err != nil {
			return fmt.Errorf("writing temp file: %w", err)
		}
		defer os.Remove(tmpFile)

		// Checksum before edit
		beforeHash := cmdutil.SHA256([]byte(version.Content))

		// Open editor
		if err := cmdutil.OpenEditor(deps.Config, tmpFile); err != nil {
			return fmt.Errorf("editor: %w", err)
		}

		// Read back
		edited, err := os.ReadFile(tmpFile)
		if err != nil {
			return fmt.Errorf("reading edited file: %w", err)
		}

		// Check if changed
		afterHash := cmdutil.SHA256(edited)
		if beforeHash == afterHash {
			fmt.Fprintln(os.Stderr, "No changes.")
			return nil
		}

		// Validate YAML
		var check interface{}
		if err := yaml.Unmarshal(edited, &check); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}

		// Save new version
		if err := deps.Store.SaveVersion(dashID, "dashboard", deps.ConnName, string(edited), "", "", "edited locally"); err != nil {
			return fmt.Errorf("saving version: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Saved new local version. Run 'ddctl dashboard push %s' to apply remotely.\n", dashID)
		return nil
	},
}
