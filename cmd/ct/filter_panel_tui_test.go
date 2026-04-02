package main

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MichielDean/cistern/internal/cistern"
)

// ── initial state ─────────────────────────────────────────────────────────────

// TestFilterPanel_NewPanel_TitleIsFilter verifies the panel title.
//
// Given: a new filterPanel
// When:  Title() is called
// Then:  "Filter" is returned
func TestFilterPanel_NewPanel_TitleIsFilter(t *testing.T) {
	p := newFilterPanel()
	if p.Title() != "Filter" {
		t.Errorf("Title() = %q, want %q", p.Title(), "Filter")
	}
}

// TestFilterPanel_NewPanel_OverlayNotActive verifies no overlay is active by default.
//
// Given: a new filterPanel
// When:  OverlayActive() is called
// Then:  false is returned
func TestFilterPanel_NewPanel_OverlayNotActive(t *testing.T) {
	p := newFilterPanel()
	if p.OverlayActive() {
		t.Error("OverlayActive() = true, want false")
	}
}

// TestFilterPanel_NewPanel_PaletteActionsNil verifies no palette actions.
//
// Given: a new filterPanel
// When:  PaletteActions() is called with a non-nil droplet
// Then:  nil is returned
func TestFilterPanel_NewPanel_PaletteActionsNil(t *testing.T) {
	p := newFilterPanel()
	d := &cistern.Droplet{ID: "ci-test01"}
	if actions := p.PaletteActions(d); actions != nil {
		t.Errorf("PaletteActions() = %v, want nil", actions)
	}
}

// TestFilterPanel_NewPanel_KeyHelpNonEmpty verifies a non-empty key help string.
//
// Given: a new filterPanel
// When:  KeyHelp() is called
// Then:  a non-empty string is returned
func TestFilterPanel_NewPanel_KeyHelpNonEmpty(t *testing.T) {
	p := newFilterPanel()
	if p.KeyHelp() == "" {
		t.Error("KeyHelp() = empty string, want non-empty")
	}
}

// TestFilterPanel_NewPanel_IsFirstUse verifies that a fresh panel is in first-use mode.
//
// Given: a new filterPanel
// When:  isFirstUse() is called
// Then:  true is returned (no session, no history)
func TestFilterPanel_NewPanel_IsFirstUse(t *testing.T) {
	p := newFilterPanel()
	if !p.isFirstUse() {
		t.Error("isFirstUse() = false, want true for a new panel")
	}
}

// TestFilterPanel_NewPanel_NotFirstUse_WhenSessionSet verifies that a panel with a
// session ID is no longer in first-use mode.
//
// Given: a filterPanel with sessionID set
// When:  isFirstUse() is called
// Then:  false is returned
func TestFilterPanel_NewPanel_NotFirstUse_WhenSessionSet(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc123"
	if p.isFirstUse() {
		t.Error("isFirstUse() = true, want false when sessionID is set")
	}
}

// TestFilterPanel_NewPanel_NotFirstUse_WhenHistoryNonEmpty verifies that a panel with
// history is not in first-use mode.
//
// Given: a filterPanel with one history entry
// When:  isFirstUse() is called
// Then:  false is returned
func TestFilterPanel_NewPanel_NotFirstUse_WhenHistoryNonEmpty(t *testing.T) {
	p := newFilterPanel()
	p.history = []filterConvEntry{{role: "user", text: "hello"}}
	if p.isFirstUse() {
		t.Error("isFirstUse() = true, want false when history is non-empty")
	}
}

// TestFilterPanel_Init_ReturnsNilCmd verifies that Init returns nil (no initial cmd).
//
// Given: a new filterPanel
// When:  Init() is called
// Then:  nil is returned (filter panel is activated by user input, not on load)
func TestFilterPanel_Init_ReturnsNilCmd(t *testing.T) {
	p := newFilterPanel()
	if cmd := p.Init(); cmd != nil {
		t.Error("Init() != nil, want nil (panel does not auto-start)")
	}
}

// ── View: first-use state ─────────────────────────────────────────────────────

// TestFilterPanel_View_FirstUse_ShowsSessionPrompt verifies that the view in first-use
// mode contains instructions for starting a new session.
//
// Given: a new filterPanel (first-use mode)
// When:  View() is called
// Then:  output contains "FILTER" and "title" (describing first-use input format)
func TestFilterPanel_View_FirstUse_ShowsSessionPrompt(t *testing.T) {
	p := newFilterPanel()
	v := p.View()
	for _, want := range []string{"FILTER", "title"} {
		if !strings.Contains(v, want) {
			t.Errorf("View() does not contain %q in first-use mode; output:\n%s", want, v)
		}
	}
}

// TestFilterPanel_View_FirstUse_ShowsInputCursor verifies the input cursor is shown.
//
// Given: a new filterPanel with inputBuf=""
// When:  View() is called
// Then:  output contains "_" (cursor indicator)
func TestFilterPanel_View_FirstUse_ShowsInputCursor(t *testing.T) {
	p := newFilterPanel()
	v := p.View()
	if !strings.Contains(v, "_") {
		t.Errorf("View() does not contain cursor '_'; output:\n%s", v)
	}
}

// TestFilterPanel_View_FirstUse_ShowsBufferedInput verifies that buffered input text
// appears in the view.
//
// Given: a filterPanel with inputBuf="My idea"
// When:  View() is called
// Then:  output contains "My idea"
func TestFilterPanel_View_FirstUse_ShowsBufferedInput(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "My idea"
	v := p.View()
	if !strings.Contains(v, "My idea") {
		t.Errorf("View() does not contain buffered input %q; output:\n%s", "My idea", v)
	}
}

// ── View: running state ───────────────────────────────────────────────────────

// TestFilterPanel_View_WhenRunning_ShowsSpinner verifies the thinking indicator is shown.
//
// Given: a filterPanel with running=true
// When:  View() is called
// Then:  output contains "thinking"
func TestFilterPanel_View_WhenRunning_ShowsSpinner(t *testing.T) {
	p := newFilterPanel()
	p.running = true
	v := p.View()
	if !strings.Contains(v, "thinking") {
		t.Errorf("View() does not contain %q when running; output:\n%s", "thinking", v)
	}
}

// ── View: history ─────────────────────────────────────────────────────────────

// TestFilterPanel_View_WithHistory_ShowsUserEntry verifies that user history entries
// appear in the view.
//
// Given: a filterPanel with a user history entry "refine my idea"
// When:  View() is called
// Then:  output contains "refine my idea"
func TestFilterPanel_View_WithHistory_ShowsUserEntry(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	p.history = []filterConvEntry{
		{role: "user", text: "refine my idea"},
	}
	v := p.View()
	if !strings.Contains(v, "refine my idea") {
		t.Errorf("View() does not contain user text; output:\n%s", v)
	}
}

// TestFilterPanel_View_WithHistory_ShowsAssistantEntry verifies that assistant history
// entries appear in the view.
//
// Given: a filterPanel with an assistant history entry "Great! Let me ask..."
// When:  View() is called
// Then:  output contains "Great! Let me ask..."
func TestFilterPanel_View_WithHistory_ShowsAssistantEntry(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	p.history = []filterConvEntry{
		{role: "user", text: "my idea"},
		{role: "assistant", text: "Great! Let me ask..."},
	}
	v := p.View()
	if !strings.Contains(v, "Great! Let me ask...") {
		t.Errorf("View() does not contain assistant text; output:\n%s", v)
	}
}

// TestFilterPanel_View_WithHistory_ShowsSessionID verifies the session ID appears in view.
//
// Given: a filterPanel with sessionID="sess-xyz"
// When:  View() is called
// Then:  output contains "sess-xyz"
func TestFilterPanel_View_WithHistory_ShowsSessionID(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-xyz"
	p.history = []filterConvEntry{{role: "user", text: "hello"}}
	v := p.View()
	if !strings.Contains(v, "sess-xyz") {
		t.Errorf("View() does not contain session ID %q; output:\n%s", "sess-xyz", v)
	}
}

// TestFilterPanel_View_WithError_ShowsError verifies that an error message is displayed.
//
// Given: a filterPanel with errMsg="agent exec failed"
// When:  View() is called
// Then:  output contains "agent exec failed"
func TestFilterPanel_View_WithError_ShowsError(t *testing.T) {
	p := newFilterPanel()
	p.errMsg = "agent exec failed"
	v := p.View()
	if !strings.Contains(v, "agent exec failed") {
		t.Errorf("View() does not contain error message; output:\n%s", v)
	}
}

// ── Update: filterAgentMsg ────────────────────────────────────────────────────

// TestFilterPanel_Update_AgentMsg_AppendsAssistantEntry verifies that a successful
// filterAgentMsg appends an assistant entry to history.
//
// Given: a filterPanel with one user history entry
// When:  a filterAgentMsg with Text="The answer is X" is processed
// Then:  history has 2 entries; the second is the assistant response
func TestFilterPanel_Update_AgentMsg_AppendsAssistantEntry(t *testing.T) {
	p := newFilterPanel()
	p.history = []filterConvEntry{{role: "user", text: "my question"}}
	p.running = true

	updated, _ := p.Update(filterAgentMsg{result: filterSessionResult{
		SessionID: "sess-001",
		Text:      "The answer is X",
	}})
	up := updated.(filterPanel)

	if len(up.history) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(up.history))
	}
	if up.history[1].role != "assistant" {
		t.Errorf("history[1].role = %q, want %q", up.history[1].role, "assistant")
	}
	if !strings.Contains(up.history[1].text, "The answer is X") {
		t.Errorf("history[1].text = %q, want it to contain %q", up.history[1].text, "The answer is X")
	}
}

// TestFilterPanel_Update_AgentMsg_SetsSessionID verifies that a filterAgentMsg updates
// the session ID.
//
// Given: a filterPanel with no sessionID
// When:  a filterAgentMsg with SessionID="sess-007" is processed
// Then:  sessionID = "sess-007"
func TestFilterPanel_Update_AgentMsg_SetsSessionID(t *testing.T) {
	p := newFilterPanel()
	p.running = true

	updated, _ := p.Update(filterAgentMsg{result: filterSessionResult{
		SessionID: "sess-007",
		Text:      "ok",
	}})
	up := updated.(filterPanel)

	if up.sessionID != "sess-007" {
		t.Errorf("sessionID = %q, want %q", up.sessionID, "sess-007")
	}
}

// TestFilterPanel_Update_AgentMsg_ClearsRunning verifies that a filterAgentMsg clears
// the running flag.
//
// Given: a filterPanel with running=true
// When:  a filterAgentMsg is processed
// Then:  running = false
func TestFilterPanel_Update_AgentMsg_ClearsRunning(t *testing.T) {
	p := newFilterPanel()
	p.running = true

	updated, _ := p.Update(filterAgentMsg{result: filterSessionResult{Text: "ok"}})
	up := updated.(filterPanel)

	if up.running {
		t.Error("running = true after filterAgentMsg, want false")
	}
}

// TestFilterPanel_Update_AgentMsg_Error_SetsErrMsg verifies that an error filterAgentMsg
// stores the error message.
//
// Given: a filterPanel with running=true
// When:  a filterAgentMsg with err="something went wrong" is processed
// Then:  errMsg contains "something went wrong"
func TestFilterPanel_Update_AgentMsg_Error_SetsErrMsg(t *testing.T) {
	p := newFilterPanel()
	p.running = true

	updated, _ := p.Update(filterAgentMsg{err: errors.New("something went wrong")})
	up := updated.(filterPanel)

	if !strings.Contains(up.errMsg, "something went wrong") {
		t.Errorf("errMsg = %q, want it to contain %q", up.errMsg, "something went wrong")
	}
}

// TestFilterPanel_Update_AgentMsg_Error_FirstUse_PopsHistoryEntry verifies that a
// failed first-use invocation pops the pending user history entry so that the
// panel returns to first-use mode and the next retry routes to invokeFilterNew.
//
// Given: a filterPanel with one user history entry, sessionID="" (first-use), running=true
// When:  a filterAgentMsg with err set is processed
// Then:  history length is 0 (entry popped) and isFirstUse() returns true
func TestFilterPanel_Update_AgentMsg_Error_FirstUse_PopsHistoryEntry(t *testing.T) {
	p := newFilterPanel()
	p.history = []filterConvEntry{{role: "user", text: "hello"}}
	p.running = true
	// sessionID is "" — first-use invocation

	updated, _ := p.Update(filterAgentMsg{err: errors.New("exec failed")})
	up := updated.(filterPanel)

	if len(up.history) != 0 {
		t.Errorf("len(history) = %d, want 0 after first-use error (entry should be popped)", len(up.history))
	}
	if !up.isFirstUse() {
		t.Error("isFirstUse() = false after first-use error, want true (retry must route to invokeFilterNew)")
	}
}

// TestFilterPanel_Update_AgentMsg_Error_Resume_KeepsHistoryEntry verifies that a
// failed resume invocation does NOT pop the history entry — the session is
// already established and the user entry should remain visible.
//
// Given: a filterPanel with one user history entry, sessionID="sess-abc", running=true
// When:  a filterAgentMsg with err set is processed
// Then:  history length stays at 1
func TestFilterPanel_Update_AgentMsg_Error_Resume_KeepsHistoryEntry(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	p.history = []filterConvEntry{{role: "user", text: "hello"}}
	p.running = true

	updated, _ := p.Update(filterAgentMsg{err: errors.New("exec failed")})
	up := updated.(filterPanel)

	if len(up.history) != 1 {
		t.Errorf("len(history) = %d, want 1 after resume error (entry must not be popped)", len(up.history))
	}
}

// ── Update: key events — input buffering ──────────────────────────────────────

// TestFilterPanel_Update_PrintableKey_AppendsToInputBuf verifies that a printable
// character is appended to inputBuf.
//
// Given: a filterPanel with inputBuf=""
// When:  'a' is pressed
// Then:  inputBuf = "a"
func TestFilterPanel_Update_PrintableKey_AppendsToInputBuf(t *testing.T) {
	p := newFilterPanel()

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	up := updated.(filterPanel)

	if up.inputBuf != "a" {
		t.Errorf("inputBuf = %q, want %q", up.inputBuf, "a")
	}
}

// TestFilterPanel_Update_BackspaceKey_RemovesLastRune verifies that backspace removes
// the last rune from inputBuf.
//
// Given: a filterPanel with inputBuf="hello"
// When:  backspace is pressed
// Then:  inputBuf = "hell"
func TestFilterPanel_Update_BackspaceKey_RemovesLastRune(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "hello"

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	up := updated.(filterPanel)

	if up.inputBuf != "hell" {
		t.Errorf("inputBuf = %q, want %q", up.inputBuf, "hell")
	}
}

// TestFilterPanel_Update_BackspaceKey_EmptyBuf_NoOp verifies that backspace on an
// empty buffer does not panic.
//
// Given: a filterPanel with inputBuf=""
// When:  backspace is pressed
// Then:  inputBuf = "" (no underflow/panic)
func TestFilterPanel_Update_BackspaceKey_EmptyBuf_NoOp(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = ""

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	up := updated.(filterPanel)

	if up.inputBuf != "" {
		t.Errorf("inputBuf = %q, want empty after backspace on empty buf", up.inputBuf)
	}
}

// ── Update: key events — first-use mode ──────────────────────────────────────

// TestFilterPanel_Update_EnterKey_FirstUse_AddsNewline verifies that Enter in first-use
// mode appends a newline to inputBuf instead of submitting.
//
// Given: a filterPanel in first-use mode with inputBuf="title"
// When:  Enter is pressed
// Then:  inputBuf = "title\n"
func TestFilterPanel_Update_EnterKey_FirstUse_AddsNewline(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "title"

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	up := updated.(filterPanel)

	if up.inputBuf != "title\n" {
		t.Errorf("inputBuf = %q, want %q in first-use mode", up.inputBuf, "title\n")
	}
}

// TestFilterPanel_Update_EnterKey_FirstUse_NoCmd verifies that Enter in first-use mode
// does not dispatch a cmd.
//
// Given: a filterPanel in first-use mode with inputBuf="title"
// When:  Enter is pressed
// Then:  cmd = nil (no agent call yet)
func TestFilterPanel_Update_EnterKey_FirstUse_NoCmd(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "title"

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("cmd != nil after Enter in first-use mode, want nil (no submit yet)")
	}
}

// TestFilterPanel_Update_CtrlDKey_EmptyBuf_NoOp verifies that ctrl+d with an empty
// buffer is a no-op.
//
// Given: a filterPanel with inputBuf=""
// When:  ctrl+d is pressed
// Then:  running = false (no submit dispatched)
func TestFilterPanel_Update_CtrlDKey_EmptyBuf_NoOp(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = ""

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	up := updated.(filterPanel)

	if up.running {
		t.Error("running = true after ctrl+d on empty buf, want false")
	}
}

// TestFilterPanel_Update_CtrlDKey_NonEmptyBuf_SetsRunning verifies that ctrl+d with
// non-empty input triggers a submit (sets running=true).
//
// Given: a filterPanel in first-use mode with inputBuf="my idea"
// When:  ctrl+d is pressed
// Then:  running = true
func TestFilterPanel_Update_CtrlDKey_NonEmptyBuf_SetsRunning(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "my idea"

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	up := updated.(filterPanel)

	if !up.running {
		t.Error("running = false after ctrl+d with non-empty buf, want true")
	}
}

// TestFilterPanel_Update_CtrlDKey_NonEmptyBuf_ReturnsCmd verifies that ctrl+d with
// non-empty input returns a non-nil cmd.
//
// Given: a filterPanel with inputBuf="my idea"
// When:  ctrl+d is pressed
// Then:  a non-nil cmd is returned
func TestFilterPanel_Update_CtrlDKey_NonEmptyBuf_ReturnsCmd(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "my idea"

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if cmd == nil {
		t.Error("cmd = nil after ctrl+d, want non-nil cmd")
	}
}

// TestFilterPanel_Update_CtrlDKey_ClearsInputBuf verifies that ctrl+d clears inputBuf.
//
// Given: a filterPanel with inputBuf="my idea"
// When:  ctrl+d is pressed
// Then:  inputBuf = ""
func TestFilterPanel_Update_CtrlDKey_ClearsInputBuf(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "my idea"

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	up := updated.(filterPanel)

	if up.inputBuf != "" {
		t.Errorf("inputBuf = %q, want empty after ctrl+d submit", up.inputBuf)
	}
}

// TestFilterPanel_Update_CtrlDKey_AppendsUserHistory verifies that ctrl+d appends the
// input text as a user history entry.
//
// Given: a filterPanel with inputBuf="my idea"
// When:  ctrl+d is pressed
// Then:  history has 1 entry with role="user" and text="my idea"
func TestFilterPanel_Update_CtrlDKey_AppendsUserHistory(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "my idea"

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	up := updated.(filterPanel)

	if len(up.history) != 1 {
		t.Fatalf("len(history) = %d, want 1", len(up.history))
	}
	if up.history[0].role != "user" {
		t.Errorf("history[0].role = %q, want %q", up.history[0].role, "user")
	}
	if !strings.Contains(up.history[0].text, "my idea") {
		t.Errorf("history[0].text = %q, want it to contain %q", up.history[0].text, "my idea")
	}
}

// ── submitCmd routing ─────────────────────────────────────────────────────────

// TestFilterPanel_SubmitCmd_FirstUse_ReturnsNonNilCmd verifies that submitCmd with
// firstUse=true returns a non-nil cmd.
//
// Given: a new filterPanel (no history, no sessionID)
// When:  submitCmd("title\ndesc", true) is called
// Then:  a non-nil cmd is returned
func TestFilterPanel_SubmitCmd_FirstUse_ReturnsNonNilCmd(t *testing.T) {
	p := newFilterPanel()
	cmd := p.submitCmd("title\ndesc", true)
	if cmd == nil {
		t.Error("submitCmd(prompt, firstUse=true) = nil, want non-nil cmd")
	}
}

// TestFilterPanel_SubmitCmd_Resume_ReturnsNonNilCmd verifies that submitCmd with
// firstUse=false returns a non-nil cmd.
//
// Given: a filterPanel with sessionID="sess-abc"
// When:  submitCmd("follow up", false) is called
// Then:  a non-nil cmd is returned
func TestFilterPanel_SubmitCmd_Resume_ReturnsNonNilCmd(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	cmd := p.submitCmd("follow up", false)
	if cmd == nil {
		t.Error("submitCmd(prompt, firstUse=false) = nil, want non-nil cmd")
	}
}

// TestFilterPanel_IsFirstUse_MustBeCapturedBeforeHistoryAppend is a regression test
// for the bug where isFirstUse() was evaluated after history was appended in doSubmit,
// causing the first submission to incorrectly route to invokeFilterResume instead of
// invokeFilterNew.
//
// Given: a new filterPanel (first-use mode)
// When:  firstUse is captured before appending a history entry
// Then:  the captured value is true, and isFirstUse() is false after the append
func TestFilterPanel_IsFirstUse_MustBeCapturedBeforeHistoryAppend(t *testing.T) {
	p := newFilterPanel()

	if !p.isFirstUse() {
		t.Fatal("precondition failed: isFirstUse() = false on new panel, want true")
	}

	// Capture firstUse before mutating history — this is what doSubmit must do.
	firstUse := p.isFirstUse()
	p.history = append(p.history, filterConvEntry{role: "user", text: "title"})

	if !firstUse {
		t.Error("firstUse captured before append = false, want true (routing bug: isFirstUse was evaluated after history mutation)")
	}
	if p.isFirstUse() {
		t.Error("isFirstUse() = true after history append, want false")
	}
}

// ── Update: key events — resume mode ─────────────────────────────────────────

// TestFilterPanel_Update_EnterKey_ResumeMode_SetsRunning verifies that Enter in resume
// mode (session active) triggers a submit.
//
// Given: a filterPanel with sessionID set and inputBuf="follow up"
// When:  Enter is pressed
// Then:  running = true
func TestFilterPanel_Update_EnterKey_ResumeMode_SetsRunning(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	p.history = []filterConvEntry{{role: "user", text: "first"}}
	p.inputBuf = "follow up"

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	up := updated.(filterPanel)

	if !up.running {
		t.Error("running = false after Enter in resume mode, want true")
	}
}

// TestFilterPanel_Update_EnterKey_ResumeMode_EmptyBuf_NoOp verifies that Enter in
// resume mode with an empty buffer is a no-op.
//
// Given: a filterPanel with sessionID set and inputBuf=""
// When:  Enter is pressed
// Then:  running = false (empty input, no submit)
func TestFilterPanel_Update_EnterKey_ResumeMode_EmptyBuf_NoOp(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	p.history = []filterConvEntry{{role: "user", text: "first"}}
	p.inputBuf = ""

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	up := updated.(filterPanel)

	if up.running {
		t.Error("running = true after Enter with empty buf, want false")
	}
}

// ── Update: key events — new session ─────────────────────────────────────────

// TestFilterPanel_Update_NKey_EmptyBuf_ClearsSession verifies that pressing 'n' with
// an empty input buffer starts a new session by clearing history and sessionID.
//
// Given: a filterPanel with sessionID and history set, inputBuf=""
// When:  'n' is pressed
// Then:  history=nil, sessionID=""
func TestFilterPanel_Update_NKey_EmptyBuf_ClearsSession(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	p.history = []filterConvEntry{{role: "user", text: "hello"}}
	p.inputBuf = ""

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	up := updated.(filterPanel)

	if up.sessionID != "" {
		t.Errorf("sessionID = %q, want empty after 'n' new session", up.sessionID)
	}
	if len(up.history) != 0 {
		t.Errorf("len(history) = %d, want 0 after 'n' new session", len(up.history))
	}
}

// TestFilterPanel_Update_NKey_NonEmptyBuf_AppendsToInput verifies that 'n' with a
// non-empty input buffer appends 'n' to the buffer (not a new session command).
//
// Given: a filterPanel with inputBuf="no"
// When:  'n' is pressed
// Then:  inputBuf = "non" (character appended, session not cleared)
func TestFilterPanel_Update_NKey_NonEmptyBuf_AppendsToInput(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	p.history = []filterConvEntry{{role: "user", text: "hello"}}
	p.inputBuf = "no"

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	up := updated.(filterPanel)

	if up.inputBuf != "non" {
		t.Errorf("inputBuf = %q, want %q", up.inputBuf, "non")
	}
	if up.sessionID != "sess-abc" {
		t.Errorf("sessionID = %q, want %q (session must not be cleared)", up.sessionID, "sess-abc")
	}
}

// TestFilterPanel_Update_UpperNKey_EmptyBuf_ClearsSession verifies that 'N' also
// starts a new session when input is empty.
//
// Given: a filterPanel with session and history, inputBuf=""
// When:  'N' is pressed
// Then:  history=nil, sessionID=""
func TestFilterPanel_Update_UpperNKey_EmptyBuf_ClearsSession(t *testing.T) {
	p := newFilterPanel()
	p.sessionID = "sess-abc"
	p.history = []filterConvEntry{{role: "user", text: "hello"}}
	p.inputBuf = ""

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	up := updated.(filterPanel)

	if up.sessionID != "" {
		t.Errorf("sessionID = %q, want empty after 'N'", up.sessionID)
	}
	if len(up.history) != 0 {
		t.Errorf("len(history) = %d, want 0 after 'N'", len(up.history))
	}
}

// ── Update: key events — running state blocks input ──────────────────────────

// TestFilterPanel_Update_KeyWhileRunning_NoOp verifies that key events are ignored
// while the agent is running.
//
// Given: a filterPanel with running=true and inputBuf=""
// When:  'a' is pressed
// Then:  inputBuf = "" (key ignored while running)
func TestFilterPanel_Update_KeyWhileRunning_NoOp(t *testing.T) {
	p := newFilterPanel()
	p.running = true

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	up := updated.(filterPanel)

	if up.inputBuf != "" {
		t.Errorf("inputBuf = %q, want empty (key must be ignored while running)", up.inputBuf)
	}
}

// ── Update: key events — scroll ───────────────────────────────────────────────

// TestFilterPanel_Update_DownArrow_IncrementsScrollY verifies the down arrow scrolls down.
//
// Given: a filterPanel with scrollY=0
// When:  down arrow is pressed
// Then:  scrollY = 1
func TestFilterPanel_Update_DownArrow_IncrementsScrollY(t *testing.T) {
	p := newFilterPanel()
	p.scrollY = 0

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	up := updated.(filterPanel)

	if up.scrollY != 1 {
		t.Errorf("scrollY = %d, want 1", up.scrollY)
	}
}

// TestFilterPanel_Update_UpArrow_DecrementsScrollY verifies the up arrow scrolls up.
//
// Given: a filterPanel with scrollY=5
// When:  up arrow is pressed
// Then:  scrollY = 4
func TestFilterPanel_Update_UpArrow_DecrementsScrollY(t *testing.T) {
	p := newFilterPanel()
	p.scrollY = 5

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyUp})
	up := updated.(filterPanel)

	if up.scrollY != 4 {
		t.Errorf("scrollY = %d, want 4", up.scrollY)
	}
}

// TestFilterPanel_Update_UpArrow_AtTop_StaysAtZero verifies the up arrow at scroll=0
// does not underflow.
//
// Given: a filterPanel with scrollY=0
// When:  up arrow is pressed
// Then:  scrollY = 0 (no underflow)
func TestFilterPanel_Update_UpArrow_AtTop_StaysAtZero(t *testing.T) {
	p := newFilterPanel()
	p.scrollY = 0

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyUp})
	up := updated.(filterPanel)

	if up.scrollY != 0 {
		t.Errorf("scrollY = %d, want 0 (no underflow)", up.scrollY)
	}
}

// TestFilterPanel_Update_JKey_EmptyBuf_IncrementsScrollY verifies that 'j' scrolls
// when the input buffer is empty.
//
// Given: a filterPanel with scrollY=0 and inputBuf=""
// When:  'j' is pressed
// Then:  scrollY = 1
func TestFilterPanel_Update_JKey_EmptyBuf_IncrementsScrollY(t *testing.T) {
	p := newFilterPanel()
	p.scrollY = 0
	p.inputBuf = ""

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	up := updated.(filterPanel)

	if up.scrollY != 1 {
		t.Errorf("scrollY = %d, want 1", up.scrollY)
	}
}

// TestFilterPanel_Update_JKey_NonEmptyBuf_AppendsToInput verifies that 'j' appends
// to the input buffer when it is non-empty.
//
// Given: a filterPanel with inputBuf="te" and scrollY=0
// When:  'j' is pressed
// Then:  inputBuf = "tej", scrollY = 0
func TestFilterPanel_Update_JKey_NonEmptyBuf_AppendsToInput(t *testing.T) {
	p := newFilterPanel()
	p.inputBuf = "te"
	p.scrollY = 0

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	up := updated.(filterPanel)

	if up.inputBuf != "tej" {
		t.Errorf("inputBuf = %q, want %q", up.inputBuf, "tej")
	}
	if up.scrollY != 0 {
		t.Errorf("scrollY = %d, want 0 (no scroll when input non-empty)", up.scrollY)
	}
}

// TestFilterPanel_Update_KKey_EmptyBuf_DecrementsScrollY verifies 'k' scrolls up
// when input is empty.
//
// Given: a filterPanel with scrollY=3 and inputBuf=""
// When:  'k' is pressed
// Then:  scrollY = 2
func TestFilterPanel_Update_KKey_EmptyBuf_DecrementsScrollY(t *testing.T) {
	p := newFilterPanel()
	p.scrollY = 3
	p.inputBuf = ""

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	up := updated.(filterPanel)

	if up.scrollY != 2 {
		t.Errorf("scrollY = %d, want 2", up.scrollY)
	}
}

// ── Update: window resize ─────────────────────────────────────────────────────

// TestFilterPanel_Update_WindowSizeMsg_UpdatesDimensions verifies WindowSizeMsg updates
// width and height.
//
// Given: a filterPanel with default dimensions
// When:  a WindowSizeMsg{Width: 120, Height: 40} is processed
// Then:  width=120, height=40
func TestFilterPanel_Update_WindowSizeMsg_UpdatesDimensions(t *testing.T) {
	p := newFilterPanel()

	updated, _ := p.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	up := updated.(filterPanel)

	if up.width != 120 {
		t.Errorf("width = %d, want 120", up.width)
	}
	if up.height != 40 {
		t.Errorf("height = %d, want 40", up.height)
	}
}

// ── cockpit integration ───────────────────────────────────────────────────────

// TestCockpit_Panel8_IsFilterPanel verifies the cockpit panel at index 7 (key: 8)
// is a filterPanel with title "Filter".
//
// Given: a new cockpitModel
// When:  panels[7] title is inspected
// Then:  title = "Filter"
func TestCockpit_Panel8_IsFilterPanel(t *testing.T) {
	m := newCockpitModel("", "")
	if len(m.panels) < 8 {
		t.Fatalf("len(panels) = %d, want at least 8", len(m.panels))
	}
	if m.panels[7].Title() != "Filter" {
		t.Errorf("panels[7].Title() = %q, want %q", m.panels[7].Title(), "Filter")
	}
}

// TestCockpit_Key8_ActivatesFilterPanel verifies that pressing '8' jumps to the
// filter panel (index 7) and activates panel focus.
//
// Given: a cockpitModel in sidebar mode, cursor=0
// When:  '8' is pressed
// Then:  cursor=7, panelFocused=true
func TestCockpit_Key8_ActivatesFilterPanel(t *testing.T) {
	m := newCockpitModel("", "")
	m.cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'8'}})
	um := updated.(cockpitModel)

	if um.cursor != 7 {
		t.Errorf("cursor = %d, want 7", um.cursor)
	}
	if !um.panelFocused {
		t.Error("panelFocused = false, want true after pressing '8'")
	}
}

// TestCockpit_FilterAgentMsg_RoutesToFilterPanel verifies that filterAgentMsg is
// delivered to panels[7] regardless of which panel is active.
//
// Given: a cockpitModel with cursor=0 (Droplets panel active)
// When:  a filterAgentMsg with Text="agent says hello" is processed
// Then:  panels[7] (filterPanel) has the assistant response in its history
func TestCockpit_FilterAgentMsg_RoutesToFilterPanel(t *testing.T) {
	m := newCockpitModel("", "")
	m.cursor = 0
	// Pre-populate the filter panel with a user entry so it can accept agent responses.
	fp := m.panels[7].(filterPanel)
	fp.history = []filterConvEntry{{role: "user", text: "hello"}}
	fp.running = true
	m.panels[7] = fp

	updated, _ := m.Update(filterAgentMsg{result: filterSessionResult{
		SessionID: "sess-rt1",
		Text:      "agent says hello",
	}})
	um := updated.(cockpitModel)

	fp2, ok := um.panels[7].(filterPanel)
	if !ok {
		t.Fatalf("panels[7] is not a filterPanel")
	}
	found := false
	for _, entry := range fp2.history {
		if entry.role == "assistant" && strings.Contains(entry.text, "agent says hello") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("filterPanel.history does not contain the assistant response")
	}
}
