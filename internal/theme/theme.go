// internal/theme/theme.go
package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name    string
	Enabled bool
}

var (
	Green      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff41"))
	DimGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00cc33"))
	Cyan       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00cccc"))
	BrightCyan = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff"))
	Yellow     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffbf00"))
	Red        = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444"))
	Dim        = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
)

func New(name string) *Theme {
	enabled := name != "none" && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
	return &Theme{Name: name, Enabled: enabled}
}

func (t *Theme) IsRetro() bool {
	return t.Enabled && t.Name == "retro"
}

func (t *Theme) Success(s string) string {
	if !t.Enabled {
		return s
	}
	return BrightCyan.Render(s)
}

func (t *Theme) Warning(s string) string {
	if !t.Enabled {
		return s
	}
	return Yellow.Render(s)
}

func (t *Theme) Err(s string) string {
	if !t.Enabled {
		return s
	}
	return Red.Render(s)
}

func (t *Theme) Highlight(s string) string {
	if !t.Enabled {
		return s
	}
	return Green.Render(s)
}

func (t *Theme) ID(s string) string {
	if !t.Enabled {
		return s
	}
	return Cyan.Render(s)
}

func (t *Theme) Label(s string) string {
	if !t.Enabled {
		return s
	}
	return DimGreen.Render(s)
}

func (t *Theme) Flavor(s string) string {
	if !t.Enabled || t.Name != "retro" {
		return ""
	}
	return DimGreen.Italic(true).Render(s)
}
