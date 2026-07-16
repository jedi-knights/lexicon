package domain

// StepKeyword is the surface vocabulary a step bullet is written with.
// Lexicon supports two equally valid dialects — Given/When/Then/And/But
// (the most widely recognized phrasing, used by Cucumber, Behave, SpecFlow,
// and other BDD tools, not just Gherkin) and Precondition/Action/Outcome/
// And/But (the dialect-neutral spelling, matching StepRole's own names
// exactly) — chosen freely per file. Both resolve to the same StepRole.
type StepKeyword string

// The BDD-style dialect.
const (
	KeywordGiven StepKeyword = "Given"
	KeywordWhen  StepKeyword = "When"
	KeywordThen  StepKeyword = "Then"
	KeywordAnd   StepKeyword = "And"
	KeywordBut   StepKeyword = "But"
)

// The dialect-neutral spelling. And/But are shared with the BDD dialect
// above rather than duplicated — they're ordinary English connectives, not
// BDD-specific vocabulary.
const (
	KeywordPrecondition StepKeyword = "Precondition"
	KeywordAction       StepKeyword = "Action"
	KeywordOutcome      StepKeyword = "Outcome"
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
