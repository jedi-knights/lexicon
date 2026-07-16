package commands_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jedi-knights/lexicon/cmd/lexicon/commands"
)

const validLexMD = `# Feature: X

## Scenario: Y

- **Given** something
- **When** something happens
- **Then** something results
`

const invalidLexMD = "just some text, no heading\n"

// writeFile creates path (and its parent directories) under dir with the
// given content.
func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
}

// runCompile executes `lexicon compile <args...>`, returning combined
// stdout/stderr and the command's error.
func runCompile(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := commands.NewRoot("test")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(append([]string{"compile"}, args...))
	err := root.Execute()
	return buf.String(), err
}

func TestCompileDirectory_NestedTree(t *testing.T) {
	// Arrange
	root := t.TempDir()
	writeFile(t, root, "a.lex.md", validLexMD)
	writeFile(t, root, "sub/b.lex.md", validLexMD)
	out := t.TempDir()

	// Act
	stdout, err := runCompile(t, root, "--to", "gherkin", "--out", out)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, stdout)
	}
	for _, rel := range []string{"a.feature", "sub/b.feature"} {
		path := filepath.Join(out, rel)
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("expected output file %s: %v", path, readErr)
		}
		if !bytes.Contains(content, []byte("Feature: X")) {
			t.Errorf("%s: got %q, want it to contain %q", path, content, "Feature: X")
		}
	}
	if !bytes.Contains([]byte(stdout), []byte("compiled 2 file(s)")) {
		t.Errorf("stdout = %q, want a summary line mentioning 2 compiled files", stdout)
	}
}

func TestCompileDirectory_GaugeTarget(t *testing.T) {
	// Arrange
	root := t.TempDir()
	writeFile(t, root, "a.lex.md", validLexMD)
	out := t.TempDir()

	// Act
	stdout, err := runCompile(t, root, "--to", "gauge", "--out", out)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, stdout)
	}
	path := filepath.Join(out, "a.spec")
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("expected output file %s: %v", path, readErr)
	}
	if !bytes.Contains(content, []byte("# X")) {
		t.Errorf("%s: got %q, want it to contain %q", path, content, "# X")
	}
}

func TestCompileDirectory_PartialFailure(t *testing.T) {
	// Arrange
	root := t.TempDir()
	writeFile(t, root, "good.lex.md", validLexMD)
	writeFile(t, root, "bad.lex.md", invalidLexMD)
	out := t.TempDir()

	// Act
	stdout, err := runCompile(t, root, "--to", "gherkin", "--out", out)

	// Assert
	if err == nil {
		t.Fatal("expected an error because one file failed to compile, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(out, "good.feature")); statErr != nil {
		t.Errorf("expected the valid file to still be compiled despite the other failing: %v", statErr)
	}
	if !bytes.Contains([]byte(stdout), []byte("bad.lex.md")) {
		t.Errorf("stdout = %q, want it to name the failing file bad.lex.md", stdout)
	}
}

func TestCompileDirectory_NoMatchingFiles(t *testing.T) {
	// Arrange
	root := t.TempDir()
	out := t.TempDir()

	// Act
	_, err := runCompile(t, root, "--to", "gherkin", "--out", out)

	// Assert
	if err == nil {
		t.Fatal("expected an error for a directory with no .lex.md files, got nil")
	}
}

func TestCompileDirectory_RequiresOut(t *testing.T) {
	// Arrange
	root := t.TempDir()
	writeFile(t, root, "a.lex.md", validLexMD)

	// Act
	_, err := runCompile(t, root, "--to", "gherkin")

	// Assert
	if err == nil {
		t.Fatal("expected an error when --out is missing for a directory target, got nil")
	}
}

func TestCompileFile_RejectsOut(t *testing.T) {
	// Arrange
	root := t.TempDir()
	writeFile(t, root, "a.lex.md", validLexMD)
	out := t.TempDir()

	// Act
	_, err := runCompile(t, filepath.Join(root, "a.lex.md"), "--to", "gherkin", "--out", out)

	// Assert
	if err == nil {
		t.Fatal("expected an error when --out is passed alongside a single-file target, got nil")
	}
}

func TestCompileDirectory_ManyFilesConcurrently(t *testing.T) {
	// Arrange
	root := t.TempDir()
	const n = 50
	for i := range n {
		writeFile(t, root, fmt.Sprintf("f%02d.lex.md", i), validLexMD)
	}
	out := t.TempDir()

	// Act
	stdout, err := runCompile(t, root, "--to", "json", "--out", out, "--workers", "4")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, stdout)
	}
	for i := range n {
		path := filepath.Join(out, fmt.Sprintf("f%02d.json", i))
		if _, statErr := os.Stat(path); statErr != nil {
			t.Errorf("expected output file %s: %v", path, statErr)
		}
	}
}
