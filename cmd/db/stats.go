// cmd/db/stats.go
package db

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database statistics",
	Long:  `Display database size, resource counts, and version counts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(config.Path())
		if err != nil {
			cfg = config.DefaultConfig()
		}

		s, err := store.New(config.DBPath())
		if err != nil {
			return fmt.Errorf("opening store: %w", err)
		}
		defer s.Close()

		// Get counts
		var connectionCount, resourceCount, versionCount int

		s.DB().QueryRow("SELECT count(*) FROM connections").Scan(&connectionCount)
		s.DB().QueryRow("SELECT count(*) FROM resources").Scan(&resourceCount)
		s.DB().QueryRow("SELECT count(*) FROM resource_versions").Scan(&versionCount)

		// Get db file size
		dbPath := config.DBPath()
		info, _ := os.Stat(dbPath)
		var dbSize int64
		if info != nil {
			dbSize = info.Size()
		}

		// Per-type counts
		type typeCount struct {
			Type     string `json:"type" yaml:"type"`
			Count    int    `json:"resources" yaml:"resources"`
			Versions int    `json:"versions" yaml:"versions"`
		}

		var typeCounts []typeCount
		typeRows, err := s.DB().Query(
			`SELECT r.resource_type,
			        count(DISTINCT r.resource_id),
			        (SELECT count(*) FROM resource_versions rv
			         WHERE rv.resource_type = r.resource_type AND rv.connection = r.connection)
			 FROM resources r
			 GROUP BY r.resource_type, r.connection`,
		)
		if err == nil {
			defer typeRows.Close()
			for typeRows.Next() {
				var tc typeCount
				typeRows.Scan(&tc.Type, &tc.Count, &tc.Versions)
				typeCounts = append(typeCounts, tc)
			}
		}

		stats := map[string]interface{}{
			"db_path":     dbPath,
			"db_size":     formatBytes(dbSize),
			"connections": connectionCount,
			"resources":   resourceCount,
			"versions":    versionCount,
			"by_type":     typeCounts,
		}

		format := cfg.Output
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Changed {
			format = f.Value.String()
		}

		switch format {
		case "json":
			output.JSON(os.Stdout, stats)
		case "yaml":
			output.YAML(os.Stdout, stats)
		default:
			fmt.Fprintf(os.Stdout, "  Database:    %s\n", dbPath)
			fmt.Fprintf(os.Stdout, "  Size:        %s\n", formatBytes(dbSize))
			fmt.Fprintf(os.Stdout, "  Connections: %d\n", connectionCount)
			fmt.Fprintf(os.Stdout, "  Resources:   %d\n", resourceCount)
			fmt.Fprintf(os.Stdout, "  Versions:    %d\n", versionCount)
			if len(typeCounts) > 0 {
				fmt.Fprintln(os.Stdout)
				headers := []string{"TYPE", "RESOURCES", "VERSIONS"}
				var rows [][]string
				for _, tc := range typeCounts {
					rows = append(rows, []string{tc.Type, fmt.Sprintf("%d", tc.Count), fmt.Sprintf("%d", tc.Versions)})
				}
				output.Table(os.Stdout, headers, rows)
			}
		}
		return nil
	},
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMG"[exp])
}
