package aqueduct

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ParseWorkflow reads a YAML file and returns a validated aqueduct.
func ParseWorkflow(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading workflow file: %w", err)
	}
	return ParseWorkflowBytes(data)
}

// ParseWorkflowBytes parses YAML bytes into a validated aqueduct.
func ParseWorkflowBytes(data []byte) (*Workflow, error) {
	var w Workflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("parsing workflow YAML: %w", err)
	}
	if err := Validate(&w); err != nil {
		return nil, err
	}
	return &w, nil
}

// ParseAqueductConfig reads a YAML file and returns a AqueductConfig.
func ParseAqueductConfig(path string) (*AqueductConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading farm config: %w", err)
	}
	var cfg AqueductConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing farm config YAML: %w", err)
	}
	if err := ValidateAqueductConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GenerateCataractaeFiles writes CLAUDE.md files for each role defined in the aqueduct.
// Files are written to <cataractaeDir>/<roleKey>/CLAUDE.md.
func GenerateCataractaeFiles(w *Workflow, cataractaeDir string) ([]string, error) {
	if len(w.CataractaeDefinitions) == 0 {
		return nil, nil
	}

	var written []string
	for key, role := range w.CataractaeDefinitions {
		dir := filepath.Join(cataractaeDir, key)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return written, fmt.Errorf("create role dir %s: %w", dir, err)
		}

		content := fmt.Sprintf("# Role: %s\n\n%s\n\n%s\n", role.Name, role.Description, role.Instructions)
		path := filepath.Join(dir, "CLAUDE.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return written, fmt.Errorf("write %s: %w", path, err)
		}
		written = append(written, path)
	}
	return written, nil
}
