package tui

import (
	"fmt"
	"strings"

	"github.com/apideck-io/cli/internal/spec"
)

// ListItemKind distinguishes between resource headers and operation rows.
type ListItemKind int

const (
	ListItemResource  ListItemKind = iota
	ListItemOperation ListItemKind = iota
)

// ListItem represents a single row in the left-panel list.
type ListItem struct {
	Kind        ListItemKind
	ResourceKey string
	Resource    *spec.Resource
	Operation   *spec.Operation
	Depth       int
	Expanded    bool
}

// BuildListItems constructs the flat list of visible items from the spec
// resources map, respecting the expanded state tracked in expandedResources.
func BuildListItems(resources map[string]*spec.Resource, expandedResources map[string]bool) []ListItem {
	if len(resources) == 0 {
		return nil
	}

	// Sort resource keys for stable ordering.
	keys := make([]string, 0, len(resources))
	for k := range resources {
		keys = append(keys, k)
	}
	sortStrings(keys)

	var items []ListItem
	for _, key := range keys {
		res := resources[key]
		expanded := expandedResources[key]
		items = append(items, ListItem{
			Kind:        ListItemResource,
			ResourceKey: key,
			Resource:    res,
			Expanded:    expanded,
		})
		if expanded {
			for _, op := range res.Operations {
				items = append(items, ListItem{
					Kind:        ListItemOperation,
					ResourceKey: key,
					Resource:    res,
					Operation:   op,
					Depth:       1,
				})
			}
		}
	}
	return items
}

// ExpandResource marks a resource as expanded.
func ExpandResource(expandedResources map[string]bool, key string) map[string]bool {
	result := make(map[string]bool, len(expandedResources)+1)
	for k, v := range expandedResources {
		result[k] = v
	}
	result[key] = true
	return result
}

// CollapseResource marks a resource as collapsed.
func CollapseResource(expandedResources map[string]bool, key string) map[string]bool {
	result := make(map[string]bool, len(expandedResources))
	for k, v := range expandedResources {
		result[k] = v
	}
	result[key] = false
	return result
}

// RenderListItem renders a single list item as a styled string.
// selected indicates whether this item is the currently focused row.
func RenderListItem(item ListItem, selected bool, width int) string {
	if item.Kind == ListItemResource {
		arrow := "▶"
		if item.Expanded {
			arrow = "▼"
		}
		label := fmt.Sprintf("%s %s", arrow, item.Resource.Name)
		if selected {
			return selectedStyle.Width(width).Render(label)
		}
		return normalStyle.Width(width).Render(label)
	}

	// Operation row
	indent := strings.Repeat("  ", item.Depth)
	method := fmt.Sprintf("%-6s", item.Operation.Method)
	methodRendered := methodStyle(item.Operation.Method).Render(method)
	verb := item.Operation.Verb()
	line := fmt.Sprintf("%s%s %s", indent, methodRendered, verb)

	if selected {
		return selectedStyle.Width(width).Render(line)
	}
	return normalStyle.Width(width).Render(line)
}

// sortStrings sorts a string slice in-place (simple insertion sort to avoid
// importing sort for a small slice).
func sortStrings(ss []string) {
	for i := 1; i < len(ss); i++ {
		key := ss[i]
		j := i - 1
		for j >= 0 && ss[j] > key {
			ss[j+1] = ss[j]
			j--
		}
		ss[j+1] = key
	}
}
