package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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
// Priority: environment variables > credential file
func newAuthenticatedClient() (*api.Client, error) {
	// Priority 1: Environment variables (for MCP server mode)
	envAccessKey := os.Getenv("RACORE_ACCESS_KEY")
	envSecretKey := os.Getenv("RACORE_SECRET_KEY")
	if envAccessKey != "" && envSecretKey != "" {
		mgr := auth.NewManager(envAccessKey, envSecretKey)
		return api.NewClient(mgr), nil
	}

	// Priority 2: Credential file (~/.racore/credentials)
	store, err := credential.NewStore()
	if err != nil {
		return nil, fmt.Errorf("cannot initialize credential store: %w", err)
	}

	creds, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load credentials: %w", err)
	}

	if creds == nil {
		return nil, fmt.Errorf("not logged in. Please run 'racore-cli login' first, or set RACORE_ACCESS_KEY and RACORE_SECRET_KEY environment variables")
	}

	mgr := auth.NewManager(creds.AccessKey, creds.SecretKey)
	mgr.SetCache(creds.Token, creds.Expire)

	return api.NewClient(mgr), nil
}

// checkAPIError returns a user-friendly error when an API response indicates failure.
// It handles cases where the API returns an empty message by including the response body,
// and appends troubleshooting hints based on common error patterns.
func checkAPIError(resp apiResponse, rawBody json.RawMessage) error {
	if resp.Code == 1 {
		return nil
	}

	msg := resp.Message

	// If message is empty, try to extract from raw body {"error":"..."} pattern
	if msg == "" {
		var errBody struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(rawBody, &errBody) == nil && errBody.Error != "" {
			msg = errBody.Error
		}
	}

	// If still empty, use truncated raw body as the message
	if msg == "" {
		body := strings.TrimSpace(string(rawBody))
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		if body != "" && body != "{}" && body != "null" {
			msg = body
		} else {
			msg = "the server returned an error with no details"
		}
	}

	// Add troubleshooting hints based on common error patterns
	hint := ""
	lowerMsg := strings.ToLower(msg)
	switch {
	case strings.Contains(lowerMsg, "no data found") || strings.Contains(lowerMsg, "there is no data"):
		hint = "\n  Hint: The requested resource may not exist, or the query parameters don't match any records.\n  Verify the domain name, time range, or other parameters are correct."
	case strings.Contains(lowerMsg, "no information modified"):
		hint = "\n  Hint: The configuration is already set to the requested value. No change was needed."
	case strings.Contains(lowerMsg, "invalid parameter"):
		hint = "\n  Hint: A parameter value is not accepted. Check the --config JSON field values.\n  Use '<command> --help' for valid options and examples."
	case strings.Contains(lowerMsg, "missing parameters") || strings.Contains(lowerMsg, "missing parameter"):
		hint = "\n  Hint: A required field is missing from the --config JSON.\n  Use '<command> --help' for the correct format and required fields."
	case strings.Contains(lowerMsg, "please enable https") || strings.Contains(lowerMsg, "please enable ssl"):
		hint = "\n  Hint: This feature requires HTTPS/SSL to be enabled first.\n  Run: racore-cli domain ssl set --domain <domain> --config '{\"is_ssl\":\"1\",\"cert_id\":\"<id>\"}'"
	case strings.Contains(lowerMsg, "only be opened when") || strings.Contains(lowerMsg, "only be closed when"):
		hint = "\n  Hint: The domain is not in the correct state for this operation.\n  Check current status: racore-cli domain list --filter <domain>"
	case strings.Contains(lowerMsg, "not supported"):
		hint = "\n  Hint: The parameter value is not in the accepted list.\n  Use '<command> --help' for valid options."
	case strings.Contains(lowerMsg, "authentication") || strings.Contains(lowerMsg, "unauthorized") || strings.Contains(lowerMsg, "token"):
		hint = "\n  Hint: Authentication issue. Try: racore-cli login"
	}

	return fmt.Errorf("API error (code %d): %s%s", resp.Code, msg, hint)
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
