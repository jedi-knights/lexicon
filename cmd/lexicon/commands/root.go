// Package commands wires the lexicon subcommands onto a cobra.Command tree.
// Each subcommand lives in its own file; this file only builds the root.
package commands

import "github.com/spf13/cobra"

// NewRoot constructs the lexicon root command with version injected.
func NewRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "lexicon",
		Short:         "A markdown-native requirements DSL, translatable to Gherkin, Gauge, and beyond.",
		Long:          "Lexicon parses .lex.md files — valid, renderable CommonMark that stays readable on GitHub and Slack with no special tooling — into a target-agnostic structured representation, then compiles that into Gherkin, Gauge, or JSON. Every step carries both its surface keyword (Given/When/Then/And/But) and its resolved semantic role (precondition/action/outcome), so Dev, QA, and Product — and an LLM reading the same file — read the same requirement the same way.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.AddCommand(newCompileCmd())
	root.AddCommand(newCheckCmd())
	root.AddCommand(newVersionCmd(version))
	return root
}
