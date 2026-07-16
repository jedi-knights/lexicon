package commands

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/jedi-knights/lexicon/internal/adapters/emitter"
	"github.com/jedi-knights/lexicon/internal/adapters/parser"
	"github.com/jedi-knights/lexicon/internal/ports"
)

func newCompileCmd() *cobra.Command {
	var to string
	var out string
	var workers int
	var listTargets bool

	cmd := &cobra.Command{
		Use:   "compile [path]",
		Short: "Translate .lex.md file(s) into another format (Gherkin, Gauge, JSON).",
		Long: "Translate .lex.md file(s) into another format (Gherkin, Gauge, JSON).\n\n" +
			"Given a file, compiles it to stdout. Given a directory (the default is " +
			"the current directory when no path is given), recursively compiles every " +
			".lex.md file found under it into --out, mirroring the input's directory " +
			"structure.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if listTargets {
				return runListTargets(cmd)
			}
			path, err := compilePathArg(args)
			if err != nil {
				return err
			}
			em, ok := emitter.Get(to)
			if !ok {
				return fmt.Errorf("lexicon compile: unknown target %q (run 'lexicon compile --list-targets' to see available targets)", to)
			}

			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("lexicon compile: %w", err)
			}
			if info.IsDir() {
				return runCompileDirectory(cmd, path, out, workers, em)
			}
			return runCompileFile(cmd, path, em)
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "target format (see --list-targets)")
	cmd.Flags().StringVarP(&out, "out", "o", "", "output directory root (required when compiling a directory)")
	cmd.Flags().IntVar(&workers, "workers", 0, "max concurrent file compilations when compiling a directory (default: number of CPUs)")
	cmd.Flags().BoolVar(&listTargets, "list-targets", false, "list every registered target format and exit")
	return cmd
}

// runListTargets prints every registered target format to cmd's stdout.
func runListTargets(cmd *cobra.Command) error {
	for _, name := range emitter.Names() {
		fmt.Fprintln(cmd.OutOrStdout(), name) //nolint:errcheck // stdout write failures surface on the next write
	}
	return nil
}

// compilePathArg resolves the compile command's optional positional
// argument, defaulting to the current directory when omitted.
func compilePathArg(args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("lexicon compile: expected at most one path argument")
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return ".", nil
}

// runCompileFile compiles a single file to cmd's stdout, the pre-existing
// compile behavior.
func runCompileFile(cmd *cobra.Command, path string, em ports.Emitter) error {
	if cmd.Flags().Changed("out") {
		return fmt.Errorf("lexicon compile: --out only applies when compiling a directory; " +
			"for a single file, redirect stdout instead (e.g. lexicon compile foo.lex.md --to gherkin > foo.feature)")
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("lexicon compile: %w", err)
	}
	defer func() { _ = f.Close() }()

	doc, err := parser.New().Parse(context.Background(), f)
	if err != nil {
		return fmt.Errorf("lexicon compile: %w", err)
	}
	return em.Emit(cmd.OutOrStdout(), doc)
}

// runCompileDirectory validates --out and --workers, then delegates to
// compileDir to walk path and compile every .lex.md file found under it.
func runCompileDirectory(cmd *cobra.Command, path, out string, workers int, em ports.Emitter) error {
	if out == "" {
		return fmt.Errorf("lexicon compile: --out is required when compiling a directory")
	}
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	return compileDir(cmd.Context(), path, out, em, workers, cmd.OutOrStdout())
}
