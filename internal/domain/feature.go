package domain

// Feature is the H1 root of a single .lex.md document: a named group of
// scenarios, with an optional Background run before each one.
type Feature struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Background  *Scenario   `json:"background,omitempty"`
	Scenarios   []*Scenario `json:"scenarios"`
}
