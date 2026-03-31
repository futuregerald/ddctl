// cmd/db/prune.go
package db

import (
	"fmt"
	"os"

	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune old resource versions",
	Long:  `Removes old versions beyond the configured versions_to_keep limit for all tracked resources.`,
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

		keep := cfg.VersionsToKeep
		if keep <= 0 {
			fmt.Fprintln(os.Stderr, "versions_to_keep is 0 (unlimited). Nothing to prune.")
			return nil
		}

		// Get all resources across all connections and types
		rows, err := s.DB().Query(
			`SELECT resource_id, resource_type, connection FROM resources`,
		)
		if err != nil {
			return fmt.Errorf("listing resources: %w", err)
		}
		defer rows.Close()

		var totalPruned int
		var resourceCount int
		for rows.Next() {
			var resourceID, resourceType, connection string
			if err := rows.Scan(&resourceID, &resourceType, &connection); err != nil {
				return err
			}

			pruned, err := s.PruneVersions(resourceID, resourceType, connection, keep)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to prune %s/%s: %v\n", resourceType, resourceID, err)
				continue
			}
			if pruned > 0 {
				resourceCount++
				totalPruned += pruned
			}
		}

		if totalPruned == 0 {
			fmt.Fprintln(os.Stderr, "Nothing to prune.")
		} else {
			fmt.Fprintf(os.Stderr, "Pruned %d versions across %d resources (keeping %d per resource).\n", totalPruned, resourceCount, keep)
		}
		return nil
	},
}
