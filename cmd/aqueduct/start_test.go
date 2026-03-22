package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MichielDean/cistern/internal/cistern"
)

// spyCisternClient records the arguments passed to Add so the test can assert
// that cisternAdder correctly maps the delivery convention (title, repo, ...) to
// the cistern.Client convention (repo, title, ...).
type spyCisternClient struct {
	gotRepo, gotTitle, gotDescription string
	gotPriority, gotComplexity        int
}

func (s *spyCisternClient) Add(repo, title, description string, priority, complexity int, deps ...string) (*cistern.Droplet, error) {
	s.gotRepo = repo
	s.gotTitle = title
	s.gotDescription = description
	s.gotPriority = priority
	s.gotComplexity = complexity
	return &cistern.Droplet{ID: "ct-test"}, nil
}

// TestResolveDeliveryDBPath_EnvSet verifies that CT_DB overrides the default path.
func TestResolveDeliveryDBPath_EnvSet(t *testing.T) {
	t.Setenv("CT_DB", "/custom/path/cistern.db")
	got := resolveDeliveryDBPath()
	if got != "/custom/path/cistern.db" {
		t.Errorf("got %q, want %q", got, "/custom/path/cistern.db")
	}
}

// TestResolveDeliveryDBPath_Default verifies that the default path is used when CT_DB is unset.
func TestResolveDeliveryDBPath_Default(t *testing.T) {
	prev, exists := os.LookupEnv("CT_DB")
	os.Unsetenv("CT_DB")
	t.Cleanup(func() {
		if exists {
			os.Setenv("CT_DB", prev)
		} else {
			os.Unsetenv("CT_DB")
		}
	})

	got := resolveDeliveryDBPath()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	want := filepath.Join(home, ".cistern", "cistern.db")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestCisternAdder_ParameterMapping verifies that cisternAdder.Add correctly
// swaps the delivery-layer argument order (title, repo) to match the
// cistern.Client convention (repo, title). A bug here would silently store
// every droplet with its title in the repo field and vice versa.
func TestCisternAdder_ParameterMapping(t *testing.T) {
	spy := &spyCisternClient{}
	a := &cisternAdder{c: spy}

	id, err := a.Add("my-title", "my-repo", "some description", 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "ct-test" {
		t.Errorf("id = %q, want %q", id, "ct-test")
	}
	if spy.gotRepo != "my-repo" {
		t.Errorf("cistern.Add received repo=%q, want %q", spy.gotRepo, "my-repo")
	}
	if spy.gotTitle != "my-title" {
		t.Errorf("cistern.Add received title=%q, want %q", spy.gotTitle, "my-title")
	}
	if spy.gotDescription != "some description" {
		t.Errorf("cistern.Add received description=%q, want %q", spy.gotDescription, "some description")
	}
}
