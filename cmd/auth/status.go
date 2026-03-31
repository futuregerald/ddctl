// cmd/auth/status.go
package auth

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/futuregerald/ddctl/internal/output"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show credential source and masked keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(config.Path())
		connName := cmd.Root().PersistentFlags().Lookup("connection").Value.String()

		if connName == "" {
			if cfg.DefaultConnection != "" {
				connName = cfg.DefaultConnection
			} else {
				s, err := store.New(config.DBPath())
				if err == nil {
					defer s.Close()
					if dc, err := s.GetDefaultConnection(); err == nil {
						connName = dc.Name
					}
				}
			}
		}

		if connName == "" {
			return fmt.Errorf("no connection specified")
		}

		creds, source, err := keyring.ResolveCredentials(connName)
		if err != nil {
			return err
		}

		status := map[string]string{
			"connection": connName,
			"source":     string(source),
			"api_key":    keyring.MaskKey(creds.APIKey),
			"app_key":    keyring.MaskKey(creds.AppKey),
		}

		if source == keyring.SourceEnvVar {
			fmt.Fprintln(os.Stderr, "Warning: credentials sourced from environment variables (visible in process listings)")
		}

		format := cfg.Output
		if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Changed {
			format = f.Value.String()
		}

		switch format {
		case "json":
			output.JSON(os.Stdout, status)
		case "yaml":
			output.YAML(os.Stdout, status)
		default:
			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"Connection", connName},
				{"Source", string(source)},
				{"API Key", keyring.MaskKey(creds.APIKey)},
				{"App Key", keyring.MaskKey(creds.AppKey)},
			}
			output.Table(os.Stdout, headers, rows)
		}
		return nil
	},
}
