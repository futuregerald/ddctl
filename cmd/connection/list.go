// cmd/connection/list.go
package connection

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all connection profiles",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := mustLoadConfig()
		s, err := store.New(config.DBPath())
		if err != nil {
			return fmt.Errorf("opening store: %w", err)
		}
		defer s.Close()

		conns, err := s.ListConnections()
		if err != nil {
			return fmt.Errorf("listing connections: %w", err)
		}

		format := resolveFormat(cmd, cfg)
		switch format {
		case "json":
			output.JSON(os.Stdout, conns)
		case "yaml":
			output.YAML(os.Stdout, conns)
		default:
			headers := []string{"NAME", "SITE", "DEFAULT", "CREATED"}
			var rows [][]string
			for _, c := range conns {
				def := ""
				if c.IsDefault {
					def = "*"
				}
				rows = append(rows, []string{c.Name, c.Site, def, c.CreatedAt.Format("2006-01-02 15:04")})
			}
			output.Table(os.Stdout, headers, rows)
		}
		return nil
	},
}

func mustLoadConfig() *config.Config {
	cfg, err := config.Load(config.Path())
	if err != nil {
		cfg = config.DefaultConfig()
	}
	return cfg
}

func resolveFormat(cmd *cobra.Command, cfg *config.Config) string {
	if f, _ := cmd.Flags().GetString("output"); f != "" {
		return f
	}
	if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Changed {
		return f.Value.String()
	}
	return cfg.Output
}
