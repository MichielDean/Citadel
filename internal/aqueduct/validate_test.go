package aqueduct

import (
	"strings"
	"testing"
)

// minimalValidConfig returns the smallest valid AqueductConfig for use as a
// base in table-driven tests.
func minimalValidConfig() AqueductConfig {
	return AqueductConfig{
		Repos: []RepoConfig{
			{Name: "test-repo", Cataractae: 1, Prefix: "t"},
		},
	}
}

// --- ArchitectiConfig validation ---

func TestValidateAqueductConfig_Architecti_Nil_NoError(t *testing.T) {
	cfg := minimalValidConfig()
	cfg.Architecti = nil
	if err := ValidateAqueductConfig(&cfg); err != nil {
		t.Errorf("unexpected error with nil Architecti: %v", err)
	}
}

func TestValidateAqueductConfig_Architecti_Valid_NoError(t *testing.T) {
	cfg := minimalValidConfig()
	cfg.Architecti = &ArchitectiConfig{
		MaxFilesPerRun: 50,
	}
	if err := ValidateAqueductConfig(&cfg); err != nil {
		t.Errorf("unexpected error with valid Architecti config: %v", err)
	}
}

func TestValidateAqueductConfig_Architecti_ZeroMaxFiles_ReturnsError(t *testing.T) {
	cfg := minimalValidConfig()
	cfg.Architecti = &ArchitectiConfig{
		MaxFilesPerRun: 0,
	}
	err := ValidateAqueductConfig(&cfg)
	if err == nil {
		t.Fatal("expected error for zero max_files_per_run, got nil")
	}
	if !strings.Contains(err.Error(), "max_files_per_run") {
		t.Errorf("error %q does not mention max_files_per_run", err.Error())
	}
}

func TestValidateAqueductConfig_Architecti_NegativeMaxFiles_ReturnsError(t *testing.T) {
	cfg := minimalValidConfig()
	cfg.Architecti = &ArchitectiConfig{
		MaxFilesPerRun: -1,
	}
	err := ValidateAqueductConfig(&cfg)
	if err == nil {
		t.Fatal("expected error for negative max_files_per_run, got nil")
	}
	if !strings.Contains(err.Error(), "max_files_per_run") {
		t.Errorf("error %q does not mention max_files_per_run", err.Error())
	}
}
