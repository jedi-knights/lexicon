package emitter

import "testing"

// TestRobotVariable locks in the translation of an Examples/outline
// placeholder into Robot Framework's own variable syntax: Gherkin and Gauge
// both use bare `<name>` natively, but RF's templated-test mechanism needs
// `${name}` to bind the value `[Arguments]` declares for the backing
// keyword — so, unlike the other two emitters, this substitution is real
// translation work, not a pass-through.
func TestRobotVariable(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		headers []string
		want    string
	}{
		{
			name:    "single placeholder matching a header",
			text:    "the user clicks the page <page> control",
			headers: []string{"page"},
			want:    "the user clicks the page ${page} control",
		},
		{
			name:    "multiple headers, each substituted",
			text:    "<first> then <second>",
			headers: []string{"first", "second"},
			want:    "${first} then ${second}",
		},
		{
			name:    "no headers leaves text unchanged",
			text:    "the second page of results replaces the first page",
			headers: nil,
			want:    "the second page of results replaces the first page",
		},
		{
			name:    "angle brackets not matching any header are left alone",
			text:    "the <page> page, not the <other> one",
			headers: []string{"page"},
			want:    "the ${page} page, not the <other> one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange — no setup beyond the table row.

			// Act
			got := robotVariables(tt.text, tt.headers)

			// Assert
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
