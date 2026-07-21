// Package emitter holds one file per translation target. Each type in this
// package is an independent, self-registering implementation of
// ports.Emitter — adding a new target later (Robot Framework, an in-house
// format) means adding one new file here with the same pattern; nothing
// elsewhere in Lexicon needs to change.
package emitter

import (
	"fmt"
	"io"
	"strings"

	"github.com/jedi-knights/lexicon/internal/domain"
)

// Gherkin emits a domain.Document as a Gherkin .feature file.
type Gherkin struct{}

// Format implements ports.Emitter.
func (Gherkin) Format() string { return "gherkin" }

// Extension implements ports.Emitter.
func (Gherkin) Extension() string { return ".feature" }

// Emit implements ports.Emitter.
func (Gherkin) Emit(w io.Writer, doc *domain.Document) error {
	f := doc.Feature
	if f == nil {
		return fmt.Errorf("lexicon: document has no Feature to emit")
	}

	var b strings.Builder
	writeGherkinTags(&b, f.Tags, "")
	fmt.Fprintf(&b, "Feature: %s\n", f.Name)
	if f.Description != "" {
		b.WriteString("\n")
		for _, line := range strings.Split(f.Description, "\n") {
			if line == "" {
				b.WriteString("\n")
				continue
			}
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}

	if f.Background != nil {
		b.WriteString("\n  Background:\n")
		writeGherkinSteps(&b, f.Background.Steps, "    ")
	}

	for _, sc := range f.Scenarios {
		b.WriteString("\n")
		writeGherkinTags(&b, sc.Tags, "  ")
		fmt.Fprintf(&b, "  Scenario: %s\n", sc.Name)
		writeGherkinSteps(&b, sc.Steps, "    ")
		if sc.Examples != nil {
			b.WriteString("\n    Examples:\n")
			writeGherkinTableRow(&b, sc.Examples.Headers, "      ")
			for _, row := range sc.Examples.Rows {
				writeGherkinTableRow(&b, row, "      ")
			}
		}
	}

	_, err := io.WriteString(w, strings.TrimRight(b.String(), "\n")+"\n")
	return err
}

func writeGherkinTags(b *strings.Builder, tags []string, indent string) {
	if len(tags) == 0 {
		return
	}
	parts := make([]string, len(tags))
	for i, t := range tags {
		parts[i] = "@" + t
	}
	fmt.Fprintf(b, "%s%s\n", indent, strings.Join(parts, " "))
}

// bddKeyword returns the Given/When/Then/And/But keyword for s: its own
// Keyword when that's already native BDD vocabulary (preserving an author's
// exact Given/When/Then/And/But), or the Given/When/Then equivalent of its
// Role when the source used the dialect-neutral spelling. Both Gherkin and
// Robot Framework require this same five-keyword vocabulary — Gherkin's
// grammar demands it outright, and RF strips the same five prefixes when
// matching keyword names — regardless of which dialect the .lex.md source
// was written in.
func bddKeyword(s *domain.Step) domain.StepKeyword {
	switch s.Keyword {
	case domain.KeywordGiven, domain.KeywordWhen, domain.KeywordThen, domain.KeywordAnd, domain.KeywordBut:
		return s.Keyword
	}
	switch s.Role {
	case domain.RolePrecondition:
		return domain.KeywordGiven
	case domain.RoleAction:
		return domain.KeywordWhen
	default:
		return domain.KeywordThen
	}
}

func writeGherkinSteps(b *strings.Builder, steps []*domain.Step, indent string) {
	for _, s := range steps {
		fmt.Fprintf(b, "%s%s %s\n", indent, bddKeyword(s), s.Text)
		if s.DocString != nil {
			fmt.Fprintf(b, "%s  \"\"\"%s\n", indent, s.DocString.ContentType)
			for _, line := range strings.Split(s.DocString.Content, "\n") {
				fmt.Fprintf(b, "%s  %s\n", indent, line)
			}
			fmt.Fprintf(b, "%s  \"\"\"\n", indent)
		}
	}
}

func writeGherkinTableRow(b *strings.Builder, cells []string, indent string) {
	fmt.Fprintf(b, "%s| %s |\n", indent, strings.Join(cells, " | "))
}
