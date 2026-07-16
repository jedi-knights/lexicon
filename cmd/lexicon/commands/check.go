package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jedi-knights/lexicon/internal/adapters/linter"
	"github.com/jedi-knights/lexicon/internal/adapters/parser"
	"github.com/jedi-knights/lexicon/internal/ports"
)

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <file>",
		Short: "Structurally lint a .lex.md file (missing outcomes, empty scenarios).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("lexicon check: %w", err)
			}
			defer func() { _ = f.Close() }()

			doc, err := parser.New().Parse(context.Background(), f)
			if err != nil {
				return fmt.Errorf("lexicon check: %w", err)
			}

			findings := linter.Structural{}.Check(doc)
			for _, finding := range findings {
				where := args[0]
				if finding.Scenario != "" {
					where += " (" + finding.Scenario + ")"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s: %s\n", where, finding.Severity, finding.Message) //nolint:errcheck // stdout write failures surface on the next write
			}
			if hasError(findings) {
				return fmt.Errorf("lexicon check: %d finding(s)", len(findings))
			}
			return nil
		},
	}
}

func hasError(findings []ports.Finding) bool {
	for _, f := range findings {
		if f.Severity == ports.SeverityError {
			return true
		}
	}
	return false
}
