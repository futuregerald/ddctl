// cmd/cmdutil/cmdutil.go
// Shared helpers for resource commands (dashboard, monitor, SLO).
package cmdutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/futuregerald/ddctl/internal/client"
	"github.com/futuregerald/ddctl/internal/config"
	"github.com/futuregerald/ddctl/internal/keyring"
	"github.com/futuregerald/ddctl/internal/store"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Deps holds the initialized dependencies for a command.
type Deps struct {
	Store      *store.Store
	Client     *client.Client
	Config     *config.Config
	ConnName   string
	Format     string
}

// InitDeps initializes the store, resolves the connection, and creates a client.
// If requireClient is false, the client will be nil (for local-only operations).
func InitDeps(cmd *cobra.Command, requireClient bool) (*Deps, error) {
	cfg, err := config.Load(config.Path())
	if err != nil {
		cfg = config.DefaultConfig()
	}

	s, err := store.New(config.DBPath())
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}

	connName := resolveConnFlag(cmd)
	if connName == "" && cfg.DefaultConnection != "" {
		connName = cfg.DefaultConnection
	}
	if connName == "" {
		if dc, err := s.GetDefaultConnection(); err == nil {
			connName = dc.Name
		}
	}
	if connName == "" {
		s.Close()
		return nil, fmt.Errorf("no connection specified. Use --connection or set a default")
	}

	format := resolveOutputFlag(cmd)
	if format == "" {
		format = cfg.Output
	}

	deps := &Deps{
		Store:    s,
		Config:   cfg,
		ConnName: connName,
		Format:   format,
	}

	if requireClient {
		conn, err := s.GetConnection(connName)
		if err != nil {
			s.Close()
			return nil, fmt.Errorf("connection %q: %w", connName, err)
		}
		creds, _, err := keyring.ResolveCredentials(connName)
		if err != nil {
			s.Close()
			return nil, err
		}
		c, err := client.New(conn.Site, creds)
		if err != nil {
			s.Close()
			return nil, fmt.Errorf("creating client: %w", err)
		}
		deps.Client = c
	}

	return deps, nil
}

// Close closes the store.
func (d *Deps) Close() {
	if d.Store != nil {
		d.Store.Close()
	}
}

// ConfirmOrSkip asks for confirmation unless --yes is set.
func ConfirmOrSkip(cmd *cobra.Command, prompt string) bool {
	if yes, _ := cmd.Flags().GetBool("yes"); yes {
		return true
	}
	// Check the root persistent flag
	if f := cmd.Root().PersistentFlags().Lookup("yes"); f != nil && f.Value.String() == "true" {
		return true
	}
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes"
}

// OpenEditor opens the specified editor on a file and waits.
func OpenEditor(cfg *config.Config, filePath string) error {
	editor := cfg.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return fmt.Errorf("no editor configured. Set 'editor' in config or $EDITOR environment variable")
	}

	var args []string
	switch editor {
	case "cursor", "code":
		args = []string{"--wait", filePath}
	default:
		args = []string{filePath}
	}

	c := exec.Command(editor, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// SHA256 computes the hex SHA256 of data.
func SHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func resolveConnFlag(cmd *cobra.Command) string {
	if f := cmd.Root().PersistentFlags().Lookup("connection"); f != nil && f.Changed {
		return f.Value.String()
	}
	return ""
}

func resolveOutputFlag(cmd *cobra.Command) string {
	if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Changed {
		return f.Value.String()
	}
	return ""
}

// StripExtension removes .yaml/.yml extension from a filename.
func StripExtension(name string) string {
	name = strings.TrimSuffix(name, ".yaml")
	name = strings.TrimSuffix(name, ".yml")
	return name
}

// EditResource opens a resource in the user's editor, validates the YAML, and saves a new version.
func EditResource(deps *Deps, resourceID, resourceType string) error {
	version, err := deps.Store.GetLatestVersion(resourceID, resourceType, deps.ConnName)
	if err != nil {
		return err
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("ddctl-%s-%s.yaml", resourceType, resourceID))
	if err := os.WriteFile(tmpFile, []byte(version.Content), 0600); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	beforeHash := SHA256([]byte(version.Content))
	if err := OpenEditor(deps.Config, tmpFile); err != nil {
		return fmt.Errorf("editor: %w", err)
	}

	edited, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	if SHA256(edited) == beforeHash {
		fmt.Fprintln(os.Stderr, "No changes.")
		return nil
	}

	var check interface{}
	if err := yaml.Unmarshal(edited, &check); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	if err := deps.Store.SaveVersion(resourceID, resourceType, deps.ConnName, string(edited), "", "", "edited locally"); err != nil {
		return fmt.Errorf("saving version: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Saved new local version. Run 'ddctl %s push %s' to apply remotely.\n", resourceType, resourceID)
	return nil
}

// RollbackResource copies a previous version's content as a new version.
func RollbackResource(deps *Deps, resourceID, resourceType string, toVersion int) error {
	targetVersion, err := deps.Store.GetVersion(resourceID, resourceType, deps.ConnName, toVersion)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("rollback to version %d", toVersion)
	if err := deps.Store.SaveVersion(resourceID, resourceType, deps.ConnName, targetVersion.Content, "", "", msg); err != nil {
		return fmt.Errorf("saving rollback version: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Rolled back to version %d. Run 'ddctl %s push %s' to apply remotely.\n", toVersion, resourceType, resourceID)
	return nil
}

// ExportResource exports a resource's latest content to a file or stdout.
func ExportResource(deps *Deps, resourceID, resourceType, outputPath string) error {
	version, err := deps.Store.GetLatestVersion(resourceID, resourceType, deps.ConnName)
	if err != nil {
		return err
	}

	if outputPath == "" {
		fmt.Print(version.Content)
		return nil
	}

	if err := os.WriteFile(outputPath, []byte(version.Content), 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Exported %s %s to %s\n", resourceType, resourceID, outputPath)
	return nil
}
