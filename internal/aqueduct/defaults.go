package aqueduct

import _ "embed"

//go:embed defaults/implementer.md
var DefaultImplementerRole string

//go:embed defaults/reviewer.md
var DefaultReviewerRole string

//go:embed defaults/qa.md
var DefaultQARole string

//go:embed defaults/security.md
var DefaultSecurityRole string

// BuiltinCataractaeDefinitions contains the default role definitions shipped with the binary.
var BuiltinCataractaeDefinitions = map[string]CataractaeDefinition{
	"implementer": {Name: "Implementer", Description: "Expert software engineer, TDD/BDD", Instructions: DefaultImplementerRole},
	"reviewer":    {Name: "Adversarial Reviewer", Description: "Adversarial code reviewer", Instructions: DefaultReviewerRole},
	"qa":          {Name: "QA Reviewer", Description: "Quality and test coverage reviewer", Instructions: DefaultQARole},
	"security":    {Name: "Security Reviewer", Description: "Security vulnerability auditor", Instructions: DefaultSecurityRole},
}
