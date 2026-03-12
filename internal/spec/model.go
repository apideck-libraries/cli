// internal/spec/model.go
package spec

import "strings"

// PermissionLevel classifies operations by safety.
type PermissionLevel int

const (
	PermissionRead      PermissionLevel = iota // GET, HEAD, OPTIONS
	PermissionWrite                            // POST, PUT, PATCH
	PermissionDangerous                        // DELETE
)

func (p PermissionLevel) String() string {
	switch p {
	case PermissionRead:
		return "read"
	case PermissionWrite:
		return "write"
	case PermissionDangerous:
		return "dangerous"
	default:
		return "unknown"
	}
}

// PermissionLevelFromMethod derives permission from HTTP method.
func PermissionLevelFromMethod(method string) PermissionLevel {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return PermissionRead
	case "POST", "PUT", "PATCH":
		return PermissionWrite
	case "DELETE":
		return PermissionDangerous
	default:
		return PermissionWrite
	}
}

// APISpec is the top-level parsed representation of the Apideck API.
type APISpec struct {
	Name        string
	Version     string
	BaseURL     string
	Description string
	APIGroups   map[string]*APIGroup
}

// APIGroup represents a domain like "accounting" or "crm".
type APIGroup struct {
	Name        string
	Description string
	Resources   map[string]*Resource
}

// Resource represents a collection like "invoices" or "customers".
type Resource struct {
	Name        string
	Description string
	Operations  []*Operation
}

// Operation represents a single API endpoint.
type Operation struct {
	ID          string
	Method      string
	Path        string
	Summary     string
	Description string
	Parameters  []*Parameter
	RequestBody *RequestBody
	Permission  PermissionLevel
	HasPathID   bool // true if path contains {id} parameter
}

// Verb returns the CLI verb for this operation (list, get, create, update, delete).
func (o *Operation) Verb() string {
	switch strings.ToUpper(o.Method) {
	case "GET":
		if o.HasPathID {
			return "get"
		}
		return "list"
	case "POST":
		return "create"
	case "PATCH", "PUT":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(o.Method)
	}
}

// Parameter represents a query, path, or header parameter.
type Parameter struct {
	Name        string
	In          string // "query", "path", "header"
	Type        string // "string", "integer", "boolean"
	Required    bool
	Default     any
	Description string
	Enum        []string
}

// RequestBody represents the request body for POST/PUT/PATCH.
type RequestBody struct {
	ContentType string
	Fields      []*BodyField
	Required    bool
}

// BodyField represents a field in a request body JSON schema.
type BodyField struct {
	Name        string
	Type        string // "string", "integer", "boolean", "object", "array"
	Required    bool
	Default     any
	Description string
	Enum        []string
	Items       *BodyField   // for array types
	Children    []*BodyField // for nested objects (max depth 3)
	Format      string       // "date", "date-time", "email", etc.
}

// APIResponse is the normalized response from an API call.
type APIResponse struct {
	StatusCode int           `json:"status_code"`
	Success    bool          `json:"success"`
	Data       any           `json:"data,omitempty"`
	Error      *APIError     `json:"error,omitempty"`
	Meta       *ResponseMeta `json:"meta,omitempty"`
	RawBody    []byte        `json:"-"`
}

// APIError represents a normalized API error.
type APIError struct {
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
	StatusCode int    `json:"status_code"`
	RequestID  string `json:"request_id,omitempty"`
}

// ResponseMeta contains pagination and rate limit info.
type ResponseMeta struct {
	Cursor     string `json:"cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	RateLimit  int    `json:"rate_limit,omitempty"`
	RatePeriod int    `json:"rate_period,omitempty"`
}

// HistoryEntry records a single API call.
type HistoryEntry struct {
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	ServiceID  string `json:"service_id,omitempty"`
}
