package cataractae

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// EnsurePrimaryClone guarantees a full git clone exists at primaryDir.
// This is the shared object store for all worktrees in the same repo.
//
// On first call: clones the repo.
// On subsequent calls: fetches latest remote refs.
func EnsurePrimaryClone(primaryDir, repoURL string) error {
	gitDir := filepath.Join(primaryDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := os.RemoveAll(primaryDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale dir %s: %w", primaryDir, err)
		}
		t0 := time.Now()
		if err := cloneSandbox(primaryDir, repoURL); err != nil {
			slog.Default().Error("git clone failed",
				"dir", primaryDir, "duration", time.Since(t0).Round(time.Millisecond).String(), "error", err)
			return err
		}
		slog.Default().Info("git clone completed",
			"dir", primaryDir, "duration", time.Since(t0).Round(time.Millisecond).String())
		return nil
	}
	t0 := time.Now()
	if err := fetchSandbox(primaryDir); err != nil {
		slog.Default().Error("git fetch failed",
			"dir", primaryDir, "duration", time.Since(t0).Round(time.Millisecond).String(), "error", err)
		return err
	}
	slog.Default().Info("git fetch completed",
		"dir", primaryDir, "duration", time.Since(t0).Round(time.Millisecond).String())
	return nil
}

// EnsureWorktree ensures a git worktree exists at worktreeDir backed by primaryDir.
// It prunes stale worktree registrations first to prevent "already in use" errors.
// If worktreeDir exists as a legacy dedicated clone (not a registered worktree), it
// is removed so a fresh worktree can be created.
// On an empty/unborn repo (no commits on HEAD), it uses --orphan to create
// an orphan worktree because --detach would fail with "fatal: invalid reference: HEAD".
func EnsureWorktree(primaryDir, worktreeDir string) error {
	// Prune stale registrations so dead paths don't block re-registration.
	prune := exec.Command("git", "worktree", "prune", "--expire=0")
	prune.Dir = primaryDir
	if out, err := prune.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree prune in %s: %w: %s", primaryDir, err, out)
	}

	// Check if this worktree is already registered.
	list := exec.Command("git", "worktree", "list", "--porcelain")
	list.Dir = primaryDir
	out, err := list.Output()
	if err != nil {
		return fmt.Errorf("git worktree list in %s: %w", primaryDir, err)
	}

	absPath, err := filepath.Abs(worktreeDir)
	if err != nil {
		return fmt.Errorf("abs path %s: %w", worktreeDir, err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "worktree "+absPath {
			return nil // already registered
		}
	}

	// If the directory exists but is not a registered worktree (e.g., a legacy
	// dedicated clone), remove it so git worktree add can create a fresh worktree.
	if _, err := os.Stat(worktreeDir); err == nil {
		if err := os.RemoveAll(worktreeDir); err != nil {
			return fmt.Errorf("remove legacy clone at %s: %w", worktreeDir, err)
		}
	}

	// Check whether HEAD has commits. On an empty/unborn repo, --detach fails
	// with "fatal: invalid reference: HEAD". Use --orphan in that case.
	hasCommits, commitsErr := repoHasCommits(primaryDir, "HEAD")
	if commitsErr != nil {
		// If we can't check, assume commits exist (conservative fallback) and
		// let the normal path fail with a descriptive error if wrong.
		slog.Default().Warn("repoHasCommits check failed, assuming commits exist",
			"dir", primaryDir, "error", commitsErr)
		hasCommits = true
	}

	t0 := time.Now()
	if !hasCommits {
		// Empty repo: create an orphan worktree. --orphan requires -b, so use
		// a named branch for the worker. The _worker_ prefix makes it distinct
		// from feature branches.
		branchName := "_worker_" + filepath.Base(worktreeDir)
		add := exec.Command("git", "worktree", "add", "--orphan", "-b", branchName, worktreeDir)
		add.Dir = primaryDir
		if addOut, err := add.CombinedOutput(); err != nil {
			slog.Default().Error("git worktree add --orphan failed",
				"dir", worktreeDir, "duration", time.Since(t0).Round(time.Millisecond).String(), "error", err)
			return fmt.Errorf("git worktree add --orphan -b %s %s: %w: %s", branchName, worktreeDir, err, addOut)
		}
		slog.Default().Info("git worktree created (orphan)",
			"dir", worktreeDir, "duration", time.Since(t0).Round(time.Millisecond).String())
	} else {
		add := exec.Command("git", "worktree", "add", "--detach", worktreeDir)
		add.Dir = primaryDir
		if addOut, err := add.CombinedOutput(); err != nil {
			slog.Default().Error("git worktree add failed",
				"dir", worktreeDir, "duration", time.Since(t0).Round(time.Millisecond).String(), "error", err)
			return fmt.Errorf("git worktree add --detach %s: %w: %s", worktreeDir, err, addOut)
		}
		slog.Default().Info("git worktree created",
			"dir", worktreeDir, "duration", time.Since(t0).Round(time.Millisecond).String())
	}
	return nil
}

// repoHasCommits reports whether the given ref resolves to a commit in dir.
// It runs git rev-parse --verify <ref> in dir. If the ref exists, it returns
// true, nil. If the ref does not exist (unknown revision), it returns false, nil.
// For any other error (dir does not exist, permission denied), it returns false, err.
func repoHasCommits(dir, ref string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		// git rev-parse --verify exits with code 1 (or 128) and prints
		// "fatal: Needed a single revision" or "unknown revision" when
		// the ref does not exist. This is the expected "no commits" case.
		lower := strings.ToLower(string(out))
		if strings.Contains(lower, "unknown revision") || strings.Contains(lower, "needed a single revision") {
			return false, nil
		}
		return false, fmt.Errorf("cataractae: repoHasCommits %s in %s: %w: %s", ref, dir, err, out)
	}
	return true, nil
}

// cloneSandbox performs a fresh git clone into dir.
func cloneSandbox(dir, repoURL string) error {
	if repoURL == "" {
		return fmt.Errorf("repo URL is required for initial clone")
	}
	cmd := exec.Command("git", "clone", repoURL, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone %s: %w", repoURL, err)
	}
	return nil
}

// fetchSandbox fetches latest remote refs without touching the working tree.
func fetchSandbox(dir string) error {
	fetch := exec.Command("git", "fetch", "origin")
	fetch.Dir = dir
	if out, err := fetch.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch in %s: %w: %s", dir, err, out)
	}
	return nil
}

// currentHead returns the current HEAD commit hash in the given directory.
func currentHead(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD in %s: %w", dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}
