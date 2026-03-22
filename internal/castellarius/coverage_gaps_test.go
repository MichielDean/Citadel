package castellarius

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MichielDean/cistern/internal/aqueduct"
	"github.com/MichielDean/cistern/internal/cistern"
)

// --- AqueductPool gap tests ---

func TestAqueductPool_IsFlowing(t *testing.T) {
	pool := NewAqueductPool("repo", []string{"alpha", "beta"})

	// Neither is flowing initially.
	if pool.IsFlowing("alpha") {
		t.Error("alpha should not be flowing before assignment")
	}
	if pool.IsFlowing("beta") {
		t.Error("beta should not be flowing before assignment")
	}
	if pool.IsFlowing("nonexistent") {
		t.Error("nonexistent aqueduct should not be flowing")
	}

	// Assign alpha — it becomes flowing.
	w := pool.AvailableAqueduct()
	pool.Assign(w, "drop-1", "implement")

	if !pool.IsFlowing("alpha") {
		t.Error("alpha should be flowing after assignment")
	}
	if pool.IsFlowing("beta") {
		t.Error("beta should still be idle")
	}

	// Release alpha — back to idle.
	pool.Release(w)
	if pool.IsFlowing("alpha") {
		t.Error("alpha should be idle after release")
	}
}

func TestAqueductPool_FindAndClaimByName(t *testing.T) {
	pool := NewAqueductPool("repo", []string{"alpha", "beta"})

	// Claim alpha by name — returns it and marks flowing.
	w := pool.FindAndClaimByName("alpha")
	if w == nil {
		t.Fatal("FindAndClaimByName(alpha) returned nil, want non-nil")
	}
	if w.Name != "alpha" {
		t.Errorf("claimed aqueduct name = %q, want %q", w.Name, "alpha")
	}
	if w.Status != AqueductFlowing {
		t.Errorf("claimed aqueduct status = %q, want flowing", w.Status)
	}

	// Trying to claim alpha again while flowing returns nil.
	w2 := pool.FindAndClaimByName("alpha")
	if w2 != nil {
		t.Error("FindAndClaimByName(alpha) on a flowing aqueduct should return nil")
	}

	// Unknown name returns nil.
	if pool.FindAndClaimByName("nonexistent") != nil {
		t.Error("FindAndClaimByName(nonexistent) should return nil")
	}

	// Beta is still available.
	wb := pool.FindAndClaimByName("beta")
	if wb == nil || wb.Name != "beta" {
		t.Errorf("FindAndClaimByName(beta) = %v, want beta", wb)
	}
}

func TestAqueductPool_AvailableAqueductExcluding(t *testing.T) {
	pool := NewAqueductPool("repo", []string{"alpha", "beta", "gamma"})

	// Exclude alpha and beta — should get gamma.
	w := pool.AvailableAqueductExcluding(map[string]bool{"alpha": true, "beta": true})
	if w == nil || w.Name != "gamma" {
		t.Errorf("AvailableAqueductExcluding = %v, want gamma", w)
	}

	// Exclude all three — returns nil.
	w2 := pool.AvailableAqueductExcluding(map[string]bool{"alpha": true, "beta": true, "gamma": true})
	if w2 != nil {
		t.Errorf("AvailableAqueductExcluding all = %v, want nil", w2)
	}

	// Assign alpha; AvailableAqueductExcluding with empty exclude skips it as flowing.
	pool.Assign(pool.AvailableAqueduct(), "drop-1", "implement")
	// alpha is now flowing — available excluding {} returns beta (first idle).
	w3 := pool.AvailableAqueductExcluding(map[string]bool{})
	if w3 == nil {
		t.Error("AvailableAqueductExcluding with empty exclude should return an idle aqueduct")
	}
	// Must be an idle one (not alpha).
	if w3 != nil && w3.Name == "alpha" {
		t.Error("AvailableAqueductExcluding should not return a flowing aqueduct")
	}
}

// --- isSupervisedProcess tests ---

func TestIsSupervisedProcess_CT_SUPERVISED(t *testing.T) {
	t.Setenv("CT_SUPERVISED", "1")
	if !isSupervisedProcess() {
		t.Error("CT_SUPERVISED=1 should be detected as supervised")
	}
}

func TestIsSupervisedProcess_INVOCATION_ID(t *testing.T) {
	t.Setenv("INVOCATION_ID", "some-systemd-id")
	if !isSupervisedProcess() {
		t.Error("INVOCATION_ID set should be detected as supervised (systemd)")
	}
}

func TestIsSupervisedProcess_SUPERVISOR_ENABLED(t *testing.T) {
	t.Setenv("SUPERVISOR_ENABLED", "1")
	if !isSupervisedProcess() {
		t.Error("SUPERVISOR_ENABLED=1 should be detected as supervised")
	}
}

func TestIsSupervisedProcess_NotSupervised(t *testing.T) {
	// Clear all supervisor environment variables.
	t.Setenv("CT_SUPERVISED", "")
	t.Setenv("INVOCATION_ID", "")
	t.Setenv("SUPERVISOR_ENABLED", "")
	// Can only test the env-var paths here (ppid==1 would be true in Docker, but
	// in a normal test environment ppid != 1, so the function returns false).
	// We just verify it doesn't panic.
	_ = isSupervisedProcess()
}

// --- WithLogger / WithPollInterval option tests ---

func TestWithLogger_Option(t *testing.T) {
	client := newMockClient()
	runner := newMockRunner(client)
	customLogger := slog.Default()

	sched := testScheduler(client, runner)
	WithLogger(customLogger)(sched)

	if sched.logger != customLogger {
		t.Error("WithLogger did not set the logger")
	}
}

func TestWithPollInterval_Option(t *testing.T) {
	client := newMockClient()
	runner := newMockRunner(client)
	interval := 42 * time.Second

	sched := testScheduler(client, runner)
	WithPollInterval(interval)(sched)

	if sched.pollInterval != interval {
		t.Errorf("WithPollInterval = %v, want %v", sched.pollInterval, interval)
	}
}

// --- purgeOldItems tests ---

// purgeTrackingClient wraps mockClient and tracks Purge calls.
type purgeTrackingClient struct {
	*mockClient
	purgeCalls int
	purgeN     int
}

func (p *purgeTrackingClient) Purge(olderThan time.Duration, dryRun bool) (int, error) {
	p.purgeCalls++
	return p.purgeN, nil
}

func TestPurgeOldItems_CallsPurgeOnAllRepos(t *testing.T) {
	mc := &purgeTrackingClient{mockClient: newMockClient(), purgeN: 2}
	config := testConfig()
	workflows := map[string]*aqueduct.Workflow{"test-repo": testWorkflow()}
	clients := map[string]CisternClient{"test-repo": mc}
	sched := NewFromParts(config, workflows, clients, newMockRunner(mc.mockClient))

	sched.purgeOldItems()

	if mc.purgeCalls != 1 {
		t.Errorf("Purge called %d times, want 1", mc.purgeCalls)
	}
}

func TestPurgeOldItems_DefaultRetentionDays(t *testing.T) {
	mc := &purgeTrackingClient{mockClient: newMockClient()}
	config := testConfig() // RetentionDays = 0 → default 90
	workflows := map[string]*aqueduct.Workflow{"test-repo": testWorkflow()}
	clients := map[string]CisternClient{"test-repo": mc}
	sched := NewFromParts(config, workflows, clients, newMockRunner(mc.mockClient))

	// Must not panic or error.
	sched.purgeOldItems()
}

// --- recoverInProgress tests ---

func TestRecoverInProgress_ItemWithOutcome_NotReset(t *testing.T) {
	client := newMockClient()
	// Item already has an outcome — observe phase should handle it.
	item := &cistern.Droplet{
		ID:               "r1",
		CurrentCataractae: "implement",
		Status:           "in_progress",
		Assignee:         "alpha",
		Outcome:          "pass",
	}
	client.items["r1"] = item

	sched := testScheduler(client, newMockRunner(client))
	sched.recoverInProgress()

	client.mu.Lock()
	defer client.mu.Unlock()
	// Should not have been reset — still has the original step.
	if client.steps["r1"] != "" {
		t.Errorf("item with outcome should not be reset, got step %q", client.steps["r1"])
	}
}

func TestRecoverInProgress_ItemWithoutOutcome_IsReset(t *testing.T) {
	client := newMockClient()
	item := &cistern.Droplet{
		ID:               "r2",
		CurrentCataractae: "review",
		Status:           "in_progress",
		Assignee:         "alpha",
		Outcome:          "", // no outcome
	}
	client.items["r2"] = item

	sched := testScheduler(client, newMockRunner(client))
	sched.recoverInProgress()

	client.mu.Lock()
	defer client.mu.Unlock()
	// Should have been reset to open at the same step.
	if client.steps["r2"] != "review" {
		t.Errorf("item without outcome should be reset to its cataractae, got %q", client.steps["r2"])
	}
}

func TestRecoverInProgress_EmptyStep_UsesWorkflowDefault(t *testing.T) {
	client := newMockClient()
	// Item has no current cataractae — should fall back to first step in workflow.
	item := &cistern.Droplet{
		ID:               "r3",
		CurrentCataractae: "",
		Status:           "in_progress",
		Assignee:         "alpha",
		Outcome:          "",
	}
	client.items["r3"] = item

	sched := testScheduler(client, newMockRunner(client))
	sched.recoverInProgress()

	client.mu.Lock()
	defer client.mu.Unlock()
	// Should have been reset to the first step in the workflow ("implement").
	if client.steps["r3"] != "implement" {
		t.Errorf("empty step item should be reset to first workflow step 'implement', got %q", client.steps["r3"])
	}
}

// --- heartbeatRepo tests ---

func TestHeartbeatRepo_ResetsStalled_NoAssignee(t *testing.T) {
	client := newMockClient()
	// Item with no assignee — tmux check is skipped, goes straight to reset.
	item := &cistern.Droplet{
		ID:               "hb-1",
		CurrentCataractae: "implement",
		Status:           "in_progress",
		Assignee:         "", // no assignee
		Outcome:          "",
	}
	client.items["hb-1"] = item

	sched := testScheduler(client, newMockRunner(client))
	sched.heartbeatRepo(context.Background(), sched.config.Repos[0])

	client.mu.Lock()
	defer client.mu.Unlock()
	// Item should have been reset to open at its current step.
	if client.steps["hb-1"] != "implement" {
		t.Errorf("stalled item should be reset to 'implement', got %q", client.steps["hb-1"])
	}
}

func TestHeartbeatRepo_SkipsItemsWithOutcome(t *testing.T) {
	client := newMockClient()
	item := &cistern.Droplet{
		ID:               "hb-2",
		CurrentCataractae: "review",
		Status:           "in_progress",
		Assignee:         "",
		Outcome:          "pass", // has outcome — observe phase handles it
	}
	client.items["hb-2"] = item

	sched := testScheduler(client, newMockRunner(client))
	sched.heartbeatRepo(context.Background(), sched.config.Repos[0])

	client.mu.Lock()
	defer client.mu.Unlock()
	// Item with outcome should NOT be reset.
	if client.steps["hb-2"] != "" {
		t.Errorf("item with outcome should not be reset by heartbeat, got step %q", client.steps["hb-2"])
	}
}

func TestHeartbeatRepo_DeadTmuxSession_ResetsItem(t *testing.T) {
	client := newMockClient()
	// Item with an assignee whose tmux session does not exist (no tmux in test env).
	item := &cistern.Droplet{
		ID:               "hb-3",
		CurrentCataractae: "implement",
		Status:           "in_progress",
		Assignee:         "alpha", // tmux session test-repo-alpha won't be alive
		Outcome:          "",
	}
	client.items["hb-3"] = item

	sched := testScheduler(client, newMockRunner(client))
	// Mark the pool worker as flowing so Release() can find it.
	pool := sched.pools["test-repo"]
	w := pool.AvailableAqueduct()
	if w != nil {
		pool.Assign(w, "hb-3", "implement")
	}

	sched.heartbeatRepo(context.Background(), sched.config.Repos[0])

	client.mu.Lock()
	defer client.mu.Unlock()
	// Dead session → item reset to open.
	if client.steps["hb-3"] != "implement" {
		t.Errorf("item with dead tmux session should be reset, got %q", client.steps["hb-3"])
	}
}

func TestHeartbeatInProgress_CallsHeartbeatForAllRepos(t *testing.T) {
	// Basic smoke test: heartbeatInProgress should iterate all repos without panic.
	client := newMockClient()
	item := &cistern.Droplet{
		ID:               "hb-all-1",
		CurrentCataractae: "implement",
		Status:           "in_progress",
		Assignee:         "",
		Outcome:          "",
	}
	client.items["hb-all-1"] = item

	sched := testScheduler(client, newMockRunner(client))
	sched.heartbeatInProgress(context.Background())

	client.mu.Lock()
	defer client.mu.Unlock()
	if client.steps["hb-all-1"] != "implement" {
		t.Errorf("heartbeatInProgress: stalled item should be reset, got %q", client.steps["hb-all-1"])
	}
}

// --- worktreeRegistered test ---

func TestWorktreeRegistered_NonGitDir_ReturnsFalse(t *testing.T) {
	dir := t.TempDir()
	// A non-git directory: git worktree list will fail → returns false.
	if worktreeRegistered(dir, "/some/worktree/path") {
		t.Error("worktreeRegistered should return false for a non-git directory")
	}
}

// --- removeDropletWorktree test ---

func TestRemoveDropletWorktree_NonGitDir_NoOp(t *testing.T) {
	// Calling removeDropletWorktree on a non-git directory ignores the error.
	// The key behavior is that it does not panic or crash.
	primaryDir := t.TempDir()
	sandboxRoot := t.TempDir()
	removeDropletWorktree(primaryDir, sandboxRoot, "myrepo", "drop-noop")
}

// --- hookTmpCleanup test ---

func TestHookTmpCleanup_NoMatchingDirs_Succeeds(t *testing.T) {
	// In a clean test environment there are no ct-diff-* dirs so this is a no-op.
	// Verifies the function runs without error regardless.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	if err := hookTmpCleanup(logger); err != nil {
		t.Errorf("hookTmpCleanup: unexpected error: %v", err)
	}
}

// --- hookDBVacuum tests ---

func TestHookDBVacuum_EmptyPath_ReturnsError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	err := hookDBVacuum("", logger)
	if err == nil {
		t.Error("hookDBVacuum with empty path should return an error")
	}
}

func TestHookDBVacuum_ValidDB_Succeeds(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	// Create a DB with the full cistern schema (which includes all needed tables).
	_, err := cistern.New(dbPath, "test")
	if err != nil {
		t.Fatalf("cistern.New: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	if err := hookDBVacuum(dbPath, logger); err != nil {
		t.Errorf("hookDBVacuum on valid DB: %v", err)
	}
}

// --- hookEventsPrune tests ---

func TestHookEventsPrune_EmptyPath_ReturnsError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	hook := aqueduct.DroughtHook{Name: "test", Action: "events_prune", KeepDays: 30}
	err := hookEventsPrune("", hook, logger)
	if err == nil {
		t.Error("hookEventsPrune with empty path should return an error")
	}
}

func TestHookEventsPrune_ValidDB_Succeeds(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, err := cistern.New(dbPath, "test")
	if err != nil {
		t.Fatalf("cistern.New: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	hook := aqueduct.DroughtHook{Name: "test", Action: "events_prune", KeepDays: 30}
	if err := hookEventsPrune(dbPath, hook, logger); err != nil {
		t.Errorf("hookEventsPrune on valid DB: %v", err)
	}
}

func TestHookEventsPrune_DefaultKeepDays(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, err := cistern.New(dbPath, "test")
	if err != nil {
		t.Fatalf("cistern.New: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	// KeepDays = 0 → default 30.
	hook := aqueduct.DroughtHook{Name: "test", Action: "events_prune", KeepDays: 0}
	if err := hookEventsPrune(dbPath, hook, logger); err != nil {
		t.Errorf("hookEventsPrune with KeepDays=0: %v", err)
	}
}

// --- RunDroughtHooks via db_vacuum/events_prune actions ---

func TestRunDroughtHooks_DbVacuumAction(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, err := cistern.New(dbPath, "test")
	if err != nil {
		t.Fatalf("cistern.New: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	cfg := &aqueduct.AqueductConfig{}
	hooks := []aqueduct.DroughtHook{{Name: "vacuum", Action: "db_vacuum"}}
	// Must not panic.
	RunDroughtHooks(hooks, cfg, dbPath, t.TempDir(), logger, time.Time{}, false, nil)
}

func TestRunDroughtHooks_EventsPruneAction(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_, err := cistern.New(dbPath, "test")
	if err != nil {
		t.Fatalf("cistern.New: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	cfg := &aqueduct.AqueductConfig{}
	hooks := []aqueduct.DroughtHook{{Name: "prune", Action: "events_prune", KeepDays: 30}}
	RunDroughtHooks(hooks, cfg, dbPath, t.TempDir(), logger, time.Time{}, false, nil)
}

func TestRunDroughtHooks_TmpCleanupAction(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	cfg := &aqueduct.AqueductConfig{}
	hooks := []aqueduct.DroughtHook{{Name: "tmp", Action: "tmp_cleanup"}}
	RunDroughtHooks(hooks, cfg, "", t.TempDir(), logger, time.Time{}, false, nil)
}

func TestRunDroughtHooks_UnknownAction_Ignored(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	cfg := &aqueduct.AqueductConfig{}
	hooks := []aqueduct.DroughtHook{{Name: "noop", Action: "completely_unknown_action"}}
	// Unknown actions are logged and skipped — must not panic.
	RunDroughtHooks(hooks, cfg, "", t.TempDir(), logger, time.Time{}, false, nil)
}

func TestRunDroughtHooks_RestartSelf_UnsupervisedNoReload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	cfg := &aqueduct.AqueductConfig{}
	hooks := []aqueduct.DroughtHook{{Name: "restart", Action: "restart_self"}}
	// restart_self + unsupervised + no workflowChanged → warns but does not exit.
	RunDroughtHooks(hooks, cfg, "", t.TempDir(), logger, time.Time{}, false, nil)
}

func TestRunDroughtHooks_RestartSelf_UnsupervisedWithReload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	cfg := &aqueduct.AqueductConfig{}
	hooks := []aqueduct.DroughtHook{{Name: "restart", Action: "restart_self"}}
	reloadCalled := false
	// restart_self + unsupervised → calls onReload if workflowChanged... but restart_self
	// doesn't set workflowChanged. The else branch fires (no supervisor, no binary update).
	// We just verify onReload is not called (since workflowChanged stays false).
	onReload := func() { reloadCalled = true }
	RunDroughtHooks(hooks, cfg, "", t.TempDir(), logger, time.Time{}, false, onReload)
	if reloadCalled {
		t.Error("onReload should not be called for restart_self without workflowChanged")
	}
}

// --- checkStuckDeliveries tests ---

func TestCheckStuckDeliveries_NoDeliveryItems_NoOp(t *testing.T) {
	client := newMockClient()
	// No items in the queue — should be a no-op.
	sched := testScheduler(client, newMockRunner(client))
	sched.checkStuckDeliveries(context.Background())
}

func TestCheckStuckDeliveries_ItemNotPastThreshold_Skipped(t *testing.T) {
	client := newMockClient()
	// An in_progress delivery item that is recent (well within threshold).
	item := &cistern.Droplet{
		ID:               "sd-skip",
		CurrentCataractae: "delivery",
		Status:           "in_progress",
		Assignee:         "alpha",
		Outcome:          "",
		UpdatedAt:        time.Now(), // just updated — not past threshold
	}
	client.items["sd-skip"] = item

	sched := testScheduler(client, newMockRunner(client))
	sched.checkStuckDeliveries(context.Background())

	client.mu.Lock()
	defer client.mu.Unlock()
	// Item should not have been touched.
	if client.steps["sd-skip"] != "" {
		t.Errorf("recent delivery item should not be reset, got step %q", client.steps["sd-skip"])
	}
}

func TestCheckStuckDeliveries_CancelledContext_Returns(t *testing.T) {
	client := newMockClient()
	sched := testScheduler(client, newMockRunner(client))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	// Should return immediately without panicking.
	sched.checkStuckDeliveries(ctx)
}

// --- doReloadWorkflows tests ---

func TestDoReloadWorkflows_ValidFile_UpdatesWorkflow(t *testing.T) {
	// Write a valid workflow YAML to a temp file.
	wfContent := `name: feature
cataractae:
  - name: implement
    type: agent
    identity: implementer
    on_pass: done
    on_fail: blocked
`
	wfPath := filepath.Join(t.TempDir(), "workflow.yaml")
	if err := os.WriteFile(wfPath, []byte(wfContent), 0o644); err != nil {
		t.Fatal(err)
	}

	config := aqueduct.AqueductConfig{
		Repos: []aqueduct.RepoConfig{
			{Name: "test-repo", WorkflowPath: wfPath, Cataractae: 1, Names: []string{"alpha"}, Prefix: "test"},
		},
		MaxCataractae: 1,
	}
	workflows := map[string]*aqueduct.Workflow{"test-repo": testWorkflow()}
	client := newMockClient()
	clients := map[string]CisternClient{"test-repo": client}
	sched := NewFromParts(config, workflows, clients, newMockRunner(client))

	sched.doReloadWorkflows()

	// Workflow should have been updated to the one from the file (1 step: implement).
	wf := sched.workflows["test-repo"]
	if wf == nil {
		t.Fatal("workflow should not be nil after reload")
	}
	if len(wf.Cataractae) != 1 {
		t.Errorf("reloaded workflow should have 1 cataractae, got %d", len(wf.Cataractae))
	}
	if wf.Cataractae[0].Name != "implement" {
		t.Errorf("reloaded workflow step = %q, want implement", wf.Cataractae[0].Name)
	}
}

func TestDoReloadWorkflows_InvalidFile_KeepsOldWorkflow(t *testing.T) {
	wfPath := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(wfPath, []byte("not: valid: yaml: {{{\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	config := aqueduct.AqueductConfig{
		Repos: []aqueduct.RepoConfig{
			{Name: "test-repo", WorkflowPath: wfPath, Cataractae: 1, Names: []string{"alpha"}, Prefix: "test"},
		},
		MaxCataractae: 1,
	}
	original := testWorkflow()
	workflows := map[string]*aqueduct.Workflow{"test-repo": original}
	client := newMockClient()
	clients := map[string]CisternClient{"test-repo": client}
	sched := NewFromParts(config, workflows, clients, newMockRunner(client))

	sched.doReloadWorkflows()

	// Old workflow should be preserved on parse error.
	if sched.workflows["test-repo"] != original {
		t.Error("workflow should not be replaced on parse failure")
	}
}

// --- dirtyNonContextFiles tests ---

// makeSimpleGitRepo creates a bare git repo at dir with one initial commit.
func makeSimpleGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("init\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestDirtyNonContextFiles_CleanRepo_Empty(t *testing.T) {
	dir := makeSimpleGitRepo(t)
	dirty := dirtyNonContextFiles(dir)
	if len(dirty) != 0 {
		t.Errorf("clean repo should have no dirty files, got %v", dirty)
	}
}

func TestDirtyNonContextFiles_UntrackedFile_Ignored(t *testing.T) {
	dir := makeSimpleGitRepo(t)
	// Untracked file ("??" prefix in git status --porcelain) should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "untracked.go"), []byte("// new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirty := dirtyNonContextFiles(dir)
	if len(dirty) != 0 {
		t.Errorf("untracked files should be ignored, got %v", dirty)
	}
}

func TestDirtyNonContextFiles_ModifiedNonContext_Reported(t *testing.T) {
	dir := makeSimpleGitRepo(t)
	// Modify a tracked file (not CONTEXT.md) — should appear as dirty.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("modified\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirty := dirtyNonContextFiles(dir)
	if len(dirty) == 0 {
		t.Error("modified tracked file should appear in dirty list, got empty")
	}
}

func TestDirtyNonContextFiles_OnlyContextMd_Empty(t *testing.T) {
	dir := makeSimpleGitRepo(t)
	// Add CONTEXT.md to the repo first so it's tracked.
	if err := os.WriteFile(filepath.Join(dir, "CONTEXT.md"), []byte("context\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "CONTEXT.md"},
		{"git", "commit", "-m", "add context"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	// Now modify CONTEXT.md — it should be excluded from the dirty list.
	if err := os.WriteFile(filepath.Join(dir, "CONTEXT.md"), []byte("updated context\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirty := dirtyNonContextFiles(dir)
	for _, f := range dirty {
		if strings.Contains(f, "CONTEXT.md") {
			t.Errorf("CONTEXT.md should be excluded from dirty list, got %v", dirty)
		}
	}
}

func TestDirtyNonContextFiles_NonGitDir_ReturnsNil(t *testing.T) {
	dir := t.TempDir() // not a git repo
	dirty := dirtyNonContextFiles(dir)
	if dirty != nil {
		t.Errorf("non-git dir should return nil, got %v", dirty)
	}
}
