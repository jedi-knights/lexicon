package emitter

import (
	"testing"

	"github.com/jedi-knights/lexicon/internal/domain"
)

// TestBDDKeyword locks in the translation from either step-keyword dialect
// to the Given/When/Then/And/But vocabulary shared by Gherkin and Robot
// Framework (RF strips the same five prefixes when matching keyword names):
// native BDD keywords pass through verbatim (preserving an author's exact
// And/But choice), and the dialect-neutral spelling is translated via Role.
func TestBDDKeyword(t *testing.T) {
	tests := []struct {
		keyword domain.StepKeyword
		role    domain.StepRole
		want    domain.StepKeyword
	}{
		{keyword: domain.KeywordGiven, role: domain.RolePrecondition, want: domain.KeywordGiven},
		{keyword: domain.KeywordWhen, role: domain.RoleAction, want: domain.KeywordWhen},
		{keyword: domain.KeywordThen, role: domain.RoleOutcome, want: domain.KeywordThen},
		{keyword: domain.KeywordAnd, role: domain.RolePrecondition, want: domain.KeywordAnd},
		{keyword: domain.KeywordBut, role: domain.RoleOutcome, want: domain.KeywordBut},
		{keyword: domain.KeywordPrecondition, role: domain.RolePrecondition, want: domain.KeywordGiven},
		{keyword: domain.KeywordAction, role: domain.RoleAction, want: domain.KeywordWhen},
		{keyword: domain.KeywordOutcome, role: domain.RoleOutcome, want: domain.KeywordThen},
	}

	for _, tt := range tests {
		t.Run(string(tt.keyword), func(t *testing.T) {
			// Arrange
			step := &domain.Step{Keyword: tt.keyword, Role: tt.role}

			// Act
			got := bddKeyword(step)

			// Assert
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
