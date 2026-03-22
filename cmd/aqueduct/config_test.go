package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRunConfigValidate_Valid verifies that a well-formed config and workflow pass validation.
func TestRunConfigValidate_Valid(t *testing.T) {
	path := filepath.Join("testdata", "test_config.yaml")
	if err := runConfigValidate(nil, []string{path}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunConfigValidate_FileNotFound verifies that a missing config file returns an error.
func TestRunConfigValidate_FileNotFound(t *testing.T) {
	if err := runConfigValidate(nil, []string{"nonexistent_config.yaml"}); err == nil {
		t.Fatal("expected error for missing config file, got nil")
	}
}

// TestRunConfigValidate_InvalidYAML verifies that malformed YAML returns an error.
func TestRunConfigValidate_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(path, []byte(":\tinvalid: yaml: ["), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runConfigValidate(nil, []string{path}); err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

// TestRunConfigValidate_MissingWorkflowPath verifies that a repo entry without
// workflow_path is rejected by runConfigValidate's per-repo checks.
func TestRunConfigValidate_MissingWorkflowPath(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	content := `
repos:
  - name: myrepo
    cataractae: 1
    prefix: mr
max_cataractae: 1
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runConfigValidate(nil, []string{path}); err == nil {
		t.Fatal("expected error for missing workflow_path, got nil")
	}
}

// TestRunConfigValidate_WorkflowFileNotFound verifies that a workflow_path
// pointing to a non-existent file is caught and reported.
func TestRunConfigValidate_WorkflowFileNotFound(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	content := `
repos:
  - name: myrepo
    workflow_path: does_not_exist.yaml
    cataractae: 1
    prefix: mr
max_cataractae: 1
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runConfigValidate(nil, []string{path}); err == nil {
		t.Fatal("expected error for missing workflow file, got nil")
	}
}
