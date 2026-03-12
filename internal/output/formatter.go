// internal/output/formatter.go
package output

import (
	"io"

	"github.com/apideck-io/cli/internal/spec"
)

// Formatter is the interface all output formatters implement.
type Formatter interface {
	Format(resp *spec.APIResponse) error
}

// NewFormatter returns the appropriate Formatter for the given format string.
// Supported formats: "json", "yaml", "csv", "table".
// Falls back to JSON for unrecognized formats.
func NewFormatter(format string, w io.Writer, fields []string) Formatter {
	switch format {
	case "yaml":
		return &YAMLFormatter{w: w}
	case "csv":
		return &CSVFormatter{w: w, fields: fields}
	case "table":
		return &TableFormatter{w: w, fields: fields}
	default:
		return &JSONFormatter{w: w}
	}
}
