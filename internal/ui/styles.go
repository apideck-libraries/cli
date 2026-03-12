package ui

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

// lightDark is a helper that picks light or dark color based on terminal background.
// It detects the background once at package init time.
var lightDark lipgloss.LightDarkFunc

func init() {
	hasDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark = lipgloss.LightDark(hasDark)
}

// Brand colors (Apideck-inspired). Each returns the appropriate color for the
// current terminal background (light or dark).
var (
	ColorPrimary = func() color.Color { return lightDark(lipgloss.Color("#6C5CE7"), lipgloss.Color("#A29BFE")) }
	ColorSuccess = func() color.Color { return lightDark(lipgloss.Color("#00B894"), lipgloss.Color("#55EFC4")) }
	ColorError   = func() color.Color { return lightDark(lipgloss.Color("#D63031"), lipgloss.Color("#FF7675")) }
	ColorWarning = func() color.Color { return lightDark(lipgloss.Color("#FDCB6E"), lipgloss.Color("#FFEAA7")) }
	ColorDim     = func() color.Color { return lightDark(lipgloss.Color("#636E72"), lipgloss.Color("#B2BEC3")) }
	ColorWhite   = func() color.Color { return lightDark(lipgloss.Color("#2D3436"), lipgloss.Color("#DFE6E9")) }
)

// Text styles
var (
	Bold        = lipgloss.NewStyle().Bold(true)
	Dim         = lipgloss.NewStyle().Foreground(ColorDim())
	Primary     = lipgloss.NewStyle().Foreground(ColorPrimary())
	PrimaryBold = lipgloss.NewStyle().Foreground(ColorPrimary()).Bold(true)
	Success     = lipgloss.NewStyle().Foreground(ColorSuccess())
	Error       = lipgloss.NewStyle().Foreground(ColorError())
	Warning     = lipgloss.NewStyle().Foreground(ColorWarning())
)

// IsTTY returns true if stdout is a terminal.
func IsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
