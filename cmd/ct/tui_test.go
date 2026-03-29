package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MichielDean/cistern/internal/cistern"
)

// ── Droplets tab cursor navigation ──────────────────────────────────────────

// TestTabApp_Droplets_CursorDown_MovesToNextItem verifies that pressing 'j'
// moves the cursor from the first to the second item.
//
// Given: a model with two cistern items and cursor=0
// When:  'j' is pressed
// Then:  cursor becomes 1
func TestTabApp_Droplets_CursorDown_MovesToNextItem(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{
		CisternItems: []*cistern.Droplet{
			{ID: "ci-aaa", Title: "First item", Status: "open"},
			{ID: "ci-bbb", Title: "Second item", Status: "open"},
		},
	}
	m.cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(tabAppModel)

	if um.cursor != 1 {
		t.Errorf("cursor = %d, want 1", um.cursor)
	}
}

// TestTabApp_Droplets_CursorDown_AtLastItem_Stays verifies that pressing 'j'
// at the last item does not advance the cursor past the end.
//
// Given: a model with two items and cursor=1 (last item)
// When:  'j' is pressed
// Then:  cursor stays at 1
func TestTabApp_Droplets_CursorDown_AtLastItem_Stays(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{
		CisternItems: []*cistern.Droplet{
			{ID: "ci-aaa", Status: "open"},
			{ID: "ci-bbb", Status: "open"},
		},
	}
	m.cursor = 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(tabAppModel)

	if um.cursor != 1 {
		t.Errorf("cursor = %d, want 1 (should not advance past last item)", um.cursor)
	}
}

// TestTabApp_Droplets_CursorUp_MovesToPreviousItem verifies that pressing 'k'
// moves the cursor from the second to the first item.
//
// Given: a model with two items and cursor=1
// When:  'k' is pressed
// Then:  cursor becomes 0
func TestTabApp_Droplets_CursorUp_MovesToPreviousItem(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{
		CisternItems: []*cistern.Droplet{
			{ID: "ci-aaa", Status: "open"},
			{ID: "ci-bbb", Status: "open"},
		},
	}
	m.cursor = 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	um := updated.(tabAppModel)

	if um.cursor != 0 {
		t.Errorf("cursor = %d, want 0", um.cursor)
	}
}

// TestTabApp_Droplets_CursorUp_AtZero_Stays verifies that pressing 'k' at
// the first item does not move the cursor to a negative index.
//
// Given: a model with cursor=0
// When:  'k' is pressed
// Then:  cursor stays at 0
func TestTabApp_Droplets_CursorUp_AtZero_Stays(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{
		CisternItems: []*cistern.Droplet{
			{ID: "ci-aaa", Status: "open"},
		},
	}
	m.cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	um := updated.(tabAppModel)

	if um.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (should not go below 0)", um.cursor)
	}
}

// ── Droplets → Detail navigation ────────────────────────────────────────────

// TestTabApp_Droplets_Enter_SwitchesToDetailTab verifies that pressing Enter
// on a selected item switches to the Detail tab, sets selectedID, and returns
// a fetch command for the detail notes.
//
// Given: a model with one cistern item and cursor=0
// When:  enter is pressed
// Then:  tab becomes tabDetail, selectedID is set, a cmd is returned
func TestTabApp_Droplets_Enter_SwitchesToDetailTab(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{
		CisternItems: []*cistern.Droplet{
			{ID: "ci-aaa", Title: "Some task", Status: "open"},
		},
	}
	m.cursor = 0
	m.tab = tabDroplets

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(tabAppModel)

	if um.tab != tabDetail {
		t.Errorf("tab = %d, want tabDetail (%d)", um.tab, tabDetail)
	}
	if um.selectedID != "ci-aaa" {
		t.Errorf("selectedID = %q, want %q", um.selectedID, "ci-aaa")
	}
	if cmd == nil {
		t.Error("expected a fetch cmd, got nil")
	}
}

// TestTabApp_Droplets_Enter_EmptyList_NoOp verifies that pressing Enter with
// an empty item list is a no-op: the tab stays on Droplets and no cmd is issued.
//
// Given: a model with no cistern items
// When:  enter is pressed
// Then:  tab remains tabDroplets and cmd is nil
func TestTabApp_Droplets_Enter_EmptyList_NoOp(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{CisternItems: nil}
	m.tab = tabDroplets

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(tabAppModel)

	if um.tab != tabDroplets {
		t.Errorf("tab = %d, want tabDroplets (%d) for empty list", um.tab, tabDroplets)
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty list, got non-nil")
	}
}

// TestTabApp_Droplets_D_Key_AlsoOpensDetail verifies that pressing 'd' is an
// alias for enter and also opens the detail tab.
//
// Given: a model with one item
// When:  'd' is pressed
// Then:  tab becomes tabDetail
func TestTabApp_Droplets_D_Key_AlsoOpensDetail(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{
		CisternItems: []*cistern.Droplet{
			{ID: "ci-aaa", Status: "open"},
		},
	}
	m.cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	um := updated.(tabAppModel)

	if um.tab != tabDetail {
		t.Errorf("tab = %d, want tabDetail (%d) on 'd' key", um.tab, tabDetail)
	}
}

// ── Detail tab navigation ────────────────────────────────────────────────────

// TestTabApp_Detail_Escape_ReturnsToDropletsAndClearsSelection verifies that
// pressing Escape in the Detail tab returns to the Droplets tab and clears
// the selectedID field.
//
// Given: a model in the Detail tab with a selectedID
// When:  esc is pressed
// Then:  tab becomes tabDroplets and selectedID is empty
func TestTabApp_Detail_Escape_ReturnsToDropletsAndClearsSelection(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.selectedID = "ci-aaa"
	m.detailDroplet = &cistern.Droplet{ID: "ci-aaa", Title: "Some task"}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(tabAppModel)

	if um.tab != tabDroplets {
		t.Errorf("tab = %d, want tabDroplets (%d) after esc", um.tab, tabDroplets)
	}
	if um.selectedID != "" {
		t.Errorf("selectedID = %q, want empty after esc", um.selectedID)
	}
}

// TestTabApp_Detail_ScrollDown_IncreasesScrollY verifies that pressing 'j'
// in the Detail tab increments the scroll offset.
//
// Given: a model in the Detail tab with detailScrollY=0
// When:  'j' is pressed
// Then:  detailScrollY becomes 1
func TestTabApp_Detail_ScrollDown_IncreasesScrollY(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{ID: "ci-aaa"}
	m.detailScrollY = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(tabAppModel)

	if um.detailScrollY != 1 {
		t.Errorf("detailScrollY = %d, want 1", um.detailScrollY)
	}
}

// TestTabApp_Detail_ScrollUp_AtZero_StaysAtZero verifies that pressing 'k'
// when already at the top does not produce a negative scroll offset.
//
// Given: a model in the Detail tab with detailScrollY=0
// When:  'k' is pressed
// Then:  detailScrollY remains 0
func TestTabApp_Detail_ScrollUp_AtZero_StaysAtZero(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{ID: "ci-aaa"}
	m.detailScrollY = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	um := updated.(tabAppModel)

	if um.detailScrollY != 0 {
		t.Errorf("detailScrollY = %d, want 0 (should not go below 0)", um.detailScrollY)
	}
}

// TestTabApp_Detail_HomeKey_ResetsScrollY verifies that pressing 'g' jumps
// the detail panel back to the top.
//
// Given: a model in the Detail tab with detailScrollY=10
// When:  'g' is pressed
// Then:  detailScrollY becomes 0
func TestTabApp_Detail_HomeKey_ResetsScrollY(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{ID: "ci-aaa"}
	m.detailScrollY = 10

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	um := updated.(tabAppModel)

	if um.detailScrollY != 0 {
		t.Errorf("detailScrollY = %d, want 0 after 'g'", um.detailScrollY)
	}
}

// ── Detail data message ──────────────────────────────────────────────────────

// TestTabApp_Detail_NotesFetched_StoredInChronologicalOrder verifies that
// when tuiDetailDataMsg arrives with notes newest-first (as returned by the DB),
// the model stores them oldest-first so the timeline reads chronologically.
//
// Given: a model in Detail tab with selectedID="ci-aaa"
// When:  tuiDetailDataMsg arrives with 2 notes, newest first
// Then:  detailNotes[0] is the older note, detailNotes[1] is the newer note
func TestTabApp_Detail_NotesFetched_StoredInChronologicalOrder(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.selectedID = "ci-aaa"
	m.detailDroplet = &cistern.Droplet{ID: "ci-aaa"}

	older := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	// DB returns newest first.
	notes := []cistern.CataractaeNote{
		{ID: 2, DropletID: "ci-aaa", CataractaeName: "reviewer", Content: "LGTM", CreatedAt: newer},
		{ID: 1, DropletID: "ci-aaa", CataractaeName: "implementer", Content: "Done", CreatedAt: older},
	}

	updated, _ := m.Update(tuiDetailDataMsg{dropletID: "ci-aaa", notes: notes})
	um := updated.(tabAppModel)

	if len(um.detailNotes) != 2 {
		t.Fatalf("detailNotes length = %d, want 2", len(um.detailNotes))
	}
	if um.detailNotes[0].CataractaeName != "implementer" {
		t.Errorf("detailNotes[0].CataractaeName = %q, want %q (oldest first)", um.detailNotes[0].CataractaeName, "implementer")
	}
	if um.detailNotes[1].CataractaeName != "reviewer" {
		t.Errorf("detailNotes[1].CataractaeName = %q, want %q (newest last)", um.detailNotes[1].CataractaeName, "reviewer")
	}
}

// TestTabApp_Detail_NotesFetched_StaleDropletID_Ignored verifies that notes
// fetched for a different droplet ID are discarded.
//
// Given: a model in Detail tab with selectedID="ci-aaa"
// When:  tuiDetailDataMsg arrives for "ci-bbb"
// Then:  detailNotes remains nil (stale response discarded)
func TestTabApp_Detail_NotesFetched_StaleDropletID_Ignored(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.selectedID = "ci-aaa"
	m.detailDroplet = &cistern.Droplet{ID: "ci-aaa"}
	m.detailNotes = nil

	notes := []cistern.CataractaeNote{
		{ID: 1, DropletID: "ci-bbb", CataractaeName: "implementer", Content: "Done"},
	}

	updated, _ := m.Update(tuiDetailDataMsg{dropletID: "ci-bbb", notes: notes})
	um := updated.(tabAppModel)

	if um.detailNotes != nil {
		t.Errorf("detailNotes should be nil for stale droplet ID, got %v", um.detailNotes)
	}
}

// ── View rendering ───────────────────────────────────────────────────────────

// TestTabApp_Detail_View_ShowsTitleAndID verifies that the Detail panel
// renders the droplet ID and title in its header.
//
// Given: a model in Detail tab with a droplet loaded
// When:  View() is called
// Then:  the output contains the droplet ID and title
func TestTabApp_Detail_View_ShowsTitleAndID(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{
		ID:                "ci-abc12",
		Title:             "Add retry logic to export pipeline",
		Repo:              "myrepo",
		Status:            "in_progress",
		CurrentCataractae: "implement",
	}
	m.width = 120
	m.height = 30

	view := m.View()

	if !strings.Contains(view, "ci-abc12") {
		t.Errorf("view should contain droplet ID 'ci-abc12', got:\n%s", view)
	}
	if !strings.Contains(view, "Add retry logic to export pipeline") {
		t.Errorf("view should contain title, got:\n%s", view)
	}
}

// TestTabApp_Detail_View_ShowsPipelineSteps verifies that the Detail panel
// renders all pipeline steps in the step position indicator.
//
// Given: a model in Detail tab with detailSteps set
// When:  View() is called
// Then:  each step name appears in the output
func TestTabApp_Detail_View_ShowsPipelineSteps(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{
		ID:                "ci-abc12",
		Title:             "Test task",
		Status:            "in_progress",
		CurrentCataractae: "review",
	}
	m.detailSteps = []string{"implement", "review", "test"}
	m.width = 120
	m.height = 30

	view := m.View()

	for _, step := range []string{"implement", "review", "test"} {
		if !strings.Contains(view, step) {
			t.Errorf("view should contain pipeline step %q, got:\n%s", step, view)
		}
	}
}

// TestTabApp_Detail_View_ShowsNotesWithAuthors verifies that the Detail panel
// renders note content and author names from the detailNotes timeline.
//
// Given: a model in Detail tab with two notes loaded
// When:  View() is called
// Then:  note content and author names are present in the output
func TestTabApp_Detail_View_ShowsNotesWithAuthors(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{
		ID:     "ci-abc12",
		Title:  "Test task",
		Status: "in_progress",
	}
	m.detailNotes = []cistern.CataractaeNote{
		{
			ID:             1,
			CataractaeName: "implementer",
			Content:        "Initial implementation done",
			CreatedAt:      time.Now().Add(-2 * time.Hour),
		},
		{
			ID:             2,
			CataractaeName: "reviewer",
			Content:        "LGTM with minor comments",
			CreatedAt:      time.Now().Add(-1 * time.Hour),
		},
	}
	m.width = 120
	m.height = 30

	view := m.View()

	for _, want := range []string{"implementer", "Initial implementation done", "reviewer", "LGTM with minor comments"} {
		if !strings.Contains(view, want) {
			t.Errorf("view should contain %q, got:\n%s", want, view)
		}
	}
}

// TestTabApp_Detail_View_ShowsEscHint verifies that the Detail panel's footer
// includes an "esc" keybinding hint to navigate back.
//
// Given: a model in Detail tab with a droplet loaded
// When:  View() is called
// Then:  "esc" appears in the output
func TestTabApp_Detail_View_ShowsEscHint(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{ID: "ci-abc12", Title: "Test"}
	m.width = 120
	m.height = 30

	view := m.View()

	if !strings.Contains(view, "esc") {
		t.Errorf("view should contain 'esc' keybinding hint, got:\n%s", view)
	}
}

// TestTabApp_Detail_View_ShowsRepoAndStatus verifies that the Detail panel
// header row contains the repo name and current status.
//
// Given: a model in Detail tab with repo="myrepo" and status="in_progress"
// When:  View() is called
// Then:  "myrepo" and "in_progress" are present in the output
func TestTabApp_Detail_View_ShowsRepoAndStatus(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{
		ID:     "ci-abc12",
		Title:  "Test task",
		Repo:   "myrepo",
		Status: "in_progress",
	}
	m.width = 120
	m.height = 30

	view := m.View()

	if !strings.Contains(view, "myrepo") {
		t.Errorf("view should contain repo 'myrepo', got:\n%s", view)
	}
	if !strings.Contains(view, "in_progress") {
		t.Errorf("view should contain status 'in_progress', got:\n%s", view)
	}
}

// TestTabApp_Droplets_View_ShowsItemIDs verifies that the Droplets tab lists
// all cistern item IDs.
//
// Given: a model in Droplets tab with two items
// When:  View() is called
// Then:  both item IDs appear in the output
func TestTabApp_Droplets_View_ShowsItemIDs(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{
		CisternItems: []*cistern.Droplet{
			{ID: "ci-abc12", Title: "First droplet", Status: "open"},
			{ID: "ci-def34", Title: "Second droplet", Status: "in_progress", CurrentCataractae: "implement"},
		},
	}
	m.tab = tabDroplets
	m.cursor = 0
	m.width = 120
	m.height = 30

	view := m.View()

	for _, id := range []string{"ci-abc12", "ci-def34"} {
		if !strings.Contains(view, id) {
			t.Errorf("view should contain item ID %q, got:\n%s", id, view)
		}
	}
}

// ── Window resize ────────────────────────────────────────────────────────────

// TestTabApp_WindowResize_UpdatesDimensions verifies that a WindowSizeMsg
// updates both width and height on the model.
//
// Given: a model with default dimensions
// When:  a WindowSizeMsg{Width:140, Height:40} arrives
// Then:  m.width=140 and m.height=40
func TestTabApp_WindowResize_UpdatesDimensions(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	um := updated.(tabAppModel)

	if um.width != 140 {
		t.Errorf("width = %d, want 140", um.width)
	}
	if um.height != 40 {
		t.Errorf("height = %d, want 40", um.height)
	}
}

// TestTabApp_Detail_WindowResize_UpdatesDimensions verifies that window resize
// works correctly when the Detail tab is active.
//
// Given: a model in Detail tab
// When:  a WindowSizeMsg arrives
// Then:  dimensions are updated
func TestTabApp_Detail_WindowResize_UpdatesDimensions(t *testing.T) {
	m := newTabAppModel("", "")
	m.data = &DashboardData{}
	m.tab = tabDetail
	m.detailDroplet = &cistern.Droplet{ID: "ci-aaa"}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 160, Height: 50})
	um := updated.(tabAppModel)

	if um.width != 160 {
		t.Errorf("width = %d, want 160", um.width)
	}
	if um.height != 50 {
		t.Errorf("height = %d, want 50", um.height)
	}
}
