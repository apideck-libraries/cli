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
