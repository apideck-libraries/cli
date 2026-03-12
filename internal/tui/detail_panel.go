package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
)

// permissionBadge returns a styled permission badge string for the operation.
func permissionBadge(op *spec.Operation) string {
	switch op.Permission {
	case spec.PermissionRead:
		return lipgloss.NewStyle().
			Foreground(ui.ColorSuccess()).
			Bold(true).
			Render("● READ")
	case spec.PermissionWrite:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#74B9FF")).
			Bold(true).
			Render("● WRITE")
	case spec.PermissionDangerous:
		return lipgloss.NewStyle().
			Foreground(ui.ColorError()).
			Bold(true).
			Render("● DANGEROUS")
	default:
		return dimStyle.Render("● UNKNOWN")
	}
}

// RenderDetailPanel renders the right-side detail panel for the selected operation.
// width and height are the inner dimensions of the panel (excluding border).
func RenderDetailPanel(op *spec.Operation, width, height int) string {
	if op == nil {
		return dimStyle.Render("Select an operation to see details.")
	}

	var sb strings.Builder

	// Method + path header
	methodStr := methodStyle(op.Method).Render(fmt.Sprintf("%-7s", op.Method))
	pathStr := normalStyle.Render(op.Path)
	sb.WriteString(methodStr + " " + pathStr + "\n")

	// Permission badge
	sb.WriteString(permissionBadge(op) + "\n\n")

	// Summary / description
	if op.Summary != "" {
		sb.WriteString(titleStyle.Render("Summary") + "\n")
		sb.WriteString(normalStyle.Render(wrapText(op.Summary, width)) + "\n\n")
	}
	if op.Description != "" {
		sb.WriteString(titleStyle.Render("Description") + "\n")
		sb.WriteString(dimStyle.Render(wrapText(op.Description, width)) + "\n\n")
	}

	// Parameters table
	queryParams := filterParams(op.Parameters, "query")
	pathParams := filterParams(op.Parameters, "path")
	headerParams := filterParams(op.Parameters, "header")

	if len(pathParams) > 0 {
		sb.WriteString(titleStyle.Render("Path Parameters") + "\n")
		sb.WriteString(renderParamTable(pathParams, width))
		sb.WriteString("\n")
	}
	if len(queryParams) > 0 {
		sb.WriteString(titleStyle.Render("Query Parameters") + "\n")
		sb.WriteString(renderParamTable(queryParams, width))
		sb.WriteString("\n")
	}
	if len(headerParams) > 0 {
		sb.WriteString(titleStyle.Render("Header Parameters") + "\n")
		sb.WriteString(renderParamTable(headerParams, width))
		sb.WriteString("\n")
	}

	// Request body
	if op.RequestBody != nil && len(op.RequestBody.Fields) > 0 {
		sb.WriteString(titleStyle.Render("Request Body") + "\n")
		sb.WriteString(renderBodyFields(op.RequestBody.Fields, width))
	}

	return sb.String()
}

func filterParams(params []*spec.Parameter, in string) []*spec.Parameter {
	var result []*spec.Parameter
	for _, p := range params {
		if p.In == in {
			result = append(result, p)
		}
	}
	return result
}

func renderParamTable(params []*spec.Parameter, width int) string {
	var sb strings.Builder
	colName := 20
	colType := 10
	colReq := 6

	header := fmt.Sprintf("  %-*s %-*s %-*s %s",
		colName, "NAME",
		colType, "TYPE",
		colReq, "REQ",
		"DESCRIPTION",
	)
	sb.WriteString(dimStyle.Render(header) + "\n")

	sep := "  " + strings.Repeat("─", minInt(width-2, 60))
	sb.WriteString(dimStyle.Render(sep) + "\n")

	for _, p := range params {
		req := "no"
		if p.Required {
			req = "yes"
		}
		desc := truncate(p.Description, width-colName-colType-colReq-6)
		row := fmt.Sprintf("  %-*s %-*s %-*s %s",
			colName, truncate(p.Name, colName),
			colType, truncate(p.Type, colType),
			colReq, req,
			desc,
		)
		sb.WriteString(normalStyle.Render(row) + "\n")
	}
	return sb.String()
}

func renderBodyFields(fields []*spec.BodyField, width int) string {
	var sb strings.Builder
	colName := 22
	colType := 10
	colReq := 6

	header := fmt.Sprintf("  %-*s %-*s %-*s %s",
		colName, "FIELD",
		colType, "TYPE",
		colReq, "REQ",
		"DESCRIPTION",
	)
	sb.WriteString(dimStyle.Render(header) + "\n")
	sep := "  " + strings.Repeat("─", minInt(width-2, 60))
	sb.WriteString(dimStyle.Render(sep) + "\n")

	for _, f := range fields {
		req := "no"
		if f.Required {
			req = "yes"
		}
		desc := truncate(f.Description, width-colName-colType-colReq-6)
		row := fmt.Sprintf("  %-*s %-*s %-*s %s",
			colName, truncate(f.Name, colName),
			colType, truncate(f.Type, colType),
			colReq, req,
			desc,
		)
		sb.WriteString(normalStyle.Render(row) + "\n")
	}
	return sb.String()
}

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	words := strings.Fields(s)
	var lines []string
	var current strings.Builder
	for _, w := range words {
		if current.Len()+len(w)+1 > width && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "\n")
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
