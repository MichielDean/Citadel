package tracker

import "testing"

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
