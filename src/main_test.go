package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("alfred_workflow_bundleid", "test-bundle")
	os.Setenv("alfred_workflow_cache", "/tmp/alfred-cache")
	os.Setenv("alfred_workflow_data", "/tmp/alfred-data")
	os.Setenv("alfred_workflow_version", "1.0.0")
	os.Setenv("alfred_debug", "1")

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
