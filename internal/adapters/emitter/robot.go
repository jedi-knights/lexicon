package emitter

import (
	"fmt"
	"io"
	"strings"

	"github.com/jedi-knights/lexicon/internal/domain"
)

// Robot emits a domain.Document as a Robot Framework test suite. A Scenario
// becomes a Test Case; Background becomes a `Test Setup` keyword (RF's own
// "runs before every test in this suite" construct) rather than repeating
// its steps in every Test Case; an Examples table becomes a `[Template]`
// Test Case whose data rows drive a backing Keyword — RF's own idiom for a
// Scenario Outline, and the reason a Scenario's steps only ever live in one
// of two places (directly under the Test Case, or under the backing
// Keyword) depending on whether Examples is set.
type Robot struct{}

// Format implements ports.Emitter.
func (Robot) Format() string { return "robot" }

// Extension implements ports.Emitter.
func (Robot) Extension() string { return ".robot" }

// Emit implements ports.Emitter.
func (Robot) Emit(w io.Writer, doc *domain.Document) error {
	f := doc.Feature
	if f == nil {
		return fmt.Errorf("lexicon: document has no Feature to emit")
	}

	var settings strings.Builder
	writeRobotSettings(&settings, f)
	testCases := robotTestCasesSection(f)
	keywords := robotKeywordsSection(f)

	var b strings.Builder
	if settings.Len() > 0 {
		b.WriteString("*** Settings ***\n")
		b.WriteString(settings.String())
	}
	if testCases != "" {
		b.WriteString("\n*** Test Cases ***\n")
		b.WriteString(testCases)
	}
	if keywords != "" {
		b.WriteString("\n*** Keywords ***\n")
		b.WriteString(keywords)
	}

	_, err := io.WriteString(w, strings.TrimLeft(strings.TrimRight(b.String(), "\n")+"\n", "\n"))
	return err
}

func robotTestCasesSection(f *domain.Feature) string {
	var b strings.Builder
	for i, sc := range f.Scenarios {
		if i > 0 {
			b.WriteString("\n")
		}
		writeRobotTestCase(&b, sc)
	}
	return b.String()
}

func robotKeywordsSection(f *domain.Feature) string {
	var b strings.Builder
	if f.Background != nil {
		writeRobotBackgroundKeyword(&b, f.Background)
	}
	for _, sc := range f.Scenarios {
		if sc.Examples == nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		writeRobotTemplateKeyword(&b, sc)
	}
	return b.String()
}

func writeRobotSettings(b *strings.Builder, f *domain.Feature) {
	if f.Description != "" {
		lines := strings.Split(f.Description, "\n")
		fmt.Fprintf(b, "Documentation    %s\n", lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(b, "...    %s\n", line)
		}
	}
	if len(f.Tags) > 0 {
		fmt.Fprintf(b, "Force Tags    %s\n", strings.Join(f.Tags, "    "))
	}
	if f.Background != nil {
		b.WriteString("Test Setup    Background\n")
	}
}

func writeRobotTestCase(b *strings.Builder, sc *domain.Scenario) {
	fmt.Fprintf(b, "%s\n", sc.Name)
	if len(sc.Tags) > 0 {
		fmt.Fprintf(b, "    [Tags]    %s\n", strings.Join(sc.Tags, "    "))
	}
	if sc.Examples == nil {
		writeRobotSteps(b, sc.Steps, nil)
		return
	}
	fmt.Fprintf(b, "    [Template]    %s\n", sc.Name)
	for _, row := range sc.Examples.Rows {
		fmt.Fprintf(b, "    %s\n", strings.Join(row, "    "))
	}
}

func writeRobotBackgroundKeyword(b *strings.Builder, background *domain.Scenario) {
	b.WriteString("Background\n")
	writeRobotSteps(b, background.Steps, nil)
}

func writeRobotTemplateKeyword(b *strings.Builder, sc *domain.Scenario) {
	fmt.Fprintf(b, "%s\n", sc.Name)
	args := make([]string, len(sc.Examples.Headers))
	for i, h := range sc.Examples.Headers {
		args[i] = "${" + h + "}"
	}
	fmt.Fprintf(b, "    [Arguments]    %s\n", strings.Join(args, "    "))
	writeRobotSteps(b, sc.Steps, sc.Examples.Headers)
}

func writeRobotSteps(b *strings.Builder, steps []*domain.Step, headers []string) {
	for _, s := range steps {
		fmt.Fprintf(b, "    %s %s\n", bddKeyword(s), robotVariables(s.Text, headers))
		if s.DocString != nil {
			for _, line := range strings.Split(s.DocString.Content, "\n") {
				fmt.Fprintf(b, "    # %s\n", line)
			}
		}
	}
}

// robotVariables translates every `<header>` placeholder in text into RF's
// own `${header}` variable syntax, for each name in headers — the binding
// that lets a `[Template]` Test Case's data rows reach the backing
// Keyword's `[Arguments]`.
func robotVariables(text string, headers []string) string {
	for _, h := range headers {
		text = strings.ReplaceAll(text, "<"+h+">", "${"+h+"}")
	}
	return text
}
