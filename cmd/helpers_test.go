package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCheckAPIError_Success(t *testing.T) {
	resp := apiResponse{Code: 1, Message: ""}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("expected nil error for code 1, got: %v", err)
	}
}

func TestCheckAPIError_NoDataFound(t *testing.T) {
	resp := apiResponse{Code: -1, Message: "There is no data found."}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "API error (code -1)") {
		t.Errorf("expected error code in message, got: %s", msg)
	}
	if !strings.Contains(msg, "There is no data found.") {
		t.Errorf("expected original message preserved, got: %s", msg)
	}
	if !strings.Contains(msg, "Hint:") {
		t.Errorf("expected hint in message, got: %s", msg)
	}
	if !strings.Contains(msg, "resource may not exist") {
		t.Errorf("expected 'resource may not exist' hint, got: %s", msg)
	}
}

func TestCheckAPIError_NoInformationModified(t *testing.T) {
	resp := apiResponse{Code: 0, Message: "No information modified"}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "already set to the requested value") {
		t.Errorf("expected 'already set' hint, got: %s", msg)
	}
}

func TestCheckAPIError_InvalidParameter(t *testing.T) {
	resp := apiResponse{Code: 0, Message: "Invalid parameter value for field 'state'"}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "parameter value is not accepted") {
		t.Errorf("expected invalid parameter hint, got: %s", msg)
	}
}

func TestCheckAPIError_MissingParameter(t *testing.T) {
	resp := apiResponse{Code: 0, Message: "Missing parameters: domain"}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "required field is missing") {
		t.Errorf("expected missing parameter hint, got: %s", msg)
	}
}

func TestCheckAPIError_Authentication(t *testing.T) {
	resp := apiResponse{Code: 0, Message: "Unauthorized access"}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Authentication issue") {
		t.Errorf("expected auth hint, got: %s", msg)
	}
}

func TestCheckAPIError_EmptyMessageWithErrorBody(t *testing.T) {
	resp := apiResponse{Code: 0, Message: ""}
	rawBody := json.RawMessage(`{"error":"Invalid parameter state"}`)
	err := checkAPIError(resp, rawBody)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Invalid parameter state") {
		t.Errorf("expected error extracted from body, got: %s", msg)
	}
	if !strings.Contains(msg, "parameter value is not accepted") {
		t.Errorf("expected invalid parameter hint, got: %s", msg)
	}
}

func TestCheckAPIError_EmptyMessageEmptyBody(t *testing.T) {
	resp := apiResponse{Code: -1, Message: ""}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "returned an error with no details") {
		t.Errorf("expected 'no details' fallback message, got: %s", msg)
	}
}

func TestCheckAPIError_NoHintForUnknownMessage(t *testing.T) {
	resp := apiResponse{Code: 0, Message: "Something completely unexpected happened"}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if strings.Contains(msg, "Hint:") {
		t.Errorf("expected no hint for unknown error pattern, got: %s", msg)
	}
	if !strings.Contains(msg, "Something completely unexpected happened") {
		t.Errorf("expected original message preserved, got: %s", msg)
	}
}

func TestCheckAPIError_EnableHttpsHint(t *testing.T) {
	resp := apiResponse{Code: 0, Message: "Please enable HTTPS first"}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "requires HTTPS/SSL to be enabled") {
		t.Errorf("expected HTTPS hint, got: %s", msg)
	}
}

func TestCheckAPIError_DomainStateHint(t *testing.T) {
	resp := apiResponse{Code: 0, Message: "Domain can only be opened when in closed state"}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "not in the correct state") {
		t.Errorf("expected domain state hint, got: %s", msg)
	}
}

func TestCheckAPIError_NotSupportedHint(t *testing.T) {
	resp := apiResponse{Code: 0, Message: "Value 'xyz' not supported for this field"}
	err := checkAPIError(resp, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "not in the accepted list") {
		t.Errorf("expected 'not supported' hint, got: %s", msg)
	}
}
