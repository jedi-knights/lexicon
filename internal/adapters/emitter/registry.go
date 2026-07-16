package emitter

import (
	"sort"

	"github.com/jedi-knights/lexicon/internal/ports"
)

// registry is the open, in-process set of translation targets. Each emitter
// registers itself in init(); adding a new target means adding one new file
// with the same init() pattern below — nothing here or in the CLI needs a
// hardcoded case for a fixed set of formats.
var registry = map[string]ports.Emitter{}

func register(e ports.Emitter) {
	registry[e.Format()] = e
}

func init() {
	register(Gherkin{})
	register(Gauge{})
	register(JSON{})
}

// Get returns the Emitter registered for format, or false if none is.
func Get(format string) (ports.Emitter, bool) {
	e, ok := registry[format]
	return e, ok
}

// Names returns every registered format name, sorted for stable CLI output.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
