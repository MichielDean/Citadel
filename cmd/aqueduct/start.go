package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MichielDean/cistern/internal/aqueduct"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Load config, validate workflows, and start the scheduler loop",
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := aqueduct.ParseAqueductConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfgDir := filepath.Dir(cfgPath)
	workflows := make(map[string]*aqueduct.Workflow, len(cfg.Repos))
	for _, repo := range cfg.Repos {
		if repo.WorkflowPath == "" {
			return fmt.Errorf("repo %q: workflow_path is required", repo.Name)
		}
		wfPath := repo.WorkflowPath
		if !filepath.IsAbs(wfPath) {
			wfPath = filepath.Join(cfgDir, wfPath)
		}
		w, err := aqueduct.ParseWorkflow(wfPath)
		if err != nil {
			return fmt.Errorf("repo %q workflow %q: %w", repo.Name, repo.WorkflowPath, err)
		}
		workflows[repo.Name] = w
	}

	fmt.Printf("aqueduct: loaded %d repo(s), max_cataractae=%d\n", len(cfg.Repos), cfg.MaxCataractae)
	for _, repo := range cfg.Repos {
		w := workflows[repo.Name]
		fmt.Printf("  %s: workflow=%q (%d cataractae), operators=%d\n",
			repo.Name, w.Name, len(w.Cataractae), repo.Cataractae)
	}

	fmt.Println("aqueduct: scheduler running (ctrl-c to stop)")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sig:
			fmt.Println("\nfarm: shutting down")
			return nil
		case <-ticker.C:
			// Aqueduct tick placeholder — will poll for ready droplets.
		}
	}
}
