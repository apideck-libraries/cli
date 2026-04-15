// internal/output/rows.go
package output

import (
	"encoding/json"
	"fmt"
	"sort"
)

// extractRows normalizes data into a slice of row maps and an ordered list of
// field names. When selectedFields is non-empty only those fields are included.
//
// Supported data shapes:
//   - []any where each element is map[string]any
//   - []map[string]any
//   - map[string]any (treated as a single row)
func extractRows(data any, selectedFields []string) ([]map[string]any, []string) {
	var rows []map[string]any

	switch v := data.(type) {
	case []map[string]any:
		rows = v
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				rows = append(rows, m)
			}
		}
	case map[string]any:
		rows = []map[string]any{v}
	default:
		// Unsupported shape -- return a single row with the stringified value.
		rows = []map[string]any{{"value": fmt.Sprintf("%v", data)}}
	}

	// Determine column order.
	fields := selectedFields
	if len(fields) == 0 && len(rows) > 0 {
		// Auto-detect from the first row. Sort scalars before complex types
		// so the most useful columns appear first in the table.
		seen := map[string]bool{}
		for k := range rows[0] {
			if !seen[k] {
				fields = append(fields, k)
				seen[k] = true
			}
		}
		first := rows[0]
		sort.SliceStable(fields, func(i, j int) bool {
			iComplex := isComplex(first[fields[i]])
			jComplex := isComplex(first[fields[j]])
			if iComplex != jComplex {
				return !iComplex // scalars first
			}
			return false // preserve relative order
		})
	}

	return rows, fields
}

// isComplex returns true for maps and slices (nested/array values).
func isComplex(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return true
	}
	return false
}

// formatValue converts a value to a human-readable string suitable for table
// and CSV output. Maps and slices are serialized as compact JSON instead of
// Go's default fmt representation. Nil values render as an empty string.
func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch v.(type) {
	case map[string]any, []any:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}
