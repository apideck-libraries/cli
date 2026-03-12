package test

import (
	"os/exec"
	"strings"
	"testing"
)

func buildCLI(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", "../bin/apideck-test", "../cmd/apideck")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return "../bin/apideck-test"
}

func TestCLIVersion(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "dev") {
		t.Errorf("expected version output, got: %s", out)
	}
}

func TestCLIHelp(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "apideck") {
		t.Errorf("expected help output, got: %s", output)
	}
	// Should list key commands
	for _, cmd := range []string{"sync", "info", "auth", "explore", "history", "permissions"} {
		if !strings.Contains(output, cmd) {
			t.Errorf("help missing command: %s", cmd)
		}
	}
}

func TestCLIListAPIs(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "--list").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	for _, api := range []string{"accounting", "crm", "ats", "hris"} {
		if !strings.Contains(output, api) {
			t.Errorf("expected %s in list, got: %s", api, output)
		}
	}
}

func TestCLIAgentPrompt(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "agent-prompt").CombinedOutput()
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

func TestCLIInfo(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "info").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "Apideck CLI") {
		t.Errorf("expected Apideck CLI in info, got: %s", output)
	}
	if !strings.Contains(output, "Operations") {
		t.Errorf("expected Operations in info, got: %s", output)
	}
}

func TestCLIAccountingList(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "accounting", "--list").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "invoices") {
		t.Errorf("expected invoices in accounting list, got: %s", output)
	}
}

func TestCLIAccountingInvoicesListHelp(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "accounting", "invoices", "list", "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "limit") {
		t.Errorf("expected limit flag in help, got: %s", output)
	}
	if !strings.Contains(output, "cursor") {
		t.Errorf("expected cursor flag in help, got: %s", output)
	}
}

func TestCLIHistory(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "history").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	// Should work even if no history yet
	if !strings.Contains(string(out), "No API calls") && len(out) == 0 {
		t.Errorf("expected history output, got empty")
	}
}

func TestCLIPermissions(t *testing.T) {
	bin := buildCLI(t)
	out, err := exec.Command(bin, "permissions").CombinedOutput()
	if err != nil {
		t.Fatalf("failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "Permission") && !strings.Contains(output, "permission") {
		t.Errorf("expected permission info, got: %s", output)
	}
}
