package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/MichielDean/cistern/internal/cistern"
	"github.com/MichielDean/cistern/internal/cataracta"
	"github.com/MichielDean/cistern/internal/castellarius"
	"github.com/MichielDean/cistern/internal/aqueduct"
	"github.com/spf13/cobra"
)

var configPath string

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Aqueduct management — start, stop, and monitor Cistern",
}

// --- flow start ---

var flowStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Open the aqueducts — load config, validate workflows, start scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfigPath()
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

		// Build per-repo queue clients for the adapter.
		dbPath := resolveDBPath()
		queueClients := make(map[string]*cistern.Client, len(cfg.Repos))
		for _, repo := range cfg.Repos {
			c, err := cistern.New(dbPath, repo.Prefix)
			if err != nil {
				return fmt.Errorf("queue for %q: %w", repo.Name, err)
			}
			queueClients[repo.Name] = c
		}

		// Build the runner adapter that implements castellarius.CataractaRunner.
		adapter, err := cataracta.NewAdapter(cfg.Repos, workflows, queueClients)
		if err != nil {
			return fmt.Errorf("runner adapter: %w", err)
		}

		// Create the castellarius.
		sched, err := castellarius.New(*cfg, dbPath, adapter)
		if err != nil {
			return fmt.Errorf("flow: %w", err)
		}

		fmt.Println("Cistern online. Aqueducts open.")
		for _, repo := range cfg.Repos {
			w := workflows[repo.Name]
			names := repoWorkerNames(repo)
			fmt.Printf("  %s: workflow=%q (%d cataractae), operators=%d (%s)\n",
				repo.Name, w.Name, len(w.Cataractae), repo.Cataractae, strings.Join(names, ", "))
		}

		fmt.Println("Ctrl-C to close aqueducts.")
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		if err := sched.Run(ctx); errors.Is(err, context.Canceled) {
			return nil
		} else {
			return err
		}
	},
}

// --- flow status ---

var flowStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cataractae, cistern levels, and droplet assignments",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfigPath()
		cfg, err := aqueduct.ParseAqueductConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Collect all configured cataracta operator names.
		var allNames []string
		for _, repo := range cfg.Repos {
			allNames = append(allNames, repoWorkerNames(repo)...)
		}

		// Open queue to count droplets.
		dbPath := resolveDBPath()
		c, err := cistern.New(dbPath, "")
		if err != nil {
			return fmt.Errorf("cistern: %w", err)
		}
		defer c.Close()

		allItems, err := c.List("", "")
		if err != nil {
			return fmt.Errorf("list droplets: %w", err)
		}

		flowing := 0
		queued := 0
		type busyInfo struct {
			name, itemID, step string
			since              time.Time
		}
		var busy []busyInfo
		for _, item := range allItems {
			switch item.Status {
			case "in_progress":
				flowing++
				if item.Assignee != "" {
					busy = append(busy, busyInfo{item.Assignee, item.ID, item.CurrentCataracta, item.UpdatedAt})
				}
			case "open":
				queued++
			}
		}

		fmt.Println("Cistern")
		fmt.Printf("Cataractae  : %d open (%s)\n", len(allNames), strings.Join(allNames, ", "))
		total := flowing + queued
		fmt.Printf("Cistern  : %d droplets (%d flowing, %d queued)\n", total, flowing, queued)

		if len(busy) > 0 {
			fmt.Println()
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			for _, bw := range busy {
				elapsed := int(time.Since(bw.since).Minutes())
				fmt.Fprintf(tw, "%s\t%s\t[%s]\t%dm\n", bw.name, bw.itemID, bw.step, elapsed)
			}
			tw.Flush()
		}

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
	Short: "Validate a config and all referenced workflow files",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := resolveConfigPath()
		if len(args) > 0 {
			path = args[0]
		}

		cfg, err := aqueduct.ParseAqueductConfig(path)
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

			if _, err := aqueduct.ParseWorkflow(wfPath); err != nil {
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

// --- hidden "aqueduct" alias (deprecated) ---

var farmAliasCmd = &cobra.Command{
	Use:                "aqueduct",
	Hidden:             true,
	Short:              "Deprecated: use 'ct flow'",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "Cistern speaks water now. Use 'ct flow' instead of 'ct farm'.")
		return nil
	},
}

func init() {
	flowStartCmd.Flags().StringVar(&configPath, "config", "", "path to config (default: ./config.yaml)")

	flowConfigCmd.AddCommand(flowConfigValidateCmd)
	flowCmd.AddCommand(flowStartCmd, flowStatusCmd, flowConfigCmd)
	rootCmd.AddCommand(flowCmd)

	// Hidden "aqueduct" alias — prints deprecation message for any usage.
	rootCmd.AddCommand(farmAliasCmd)
}

func resolveConfigPath() string {
	if configPath != "" {
		return configPath
	}
	if env := os.Getenv("CT_CONFIG"); env != "" {
		return env
	}
	return "config.yaml"
}

// repoWorkerNames returns the configured worker names for a repo,
// falling back to worker-0, worker-1, etc.
func repoWorkerNames(repo aqueduct.RepoConfig) []string {
	if len(repo.Names) > 0 {
		return repo.Names
	}
	names := make([]string, repo.Cataractae)
	for i := range names {
		names[i] = fmt.Sprintf("worker-%d", i)
	}
	return names
}
