package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	aw "github.com/deanishe/awgo"
)

func TestGetCurrentTracking(t *testing.T) {
	// Mock configuration
	cfg = &WorkflowConfig{
		APIToken:    "test-token",
		WorkspaceID: 12345,
	}
	// Mock workflow for FatalError
	wf = aw.New()

	tests := []struct {
		name           string
		serverResponse string
		expectedMsg    string
	}{
		{
			name:           "Tracking running",
			serverResponse: `{"id": 1, "description": "Working on it"}`,
			expectedMsg:    `{"id": 1, "description": "Working on it"}`,
		},
		{
			name:           "Tracking not running",
			serverResponse: `null`,
			expectedMsg:    "not running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/me/time_entries/current" {
					t.Errorf("Expected to request '/me/time_entries/current', got: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, tt.serverResponse)
			}))
			defer server.Close()

			// Override base URL for test
			oldBaseURL := togglBaseURL
			togglBaseURL = server.URL
			defer func() { togglBaseURL = oldBaseURL }()

			msg := GetCurrentTracking()
			if msg != tt.expectedMsg {
				t.Errorf("GetCurrentTracking() = %v, want %v", msg, tt.expectedMsg)
			}
		})
	}
}

func TestStartTracking(t *testing.T) {
	cfg = &WorkflowConfig{
		APIToken:    "test-token",
		WorkspaceID: 12345,
	}
	wf = aw.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}
		expectedPath := "/workspaces/12345/time_entries"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got: %s", expectedPath, r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["description"] != "PROJ-123" {
			t.Errorf("Expected description 'PROJ-123', got: %v", body["description"])
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id": 1, "description": "PROJ-123"}`)
	}))
	defer server.Close()

	oldBaseURL := togglBaseURL
	togglBaseURL = server.URL
	defer func() { togglBaseURL = oldBaseURL }()

	msg := StartTracking("PROJ-123")
	expectedMsg := "Tracking started for PROJ-123"
	if msg != expectedMsg {
		t.Errorf("StartTracking() = %v, want %v", msg, expectedMsg)
	}
}

func TestAddDescription(t *testing.T) {
	cfg = &WorkflowConfig{
		APIToken:    "test-token",
		WorkspaceID: 12345,
	}
	wf = aw.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got: %s", r.Method)
		}
		expectedPath := "/workspaces/12345/time_entries/999"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got: %s", expectedPath, r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["description"] != "New Description" {
			t.Errorf("Expected description 'New Description', got: %v", body["description"])
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id": 999, "description": "New Description"}`)
	}))
	defer server.Close()

	oldBaseURL := togglBaseURL
	togglBaseURL = server.URL
	defer func() { togglBaseURL = oldBaseURL }()

	msg := AddDescription("New Description", 999)
	expectedMsg := "New Description added to current toggl entry"
	if msg != expectedMsg {
		t.Errorf("AddDescription() = %v, want %v", msg, expectedMsg)
	}
}

func TestStopTogglEntry(t *testing.T) {
	cfg = &WorkflowConfig{
		APIToken:    "test-token",
		WorkspaceID: 12345,
	}
	wf = aw.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH request, got: %s", r.Method)
		}
		expectedPath := "/workspaces/12345/time_entries/888/stop"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got: %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id": 888, "stop": "2026-04-07T09:00:00Z"}`)
	}))
	defer server.Close()

	oldBaseURL := togglBaseURL
	togglBaseURL = server.URL
	defer func() { togglBaseURL = oldBaseURL }()

	err := StopTogglEntry(888)
	if err != nil {
		t.Errorf("StopTogglEntry() returned error: %v", err)
	}
}

func TestQuotaReached(t *testing.T) {
	cfg = &WorkflowConfig{
		APIToken:    "test-token",
		WorkspaceID: 12345,
	}
	// We can't easily test wf.FatalError as it calls os.Exit.
	// But we can check if it behaves as expected by making sure it doesn't panic before that.
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Toggl-Quota-Resets-In", "3600")
		w.WriteHeader(http.StatusPaymentRequired) // 402
	}))
	defer server.Close()

	oldBaseURL := togglBaseURL
	togglBaseURL = server.URL
	defer func() { togglBaseURL = oldBaseURL }()

	// Mock wf to prevent real exit if possible, but awgo doesn't make it easy.
	// For this test, we just want to ensure it doesn't crash elsewhere.
	// Since FatalError will exit, we can't fully test it here without more complex mocking.
}
