package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MichielDean/bullet-farm/internal/bd"
	"github.com/MichielDean/bullet-farm/internal/workflow"
)

// --- mocks ---

type mockClient struct {
	mu         sync.Mutex
	readyBeads []*bd.Bead
	readyCalls int
	steps      map[string]string
	attempts   map[string]int
	notes      map[string][]bd.StepNote
	escalated  map[string]string
	attached   []attachedNote
}

type attachedNote struct {
	id, fromStep, notes string
}

func newMockClient() *mockClient {
	return &mockClient{
		steps:     make(map[string]string),
		attempts:  make(map[string]int),
		notes:     make(map[string][]bd.StepNote),
		escalated: make(map[string]string),
	}
}

func (m *mockClient) GetReady(rig string) (*bd.Bead, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readyCalls++
	if len(m.readyBeads) == 0 {
		return nil, nil
	}
	b := m.readyBeads[0]
	m.readyBeads = m.readyBeads[1:]
	return b, nil
}

func (m *mockClient) UpdateStep(id, step string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.steps[id] = step
	return nil
}

func (m *mockClient) IncrementAttempts(id, step string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := id + ":" + step
	m.attempts[key]++
	return m.attempts[key], nil
}

func (m *mockClient) AttachNotes(id, fromStep, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attached = append(m.attached, attachedNote{id, fromStep, notes})
	return nil
}

func (m *mockClient) GetNotes(id string) ([]bd.StepNote, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.notes[id], nil
}

func (m *mockClient) Escalate(id, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.escalated[id] = reason
	return nil
}

type mockRunner struct {
	mu       sync.Mutex
	outcomes map[string]*Outcome
	calls    []StepRequest
	err      error
	done     chan struct{} // closed after each Run call
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		outcomes: make(map[string]*Outcome),
		done:     make(chan struct{}, 16),
	}
}

func (r *mockRunner) Run(_ context.Context, req StepRequest) (*Outcome, error) {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		r.done <- struct{}{}
	}()
	r.calls = append(r.calls, req)
	if r.err != nil {
		return nil, r.err
	}
	if o, ok := r.outcomes[req.Step.Name]; ok {
		return o, nil
	}
	return &Outcome{Result: ResultPass}, nil
}

func (r *mockRunner) waitCalls(n int, timeout time.Duration) bool {
	deadline := time.After(timeout)
	for range n {
		select {
		case <-r.done:
		case <-deadline:
			return false
		}
	}
	return true
}

type blockingRunner struct {
	ch   chan struct{}
	done chan struct{}
}

func newBlockingRunner() *blockingRunner {
	return &blockingRunner{
		ch:   make(chan struct{}),
		done: make(chan struct{}, 16),
	}
}

func (r *blockingRunner) Run(ctx context.Context, _ StepRequest) (*Outcome, error) {
	r.done <- struct{}{}
	select {
	case <-r.ch:
		return &Outcome{Result: ResultPass}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// --- helpers ---

func testWorkflow() *workflow.Workflow {
	return &workflow.Workflow{
		Name: "test",
		Steps: []workflow.WorkflowStep{
			{
				Name:   "implement",
				Type:   workflow.StepTypeAgent,
				OnPass: "review",
				OnFail: "blocked",
			},
			{
				Name:       "review",
				Type:       workflow.StepTypeAgent,
				OnPass:     "done",
				OnFail:     "implement",
				OnRevision: "implement",
			},
		},
	}
}

func testConfig() workflow.FarmConfig {
	return workflow.FarmConfig{
		Repos: []workflow.RepoConfig{
			{
				Name:     "test-repo",
				Workers:  2,
				Names:    []string{"alpha", "beta"},
				BdPrefix: "test",
			},
		},
		MaxTotalWorkers: 4,
	}
}

func testScheduler(client BeadClient, runner StepRunner) *Scheduler {
	config := testConfig()
	workflows := map[string]*workflow.Workflow{"test-repo": testWorkflow()}
	clients := map[string]BeadClient{"test-repo": client}
	return NewFromParts(config, workflows, clients, runner)
}

// --- tests ---

func TestRoute(t *testing.T) {
	step := workflow.WorkflowStep{
		OnPass:     "review",
		OnFail:     "blocked",
		OnRevision: "implement",
		OnEscalate: "human",
	}

	tests := []struct {
		result Result
		want   string
	}{
		{ResultPass, "review"},
		{ResultFail, "blocked"},
		{ResultRevision, "implement"},
		{ResultEscalate, "human"},
		{Result("unknown"), "blocked"},
	}

	for _, tt := range tests {
		got := route(step, tt.result)
		if got != tt.want {
			t.Errorf("route(%q) = %q, want %q", tt.result, got, tt.want)
		}
	}
}

func TestIsTerminal(t *testing.T) {
	for _, name := range []string{"done", "blocked", "human", "escalate", "Done", "BLOCKED"} {
		if !isTerminal(name) {
			t.Errorf("isTerminal(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"implement", "review", "qa", ""} {
		if isTerminal(name) {
			t.Errorf("isTerminal(%q) = true, want false", name)
		}
	}
}

func TestCurrentStep_FirstStep(t *testing.T) {
	wf := testWorkflow()
	bead := &bd.Bead{ID: "b1"}

	step := currentStep(bead, wf)
	if step == nil || step.Name != "implement" {
		t.Fatalf("expected first step 'implement', got %v", step)
	}
}

func TestCurrentStep_FromMetadata(t *testing.T) {
	wf := testWorkflow()
	bead := &bd.Bead{
		ID:       "b1",
		Metadata: map[string]any{"step": "review"},
	}

	step := currentStep(bead, wf)
	if step == nil || step.Name != "review" {
		t.Fatalf("expected step 'review' from metadata, got %v", step)
	}
}

func TestCurrentStep_UnknownMetadata(t *testing.T) {
	wf := testWorkflow()
	bead := &bd.Bead{
		ID:       "b1",
		Metadata: map[string]any{"step": "nonexistent"},
	}

	step := currentStep(bead, wf)
	if step != nil {
		t.Fatalf("expected nil for unknown step, got %v", step)
	}
}

func TestTick_AssignsWork(t *testing.T) {
	client := newMockClient()
	client.readyBeads = []*bd.Bead{{ID: "b1", Title: "test bead"}}

	runner := newMockRunner()
	runner.outcomes["implement"] = &Outcome{Result: ResultPass, Notes: "done"}

	sched := testScheduler(client, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(1, time.Second) {
		t.Fatal("timed out waiting for runner call")
	}

	runner.mu.Lock()
	defer runner.mu.Unlock()
	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 runner call, got %d", len(runner.calls))
	}
	if runner.calls[0].Step.Name != "implement" {
		t.Errorf("expected step 'implement', got %q", runner.calls[0].Step.Name)
	}
	if runner.calls[0].WorkerName != "alpha" {
		t.Errorf("expected worker 'alpha', got %q", runner.calls[0].WorkerName)
	}
}

func TestTick_RoutesToNextStep(t *testing.T) {
	client := newMockClient()
	client.readyBeads = []*bd.Bead{{ID: "b1", Title: "test"}}

	runner := newMockRunner()
	runner.outcomes["implement"] = &Outcome{Result: ResultPass, Notes: "impl done"}

	sched := testScheduler(client, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(1, time.Second) {
		t.Fatal("timed out")
	}
	// Small sleep to let post-Run routing complete.
	time.Sleep(50 * time.Millisecond)

	client.mu.Lock()
	defer client.mu.Unlock()
	if client.steps["b1"] != "review" {
		t.Errorf("expected step 'review', got %q", client.steps["b1"])
	}
}

func TestTick_TerminalDone(t *testing.T) {
	client := newMockClient()
	client.readyBeads = []*bd.Bead{
		{ID: "b1", Metadata: map[string]any{"step": "review"}},
	}

	runner := newMockRunner()
	runner.outcomes["review"] = &Outcome{Result: ResultPass}

	sched := testScheduler(client, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(1, time.Second) {
		t.Fatal("timed out")
	}
	time.Sleep(50 * time.Millisecond)

	client.mu.Lock()
	defer client.mu.Unlock()
	if client.steps["b1"] != "done" {
		t.Errorf("expected terminal 'done', got %q", client.steps["b1"])
	}
}

func TestTick_TerminalBlocked(t *testing.T) {
	client := newMockClient()
	client.readyBeads = []*bd.Bead{
		{ID: "b1", Metadata: map[string]any{"step": "implement"}},
	}

	runner := newMockRunner()
	runner.outcomes["implement"] = &Outcome{Result: ResultFail}

	sched := testScheduler(client, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(1, time.Second) {
		t.Fatal("timed out")
	}
	time.Sleep(50 * time.Millisecond)

	client.mu.Lock()
	defer client.mu.Unlock()
	if _, ok := client.escalated["b1"]; !ok {
		t.Error("expected bead escalated for terminal 'blocked'")
	}
}

func TestTick_GlobalCap(t *testing.T) {
	client := newMockClient()
	for i := range 10 {
		client.readyBeads = append(client.readyBeads, &bd.Bead{
			ID: fmt.Sprintf("b%d", i),
		})
	}

	runner := newBlockingRunner()

	config := workflow.FarmConfig{
		Repos: []workflow.RepoConfig{
			{Name: "r1", Workers: 3, Names: []string{"w1", "w2", "w3"}, BdPrefix: "r1"},
		},
		MaxTotalWorkers: 2,
	}
	wf := testWorkflow()
	clients := map[string]BeadClient{"r1": client}
	workflows := map[string]*workflow.Workflow{"r1": wf}
	sched := NewFromParts(config, workflows, clients, runner)

	// Tick multiple times.
	for range 5 {
		sched.Tick(context.Background())
	}

	total := sched.totalBusy()
	if total > 2 {
		t.Errorf("global cap violated: %d busy workers (cap=2)", total)
	}

	close(runner.ch)
}

func TestTick_RetryBudgetOK(t *testing.T) {
	client := newMockClient()
	client.attempts["b1:implement"] = 2 // will become 3
	client.readyBeads = []*bd.Bead{{ID: "b1"}}

	runner := newMockRunner()

	wf := &workflow.Workflow{
		Name: "test",
		Steps: []workflow.WorkflowStep{
			{
				Name:          "implement",
				Type:          workflow.StepTypeAgent,
				MaxIterations: 3,
				OnPass:        "done",
				OnFail:        "blocked",
			},
		},
	}

	config := testConfig()
	clients := map[string]BeadClient{"test-repo": client}
	workflows := map[string]*workflow.Workflow{"test-repo": wf}
	sched := NewFromParts(config, workflows, clients, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(1, time.Second) {
		t.Fatal("timed out — runner should be called when attempts <= max")
	}
}

func TestTick_RetryBudgetExceeded(t *testing.T) {
	client := newMockClient()
	client.attempts["b1:implement"] = 3 // will become 4, exceeds max of 3
	client.readyBeads = []*bd.Bead{{ID: "b1"}}

	runner := newMockRunner()

	wf := &workflow.Workflow{
		Name: "test",
		Steps: []workflow.WorkflowStep{
			{
				Name:          "implement",
				Type:          workflow.StepTypeAgent,
				MaxIterations: 3,
				OnPass:        "done",
				OnFail:        "blocked",
			},
		},
	}

	config := testConfig()
	clients := map[string]BeadClient{"test-repo": client}
	workflows := map[string]*workflow.Workflow{"test-repo": wf}
	sched := NewFromParts(config, workflows, clients, runner)
	sched.Tick(context.Background())

	// Wait a beat for the goroutine to complete.
	time.Sleep(200 * time.Millisecond)

	client.mu.Lock()
	defer client.mu.Unlock()
	if _, ok := client.escalated["b1"]; !ok {
		t.Error("expected escalation when retry budget exceeded")
	}

	runner.mu.Lock()
	defer runner.mu.Unlock()
	if len(runner.calls) != 0 {
		t.Error("runner should not be called when retry budget exceeded")
	}
}

func TestTick_CrashRequeue(t *testing.T) {
	client := newMockClient()
	client.readyBeads = []*bd.Bead{{ID: "b1"}}

	runner := newMockRunner()
	runner.err = fmt.Errorf("agent crashed")

	sched := testScheduler(client, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(1, time.Second) {
		t.Fatal("timed out")
	}
	time.Sleep(50 * time.Millisecond)

	client.mu.Lock()
	defer client.mu.Unlock()
	// Bead stays at "implement" — not advanced, not escalated.
	if client.steps["b1"] != "implement" {
		t.Errorf("expected step to remain 'implement' after crash, got %q", client.steps["b1"])
	}
	if _, ok := client.escalated["b1"]; ok {
		t.Error("should not escalate on crash — just requeue")
	}
}

func TestTick_NotesForwarding(t *testing.T) {
	client := newMockClient()
	client.readyBeads = []*bd.Bead{{ID: "b1"}}
	client.notes["b1"] = []bd.StepNote{
		{ID: 1, IssueID: "b1", FromStep: "refine", Text: "specs clarified"},
	}

	runner := newMockRunner()
	runner.outcomes["implement"] = &Outcome{Result: ResultPass, Notes: "code written"}

	sched := testScheduler(client, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(1, time.Second) {
		t.Fatal("timed out")
	}
	time.Sleep(50 * time.Millisecond)

	runner.mu.Lock()
	req := runner.calls[0]
	runner.mu.Unlock()

	if len(req.Notes) != 1 || req.Notes[0].FromStep != "refine" {
		t.Errorf("expected prior notes forwarded, got %v", req.Notes)
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	if len(client.attached) != 1 || client.attached[0].fromStep != "implement" {
		t.Errorf("expected notes attached from 'implement', got %v", client.attached)
	}
}

func TestTick_NoRoute(t *testing.T) {
	client := newMockClient()
	client.readyBeads = []*bd.Bead{{ID: "b1"}}

	runner := newMockRunner()
	runner.outcomes["implement"] = &Outcome{Result: ResultRevision} // no OnRevision set

	sched := testScheduler(client, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(1, time.Second) {
		t.Fatal("timed out")
	}
	time.Sleep(50 * time.Millisecond)

	client.mu.Lock()
	defer client.mu.Unlock()
	// implement step has no OnRevision → empty route → escalate.
	if _, ok := client.escalated["b1"]; !ok {
		t.Error("expected escalation when no route exists")
	}
}

func TestTick_NoWorkAvailable(t *testing.T) {
	client := newMockClient()
	runner := newMockRunner()

	sched := testScheduler(client, runner)
	sched.Tick(context.Background())
	time.Sleep(50 * time.Millisecond)

	runner.mu.Lock()
	defer runner.mu.Unlock()
	if len(runner.calls) != 0 {
		t.Error("expected no runner calls when no work available")
	}
}

func TestTick_PerRepoIsolation(t *testing.T) {
	client1 := newMockClient()
	client1.readyBeads = []*bd.Bead{{ID: "r1-b1"}}
	client2 := newMockClient()
	client2.readyBeads = []*bd.Bead{{ID: "r2-b1"}}

	runner := newMockRunner()

	config := workflow.FarmConfig{
		Repos: []workflow.RepoConfig{
			{Name: "repo1", Workers: 1, Names: []string{"w1"}, BdPrefix: "r1"},
			{Name: "repo2", Workers: 1, Names: []string{"w2"}, BdPrefix: "r2"},
		},
		MaxTotalWorkers: 10,
	}
	wf := testWorkflow()
	clients := map[string]BeadClient{"repo1": client1, "repo2": client2}
	workflows := map[string]*workflow.Workflow{"repo1": wf, "repo2": wf}
	sched := NewFromParts(config, workflows, clients, runner)
	sched.Tick(context.Background())

	if !runner.waitCalls(2, time.Second) {
		t.Fatal("timed out waiting for 2 runner calls")
	}

	runner.mu.Lock()
	defer runner.mu.Unlock()
	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 runner calls (one per repo), got %d", len(runner.calls))
	}

	for _, call := range runner.calls {
		if call.Bead.ID == "r1-b1" && call.WorkerName != "w1" {
			t.Errorf("repo1 bead assigned to wrong worker: %s", call.WorkerName)
		}
		if call.Bead.ID == "r2-b1" && call.WorkerName != "w2" {
			t.Errorf("repo2 bead assigned to wrong worker: %s", call.WorkerName)
		}
	}
}

func TestRun_CancelledContext(t *testing.T) {
	client := newMockClient()
	runner := newMockRunner()
	sched := testScheduler(client, runner)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sched.Run(ctx)
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got %v", err)
	}
}

func TestWorkerPool_Basic(t *testing.T) {
	pool := NewWorkerPool("repo", []string{"a", "b"})

	w := pool.IdleWorker()
	if w == nil || w.Name != "a" {
		t.Fatalf("expected first idle worker 'a', got %v", w)
	}

	pool.Assign(w, "bead-1", "implement")
	if pool.BusyCount() != 1 {
		t.Errorf("expected 1 busy, got %d", pool.BusyCount())
	}

	w2 := pool.IdleWorker()
	if w2 == nil || w2.Name != "b" {
		t.Fatalf("expected second idle worker 'b', got %v", w2)
	}

	pool.Assign(w2, "bead-2", "review")
	if pool.BusyCount() != 2 {
		t.Errorf("expected 2 busy, got %d", pool.BusyCount())
	}

	if pool.IdleWorker() != nil {
		t.Error("expected nil when all workers busy")
	}

	pool.Release(w)
	if pool.BusyCount() != 1 {
		t.Errorf("expected 1 busy after release, got %d", pool.BusyCount())
	}

	w3 := pool.IdleWorker()
	if w3 == nil || w3.Name != "a" {
		t.Fatalf("expected 'a' available after release, got %v", w3)
	}
}

func TestDefaultWorkerNames(t *testing.T) {
	names := defaultWorkerNames(3)
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	if names[0] != "worker-0" {
		t.Errorf("expected 'worker-0', got %q", names[0])
	}

	names = defaultWorkerNames(0)
	if len(names) != 1 {
		t.Errorf("expected 1 name for n=0, got %d", len(names))
	}
}

func TestWriteContext(t *testing.T) {
	dir := t.TempDir()
	notes := []bd.StepNote{
		{FromStep: "implement", Text: "wrote the feature"},
		{FromStep: "review", Text: "needs error handling"},
	}

	if err := WriteContext(dir, notes); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CONTEXT.md"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "## Step: implement") {
		t.Error("missing implement step header")
	}
	if !strings.Contains(content, "wrote the feature") {
		t.Error("missing implement note text")
	}
	if !strings.Contains(content, "## Step: review") {
		t.Error("missing review step header")
	}
	if !strings.Contains(content, "needs error handling") {
		t.Error("missing review note text")
	}
}

func TestWriteContext_Empty(t *testing.T) {
	dir := t.TempDir()
	if err := WriteContext(dir, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "CONTEXT.md")); !os.IsNotExist(err) {
		t.Error("expected no CONTEXT.md for empty notes")
	}
}

func TestReadOutcome(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "outcome.json")
	o := Outcome{
		Result: ResultPass,
		Notes:  "all good",
		Annotations: []Annotation{
			{File: "main.go", Line: 10, Comment: "nice"},
		},
	}
	data, _ := json.Marshal(o)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadOutcome(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Result != ResultPass {
		t.Errorf("result = %q, want %q", got.Result, ResultPass)
	}
	if got.Notes != "all good" {
		t.Errorf("notes = %q, want 'all good'", got.Notes)
	}
	if len(got.Annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(got.Annotations))
	}
}

func TestReadOutcome_NotFound(t *testing.T) {
	_, err := ReadOutcome("/nonexistent/outcome.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLookupStep(t *testing.T) {
	wf := testWorkflow()

	if s := lookupStep(wf, "implement"); s == nil || s.Name != "implement" {
		t.Error("expected to find 'implement'")
	}
	if s := lookupStep(wf, "nonexistent"); s != nil {
		t.Error("expected nil for unknown step")
	}
}
