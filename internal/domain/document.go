// Package domain holds Lexicon's target-agnostic representation of a parsed
// .lex.md file. Nothing here names Gherkin or Gauge: any future translation
// target that has some notion of "a named scenario made of ordered steps
// with optional tags/examples/context" can be satisfied by a new emitter
// with no changes to these types.
package domain

// Document is the root of a parsed .lex.md file.
type Document struct {
	FrontMatter *FrontMatter `json:"front_matter,omitempty"`
	Feature     *Feature     `json:"feature"`
}
