package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/MichielDean/citadel/internal/queue"
	"github.com/MichielDean/citadel/internal/runner"
	"github.com/MichielDean/citadel/internal/scheduler"
	"github.com/MichielDean/citadel/internal/workflow"
	"github.com/spf13/cobra"
)

var configPath string

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Open and close the aqueducts",
}

// Deprecated alias — "ct farm" still works but warns.
var farmCmd = &cobra.Command{
	Use:    "farm",
	Short:  "Deprecated: use 'ct flow'",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "The Citadel speaks water now. Use 'ct flow' instead of 'ct farm'.")
		return cmd.Help()
	},
}

// --- flow start ---

var flowStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Open the aqueducts and start the scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfigPath()
		cfg, err := workflow.ParseFarmConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		cfgDir := filepath.Dir(cfgPath)
		workflows := make(map[string]*workflow.Workflow, len(cfg.Repos))
		for _, repo := range cfg.Repos {
			if repo.WorkflowPath == "" {
				return fmt.Errorf("repo %q: workflow_path is required", repo.Name)
			}
			wfPath := repo.WorkflowPath
			if !filepath.IsAbs(wfPath) {
				wfPath = filepath.Join(cfgDir, wfPath)
			}
			w, err := workflow.ParseWorkflow(wfPath)
			if err != nil {
				return fmt.Errorf("repo %q workflow %q: %w", repo.Name, repo.WorkflowPath, err)
			}
			workflows[repo.Name] = w
		}

		dbPath := resolveDBPath()
		queueClients := make(map[string]*queue.Client, len(cfg.Repos))
		for _, repo := range cfg.Repos {
			c, err := queue.New(dbPath, repo.Prefix)
			if err != nil {
				return fmt.Errorf("queue for %q: %w", repo.Name, err)
			}
			queueClients[repo.Name] = c
		}

		adapter, err := runner.NewAdapter(cfg.Repos, workflows, queueClients)
		if err != nil {
			return fmt.Errorf("runner adapter: %w", err)
		}

		sched, err := scheduler.New(*cfg, dbPath, adapter)
		if err != nil {
			return fmt.Errorf("scheduler: %w", err)
		}

		fmt.Println("Citadel online. Aqueducts open.")
		for _, repo := range cfg.Repos {
			w := workflows[repo.Name]
			fmt.Printf("  %s: %d valves, %d channels\n", repo.Name, len(w.Steps), repo.Workers)
		}

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		if err := sched.Run(ctx); errors.Is(err, context.Canceled) {
			fmt.Println("Aqueducts closed.")
			return nil
		} else {
			return err
		}
	},
}

// --- flow status ---

var flowStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show channels and cistern state",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfigPath()
		cfg, err := workflow.ParseFarmConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		cfgDir := filepath.Dir(cfgPath)

		// Collect channel names.
		var allNames []string
		totalWorkers := 0
		for _, repo := range cfg.Repos {
			names := repo.Names
			if len(names) == 0 {
				for i := 0; i < repo.Workers; i++ {
					names = append(names, fmt.Sprintf("worker-%d", i))
				}
			}
			allNames = append(allNames, names...)
			totalWorkers += repo.Workers

			wfPath := repo.WorkflowPath
			if wfPath != "" && !filepath.IsAbs(wfPath) {
				wfPath = filepath.Join(cfgDir, wfPath)
			}
			if wfPath != "" {
				if _, err := workflow.ParseWorkflow(wfPath); err != nil {
					fmt.Fprintf(os.Stderr, "  warning: %s workflow error: %v\n", repo.Name, err)
				}
			}
		}

		// Count drops in cistern.
		dbPath := resolveDBPath()
		var totalDrops, flowingDrops, queuedDrops int
		for _, repo := range cfg.Repos {
			c, err := queue.New(dbPath, repo.Prefix)
			if err != nil {
				continue
			}
			items, err := c.List("", "")
			c.Close()
			if err != nil {
				continue
			}
			for _, item := range items {
				if item.Status == "closed" || item.Status == "escalated" {
					continue
				}
				totalDrops++
				if item.Status == "in_progress" {
					flowingDrops++
				} else {
					queuedDrops++
				}
			}
		}

		fmt.Println("Citadel")
		fmt.Printf("Channels : %d open (%s)\n", totalWorkers, strings.Join(allNames, ", "))
		fmt.Printf("Cistern  : %d drops (%d flowing, %d queued)\n", totalDrops, flowingDrops, queuedDrops)
		fmt.Println()

		// Show active drops.
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		for _, repo := range cfg.Repos {
			c, err := queue.New(dbPath, repo.Prefix)
			if err != nil {
				continue
			}
			items, err := c.List("", "in_progress")
			c.Close()
			if err != nil {
				continue
			}
			for _, item := range items {
				assignee := item.Assignee
				if assignee == "" {
					assignee = "\u2014"
				}
				step := item.CurrentStep
				if step == "" {
					step = "\u2014"
				}
				fmt.Fprintf(tw, "%s\t%s\t[%s]\n", assignee, item.ID, step)
			}
		}
		tw.Flush()

		return nil
	},
}

// --- flow config validate ---

var flowConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Config management",
}

var flowConfigValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate config and all referenced workflow files",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := resolveConfigPath()
		if len(args) > 0 {
			path = args[0]
		}

		cfg, err := workflow.ParseFarmConfig(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config error: %v\n", err)
			return err
		}

		cfgDir := filepath.Dir(path)
		var errs []error
		for _, repo := range cfg.Repos {
			if repo.Name == "" {
				e := fmt.Errorf("repo entry missing name")
				fmt.Fprintf(os.Stderr, "  error: %v\n", e)
				errs = append(errs, e)
				continue
			}
			if repo.WorkflowPath == "" {
				e := fmt.Errorf("repo %q: workflow_path is required", repo.Name)
				fmt.Fprintf(os.Stderr, "  error: %v\n", e)
				errs = append(errs, e)
				continue
			}

			wfPath := repo.WorkflowPath
			if !filepath.IsAbs(wfPath) {
				wfPath = filepath.Join(cfgDir, wfPath)
			}

			if _, err := workflow.ParseWorkflow(wfPath); err != nil {
				e := fmt.Errorf("repo %q workflow %q: %w", repo.Name, repo.WorkflowPath, err)
				fmt.Fprintf(os.Stderr, "  error: %v\n", e)
				errs = append(errs, e)
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("validation found %d error(s)", len(errs))
		}

		fmt.Println("config valid:", path)
		return nil
	},
}

func init() {
	flowStartCmd.Flags().StringVar(&configPath, "config", "", "path to citadel config (default: ~/.citadel/citadel.yaml)")

	flowConfigCmd.AddCommand(flowConfigValidateCmd)
	flowCmd.AddCommand(flowStartCmd, flowStatusCmd, flowConfigCmd)
	rootCmd.AddCommand(flowCmd)

	// Deprecated alias — mirrors flow subcommands so "ct farm start" still routes.
	farmStartAlias := *flowStartCmd
	farmStatusAlias := *flowStatusCmd
	farmConfigAlias := *flowConfigCmd
	farmCmd.AddCommand(&farmStartAlias, &farmStatusAlias, &farmConfigAlias)
	rootCmd.AddCommand(farmCmd)
}

func resolveConfigPath() string {
	if configPath != "" {
		return configPath
	}
	if env := os.Getenv("CT_CONFIG"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "citadel.yaml"
	}
	return filepath.Join(home, ".citadel", "citadel.yaml")
}
