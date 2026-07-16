// Package parser adapts a real CommonMark/GFM engine (goldmark) into
// Lexicon's domain.Document. A hand-rolled line scanner was deliberately
// rejected: Lexicon's core promise is that a .lex.md file is valid,
// renderable markdown, and that promise is a structural property (list
// continuation, table shape) that a regex-over-lines approach can't verify.
// goldmark is also the engine under Hugo, so the choice is consistent with
// the stack omarcrosby.com already runs on.
package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	gext "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"

	"github.com/jedi-knights/lexicon/internal/domain"
)

var md = goldmark.New(goldmark.WithExtensions(extension.Table))

var (
	tagLineRe = regexp.MustCompile(`^(@[A-Za-z0-9_-]+)(\s+@[A-Za-z0-9_-]+)*$`)
	keywordRe = regexp.MustCompile(`^(Given|When|Then|Precondition|Action|Outcome|And|But)$`)
	quotedRe  = regexp.MustCompile(`"([^"]*)"`)
)

// GoldmarkParser is the ports.DocumentParser implementation backing every
// Lexicon command.
type GoldmarkParser struct{}

// New returns a ready-to-use GoldmarkParser.
func New() *GoldmarkParser { return &GoldmarkParser{} }

// Parse implements ports.DocumentParser.
func (p *GoldmarkParser) Parse(_ context.Context, r io.Reader) (*domain.Document, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("lexicon: read source: %w", err)
	}

	fm, body, err := splitFrontMatter(raw)
	if err != nil {
		return nil, err
	}

	source := []byte(body)
	root := md.Parser().Parse(text.NewReader(source))

	doc := &domain.Document{FrontMatter: fm}

	var pendingTags []string
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		var err error
		switch node := n.(type) {
		case *gast.Heading:
			err = handleHeading(doc, node, source, &pendingTags)
		case *gast.Paragraph, *gast.TextBlock:
			handleParagraph(doc, n, source, &pendingTags)
		case *gast.List:
			err = handleList(doc, node, source)
		case *gext.Table:
			err = handleTable(doc, node, source)
		}
		if err != nil {
			return nil, err
		}
	}

	if doc.Feature == nil {
		return nil, fmt.Errorf("lexicon: no Feature (H1 heading) found")
	}
	return doc, nil
}

func handleList(doc *domain.Document, node *gast.List, source []byte) error {
	sc := currentScenario(doc.Feature)
	if sc == nil {
		return fmt.Errorf("lexicon: found step bullets before any Scenario or Background heading")
	}
	steps, err := parseSteps(node, source)
	if err != nil {
		return err
	}
	sc.Steps = append(sc.Steps, steps...)
	return nil
}

func handleTable(doc *domain.Document, node *gext.Table, source []byte) error {
	sc := currentScenario(doc.Feature)
	if sc == nil {
		return fmt.Errorf("lexicon: found a table before any Scenario heading")
	}
	sc.Examples = parseTable(node, source)
	return nil
}

func handleHeading(doc *domain.Document, node *gast.Heading, source []byte, pendingTags *[]string) error {
	title := strings.TrimSpace(renderText(node, source))
	switch node.Level {
	case 1:
		doc.Feature = &domain.Feature{
			Name: strings.TrimPrefix(title, "Feature: "),
			Tags: *pendingTags,
		}
		*pendingTags = nil
	case 2:
		if doc.Feature == nil {
			return fmt.Errorf("lexicon: found a Scenario/Background heading (%q) before any Feature (H1) heading", title)
		}
		isBackground := title == "Background"
		sc := &domain.Scenario{
			Name:         strings.TrimPrefix(title, "Scenario: "),
			Tags:         *pendingTags,
			IsBackground: isBackground,
		}
		*pendingTags = nil
		if isBackground {
			if doc.Feature.Background != nil {
				return fmt.Errorf("lexicon: a Feature may have only one Background")
			}
			doc.Feature.Background = sc
		} else {
			doc.Feature.Scenarios = append(doc.Feature.Scenarios, sc)
		}
	default:
		return fmt.Errorf("lexicon: heading level %d is not part of the Lexicon grammar (only H1 Feature and H2 Scenario/Background are)", node.Level)
	}
	return nil
}

func handleParagraph(doc *domain.Document, node gast.Node, source []byte, pendingTags *[]string) {
	txt := strings.TrimSpace(renderText(node, source))
	if tags := parseTagLine(txt); tags != nil {
		*pendingTags = tags
		return
	}
	if doc.Feature != nil && len(doc.Feature.Scenarios) == 0 && doc.Feature.Background == nil {
		if doc.Feature.Description != "" {
			doc.Feature.Description += "\n\n"
		}
		doc.Feature.Description += txt
	}
}

// currentScenario returns the Scenario that a subsequent List or Table block
// belongs to: the most recently opened real Scenario, or the Background if
// no real Scenario has been opened yet.
func currentScenario(f *domain.Feature) *domain.Scenario {
	if f == nil {
		return nil
	}
	if n := len(f.Scenarios); n > 0 {
		return f.Scenarios[n-1]
	}
	return f.Background
}

func parseTagLine(s string) []string {
	if !tagLineRe.MatchString(s) {
		return nil
	}
	fields := strings.Fields(s)
	tags := make([]string, len(fields))
	for i, field := range fields {
		tags[i] = strings.TrimPrefix(field, "@")
	}
	return tags
}

func parseSteps(list *gast.List, source []byte) ([]*domain.Step, error) {
	var steps []*domain.Step
	var lastRole domain.StepRole

	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		li, ok := item.(*gast.ListItem)
		if !ok {
			continue
		}
		step, ok, err := parseStep(li, source, lastRole)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		lastRole = step.Role
		steps = append(steps, step)
	}
	return steps, nil
}

// parseStep parses a single step bullet's list item. The bool return is
// false when the item has no block content at all (an edge case goldmark
// itself shouldn't produce, but handled rather than assumed impossible).
func parseStep(li *gast.ListItem, source []byte, lastRole domain.StepRole) (*domain.Step, bool, error) {
	content, docStringNode := stepListItemChildren(li)
	if content == nil {
		return nil, false, nil
	}

	emph, ok := content.FirstChild().(*gast.Emphasis)
	if !ok || emph.Level != 2 {
		return nil, false, fmt.Errorf("lexicon: step bullet %q does not start with a bold keyword (- **Given** ...)", strings.TrimSpace(renderText(content, source)))
	}
	keyword := strings.TrimSpace(renderText(emph, source))
	if !keywordRe.MatchString(keyword) {
		return nil, false, fmt.Errorf("lexicon: %q is not a recognized step keyword (Given/When/Then or Precondition/Action/Outcome, plus And/But)", keyword)
	}

	var rest strings.Builder
	for c := emph.NextSibling(); c != nil; c = c.NextSibling() {
		rest.WriteString(renderText(c, source))
	}
	stepText := strings.TrimSpace(rest.String())
	role := roleFor(domain.StepKeyword(keyword), lastRole)

	step := &domain.Step{
		Keyword:    domain.StepKeyword(keyword),
		Role:       role,
		Text:       stepText,
		Parameters: extractParameters(stepText),
	}
	if docStringNode != nil {
		step.DocString = &domain.DocString{
			ContentType: string(docStringNode.Language(source)),
			Content:     strings.TrimRight(string(docStringNode.Lines().Value(source)), "\n"),
		}
	}
	return step, true, nil
}

// stepListItemChildren splits a step bullet's list-item children into its
// text content block (the bold-keyword paragraph/text-block) and an
// optional trailing doc-string fence.
func stepListItemChildren(li *gast.ListItem) (content gast.Node, docString *gast.FencedCodeBlock) {
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		switch v := c.(type) {
		case *gast.TextBlock, *gast.Paragraph:
			if content == nil {
				content = c
			}
		case *gast.FencedCodeBlock:
			docString = v
		}
	}
	return content, docString
}

func roleFor(kw domain.StepKeyword, last domain.StepRole) domain.StepRole {
	switch kw {
	case domain.KeywordGiven, domain.KeywordPrecondition:
		return domain.RolePrecondition
	case domain.KeywordWhen, domain.KeywordAction:
		return domain.RoleAction
	case domain.KeywordThen, domain.KeywordOutcome:
		return domain.RoleOutcome
	default: // And, But — inherit the role of the step they continue
		return last
	}
}

func extractParameters(s string) []string {
	matches := quotedRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}
	params := make([]string, 0, len(matches))
	for _, m := range matches {
		params = append(params, m[1])
	}
	return params
}

func parseTable(tbl *gext.Table, source []byte) *domain.Table {
	t := &domain.Table{}
	for row := tbl.FirstChild(); row != nil; row = row.NextSibling() {
		var cells []string
		for c := row.FirstChild(); c != nil; c = c.NextSibling() {
			cells = append(cells, strings.TrimSpace(renderText(c, source)))
		}
		switch row.(type) {
		case *gext.TableHeader:
			t.Headers = cells
		case *gext.TableRow:
			t.Rows = append(t.Rows, cells)
		}
	}
	return t
}

// renderText concatenates the literal text of every descendant leaf of n.
// Written by hand rather than relying on ast.BaseNode.Text (deprecated
// upstream) — the leaf kinds Lexicon cares about (Text inside a paragraph,
// TextBlock, or CodeSpan) are exactly the two cases handled below. A
// CodeSpan's only children are Text nodes covering its content with the
// backtick delimiters already excluded by goldmark's own inline parser,
// which is exactly what lets “ `<name>` “ round-trip to a bare `<name>`
// token without any special-casing here.
func renderText(n gast.Node, source []byte) string {
	var buf bytes.Buffer
	var walk func(n gast.Node)
	walk = func(n gast.Node) {
		if t, ok := n.(*gast.Text); ok {
			buf.Write(t.Segment.Value(source))
			if t.SoftLineBreak() {
				buf.WriteByte(' ')
			}
			return
		}
		if s, ok := n.(*gast.String); ok {
			buf.Write(s.Value)
			return
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			walk(c)
		}
	}
	walk(n)
	return buf.String()
}

// splitFrontMatter strips an optional leading YAML front matter block
// (`---`-delimited) and returns it decoded, along with the remaining body to
// hand to goldmark. Front matter is intentionally NOT used for tags — see
// domain.FrontMatter's doc comment — so only a fixed set of bookkeeping keys
// is recognized; anything else lands in Extra rather than being dropped.
func splitFrontMatter(raw []byte) (*domain.FrontMatter, string, error) {
	s := string(raw)
	if !strings.HasPrefix(s, "---\n") {
		return nil, s, nil
	}
	rest := s[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return nil, "", fmt.Errorf("lexicon: unterminated front matter block (missing closing ---)")
	}
	yamlBlock := rest[:end]
	body := strings.TrimPrefix(rest[end+len("\n---"):], "\n")

	var raw2 map[string]any
	if err := yaml.Unmarshal([]byte(yamlBlock), &raw2); err != nil {
		return nil, "", fmt.Errorf("lexicon: parse front matter: %w", err)
	}

	fm := &domain.FrontMatter{Extra: map[string]string{}}
	for k, v := range raw2 {
		applyFrontMatterField(fm, k, v)
	}
	if len(fm.Extra) == 0 {
		fm.Extra = nil
	}
	return fm, body, nil
}

// applyFrontMatterField assigns one decoded YAML front matter key into fm,
// recognizing a fixed bookkeeping set (id/owner/status/links) and routing
// anything else into Extra rather than dropping it.
func applyFrontMatterField(fm *domain.FrontMatter, key string, value any) {
	switch key {
	case "id":
		fm.ID, _ = value.(string)
	case "owner":
		fm.Owner, _ = value.(string)
	case "status":
		fm.Status, _ = value.(string)
	case "links":
		list, ok := value.([]any)
		if !ok {
			return
		}
		for _, item := range list {
			if str, ok := item.(string); ok {
				fm.Links = append(fm.Links, str)
			}
		}
	default:
		fm.Extra[key] = fmt.Sprintf("%v", value)
	}
}
