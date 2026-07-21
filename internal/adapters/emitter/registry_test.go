package emitter_test

import (
	"testing"

	"github.com/jedi-knights/lexicon/internal/adapters/emitter"
)

// TestExtension locks in the file extension each registered target produces
// for directory-mode compilation, so a future target that forgets to
// implement Extension fails to compile rather than silently writing
// extensionless files.
func TestExtension(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{format: "gherkin", want: ".feature"},
		{format: "gauge", want: ".spec"},
		{format: "json", want: ".json"},
		{format: "robot", want: ".robot"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			// Arrange
			em, ok := emitter.Get(tt.format)
			if !ok {
				t.Fatalf("no emitter registered for %q", tt.format)
			}

			// Act
			got := em.Extension()

			// Assert
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
