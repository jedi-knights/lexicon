package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/jedi-knights/lexicon/internal/domain"
)

func mustParse(t *testing.T, src string) *domain.Document {
	t.Helper()
	doc, err := New().Parse(context.Background(), strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse: unexpected error: %v", err)
	}
	return doc
}

func TestParse_DocString(t *testing.T) {
	src := `# Feature: Empty search state

## Scenario: Rendering an empty state

- **Given** the search returns zero results
- **When** the results page renders
- **Then** the page shows the following message

  ` + "```text" + `
  No results found for your search.
  ` + "```" + `
`
	doc := mustParse(t, src)

	sc := doc.Feature.Scenarios[0]
	last := sc.Steps[len(sc.Steps)-1]
	if last.DocString == nil {
		t.Fatalf("expected a DocString on the final Then step, got none")
	}
	if got, want := last.DocString.ContentType, "text"; got != want {
		t.Errorf("DocString.ContentType = %q, want %q", got, want)
	}
	if got, want := last.DocString.Content, "No results found for your search."; got != want {
		t.Errorf("DocString.Content = %q, want %q", got, want)
	}
}

func TestParse_RoleInheritance(t *testing.T) {
	src := `# Feature: X

## Scenario: Y

- **Given** a
- **And** b
- **When** c
- **And** d
- **Then** e
- **But** f
`
	doc := mustParse(t, src)
	steps := doc.Feature.Scenarios[0].Steps
	want := []domain.StepRole{
		domain.RolePrecondition, domain.RolePrecondition,
		domain.RoleAction, domain.RoleAction,
		domain.RoleOutcome, domain.RoleOutcome,
	}
	if len(steps) != len(want) {
		t.Fatalf("got %d steps, want %d", len(steps), len(want))
	}
	for i, s := range steps {
		if s.Role != want[i] {
			t.Errorf("step %d (%s): Role = %q, want %q", i, s.Keyword, s.Role, want[i])
		}
	}
}

func TestParse_QuotedParameters(t *testing.T) {
	src := `# Feature: X

## Scenario: Y

- **When** the user searches for "cupcakes" in "bakery"
`
	doc := mustParse(t, src)
	step := doc.Feature.Scenarios[0].Steps[0]
	want := []string{"cupcakes", "bakery"}
	if len(step.Parameters) != len(want) {
		t.Fatalf("Parameters = %v, want %v", step.Parameters, want)
	}
	for i, p := range want {
		if step.Parameters[i] != p {
			t.Errorf("Parameters[%d] = %q, want %q", i, step.Parameters[i], p)
		}
	}
}

func TestParse_NoFeature(t *testing.T) {
	_, err := New().Parse(context.Background(), strings.NewReader("just some text, no heading\n"))
	if err == nil {
		t.Fatal("expected an error for a document with no Feature heading, got nil")
	}
}

func TestParse_ScenarioBeforeFeature(t *testing.T) {
	src := "## Scenario: orphaned\n\n- **Given** a\n"
	_, err := New().Parse(context.Background(), strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a Scenario heading before any Feature heading, got nil")
	}
}

func TestParse_StepWithoutBoldKeyword(t *testing.T) {
	src := "# Feature: X\n\n## Scenario: Y\n\n- the user does something\n"
	_, err := New().Parse(context.Background(), strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a step bullet with no bold leading keyword, got nil")
	}
}

func TestParse_UnknownKeyword(t *testing.T) {
	src := "# Feature: X\n\n## Scenario: Y\n\n- **Whenever** something happens\n"
	_, err := New().Parse(context.Background(), strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for an unrecognized step keyword, got nil")
	}
}

func TestParse_DuplicateBackground(t *testing.T) {
	src := `# Feature: X

## Background

- **Given** a

## Background

- **Given** b
`
	_, err := New().Parse(context.Background(), strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a second Background heading, got nil")
	}
}

func TestParse_FrontMatter(t *testing.T) {
	src := `---
id: SEARCH-042
owner: product
status: draft
links:
  - https://example.com/ticket/42
custom: keep-me
---

# Feature: X

## Scenario: Y

- **Given** a
- **Then** b
`
	doc := mustParse(t, src)
	fm := doc.FrontMatter
	if fm == nil {
		t.Fatal("expected non-nil FrontMatter")
	}
	if fm.ID != "SEARCH-042" || fm.Owner != "product" || fm.Status != "draft" {
		t.Errorf("FrontMatter = %+v, want id/owner/status set", fm)
	}
	if len(fm.Links) != 1 || fm.Links[0] != "https://example.com/ticket/42" {
		t.Errorf("FrontMatter.Links = %v, want one ticket link", fm.Links)
	}
	if fm.Extra["custom"] != "keep-me" {
		t.Errorf("FrontMatter.Extra[%q] = %q, want %q", "custom", fm.Extra["custom"], "keep-me")
	}
}

func TestParse_UnterminatedFrontMatter(t *testing.T) {
	src := "---\nid: X\n\n# Feature: X\n"
	_, err := New().Parse(context.Background(), strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for an unterminated front matter block, got nil")
	}
}
