// internal/spec/parser.go
package spec

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// internalParams are Apideck-specific parameters that should be excluded from CLI-visible params.
var internalParams = map[string]bool{
	"consumerId":    true,
	"applicationId": true,
	"serviceId":     true,
	"companyId":     true,
	"raw":           true,
}

// apiGroupAliases maps short x-apideck-api values to full group names.
var apiGroupAliases = map[string]string{
	"fileStorage":    "file-storage",
	"issueTracking":  "issue-tracking",
}

// hasPathIDRegexp matches paths with {id} or {something_id} patterns.
var hasPathIDRegexp = regexp.MustCompile(`\{[a-zA-Z]*[Ii]d\}`)

// ParseSpec parses an OpenAPI 3.x spec YAML and returns an APISpec.
func ParseSpec(data []byte) (*APISpec, error) {
	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	model, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("failed to build v3 model: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("failed to build v3 model: nil model returned")
	}

	v3doc := model.Model

	apiSpec := &APISpec{
		APIGroups: make(map[string]*APIGroup),
	}

	// Extract info
	if v3doc.Info != nil {
		apiSpec.Name = v3doc.Info.Title
		apiSpec.Version = v3doc.Info.Version
		apiSpec.Description = v3doc.Info.Description
	}

	// Extract base URL from first server
	if len(v3doc.Servers) > 0 {
		apiSpec.BaseURL = v3doc.Servers[0].URL
	}

	// Iterate paths
	if v3doc.Paths != nil && v3doc.Paths.PathItems != nil {
		for path, pathItem := range v3doc.Paths.PathItems.FromOldest() {
			processPathItem(apiSpec, path, pathItem)
		}
	}

	return apiSpec, nil
}

// processPathItem processes all operations in a path item.
func processPathItem(apiSpec *APISpec, path string, pathItem *v3.PathItem) {
	type methodOp struct {
		method    string
		operation *v3.Operation
	}
	ops := []methodOp{
		{"GET", pathItem.Get},
		{"POST", pathItem.Post},
		{"PUT", pathItem.Put},
		{"PATCH", pathItem.Patch},
		{"DELETE", pathItem.Delete},
		{"HEAD", pathItem.Head},
		{"OPTIONS", pathItem.Options},
	}

	for _, mo := range ops {
		if mo.operation == nil {
			continue
		}
		processOperation(apiSpec, path, mo.method, mo.operation)
	}
}

// processOperation extracts info from a single operation and places it in the correct group/resource.
func processOperation(apiSpec *APISpec, path, method string, op *v3.Operation) {
	// Get api group from x-apideck-api extension
	groupName := extractStringExtension(op.Extensions, "x-apideck-api")
	if groupName == "" {
		return
	}

	// Normalize group name using aliases
	if alias, ok := apiGroupAliases[groupName]; ok {
		groupName = alias
	}

	// Determine resource name from x-speakeasy-group (e.g., "accounting.invoices" -> "invoices")
	// Fall back to tags
	resourceName := extractResourceName(op, path)
	if resourceName == "" {
		return
	}

	// Get or create group
	group, ok := apiSpec.APIGroups[groupName]
	if !ok {
		group = &APIGroup{
			Name:      groupName,
			Resources: make(map[string]*Resource),
		}
		apiSpec.APIGroups[groupName] = group
	}

	// Get or create resource
	resource, ok := group.Resources[resourceName]
	if !ok {
		resource = &Resource{
			Name: resourceName,
		}
		group.Resources[resourceName] = resource
	}

	// Build operation
	operation := &Operation{
		ID:          op.OperationId,
		Method:      method,
		Path:        path,
		Summary:     op.Summary,
		Description: op.Description,
		Permission:  PermissionLevelFromMethod(method),
		HasPathID:   hasPathIDRegexp.MatchString(path),
	}

	// Process parameters
	for _, param := range op.Parameters {
		if param == nil {
			continue
		}
		if internalParams[param.Name] {
			continue
		}
		// Skip header params (Authorization, etc.)
		if strings.ToLower(param.In) == "header" {
			continue
		}

		p := &Parameter{
			Name:        param.Name,
			In:          param.In,
			Description: param.Description,
		}

		// Required
		if param.Required != nil {
			p.Required = *param.Required
		}

		// Extract type from schema
		if param.Schema != nil {
			schema := param.Schema.Schema()
			if schema != nil {
				if len(schema.Type) > 0 {
					p.Type = schema.Type[0]
				}
				// Extract enum values
				for _, enumNode := range schema.Enum {
					if enumNode != nil {
						p.Enum = append(p.Enum, enumNode.Value)
					}
				}
				// Extract default
				if schema.Default != nil {
					p.Default = schema.Default.Value
				}
			}
		}

		operation.Parameters = append(operation.Parameters, p)
	}

	// Process request body
	if op.RequestBody != nil {
		rb := op.RequestBody
		reqBody := &RequestBody{}
		if rb.Required != nil {
			reqBody.Required = *rb.Required
		}
		// Get first content type
		if rb.Content != nil {
			for ct := range rb.Content.FromOldest() {
				reqBody.ContentType = ct
				break
			}
		}
		// flattenSchema returns nil for now (known limitation)
		reqBody.Fields = nil
		operation.RequestBody = reqBody
	}

	resource.Operations = append(resource.Operations, operation)
}

// extractStringExtension retrieves a string value from an extensions orderedmap.
func extractStringExtension(extensions *orderedmap.Map[string, *yaml.Node], key string) string {
	if extensions == nil {
		return ""
	}
	node := extensions.GetOrZero(key)
	if node == nil {
		return ""
	}
	return node.Value
}

// extractResourceName derives the resource name from the operation.
// Prefers x-speakeasy-group (e.g., "accounting.invoices" -> "invoices"),
// falls back to the first tag.
func extractResourceName(op *v3.Operation, path string) string {
	if op.Extensions != nil {
		group := extractStringExtension(op.Extensions, "x-speakeasy-group")
		if group != "" {
			parts := strings.SplitN(group, ".", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}

	if len(op.Tags) > 0 {
		return op.Tags[0]
	}

	// Fall back: derive from path segments
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) > 1 {
		return segments[1]
	}
	if len(segments) == 1 {
		return segments[0]
	}
	return ""
}
