// Package bd provides a Go client for the bd CLI issue tracker.
//
// All methods exec the bd binary and parse JSON output.
// No direct Dolt access. Each repo config gets its own Client instance.
package bd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Bead represents an issue returned by bd.
type Bead struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Priority    int               `json:"priority"`
	Type        string            `json:"issue_type"`
	Assignee    string            `json:"assignee,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Notes       string            `json:"notes,omitempty"`
	Design      string            `json:"design,omitempty"`
	Dependencies []BeadDependency `json:"dependencies,omitempty"`
}

// BeadDependency is a dependency edge on a bead.
type BeadDependency struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Type   string `json:"dependency_type"`
}

// StepNote is a note attached by a workflow step.
type StepNote struct {
	ID        int       `json:"id"`
	IssueID   string    `json:"issue_id"`
	Author    string    `json:"author"`
	FromStep  string    `json:"-"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// Client wraps the bd CLI for a specific database path.
type Client struct {
	// BdPath is the path to the bd binary. Defaults to "bd".
	BdPath string

	// DBPath is passed as --db to scope the client to a specific repo's beads.
	// If empty, bd auto-discovers the database.
	DBPath string
}

// NewClient creates a Client. If bdPath is empty, "bd" is used.
func NewClient(bdPath, dbPath string) *Client {
	if bdPath == "" {
		bdPath = "bd"
	}
	return &Client{BdPath: bdPath, DBPath: dbPath}
}

func (c *Client) baseArgs() []string {
	var args []string
	if c.DBPath != "" {
		args = append(args, "--db", c.DBPath)
	}
	args = append(args, "--json")
	return args
}

func (c *Client) run(args ...string) ([]byte, error) {
	cmd := exec.Command(c.BdPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bd %s: %w: %s", strings.Join(args, " "), err, out)
	}
	return out, nil
}

// GetReady returns the next bead with no active blockers, ready for work.
// rig scopes the query to a specific rig's database.
func (c *Client) GetReady(rig string) (*Bead, error) {
	args := c.baseArgs()
	args = append(args, "ready", "--limit", "1")
	if rig != "" {
		args = append(args, "--rig", rig)
	}

	out, err := c.run(args...)
	if err != nil {
		return nil, fmt.Errorf("GetReady: %w", err)
	}

	var beads []Bead
	if err := json.Unmarshal(out, &beads); err != nil {
		return nil, fmt.Errorf("GetReady: parse: %w", err)
	}
	if len(beads) == 0 {
		return nil, nil
	}
	return &beads[0], nil
}

// UpdateStep sets the current workflow step on a bead via metadata.
func (c *Client) UpdateStep(id, step string) error {
	args := c.baseArgs()
	args = append(args, "update", id, "--set-metadata", "step="+step)

	if _, err := c.run(args...); err != nil {
		return fmt.Errorf("UpdateStep: %w", err)
	}
	return nil
}

// IncrementAttempts increments the attempt counter for a step and returns the new count.
// The count is stored in bead metadata as "attempts:<step>".
func (c *Client) IncrementAttempts(id, step string) (int, error) {
	current, err := c.getAttemptCount(id, step)
	if err != nil {
		return 0, fmt.Errorf("IncrementAttempts: %w", err)
	}

	next := current + 1
	key := "attempts:" + step
	args := c.baseArgs()
	args = append(args, "update", id, "--set-metadata", key+"="+strconv.Itoa(next))

	if _, err := c.run(args...); err != nil {
		return 0, fmt.Errorf("IncrementAttempts: %w", err)
	}
	return next, nil
}

func (c *Client) getAttemptCount(id, step string) (int, error) {
	bead, err := c.show(id)
	if err != nil {
		return 0, err
	}

	key := "attempts:" + step
	if bead.Metadata == nil {
		return 0, nil
	}
	v, ok := bead.Metadata[key]
	if !ok {
		return 0, nil
	}

	switch n := v.(type) {
	case float64:
		return int(n), nil
	case json.Number:
		i, err := n.Int64()
		return int(i), err
	case string:
		return strconv.Atoi(n)
	default:
		return 0, fmt.Errorf("unexpected attempt count type %T", v)
	}
}

func (c *Client) show(id string) (*Bead, error) {
	args := c.baseArgs()
	args = append(args, "show", id, "--long")

	out, err := c.run(args...)
	if err != nil {
		return nil, fmt.Errorf("show %s: %w", id, err)
	}

	var beads []Bead
	if err := json.Unmarshal(out, &beads); err != nil {
		return nil, fmt.Errorf("show %s: parse: %w", id, err)
	}
	if len(beads) == 0 {
		return nil, fmt.Errorf("show %s: not found", id)
	}
	return &beads[0], nil
}

// AttachNotes stores step output as a comment on the bead.
// The comment is prefixed with "[fromStep]" so GetNotes can parse the source step.
func (c *Client) AttachNotes(id, fromStep, notes string) error {
	text := "[" + fromStep + "] " + notes

	args := []string{}
	if c.DBPath != "" {
		args = append(args, "--db", c.DBPath)
	}
	args = append(args, "comments", "add", id, text)

	if _, err := c.run(args...); err != nil {
		return fmt.Errorf("AttachNotes: %w", err)
	}
	return nil
}

// GetNotes returns all step notes (comments) attached to a bead.
// Comments prefixed with "[step] " are parsed into StepNote with the FromStep field set.
func (c *Client) GetNotes(id string) ([]StepNote, error) {
	args := c.baseArgs()
	args = append(args, "comments", id)

	out, err := c.run(args...)
	if err != nil {
		return nil, fmt.Errorf("GetNotes: %w", err)
	}

	var notes []StepNote
	if err := json.Unmarshal(out, &notes); err != nil {
		return nil, fmt.Errorf("GetNotes: parse: %w", err)
	}

	for i := range notes {
		notes[i].FromStep, notes[i].Text = parseStepPrefix(notes[i].Text)
	}
	return notes, nil
}

// parseStepPrefix extracts a "[step] " prefix from text.
// Returns (step, remaining_text). If no prefix, step is empty.
func parseStepPrefix(text string) (string, string) {
	if !strings.HasPrefix(text, "[") {
		return "", text
	}
	end := strings.Index(text, "] ")
	if end <= 1 {
		return "", text
	}
	return text[1:end], text[end+2:]
}

// Escalate marks a bead as needing human attention.
// Sets status to "blocked", adds the "needs-human" label, and records the reason as a comment.
func (c *Client) Escalate(id, reason string) error {
	args := c.baseArgs()
	args = append(args, "update", id, "--status", "blocked", "--add-label", "needs-human")

	if _, err := c.run(args...); err != nil {
		return fmt.Errorf("Escalate: update: %w", err)
	}

	commentArgs := []string{}
	if c.DBPath != "" {
		commentArgs = append(commentArgs, "--db", c.DBPath)
	}
	commentArgs = append(commentArgs, "comments", "add", id, "ESCALATE: "+reason)

	if _, err := c.run(commentArgs...); err != nil {
		return fmt.Errorf("Escalate: comment: %w", err)
	}
	return nil
}
