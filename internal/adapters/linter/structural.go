// Package linter backs `lexicon check`: structural validity independent of
// any translation target (missing outcomes, empty scenarios), as distinct
// from whether a document translates cleanly into Gherkin or Gauge.
package linter

import (
	"github.com/jedi-knights/lexicon/internal/domain"
	"github.com/jedi-knights/lexicon/internal/ports"
)

// Structural is the ports.Checker implementation backing `lexicon check`.
type Structural struct{}

// Check implements ports.Checker.
func (Structural) Check(doc *domain.Document) []ports.Finding {
	if doc.Feature == nil {
		return []ports.Finding{{Severity: ports.SeverityError, Message: "document has no Feature (H1) heading"}}
	}

	f := doc.Feature
	var findings []ports.Finding
	if len(f.Scenarios) == 0 {
		findings = append(findings, ports.Finding{Severity: ports.SeverityError, Message: "Feature has no Scenario"})
	}
	for _, sc := range f.Scenarios {
		findings = append(findings, checkScenario(sc)...)
	}
	if f.Background != nil {
		findings = append(findings, checkScenario(f.Background)...)
	}
	return findings
}

func checkScenario(sc *domain.Scenario) []ports.Finding {
	if len(sc.Steps) == 0 {
		return []ports.Finding{{Severity: ports.SeverityError, Scenario: sc.Name, Message: "scenario has no steps"}}
	}

	var hasOutcome bool
	for _, s := range sc.Steps {
		if s.Role == domain.RoleOutcome {
			hasOutcome = true
			break
		}
	}

	var findings []ports.Finding
	if !sc.IsBackground && !hasOutcome {
		findings = append(findings, ports.Finding{Severity: ports.SeverityError, Scenario: sc.Name, Message: "scenario has no Then (outcome) step"})
	}
	return findings
}
