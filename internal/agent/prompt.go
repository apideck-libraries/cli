package agent

import (
	"fmt"
	"sort"
	"strings"

	"github.com/apideck-io/cli/internal/spec"
)

func GlobalPrompt(apiSpec *spec.APISpec) string {
	var groups []string
	for name := range apiSpec.APIGroups {
		groups = append(groups, name)
	}
	sort.Strings(groups)

	return fmt.Sprintf("Use `apideck` to interact with the Apideck Unified API.\n"+
		"Available APIs: `apideck --list`\n"+
		"List resources: `apideck <api> --list`\n"+
		"Operation help: `apideck <api> <resource> <verb> --help`\n"+
		"APIs: %s\n"+
		"Auth is pre-configured. GET auto-approved. POST/PUT/PATCH prompt (use --yes). DELETE blocked (use --force).\n"+
		"Use --service-id <connector> to target a specific integration.\n"+
		"For clean output: -q -o json", strings.Join(groups, ", "))
}

func ScopedPrompt(groupName string, group *spec.APIGroup) string {
	var resources []string
	for name := range group.Resources {
		resources = append(resources, name)
	}
	sort.Strings(resources)

	return fmt.Sprintf("Use `apideck %s` to interact with %s resources.\n"+
		"Resources: %s\n"+
		"List operations: `apideck %s --list`\n"+
		"Details: `apideck %s <resource> <verb> --help`\n"+
		"Auth pre-configured. GET auto-approved. POST/PUT/PATCH: use --yes. DELETE: use --force. For JSON: -q -o json",
		groupName, groupName,
		strings.Join(resources, ", "),
		groupName, groupName)
}
