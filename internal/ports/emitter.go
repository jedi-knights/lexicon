package ports

import (
	"io"

	"github.com/jedi-knights/lexicon/internal/domain"
)

// Emitter renders a domain.Document into one translation target. It is the
// extensibility seam: adding a target later (Robot Framework, an in-house
// format, whatever comes next) means writing one new Emitter implementation
// and registering it — nothing else in Lexicon changes.
type Emitter interface {
	// Format is the registry key this Emitter is selected by, e.g. "gherkin".
	Format() string
	Emit(w io.Writer, doc *domain.Document) error
}
