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
	if apiSpec.Name == "" {
		t.Error("Name is empty")
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
	// Should have at least these groups
	for _, g := range []string{"accounting", "ats", "crm", "hris"} {
		if _, ok := apiSpec.APIGroups[g]; !ok {
			t.Errorf("missing API group: %s (found: %v)", g, groupNames(apiSpec))
		}
	}
}

func groupNames(s *APISpec) []string {
	var names []string
	for name := range s.APIGroups {
		names = append(names, name)
	}
	return names
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
	// Should have some resources with operations
	totalOps := 0
	for _, r := range acct.Resources {
		totalOps += len(r.Operations)
	}
	if totalOps == 0 {
		t.Error("accounting has no operations")
	}
}
