package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/MichielDean/citadel/internal/queue"
	"github.com/spf13/cobra"
)

var cisternCmd = &cobra.Command{
	Use:     "cistern",
	Aliases: []string{"queue"},
	Short:   "Manage drops in the cistern",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if len(os.Args) > 1 && os.Args[1] == "queue" {
			fmt.Fprintln(os.Stderr, "The Citadel speaks water now. Use 'ct cistern' instead of 'ct queue'.")
		}
	},
}

// --- cistern add ---

var (
	addTitle       string
	addDescription string
	addPriority    int
	addRepo        string
)

var cisternAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new drop to the cistern",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addTitle == "" {
			return fmt.Errorf("--title is required")
		}
		if addRepo == "" {
			return fmt.Errorf("--repo is required")
		}
		c, err := queue.New(resolveDBPath(), inferPrefix(addRepo))
		if err != nil {
			return err
		}
		defer c.Close()

		item, err := c.Add(addRepo, addTitle, addDescription, addPriority)
		if err != nil {
			return err
		}
		fmt.Printf("Drop added to cistern. %s: %s\n", item.ID, item.Title)
		return nil
	},
}

// --- cistern list ---

var (
	listRepo   string
	listStatus string
	listOutput string
)

func waterStatus(s string) string {
	switch s {
	case "in_progress":
		return "flowing"
	case "open":
		return "queued"
	case "closed":
		return "free"
	case "escalated":
		return "poisoned"
	default:
		return s
	}
}

var cisternListCmd = &cobra.Command{
	Use:   "list",
	Short: "List drops in the cistern",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listOutput != "table" && listOutput != "json" {
			return fmt.Errorf("--output must be table or json")
		}
		c, err := queue.New(resolveDBPath(), "")
		if err != nil {
			return err
		}
		defer c.Close()

		items, err := c.List(listRepo, listStatus)
		if err != nil {
			return err
		}

		if listOutput == "json" {
			if items == nil {
				items = []*queue.WorkItem{}
			}
			out, err := json.MarshalIndent(items, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		}

		if len(items) == 0 {
			fmt.Println("Cistern dry.")
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "ID\tTITLE\tSTATUS\tVALVE")
		for _, item := range items {
			step := item.CurrentStep
			if step == "" {
				step = "\u2014"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
				item.ID, item.Title, waterStatus(item.Status), step)
		}
		return tw.Flush()
	},
}

// --- cistern show ---

var cisternShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show details of a drop",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := queue.New(resolveDBPath(), "")
		if err != nil {
			return err
		}
		defer c.Close()

		item, err := c.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("ID:          %s\n", item.ID)
		fmt.Printf("Title:       %s\n", item.Title)
		fmt.Printf("Repo:        %s\n", item.Repo)
		fmt.Printf("Status:      %s\n", waterStatus(item.Status))
		fmt.Printf("Priority:    %d\n", item.Priority)
		fmt.Printf("Channel:     %s\n", item.Assignee)
		fmt.Printf("Valve:       %s\n", item.CurrentStep)

		fmt.Printf("Created:     %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:     %s\n", item.UpdatedAt.Format("2006-01-02 15:04:05"))

		if item.Description != "" {
			fmt.Printf("\nDescription:\n%s\n", item.Description)
		}

		notes, err := c.GetNotes(item.ID)
		if err != nil {
			return err
		}
		if len(notes) > 0 {
			fmt.Printf("\nNotes:\n")
			for _, n := range notes {
				fmt.Printf("  [%s] %s\n", n.StepName, n.Content)
			}
		}

		return nil
	},
}

// --- cistern note ---

var cisternNoteCmd = &cobra.Command{
	Use:   "note <id> <content>",
	Short: "Add a note to a drop",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := queue.New(resolveDBPath(), "")
		if err != nil {
			return err
		}
		defer c.Close()

		if err := c.AddNote(args[0], "manual", args[1]); err != nil {
			return err
		}
		fmt.Printf("note added to drop %s\n", args[0])
		return nil
	},
}

// --- cistern close ---

var cisternCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a drop",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := queue.New(resolveDBPath(), "")
		if err != nil {
			return err
		}
		defer c.Close()

		if err := c.CloseItem(args[0]); err != nil {
			return err
		}
		fmt.Printf("Drop %s flows free.\n", args[0])
		return nil
	},
}

// --- cistern reopen ---

var cisternReopenCmd = &cobra.Command{
	Use:   "reopen <id>",
	Short: "Reopen a closed or poisoned drop",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := queue.New(resolveDBPath(), "")
		if err != nil {
			return err
		}
		defer c.Close()

		if err := c.UpdateStatus(args[0], "open"); err != nil {
			return err
		}
		fmt.Printf("Drop %s returned to cistern.\n", args[0])
		return nil
	},
}

// --- cistern escalate ---

var escalateReason string

var cisternEscalateCmd = &cobra.Command{
	Use:   "escalate <id>",
	Short: "Escalate a drop for human attention",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if escalateReason == "" {
			return fmt.Errorf("--reason is required")
		}
		c, err := queue.New(resolveDBPath(), "")
		if err != nil {
			return err
		}
		defer c.Close()

		if err := c.Escalate(args[0], escalateReason); err != nil {
			return err
		}
		fmt.Printf("Drop %s poisoned \u2014 escalated for human.\n", args[0])
		return nil
	},
}

// --- cistern purge ---

var (
	purgeOlderThan string
	purgeDryRun    bool
)

var cisternPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Delete old drops that have flowed free or been poisoned",
	RunE: func(cmd *cobra.Command, args []string) error {
		if purgeOlderThan == "" {
			return fmt.Errorf("--older-than is required")
		}
		d, err := parseDuration(purgeOlderThan)
		if err != nil {
			return fmt.Errorf("invalid --older-than value: %w", err)
		}
		c, err := queue.New(resolveDBPath(), "")
		if err != nil {
			return err
		}
		defer c.Close()

		n, err := c.Purge(d, purgeDryRun)
		if err != nil {
			return err
		}
		if purgeDryRun {
			fmt.Printf("dry-run: would purge %d drop(s)\n", n)
		} else {
			fmt.Printf("purged %d drop(s)\n", n)
		}
		return nil
	},
}

// --- cistern get (alias for show) ---

var cisternGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a drop (alias for show)",
	Args:  cobra.ExactArgs(1),
	RunE:  cisternShowCmd.RunE,
}

// parseDuration parses a duration string, supporting 'd' suffix for days
// in addition to standard Go duration units (e.g., "30d", "24h", "1h30m").
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, fmt.Errorf("invalid days value: %q", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func init() {
	cisternAddCmd.Flags().StringVar(&addTitle, "title", "", "drop title (required)")
	cisternAddCmd.Flags().StringVar(&addDescription, "description", "", "drop description")
	cisternAddCmd.Flags().IntVar(&addPriority, "priority", 2, "priority (1=highest)")
	cisternAddCmd.Flags().StringVar(&addRepo, "repo", "", "target repository (required)")

	cisternListCmd.Flags().StringVar(&listRepo, "repo", "", "filter by repo")
	cisternListCmd.Flags().StringVar(&listStatus, "status", "", "filter by status (open|in_progress|closed|escalated)")
	cisternListCmd.Flags().StringVar(&listOutput, "output", "table", "output format: table or json")

	cisternEscalateCmd.Flags().StringVar(&escalateReason, "reason", "", "escalation reason (required)")

	cisternPurgeCmd.Flags().StringVar(&purgeOlderThan, "older-than", "", "delete drops older than this duration (e.g. 30d, 24h) (required)")
	cisternPurgeCmd.Flags().BoolVar(&purgeDryRun, "dry-run", false, "show what would be deleted without deleting")

	cisternCmd.AddCommand(cisternAddCmd, cisternListCmd, cisternShowCmd, cisternGetCmd,
		cisternNoteCmd, cisternCloseCmd, cisternReopenCmd, cisternEscalateCmd, cisternPurgeCmd)
	rootCmd.AddCommand(cisternCmd)
}

// inferPrefix extracts a short prefix from a repo path for ID generation.
func inferPrefix(repo string) string {
	name := repo
	for i := len(repo) - 1; i >= 0; i-- {
		if repo[i] == '/' {
			name = repo[i+1:]
			break
		}
	}
	if len(name) == 0 {
		return "ct"
	}
	if len(name) <= 2 {
		return name
	}
	r := []byte{name[0], name[1]}
	for i := range r {
		if r[i] >= 'A' && r[i] <= 'Z' {
			r[i] += 32
		}
	}
	return string(r)
}
