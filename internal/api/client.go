package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"racore-cli/internal/auth"
)

const (
	RequestTimeout = 30 * time.Second
	BaseURL        = "https://portal.racorecloud.com"
)

// AuthError represents an authentication failure after retry.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// Client encapsulates API calls with Bearer token auth and 401 retry.
type Client struct {
	authMgr *auth.Manager
	baseURL string
	client  *http.Client
}

// NewClient creates a new API client.
func NewClient(authMgr *auth.Manager) *Client {
	return &Client{
		authMgr: authMgr,
		baseURL: BaseURL,
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// Get sends a GET request with Bearer auth and 401 retry.
func (c *Client) Get(endpoint string, queryParams map[string]string) (json.RawMessage, error) {
	return c.doWithRetry(func(token string) (*http.Response, error) {
		url := c.baseURL + endpoint
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add query parameters
		if len(queryParams) > 0 {
			q := req.URL.Query()
			for k, v := range queryParams {
				q.Set(k, v)
			}
			req.URL.RawQuery = q.Encode()
		}

		req.Header.Set("Authorization", "Bearer "+token)
		return c.client.Do(req)
	})
}

// Post sends a POST request with Bearer auth and 401 retry.
func (c *Client) Post(endpoint string, body interface{}) (json.RawMessage, error) {
	return c.doWithRetry(func(token string) (*http.Response, error) {
		url := c.baseURL + endpoint

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		return c.client.Do(req)
	})
}

// Put sends a PUT request with Bearer auth and 401 retry.
func (c *Client) Put(endpoint string, body interface{}) (json.RawMessage, error) {
	return c.doWithRetry(func(token string) (*http.Response, error) {
		url := c.baseURL + endpoint

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		return c.client.Do(req)
	})
}

// Delete sends a DELETE request with a JSON body, Bearer auth and 401 retry.
func (c *Client) Delete(endpoint string, body interface{}) (json.RawMessage, error) {
	return c.doWithRetry(func(token string) (*http.Response, error) {
		url := c.baseURL + endpoint

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		return c.client.Do(req)
	})
}

// doWithRetry handles the 401 retry flow:
// 1. Get token via authMgr.GetValidToken()
// 2. Send request with Authorization: Bearer <token>
// 3. If 401: ClearCache() → GetValidToken() → retry
// 4. If retry also 401: return AuthError
// 5. Otherwise: read body and return as json.RawMessage
func (c *Client) doWithRetry(sendRequest func(token string) (*http.Response, error)) (json.RawMessage, error) {
	// Step 1: Get token
	token, err := c.authMgr.GetValidToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Step 2: Send request
	resp, err := sendRequest(token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Step 3: If 401, clear cache and retry
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		c.authMgr.ClearCache()
		token, err = c.authMgr.GetValidToken()
		if err != nil {
			return nil, fmt.Errorf("failed to re-authenticate: %w", err)
		}

		// Retry request
		resp, err = sendRequest(token)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Step 4: If retry also 401, return AuthError
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, &AuthError{
				Message: "Authentication failed: please run 'racore-cli login' to re-authenticate",
			}
		}
	}

	// Step 5: Read body and return
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return json.RawMessage(respBody), nil
}
