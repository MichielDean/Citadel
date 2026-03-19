package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit_CreatesDirectoryStructure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
t.Setenv("USERPROFILE", home)
	initForce = false

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cisternDir := filepath.Join(home, ".cistern")
	for _, dir := range []string{
		cisternDir,
		filepath.Join(cisternDir, "aqueduct"),
		filepath.Join(cisternDir, "cataractae"),
	} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory to exist: %s", dir)
		}
	}
}

func TestInit_WritesCisternYAML(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
t.Setenv("USERPROFILE", home)
	initForce = false

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configFile := filepath.Join(home, ".cistern", "cistern.yaml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("cistern.yaml not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("cistern.yaml is empty")
	}
	// Verify it matches the embedded template.
	if string(data) != string(defaultCisternConfig) {
		t.Error("cistern.yaml content does not match embedded template")
	}
}

func TestInit_CopiesWorkflowFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
t.Setenv("USERPROFILE", home)
	initForce = false

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	aqueductDir := filepath.Join(home, ".cistern", "aqueduct")
	aqueductYAML := filepath.Join(aqueductDir, "aqueduct.yaml")
	if _, err := os.Stat(aqueductYAML); os.IsNotExist(err) {
		t.Errorf("expected workflow file to exist: aqueduct.yaml")
	}

	// Verify aqueduct.yaml content matches embedded template.
	aqueductData, err := os.ReadFile(aqueductYAML)
	if err != nil {
		t.Fatalf("read aqueduct.yaml: %v", err)
	}
	if string(aqueductData) != string(defaultAqueductWorkflow) {
		t.Error("aqueduct.yaml content does not match embedded template")
	}
}

func TestInit_GeneratesRoles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
t.Setenv("USERPROFILE", home)
	initForce = false

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cataractaeDir := filepath.Join(home, ".cistern", "cataractae")
	entries, err := os.ReadDir(cataractaeDir)
	if err != nil {
		t.Fatalf("read roles dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no roles were generated")
	}

	// Each role should have a CLAUDE.md.
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		claudeMD := filepath.Join(cataractaeDir, entry.Name(), "CLAUDE.md")
		if _, err := os.Stat(claudeMD); os.IsNotExist(err) {
			t.Errorf("missing CLAUDE.md for role %q", entry.Name())
		}
	}
}

func TestInit_SkipsExistingFilesWithoutForce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
t.Setenv("USERPROFILE", home)
	initForce = false

	// First run to create files.
	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Overwrite cistern.yaml with sentinel content.
	configFile := filepath.Join(home, ".cistern", "cistern.yaml")
	sentinel := []byte("# sentinel — must not be overwritten")
	if err := os.WriteFile(configFile, sentinel, 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	// Second run without --force must not overwrite.
	initForce = false
	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("second run error: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(sentinel) {
		t.Error("cistern.yaml was overwritten without --force")
	}
}

func TestInit_ForceOverwritesExistingFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
t.Setenv("USERPROFILE", home)

	// First run to create files.
	initForce = false
	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Overwrite cistern.yaml with sentinel content.
	configFile := filepath.Join(home, ".cistern", "cistern.yaml")
	sentinel := []byte("# sentinel — must be overwritten with --force")
	if err := os.WriteFile(configFile, sentinel, 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	// Run with --force — must overwrite.
	initForce = true
	defer func() { initForce = false }()

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("force run error: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == string(sentinel) {
		t.Error("cistern.yaml was not overwritten with --force")
	}
	if string(data) != string(defaultCisternConfig) {
		t.Error("cistern.yaml does not match embedded template after --force")
	}
}

func TestInit_IdempotentWithoutForce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
t.Setenv("USERPROFILE", home)
	initForce = false

	// Run twice — both must succeed with no errors.
	for i := 0; i < 2; i++ {
		if err := initCmd.RunE(initCmd, nil); err != nil {
			t.Fatalf("run %d error: %v", i+1, err)
		}
	}
}
