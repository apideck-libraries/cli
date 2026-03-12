package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(ColorPrimary())

	errorBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(ColorError())

	successIcon = Success.Render("✓")
	errorIcon   = Error.Render("✖")
	warningIcon = Warning.Render("⚠")
	spinnerIcon = Primary.Render("⠋")
)

// SuccessMsg renders a success message.
func SuccessMsg(msg string) string {
	return fmt.Sprintf("%s %s", successIcon, msg)
}

// ErrorMsg renders an error message.
func ErrorMsg(msg string) string {
	return fmt.Sprintf("%s %s", errorIcon, Error.Render(msg))
}

// WarningMsg renders a warning message.
func WarningMsg(msg string) string {
	return fmt.Sprintf("%s %s", warningIcon, Warning.Render(msg))
}

// ErrorBox renders an error in a styled box with what/why/fix sections.
func ErrorBox(what, why, fix string) string {
	content := Error.Render(what)
	if why != "" {
		content += "\n" + Dim.Render(why)
	}
	if fix != "" {
		content += "\n" + PrimaryBold.Render(fix)
	}
	return errorBoxStyle.Render(content)
}

// InfoBox renders content in a styled box.
func InfoBox(title, content string) string {
	header := PrimaryBold.Render(title)
	return boxStyle.Render(header + "\n" + content)
}

// StepProgress renders a step with status icon.
func StepProgress(done bool, msg string) string {
	if done {
		return fmt.Sprintf("%s %s", successIcon, msg)
	}
	return fmt.Sprintf("%s %s", spinnerIcon, Dim.Render(msg))
}
