package main

import (
	"path/filepath"
	"testing"
)

// TestRunStatus_ValidConfig verifies that runStatus returns no error when given
// a well-formed config with resolvable workflow files.
func TestRunStatus_ValidConfig(t *testing.T) {
	saved := cfgPath
	defer func() { cfgPath = saved }()

	cfgPath = filepath.Join("testdata", "test_config.yaml")
	if err := runStatus(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunStatus_InvalidConfigPath verifies that runStatus returns an error when
// the config file does not exist.
func TestRunStatus_InvalidConfigPath(t *testing.T) {
	saved := cfgPath
	defer func() { cfgPath = saved }()

	cfgPath = "nonexistent_config.yaml"
	if err := runStatus(nil, nil); err == nil {
		t.Fatal("expected error for nonexistent config path, got nil")
	}
}
