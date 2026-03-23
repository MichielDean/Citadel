package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/MichielDean/cistern/internal/aqueduct"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show repos, cataractae operators, and global cataractae count",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := aqueduct.ParseAqueductConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfgDir := filepath.Dir(cfgPath)
	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "REPO\tWORKFLOW\tSLUICES\tSTATUS")

	totalWorkers := 0
	for _, repo := range cfg.Repos {
		wfName := "-"
		status := "ok"

		wfPath := repo.WorkflowPath
		if wfPath != "" {
			if !filepath.IsAbs(wfPath) {
				wfPath = filepath.Join(cfgDir, wfPath)
			}
			w, err := aqueduct.ParseWorkflow(wfPath)
			if err != nil {
				wfName = repo.WorkflowPath
				status = "error: " + err.Error()
			} else {
				wfName = w.Name
			}
		}

		totalWorkers += repo.Cataractae
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", repo.Name, wfName, repo.Cataractae, status)
	}
	tw.Flush()

	fmt.Printf("\ntotal cataractae: %d\n", totalWorkers)
	return nil
}
