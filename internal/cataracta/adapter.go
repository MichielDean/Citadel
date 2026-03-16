package cataracta

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MichielDean/cistern/internal/gates"
	"github.com/MichielDean/cistern/internal/cistern"
	"github.com/MichielDean/cistern/internal/castellarius"
	"github.com/MichielDean/cistern/internal/aqueduct"
)

// Adapter wraps Runner instances to implement castellarius.CataractaRunner.
type Adapter struct {
	runners  map[string]*Runner // keyed by repo name
	executor *gates.Executor
}

// NewAdapter creates an Adapter with a Runner for each configured repo.
func NewAdapter(configs []aqueduct.RepoConfig, workflows map[string]*aqueduct.Workflow, queueClients map[string]*cistern.Client) (*Adapter, error) {
	runners := make(map[string]*Runner, len(configs))
	for _, repo := range configs {
		wf, ok := workflows[repo.Name]
		if !ok {
			return nil, fmt.Errorf("adapter: no workflow for repo %q", repo.Name)
		}
		client, ok := queueClients[repo.Name]
		if !ok {
			return nil, fmt.Errorf("adapter: no queue client for repo %q", repo.Name)
		}
		r, err := New(Config{
			Repo:        repo,
			Workflow:    wf,
			CisternClient: client,
		})
		if err != nil {
			return nil, fmt.Errorf("adapter: runner for %q: %w", repo.Name, err)
		}
		runners[repo.Name] = r
	}
	return &Adapter{
		runners:  runners,
		executor: gates.New(),
	}, nil
}

// Run implements castellarius.CataractaRunner by delegating to the appropriate cataracta.
func (a *Adapter) Run(ctx context.Context, req castellarius.CataractaRequest) (*castellarius.Outcome, error) {
	// Automated steps are handled by the automated executor, not Claude.
	if req.Step.Type == aqueduct.CataractaTypeAutomated {
		return a.runAutomated(ctx, req), nil
	}

	r, ok := a.runners[req.RepoConfig.Name]
	if !ok {
		return nil, fmt.Errorf("adapter: no runner for repo %q", req.RepoConfig.Name)
	}

	// Find the worker by name.
	var worker *Worker
	r.mu.Lock()
	for _, w := range r.workers {
		if w.Name == req.WorkerName {
			worker = w
			break
		}
	}
	r.mu.Unlock()

	if worker == nil {
		return nil, fmt.Errorf("adapter: worker %q not found in repo %q", req.WorkerName, req.RepoConfig.Name)
	}

	step := req.Step
	outcome, err := r.RunStep(worker, req.Item, &step)
	if err != nil {
		return nil, err
	}

	return convertOutcome(outcome), nil
}

// runAutomated dispatches an automated step through the automated executor.
func (a *Adapter) runAutomated(ctx context.Context, req castellarius.CataractaRequest) *castellarius.Outcome {
	home, _ := os.UserHomeDir()
	sandboxDir := filepath.Join(home, ".cistern", "sandboxes", req.RepoConfig.Name, req.WorkerName)
	branch := "feat/" + req.Item.ID

	// Build metadata from prior annotations stored as step notes with "meta:" prefix.
	metadata := make(map[string]any)
	for _, n := range req.Notes {
		if len(n.Content) > 5 && n.Content[:5] == "meta:" {
			// Format: "meta:key=value"
			kv := n.Content[5:]
			for i := 0; i < len(kv); i++ {
				if kv[i] == '=' {
					metadata[kv[:i]] = kv[i+1:]
					break
				}
			}
		}
	}

	bc := gates.DropletContext{
		ID:          req.WorkerName + "-" + req.Item.ID,
		Title:       req.Item.Title,
		Description: req.Item.Description,
		WorkDir:     sandboxDir,
		Branch:      branch,
		BaseBranch:  "main",
		Metadata:    metadata,
	}
	result := a.executor.RunStep(ctx, req.Step.Name, bc)
	out := &castellarius.Outcome{
		Result: castellarius.Result(result.Result),
		Notes:  result.Notes,
	}
	// Convert annotations to metadata notes for persistence across steps.
	if len(result.Annotations) > 0 {
		for k, v := range result.Annotations {
			out.MetaNotes = append(out.MetaNotes, fmt.Sprintf("meta:%s=%s", k, v))
		}
	}
	return out
}

// convertOutcome maps a cataracta.Outcome to a castellarius.Outcome.
func convertOutcome(ro *Outcome) *castellarius.Outcome {
	so := &castellarius.Outcome{
		Result: castellarius.Result(ro.Result),
		Notes:  ro.Notes,
	}
	if len(ro.Annotations) > 0 {
		for k, v := range ro.Annotations {
			so.Annotations = append(so.Annotations, castellarius.Annotation{
				File:    k,
				Comment: v,
			})
		}
	}
	return so
}
