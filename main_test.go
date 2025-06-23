package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestJSONRPCRequest_Marshal(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "textDocument/hover",
		Params: map[string]interface{}{
			"textDocument": map[string]interface{}{
				"uri": "file:///test.go",
			},
			"position": map[string]interface{}{
				"line":      10,
				"character": 5,
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaled JSONRPCRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if unmarshaled.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", unmarshaled.JSONRPC)
	}
	if unmarshaled.ID != 1 {
		t.Errorf("Expected ID 1, got %d", unmarshaled.ID)
	}
	if unmarshaled.Method != "textDocument/hover" {
		t.Errorf("Expected method textDocument/hover, got %s", unmarshaled.Method)
	}
}

func TestJSONRPCResponse_Unmarshal(t *testing.T) {
	responseJSON := `{
		"jsonrpc": "2.0",
		"id": 1,
		"result": {
			"contents": {
				"kind": "markdown",
				"value": "Test hover content"
			}
		}
	}`

	var response JSONRPCResponse
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", response.JSONRPC)
	}
	if response.ID != 1 {
		t.Errorf("Expected ID 1, got %d", response.ID)
	}
	if response.Result == nil {
		t.Error("Expected result to be non-nil")
	}
}

func TestJSONRPCResponse_Error(t *testing.T) {
	errorJSON := `{
		"jsonrpc": "2.0",
		"id": 1,
		"error": {
			"code": -32601,
			"message": "Method not found"
		}
	}`

	var response JSONRPCResponse
	if err := json.Unmarshal([]byte(errorJSON), &response); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if response.Error == nil {
		t.Error("Expected error to be non-nil")
	}
	if response.Result != nil {
		t.Error("Expected result to be nil for error response")
	}
}

func TestLSPClient_IDIncrement(t *testing.T) {
	client := &LSPClient{id: 1}

	// Simulate creating multiple requests
	req1 := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      client.id,
		Method:  "initialize",
	}
	client.id++

	req2 := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      client.id,
		Method:  "textDocument/hover",
	}
	client.id++

	if req1.ID != 1 {
		t.Errorf("Expected first request ID to be 1, got %d", req1.ID)
	}
	if req2.ID != 2 {
		t.Errorf("Expected second request ID to be 2, got %d", req2.ID)
	}
	if client.id != 3 {
		t.Errorf("Expected client ID to be 3 after two requests, got %d", client.id)
	}
}

func TestContentLengthParsing(t *testing.T) {
	// Test the logic used in ReadResponse for parsing Content-Length header
	testCases := []struct {
		line     string
		expected int
		hasError bool
	}{
		{"Content-Length: 123", 123, false},
		{"Content-Length:456", 456, false},
		{"Content-Length: 0", 0, false},
		{"Content-Type: application/json", 0, true}, // Not a Content-Length header
		{"Content-Length: abc", 0, true},            // Invalid number
	}

	for _, tc := range testCases {
		t.Run(tc.line, func(t *testing.T) {
			var contentLength int
			var err error

			if strings.HasPrefix(tc.line, "Content-Length:") {
				lengthStr := strings.TrimSpace(strings.TrimPrefix(tc.line, "Content-Length:"))
				contentLength, err = strconv.Atoi(lengthStr)
			} else {
				// Simulate not finding Content-Length
				err = fmt.Errorf("not a Content-Length header")
			}

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for line %q, but got none", tc.line)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for line %q: %v", tc.line, err)
				}
				if contentLength != tc.expected {
					t.Errorf("Expected content length %d, got %d", tc.expected, contentLength)
				}
			}
		})
	}
}

func TestInitializeParams(t *testing.T) {
	// Test the initialize parameters structure
	params := map[string]interface{}{
		"processId": 12345,
		"rootUri":   "file:///test/project",
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"completion": map[string]interface{}{
					"completionItem": map[string]interface{}{
						"snippetSupport": true,
					},
				},
			},
		},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal initialize params: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal initialize params: %v", err)
	}

	if unmarshaled["processId"].(float64) != 12345 {
		t.Errorf("Expected processId 12345, got %v", unmarshaled["processId"])
	}
	if unmarshaled["rootUri"].(string) != "file:///test/project" {
		t.Errorf("Expected rootUri file:///test/project, got %v", unmarshaled["rootUri"])
	}

	capabilities, ok := unmarshaled["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected capabilities to be a map")
	}

	textDocument, ok := capabilities["textDocument"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected textDocument to be a map")
	}

	completion, ok := textDocument["completion"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected completion to be a map")
	}

	completionItem, ok := completion["completionItem"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected completionItem to be a map")
	}

	if completionItem["snippetSupport"].(bool) != true {
		t.Error("Expected snippetSupport to be true")
	}
}
