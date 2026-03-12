// internal/output/csv.go
package output

import (
	"encoding/csv"
	"io"

	"github.com/apideck-io/cli/internal/spec"
)

// CSVFormatter formats an APIResponse as comma-separated values.
type CSVFormatter struct {
	w      io.Writer
	fields []string
}

// Format writes the response data as CSV with a header row.
// When fields is non-empty only those columns are emitted.
func (f *CSVFormatter) Format(resp *spec.APIResponse) error {
	rows, fields := extractRows(resp.Data, f.fields)
	if len(fields) == 0 {
		return nil
	}

	w := csv.NewWriter(f.w)

	// Header row.
	if err := w.Write(fields); err != nil {
		return err
	}

	// Data rows.
	record := make([]string, len(fields))
	for _, row := range rows {
		for i, field := range fields {
			record[i] = formatValue(row[field])
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
