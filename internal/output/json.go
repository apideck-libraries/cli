// internal/output/json.go
package output

import (
	"encoding/json"
	"io"

	"github.com/apideck-io/cli/internal/spec"
)

// JSONFormatter formats an APIResponse as pretty-printed JSON.
type JSONFormatter struct {
	w io.Writer
}

// Format writes the response data as indented JSON to the writer.
// If resp.Data is non-nil, only the data field is serialized;
// otherwise the full APIResponse struct is serialized.
func (f *JSONFormatter) Format(resp *spec.APIResponse) error {
	var target any
	if resp.Data != nil {
		target = resp.Data
	} else {
		target = resp
	}

	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(target)
}
