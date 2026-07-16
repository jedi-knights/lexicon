package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jedi-knights/lexicon/internal/adapters/emitter"
	"github.com/jedi-knights/lexicon/internal/adapters/parser"
)

func newCompileCmd() *cobra.Command {
	var to string
	var listTargets bool

	cmd := &cobra.Command{
		Use:   "compile <file>",
		Short: "Translate a .lex.md file into another format (Gherkin, Gauge, JSON).",
		RunE: func(cmd *cobra.Command, args []string) error {
			if listTargets {
				for _, name := range emitter.Names() {
					fmt.Fprintln(cmd.OutOrStdout(), name) //nolint:errcheck // stdout write failures surface on the next write
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("lexicon compile: expected exactly one file argument")
			}
			em, ok := emitter.Get(to)
			if !ok {
				return fmt.Errorf("lexicon compile: unknown target %q (run 'lexicon compile --list-targets' to see available targets)", to)
			}

			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("lexicon compile: %w", err)
			}
			defer func() { _ = f.Close() }()

			doc, err := parser.New().Parse(context.Background(), f)
			if err != nil {
				return fmt.Errorf("lexicon compile: %w", err)
			}
			return em.Emit(cmd.OutOrStdout(), doc)
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "target format (see --list-targets)")
	cmd.Flags().BoolVar(&listTargets, "list-targets", false, "list every registered target format and exit")
	return cmd
}
