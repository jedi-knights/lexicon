package domain

// FrontMatter holds document bookkeeping that has no equivalent in Gherkin
// or Gauge — ticket linkage, ownership, review status. It is deliberately
// kept separate from Tags, which stays the single mechanism for anything
// BDD tooling needs to see.
type FrontMatter struct {
	ID     string            `json:"id,omitempty"`
	Owner  string            `json:"owner,omitempty"`
	Status string            `json:"status,omitempty"`
	Links  []string          `json:"links,omitempty"`
	Extra  map[string]string `json:"extra,omitempty"`
}
