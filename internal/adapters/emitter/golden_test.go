package emitter_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gherkin "github.com/cucumber/gherkin/go/v28"

	"github.com/jedi-knights/lexicon/internal/adapters/emitter"
	"github.com/jedi-knights/lexicon/internal/adapters/parser"
)

// TestGolden parses every testdata/<case>/input.lex.md fixture and compiles
// it to every registered target, comparing byte-for-byte against
// expected.<ext>. It also feeds the emitted Gherkin through the real
// cucumber/gherkin library — the single highest-value guard against
// "looks right, doesn't actually parse" per the plan's verification section.
// Gauge fidelity has no equivalent cheap real-parser check available (see
// README), so expected.spec is reviewed by hand against the Gauge docs.
func TestGolden(t *testing.T) {
	cases, err := filepath.Glob("testdata/*")
	if err != nil || len(cases) == 0 {
		t.Fatalf("no testdata cases found: %v", err)
	}

	targets := map[string]string{
		"gherkin": "expected.feature",
		"gauge":   "expected.spec",
		"json":    "expected.json",
		"robot":   "expected.robot",
	}

	for _, dir := range cases {
		dir := dir
		t.Run(filepath.Base(dir), func(t *testing.T) {
			input, err := os.Open(filepath.Join(dir, "input.lex.md"))
			if err != nil {
				t.Fatalf("open input: %v", err)
			}
			defer input.Close()

			doc, err := parser.New().Parse(context.Background(), input)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			for format, filename := range targets {
				em, ok := emitter.Get(format)
				if !ok {
					t.Fatalf("no emitter registered for %q", format)
				}

				var buf bytes.Buffer
				if err := em.Emit(&buf, doc); err != nil {
					t.Fatalf("emit %s: %v", format, err)
				}

				wantPath := filepath.Join(dir, filename)
				want, err := os.ReadFile(wantPath)
				if err != nil {
					t.Fatalf("read %s: %v", wantPath, err)
				}
				if buf.String() != string(want) {
					t.Errorf("%s output mismatch\n--- got ---\n%s\n--- want ---\n%s", format, buf.String(), string(want))
				}

				if format == "gherkin" {
					if _, err := gherkin.ParseGherkinDocument(strings.NewReader(buf.String()), func() string { return "0" }); err != nil {
						t.Errorf("emitted .feature does not parse as real Gherkin: %v\n%s", err, buf.String())
					}
				}
			}
		})
	}
}
