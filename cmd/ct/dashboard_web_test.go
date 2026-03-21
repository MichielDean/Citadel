package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MichielDean/cistern/internal/cistern"
)

func TestDashboardWebMux_RootServesHTML(t *testing.T) {
	mux := newDashboardMux(tempCfg(t), tempDB(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want 200", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("body should contain <!DOCTYPE html>")
	}
	if !strings.Contains(body, "EventSource") {
		t.Error("body should contain EventSource (SSE client code)")
	}
}

func TestDashboardWebMux_NotFoundForUnknownPaths(t *testing.T) {
	mux := newDashboardMux(tempCfg(t), tempDB(t))

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent status = %d, want 404", w.Code)
	}
}

func TestDashboardWebMux_APIReturnsJSON(t *testing.T) {
	mux := newDashboardMux(tempCfg(t), tempDB(t))

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/dashboard status = %d, want 200", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var data DashboardData
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if data.FetchedAt.IsZero() {
		t.Error("FetchedAt should be set in JSON response")
	}
}

func TestDashboardWebMux_APIMethodNotAllowed(t *testing.T) {
	mux := newDashboardMux(tempCfg(t), tempDB(t))

	req := httptest.NewRequest(http.MethodPost, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/dashboard status = %d, want 405", w.Code)
	}
}

func TestDashboardWebMux_EventsSSEHeaders(t *testing.T) {
	mux := newDashboardMux(tempCfg(t), tempDB(t))

	// Pre-cancel the context so the SSE handler exits after sending the first event.
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/events", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
	body := w.Body.String()
	if !strings.HasPrefix(body, "data: ") {
		t.Errorf("SSE body should start with 'data: ', got %q", truncateStr(body, 60))
	}
	// The first SSE line's payload must be valid JSON.
	firstLine := strings.SplitN(body, "\n", 2)[0]
	payload := strings.TrimPrefix(firstLine, "data: ")
	var d DashboardData
	if err := json.Unmarshal([]byte(payload), &d); err != nil {
		t.Errorf("SSE payload is not valid DashboardData JSON: %v — payload: %q", err, truncateStr(payload, 80))
	}
}

func TestDashboardWebMux_APIReturnsCorrectCounts(t *testing.T) {
	cfgPath := tempCfg(t)
	dbPath := tempDB(t)

	// Seed: 1 flowing (virgo/implement), 1 queued.
	c, err := cistern.New(dbPath, "mr")
	if err != nil {
		t.Fatal(err)
	}
	flowing, _ := c.Add("myrepo", "Feature A", "", 1, 2)
	c.GetReady("myrepo")
	c.Assign(flowing.ID, "virgo", "implement")
	c.Add("myrepo", "Feature B", "", 2, 2)
	c.Close()

	mux := newDashboardMux(cfgPath, dbPath)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var data DashboardData
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if data.FlowingCount != 1 {
		t.Errorf("FlowingCount = %d, want 1", data.FlowingCount)
	}
	if data.QueuedCount != 1 {
		t.Errorf("QueuedCount = %d, want 1", data.QueuedCount)
	}
}

func TestDashboardWebMux_NoteFieldsRoundTrip(t *testing.T) {
	cfgPath := tempCfg(t)
	dbPath := tempDB(t)

	c, err := cistern.New(dbPath, "mr")
	if err != nil {
		t.Fatal(err)
	}
	droplet, _ := c.Add("myrepo", "Note Test", "", 1, 2)
	c.GetReady("myrepo")
	c.Assign(droplet.ID, "virgo", "implement")
	if err := c.AddNote(droplet.ID, "implementer", "hello world"); err != nil {
		t.Fatal(err)
	}
	c.Close()

	mux := newDashboardMux(cfgPath, dbPath)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var data DashboardData
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(data.FlowActivities) == 0 {
		t.Fatal("expected at least one flow activity")
	}
	notes := data.FlowActivities[0].RecentNotes
	if len(notes) == 0 {
		t.Fatal("expected at least one recent note")
	}
	if notes[0].CataractaeName != "implementer" {
		t.Errorf("CataractaeName = %q, want %q", notes[0].CataractaeName, "implementer")
	}
	if notes[0].Content != "hello world" {
		t.Errorf("Content = %q, want %q", notes[0].Content, "hello world")
	}
}

func TestDashboardWebMux_NoteFieldsSnakeCaseJSON(t *testing.T) {
	cfgPath := tempCfg(t)
	dbPath := tempDB(t)

	c, err := cistern.New(dbPath, "mr")
	if err != nil {
		t.Fatal(err)
	}
	droplet, _ := c.Add("myrepo", "Snake Case Test", "", 1, 2)
	c.GetReady("myrepo")
	c.Assign(droplet.ID, "virgo", "implement")
	if err := c.AddNote(droplet.ID, "implementer", "snake test"); err != nil {
		t.Fatal(err)
	}
	c.Close()

	mux := newDashboardMux(cfgPath, dbPath)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Unmarshal into a generic map to verify raw JSON key names (not Go field names).
	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw JSON: %v", err)
	}
	activities, _ := raw["flow_activities"].([]interface{})
	if len(activities) == 0 {
		t.Fatal("expected at least one flow_activity in raw JSON")
	}
	act, _ := activities[0].(map[string]interface{})
	notes, _ := act["recent_notes"].([]interface{})
	if len(notes) == 0 {
		t.Fatal("expected at least one recent_note in raw JSON")
	}
	note, _ := notes[0].(map[string]interface{})
	if _, ok := note["cataractae_name"]; !ok {
		t.Errorf("raw JSON note missing key %q (got %v)", "cataractae_name", note)
	}
	if _, ok := note["content"]; !ok {
		t.Errorf("raw JSON note missing key %q (got %v)", "content", note)
	}
}

// truncateStr returns at most n runes of s for safe display in test messages.
func truncateStr(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
