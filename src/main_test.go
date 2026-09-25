package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	aw "github.com/deanishe/awgo"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("alfred_workflow_bundleid", "test-bundle")
	_ = os.Setenv("alfred_workflow_cache", "/tmp/alfred-cache")
	_ = os.Setenv("alfred_workflow_data", "/tmp/alfred-data")
	_ = os.Setenv("alfred_workflow_version", "1.0.0")
	_ = os.Setenv("alfred_debug", "1")

	os.Exit(m.Run())
}

func TestExtractIssueFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		jiraURL  string
		expected string
	}{
		{
			name:     "Valid Jira URL with issue",
			url:      "https://mycompany.atlassian.net/browse/PROJ-123",
			jiraURL:  "https://mycompany.atlassian.net",
			expected: "PROJ-123",
		},
		{
			name:     "Valid Jira URL with issue and query params",
			url:      "https://mycompany.atlassian.net/browse/PROJ-123?filter=all",
			jiraURL:  "https://mycompany.atlassian.net",
			expected: "PROJ-123",
		},
		{
			name:     "Invalid URL (not Jira)",
			url:      "https://google.com/browse/PROJ-123",
			jiraURL:  "https://mycompany.atlassian.net",
			expected: "",
		},
		{
			name:     "Jira URL but not browse",
			url:      "https://mycompany.atlassian.net/dashboard",
			jiraURL:  "https://mycompany.atlassian.net",
			expected: "",
		},
		{
			name:     "Empty jiraURL",
			url:      "https://mycompany.atlassian.net/browse/PROJ-123",
			jiraURL:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractIssueFromURL(tt.url, tt.jiraURL)
			if got != tt.expected {
				t.Errorf("ExtractIssueFromURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResolveCurrentTrack(t *testing.T) {
	cfg = &WorkflowConfig{APIToken: "test-token", WorkspaceID: 12345}
	wf = aw.New()

	tests := []struct {
		name        string
		trackID     string
		description string
		projectID   string
		response    string
		wantCalls   int
		wantRunning bool
		wantID      int
		wantDesc    string
		wantProject int
	}{
		{
			name:        "reuses the entry passed down from the previous step",
			trackID:     "42",
			description: "PROJ-123",
			projectID:   "7",
			response:    `{"id": 99, "description": "stale", "project_id": 1}`,
			wantCalls:   0,
			wantRunning: true,
			wantID:      42,
			wantDesc:    "PROJ-123",
			wantProject: 7,
		},
		{
			name:        "keeps an empty description that was passed down",
			trackID:     "42",
			description: "",
			projectID:   "",
			wantCalls:   0,
			wantRunning: true,
			wantID:      42,
			wantDesc:    "",
			wantProject: 0,
		},
		{
			name:        "falls back to toggl when no entry was passed down",
			response:    `{"id": 99, "description": "Working on it", "project_id": 1}`,
			wantCalls:   1,
			wantRunning: true,
			wantID:      99,
			wantDesc:    "Working on it",
			wantProject: 1,
		},
		{
			name:        "falls back to toggl when the passed down id is unusable",
			trackID:     "not-an-id",
			response:    `{"id": 99, "description": "Working on it", "project_id": 1}`,
			wantCalls:   1,
			wantRunning: true,
			wantID:      99,
			wantDesc:    "Working on it",
			wantProject: 1,
		},
		{
			name:        "reports nothing running",
			response:    `null`,
			wantCalls:   1,
			wantRunning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, tt.response)
			}))
			defer server.Close()

			oldBaseURL, oldTrackID := togglBaseURL, trackIDFlag
			togglBaseURL, trackIDFlag = server.URL, tt.trackID
			t.Setenv("track_description", tt.description)
			t.Setenv("track_project_id", tt.projectID)
			defer func() { togglBaseURL, trackIDFlag = oldBaseURL, oldTrackID }()

			track, running := resolveCurrentTrack()

			if calls != tt.wantCalls {
				t.Errorf("hit /me/time_entries/current %d times, want %d", calls, tt.wantCalls)
			}
			if running != tt.wantRunning {
				t.Fatalf("running = %v, want %v", running, tt.wantRunning)
			}
			if !running {
				return
			}
			if track.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", track.ID, tt.wantID)
			}
			if track.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", track.Description, tt.wantDesc)
			}
			if track.ProjectID != tt.wantProject {
				t.Errorf("ProjectID = %d, want %d", track.ProjectID, tt.wantProject)
			}
		})
	}
}

func TestIssueKeyProvided(t *testing.T) {
	if issueKeyProvided() {
		t.Fatal("issueKeyProvided() = true before parsing, want false")
	}
	if err := cli.Parse([]string{"-issue-key", ""}); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !issueKeyProvided() {
		t.Error("issueKeyProvided() = false for an explicitly empty -issue-key, want true")
	}
}
