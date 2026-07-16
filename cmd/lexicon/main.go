// Command lexicon is the CLI entrypoint: parse a .lex.md file, compile it to
// another format, or structurally lint it.
package main

import (
	"fmt"
	"os"

	"github.com/jedi-knights/lexicon/cmd/lexicon/commands"
)

var version = "dev"

func main() {
	if err := commands.NewRoot(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err) //nolint:errcheck // nothing meaningful to do if stderr itself fails
		os.Exit(1)
	}
}
