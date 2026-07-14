package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"racore-cli/internal/api"
	"racore-cli/internal/auth"
	"racore-cli/internal/credential"
)

// apiResponse represents the standard API response envelope.
type apiResponse struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

// newAuthenticatedClient loads credentials and returns a ready-to-use API client.
func newAuthenticatedClient() (*api.Client, error) {
	store, err := credential.NewStore()
	if err != nil {
		return nil, fmt.Errorf("cannot initialize credential store: %w", err)
	}

	creds, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load credentials: %w", err)
	}

	if creds == nil {
		return nil, fmt.Errorf("not logged in. Please run 'racore-cli login' first")
	}

	mgr := auth.NewManager(creds.AccessKey, creds.SecretKey)
	mgr.SetCache(creds.Token, creds.Expire)

	return api.NewClient(mgr), nil
}

// checkAPIError returns a user-friendly error when an API response indicates failure.
// It handles cases where the API returns an empty message by including the response body.
func checkAPIError(resp apiResponse, rawBody json.RawMessage) error {
	if resp.Code == 1 {
		return nil
	}
	if resp.Message != "" {
		return fmt.Errorf("API error (code %d): %s", resp.Code, resp.Message)
	}
	// Try to provide useful context from the raw response
	// Truncate if too long
	body := strings.TrimSpace(string(rawBody))
	if len(body) > 200 {
		body = body[:200] + "..."
	}
	if body == "" || body == "{}" || body == "null" {
		return fmt.Errorf("API request failed (code %d): the server returned an error with no details. Check that the resource exists and parameters are correct", resp.Code)
	}
	return fmt.Errorf("API request failed (code %d): %s", resp.Code, body)
}

// formatTable formats headers and rows into an aligned table using tabwriter.
func formatTable(headers []string, rows [][]string) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\n", strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintf(w, "%s\n", strings.Join(row, "\t"))
	}
	w.Flush()
	return buf.String()
}

// formatKeyValue formats key-value pairs for single-resource display.
func formatKeyValue(pairs map[string]string) string {
	var buf strings.Builder
	for k, v := range pairs {
		buf.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	return buf.String()
}
