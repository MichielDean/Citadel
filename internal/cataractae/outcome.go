package cataractae

import "fmt"

// Outcome is the result of a step execution, parsed from outcome.json.
type Outcome struct {
	// Result is the step outcome: pass, fail, revision, or escalate.
	Result string `json:"result"`

	// Notes is a free-text explanation of what the agent did or found.
	Notes string `json:"notes"`

	// Annotations holds optional key-value data from the step execution.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Valid result values.
const (
	ResultPass     = "pass"
	ResultFail     = "fail"
	ResultRecirculate = "recirculate"
	ResultEscalate = "escalate"
)

// Validate checks that the outcome has a recognized result value.
func (o *Outcome) Validate() error {
	switch o.Result {
	case ResultPass, ResultFail, ResultRecirculate, ResultEscalate:
		return nil
	case "":
		return fmt.Errorf("outcome: result is required")
	default:
			return fmt.Errorf("outcome: unknown result %q (want pass|fail|recirculate|escalate)", o.Result)
	}
}

// RouteField returns the workflow routing field name for this outcome.
// Maps result → on_pass, on_fail, on_recirculate, on_escalate.
func (o *Outcome) RouteField() string {
	switch o.Result {
	case ResultPass:
		return "on_pass"
	case ResultFail:
		return "on_fail"
	case ResultRecirculate:
		return "on_recirculate"
	case ResultEscalate:
		return "on_escalate"
	default:
		return ""
	}
}
