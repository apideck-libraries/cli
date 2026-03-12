// internal/output/yaml.go
package output

import (
	"io"

	"github.com/apideck-io/cli/internal/spec"
	"gopkg.in/yaml.v3"
)

// YAMLFormatter formats an APIResponse as YAML.
type YAMLFormatter struct {
	w io.Writer
}

// Format writes the response data as YAML to the writer.
// If resp.Data is non-nil, only the data field is serialized;
// otherwise the full APIResponse struct is serialized.
func (f *YAMLFormatter) Format(resp *spec.APIResponse) error {
	var target any
	if resp.Data != nil {
		target = resp.Data
	} else {
		target = resp
	}

	enc := yaml.NewEncoder(f.w)
	enc.SetIndent(2)
	return enc.Encode(target)
}
