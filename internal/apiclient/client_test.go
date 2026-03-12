// internal/apiclient/client_test.go
package apiclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClientGet(t *testing.T) {
	// Set up a test server that verifies headers and returns a valid response.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth headers are forwarded.
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("expected Authorization header 'Bearer test-token', got %q", got)
		}
		if got := r.Header.Get("x-apideck-app-id"); got != "test-app-id" {
			t.Errorf("expected x-apideck-app-id header 'test-app-id', got %q", got)
		}
		if got := r.Header.Get("x-apideck-consumer-id"); got != "test-consumer" {
			t.Errorf("expected x-apideck-consumer-id header 'test-consumer', got %q", got)
		}

		// Return a typical Apideck paginated response.
		resp := map[string]any{
			"status_code": 200,
			"status":      "OK",
			"data": []map[string]any{
				{"id": "inv-1", "total": 100},
				{"id": "inv-2", "total": 200},
			},
			"meta": map[string]any{
				"items_on_page": 2,
				"cursors": map[string]any{
					"previous": "",
					"current":  "cursor-abc",
					"next":     "cursor-xyz",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := ClientConfig{
		BaseURL: srv.URL,
		Headers: map[string]string{
			"Authorization":          "Bearer test-token",
			"x-apideck-app-id":       "test-app-id",
			"x-apideck-consumer-id":  "test-consumer",
		},
		TimeoutSecs: 5,
	}

	client := NewClient(cfg)
	apiResp, err := client.Do("GET", "/crm/invoices", url.Values{"limit": {"10"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify response normalization.
	if apiResp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", apiResp.StatusCode)
	}
	if !apiResp.Success {
		t.Error("expected Success to be true")
	}
	if apiResp.Error != nil {
		t.Errorf("expected no error, got %+v", apiResp.Error)
	}
	if apiResp.Data == nil {
		t.Error("expected Data to be non-nil")
	}

	// Verify pagination metadata.
	if apiResp.Meta == nil {
		t.Fatal("expected Meta to be non-nil")
	}
	if apiResp.Meta.Cursor != "cursor-xyz" {
		t.Errorf("expected cursor 'cursor-xyz', got %q", apiResp.Meta.Cursor)
	}
	if !apiResp.Meta.HasMore {
		t.Error("expected HasMore to be true")
	}

	// Verify raw body is stored.
	if len(apiResp.RawBody) == 0 {
		t.Error("expected RawBody to be non-empty")
	}
}

func TestClientErrorResponse(t *testing.T) {
	// Set up a test server that returns a 404 with error details.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errResp := map[string]any{
			"status_code": 404,
			"error": map[string]any{
				"status_code": 404,
				"error":       "NotFoundError",
				"message":     "The requested resource was not found",
				"detail":      "No invoice with id 'inv-999' exists",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-apideck-request-id", "req-abc-123")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(errResp)
	}))
	defer srv.Close()

	cfg := ClientConfig{
		BaseURL: srv.URL,
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
		TimeoutSecs: 5,
	}

	client := NewClient(cfg)
	apiResp, err := client.Do("GET", "/crm/invoices/inv-999", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify error normalization.
	if apiResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiResp.StatusCode)
	}
	if apiResp.Success {
		t.Error("expected Success to be false")
	}
	if apiResp.Error == nil {
		t.Fatal("expected Error to be non-nil")
	}
	if apiResp.Error.StatusCode != http.StatusNotFound {
		t.Errorf("expected error status 404, got %d", apiResp.Error.StatusCode)
	}
	if apiResp.Error.Message == "" {
		t.Error("expected error message to be non-empty")
	}
	if apiResp.Error.RequestID != "req-abc-123" {
		t.Errorf("expected request ID 'req-abc-123', got %q", apiResp.Error.RequestID)
	}
	if apiResp.Data != nil {
		t.Error("expected Data to be nil on error response")
	}
}

func TestClientNoMetaResponse(t *testing.T) {
	// Test a single-resource response without pagination (no meta.cursors).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"status_code": 200,
			"status":      "OK",
			"data": map[string]any{
				"id":   "inv-1",
				"total": 100,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := ClientConfig{
		BaseURL:     srv.URL,
		TimeoutSecs: 5,
	}

	client := NewClient(cfg)
	apiResp, err := client.Do("GET", "/crm/invoices/inv-1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !apiResp.Success {
		t.Error("expected Success to be true")
	}
	if apiResp.Meta != nil {
		t.Errorf("expected Meta to be nil when no cursors present, got %+v", apiResp.Meta)
	}
	if apiResp.Data == nil {
		t.Error("expected Data to be non-nil")
	}
}
