package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MichielDean/cistern/internal/cistern"
	"github.com/MichielDean/cistern/internal/aqueduct"
)

// --- helpers ---

// tempDB creates a temporary SQLite database and returns its path and a cleanup func.
func tempDB(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.db")
}

// tempCfg writes a minimal cistern.yaml referencing a feature.yaml stub.
// Returns the path to the config file.
func tempCfg(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Minimal workflow YAML.
	wfContent := `name: test
cataractae:
  - name: implement
    type: agent
  - name: review
    type: agent
  - name: merge
    type: automated
`
	wfPath := filepath.Join(dir, "feature.yaml")
	if err := os.WriteFile(wfPath, []byte(wfContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Config referencing two operators named "upstream" and "tributary".
	cfgContent := `repos:
  - name: myrepo
    url: https://example.com/repo
    workflow_path: feature.yaml
    cataractae: 2
    names:
      - upstream
      - tributary
    prefix: mr
max_cataractae: 4
`
	cfgPath := filepath.Join(dir, "cistern.yaml")
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	return cfgPath
}

// --- TestFetchDashboardData_FeedsDataCorrectly ---

func TestFetchDashboardData_FeedsDataCorrectly(t *testing.T) {
	cfgPath := tempCfg(t)
	dbPath := tempDB(t)

	// Seed the queue with known items.
	c, err := cistern.New(dbPath, "mr")
	if err != nil {
		t.Fatal(err)
	}

	// Add: 1 flowing assigned to "upstream", 1 queued, 1 closed.
	flowing, _ := c.Add("myrepo", "Feature A", "", 1, 2)
	c.GetReady("myrepo") // marks it in_progress
	c.Assign(flowing.ID, "upstream", "implement")

	_, _ = c.Add("myrepo", "Feature B", "", 2, 2) // stays open/queued

	closed, _ := c.Add("myrepo", "Feature C", "", 1, 2)
	c.CloseItem(closed.ID)
	c.Close()

	data := fetchDashboardData(cfgPath, dbPath)

	if data.CataractaCount != 2 {
		t.Errorf("CataractaCount = %d, want 2", data.CataractaCount)
	}
	if data.FlowingCount != 1 {
		t.Errorf("FlowingCount = %d, want 1", data.FlowingCount)
	}
	if data.QueuedCount != 1 {
		t.Errorf("QueuedCount = %d, want 1", data.QueuedCount)
	}
	if data.DoneCount != 1 {
		t.Errorf("DoneCount = %d, want 1", data.DoneCount)
	}
	if !data.FarmRunning {
		t.Error("FarmRunning should be true when queue is accessible")
	}
	if data.FetchedAt.IsZero() {
		t.Error("FetchedAt should be set")
	}

	// Cataracta "upstream" should be assigned to the flowing item.
	var upstream *CataractaInfo
	for i := range data.Cataractae {
		if data.Cataractae[i].Name == "upstream" {
			upstream = &data.Cataractae[i]
		}
	}
	if upstream == nil {
		t.Fatal("cataracta upstream not found in data.Cataractae")
	}
	if upstream.DropletID != flowing.ID {
		t.Errorf("upstream.DropletID = %q, want %q", upstream.DropletID, flowing.ID)
	}
	if upstream.Step != "implement" {
		t.Errorf("upstream.Step = %q, want %q", upstream.Step, "implement")
	}
	if upstream.CataractaIndex != 1 {
		t.Errorf("upstream.CataractaIndex = %d, want 1", upstream.CataractaIndex)
	}
	if upstream.TotalCataractae != 3 {
		t.Errorf("upstream.TotalCataractae = %d, want 3", upstream.TotalCataractae)
	}

	// Cataracta "tributary" should be dry.
	var tributary *CataractaInfo
	for i := range data.Cataractae {
		if data.Cataractae[i].Name == "tributary" {
			tributary = &data.Cataractae[i]
		}
	}
	if tributary == nil {
		t.Fatal("cataracta tributary not found in data.Cataractae")
	}
	if tributary.DropletID != "" {
		t.Errorf("tributary.DropletID = %q, want empty (dry)", tributary.DropletID)
	}

	// Cistern should contain flowing + queued (2 items).
	if len(data.CisternItems) != 2 {
		t.Errorf("len(CisternItems) = %d, want 2", len(data.CisternItems))
	}

	// Recent should contain the 1 closed item.
	if len(data.RecentItems) != 1 {
		t.Errorf("len(RecentItems) = %d, want 1", len(data.RecentItems))
	}
}

// --- TestFetchDashboardData_FarmNotRunning_ShowsDroughtState ---

func TestFetchDashboardData_FarmNotRunning_ShowsDroughtState(t *testing.T) {
	t.Run("missing config returns empty data", func(t *testing.T) {
		data := fetchDashboardData("/nonexistent/cistern.yaml", "/nonexistent/cistern.db")

		if data == nil {
			t.Fatal("expected non-nil DashboardData even on error")
		}
		if data.FarmRunning {
			t.Error("FarmRunning should be false when config is missing")
		}
		if data.CataractaCount != 0 {
			t.Errorf("CataractaCount = %d, want 0", data.CataractaCount)
		}
		if data.FetchedAt.IsZero() {
			t.Error("FetchedAt should always be set")
		}
	})

	t.Run("valid config but missing DB shows cataractae dry", func(t *testing.T) {
		cfgPath := tempCfg(t)
		dbPath := filepath.Join(t.TempDir(), "nonexistent.db")
		// Don't create the DB — remove it if it exists.
		os.Remove(dbPath)

		// cistern.New creates the DB if missing, so we can't test a truly missing DB
		// at the queue level without making the path unwritable. Instead, test
		// that a fresh empty DB yields all-dry cataractae and zero counts.
		data := fetchDashboardData(cfgPath, dbPath)

		if data.CataractaCount != 2 {
			t.Errorf("CataractaCount = %d, want 2 (from config)", data.CataractaCount)
		}
		if data.FlowingCount != 0 {
			t.Errorf("FlowingCount = %d, want 0 for empty DB", data.FlowingCount)
		}
		if data.QueuedCount != 0 {
			t.Errorf("QueuedCount = %d, want 0 for empty DB", data.QueuedCount)
		}
		for _, ch := range data.Cataractae {
			if ch.DropletID != "" {
				t.Errorf("cataracta %q should be dry (empty DropletID), got %q", ch.Name, ch.DropletID)
			}
		}
	})
}

// --- TestDashboard_ExitsCleanlyOnQ ---

func TestDashboard_ExitsCleanlyOnQ(t *testing.T) {
	cfgPath := tempCfg(t)
	dbPath := tempDB(t)

	inputCh := make(chan byte, 1)
	var out bytes.Buffer

	// Send 'q' after a short delay so the dashboard renders at least once first.
	go func() {
		time.Sleep(50 * time.Millisecond)
		inputCh <- 'q'
	}()

	err := RunDashboard(cfgPath, dbPath, inputCh, &out)
	if err != nil {
		t.Errorf("RunDashboard returned error on q: %v", err)
	}

	// Dashboard should have rendered at least one frame.
	if out.Len() == 0 {
		t.Error("expected some output before exit")
	}
}

func TestDashboard_ExitsCleanlyOnCtrlC(t *testing.T) {
	cfgPath := tempCfg(t)
	dbPath := tempDB(t)

	inputCh := make(chan byte, 1)
	var out bytes.Buffer

	go func() {
		time.Sleep(50 * time.Millisecond)
		inputCh <- 3 // Ctrl-C byte
	}()

	err := RunDashboard(cfgPath, dbPath, inputCh, &out)
	if err != nil {
		t.Errorf("RunDashboard returned error on Ctrl-C: %v", err)
	}
}

func TestDashboard_ExitsWhenInputClosed(t *testing.T) {
	cfgPath := tempCfg(t)
	dbPath := tempDB(t)

	inputCh := make(chan byte)
	var out bytes.Buffer

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(inputCh)
	}()

	err := RunDashboard(cfgPath, dbPath, inputCh, &out)
	if err != nil {
		t.Errorf("RunDashboard returned error when channel closed: %v", err)
	}
}

// --- TestRenderDashboard ---

func TestRenderDashboard_ContainsExpectedSections(t *testing.T) {
	data := &DashboardData{
		CataractaCount:  2,
		FlowingCount: 1,
		QueuedCount:  1,
		DoneCount:    3,
		Cataractae: []CataractaInfo{
			{Name: "upstream", DropletID: "ci-abc12", Step: "implement", CataractaIndex: 1, TotalCataractae: 6, Elapsed: 2*time.Minute + 14*time.Second},
			{Name: "tributary"},
		},
		CisternItems: []*cistern.Droplet{
			{ID: "ci-abc12", Repo: "cistern", Status: "in_progress", CurrentCataracta: "implement", Complexity: 2},
		},
		RecentItems: []*cistern.Droplet{
			{ID: "ci-xyz99", Status: "closed", CurrentCataracta: "merge", UpdatedAt: time.Now()},
		},
		FarmRunning: true,
		FetchedAt:   time.Date(2026, 3, 14, 15, 7, 42, 0, time.UTC),
	}

	out := renderDashboard(data)

	sections := []string{"CISTERN", "SLUICES", "CISTERN", "RECENT FLOW"}
	for _, s := range sections {
		if !strings.Contains(out, s) {
			t.Errorf("output missing section %q", s)
		}
	}
	if !strings.Contains(out, "upstream") {
		t.Error("output missing cataracta name upstream")
	}
	if !strings.Contains(out, "tributary") {
		t.Error("output missing cataracta name tributary")
	}
	if !strings.Contains(out, "15:07:42") {
		t.Error("output missing last update timestamp")
	}
	if !strings.Contains(out, "q to quit") {
		t.Error("output missing footer hint")
	}
}

func TestRenderDashboard_AqueductsClosedWhenNoCataractae(t *testing.T) {
	data := &DashboardData{
		Cataractae:   []CataractaInfo{},
		FetchedAt: time.Now(),
	}
	out := renderDashboard(data)
	if !strings.Contains(out, "Aqueducts closed") {
		t.Error("expected 'Aqueducts closed' when no channels configured")
	}
}

func TestRenderDashboardHTML_ContainsEasterEggHoverText(t *testing.T) {
	dropletID := "ct-ab123"
	stage := "implement"
	elapsed := 42
	snapshot := inspectOutput{
		Cataractae: []cataractaInfo{
			{Name: "upstream", DropletID: &dropletID, Stage: &stage, ElapsedSeconds: &elapsed},
		},
		Queue: cisternInfo{Flowing: 1, Queued: 0, Closed: 0},
	}

	out := renderDashboardHTML(snapshot)

	if !strings.Contains(out, "id=\"easter-egg\"") {
		t.Error("html dashboard should include subtle easter egg icon")
	}
	if !strings.Contains(out, "Four letters guard the gate you seek") {
		t.Error("html dashboard should include easter egg hover text")
	}
	if !strings.Contains(out, "CT Dashboard") {
		t.Error("html dashboard should include title")
	}
}

func TestDashboardListenAddr_UsesProvidedPort(t *testing.T) {
	got := dashboardListenAddr(defaultDashboardHTMLPort)
	if got != ":5737" {
		t.Errorf("dashboardListenAddr(defaultDashboardHTMLPort) = %q, want %q", got, ":5737")
	}
}

func TestRenderDashboardHTML_ShowsEscalatedDropFromInspectSnapshot(t *testing.T) {
	snapshot := inspectOutput{
		Droplets: []dropletInfo{{ID: "ct-poisn1", Status: "escalated", Stage: "qa", UpdatedAt: time.Now()}},
	}

	out := renderDashboardHTML(snapshot)
	if !strings.Contains(out, "ct-poisn1") {
		t.Error("html dashboard should render droplet id from inspect snapshot")
	}
	if !strings.Contains(out, displayStatus("escalated")) {
		t.Error("html dashboard should render escalated display status from inspect snapshot")
	}
}

// --- TestProgressBar ---

func TestProgressBar_FilledAndEmpty(t *testing.T) {
	tests := []struct {
		step, total, width int
		wantFilled         int
	}{
		{1, 6, 6, 1},
		{3, 6, 6, 3},
		{6, 6, 6, 6},
		{0, 6, 6, 0},
		{1, 0, 6, 0}, // zero total → all empty
	}
	for _, tt := range tests {
		bar := progressBar(tt.step, tt.total, tt.width)
		if len([]rune(bar)) != tt.width {
			t.Errorf("progressBar(%d,%d,%d) length = %d, want %d",
				tt.step, tt.total, tt.width, len([]rune(bar)), tt.width)
		}
		filled := strings.Count(bar, "█")
		if filled != tt.wantFilled {
			t.Errorf("progressBar(%d,%d,%d) filled = %d, want %d",
				tt.step, tt.total, tt.width, filled, tt.wantFilled)
		}
	}
}

// --- TestStepIndexInWorkflow ---

func TestStepIndexInWorkflow_ReturnsCorrectIndex(t *testing.T) {
	steps := []aqueduct.WorkflowCataracta{
		{Name: "implement"},
		{Name: "review"},
		{Name: "merge"},
	}

	if idx := cataractaIndexInWorkflow("implement", steps); idx != 1 {
		t.Errorf("stepIndex(implement) = %d, want 1", idx)
	}
	if idx := cataractaIndexInWorkflow("merge", steps); idx != 3 {
		t.Errorf("stepIndex(merge) = %d, want 3", idx)
	}
	if idx := cataractaIndexInWorkflow("unknown", steps); idx != 0 {
		t.Errorf("stepIndex(unknown) = %d, want 0", idx)
	}
}

// --- TestFormatElapsed ---

func TestFormatElapsed_Seconds(t *testing.T) {
	got := formatElapsed(45 * time.Second)
	if got != "45s" {
		t.Errorf("formatElapsed(45s) = %q, want %q", got, "45s")
	}
}

func TestFormatElapsed_MinutesAndSeconds(t *testing.T) {
	got := formatElapsed(2*time.Minute + 14*time.Second)
	if got != "2m 14s" {
		t.Errorf("formatElapsed(2m14s) = %q, want %q", got, "2m 14s")
	}
}
