package tui

import (
	"github.com/apideck-io/cli/internal/ui"
	"charm.land/lipgloss/v2"
)

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary()).Padding(0, 1)
	selectedStyle    = lipgloss.NewStyle().Foreground(ui.ColorPrimary()).Bold(true)
	normalStyle      = lipgloss.NewStyle().Foreground(ui.ColorWhite())
	dimStyle         = lipgloss.NewStyle().Foreground(ui.ColorDim())
	panelStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ui.ColorDim()).Padding(0, 1)
	activePanelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ui.ColorPrimary()).Padding(0, 1)
)

// methodStyle returns the lipgloss style for an HTTP method.
func methodStyle(method string) lipgloss.Style {
	switch method {
	case "GET":
		return lipgloss.NewStyle().Foreground(ui.ColorSuccess()).Bold(true)
	case "POST":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#74B9FF")).Bold(true)
	case "PUT", "PATCH":
		return lipgloss.NewStyle().Foreground(ui.ColorWarning()).Bold(true)
	case "DELETE":
		return lipgloss.NewStyle().Foreground(ui.ColorError()).Bold(true)
	default:
		return normalStyle
	}
}
