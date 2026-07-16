package linter_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jedi-knights/lexicon/internal/adapters/linter"
	"github.com/jedi-knights/lexicon/internal/adapters/parser"
	"github.com/jedi-knights/lexicon/internal/ports"
)

func TestCheck_Clean(t *testing.T) {
	src := "# Feature: X\n\n## Scenario: Y\n\n- **Given** a\n- **When** b\n- **Then** c\n"
	doc, err := parser.New().Parse(context.Background(), strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if findings := (linter.Structural{}).Check(doc); len(findings) != 0 {
		t.Errorf("Check() = %v, want no findings", findings)
	}
}

func TestCheck_MissingOutcome(t *testing.T) {
	src := "# Feature: X\n\n## Scenario: Y\n\n- **Given** a\n- **When** b\n"
	doc, err := parser.New().Parse(context.Background(), strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	findings := linter.Structural{}.Check(doc)
	if len(findings) != 1 || findings[0].Severity != ports.SeverityError {
		t.Fatalf("Check() = %v, want exactly one error finding", findings)
	}
	if !strings.Contains(findings[0].Message, "Then") {
		t.Errorf("finding message = %q, want it to mention the missing Then step", findings[0].Message)
	}
}

func TestCheck_NoScenarios(t *testing.T) {
	src := "# Feature: X\n\nJust a description, no scenarios.\n"
	doc, err := parser.New().Parse(context.Background(), strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	findings := linter.Structural{}.Check(doc)
	if len(findings) != 1 {
		t.Fatalf("Check() = %v, want exactly one finding for a Feature with no Scenario", findings)
	}
}
