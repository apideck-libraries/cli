# Apideck CLI Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a beautiful, agent-friendly CLI for the Apideck Unified API that parses OpenAPI specs directly.

**Architecture:** Go CLI using Cobra for commands, libopenapi for OpenAPI parsing, Charmbracelet stack (bubbletea/lipgloss/huh) for TUI and beautiful output. Single unified spec from S3, cached as gob-encoded parse tree. Auth via env vars > config file.

**Tech Stack:** Go, Cobra, libopenapi (pb33f), bubbletea, lipgloss, bubbles, huh, glamour, retryablehttp

**Spec:** `docs/superpowers/specs/2026-03-12-apideck-cli-design.md`

---

## Chunk 1: Project Scaffolding + Internal Model + UI Styles

### Task 1: Initialize Go module and dependencies

**Files:**
- Create: `go.mod`
- Create: `cmd/apideck/main.go`
- Create: `Makefile`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/samir.amzani/Projects/apideck/cli
go mod init github.com/apideck-io/cli
```

- [ ] **Step 2: Add all dependencies**

```bash
go get github.com/spf13/cobra@latest
go get github.com/pb33f/libopenapi@latest
go get github.com/charmbracelet/bubbletea/v2@latest
go get github.com/charmbracelet/lipgloss/v2@latest
go get github.com/charmbracelet/bubbles/v2@latest
go get github.com/charmbracelet/huh@latest
go get github.com/charmbracelet/glamour@latest
go get github.com/hashicorp/go-retryablehttp@latest
go get github.com/olekukonez/tablewriter@latest
go get gopkg.in/yaml.v3@latest
```

- [ ] **Step 3: Create minimal main.go**

```go
// cmd/apideck/main.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "apideck",
		Short: "Beautiful, agent-friendly CLI for the Apideck Unified API",
		Long:  "apideck turns the Apideck Unified API into a beautiful, secure, AI-agent-friendly command-line experience.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create Makefile**

```makefile
# Makefile
BINARY_NAME=apideck
VERSION?=dev

.PHONY: build run test clean

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/$(BINARY_NAME) ./cmd/apideck

run: build
	./bin/$(BINARY_NAME)

test:
	go test ./... -v

clean:
	rm -rf bin/
```

- [ ] **Step 5: Verify it builds and runs**

```bash
make build && ./bin/apideck --version
make run
```

Expected: prints version "dev" and help output.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum cmd/ Makefile
git commit -m "feat: initialize Go project with Cobra CLI skeleton"
```

---

### Task 2: Define internal model types

**Files:**
- Create: `internal/spec/model.go`
- Create: `internal/spec/model_test.go`

- [ ] **Step 1: Write test for model types**

```go
// internal/spec/model_test.go
package spec

import "testing"

func TestPermissionLevelFromMethod(t *testing.T) {
	tests := []struct {
		method string
		want   PermissionLevel
	}{
		{"GET", PermissionRead},
		{"HEAD", PermissionRead},
		{"OPTIONS", PermissionRead},
		{"POST", PermissionWrite},
		{"PUT", PermissionWrite},
		{"PATCH", PermissionWrite},
		{"DELETE", PermissionDangerous},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := PermissionLevelFromMethod(tt.method)
			if got != tt.want {
				t.Errorf("PermissionLevelFromMethod(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestOperationVerb(t *testing.T) {
	tests := []struct {
		method string
		hasID  bool
		want   string
	}{
		{"GET", false, "list"},
		{"GET", true, "get"},
		{"POST", false, "create"},
		{"PATCH", false, "update"},
		{"PUT", false, "update"},
		{"DELETE", true, "delete"},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			op := &Operation{Method: tt.method, HasPathID: tt.hasID}
			if got := op.Verb(); got != tt.want {
				t.Errorf("Operation.Verb() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/spec/ -v
```

Expected: FAIL — types not defined.

- [ ] **Step 3: Implement model types**

```go
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
	StatusCode int              `json:"status_code"`
	Success    bool             `json:"success"`
	Data       any              `json:"data,omitempty"`
	Error      *APIError        `json:"error,omitempty"`
	Meta       *ResponseMeta    `json:"meta,omitempty"`
	RawBody    []byte           `json:"-"`
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
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/spec/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/spec/
git commit -m "feat: add internal model types for API spec, operations, permissions"
```

---

### Task 3: Create shared UI styles

**Files:**
- Create: `internal/ui/styles.go`
- Create: `internal/ui/messages.go`

- [ ] **Step 1: Create brand colors and typography**

```go
// internal/ui/styles.go
package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss/v2"
)

// Brand colors (Apideck-inspired)
var (
	ColorPrimary   = lipgloss.AdaptiveColor{Light: "#6C5CE7", Dark: "#A29BFE"}
	ColorSuccess   = lipgloss.AdaptiveColor{Light: "#00B894", Dark: "#55EFC4"}
	ColorError     = lipgloss.AdaptiveColor{Light: "#D63031", Dark: "#FF7675"}
	ColorWarning   = lipgloss.AdaptiveColor{Light: "#FDCB6E", Dark: "#FFEAA7"}
	ColorDim       = lipgloss.AdaptiveColor{Light: "#636E72", Dark: "#B2BEC3"}
	ColorWhite     = lipgloss.AdaptiveColor{Light: "#2D3436", Dark: "#DFE6E9"}
)

// Text styles
var (
	Bold          = lipgloss.NewStyle().Bold(true)
	Dim           = lipgloss.NewStyle().Foreground(ColorDim)
	Primary       = lipgloss.NewStyle().Foreground(ColorPrimary)
	PrimaryBold   = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	Success       = lipgloss.NewStyle().Foreground(ColorSuccess)
	Error         = lipgloss.NewStyle().Foreground(ColorError)
	Warning       = lipgloss.NewStyle().Foreground(ColorWarning)
)

// IsTTY returns true if stdout is a terminal.
func IsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
```

- [ ] **Step 2: Create styled message helpers**

```go
// internal/ui/messages.go
package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss/v2"
)

var (
	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(ColorPrimary)

	errorBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(ColorError)

	successIcon = Success.Render("✓")
	errorIcon   = Error.Render("✖")
	warningIcon = Warning.Render("⚠")
	spinnerIcon = Primary.Render("⠋")
)

// SuccessMsg renders a success message.
func SuccessMsg(msg string) string {
	return fmt.Sprintf("%s %s", successIcon, msg)
}

// ErrorMsg renders an error message.
func ErrorMsg(msg string) string {
	return fmt.Sprintf("%s %s", errorIcon, Error.Render(msg))
}

// WarningMsg renders a warning message.
func WarningMsg(msg string) string {
	return fmt.Sprintf("%s %s", warningIcon, Warning.Render(msg))
}

// ErrorBox renders an error in a styled box with what/why/fix sections.
func ErrorBox(what, why, fix string) string {
	content := Error.Render(what)
	if why != "" {
		content += "\n" + Dim.Render(why)
	}
	if fix != "" {
		content += "\n" + PrimaryBold.Render(fix)
	}
	return errorBoxStyle.Render(content)
}

// InfoBox renders content in a styled box.
func InfoBox(title, content string) string {
	header := PrimaryBold.Render(title)
	return boxStyle.Render(header + "\n" + content)
}

// StepProgress renders a step with status icon.
func StepProgress(done bool, msg string) string {
	if done {
		return fmt.Sprintf("%s %s", successIcon, msg)
	}
	return fmt.Sprintf("%s %s", spinnerIcon, Dim.Render(msg))
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/ui/
```

- [ ] **Step 4: Commit**

```bash
git add internal/ui/
git commit -m "feat: add shared UI styles, brand colors, and message helpers"
```

---

## Chunk 2: OpenAPI Parser + Spec Caching

### Task 4: Download and embed baseline spec

**Files:**
- Create: `specs/speakeasy-spec.yml`
- Create: `internal/spec/embed.go`

- [ ] **Step 1: Download the unified spec**

```bash
curl -s "https://ci-spec-unify.s3.eu-central-1.amazonaws.com/speakeasy-spec.yml" > specs/speakeasy-spec.yml
```

- [ ] **Step 2: Create embed.go**

```go
// internal/spec/embed.go
package spec

import _ "embed"

//go:embed ../../specs/speakeasy-spec.yml
var EmbeddedSpec []byte
```

Wait — `go:embed` paths are relative to the source file. Since `embed.go` is in `internal/spec/`, the path `../../specs/speakeasy-spec.yml` is correct.

Actually, `go:embed` only supports embedding files within the module's own package directory or subdirectories. Cross-package embedding requires the embed file to be in the same directory or a subdirectory.

Fix: move the embed to a top-level package or use a different approach.

```go
// internal/spec/embed.go
package spec

// EmbeddedSpec will be set at init time from the top-level embed package.
var EmbeddedSpec []byte
```

```go
// embed.go (root of project)
package cli

import _ "embed"

//go:embed specs/speakeasy-spec.yml
var EmbeddedSpecData []byte
```

Actually simpler approach — embed from `cmd/apideck/main.go` and pass it down:

```go
// cmd/apideck/main.go — add embed directive
import "embed"

//go:embed ../../specs/speakeasy-spec.yml
```

No — same problem. Let's create a dedicated `specs` Go package:

```go
// specs/embed.go
package specs

import _ "embed"

//go:embed speakeasy-spec.yml
var EmbeddedSpec []byte
```

This works because the YAML file and the Go file are in the same directory.

- [ ] **Step 3: Create specs/embed.go**

```go
// specs/embed.go
package specs

import _ "embed"

//go:embed speakeasy-spec.yml
var EmbeddedSpec []byte
```

- [ ] **Step 4: Verify it compiles**

```bash
go build ./specs/
```

- [ ] **Step 5: Commit**

```bash
git add specs/
git commit -m "feat: embed baseline Apideck OpenAPI spec"
```

---

### Task 5: Implement OpenAPI parser

**Files:**
- Create: `internal/spec/parser.go`
- Create: `internal/spec/parser_test.go`

- [ ] **Step 1: Write parser test**

```go
// internal/spec/parser_test.go
package spec

import (
	"os"
	"testing"
)

func loadTestSpec(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../specs/speakeasy-spec.yml")
	if err != nil {
		t.Fatalf("failed to read spec: %v", err)
	}
	return data
}

func TestParseSpec(t *testing.T) {
	data := loadTestSpec(t)
	apiSpec, err := ParseSpec(data)
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	if apiSpec.Name != "Apideck" {
		t.Errorf("Name = %q, want %q", apiSpec.Name, "Apideck")
	}
	if apiSpec.Version == "" {
		t.Error("Version is empty")
	}
	if apiSpec.BaseURL == "" {
		t.Error("BaseURL is empty")
	}
	if len(apiSpec.APIGroups) == 0 {
		t.Error("APIGroups is empty")
	}
}

func TestParseSpecAPIGroups(t *testing.T) {
	data := loadTestSpec(t)
	apiSpec, err := ParseSpec(data)
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	expectedGroups := []string{"accounting", "ats", "crm", "hris", "ecommerce", "file-storage", "vault"}
	for _, g := range expectedGroups {
		if _, ok := apiSpec.APIGroups[g]; !ok {
			t.Errorf("missing API group: %s", g)
		}
	}
}

func TestParseSpecOperations(t *testing.T) {
	data := loadTestSpec(t)
	apiSpec, err := ParseSpec(data)
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	acct, ok := apiSpec.APIGroups["accounting"]
	if !ok {
		t.Fatal("missing accounting group")
	}

	invoices, ok := acct.Resources["invoices"]
	if !ok {
		t.Fatal("missing invoices resource")
	}

	if len(invoices.Operations) == 0 {
		t.Error("invoices has no operations")
	}

	// Check that operations have correct verbs
	verbs := make(map[string]bool)
	for _, op := range invoices.Operations {
		verbs[op.Verb()] = true
	}
	for _, v := range []string{"list", "get", "create", "update", "delete"} {
		if !verbs[v] {
			t.Errorf("invoices missing verb: %s", v)
		}
	}
}

func TestParseSpecParameters(t *testing.T) {
	data := loadTestSpec(t)
	apiSpec, err := ParseSpec(data)
	if err != nil {
		t.Fatalf("ParseSpec failed: %v", err)
	}

	acct := apiSpec.APIGroups["accounting"]
	invoices := acct.Resources["invoices"]

	// Find the list operation
	var listOp *Operation
	for _, op := range invoices.Operations {
		if op.Verb() == "list" {
			listOp = op
			break
		}
	}
	if listOp == nil {
		t.Fatal("no list operation found")
	}

	if len(listOp.Parameters) == 0 {
		t.Error("list operation has no parameters")
	}

	// Should have limit and cursor parameters
	paramNames := make(map[string]bool)
	for _, p := range listOp.Parameters {
		paramNames[p.Name] = true
	}
	if !paramNames["limit"] {
		t.Error("missing limit parameter")
	}
	if !paramNames["cursor"] {
		t.Error("missing cursor parameter")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/spec/ -v -run TestParse
```

Expected: FAIL — `ParseSpec` not defined.

- [ ] **Step 3: Implement parser**

```go
// internal/spec/parser.go
package spec

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ParseSpec parses an OpenAPI 3.x YAML spec into our internal model.
func ParseSpec(data []byte) (*APISpec, error) {
	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI doc: %w", err)
	}

	v3Model, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to build v3 model: %v", errs[0])
	}

	model := v3Model.Model

	apiSpec := &APISpec{
		Name:        model.Info.Title,
		Version:     model.Info.Version,
		Description: model.Info.Description,
		APIGroups:   make(map[string]*APIGroup),
	}

	if len(model.Servers) > 0 {
		apiSpec.BaseURL = model.Servers[0].URL
	}

	// Iterate through all paths and group by x-apideck-api extension
	if model.Paths == nil {
		return apiSpec, nil
	}

	for pathPair := range model.Paths.PathItems.Iterate() {
		pathName := pathPair.Key()
		pathItem := pathPair.Value()

		operations := map[string]*v3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"DELETE":  pathItem.Delete,
			"PATCH":   pathItem.Patch,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			groupName := extractAPIGroup(operation, pathName)
			resourceName := extractResourceName(pathName, groupName)
			hasPathID := strings.Contains(pathName, "{id}") || strings.Contains(pathName, "{"+resourceName+"_id}")

			// Ensure API group exists
			group, ok := apiSpec.APIGroups[groupName]
			if !ok {
				group = &APIGroup{
					Name:      groupName,
					Resources: make(map[string]*Resource),
				}
				apiSpec.APIGroups[groupName] = group
			}

			// Ensure resource exists
			resource, ok := group.Resources[resourceName]
			if !ok {
				resource = &Resource{
					Name: resourceName,
				}
				group.Resources[resourceName] = resource
			}

			op := &Operation{
				ID:          operation.OperationId,
				Method:      method,
				Path:        pathName,
				Summary:     operation.Summary,
				Description: operation.Description,
				Permission:  PermissionLevelFromMethod(method),
				HasPathID:   hasPathID,
			}

			// Parse parameters (skip internal Apideck headers)
			for _, param := range operation.Parameters {
				if param == nil {
					continue
				}
				// Skip Apideck internal headers (consumerId, applicationId, serviceId, companyId)
				if isInternalParam(param.Name) {
					continue
				}
				p := &Parameter{
					Name:        param.Name,
					In:          param.In,
					Required:    param.Required != nil && *param.Required,
					Description: param.Description,
				}
				if param.Schema != nil {
					schema := param.Schema.Schema()
					if schema != nil {
						if len(schema.Type) > 0 {
							p.Type = schema.Type[0]
						}
						p.Default = schema.Default
						for _, e := range schema.Enum {
							p.Enum = append(p.Enum, fmt.Sprintf("%v", e.Value))
						}
					}
				}
				op.Parameters = append(op.Parameters, p)
			}

			// Parse request body
			if operation.RequestBody != nil {
				op.RequestBody = parseRequestBody(operation.RequestBody)
			}

			resource.Operations = append(resource.Operations, op)
		}
	}

	return apiSpec, nil
}

// extractAPIGroup gets the API group name from x-apideck-api extension or path.
func extractAPIGroup(op *v3.Operation, path string) string {
	if op.Extensions != nil {
		for extPair := range op.Extensions.Iterate() {
			if extPair.Key() == "x-apideck-api" {
				val := strings.Trim(extPair.Value().Value, "\"")
				// Map short names to full names
				switch val {
				case "file":
					return "file-storage"
				case "issue":
					return "issue-tracking"
				default:
					return val
				}
			}
		}
	}
	// Fallback: extract from path
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// extractResourceName extracts the resource name from the path.
func extractResourceName(path, groupName string) string {
	// Path format: /<group>/<resource> or /<group>/<resource>/{id}
	trimmed := strings.TrimPrefix(path, "/")
	parts := strings.Split(trimmed, "/")

	// Skip the group prefix (may be multi-segment like "file-storage")
	// Find the resource part after the group
	groupParts := strings.Split(groupName, "-")
	startIdx := 1
	if len(groupParts) > 1 {
		// For "file-storage", the path is /file-storage/...
		startIdx = 1
	}

	if len(parts) > startIdx {
		resource := parts[startIdx]
		// Remove trailing {id} segments
		if !strings.HasPrefix(resource, "{") {
			return resource
		}
	}
	return "unknown"
}

// isInternalParam checks if a parameter is an Apideck internal header.
func isInternalParam(name string) bool {
	internal := map[string]bool{
		"consumerId":    true,
		"applicationId": true,
		"serviceId":     true,
		"companyId":     true,
		"raw":           true,
	}
	return internal[name]
}

// parseRequestBody extracts fields from a request body schema.
func parseRequestBody(rb *v3.RequestBody) *RequestBody {
	if rb == nil || rb.Content == nil {
		return nil
	}

	reqBody := &RequestBody{}

	for contentPair := range rb.Content.Iterate() {
		contentType := contentPair.Key()
		mediaType := contentPair.Value()

		reqBody.ContentType = contentType
		if rb.Required != nil {
			reqBody.Required = *rb.Required
		}

		if mediaType.Schema != nil {
			schema := mediaType.Schema.Schema()
			if schema != nil && schema.Properties != nil {
				reqBody.Fields = flattenSchema(schema, 0)
			}
		}
		break // Use first content type
	}

	return reqBody
}

// flattenSchema recursively extracts fields from a JSON schema (max depth 3).
func flattenSchema(schema interface{ }, depth int) []*BodyField {
	// This is a simplified version - full implementation would handle
	// all schema types including oneOf/anyOf
	return nil // Placeholder - will be implemented with proper schema walking
}
```

Note: The `flattenSchema` function is a placeholder. The full implementation requires careful handling of libopenapi's schema proxy types. We'll iterate on this in a subsequent task.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/spec/ -v -run TestParse
```

Expected: Most tests PASS. Some may need iteration on the parser logic.

- [ ] **Step 5: Iterate until all tests pass**

Debug any failures by examining the actual spec structure. Common issues:
- `HasPathID` detection may need to check for `{id}` patterns more broadly
- Resource name extraction from paths may need adjustment for the Apideck path structure
- Parameter type extraction from libopenapi's schema proxy

- [ ] **Step 6: Commit**

```bash
git add internal/spec/parser.go internal/spec/parser_test.go
git commit -m "feat: implement OpenAPI parser with libopenapi"
```

---

### Task 6: Implement spec cache

**Files:**
- Create: `internal/spec/cache.go`
- Create: `internal/spec/cache_test.go`

- [ ] **Step 1: Write cache test**

```go
// internal/spec/cache_test.go
package spec

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	apiSpec := &APISpec{
		Name:    "test",
		Version: "1.0.0",
		BaseURL: "https://example.com",
		APIGroups: map[string]*APIGroup{
			"accounting": {
				Name:      "accounting",
				Resources: map[string]*Resource{},
			},
		},
	}

	err := cache.Save(apiSpec, []byte("raw spec data"))
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := cache.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Name != "test" {
		t.Errorf("Name = %q, want %q", loaded.Name, "test")
	}
	if loaded.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", loaded.Version, "1.0.0")
	}
}

func TestCacheIsFresh(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	// No cache yet
	if cache.IsFresh() {
		t.Error("empty cache should not be fresh")
	}

	apiSpec := &APISpec{Name: "test", Version: "1.0.0", APIGroups: map[string]*APIGroup{}}
	cache.Save(apiSpec, []byte("data"))

	if !cache.IsFresh() {
		t.Error("just-saved cache should be fresh")
	}
}

func TestCacheMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	apiSpec := &APISpec{Name: "test", Version: "2.0.0", APIGroups: map[string]*APIGroup{}}
	cache.Save(apiSpec, []byte("data"))

	meta, err := cache.LoadMeta()
	if err != nil {
		t.Fatalf("LoadMeta failed: %v", err)
	}
	if meta.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", meta.Version, "2.0.0")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/spec/ -v -run TestCache
```

- [ ] **Step 3: Implement cache**

```go
// internal/spec/cache.go
package spec

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 24 * time.Hour

// CacheMeta stores metadata about the cached spec.
type CacheMeta struct {
	Version   string    `json:"version"`
	FetchedAt time.Time `json:"fetched_at"`
	Source    string    `json:"source"`
	TTLHours  int       `json:"ttl_hours"`
}

// Cache manages the spec cache on disk.
type Cache struct {
	dir string
}

// NewCache creates a new cache in the given directory.
func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

// DefaultCacheDir returns ~/.apideck-cli/cache/
func DefaultCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".apideck-cli", "cache")
}

// Save writes the parsed spec and raw data to cache atomically.
func (c *Cache) Save(apiSpec *APISpec, rawSpec []byte) error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	// Write raw spec
	if err := atomicWrite(filepath.Join(c.dir, "spec.yml"), rawSpec); err != nil {
		return fmt.Errorf("write raw spec: %w", err)
	}

	// Write parsed model as gob
	parsedPath := filepath.Join(c.dir, "parsed.bin")
	tmpParsed := parsedPath + ".tmp"
	f, err := os.Create(tmpParsed)
	if err != nil {
		return fmt.Errorf("create parsed file: %w", err)
	}
	enc := gob.NewEncoder(f)
	if err := enc.Encode(apiSpec); err != nil {
		f.Close()
		os.Remove(tmpParsed)
		return fmt.Errorf("encode parsed spec: %w", err)
	}
	f.Close()
	if err := os.Rename(tmpParsed, parsedPath); err != nil {
		return fmt.Errorf("rename parsed file: %w", err)
	}

	// Write metadata
	meta := CacheMeta{
		Version:   apiSpec.Version,
		FetchedAt: time.Now(),
		Source:    "https://ci-spec-unify.s3.eu-central-1.amazonaws.com/speakeasy-spec.yml",
		TTLHours:  24,
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	return atomicWrite(filepath.Join(c.dir, "meta.json"), metaBytes)
}

// Load reads the cached parsed spec from disk.
func (c *Cache) Load() (*APISpec, error) {
	parsedPath := filepath.Join(c.dir, "parsed.bin")
	f, err := os.Open(parsedPath)
	if err != nil {
		return nil, fmt.Errorf("open parsed cache: %w", err)
	}
	defer f.Close()

	var apiSpec APISpec
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&apiSpec); err != nil {
		return nil, fmt.Errorf("decode parsed cache: %w", err)
	}
	return &apiSpec, nil
}

// LoadMeta reads the cache metadata.
func (c *Cache) LoadMeta() (*CacheMeta, error) {
	data, err := os.ReadFile(filepath.Join(c.dir, "meta.json"))
	if err != nil {
		return nil, err
	}
	var meta CacheMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// IsFresh returns true if the cache exists and is within TTL.
func (c *Cache) IsFresh() bool {
	meta, err := c.LoadMeta()
	if err != nil {
		return false
	}
	return time.Since(meta.FetchedAt) < cacheTTL
}

// LoadRawSpec reads the raw YAML spec from cache.
func (c *Cache) LoadRawSpec() ([]byte, error) {
	return os.ReadFile(filepath.Join(c.dir, "spec.yml"))
}

// atomicWrite writes data to a file atomically (write to temp, then rename).
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/spec/ -v -run TestCache
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/spec/cache.go internal/spec/cache_test.go
git commit -m "feat: implement spec cache with gob encoding and atomic writes"
```

---

## Chunk 3: Auth Manager + Permission Engine

### Task 7: Implement auth manager

**Files:**
- Create: `internal/auth/manager.go`
- Create: `internal/auth/manager_test.go`
- Create: `internal/auth/config.go`

- [ ] **Step 1: Write auth manager test**

```go
// internal/auth/manager_test.go
package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveFromEnv(t *testing.T) {
	t.Setenv("APIDECK_API_KEY", "test-key")
	t.Setenv("APIDECK_APP_ID", "test-app")
	t.Setenv("APIDECK_CONSUMER_ID", "test-consumer")
	t.Setenv("APIDECK_SERVICE_ID", "quickbooks")

	mgr := NewManager("")
	creds, err := mgr.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if creds.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want %q", creds.APIKey, "test-key")
	}
	if creds.AppID != "test-app" {
		t.Errorf("AppID = %q, want %q", creds.AppID, "test-app")
	}
	if creds.ConsumerID != "test-consumer" {
		t.Errorf("ConsumerID = %q, want %q", creds.ConsumerID, "test-consumer")
	}
	if creds.ServiceID != "quickbooks" {
		t.Errorf("ServiceID = %q, want %q", creds.ServiceID, "quickbooks")
	}
}

func TestResolveFromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte(`api_key: "file-key"
app_id: "file-app"
consumer_id: "file-consumer"
service_id: "xero"
`), 0644)

	// Clear env vars
	t.Setenv("APIDECK_API_KEY", "")
	t.Setenv("APIDECK_APP_ID", "")
	t.Setenv("APIDECK_CONSUMER_ID", "")
	t.Setenv("APIDECK_SERVICE_ID", "")

	mgr := NewManager(configPath)
	creds, err := mgr.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if creds.APIKey != "file-key" {
		t.Errorf("APIKey = %q, want %q", creds.APIKey, "file-key")
	}
}

func TestResolveEnvOverridesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte(`api_key: "file-key"
app_id: "file-app"
consumer_id: "file-consumer"
`), 0644)

	t.Setenv("APIDECK_API_KEY", "env-key")
	t.Setenv("APIDECK_APP_ID", "")
	t.Setenv("APIDECK_CONSUMER_ID", "")

	mgr := NewManager(configPath)
	creds, err := mgr.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Env should override
	if creds.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q (env override)", creds.APIKey, "env-key")
	}
	// Config should fill in the rest
	if creds.AppID != "file-app" {
		t.Errorf("AppID = %q, want %q (from config)", creds.AppID, "file-app")
	}
}

func TestResolveNoCredentials(t *testing.T) {
	t.Setenv("APIDECK_API_KEY", "")
	t.Setenv("APIDECK_APP_ID", "")
	t.Setenv("APIDECK_CONSUMER_ID", "")
	t.Setenv("APIDECK_SERVICE_ID", "")

	mgr := NewManager("/nonexistent/config.yaml")
	_, err := mgr.Resolve()
	if err == nil {
		t.Error("expected error when no credentials available")
	}
}

func TestCredentialsHeaders(t *testing.T) {
	creds := &Credentials{
		APIKey:     "key",
		AppID:      "app",
		ConsumerID: "consumer",
		ServiceID:  "qb",
	}
	headers := creds.Headers()

	if headers["Authorization"] != "Bearer key" {
		t.Errorf("Authorization = %q", headers["Authorization"])
	}
	if headers["x-apideck-app-id"] != "app" {
		t.Errorf("x-apideck-app-id = %q", headers["x-apideck-app-id"])
	}
	if headers["x-apideck-consumer-id"] != "consumer" {
		t.Errorf("x-apideck-consumer-id = %q", headers["x-apideck-consumer-id"])
	}
	if headers["x-apideck-service-id"] != "qb" {
		t.Errorf("x-apideck-service-id = %q", headers["x-apideck-service-id"])
	}
}

func TestCredentialsHeadersNoServiceID(t *testing.T) {
	creds := &Credentials{
		APIKey:     "key",
		AppID:      "app",
		ConsumerID: "consumer",
	}
	headers := creds.Headers()

	if _, ok := headers["x-apideck-service-id"]; ok {
		t.Error("service-id header should be omitted when empty")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/auth/ -v
```

- [ ] **Step 3: Implement auth config**

```go
// internal/auth/config.go
package auth

import (
	"os"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the config.yaml file structure.
type FileConfig struct {
	APIKey     string `yaml:"api_key"`
	AppID      string `yaml:"app_id"`
	ConsumerID string `yaml:"consumer_id"`
	ServiceID  string `yaml:"service_id,omitempty"`
}

// LoadConfig reads the config file.
func LoadConfig(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes the config file.
func SaveConfig(path string, cfg *FileConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	dir := path[:len(path)-len("/config.yaml")]
	os.MkdirAll(dir, 0755)
	return os.WriteFile(path, data, 0600)
}
```

- [ ] **Step 4: Implement auth manager**

```go
// internal/auth/manager.go
package auth

import (
	"fmt"
	"os"
	"path/filepath"
)

// Credentials holds the resolved Apideck credentials.
type Credentials struct {
	APIKey     string
	AppID      string
	ConsumerID string
	ServiceID  string
}

// Headers returns the HTTP headers for authentication.
func (c *Credentials) Headers() map[string]string {
	h := map[string]string{
		"Authorization":        "Bearer " + c.APIKey,
		"x-apideck-app-id":    c.AppID,
		"x-apideck-consumer-id": c.ConsumerID,
	}
	if c.ServiceID != "" {
		h["x-apideck-service-id"] = c.ServiceID
	}
	return h
}

// Manager resolves credentials from env vars and config file.
type Manager struct {
	configPath string
}

// NewManager creates an auth manager. If configPath is empty, uses default.
func NewManager(configPath string) *Manager {
	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".apideck-cli", "config.yaml")
	}
	return &Manager{configPath: configPath}
}

// Resolve returns credentials from env vars > config file.
func (m *Manager) Resolve() (*Credentials, error) {
	creds := &Credentials{}

	// Load config file first (lowest priority)
	cfg, _ := LoadConfig(m.configPath) // ignore error — file may not exist
	if cfg != nil {
		creds.APIKey = cfg.APIKey
		creds.AppID = cfg.AppID
		creds.ConsumerID = cfg.ConsumerID
		creds.ServiceID = cfg.ServiceID
	}

	// Env vars override config
	if v := os.Getenv("APIDECK_API_KEY"); v != "" {
		creds.APIKey = v
	}
	if v := os.Getenv("APIDECK_APP_ID"); v != "" {
		creds.AppID = v
	}
	if v := os.Getenv("APIDECK_CONSUMER_ID"); v != "" {
		creds.ConsumerID = v
	}
	if v := os.Getenv("APIDECK_SERVICE_ID"); v != "" {
		creds.ServiceID = v
	}

	// Validate required fields
	if creds.APIKey == "" {
		return nil, fmt.Errorf("API key not configured. Set APIDECK_API_KEY or run: apideck auth setup")
	}
	if creds.AppID == "" {
		return nil, fmt.Errorf("App ID not configured. Set APIDECK_APP_ID or run: apideck auth setup")
	}
	if creds.ConsumerID == "" {
		return nil, fmt.Errorf("Consumer ID not configured. Set APIDECK_CONSUMER_ID or run: apideck auth setup")
	}

	return creds, nil
}

// ConfigPath returns the path to the config file.
func (m *Manager) ConfigPath() string {
	return m.configPath
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/auth/ -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/auth/
git commit -m "feat: implement auth manager with env var and config file resolution"
```

---

### Task 8: Implement permission engine

**Files:**
- Create: `internal/permission/engine.go`
- Create: `internal/permission/engine_test.go`
- Create: `internal/permission/config.go`

- [ ] **Step 1: Write permission engine test**

```go
// internal/permission/engine_test.go
package permission

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apideck-io/cli/internal/spec"
)

func TestClassify(t *testing.T) {
	engine := NewEngine("")

	tests := []struct {
		method string
		want   Action
	}{
		{"GET", ActionAllow},
		{"POST", ActionPrompt},
		{"DELETE", ActionBlock},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			op := &spec.Operation{Method: tt.method, ID: "test.op"}
			got := engine.Classify(op)
			if got != tt.want {
				t.Errorf("Classify(%s) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "permissions.yaml")
	os.WriteFile(configPath, []byte(`defaults:
  read: allow
  write: prompt
  dangerous: block
overrides:
  accounting.payments.create: dangerous
  crm.contacts.delete: prompt
`), 0644)

	engine := NewEngine(configPath)

	// Override: payments.create upgraded to dangerous (block)
	op1 := &spec.Operation{Method: "POST", ID: "accounting.payments.create"}
	if got := engine.Classify(op1); got != ActionBlock {
		t.Errorf("overridden POST = %v, want block", got)
	}

	// Override: contacts.delete downgraded to write (prompt)
	op2 := &spec.Operation{Method: "DELETE", ID: "crm.contacts.delete"}
	if got := engine.Classify(op2); got != ActionPrompt {
		t.Errorf("overridden DELETE = %v, want prompt", got)
	}

	// No override: normal behavior
	op3 := &spec.Operation{Method: "GET", ID: "accounting.invoices.list"}
	if got := engine.Classify(op3); got != ActionAllow {
		t.Errorf("normal GET = %v, want allow", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/permission/ -v
```

- [ ] **Step 3: Implement permission config**

```go
// internal/permission/config.go
package permission

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PermConfig represents the permissions.yaml structure.
type PermConfig struct {
	Defaults  map[string]string `yaml:"defaults"`
	Overrides map[string]string `yaml:"overrides"`
}

// LoadPermConfig loads the permissions config file.
func LoadPermConfig(path string) (*PermConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg PermConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// DefaultPermConfigPath returns ~/.apideck-cli/permissions.yaml
func DefaultPermConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".apideck-cli", "permissions.yaml")
}
```

- [ ] **Step 4: Implement permission engine**

```go
// internal/permission/engine.go
package permission

import (
	"github.com/apideck-io/cli/internal/spec"
)

// Action represents what happens when an operation is invoked.
type Action int

const (
	ActionAllow  Action = iota // execute immediately
	ActionPrompt               // ask for confirmation
	ActionBlock                // blocked unless --force
)

func (a Action) String() string {
	switch a {
	case ActionAllow:
		return "allow"
	case ActionPrompt:
		return "prompt"
	case ActionBlock:
		return "block"
	default:
		return "unknown"
	}
}

// Engine classifies operations and checks overrides.
type Engine struct {
	config *PermConfig
}

// NewEngine creates a permission engine, loading overrides from config.
func NewEngine(configPath string) *Engine {
	var cfg *PermConfig
	if configPath != "" {
		cfg, _ = LoadPermConfig(configPath)
	}
	return &Engine{config: cfg}
}

// Classify returns the action for an operation.
func (e *Engine) Classify(op *spec.Operation) Action {
	// Check overrides first
	if e.config != nil && e.config.Overrides != nil {
		if override, ok := e.config.Overrides[op.ID]; ok {
			return actionFromLevel(override)
		}
	}

	// Default classification from permission level
	switch op.Permission {
	case spec.PermissionRead:
		return ActionAllow
	case spec.PermissionWrite:
		return ActionPrompt
	case spec.PermissionDangerous:
		return ActionBlock
	default:
		return ActionPrompt
	}
}

func actionFromLevel(level string) Action {
	switch level {
	case "read", "allow":
		return ActionAllow
	case "write", "prompt":
		return ActionPrompt
	case "dangerous", "block":
		return ActionBlock
	default:
		return ActionPrompt
	}
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/permission/ -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/permission/
git commit -m "feat: implement permission engine with override support"
```

---

## Chunk 4: HTTP Client + Output Formatters

### Task 9: Implement HTTP client

**Files:**
- Create: `internal/http/client.go`
- Create: `internal/http/client_test.go`

- [ ] **Step 1: Write HTTP client test**

```go
// internal/http/client_test.go
package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apideck-io/cli/internal/spec"
)

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing auth header")
		}
		if r.Header.Get("x-apideck-app-id") != "test-app" {
			t.Errorf("missing app-id header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status_code":200,"data":[{"id":"inv_1"}],"meta":{"cursors":{"next":"abc"}}}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		Headers: map[string]string{
			"Authorization":          "Bearer test-key",
			"x-apideck-app-id":      "test-app",
			"x-apideck-consumer-id": "test-consumer",
		},
		TimeoutSecs: 30,
	})

	resp, err := client.Do("GET", "/accounting/invoices", nil, nil)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if !resp.Success {
		t.Error("expected Success = true")
	}
}

func TestClientErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"status_code":401,"error":"Unauthorized","message":"Invalid API key"}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		BaseURL:     server.URL,
		Headers:     map[string]string{},
		TimeoutSecs: 30,
	})

	resp, err := client.Do("GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	if resp.Success {
		t.Error("expected Success = false")
	}
	if resp.Error == nil {
		t.Error("expected Error to be set")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/http/ -v
```

- [ ] **Step 3: Implement HTTP client**

```go
// internal/http/client.go
package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/hashicorp/go-retryablehttp"
)

// ClientConfig holds HTTP client configuration.
type ClientConfig struct {
	BaseURL     string
	Headers     map[string]string
	TimeoutSecs int
}

// Client wraps retryablehttp with Apideck-specific behavior.
type Client struct {
	httpClient *retryablehttp.Client
	baseURL    string
	headers    map[string]string
}

// NewClient creates a configured HTTP client.
func NewClient(cfg ClientConfig) *Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 3
	rc.Logger = nil // suppress default logging
	rc.HTTPClient.Timeout = time.Duration(cfg.TimeoutSecs) * time.Second

	// Custom retry policy: retry on 429, 500, 502, 503, 504
	rc.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy

	return &Client{
		httpClient: rc,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		headers:    cfg.Headers,
	}
}

// Do executes an HTTP request and returns a normalized response.
func (c *Client) Do(method, path string, queryParams map[string]string, body []byte) (*spec.APIResponse, error) {
	fullURL := c.baseURL + path

	// Add query parameters
	if len(queryParams) > 0 {
		params := url.Values{}
		for k, v := range queryParams {
			params.Set(k, v)
		}
		fullURL += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, err := retryablehttp.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()
	duration := time.Since(start)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	apiResp := &spec.APIResponse{
		StatusCode: resp.StatusCode,
		Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		RawBody:    respBody,
	}

	// Parse JSON response
	var rawJSON map[string]any
	if err := json.Unmarshal(respBody, &rawJSON); err == nil {
		if data, ok := rawJSON["data"]; ok {
			apiResp.Data = data
		} else {
			apiResp.Data = rawJSON
		}

		// Extract pagination meta
		if meta, ok := rawJSON["meta"].(map[string]any); ok {
			apiResp.Meta = &spec.ResponseMeta{}
			if cursors, ok := meta["cursors"].(map[string]any); ok {
				if next, ok := cursors["next"].(string); ok {
					apiResp.Meta.Cursor = next
					apiResp.Meta.HasMore = next != ""
				}
			}
		}

		// Extract error
		if !apiResp.Success {
			apiResp.Error = &spec.APIError{
				StatusCode: resp.StatusCode,
				RequestID:  resp.Header.Get("x-request-id"),
			}
			if msg, ok := rawJSON["message"].(string); ok {
				apiResp.Error.Message = msg
			}
			if detail, ok := rawJSON["detail"].(string); ok {
				apiResp.Error.Detail = detail
			}
		}
	}

	_ = duration // Will be used for history logging
	return apiResp, nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/http/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/http/
git commit -m "feat: implement HTTP client with retryablehttp and response normalization"
```

---

### Task 10: Implement output formatters

**Files:**
- Create: `internal/output/formatter.go`
- Create: `internal/output/json.go`
- Create: `internal/output/table.go`
- Create: `internal/output/yaml.go`
- Create: `internal/output/csv.go`
- Create: `internal/output/formatter_test.go`

- [ ] **Step 1: Write formatter test**

```go
// internal/output/formatter_test.go
package output

import (
	"bytes"
	"testing"

	"github.com/apideck-io/cli/internal/spec"
)

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{Writer: &buf, Pretty: true}

	resp := &spec.APIResponse{
		StatusCode: 200,
		Success:    true,
		Data:       []map[string]any{{"id": "inv_1", "total": 100}},
	}

	err := f.Format(resp)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("empty output")
	}
}

func TestFormatDispatch(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"json", "*output.JSONFormatter"},
		{"yaml", "*output.YAMLFormatter"},
		{"csv", "*output.CSVFormatter"},
		{"table", "*output.TableFormatter"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewFormatter(tt.format, &buf, nil)
			if f == nil {
				t.Fatalf("NewFormatter(%q) returned nil", tt.format)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/output/ -v
```

- [ ] **Step 3: Implement formatter interface and dispatch**

```go
// internal/output/formatter.go
package output

import (
	"io"

	"github.com/apideck-io/cli/internal/spec"
)

// Formatter formats API responses for output.
type Formatter interface {
	Format(resp *spec.APIResponse) error
}

// NewFormatter creates a formatter for the given format string.
func NewFormatter(format string, w io.Writer, fields []string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{Writer: w, Pretty: true}
	case "yaml":
		return &YAMLFormatter{Writer: w}
	case "csv":
		return &CSVFormatter{Writer: w, Fields: fields}
	case "table":
		return &TableFormatter{Writer: w, Fields: fields}
	default:
		return &JSONFormatter{Writer: w, Pretty: true}
	}
}
```

- [ ] **Step 4: Implement JSON formatter**

```go
// internal/output/json.go
package output

import (
	"encoding/json"
	"io"

	"github.com/apideck-io/cli/internal/spec"
)

// JSONFormatter outputs API responses as JSON.
type JSONFormatter struct {
	Writer io.Writer
	Pretty bool
}

func (f *JSONFormatter) Format(resp *spec.APIResponse) error {
	var data any
	if resp.Data != nil {
		data = resp.Data
	} else {
		data = resp
	}

	enc := json.NewEncoder(f.Writer)
	if f.Pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(data)
}
```

- [ ] **Step 5: Implement YAML formatter**

```go
// internal/output/yaml.go
package output

import (
	"io"

	"github.com/apideck-io/cli/internal/spec"
	"gopkg.in/yaml.v3"
)

// YAMLFormatter outputs API responses as YAML.
type YAMLFormatter struct {
	Writer io.Writer
}

func (f *YAMLFormatter) Format(resp *spec.APIResponse) error {
	var data any
	if resp.Data != nil {
		data = resp.Data
	} else {
		data = resp
	}

	enc := yaml.NewEncoder(f.Writer)
	defer enc.Close()
	return enc.Encode(data)
}
```

- [ ] **Step 6: Implement CSV formatter**

```go
// internal/output/csv.go
package output

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/apideck-io/cli/internal/spec"
)

// CSVFormatter outputs API responses as CSV.
type CSVFormatter struct {
	Writer io.Writer
	Fields []string
}

func (f *CSVFormatter) Format(resp *spec.APIResponse) error {
	w := csv.NewWriter(f.Writer)
	defer w.Flush()

	rows, fields := extractRows(resp.Data, f.Fields)
	if len(rows) == 0 {
		return nil
	}

	// Write header
	w.Write(fields)

	// Write rows
	for _, row := range rows {
		record := make([]string, len(fields))
		for i, field := range fields {
			if v, ok := row[field]; ok {
				record[i] = fmt.Sprintf("%v", v)
			}
		}
		w.Write(record)
	}

	return nil
}

// extractRows converts API response data into a slice of maps.
func extractRows(data any, selectedFields []string) ([]map[string]any, []string) {
	var rows []map[string]any

	switch d := data.(type) {
	case []any:
		for _, item := range d {
			if m, ok := item.(map[string]any); ok {
				rows = append(rows, m)
			}
		}
	case []map[string]any:
		rows = d
	case map[string]any:
		rows = []map[string]any{d}
	default:
		return nil, nil
	}

	if len(rows) == 0 {
		return nil, nil
	}

	// Determine fields
	fields := selectedFields
	if len(fields) == 0 {
		// Auto-detect from first row
		for k := range rows[0] {
			fields = append(fields, k)
		}
	}

	return rows, fields
}
```

- [ ] **Step 7: Implement table formatter**

```go
// internal/output/table.go
package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/charmbracelet/lipgloss/v2"
)

// TableFormatter outputs API responses as styled tables.
type TableFormatter struct {
	Writer io.Writer
	Fields []string
}

func (f *TableFormatter) Format(resp *spec.APIResponse) error {
	rows, fields := extractRows(resp.Data, f.Fields)
	if len(rows) == 0 {
		fmt.Fprintln(f.Writer, ui.Dim.Render("No results."))
		return nil
	}

	// Calculate column widths
	widths := make(map[string]int)
	for _, field := range fields {
		widths[field] = len(field)
	}
	for _, row := range rows {
		for _, field := range fields {
			if v, ok := row[field]; ok {
				s := fmt.Sprintf("%v", v)
				if len(s) > widths[field] {
					widths[field] = len(s)
				}
			}
		}
	}

	// Cap column widths
	for k, v := range widths {
		if v > 40 {
			widths[k] = 40
		}
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary)
	dimStyle := lipgloss.NewStyle().Foreground(ui.ColorDim)

	// Header
	var headerParts []string
	for _, field := range fields {
		headerParts = append(headerParts, headerStyle.Render(padRight(field, widths[field])))
	}
	fmt.Fprintln(f.Writer, strings.Join(headerParts, "  "))

	// Separator
	var sepParts []string
	for _, field := range fields {
		sepParts = append(sepParts, dimStyle.Render(strings.Repeat("─", widths[field])))
	}
	fmt.Fprintln(f.Writer, strings.Join(sepParts, "  "))

	// Rows
	for _, row := range rows {
		var rowParts []string
		for _, field := range fields {
			val := ""
			if v, ok := row[field]; ok {
				val = fmt.Sprintf("%v", v)
			}
			if len(val) > widths[field] {
				val = val[:widths[field]-1] + "…"
			}
			rowParts = append(rowParts, padRight(val, widths[field]))
		}
		fmt.Fprintln(f.Writer, strings.Join(rowParts, "  "))
	}

	// Footer
	fmt.Fprintln(f.Writer)
	footer := fmt.Sprintf("%d results", len(rows))
	if resp.Meta != nil && resp.Meta.HasMore {
		footer += " · More available (use --cursor or --all)"
	}
	fmt.Fprintln(f.Writer, dimStyle.Render(footer))

	return nil
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
```

- [ ] **Step 8: Run tests**

```bash
go test ./internal/output/ -v
```

Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add internal/output/
git commit -m "feat: implement output formatters (JSON, YAML, CSV, table)"
```

---

## Chunk 5: Command Router + Static Commands

### Task 11: Implement dynamic command router

**Files:**
- Create: `internal/router/router.go`
- Create: `internal/router/router_test.go`
- Create: `internal/router/executor.go`

- [ ] **Step 1: Write router test**

```go
// internal/router/router_test.go
package router

import (
	"os"
	"testing"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/spf13/cobra"
)

func testSpec() *spec.APISpec {
	return &spec.APISpec{
		Name:    "Apideck",
		Version: "1.0.0",
		BaseURL: "https://unify.apideck.com",
		APIGroups: map[string]*spec.APIGroup{
			"accounting": {
				Name: "accounting",
				Resources: map[string]*spec.Resource{
					"invoices": {
						Name: "invoices",
						Operations: []*spec.Operation{
							{ID: "accounting.invoices.list", Method: "GET", Path: "/accounting/invoices", Summary: "List invoices", Permission: spec.PermissionRead},
							{ID: "accounting.invoices.get", Method: "GET", Path: "/accounting/invoices/{id}", Summary: "Get invoice", Permission: spec.PermissionRead, HasPathID: true},
							{ID: "accounting.invoices.create", Method: "POST", Path: "/accounting/invoices", Summary: "Create invoice", Permission: spec.PermissionWrite},
							{ID: "accounting.invoices.delete", Method: "DELETE", Path: "/accounting/invoices/{id}", Summary: "Delete invoice", Permission: spec.PermissionDangerous, HasPathID: true},
						},
					},
				},
			},
		},
	}
}

func TestBuildCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "apideck"}
	apiSpec := testSpec()

	BuildCommands(rootCmd, apiSpec, nil)

	// Should have "accounting" subcommand
	acctCmd, _, err := rootCmd.Find([]string{"accounting"})
	if err != nil || acctCmd == nil {
		t.Fatal("accounting command not found")
	}

	// Should have "invoices" under "accounting"
	invCmd, _, err := rootCmd.Find([]string{"accounting", "invoices"})
	if err != nil || invCmd == nil {
		t.Fatal("accounting invoices command not found")
	}

	// Should have verb commands under "invoices"
	for _, verb := range []string{"list", "get", "create", "delete"} {
		cmd, _, err := rootCmd.Find([]string{"accounting", "invoices", verb})
		if err != nil || cmd == nil {
			t.Errorf("accounting invoices %s command not found", verb)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/router/ -v
```

- [ ] **Step 3: Implement router**

```go
// internal/router/router.go
package router

import (
	"fmt"
	"strings"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/spf13/cobra"
)

// ExecutorFunc is called when a dynamic command is run.
type ExecutorFunc func(op *spec.Operation, flags map[string]string) error

// BuildCommands generates Cobra subcommands from the API spec.
func BuildCommands(root *cobra.Command, apiSpec *spec.APISpec, executor ExecutorFunc) {
	for _, group := range apiSpec.APIGroups {
		groupCmd := &cobra.Command{
			Use:   group.Name,
			Short: fmt.Sprintf("Interact with Apideck %s API", group.Name),
		}

		// --list flag for the group
		var listResources bool
		groupCmd.Flags().BoolVar(&listResources, "list", false, "List all resources and operations")
		groupCmd.RunE = func(cmd *cobra.Command, args []string) error {
			if listResources {
				printResourceList(cmd, group)
				return nil
			}
			return cmd.Help()
		}

		for _, resource := range group.Resources {
			resourceCmd := &cobra.Command{
				Use:   resource.Name,
				Short: fmt.Sprintf("Manage %s", resource.Name),
			}

			for _, op := range resource.Operations {
				opCmd := buildOperationCommand(op, executor)
				resourceCmd.AddCommand(opCmd)
			}

			groupCmd.AddCommand(resourceCmd)
		}

		root.AddCommand(groupCmd)
	}
}

func buildOperationCommand(op *spec.Operation, executor ExecutorFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   op.Verb(),
		Short: op.Summary,
		Long:  op.Description,
	}

	// Register flags from parameters
	flagValues := make(map[string]*string)
	for _, param := range op.Parameters {
		if param.In == "header" {
			continue // Skip header params, handled by auth
		}
		val := cmd.Flags().String(param.Name, fmt.Sprintf("%v", paramDefault(param)), param.Description)
		flagValues[param.Name] = val
		if param.Required {
			cmd.MarkFlagRequired(param.Name)
		}
	}

	// --id flag for get/update/delete operations
	var idFlag *string
	if op.HasPathID {
		idFlag = cmd.Flags().String("id", "", "Resource ID (required)")
		cmd.MarkFlagRequired("id")
	}

	// --data flag for create/update operations
	var dataFlag *string
	if op.Method == "POST" || op.Method == "PUT" || op.Method == "PATCH" {
		dataFlag = cmd.Flags().String("data", "", "Raw JSON body or @file.json")
	}

	// Capture op in closure
	operation := op
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if executor == nil {
			return fmt.Errorf("no executor configured")
		}
		flags := make(map[string]string)
		for name, val := range flagValues {
			if val != nil && *val != "" {
				flags[name] = *val
			}
		}
		if idFlag != nil && *idFlag != "" {
			flags["id"] = *idFlag
		}
		if dataFlag != nil && *dataFlag != "" {
			flags["__data"] = *dataFlag
		}
		return executor(operation, flags)
	}

	return cmd
}

func paramDefault(p *spec.Parameter) string {
	if p.Default != nil {
		return fmt.Sprintf("%v", p.Default)
	}
	return ""
}

func printResourceList(cmd *cobra.Command, group *spec.APIGroup) {
	for _, resource := range group.Resources {
		var verbs []string
		for _, op := range resource.Operations {
			verbs = append(verbs, op.Verb())
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-20s %s\n",
			resource.Name, strings.Join(verbs, ", "))
	}
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/router/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/router/
git commit -m "feat: implement dynamic command router from OpenAPI spec"
```

---

### Task 12: Wire up main.go with all components

**Files:**
- Modify: `cmd/apideck/main.go`

- [ ] **Step 1: Wire up the full CLI**

Update `cmd/apideck/main.go` to:
1. Load spec (cache → embedded fallback)
2. Parse spec
3. Create auth manager
4. Create permission engine
5. Create HTTP client
6. Build dynamic commands from spec
7. Add static commands (auth setup, auth status, sync, info, history, permissions, agent-prompt, explore, skill install, completion)

This is the integration point. Each static command should be a separate function that creates a `*cobra.Command`.

```go
// cmd/apideck/main.go
package main

import (
	"fmt"
	"os"

	"github.com/apideck-io/cli/internal/auth"
	apidehttp "github.com/apideck-io/cli/internal/http"
	"github.com/apideck-io/cli/internal/output"
	"github.com/apideck-io/cli/internal/permission"
	"github.com/apideck-io/cli/internal/router"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/apideck-io/cli/specs"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "apideck",
		Short: "Beautiful, agent-friendly CLI for the Apideck Unified API",
		Long:  "apideck turns the Apideck Unified API into a beautiful, secure, AI-agent-friendly command-line experience.",
	}

	rootCmd.Version = version

	// Global flags
	var outputFormat string
	var fieldsFlag string
	var quietFlag bool
	var rawFlag bool
	var serviceIDFlag string
	var yesFlag bool
	var forceFlag bool
	var timeoutFlag int

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format: json|table|yaml|csv")
	rootCmd.PersistentFlags().StringVar(&fieldsFlag, "fields", "", "Comma-separated list of fields to display")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress non-data output")
	rootCmd.PersistentFlags().BoolVar(&rawFlag, "raw", false, "Output raw API response")
	rootCmd.PersistentFlags().StringVar(&serviceIDFlag, "service-id", "", "Target a specific connector")
	rootCmd.PersistentFlags().BoolVar(&yesFlag, "yes", false, "Skip write confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&forceFlag, "force", false, "Override dangerous operation blocks")
	rootCmd.PersistentFlags().IntVar(&timeoutFlag, "timeout", 30, "Request timeout in seconds")

	// Load spec: cache → embedded
	cache := spec.NewCache(spec.DefaultCacheDir())
	var apiSpec *spec.APISpec
	var loadErr error

	if cache.IsFresh() {
		apiSpec, loadErr = cache.Load()
	}
	if apiSpec == nil {
		// Parse embedded spec
		apiSpec, loadErr = spec.ParseSpec(specs.EmbeddedSpec)
		if loadErr != nil {
			fmt.Fprintln(os.Stderr, ui.ErrorBox(
				"Failed to parse OpenAPI spec",
				loadErr.Error(),
				"Try: apideck sync",
			))
			os.Exit(1)
		}
	}

	// --list flag on root
	var listAPIs bool
	rootCmd.Flags().BoolVar(&listAPIs, "list", false, "List available API groups")
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if listAPIs {
			for name := range apiSpec.APIGroups {
				fmt.Println(name)
			}
			return nil
		}
		return cmd.Help()
	}

	// Auth + Permission
	authMgr := auth.NewManager("")
	permEngine := permission.NewEngine(permission.DefaultPermConfigPath())

	// Build executor
	executor := func(op *spec.Operation, flags map[string]string) error {
		// Permission check
		action := permEngine.Classify(op)
		switch action {
		case permission.ActionPrompt:
			if !yesFlag {
				if !ui.IsTTY() {
					return fmt.Errorf("write operation %s requires --yes flag in non-interactive mode", op.ID)
				}
				// TODO: huh confirmation prompt
				fmt.Printf("⚠ Write operation: %s %s\nProceed? [y/N] ", op.Method, op.Path)
			}
		case permission.ActionBlock:
			if !forceFlag {
				fmt.Fprintln(os.Stderr, ui.ErrorBox(
					fmt.Sprintf("Blocked: %s %s", op.Method, op.Path),
					"This operation is classified as dangerous.",
					"Use --force to override, or: apideck permissions",
				))
				os.Exit(1)
			}
		}

		// Resolve auth
		creds, err := authMgr.Resolve()
		if err != nil {
			return fmt.Errorf("auth: %w", err)
		}

		// Apply --service-id override
		if serviceIDFlag != "" {
			creds.ServiceID = serviceIDFlag
		}

		// Create HTTP client
		client := apidehttp.NewClient(apidehttp.ClientConfig{
			BaseURL:     apiSpec.BaseURL,
			Headers:     creds.Headers(),
			TimeoutSecs: timeoutFlag,
		})

		// Build query params and body from flags
		queryParams := make(map[string]string)
		var body []byte
		for k, v := range flags {
			if k == "__data" {
				body = []byte(v) // TODO: handle @file.json
				continue
			}
			if k == "id" {
				continue // handled in path
			}
			queryParams[k] = v
		}

		// Build path with ID substitution
		path := op.Path
		if id, ok := flags["id"]; ok {
			path = strings.Replace(path, "{id}", id, 1)
		}

		// Execute
		resp, err := client.Do(op.Method, path, queryParams, body)
		if err != nil {
			return err
		}

		// Determine output format
		format := outputFormat
		if format == "" {
			if ui.IsTTY() {
				format = "table"
			} else {
				format = "json"
			}
		}

		if rawFlag {
			os.Stdout.Write(resp.RawBody)
			fmt.Println()
			return nil
		}

		// Parse fields
		var fields []string
		if fieldsFlag != "" {
			fields = strings.Split(fieldsFlag, ",")
		}

		formatter := output.NewFormatter(format, os.Stdout, fields)
		return formatter.Format(resp)
	}

	// Build dynamic commands
	router.BuildCommands(rootCmd, apiSpec, executor)

	// Static commands
	rootCmd.AddCommand(newSyncCmd(cache))
	rootCmd.AddCommand(newAgentPromptCmd(apiSpec))
	rootCmd.AddCommand(newInfoCmd(apiSpec, cache))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

Note: The static command builder functions (`newSyncCmd`, `newAgentPromptCmd`, `newInfoCmd`) should be defined in separate files under `cmd/apideck/`. This is intentionally a simplified version — each static command will be its own file.

- [ ] **Step 2: Create static command files**

Create these files under `cmd/apideck/`:
- `cmd/apideck/sync.go` — `newSyncCmd` fetches spec from S3, parses, caches
- `cmd/apideck/agent_prompt.go` — `newAgentPromptCmd` outputs token-optimized prompt
- `cmd/apideck/info.go` — `newInfoCmd` shows spec version, cache status
- `cmd/apideck/auth_cmds.go` — `newAuthCmd` with setup/status subcommands
- `cmd/apideck/history.go` — `newHistoryCmd` shows recent API calls
- `cmd/apideck/permissions_cmd.go` — `newPermissionsCmd` shows/edits permission config

Each command follows the same pattern: create a `*cobra.Command`, return it.

- [ ] **Step 3: Verify it builds**

```bash
make build
```

- [ ] **Step 4: Test basic commands**

```bash
./bin/apideck --version
./bin/apideck --list
./bin/apideck --help
./bin/apideck accounting --list
./bin/apideck accounting invoices list --help
./bin/apideck agent-prompt
```

- [ ] **Step 5: Commit**

```bash
git add cmd/apideck/
git commit -m "feat: wire up CLI with all components and static commands"
```

---

## Chunk 6: Agent Interface + Auth Setup Wizard

### Task 13: Implement agent-prompt command

**Files:**
- Create: `internal/agent/prompt.go`
- Create: `internal/agent/prompt_test.go`

- [ ] **Step 1: Write agent prompt test**

```go
// internal/agent/prompt_test.go
package agent

import (
	"strings"
	"testing"

	"github.com/apideck-io/cli/internal/spec"
)

func TestGlobalPrompt(t *testing.T) {
	apiSpec := &spec.APISpec{
		APIGroups: map[string]*spec.APIGroup{
			"accounting": {Name: "accounting"},
			"crm":        {Name: "crm"},
		},
	}

	prompt := GlobalPrompt(apiSpec)

	if !strings.Contains(prompt, "apideck") {
		t.Error("prompt should mention apideck")
	}
	if !strings.Contains(prompt, "accounting") {
		t.Error("prompt should list accounting")
	}
	if !strings.Contains(prompt, "--yes") {
		t.Error("prompt should mention --yes")
	}
	if !strings.Contains(prompt, "--force") {
		t.Error("prompt should mention --force")
	}
	if !strings.Contains(prompt, "-q -o json") {
		t.Error("prompt should mention quiet json output")
	}
}

func TestScopedPrompt(t *testing.T) {
	group := &spec.APIGroup{
		Name: "accounting",
		Resources: map[string]*spec.Resource{
			"invoices":  {Name: "invoices"},
			"customers": {Name: "customers"},
		},
	}

	prompt := ScopedPrompt("accounting", group)

	if !strings.Contains(prompt, "apideck accounting") {
		t.Error("prompt should mention apideck accounting")
	}
	if !strings.Contains(prompt, "invoices") {
		t.Error("prompt should list invoices")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

- [ ] **Step 3: Implement prompt generation**

```go
// internal/agent/prompt.go
package agent

import (
	"fmt"
	"sort"
	"strings"

	"github.com/apideck-io/cli/internal/spec"
)

// GlobalPrompt generates a token-optimized prompt for all APIs.
func GlobalPrompt(apiSpec *spec.APISpec) string {
	var groups []string
	for name := range apiSpec.APIGroups {
		groups = append(groups, name)
	}
	sort.Strings(groups)

	return fmt.Sprintf(`Use `+"`apideck`"+` to interact with the Apideck Unified API.
Available APIs: `+"`apideck --list`"+`
List resources: `+"`apideck <api> --list`"+`
Operation help: `+"`apideck <api> <resource> <verb> --help`"+`
APIs: %s
Auth is pre-configured. GET auto-approved. POST/PUT/PATCH prompt (use --yes). DELETE blocked (use --force).
Use --service-id <connector> to target a specific integration.
For clean output: -q -o json`, strings.Join(groups, ", "))
}

// ScopedPrompt generates a token-optimized prompt for a single API group.
func ScopedPrompt(groupName string, group *spec.APIGroup) string {
	var resources []string
	for name := range group.Resources {
		resources = append(resources, name)
	}
	sort.Strings(resources)

	return fmt.Sprintf(`Use `+"`apideck %s`"+` to interact with %s resources.
Resources: %s
List operations: `+"`apideck %s --list`"+`
Details: `+"`apideck %s <resource> <verb> --help`"+`
Auth pre-configured. GET auto-approved. POST/PUT/PATCH: use --yes. DELETE: use --force. For JSON: -q -o json`,
		groupName, groupName,
		strings.Join(resources, ", "),
		groupName, groupName)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/agent/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/
git commit -m "feat: implement token-optimized agent prompt generation"
```

---

### Task 14: Implement auth setup wizard

**Files:**
- Create: `internal/auth/setup.go`

- [ ] **Step 1: Implement setup wizard using huh**

```go
// internal/auth/setup.go
package auth

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/apideck-io/cli/internal/ui"
)

// RunSetup launches the interactive auth setup wizard.
func RunSetup(configPath string) error {
	var apiKey, appID, consumerID, serviceID string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("API Key").
				Description("Your Apideck API key").
				Value(&apiKey).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("API key is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("App ID").
				Description("Your Apideck application ID").
				Value(&appID).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("App ID is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Consumer ID").
				Description("The consumer ID to use").
				Value(&consumerID).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("Consumer ID is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Service ID (optional)").
				Description("Default connector (e.g., quickbooks, xero)").
				Value(&serviceID),
		),
	)

	err := form.Run()
	if err != nil {
		return fmt.Errorf("setup cancelled: %w", err)
	}

	cfg := &FileConfig{
		APIKey:     apiKey,
		AppID:      appID,
		ConsumerID: consumerID,
		ServiceID:  serviceID,
	}

	if err := SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Credentials saved to %s", configPath)))
	return nil
}

// RunStatus shows where credentials are sourced from.
func RunStatus(configPath string) {
	mgr := NewManager(configPath)
	creds, err := mgr.Resolve()

	sources := map[string]string{
		"API Key":     checkSource("APIDECK_API_KEY", configPath, "api_key"),
		"App ID":      checkSource("APIDECK_APP_ID", configPath, "app_id"),
		"Consumer ID": checkSource("APIDECK_CONSUMER_ID", configPath, "consumer_id"),
		"Service ID":  checkSource("APIDECK_SERVICE_ID", configPath, "service_id"),
	}

	if err != nil {
		fmt.Println(ui.ErrorMsg(err.Error()))
		return
	}

	_ = creds
	for name, source := range sources {
		if source != "" {
			fmt.Printf("%s %-14s %s\n", ui.Success.Render("✓"), name+":", source)
		} else {
			fmt.Printf("%s %-14s %s\n", ui.Dim.Render("○"), name+":", ui.Dim.Render("not set"))
		}
	}
}

func checkSource(envVar, configPath, configKey string) string {
	if v := getEnv(envVar); v != "" {
		return fmt.Sprintf("from env (%s)", envVar)
	}
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return ""
	}
	switch configKey {
	case "api_key":
		if cfg.APIKey != "" {
			return fmt.Sprintf("from config (%s)", filepath.Base(configPath))
		}
	case "app_id":
		if cfg.AppID != "" {
			return fmt.Sprintf("from config (%s)", filepath.Base(configPath))
		}
	case "consumer_id":
		if cfg.ConsumerID != "" {
			return fmt.Sprintf("from config (%s)", filepath.Base(configPath))
		}
	case "service_id":
		if cfg.ServiceID != "" {
			return fmt.Sprintf("from config (%s)", filepath.Base(configPath))
		}
	}
	return ""
}

func getEnv(key string) string {
	return fmt.Sprintf("%s", key) // placeholder — use os.Getenv in real impl
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/auth/
```

- [ ] **Step 3: Commit**

```bash
git add internal/auth/setup.go
git commit -m "feat: implement interactive auth setup wizard with huh"
```

---

## Chunk 7: TUI Explorer

### Task 15: Implement TUI explorer

**Files:**
- Create: `internal/tui/explorer.go`
- Create: `internal/tui/endpoint_list.go`
- Create: `internal/tui/detail_panel.go`
- Create: `internal/tui/styles.go`

This is the most complex task. The TUI has two panels: resource/operation list on the left, detail view on the right.

- [ ] **Step 1: Create TUI styles**

```go
// internal/tui/styles.go
package tui

import (
	"github.com/apideck-io/cli/internal/ui"
	"github.com/charmbracelet/lipgloss/v2"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.ColorPrimary).
		Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
		Foreground(ui.ColorPrimary).
		Bold(true)

	normalStyle = lipgloss.NewStyle().
		Foreground(ui.ColorWhite)

	dimStyle = lipgloss.NewStyle().
		Foreground(ui.ColorDim)

	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorDim).
		Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(0, 1)

	methodColors = map[string]lipgloss.Style{
		"GET":    lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true),
		"POST":   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0984E3", Dark: "#74B9FF"}).Bold(true),
		"PUT":    lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true),
		"PATCH":  lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true),
		"DELETE": lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true),
	}
)

func methodStyle(method string) lipgloss.Style {
	if s, ok := methodColors[method]; ok {
		return s
	}
	return normalStyle
}
```

- [ ] **Step 2: Create endpoint list model (left panel)**

```go
// internal/tui/endpoint_list.go
package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/apideck-io/cli/internal/spec"
)

// ListItem represents an item in the endpoint list.
type ListItem struct {
	Resource  *spec.Resource
	Operation *spec.Operation // nil for resource headers
	Expanded  bool
	Depth     int
}

// BuildListItems creates a flat list from API groups for display.
func BuildListItems(apiSpec *spec.APISpec, filterGroup string) []ListItem {
	var items []ListItem

	var groupNames []string
	for name := range apiSpec.APIGroups {
		if filterGroup != "" && name != filterGroup {
			continue
		}
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		group := apiSpec.APIGroups[groupName]

		var resourceNames []string
		for name := range group.Resources {
			resourceNames = append(resourceNames, name)
		}
		sort.Strings(resourceNames)

		for _, resName := range resourceNames {
			resource := group.Resources[resName]
			items = append(items, ListItem{
				Resource: resource,
				Depth:    0,
			})
		}
	}

	return items
}

// ExpandResource inserts operation items below a resource.
func ExpandResource(items []ListItem, idx int) []ListItem {
	resource := items[idx].Resource
	items[idx].Expanded = true

	opItems := make([]ListItem, len(resource.Operations))
	for i, op := range resource.Operations {
		opItems[i] = ListItem{
			Resource:  resource,
			Operation: op,
			Depth:     1,
		}
	}

	// Insert after the resource item
	result := make([]ListItem, 0, len(items)+len(opItems))
	result = append(result, items[:idx+1]...)
	result = append(result, opItems...)
	result = append(result, items[idx+1:]...)
	return result
}

// CollapseResource removes operation items below a resource.
func CollapseResource(items []ListItem, idx int) []ListItem {
	items[idx].Expanded = false

	result := []ListItem{items[idx]}
	// Skip operation items
	i := idx + 1
	for i < len(items) && items[i].Depth > 0 {
		i++
	}
	result = append(items[:idx+1], items[i:]...)
	return result
}

// RenderListItem renders a single list item.
func RenderListItem(item ListItem, selected bool) string {
	style := normalStyle
	if selected {
		style = selectedStyle
	}

	if item.Operation != nil {
		// Operation item (indented)
		method := methodStyle(item.Operation.Method).Render(fmt.Sprintf("%-6s", item.Operation.Method))
		verb := style.Render(item.Operation.Verb())
		return fmt.Sprintf("  %s %s", method, verb)
	}

	// Resource item
	prefix := "▸"
	if item.Expanded {
		prefix = "▾"
	}
	return fmt.Sprintf("%s %s", dimStyle.Render(prefix), style.Render(item.Resource.Name))
}
```

- [ ] **Step 3: Create detail panel (right panel)**

```go
// internal/tui/detail_panel.go
package tui

import (
	"fmt"
	"strings"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/charmbracelet/lipgloss/v2"
)

// RenderDetailPanel renders the right panel for an operation.
func RenderDetailPanel(op *spec.Operation, width int) string {
	if op == nil {
		return dimStyle.Render("Select an operation to view details")
	}

	var b strings.Builder

	// Header: METHOD /path
	method := methodStyle(op.Method).Render(op.Method)
	b.WriteString(fmt.Sprintf("%s %s\n", method, op.Path))
	b.WriteString(dimStyle.Render(op.Summary) + "\n\n")

	// Parameters table
	if len(op.Parameters) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("PARAMETERS") + "\n")

		header := fmt.Sprintf("  %-15s %-10s %-10s %s\n",
			dimStyle.Render("Name"),
			dimStyle.Render("Type"),
			dimStyle.Render("Required"),
			dimStyle.Render("Default"))
		b.WriteString(header)
		b.WriteString(dimStyle.Render("  "+strings.Repeat("─", 50)) + "\n")

		for _, p := range op.Parameters {
			req := "no"
			if p.Required {
				req = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D63031", Dark: "#FF7675"}).Render("yes")
			}
			def := "—"
			if p.Default != nil {
				def = fmt.Sprintf("%v", p.Default)
			}
			b.WriteString(fmt.Sprintf("  %-15s %-10s %-10s %s\n",
				p.Name, p.Type, req, def))
		}
		b.WriteString("\n")
	}

	// Permission badge
	perm := op.Permission.String()
	var permBadge string
	switch op.Permission {
	case spec.PermissionRead:
		permBadge = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00B894", Dark: "#55EFC4"}).Render("● read (auto-approved)")
	case spec.PermissionWrite:
		permBadge = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0984E3", Dark: "#74B9FF"}).Render("● write (confirmation required)")
	case spec.PermissionDangerous:
		permBadge = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D63031", Dark: "#FF7675"}).Render("● dangerous (blocked)")
	default:
		permBadge = perm
	}
	b.WriteString(fmt.Sprintf("Permission: %s\n\n", permBadge))

	// Action buttons
	b.WriteString(dimStyle.Render("[Enter: Try it]  [c: Copy curl]  [C: Copy CLI]"))

	return b.String()
}
```

- [ ] **Step 4: Create main explorer model**

```go
// internal/tui/explorer.go
package tui

import (
	"fmt"

	"github.com/apideck-io/cli/internal/spec"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// ExplorerModel is the top-level bubbletea model for the TUI explorer.
type ExplorerModel struct {
	apiSpec     *spec.APISpec
	filterGroup string
	items       []ListItem
	cursor      int
	focusLeft   bool
	width       int
	height      int
	response    string // last response text
}

// NewExplorerModel creates a new explorer model.
func NewExplorerModel(apiSpec *spec.APISpec, filterGroup string) ExplorerModel {
	items := BuildListItems(apiSpec, filterGroup)
	return ExplorerModel{
		apiSpec:     apiSpec,
		filterGroup: filterGroup,
		items:       items,
		focusLeft:   true,
	}
}

func (m ExplorerModel) Init() tea.Cmd {
	return nil
}

func (m ExplorerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if item.Operation == nil {
					// Toggle resource expansion
					if item.Expanded {
						m.items = CollapseResource(m.items, m.cursor)
					} else {
						m.items = ExpandResource(m.items, m.cursor)
					}
				}
			}
		case "left", "h":
			m.focusLeft = true
		case "right", "l":
			m.focusLeft = false
		}
	}
	return m, nil
}

func (m ExplorerModel) View() tea.View {
	if m.width == 0 {
		return tea.NewView("Loading...")
	}

	leftWidth := m.width/3 - 2
	rightWidth := m.width - leftWidth - 6

	// Title bar
	title := titleStyle.Render(fmt.Sprintf("apideck explore · %s v%s",
		m.apiSpec.Name, m.apiSpec.Version))

	// Left panel: endpoint list
	var leftContent string
	leftContent += lipgloss.NewStyle().Bold(true).Render("RESOURCES") + "\n"
	leftContent += dimStyle.Render(strings.Repeat("─", leftWidth)) + "\n"

	visibleStart := 0
	visibleEnd := len(m.items)
	maxVisible := m.height - 8
	if maxVisible > 0 && len(m.items) > maxVisible {
		if m.cursor >= maxVisible {
			visibleStart = m.cursor - maxVisible + 1
		}
		visibleEnd = visibleStart + maxVisible
		if visibleEnd > len(m.items) {
			visibleEnd = len(m.items)
		}
	}

	for i := visibleStart; i < visibleEnd; i++ {
		leftContent += RenderListItem(m.items[i], i == m.cursor) + "\n"
	}

	leftContent += "\n" + dimStyle.Render("/ search  ? help")

	// Right panel: detail view
	var selectedOp *spec.Operation
	if m.cursor < len(m.items) && m.items[m.cursor].Operation != nil {
		selectedOp = m.items[m.cursor].Operation
	}
	rightContent := RenderDetailPanel(selectedOp, rightWidth)

	if m.response != "" {
		rightContent += "\n\n" + lipgloss.NewStyle().Bold(true).Render("RESPONSE") + "\n"
		rightContent += m.response
	}

	// Build panels
	leftStyle := panelStyle.Width(leftWidth)
	rightStyle := panelStyle.Width(rightWidth)
	if m.focusLeft {
		leftStyle = activePanelStyle.Width(leftWidth)
	} else {
		rightStyle = activePanelStyle.Width(rightWidth)
	}

	left := leftStyle.Render(leftContent)
	right := rightStyle.Render(rightContent)

	content := title + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
```

Note: This uses the bubbletea v2 API with `tea.View` return type and `tea.KeyPressMsg`. Need to add `"strings"` import.

- [ ] **Step 5: Verify it compiles**

```bash
go build ./internal/tui/
```

- [ ] **Step 6: Commit**

```bash
git add internal/tui/
git commit -m "feat: implement TUI explorer with two-panel layout"
```

---

## Chunk 8: Distribution + Final Integration

### Task 16: Create Dockerfile

**Files:**
- Create: `Dockerfile`

- [ ] **Step 1: Write Dockerfile**

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w -X main.version=$(cat VERSION 2>/dev/null || echo dev)" -o /apideck ./cmd/apideck

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /apideck /apideck
ENTRYPOINT ["/apideck"]
```

- [ ] **Step 2: Verify Docker build**

```bash
docker build -t apideck-cli .
docker run apideck-cli --version
```

- [ ] **Step 3: Commit**

```bash
git add Dockerfile
git commit -m "feat: add Dockerfile with distroless base image"
```

---

### Task 17: Create goreleaser config

**Files:**
- Create: `.goreleaser.yml`

- [ ] **Step 1: Write goreleaser config**

```yaml
# .goreleaser.yml
version: 2
project_name: apideck

builds:
  - id: apideck
    main: ./cmd/apideck
    binary: apideck
    ldflags:
      - -s -w -X main.version={{.Version}}
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0

archives:
  - id: apideck
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

brews:
  - repository:
      owner: apideck-io
      name: homebrew-tap
    name: apideck
    homepage: "https://github.com/apideck-io/cli"
    description: "Beautiful, agent-friendly CLI for the Apideck Unified API"
    install: |
      bin.install "apideck"

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
```

- [ ] **Step 2: Commit**

```bash
git add .goreleaser.yml
git commit -m "feat: add goreleaser config for multi-platform distribution"
```

---

### Task 18: End-to-end integration test

**Files:**
- Create: `test/integration_test.go`

- [ ] **Step 1: Write integration test**

```go
// test/integration_test.go
package test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCLIVersion(t *testing.T) {
	out, err := exec.Command("go", "run", "../cmd/apideck", "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "dev") {
		t.Errorf("expected version output, got: %s", out)
	}
}

func TestCLIHelp(t *testing.T) {
	out, err := exec.Command("go", "run", "../cmd/apideck", "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "apideck") {
		t.Errorf("expected help output, got: %s", out)
	}
}

func TestCLIListAPIs(t *testing.T) {
	out, err := exec.Command("go", "run", "../cmd/apideck", "--list").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "accounting") {
		t.Errorf("expected accounting in list, got: %s", output)
	}
}

func TestCLIAgentPrompt(t *testing.T) {
	out, err := exec.Command("go", "run", "../cmd/apideck", "agent-prompt").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "apideck") {
		t.Errorf("expected apideck in prompt, got: %s", output)
	}
	if !strings.Contains(output, "--yes") {
		t.Errorf("expected --yes in prompt, got: %s", output)
	}
}

func TestCLIAccountingList(t *testing.T) {
	out, err := exec.Command("go", "run", "../cmd/apideck", "accounting", "--list").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "invoices") {
		t.Errorf("expected invoices in accounting list, got: %s", output)
	}
}
```

- [ ] **Step 2: Run integration tests**

```bash
go test ./test/ -v
```

- [ ] **Step 3: Commit**

```bash
git add test/
git commit -m "test: add integration tests for CLI commands"
```

---

## Chunk 9: Missing Static Commands + Bug Fixes

### Task 19: Fix known code bugs

**Files:**
- Modify: `internal/auth/setup.go`
- Modify: `internal/http/client.go` (rename package to `apiclient`)
- Modify: `cmd/apideck/main.go`

- [ ] **Step 1: Fix `getEnv` bug in setup.go**

Replace the broken `getEnv` function:
```go
// WRONG: return fmt.Sprintf("%s", key)
// CORRECT:
func getEnv(key string) string {
	return os.Getenv(key)
}
```

- [ ] **Step 2: Fix `SaveConfig` path bug**

Replace brittle path calculation:
```go
// WRONG: dir := path[:len(path)-len("/config.yaml")]
// CORRECT:
dir := filepath.Dir(path)
```

- [ ] **Step 3: Rename `internal/http` to `internal/apiclient`**

Rename the package to avoid shadowing `net/http`. Update all imports from `apidehttp "github.com/apideck-io/cli/internal/http"` to `"github.com/apideck-io/cli/internal/apiclient"`.

- [ ] **Step 4: Add missing `strings` imports**

Add `"strings"` to import blocks in:
- `cmd/apideck/main.go`
- `internal/tui/explorer.go`

- [ ] **Step 5: Fix retry policy in HTTP client**

Replace `retryablehttp.ErrorPropagatedRetryPolicy` with a custom policy:
```go
rc.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil {
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}
	switch resp.StatusCode {
	case 429, 500, 502, 503, 504:
		return true, nil
	}
	return false, nil
}
```

- [ ] **Step 6: Guard against empty `schema.Type` slice**

```go
if schema != nil && len(schema.Type) > 0 {
	p.Type = schema.Type[0]
}
```

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "fix: resolve code bugs (getEnv, SaveConfig, retry policy, imports)"
```

---

### Task 20: Implement static commands (history, permissions, completion, skill install, explore)

**Files:**
- Create: `cmd/apideck/history.go`
- Create: `cmd/apideck/permissions_cmd.go`
- Create: `cmd/apideck/auth_cmds.go`
- Create: `cmd/apideck/explore.go`
- Create: `cmd/apideck/skill.go`
- Create: `internal/history/history.go`

- [ ] **Step 1: Implement history logger and command**

```go
// internal/history/history.go
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apideck-io/cli/internal/spec"
)

const maxEntries = 100

// Log appends a history entry with atomic write and FIFO rotation.
func Log(entry spec.HistoryEntry) error {
	path := DefaultPath()
	entries := load(path)
	entries = append(entries, entry)
	if len(entries) > maxEntries {
		entries = entries[len(entries)-maxEntries:]
	}
	data, _ := json.MarshalIndent(entries, "", "  ")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load returns all history entries.
func Load() []spec.HistoryEntry {
	return load(DefaultPath())
}

func load(path string) []spec.HistoryEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []spec.HistoryEntry
	json.Unmarshal(data, &entries)
	return entries
}

// DefaultPath returns ~/.apideck-cli/history.json
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".apideck-cli")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "history.json")
}

// NewEntry creates a history entry from a completed request.
func NewEntry(method, path string, status int, duration time.Duration, serviceID string) spec.HistoryEntry {
	return spec.HistoryEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Method:     method,
		Path:       path,
		Status:     status,
		DurationMs: duration.Milliseconds(),
		ServiceID:  serviceID,
	}
}
```

- [ ] **Step 2: Create `cmd/apideck/history.go`**

```go
// cmd/apideck/history.go
package main

import (
	"fmt"

	"github.com/apideck-io/cli/internal/history"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history",
		Short: "Show recent API calls",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := history.Load()
			if len(entries) == 0 {
				fmt.Println(ui.Dim.Render("No API calls recorded yet."))
				return nil
			}
			fmt.Printf("%-22s %-7s %-40s %s  %s\n",
				ui.Dim.Render("Timestamp"),
				ui.Dim.Render("Method"),
				ui.Dim.Render("Path"),
				ui.Dim.Render("Status"),
				ui.Dim.Render("Duration"))
			for _, e := range entries {
				fmt.Printf("%-22s %-7s %-40s %d     %dms\n",
					e.Timestamp[:19], e.Method, e.Path, e.Status, e.DurationMs)
			}
			return nil
		},
	}
}
```

- [ ] **Step 3: Create `cmd/apideck/auth_cmds.go`**

```go
// cmd/apideck/auth_cmds.go
package main

import (
	"github.com/apideck-io/cli/internal/auth"
	"github.com/spf13/cobra"
)

func newAuthCmd(configPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication credentials",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "setup",
		Short: "Interactive credential setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.RunSetup(configPath)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show credential sources",
		Run: func(cmd *cobra.Command, args []string) {
			auth.RunStatus(configPath)
		},
	})
	return cmd
}
```

- [ ] **Step 4: Create `cmd/apideck/explore.go`**

```go
// cmd/apideck/explore.go
package main

import (
	"fmt"

	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/spf13/cobra"
)

func newExploreCmd(apiSpec *spec.APISpec) *cobra.Command {
	return &cobra.Command{
		Use:   "explore",
		Short: "Launch interactive TUI API explorer",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := tui.NewExplorerModel(apiSpec, "")
			p := tea.NewProgram(m)
			_, err := p.Run()
			if err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		},
	}
}

// newGroupExploreCmd creates an "explore" subcommand scoped to an API group.
func newGroupExploreCmd(apiSpec *spec.APISpec, groupName string) *cobra.Command {
	return &cobra.Command{
		Use:   "explore",
		Short: fmt.Sprintf("Launch TUI explorer for %s API", groupName),
		RunE: func(cmd *cobra.Command, args []string) error {
			m := tui.NewExplorerModel(apiSpec, groupName)
			p := tea.NewProgram(m)
			_, err := p.Run()
			return err
		},
	}
}
```

- [ ] **Step 5: Create `cmd/apideck/skill.go`**

```go
// cmd/apideck/skill.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apideck-io/cli/internal/agent"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/spf13/cobra"
)

func newSkillCmd(apiSpec *spec.APISpec) *cobra.Command {
	cmd := &cobra.Command{Use: "skill", Short: "Manage AI agent skills"}
	cmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install Claude Code skill to ~/.claude/skills/",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			dir := filepath.Join(home, ".claude", "skills")
			os.MkdirAll(dir, 0755)
			path := filepath.Join(dir, "apideck.md")

			content := fmt.Sprintf("---\ndescription: Interact with Apideck Unified API via CLI\n---\n\n%s\n",
				agent.GlobalPrompt(apiSpec))

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Skill installed to %s", path)))
			return nil
		},
	})
	return cmd
}
```

- [ ] **Step 6: Create `cmd/apideck/permissions_cmd.go`**

```go
// cmd/apideck/permissions_cmd.go
package main

import (
	"fmt"
	"os"

	"github.com/apideck-io/cli/internal/permission"
	"github.com/apideck-io/cli/internal/ui"
	"github.com/spf13/cobra"
)

func newPermissionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "permissions",
		Short: "Show permission configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := permission.DefaultPermConfigPath()
			cfg, err := permission.LoadPermConfig(path)
			if err != nil {
				fmt.Println(ui.Dim.Render("No permissions config found."))
				fmt.Println(ui.Dim.Render(fmt.Sprintf("Create one at: %s", path)))
				fmt.Println()
				fmt.Println("Default permissions:")
				fmt.Println("  read (GET):       auto-approved")
				fmt.Println("  write (POST/PUT): confirmation prompt")
				fmt.Println("  dangerous (DELETE): blocked")
				return nil
			}
			fmt.Println(ui.PrimaryBold.Render("Permission Defaults:"))
			for level, action := range cfg.Defaults {
				fmt.Printf("  %-12s %s\n", level+":", action)
			}
			if len(cfg.Overrides) > 0 {
				fmt.Println()
				fmt.Println(ui.PrimaryBold.Render("Overrides:"))
				for op, level := range cfg.Overrides {
					fmt.Printf("  %-40s %s\n", op, level)
				}
			}
			fmt.Printf("\nConfig: %s\n", ui.Dim.Render(path))
			return nil
		},
	}
}
```

- [ ] **Step 7: Add `completion` command**

Add to `main.go`:
```go
rootCmd.AddCommand(&cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletion(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
})
```

- [ ] **Step 8: Wire scoped `agent-prompt` and `explore` into group commands**

In `internal/router/router.go`, update `BuildCommands` to add `agent-prompt` and `explore` subcommands to each API group:
```go
// After building resource commands for each group:
groupCmd.AddCommand(newGroupAgentPromptCmd(group))
groupCmd.AddCommand(newGroupExploreCmd(apiSpec, group.Name))
```

Pass the necessary builder functions into `BuildCommands` or use callback hooks.

- [ ] **Step 9: Commit**

```bash
git add cmd/apideck/ internal/history/
git commit -m "feat: implement all static commands (history, permissions, auth, explore, skill, completion)"
```

---

### Task 21: Implement write confirmation with huh

**Files:**
- Modify: `cmd/apideck/main.go` (executor function)

- [ ] **Step 1: Replace raw fmt.Printf with huh confirm**

```go
import "github.com/charmbracelet/huh"

// In the executor, replace the TODO:
case permission.ActionPrompt:
	if !yesFlag {
		if !ui.IsTTY() {
			return fmt.Errorf("write operation %s requires --yes flag in non-interactive mode", op.ID)
		}
		var confirm bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Write operation: %s %s", op.Method, op.Path)).
			Description(op.Summary).
			Value(&confirm).
			Run()
		if err != nil || !confirm {
			return fmt.Errorf("operation cancelled")
		}
	}
```

- [ ] **Step 2: Commit**

```bash
git add cmd/apideck/main.go
git commit -m "feat: add styled huh confirmation for write operations"
```

---

### Task 22: Implement `--data @file.json` support

**Files:**
- Modify: `cmd/apideck/main.go` (executor function)

- [ ] **Step 1: Handle @file.json syntax**

```go
// In the executor, replace the TODO:
if dataFlag != nil && *dataFlag != "" {
	data := *dataFlag
	if strings.HasPrefix(data, "@") {
		filePath := strings.TrimPrefix(data, "@")
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read data file %s: %w", filePath, err)
		}
		flags["__data"] = string(fileData)
	} else {
		flags["__data"] = data
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add cmd/apideck/main.go
git commit -m "feat: support --data @file.json for request body input"
```

---

### Task 23: Wire history logging into executor

**Files:**
- Modify: `cmd/apideck/main.go`

- [ ] **Step 1: Add history logging after each API call**

```go
import "github.com/apideck-io/cli/internal/history"

// After client.Do() returns successfully:
history.Log(history.NewEntry(op.Method, path, resp.StatusCode, duration, creds.ServiceID))
```

Track duration by recording `time.Now()` before `client.Do()`.

- [ ] **Step 2: Commit**

```bash
git add cmd/apideck/main.go
git commit -m "feat: log API calls to history"
```

---

## Task Dependencies (for parallel execution)

```
Task 1 (scaffold) ─────────────────────────────────────────────┐
    │                                                           │
    ├── Task 2 (model) ──┬── Task 5 (parser) ── Task 6 (cache) │
    │                    │                                      │
    ├── Task 3 (UI)      ├── Task 7 (auth) ── Task 14 (wizard) │
    │                    │                                      │
    │                    ├── Task 8 (permission)                │
    │                    │                                      │
    │                    ├── Task 9 (HTTP client)               │
    │                    │                                      │
    │                    └── Task 10 (output)                   │
    │                                                           │
    │   Task 5+6+7+8+9+10 ──→ Task 11 (router)                │
    │                              │                            │
    │                              ▼                            │
    │                         Task 12 (wire up)                │
    │                              │                            │
    │                              ▼                            │
    │                    Task 19 (bug fixes)                    │
    │                         │                                 │
    │                         ▼                                 │
    │     ┌── Task 20 (static cmds)                            │
    │     ├── Task 21 (huh confirm)                            │
    │     ├── Task 22 (--data @file)                           │
    │     └── Task 23 (history logging)                        │
    │                         │                                 │
    │                         ▼                                 │
    ├── Task 13 (agent prompt) ── depends on Task 2            │
    ├── Task 15 (TUI explorer) ── depends on Tasks 2, 3       │
    ├── Task 16 (Dockerfile) ── independent                    │
    ├── Task 17 (goreleaser) ── independent                    │
    └── Task 18 (integration tests) ── depends on Task 20     │
```

**Parallelizable groups:**
- **Group A** (after Task 2): Tasks 5, 7, 8, 9, 10 can run in parallel
- **Group B** (after Tasks 2, 3): Tasks 13, 15, 16, 17 can run in parallel
- **Group C** (sequential, after Group A): Tasks 11 → 12 → 19
- **Group D** (after Task 19): Tasks 20, 21, 22, 23 can run in parallel
- **Group E** (after Group D): Task 18 (integration tests)
