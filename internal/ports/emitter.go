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
	// Extension is the file extension (including the leading dot, e.g.
	// ".feature") this Emitter's output conventionally uses. Directory-mode
	// compilation uses it to name each compiled file.
	Extension() string
}
