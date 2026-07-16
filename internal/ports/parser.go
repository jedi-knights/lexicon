// Package ports defines the interfaces by which the domain interacts with
// the outside world: parsing .lex.md source, emitting translation targets,
// checking structural validity. Concrete implementations live in
// internal/adapters/.
package ports

import (
	"context"
	"io"

	"github.com/jedi-knights/lexicon/internal/domain"
)

// DocumentParser parses .lex.md source into a domain.Document. It is the
// only seam through which raw markdown enters Lexicon; everything
// downstream (emitters, checkers) operates on domain.Document values.
type DocumentParser interface {
	Parse(ctx context.Context, r io.Reader) (*domain.Document, error)
}
