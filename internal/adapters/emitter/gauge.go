package emitter

import (
	"fmt"
	"io"
	"strings"

	"github.com/jedi-knights/lexicon/internal/domain"
)

// Gauge emits a domain.Document as a Gauge specification (itself already a
// markdown dialect: H1 Specification, H2 Scenario, `*` step bullets).
//
// Gauge has no doc-string construct — a step's DocString content is emitted
// as a visible HTML comment rather than silently dropped, so translation
// loss is legible in the output instead of hidden. See the README's
// verification-asymmetry note: Gauge fidelity is checked against
// hand-authored fixtures for v1, not a real gauge-binary round-trip.
type Gauge struct{}

// Format implements ports.Emitter.
func (Gauge) Format() string { return "gauge" }

// Emit implements ports.Emitter.
func (Gauge) Emit(w io.Writer, doc *domain.Document) error {
	f := doc.Feature
	if f == nil {
		return fmt.Errorf("lexicon: document has no Feature to emit")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", f.Name)
	writeGaugeTags(&b, f.Tags)
	if f.Description != "" {
		fmt.Fprintf(&b, "\n%s\n", f.Description)
	}

	if f.Background != nil {
		b.WriteString("\n")
		writeGaugeSteps(&b, f.Background.Steps)
	}

	for _, sc := range f.Scenarios {
		fmt.Fprintf(&b, "\n## %s\n", sc.Name)
		writeGaugeTags(&b, sc.Tags)
		b.WriteString("\n")
		writeGaugeSteps(&b, sc.Steps)
		if sc.Examples != nil {
			writeGaugeTable(&b, sc.Examples)
		}
	}

	_, err := io.WriteString(w, strings.TrimRight(b.String(), "\n")+"\n")
	return err
}

func writeGaugeTags(b *strings.Builder, tags []string) {
	if len(tags) == 0 {
		return
	}
	fmt.Fprintf(b, "Tags: %s\n", strings.Join(tags, ", "))
}

func writeGaugeSteps(b *strings.Builder, steps []*domain.Step) {
	for _, s := range steps {
		fmt.Fprintf(b, "* %s\n", s.Text)
		if s.DocString != nil {
			fmt.Fprintf(b, "<!-- lexicon: doc string not representable in Gauge syntax, kept for reference:\n%s\n-->\n", s.DocString.Content)
		}
	}
}

func writeGaugeTable(b *strings.Builder, t *domain.Table) {
	b.WriteString(gaugeTableRow(t.Headers))
	for _, row := range t.Rows {
		b.WriteString(gaugeTableRow(row))
	}
}

func gaugeTableRow(cells []string) string {
	return "|" + strings.Join(cells, "|") + "|\n"
}
