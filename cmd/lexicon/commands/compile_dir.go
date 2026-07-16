package commands

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/jedi-knights/lexicon/internal/adapters/parser"
	"github.com/jedi-knights/lexicon/internal/ports"
)

// findLexFiles returns every ".lex.md" file under root, walked recursively.
// filepath.WalkDir visits entries in lexical order within each directory,
// so the result is deterministic without an extra sort.
func findLexFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".lex.md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("lexicon compile: walking %s: %w", root, err)
	}
	return files, nil
}

// compileOneFile parses srcPath and emits it under outDir, mirroring
// srcPath's position relative to root and swapping the ".lex.md" suffix for
// em's target extension.
func compileOneFile(ctx context.Context, root, srcPath, outDir string, em ports.Emitter) error {
	relPath, err := filepath.Rel(root, srcPath)
	if err != nil {
		return fmt.Errorf("%s: %w", srcPath, err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("%s: %w", relPath, err)
	}
	defer func() { _ = src.Close() }()

	doc, err := parser.New().Parse(ctx, src)
	if err != nil {
		return fmt.Errorf("%s: %w", relPath, err)
	}

	dstRel := strings.TrimSuffix(relPath, ".lex.md") + em.Extension()
	dstPath := filepath.Join(outDir, dstRel)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("%s: %w", relPath, err)
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("%s: %w", relPath, err)
	}
	defer func() { _ = dst.Close() }()

	if err := em.Emit(dst, doc); err != nil {
		return fmt.Errorf("%s: %w", relPath, err)
	}
	return nil
}

// compileDir compiles every ".lex.md" file under root to em's target format,
// writing results under outDir in a tree that mirrors root. Up to workers
// files are compiled concurrently. Every discovered file is attempted even
// if others fail; failures are reported by relative path to w, and a
// summary error is returned if any file failed.
func compileDir(ctx context.Context, root, outDir string, em ports.Emitter, workers int, w io.Writer) error {
	files, err := findLexFiles(root)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("lexicon compile: no .lex.md files found under %s", root)
	}

	errs := make([]error, len(files))
	g := new(errgroup.Group)
	g.SetLimit(workers)
	for i, file := range files {
		g.Go(func() error {
			errs[i] = compileOneFile(ctx, root, file, outDir, em)
			return nil
		})
	}
	_ = g.Wait() // errs[i] carries each file's outcome; compileOneFile goroutines never return an error here.

	var failed int
	for _, err := range errs {
		if err == nil {
			continue
		}
		failed++
		fmt.Fprintf(w, "%v\n", err) //nolint:errcheck // stdout write failures surface on the next write
	}
	if failed > 0 {
		return fmt.Errorf("lexicon compile: %d of %d file(s) failed", failed, len(files))
	}

	fmt.Fprintf(w, "compiled %d file(s) to %s\n", len(files), outDir) //nolint:errcheck // stdout write failures surface on the next write
	return nil
}
