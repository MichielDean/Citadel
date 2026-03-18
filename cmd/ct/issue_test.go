package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MichielDean/cistern/internal/cistern"
)

func execCmd(t *testing.T, args ...string) error {
	t.Helper()
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func TestDropletIssueAdd_CreatesIssue(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("CT_DB", db)
	t.Setenv("CT_NO_ASCII_LOGO", "1")
	t.Setenv("CT_CATARACTA_NAME", "reviewer")

	c, err := cistern.New(db, "ct")
	if err != nil {
		t.Fatal(err)
	}
	item, _ := c.Add("myrepo", "Task", "", 1, 3)
	c.Close()

	if err := execCmd(t, "droplet", "issue", "add", item.ID, "missing error handling"); err != nil {
		t.Fatalf("issue add failed: %v", err)
	}

	c2, _ := cistern.New(db, "ct")
	defer c2.Close()
	issues, _ := c2.ListIssues(item.ID, false)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Description != "missing error handling" {
		t.Errorf("description = %q", issues[0].Description)
	}
	if issues[0].Status != "open" {
		t.Errorf("status = %q, want open", issues[0].Status)
	}
	if issues[0].FlaggedBy != "reviewer" {
		t.Errorf("flagged_by = %q, want reviewer", issues[0].FlaggedBy)
	}
}

func TestDropletIssueResolve_UpdatesStatus(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("CT_DB", db)
	t.Setenv("CT_NO_ASCII_LOGO", "1")
	t.Setenv("CT_CATARACTA_NAME", "reviewer")

	c, _ := cistern.New(db, "ct")
	item, _ := c.Add("myrepo", "Task", "", 1, 3)
	iss, _ := c.AddIssue(item.ID, "reviewer", "some issue")
	c.Close()

	if err := execCmd(t, "droplet", "issue", "resolve", iss.ID, "--evidence", "grep output"); err != nil {
		t.Fatalf("issue resolve failed: %v", err)
	}

	c2, _ := cistern.New(db, "ct")
	defer c2.Close()
	issues, _ := c2.ListIssues(item.ID, false)
	if issues[0].Status != "resolved" {
		t.Errorf("status = %q, want resolved", issues[0].Status)
	}
}

func TestDropletIssueResolve_ImplementerForbidden(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("CT_DB", db)
	t.Setenv("CT_NO_ASCII_LOGO", "1")
	t.Setenv("CT_CATARACTA_NAME", "implementer")

	c, _ := cistern.New(db, "ct")
	item, _ := c.Add("myrepo", "Task", "", 1, 3)
	iss, _ := c.AddIssue(item.ID, "reviewer", "some issue")
	c.Close()

	err := execCmd(t, "droplet", "issue", "resolve", iss.ID, "--evidence", "trust me")
	if err == nil {
		t.Error("expected error: implementer should be forbidden from resolving issues")
	}
	if !strings.Contains(err.Error(), "only reviewer") {
		t.Errorf("unexpected error message: %v", err)
	}

	// Verify DB state unchanged.
	c2, _ := cistern.New(db, "ct")
	defer c2.Close()
	issues, _ := c2.ListIssues(item.ID, false)
	if issues[0].Status != "open" {
		t.Errorf("status should remain open, got %q", issues[0].Status)
	}
}

func TestDropletIssueResolve_ImplementShortName(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("CT_DB", db)
	t.Setenv("CT_NO_ASCII_LOGO", "1")
	t.Setenv("CT_CATARACTA_NAME", "implement")

	c, _ := cistern.New(db, "ct")
	item, _ := c.Add("myrepo", "Task", "", 1, 3)
	iss, _ := c.AddIssue(item.ID, "reviewer", "some issue")
	c.Close()

	err := execCmd(t, "droplet", "issue", "resolve", iss.ID, "--evidence", "proof")
	if err == nil {
		t.Error("expected error for CT_CATARACTA_NAME=implement")
	}
}

func TestDropletIssueReject_UpdatesStatus(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("CT_DB", db)
	t.Setenv("CT_NO_ASCII_LOGO", "1")

	c, _ := cistern.New(db, "ct")
	item, _ := c.Add("myrepo", "Task", "", 1, 3)
	iss, _ := c.AddIssue(item.ID, "reviewer", "unfixed issue")
	c.Close()

	if err := execCmd(t, "droplet", "issue", "reject", iss.ID, "--evidence", "still broken"); err != nil {
		t.Fatalf("issue reject failed: %v", err)
	}

	c2, _ := cistern.New(db, "ct")
	defer c2.Close()
	issues, _ := c2.ListIssues(item.ID, false)
	if issues[0].Status != "unresolved" {
		t.Errorf("status = %q, want unresolved", issues[0].Status)
	}
	if issues[0].Evidence != "still broken" {
		t.Errorf("evidence = %q", issues[0].Evidence)
	}
}

func TestDropletPass_BlockedByOpenIssues(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("CT_DB", db)
	t.Setenv("CT_NO_ASCII_LOGO", "1")

	c, _ := cistern.New(db, "ct")
	item, _ := c.Add("myrepo", "Task", "", 1, 3)
	c.AddIssue(item.ID, "reviewer", "open issue blocking pass")
	c.Close()

	err := execCmd(t, "droplet", "pass", item.ID)
	if err == nil {
		t.Error("expected error: pass should be blocked by open issues")
	}
	if !strings.Contains(err.Error(), "open issue") {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify outcome was NOT set.
	c2, _ := cistern.New(db, "ct")
	defer c2.Close()
	d, _ := c2.Get(item.ID)
	if d.Outcome == "pass" {
		t.Error("outcome should not be set to pass when open issues exist")
	}
}

func TestDropletPass_AllowedWhenIssuesResolved(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("CT_DB", db)
	t.Setenv("CT_NO_ASCII_LOGO", "1")
	t.Setenv("CT_CATARACTA_NAME", "reviewer")

	c, _ := cistern.New(db, "ct")
	item, _ := c.Add("myrepo", "Task", "", 1, 3)
	iss, _ := c.AddIssue(item.ID, "reviewer", "a finding")
	c.ResolveIssue(iss.ID, "fixed")
	c.Close()

	// Temporarily clear CT_CATARACTA_NAME so pass doesn't get confused.
	os.Unsetenv("CT_CATARACTA_NAME")
	defer os.Setenv("CT_CATARACTA_NAME", "reviewer")

	if err := execCmd(t, "droplet", "pass", item.ID); err != nil {
		t.Fatalf("pass should succeed when all issues resolved: %v", err)
	}

	c2, _ := cistern.New(db, "ct")
	defer c2.Close()
	d, _ := c2.Get(item.ID)
	if d.Outcome != "pass" {
		t.Errorf("outcome = %q, want pass", d.Outcome)
	}
}

func TestDropletPass_NoIssues(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("CT_DB", db)
	t.Setenv("CT_NO_ASCII_LOGO", "1")

	c, _ := cistern.New(db, "ct")
	item, _ := c.Add("myrepo", "Task", "", 1, 3)
	c.Close()

	if err := execCmd(t, "droplet", "pass", item.ID); err != nil {
		t.Fatalf("pass with no issues should succeed: %v", err)
	}

	c2, _ := cistern.New(db, "ct")
	defer c2.Close()
	d, _ := c2.Get(item.ID)
	if d.Outcome != "pass" {
		t.Errorf("outcome = %q, want pass", d.Outcome)
	}
}
