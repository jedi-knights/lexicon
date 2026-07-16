package ports

import "github.com/jedi-knights/lexicon/internal/domain"

// Severity classifies a Finding.
type Severity string

// The two recognized Finding severities.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Finding is one structural issue reported by a Checker.
type Finding struct {
	Severity Severity
	Message  string
	Scenario string // empty when the finding is document-level
}

// Checker inspects a parsed Document for structural problems — an empty
// Feature, a Scenario with no Then, a malformed tag — independent of any
// translation target.
type Checker interface {
	Check(doc *domain.Document) []Finding
}
