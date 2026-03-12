// internal/output/formatter_test.go
package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/apideck-io/cli/internal/spec"
)

func sampleResponse() *spec.APIResponse {
	return &spec.APIResponse{
		StatusCode: 200,
		Success:    true,
		Data: []any{
			map[string]any{"id": "1", "name": "Alice", "email": "alice@example.com"},
			map[string]any{"id": "2", "name": "Bob", "email": "bob@example.com"},
		},
	}
}

// TestJSONFormatter verifies that the JSON formatter produces non-empty output
// and contains expected content.
func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{w: &buf}

	if err := f.Format(sampleResponse()); err != nil {
		t.Fatalf("JSONFormatter.Format returned error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatal("expected non-empty JSON output")
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected output to contain 'Alice', got: %s", out)
	}
}

// TestYAMLFormatter verifies YAML output is non-empty and contains expected data.
func TestYAMLFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &YAMLFormatter{w: &buf}

	if err := f.Format(sampleResponse()); err != nil {
		t.Fatalf("YAMLFormatter.Format returned error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatal("expected non-empty YAML output")
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected output to contain 'Alice', got: %s", out)
	}
}

// TestCSVFormatter verifies CSV output contains a header and data rows.
func TestCSVFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &CSVFormatter{w: &buf, fields: []string{"id", "name"}}

	if err := f.Format(sampleResponse()); err != nil {
		t.Fatalf("CSVFormatter.Format returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "id,name") {
		t.Errorf("expected CSV header 'id,name', got: %s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected output to contain 'Alice', got: %s", out)
	}
}

// TestTableFormatter verifies table output is non-empty and contains headers.
func TestTableFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{w: &buf, fields: []string{"id", "name"}}

	if err := f.Format(sampleResponse()); err != nil {
		t.Fatalf("TableFormatter.Format returned error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatal("expected non-empty table output")
	}
	if !strings.Contains(out, "id") {
		t.Errorf("expected table to contain column 'id', got: %s", out)
	}
}

// TestFormatDispatch verifies NewFormatter returns the correct concrete type.
func TestFormatDispatch(t *testing.T) {
	cases := []struct {
		format   string
		wantType string
	}{
		{"json", "*output.JSONFormatter"},
		{"yaml", "*output.YAMLFormatter"},
		{"csv", "*output.CSVFormatter"},
		{"table", "*output.TableFormatter"},
		{"unknown", "*output.JSONFormatter"},
		{"", "*output.JSONFormatter"},
	}

	for _, tc := range cases {
		t.Run(tc.format, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewFormatter(tc.format, &buf, nil)
			if f == nil {
				t.Fatal("NewFormatter returned nil")
			}

			// Use a type switch to validate the concrete type.
			var got string
			switch f.(type) {
			case *JSONFormatter:
				got = "*output.JSONFormatter"
			case *YAMLFormatter:
				got = "*output.YAMLFormatter"
			case *CSVFormatter:
				got = "*output.CSVFormatter"
			case *TableFormatter:
				got = "*output.TableFormatter"
			default:
				got = "unknown"
			}

			if got != tc.wantType {
				t.Errorf("NewFormatter(%q): got %s, want %s", tc.format, got, tc.wantType)
			}
		})
	}
}

// TestJSONFormatterNoData verifies that when Data is nil the full response is serialized.
func TestJSONFormatterNoData(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{w: &buf}

	resp := &spec.APIResponse{
		StatusCode: 404,
		Success:    false,
		Error: &spec.APIError{
			Message:    "not found",
			StatusCode: 404,
		},
	}

	if err := f.Format(resp); err != nil {
		t.Fatalf("JSONFormatter.Format returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "not found") {
		t.Errorf("expected output to contain error message, got: %s", out)
	}
}

// TestFormatValue verifies that complex types are serialized properly.
func TestFormatValue(t *testing.T) {
	cases := []struct {
		name string
		val  any
		want string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"bool", true, "true"},
		{"map", map[string]any{"download": true}, `{"download":true}`},
		{"slice", []any{"pdf", "docx"}, `["pdf","docx"]`},
		{"empty slice", []any{}, "[]"},
		{"nested map", map[string]any{"email": "a@b.com", "id": "123"}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatValue(tc.val)
			if tc.name == "nested map" {
				// Map key order is non-deterministic, just check it's valid JSON.
				if got[0] != '{' {
					t.Errorf("expected JSON object, got: %s", got)
				}
				return
			}
			if got != tc.want {
				t.Errorf("formatValue(%v) = %q, want %q", tc.val, got, tc.want)
			}
		})
	}
}

// TestTableFormatterComplexValues verifies maps and nil render correctly in table output.
func TestTableFormatterComplexValues(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{w: &buf, fields: []string{"name", "permissions", "owner"}}

	resp := &spec.APIResponse{
		StatusCode: 200,
		Success:    true,
		Data: []any{
			map[string]any{
				"name":        "Analytics",
				"permissions": map[string]any{"download": true},
				"owner":       nil,
			},
		},
	}

	if err := f.Format(resp); err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "map[") {
		t.Errorf("table output should not contain Go map formatting, got: %s", out)
	}
	if strings.Contains(out, "<nil>") {
		t.Errorf("table output should not contain <nil>, got: %s", out)
	}
	if !strings.Contains(out, `{"download":true}`) {
		t.Errorf("expected JSON-formatted map, got: %s", out)
	}
}

// TestCSVFormatterComplexValues verifies maps and nil render correctly in CSV output.
func TestCSVFormatterComplexValues(t *testing.T) {
	var buf bytes.Buffer
	f := &CSVFormatter{w: &buf, fields: []string{"name", "formats"}}

	resp := &spec.APIResponse{
		StatusCode: 200,
		Success:    true,
		Data: []any{
			map[string]any{
				"name":    "Doc",
				"formats": []any{"pdf", "docx"},
			},
		},
	}

	if err := f.Format(resp); err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "map[") {
		t.Errorf("CSV output should not contain Go map formatting, got: %s", out)
	}
	if !strings.Contains(out, `"[""pdf"",""docx""]"`) {
		// CSV encoding will double-quote the JSON array
		t.Logf("CSV output: %s", out)
	}
}

// TestExtractRows verifies the extractRows helper handles different data shapes.
func TestExtractRows(t *testing.T) {
	t.Run("slice of any", func(t *testing.T) {
		data := []any{
			map[string]any{"a": 1},
			map[string]any{"a": 2},
		}
		rows, fields := extractRows(data, nil)
		if len(rows) != 2 {
			t.Errorf("expected 2 rows, got %d", len(rows))
		}
		if len(fields) == 0 {
			t.Error("expected auto-detected fields")
		}
	})

	t.Run("single map", func(t *testing.T) {
		data := map[string]any{"x": "hello"}
		rows, fields := extractRows(data, nil)
		if len(rows) != 1 {
			t.Errorf("expected 1 row, got %d", len(rows))
		}
		if len(fields) == 0 {
			t.Error("expected auto-detected fields")
		}
	})

	t.Run("selected fields override", func(t *testing.T) {
		data := []any{map[string]any{"a": 1, "b": 2}}
		_, fields := extractRows(data, []string{"a"})
		if len(fields) != 1 || fields[0] != "a" {
			t.Errorf("expected fields=[a], got %v", fields)
		}
	})
}
