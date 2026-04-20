package tracker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractJiraDescription_WithNilDescription_ReturnsEmpty(t *testing.T) {
	got := extractJiraDescription(nil)
	if got != "" {
		t.Errorf("extractJiraDescription(nil) = %q, want %q", got, "")
	}
}

func TestExtractJiraDescription_WithADFDoc_ExtractsText(t *testing.T) {
	// Given: a minimal ADF document with a paragraph containing text.
	adf := map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{
						"type": "text",
						"text": "Hello ADF world",
					},
				},
			},
		},
	}

	// When: extractJiraDescription is called with the ADF object.
	got := extractJiraDescription(adf)

	// Then: the plain text is returned.
	want := "Hello ADF world"
	if got != want {
		t.Errorf("extractJiraDescription(ADF) = %q, want %q", got, want)
	}
}

func TestFetchIssue_PopulatesKeyLabelsSourceURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"key": "PROJ-123",
			"fields": map[string]any{
				"summary": "Test issue",
				"description": map[string]any{
					"type":    "doc",
					"version": 1,
					"content": []any{
						map[string]any{
							"type": "paragraph",
							"content": []any{
								map[string]any{
									"type": "text",
									"text": "Body text",
								},
							},
						},
					},
				},
				"priority": map[string]any{"name": "High"},
				"labels":   []string{"bug", "frontend"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := newJiraProvider(TrackerConfig{
		BaseURL:  srv.URL,
		TokenEnv: "TEST_JIRA_TOKEN",
	})
	if err != nil {
		t.Fatalf("newJiraProvider: %v", err)
	}

	t.Setenv("TEST_JIRA_TOKEN", "fake-token")

	issue, err := p.FetchIssue("PROJ-123")
	if err != nil {
		t.Fatalf("FetchIssue: %v", err)
	}

	if issue.Key != "PROJ-123" {
		t.Errorf("Key = %q, want %q", issue.Key, "PROJ-123")
	}
	if len(issue.Labels) != 2 || issue.Labels[0] != "bug" || issue.Labels[1] != "frontend" {
		t.Errorf("Labels = %v, want [bug frontend]", issue.Labels)
	}
	wantURL := srv.URL + "/browse/PROJ-123"
	if issue.SourceURL != wantURL {
		t.Errorf("SourceURL = %q, want %q", issue.SourceURL, wantURL)
	}
}
