// internal/apiclient/client.go
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"

	"github.com/apideck-io/cli/internal/spec"
)

// ClientConfig holds configuration for the HTTP client.
type ClientConfig struct {
	BaseURL    string
	Headers    map[string]string
	TimeoutSecs int
}

// Client wraps retryablehttp.Client with Apideck-specific behavior.
type Client struct {
	cfg        ClientConfig
	retryClient *retryablehttp.Client
}

// NewClient creates a new Client with a custom retry policy.
func NewClient(cfg ClientConfig) *Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 3
	rc.RetryWaitMin = 1 * time.Second
	rc.RetryWaitMax = 10 * time.Second
	rc.Logger = nil // suppress default retryablehttp logging

	timeout := 30
	if cfg.TimeoutSecs > 0 {
		timeout = cfg.TimeoutSecs
	}
	rc.HTTPClient = &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

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

	return &Client{
		cfg:         cfg,
		retryClient: rc,
	}
}

// Do executes an HTTP request and returns a normalized APIResponse.
func (c *Client) Do(method, path string, queryParams url.Values, body any) (*spec.APIResponse, error) {
	fullURL := c.cfg.BaseURL + path
	if len(queryParams) > 0 {
		fullURL = fullURL + "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := retryablehttp.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Inject Apideck auth headers
	for k, v := range c.cfg.Headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.retryClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	apiResp := &spec.APIResponse{
		StatusCode: resp.StatusCode,
		Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		RawBody:    rawBody,
	}

	if apiResp.Success {
		apiResp.Data, apiResp.Meta = extractDataAndMeta(rawBody)
	} else {
		apiResp.Error = extractError(rawBody, resp.StatusCode, resp.Header.Get("x-apideck-request-id"))
	}

	return apiResp, nil
}

// apideckEnvelope is the common Apideck response envelope.
type apideckEnvelope struct {
	StatusCode int             `json:"status_code"`
	Status     string          `json:"status"`
	Data       json.RawMessage `json:"data"`
	Meta       *apideckMeta    `json:"meta"`
	Error      *apideckError   `json:"error"`
	Message    string          `json:"message"`
	Detail     any             `json:"detail"`
}

type apideckMeta struct {
	ItemsOnPage int             `json:"items_on_page"`
	Cursors     *apideckCursors `json:"cursors"`
}

type apideckCursors struct {
	Previous string `json:"previous"`
	Current  string `json:"current"`
	Next     string `json:"next"`
}

type apideckError struct {
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
	Message    string `json:"message"`
	Detail     any    `json:"detail"`
	Ref        string `json:"ref"`
}

// extractDataAndMeta parses the response body to extract data and pagination metadata.
func extractDataAndMeta(rawBody []byte) (any, *spec.ResponseMeta) {
	if len(rawBody) == 0 {
		return nil, nil
	}

	var envelope apideckEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		// Return raw body as string if not valid JSON envelope
		return string(rawBody), nil
	}

	var data any
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			data = string(envelope.Data)
		}
	}

	var meta *spec.ResponseMeta
	if envelope.Meta != nil {
		meta = &spec.ResponseMeta{}
		if envelope.Meta.Cursors != nil {
			meta.Cursor = envelope.Meta.Cursors.Next
			meta.HasMore = envelope.Meta.Cursors.Next != ""
		}
	}

	return data, meta
}

// extractError parses a non-2xx response body into an APIError.
func extractError(rawBody []byte, statusCode int, requestID string) *spec.APIError {
	apiErr := &spec.APIError{
		StatusCode: statusCode,
		RequestID:  requestID,
	}

	if len(rawBody) == 0 {
		apiErr.Message = http.StatusText(statusCode)
		return apiErr
	}

	var envelope apideckEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		apiErr.Message = string(rawBody)
		return apiErr
	}

	if envelope.Error != nil {
		apiErr.Message = envelope.Error.Message
		if envelope.Error.Detail != nil {
			detail, _ := json.Marshal(envelope.Error.Detail)
			apiErr.Detail = string(detail)
		}
		if envelope.Error.StatusCode != 0 {
			apiErr.StatusCode = envelope.Error.StatusCode
		}
	} else if envelope.Message != "" {
		apiErr.Message = envelope.Message
		if envelope.Detail != nil {
			detail, _ := json.Marshal(envelope.Detail)
			apiErr.Detail = string(detail)
		}
	} else {
		apiErr.Message = http.StatusText(statusCode)
	}

	return apiErr
}
