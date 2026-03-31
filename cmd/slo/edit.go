// cmd/slo/edit.go
package slo

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
	Short: "Edit an SLO in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sloID := args[0]

		deps, err := cmdutil.InitDeps(cmd, false)
		if err != nil {
			return err
		}
		defer deps.Close()

		version, err := deps.Store.GetLatestVersion(sloID, "slo", deps.ConnName)
		if err != nil {
			return err
		}

		tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("ddctl-slo-%s.yaml", sloID))
		if err := os.WriteFile(tmpFile, []byte(version.Content), 0600); err != nil {
			return err
		}
		defer os.Remove(tmpFile)

		beforeHash := cmdutil.SHA256([]byte(version.Content))
		if err := cmdutil.OpenEditor(deps.Config, tmpFile); err != nil {
			return err
		}

		edited, err := os.ReadFile(tmpFile)
		if err != nil {
			return err
		}

		if cmdutil.SHA256(edited) == beforeHash {
			fmt.Fprintln(os.Stderr, "No changes.")
			return nil
		}

		var check interface{}
		if err := yaml.Unmarshal(edited, &check); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}

		if err := deps.Store.SaveVersion(sloID, "slo", deps.ConnName, string(edited), "", "", "edited locally"); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Saved new local version. Run 'ddctl slo push %s' to apply remotely.\n", sloID)
		return nil
	},
}
