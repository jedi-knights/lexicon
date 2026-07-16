package domain

// Scenario is one H2 block within a Feature: a named sequence of steps,
// optionally tagged, optionally carrying an Examples table that turns it
// into an outline. A Scenario with IsBackground set represents the optional
// H2 "Background" block, whose steps run before every other Scenario in the
// Feature.
type Scenario struct {
	Name         string   `json:"name"`
	Tags         []string `json:"tags,omitempty"`
	Steps        []*Step  `json:"steps"`
	Examples     *Table   `json:"examples,omitempty"`
	IsBackground bool     `json:"is_background,omitempty"`
}
