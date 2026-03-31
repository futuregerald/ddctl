// cmd/dashboard/diff.go
package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/futuregerald/ddctl/cmd/cmdutil"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var diffFlagVersion int

var diffCmd = &cobra.Command{
	Use:   "diff [id]",
	Short: "Diff local dashboard vs remote or a specific version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dashID := args[0]

		requireRemote := diffFlagVersion == 0
		deps, err := cmdutil.InitDeps(cmd, requireRemote)
		if err != nil {
			return err
		}
		defer deps.Close()

		// Get local latest
		localVersion, err := deps.Store.GetLatestVersion(dashID, "dashboard", deps.ConnName)
		if err != nil {
			return err
		}

		var compareContent string
		var compareLabel string

		if diffFlagVersion > 0 {
			// Compare against specific version
			v, err := deps.Store.GetVersion(dashID, "dashboard", deps.ConnName, diffFlagVersion)
			if err != nil {
				return err
			}
			compareContent = v.Content
			compareLabel = fmt.Sprintf("version %d", diffFlagVersion)
		} else {
			// Compare against remote
			jsonBytes, err := deps.Client.GetDashboard(dashID)
			if err != nil {
				return err
			}
			var raw interface{}
			json.Unmarshal(jsonBytes, &raw)
			yamlBytes, _ := yaml.Marshal(raw)
			compareContent = string(yamlBytes)
			compareLabel = "remote"
		}

		// Simple line-by-line diff
		localLines := strings.Split(localVersion.Content, "\n")
		compareLines := strings.Split(compareContent, "\n")

		hasDiff := false
		maxLines := len(localLines)
		if len(compareLines) > maxLines {
			maxLines = len(compareLines)
		}

		for i := 0; i < maxLines; i++ {
			var local, compare string
			if i < len(localLines) {
				local = localLines[i]
			}
			if i < len(compareLines) {
				compare = compareLines[i]
			}
			if local != compare {
				if !hasDiff {
					fmt.Fprintf(os.Stdout, "--- local (version %d)\n+++ %s\n", localVersion.Version, compareLabel)
					hasDiff = true
				}
				if local != "" {
					fmt.Fprintf(os.Stdout, "- %s\n", local)
				}
				if compare != "" {
					fmt.Fprintf(os.Stdout, "+ %s\n", compare)
				}
			}
		}

		if !hasDiff {
			fmt.Fprintln(os.Stderr, "No differences.")
		}
		return nil
	},
}

func init() {
	diffCmd.Flags().IntVar(&diffFlagVersion, "version", 0, "Compare against a specific local version")
}
