// Package castellarius implements the Castellarius — the overseer of all aqueducts.
//
// It polls the work cistern for each configured repo, assigns droplets to
// named operators, runs workflow cataractae via an injected CataractaRunner, reads
// outcomes, and routes to the next cataracta via deterministic workflow rules.
// No AI in the Castellarius — pure state machine.
package castellarius

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MichielDean/cistern/internal/cistern"
	"github.com/MichielDean/cistern/internal/aqueduct"
)

// CisternClient is the interface for interacting with the work cistern.
// *cistern.Client satisfies this interface.
type CisternClient interface {
	GetReady(repo string) (*cistern.Droplet, error)
	Assign(id, worker, step string) error

	AddNote(id, step, content string) error
	GetNotes(id string) ([]cistern.CataractaNote, error)
	Escalate(id, reason string) error
	CloseItem(id string) error
	List(repo, status string) ([]*cistern.Droplet, error)
	Purge(olderThan time.Duration, dryRun bool) (int, error)
}

// CataractaRunner executes a single workflow step.
// The scheduler calls Run and reads the returned Outcome to decide routing.
// Implementations handle agent spawning, automated commands, etc.
type CataractaRunner interface {
	Run(ctx context.Context, req CataractaRequest) (*Outcome, error)
}

// CataractaRequest contains everything needed to execute a workflow step.
type CataractaRequest struct {
	Item       *cistern.Droplet
	Step       aqueduct.WorkflowCataracta
	Workflow   *aqueduct.Workflow
	RepoConfig aqueduct.RepoConfig
	WorkerName string
	Notes      []cistern.CataractaNote // context from previous steps
}

// Castellarius is the core loop that polls for work, assigns it to operators,
// and routes outcomes through workflow cataractae.
// and routes outcomes through workflow cataractae.
type Castellarius struct {
	config          aqueduct.AqueductConfig
	workflows       map[string]*aqueduct.Workflow
	clients         map[string]CisternClient
	pools           map[string]*WorkerPool
	runner          CataractaRunner
	logger          *slog.Logger
	pollInterval    time.Duration
	sandboxRoot     string
	cleanupInterval time.Duration
	dbPath          string
	wasDrought         bool
}

// Option configures a flow.
type Option func(*Castellarius)

// WithLogger sets the logger.
func WithLogger(l *slog.Logger) Option {
	return func(s *Castellarius) { s.logger = l }
}

// WithPollInterval sets how often the scheduler polls for work.
func WithPollInterval(d time.Duration) Option {
	return func(s *Castellarius) { s.pollInterval = d }
}

// WithSandboxRoot sets the root directory for worker sandboxes.
func WithSandboxRoot(root string) Option {
	return func(s *Castellarius) { s.sandboxRoot = root }
}

// New creates a Castellarius from an AqueductConfig.
// Workflows are loaded from each RepoConfig.WorkflowPath.
// Each repo gets its own cistern.Client scoped by prefix.
func New(config aqueduct.AqueductConfig, dbPath string, runner CataractaRunner, opts ...Option) (*Castellarius, error) {
	s := &Castellarius{
		config:       config,
		workflows:    make(map[string]*aqueduct.Workflow),
		clients:      make(map[string]CisternClient),
		pools:        make(map[string]*WorkerPool),
		runner:       runner,
		logger:       slog.Default(),
		pollInterval: 10 * time.Second,
		dbPath:       dbPath,
	}
	for _, o := range opts {
		o(s)
	}

	if s.sandboxRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("castellarius: home dir: %w", err)
		}
		s.sandboxRoot = filepath.Join(home, ".cistern", "sandboxes")
	}

	if config.CleanupInterval != "" {
		d, err := time.ParseDuration(config.CleanupInterval)
		if err != nil {
			return nil, fmt.Errorf("castellarius: invalid cleanup_interval %q: %w", config.CleanupInterval, err)
		}
		s.cleanupInterval = d
	} else {
		s.cleanupInterval = 24 * time.Hour
	}

	for _, repo := range config.Repos {
		wf, err := aqueduct.ParseWorkflow(repo.WorkflowPath)
		if err != nil {
			return nil, fmt.Errorf("load workflow for %s: %w", repo.Name, err)
		}
		s.workflows[repo.Name] = wf

		client, err := cistern.New(dbPath, repo.Prefix)
		if err != nil {
			return nil, fmt.Errorf("queue for %s: %w", repo.Name, err)
		}
		s.clients[repo.Name] = client

		names := repo.Names
		if len(names) == 0 {
			names = defaultWorkerNames(repo.Cataractae)
		}
		s.pools[repo.Name] = NewWorkerPool(repo.Name, names)
	}

	return s, nil
}

// NewFromParts creates a Castellarius with pre-built components (for testing).
func NewFromParts(
	config aqueduct.AqueductConfig,
	workflows map[string]*aqueduct.Workflow,
	clients map[string]CisternClient,
	runner CataractaRunner,
	opts ...Option,
) *Castellarius {
	s := &Castellarius{
		config:       config,
		workflows:    workflows,
		clients:      clients,
		pools:        make(map[string]*WorkerPool),
		runner:       runner,
		logger:       slog.Default(),
		pollInterval: 10 * time.Second,
	}
	for _, o := range opts {
		o(s)
	}

	for _, repo := range config.Repos {
		names := repo.Names
		if len(names) == 0 {
			names = defaultWorkerNames(repo.Cataractae)
		}
		s.pools[repo.Name] = NewWorkerPool(repo.Name, names)
	}

	return s
}

func defaultWorkerNames(n int) []string {
	if n <= 0 {
		n = 1
	}
	names := make([]string, n)
	for i := range names {
		names[i] = fmt.Sprintf("worker-%d", i)
	}
	return names
}

// Run starts the scheduler loop. It blocks until ctx is cancelled.
func (s *Castellarius) Run(ctx context.Context) error {
	s.logger.Info("Cistern online. Aqueducts open.",
		"repos", len(s.config.Repos),
		"cataractae", s.config.MaxCataractae,
	)

	s.recoverInProgress()

	if s.cleanupInterval > 0 {
		go func() {
			ticker := time.NewTicker(s.cleanupInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					s.purgeOldItems()
				}
			}
		}()
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Aqueducts closed.")
			return ctx.Err()
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// purgeOldItems deletes closed/escalated items older than retention_days across all repos.
func (s *Castellarius) purgeOldItems() {
	retentionDays := s.config.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 90
	}
	olderThan := time.Duration(retentionDays) * 24 * time.Hour

	total := 0
	for _, repo := range s.config.Repos {
		client := s.clients[repo.Name]
		n, err := client.Purge(olderThan, false)
		if err != nil {
			s.logger.Error("purge failed", "repo", repo.Name, "error", err)
			continue
		}
		if n > 0 {
			s.logger.Info("purged items", "repo", repo.Name, "count", n)
		}
		total += n
	}
	s.logger.Info("purge complete", "total", total)
}

// Tick runs a single poll cycle across all repos. Exported for testing.
func (s *Castellarius) Tick(ctx context.Context) {
	s.tick(ctx)
}

func (s *Castellarius) tick(ctx context.Context) {
	for _, repo := range s.config.Repos {
		if err := ctx.Err(); err != nil {
			return
		}
		s.tickRepo(ctx, repo)
	}

	// Drought edge detection: fire hooks on transition from busy → drought.
	isDrought := s.totalBusy() == 0
	if isDrought && !s.wasDrought {
			// Entering drought state — run drought hooks.
		if len(s.config.DroughtHooks) > 0 {
			s.logger.Info("Drought protocols running.")
			go RunDroughtHooks(s.config.DroughtHooks, &s.config, s.dbPath, s.sandboxRoot, s.logger)
		}
	}
	s.wasDrought = isDrought
}

func (s *Castellarius) tickRepo(ctx context.Context, repo aqueduct.RepoConfig) {
	pool := s.pools[repo.Name]
	client := s.clients[repo.Name]
	wf := s.workflows[repo.Name]

	for {
		worker := pool.AvailableWorker()
		if worker == nil {
			return
		}

		if s.totalBusy() >= s.config.MaxCataractae {
			return
		}

		item, err := client.GetReady(repo.Name)
		if err != nil {
			s.logger.Error("poll failed", "repo", repo.Name, "error", err)
			return
		}
		if item == nil {
			return
		}

		step := currentCataracta(item, wf)
		if step == nil {
			s.logger.Error("no step found", "repo", repo.Name, "droplet", item.ID)
			return
		}

		pool.Assign(worker, item.ID, step.Name)
		go s.runStep(ctx, worker, pool, item, *step, repo)
	}
}

func (s *Castellarius) totalBusy() int {
	total := 0
	for _, pool := range s.pools {
		total += pool.BusyCount()
	}
	return total
}

// currentCataracta determines which workflow step a work item is at.
// If the item has a current_step, look up that step.
// Otherwise, start at the first step in the aqueduct.
func currentCataracta(item *cistern.Droplet, wf *aqueduct.Workflow) *aqueduct.WorkflowCataracta {
	if item.CurrentCataracta != "" {
		return lookupCataracta(wf, item.CurrentCataracta)
	}
	if len(wf.Cataractae) > 0 {
		return &wf.Cataractae[0]
	}
	return nil
}

func lookupCataracta(wf *aqueduct.Workflow, name string) *aqueduct.WorkflowCataracta {
	for i := range wf.Cataractae {
		if wf.Cataractae[i].Name == name {
			return &wf.Cataractae[i]
		}
	}
	return nil
}

func (s *Castellarius) runStep(
	ctx context.Context,
	worker *Worker,
	pool *WorkerPool,
	item *cistern.Droplet,
	step aqueduct.WorkflowCataracta,
	repo aqueduct.RepoConfig,
) {
	defer pool.Release(worker)

	client := s.clients[repo.Name]
	wf := s.workflows[repo.Name]

	s.logger.Info("Droplet entering cataracta",
		"droplet", item.ID,
		"operator", worker.Name,
		"cataracta", step.Name,
	)

	// Mark item as in-progress with the assigned worker and step.
	if err := client.Assign(item.ID, worker.Name, step.Name); err != nil {
		s.logger.Error("assign failed", "droplet", item.ID, "error", err)
		return
	}

	// Gather prior notes for context forwarding.
	notes, err := client.GetNotes(item.ID)
	if err != nil {
		s.logger.Error("get notes failed", "droplet", item.ID, "error", err)
		notes = nil
	}

	req := CataractaRequest{
		Item:       item,
		Step:       step,
		Workflow:   wf,
		RepoConfig: repo,
		WorkerName: worker.Name,
		Notes:      notes,
	}

	// Apply step timeout.
	stepCtx := ctx
	if step.TimeoutMinutes > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, time.Duration(step.TimeoutMinutes)*time.Minute)
		defer cancel()
	}

	// Execute the step.
	outcome, err := s.runner.Run(stepCtx, req)
	if err != nil {
		// Agent crash or timeout: item stays at current cataracta for requeue.
		s.logger.Error("step execution failed",
			"repo", repo.Name,
			"droplet", item.ID,
			"cataracta", step.Name,
			"operator", worker.Name,
			"error", err,
		)
		return
	}

	switch outcome.Result {
	case ResultPass:
		s.logger.Info("Droplet cleared cataracta", "droplet", item.ID, "cataracta", step.Name)
	case ResultRecirculate:
		s.logger.Info("Droplet recirculated \u2014 cataracta returned it upstream", "droplet", item.ID, "cataracta", step.Name)
	case ResultFail:
		s.logger.Info("Droplet stagnant at cataracta", "droplet", item.ID, "cataracta", step.Name)
	default:
		s.logger.Info("Droplet outcome", "droplet", item.ID, "cataracta", step.Name, "result", outcome.Result)
	}

	// Attach notes from this step.
	if outcome.Notes != "" {
		if err := client.AddNote(item.ID, step.Name, outcome.Notes); err != nil {
			s.logger.Error("add note failed", "droplet", item.ID, "error", err)
		}
	}

	// Persist metadata notes (e.g., pr_url from pr-create) for downstream steps.
	for _, mn := range outcome.MetaNotes {
		if err := client.AddNote(item.ID, step.Name, mn); err != nil {
			s.logger.Error("add meta note failed", "droplet", item.ID, "error", err)
		}
	}

	// Route to next step.
	next := route(step, outcome.Result)
	if next == "" {
		reason := fmt.Sprintf("no route from step %q for result %q", step.Name, outcome.Result)
		s.logger.Warn("no route", "droplet", item.ID, "cataracta", step.Name, "result", outcome.Result)
		if err := client.Escalate(item.ID, reason); err != nil {
			s.logger.Error("escalate failed", "droplet", item.ID, "error", err)
		}
		return
	}

	// Apply complexity skip rules: advance past skipped steps.
	// Derived from each step's skip_for field in the workflow YAML.
	skipSteps := wf.SkipCataractaeForLevel(item.Complexity)
	next = advanceSkippedCataractae(next, wf, skipSteps)

	// For critical droplets (complexity 4), insert a human gate before merge.
	if wf.Complexity.RequireHumanForLevel(item.Complexity) && next == "merge" {
		next = "human"
	}

	if isTerminal(next) {
		s.handleTerminal(client, item.ID, next, step.Name)
		return
	}

	// Advance item to next step (open for the next poll cycle).
	if err := client.Assign(item.ID, "", next); err != nil {
		s.logger.Error("advance step failed", "droplet", item.ID, "next", next, "error", err)
	}
}

// route determines the next step name based on the outcome result.
func route(step aqueduct.WorkflowCataracta, result Result) string {
	switch result {
	case ResultPass:
		return step.OnPass
	case ResultFail:
		return step.OnFail
	case ResultRecirculate:
		return step.OnRecirculate
	case ResultEscalate:
		return step.OnEscalate
	default:
		return step.OnFail
	}
}

// advanceSkippedCataractae walks the workflow from nextStep, skipping any step whose name
// appears in skipSteps. It follows on_pass links to find the next non-skipped step.
// Returns "done" if all remaining steps are skipped.
func advanceSkippedCataractae(nextStep string, wf *aqueduct.Workflow, skipSteps []string) string {
	if len(skipSteps) == 0 {
		return nextStep
	}
	skip := make(map[string]bool, len(skipSteps))
	for _, s := range skipSteps {
		skip[s] = true
	}
	current := nextStep
	for skip[current] {
		step := lookupCataracta(wf, current)
		if step == nil || step.OnPass == "" {
			return "done"
		}
		current = step.OnPass
	}
	return current
}

// isTerminal returns true if the target is a terminal state.
func isTerminal(name string) bool {
	switch strings.ToLower(name) {
	case "done", "blocked", "human", "escalate":
		return true
	}
	return false
}

func (s *Castellarius) handleTerminal(client CisternClient, itemID, terminal, fromStep string) {
	switch strings.ToLower(terminal) {
	case "done":
		s.logger.Info("Droplet flows free", "droplet", itemID)
		if err := client.CloseItem(itemID); err != nil {
			s.logger.Error("close failed", "droplet", itemID, "error", err)
		}
	case "blocked", "human", "escalate":
		s.logger.Info("Droplet stagnant at terminal", "droplet", itemID, "terminal", terminal, "from_cataracta", fromStep)
		reason := fmt.Sprintf("reached terminal %q from cataracta %q", terminal, fromStep)
		if err := client.Escalate(itemID, reason); err != nil {
			s.logger.Error("escalate at terminal failed", "droplet", itemID, "error", err)
		}
	}
}

// recoverInProgress recovers items left in_progress after a restart.
// For each in_progress item, it checks for an outcome.json in the worker
// sandbox directory. If found, the outcome is processed and the item is
// advanced. If not found, the item is reset to open at its current step.
func (s *Castellarius) recoverInProgress() {
	for _, repo := range s.config.Repos {
		client := s.clients[repo.Name]
		wf := s.workflows[repo.Name]

		items, err := client.List(repo.Name, "in_progress")
		if err != nil {
			s.logger.Error("recovery: list in_progress failed", "repo", repo.Name, "error", err)
			continue
		}

		for _, item := range items {
			step := currentCataracta(item, wf)
			if step == nil {
				s.logger.Warn("recovery: no step found", "repo", repo.Name, "droplet", item.ID, "cataracta", item.CurrentCataracta)
				continue
			}

			// Check for outcome.json in the worker's sandbox directory.
			sandboxDir := filepath.Join(s.sandboxRoot, repo.Name, item.Assignee)
			outcomePath := filepath.Join(sandboxDir, "outcome.json")

			outcome, err := ReadOutcome(outcomePath)
			if err != nil {
				// No outcome found — reset item to open at current step for retry.
				s.logger.Info("recovery: resetting to open",
					"repo", repo.Name,
					"droplet", item.ID,
					"cataracta", item.CurrentCataracta,
				)
				if err := client.Assign(item.ID, "", item.CurrentCataracta); err != nil {
					s.logger.Error("recovery: reset failed", "droplet", item.ID, "error", err)
				}
				continue
			}

			s.logger.Info("recovery: processing leftover outcome",
				"repo", repo.Name,
				"droplet", item.ID,
				"cataracta", item.CurrentCataracta,
				"result", outcome.Result,
			)

			// Attach notes from the recovered outcome.
			if outcome.Notes != "" {
				if err := client.AddNote(item.ID, step.Name, outcome.Notes); err != nil {
					s.logger.Error("recovery: add note failed", "droplet", item.ID, "error", err)
				}
			}

			// Route to next step.
			next := route(*step, outcome.Result)
			if next == "" {
				reason := fmt.Sprintf("recovery: no route from step %q for result %q", step.Name, outcome.Result)
				s.logger.Warn("recovery: no route", "droplet", item.ID)
				if err := client.Escalate(item.ID, reason); err != nil {
					s.logger.Error("recovery: escalate failed", "droplet", item.ID, "error", err)
				}
				continue
			}

			if isTerminal(next) {
				s.handleTerminal(client, item.ID, next, step.Name)
				continue
			}

			if err := client.Assign(item.ID, "", next); err != nil {
				s.logger.Error("recovery: advance failed", "droplet", item.ID, "next", next, "error", err)
			}
		}
	}
}

// WriteContext writes a CONTEXT.md file with notes from previous steps.
// Call this before spawning the next agent to provide context from prior steps.
func WriteContext(dir string, notes []cistern.CataractaNote) error {
	if len(notes) == 0 {
		return nil
	}

	var b []byte
	b = append(b, "# Context from Previous Steps\n\n"...)
	for _, n := range notes {
		header := n.CataractaName
		if header == "" {
			header = "unknown"
		}
		b = append(b, fmt.Sprintf("## Step: %s\n\n%s\n\n", header, n.Content)...)
	}

	return os.WriteFile(filepath.Join(dir, "CONTEXT.md"), b, 0o644)
}
