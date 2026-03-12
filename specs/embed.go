// specs/embed.go
package specs

import _ "embed"

//go:embed speakeasy-spec.yml
var EmbeddedSpec []byte
