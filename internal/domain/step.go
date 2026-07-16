package domain

// StepKeyword is the surface BDD vocabulary a step bullet is written with —
// reused verbatim from Gherkin because an already-known vocabulary is the
// most memorable choice available, not because Lexicon is Gherkin-specific.
type StepKeyword string

// The five recognized surface step keywords.
const (
	KeywordGiven StepKeyword = "Given"
	KeywordWhen  StepKeyword = "When"
	KeywordThen  StepKeyword = "Then"
	KeywordAnd   StepKeyword = "And"
	KeywordBut   StepKeyword = "But"
)

// StepRole is the semantic category a step belongs to, independent of which
// surface keyword it was written with — And/But steps inherit the Role of
// the nearest preceding Given/When/Then in the same scenario. This is the
// layer that carries the actual thesis (a requirement read differently by
// Dev, QA, and Product because its precondition, action, and outcome were
// left implicit): Role is what a consuming human or LLM should key off of,
// with Keyword as just the surface spelling.
type StepRole string

// The three recognized semantic roles.
const (
	RolePrecondition StepRole = "precondition"
	RoleAction       StepRole = "action"
	RoleOutcome      StepRole = "outcome"
)

// DocString is a fenced-code-block argument attached to a step, mirroring
// Gherkin's doc-string step argument. Gauge has no equivalent construct —
// see the Gauge emitter for how this is represented there.
type DocString struct {
	ContentType string `json:"content_type,omitempty"`
	Content     string `json:"content"`
}

// Step is one bullet within a Scenario or Background.
type Step struct {
	Keyword StepKeyword `json:"keyword"`
	Role    StepRole    `json:"role"`
	Text    string      `json:"text"`
	// Parameters holds every quoted-string substring found in Text, mirroring
	// Gauge's own convention of treating quoted text as a step parameter —
	// applied universally so JSON consumers get structured values instead of
	// having to re-extract them from prose.
	Parameters []string   `json:"parameters,omitempty"`
	DocString  *DocString `json:"doc_string,omitempty"`
}
