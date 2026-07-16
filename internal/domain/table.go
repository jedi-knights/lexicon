package domain

// Table is a GFM data table. Its only use in v1 is a Scenario's Examples —
// the table's mere presence, combined with backtick-escaped `<name>`
// placeholders in step text, implies outline behavior. No separate
// "Scenario Outline"/"Examples:" heading is required, since modern Gherkin
// (v6+) doesn't need one and Gauge never did.
type Table struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}
