// internal/output/table.go
package output

import (
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
)

// TableFormatter formats an APIResponse as a styled terminal table.
type TableFormatter struct {
	w      io.Writer
	fields []string
}

// Format writes the response data as a lipgloss-styled table.
// The header row is rendered in the primary brand color (bold) and the
// separator line is rendered in the dim color.
func (f *TableFormatter) Format(resp *spec.APIResponse) error {
	rows, fields := extractRows(resp.Data, f.fields)
	if len(fields) == 0 {
		return nil
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary())
	dimStyle := lipgloss.NewStyle().Foreground(ui.ColorDim())

	// Calculate column widths (minimum = header width).
	colWidths := make([]int, len(fields))
	for i, h := range fields {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, field := range fields {
			cell := formatValue(row[field])
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Build and print header.
	headerCells := make([]string, len(fields))
	for i, h := range fields {
		headerCells[i] = headerStyle.Render(padRight(h, colWidths[i]))
	}
	fmt.Fprintln(f.w, strings.Join(headerCells, "  "))

	// Separator line.
	sepParts := make([]string, len(fields))
	for i, w := range colWidths {
		sepParts[i] = dimStyle.Render(strings.Repeat("─", w))
	}
	fmt.Fprintln(f.w, strings.Join(sepParts, "  "))

	// Data rows.
	for _, row := range rows {
		cells := make([]string, len(fields))
		for i, field := range fields {
			cell := formatValue(row[field])
			cells[i] = padRight(cell, colWidths[i])
		}
		fmt.Fprintln(f.w, strings.Join(cells, "  "))
	}

	return nil
}

// padRight pads s with trailing spaces until it reaches width w.
func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
