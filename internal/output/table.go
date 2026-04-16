// internal/output/table.go
package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/charmbracelet/x/term"
)

const (
	colGap     = 2  // spaces between columns
	colAbsMax  = 40 // hard cap per column before budget distribution
	colMin     = 6  // minimum usable column width
	defaultTTY = 120
)

// TableFormatter formats an APIResponse as a styled terminal table.
type TableFormatter struct {
	w       io.Writer
	fields  []string
	widthFn func() int // optional; returns terminal width — nil means auto-detect
}

// Format writes the response data as a lipgloss-styled table.
// The header row is rendered in the primary brand color (bold) and the
// separator line is rendered in the dim color.
func (f *TableFormatter) Format(resp *spec.APIResponse) error {
	rows, fields := extractRows(resp.Data, f.fields)
	if len(fields) == 0 {
		return nil
	}

	tw := f.getWidth()

	// When fields weren't explicitly selected, drop complex (nested object/array)
	// columns if there are too many columns to display readably.
	if len(f.fields) == 0 && len(rows) > 0 {
		fields = pruneComplexFields(fields, rows[0], tw)
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary())
	dimStyle := lipgloss.NewStyle().Foreground(ui.ColorDim())

	// Calculate natural column widths in runes (minimum = header width), capped at colAbsMax.
	colWidths := make([]int, len(fields))
	for i, h := range fields {
		colWidths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, field := range fields {
			cell := formatValue(row[field])
			if n := utf8.RuneCountInString(cell); n > colWidths[i] {
				colWidths[i] = n
			}
		}
	}
	for i := range colWidths {
		if colWidths[i] > colAbsMax {
			colWidths[i] = colAbsMax
		}
	}

	// Shrink columns to fit the terminal width.
	budgetColumns(colWidths, tw)

	// Build and print header.
	headerCells := make([]string, len(fields))
	for i, h := range fields {
		headerCells[i] = headerStyle.Render(padRight(truncate(h, colWidths[i]), colWidths[i]))
	}
	fmt.Fprintln(f.w, strings.Join(headerCells, "  "))

	// Separator line.
	sepParts := make([]string, len(fields))
	for i, w := range colWidths {
		sepParts[i] = dimStyle.Render(strings.Repeat("\u2500", w))
	}
	fmt.Fprintln(f.w, strings.Join(sepParts, "  "))

	// Data rows.
	for _, row := range rows {
		cells := make([]string, len(fields))
		for i, field := range fields {
			cell := formatValue(row[field])
			cells[i] = padRight(truncate(cell, colWidths[i]), colWidths[i])
		}
		fmt.Fprintln(f.w, strings.Join(cells, "  "))
	}

	return nil
}

// getWidth returns the terminal width to use for this formatter.
// If a custom widthFn was injected it is used; otherwise the width is derived
// from f.w when it is an *os.File TTY, falling back to defaultTTY.
func (f *TableFormatter) getWidth() int {
	if f.widthFn != nil {
		return f.widthFn()
	}
	if file, ok := f.w.(*os.File); ok {
		if w, _, err := term.GetSize(file.Fd()); err == nil && w > 0 {
			return w
		}
	}
	return defaultTTY
}

// pruneComplexFields removes complex (nested object/array) fields when the
// total number of columns would make the table unreadable. It keeps all scalar
// fields and only adds complex fields if there's room for at least colReadable
// characters per column.
func pruneComplexFields(fields []string, sampleRow map[string]any, tw int) []string {
	const colReadable = 10 // minimum chars per column to be useful

	// Separate scalar and complex fields.
	var scalar, complex_ []string
	for _, f := range fields {
		if isComplex(sampleRow[f]) {
			complex_ = append(complex_, f)
		} else {
			scalar = append(scalar, f)
		}
	}

	// Start with scalars. Add complex fields one by one if they fit.
	result := scalar
	for _, f := range complex_ {
		n := len(result) + 1
		available := tw - colGap*(n-1)
		if available/n >= colReadable {
			result = append(result, f)
		}
	}

	// If we pruned something, the result is already good.
	// If nothing was pruned, return original to preserve order.
	if len(result) == len(fields) {
		return fields
	}
	return result
}

// budgetColumns shrinks column widths so the total table fits within maxWidth.
// It repeatedly trims the widest column by one until the table fits, but never
// shrinks a column below colMin.
func budgetColumns(widths []int, maxWidth int) {
	n := len(widths)
	if n == 0 {
		return
	}
	gapTotal := colGap * (n - 1)

	for {
		total := gapTotal
		for _, w := range widths {
			total += w
		}
		if total <= maxWidth {
			return
		}

		// Find the widest column that is still above colMin.
		widest := -1
		for i, w := range widths {
			if w > colMin && (widest == -1 || w > widths[widest]) {
				widest = i
			}
		}
		if widest == -1 {
			return // all columns at minimum, nothing more to shrink
		}
		widths[widest]--
	}
}

// truncate shortens s to fit within max runes, adding "\u2026" when trimmed.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "\u2026"
}

// padRight pads s with trailing spaces until it reaches w runes.
func padRight(s string, w int) string {
	n := utf8.RuneCountInString(s)
	if n >= w {
		return s
	}
	return s + strings.Repeat(" ", w-n)
}
