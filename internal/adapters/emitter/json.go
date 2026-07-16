package emitter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jedi-knights/lexicon/internal/domain"
)

// JSON emits a domain.Document as a schema-stable JSON structure — the
// concrete answer to "optimized for LLM interpretation": every step
// surfaces both its surface Keyword (Given/When/Then/And/But) and its
// resolved semantic Role (precondition/action/outcome), so a consuming
// model reads structured fields instead of re-deriving BDD convention from
// prose.
type JSON struct{}

// Format implements ports.Emitter.
func (JSON) Format() string { return "json" }

// Emit implements ports.Emitter.
func (JSON) Emit(w io.Writer, doc *domain.Document) error {
	if doc.Feature == nil {
		return fmt.Errorf("lexicon: document has no Feature to emit")
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	// Step text routinely contains `<placeholder>` tokens; Go's default
	// HTML-escaping would render the angle brackets as < and >,
	// which is correct-but-unreadable for a format whose whole point is
	// that an LLM (or a human) can read it directly.
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}
