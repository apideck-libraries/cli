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
	for _, want := range []string{"apideck", "accounting", "--yes", "--force", "-q -o json"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
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
	for _, want := range []string{"apideck accounting", "invoices"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}
