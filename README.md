<div align="center">

# lexicon

**A markdown-native requirements DSL, translatable to Gauge, Gherkin, and beyond.**

[![CI](https://github.com/jedi-knights/lexicon/actions/workflows/ci.yml/badge.svg)](https://github.com/jedi-knights/lexicon/actions/workflows/ci.yml)
[![Release](https://github.com/jedi-knights/lexicon/actions/workflows/release.yml/badge.svg)](https://github.com/jedi-knights/lexicon/actions/workflows/release.yml)
[![GoReleaser](https://github.com/jedi-knights/lexicon/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/jedi-knights/lexicon/actions/workflows/goreleaser.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[Why](#why-lexicon) · [The format](#the-lexmd-format) · [Installation](#installation) · [Usage](#usage) · [GitHub Action](#github-action) · [Verification](#verification) · [Roadmap](#roadmap--explicitly-deferred) · [Development](#development)

</div>

---

Dev, QA, and Product routinely read the same requirement and walk away with three different mental models of what's being built — not from carelessness, but because prose leaves three things implicit: the **precondition**, the **action**, and the **outcome**. Gherkin's Given/When/Then is one well-known way to force those three things into the open. Lexicon generalizes the idea: a small, memorable, plain-markdown structure that stays readable on GitHub and Slack with zero special tooling, and compiles deterministically into Gauge, Gherkin, or a schema-stable JSON structure built for LLM consumption.

## Why lexicon

- **It's just markdown.** A `.lex.md` file is valid CommonMark/GFM — headings, bullet lists, tables, fenced code blocks. It renders correctly with no plugin, on GitHub, in a PR diff, or pasted into Slack.
- **It's small on purpose.** One heading pattern (`# Feature`, `## Scenario`), one step pattern (`- **Precondition** ...`). Nothing else to memorize.
- **The keyword and the concept are separate.** Every step carries both a surface `Keyword` and a resolved `Role` (`precondition`/`action`/`outcome`) — the actual thing that causes Dev/QA/Product to diverge. Write the `Keyword` in whichever of Lexicon's two dialects you prefer — `Precondition`/`Action`/`Outcome`/`And`/`But` (dialect-neutral, matching `Role`'s own names, used throughout this README) or `Given`/`When`/`Then`/`And`/`But` (widely recognized; used by Cucumber, Behave, SpecFlow, and other BDD tools) — both resolve to the same `Role`. An LLM (or a human) reading the JSON output gets the concept directly either way.
- **Targets are an open set, not a hardcoded pair.** Gauge, Gherkin, and Robot Framework ship today, but adding a new target later (an in-house format, whatever comes next) means writing one new `Emitter` implementation — nothing else in Lexicon changes. See [`internal/adapters/emitter/registry.go`](internal/adapters/emitter/registry.go).

## The `.lex.md` format

```markdown
---
id: SEARCH-042
owner: product
status: draft
---

# Feature: Search results pagination

As a user, I want to page through search results so I can browse more than the first page of matches.

## Background

- **Precondition** the catalog has more than one page of results

@search @pagination
## Scenario: Navigating to the next page

- **Precondition** the user has performed a search that returns more than one page of results
- **Action** the user clicks the "Next" pagination control
- **Outcome** the second page of results replaces the first page
- **And** the pagination control shows page 2 as active

@search @pagination
## Scenario: Jumping to a specific page

- **Precondition** the user has performed a search that returns more than one page of results
- **Action** the user clicks the page `<page>` control
- **Outcome** the `<page>` page of results replaces the current page

| page |
| ---- |
| 2    |
| 3    |
```

- **One `# Feature: <name>` per file**, with an optional free-text description paragraph directly beneath it.
- **One or more `## Scenario: <name>` blocks.** A `## Background` block (no `Scenario:` prefix) runs before every other scenario.
- **Steps are bullets with a bold leading keyword**, in either of two dialects: `- **Precondition** ...` / `**Action**` / `**Outcome**` / `**And**` / `**But**` (dialect-neutral, used above), or `- **Given** ...` / `**When**` / `**Then**` / `**And**` / `**But**` (widely recognized BDD phrasing). Both resolve to the same `Role` and can be mixed freely across scenarios in the same file. Plain CommonMark; renders as a normal list everywhere. For example, the Background step above is equally valid written in the BDD dialect as:

  ```markdown
  - **Given** the catalog has more than one page of results
  ```
- **Tags** are a bare `@tag1 @tag2` line directly above a heading — the only tagging mechanism (front matter is never used for tags, to avoid two parallel ways to do the same thing).
- **Examples/outline data** is a plain GFM table placed right after a scenario's steps. Its presence — combined with `` `<name>` `` placeholders in step text — implies outline behavior; no separate `Scenario Outline`/`Examples:` heading is needed, since modern Gherkin (v6+) doesn't require one and Gauge never did.
- **Placeholders must be backtick-escaped** (`` `<page>` ``, not bare `<page>`). This isn't stylistic: a bare `<name>` is inline raw HTML in CommonMark, and GitHub's renderer silently **drops** unrecognized tags — a bare placeholder in an Examples row vanishes when viewed on GitHub. Verified against GitHub's real markdown API, not just goldmark's opinion — see [Verification](#verification).
- **Doc strings** are a fenced code block indented two spaces under a step bullet, so CommonMark parses it as part of that list item rather than ending the list.
- **Front matter** (`---`-delimited YAML) is for bookkeeping with no BDD equivalent — `id`, `owner`, `status`, `links`. Anything else you add lands in a catch-all `extra` map rather than being silently dropped.

## Installation

### Homebrew (macOS, Linux)

```bash
brew install jedi-knights/tap/lexicon
```

Pulls a prebuilt binary from the [`jedi-knights/homebrew-tap`](https://github.com/jedi-knights/homebrew-tap) — no Go toolchain required. The formula is regenerated by GoReleaser on every release.

### Go install

Requires Go 1.23 or later.

```bash
# Latest release
go install github.com/jedi-knights/lexicon/cmd/lexicon@latest

# Pinned to a specific version
go install github.com/jedi-knights/lexicon/cmd/lexicon@v0.1.0
```

`go install` places the binary in `$GOBIN` if set, otherwise `$GOPATH/bin`.

## Usage

```bash
# List every registered translation target
$ lexicon compile --list-targets
gauge
gherkin
json
robot

# Translate to Gauge
$ lexicon compile --to gauge requirements/pagination.lex.md

# Translate to Gherkin
$ lexicon compile --to gherkin requirements/pagination.lex.md

# Translate to Robot Framework
$ lexicon compile --to robot requirements/pagination.lex.md

# Translate to schema-stable JSON (role/keyword both present, for LLM consumption)
$ lexicon compile --to json requirements/pagination.lex.md

# Structural lint: missing outcomes, empty scenarios
$ lexicon check requirements/pagination.lex.md
```

### Directory compilation

Point `compile` at a directory (or nothing, to default to the current directory) and it recursively compiles every `.lex.md` file it finds, mirroring the input tree under `--out`:

```bash
# Compile every .lex.md under requirements/ to Gherkin, mirroring structure under gherkin-out/
lexicon compile requirements --to gherkin --out gherkin-out

# Default to the current directory, cap concurrency at 4 workers
lexicon compile --to json --out json-out --workers 4
```

Files compile concurrently (default: one per CPU, override with `--workers`). Every matching file is attempted even if others fail — a failing file's error is printed by relative path and the command exits non-zero, but files that succeeded are still written.

## GitHub Action

```yaml
- uses: jedi-knights/lexicon@v1
  with:
    path: "requirements/*.lex.md"
```

Wraps `lexicon check` in CI. See [`action.yml`](action.yml).

## Verification

Lexicon's core promise — "this is valid, renderable markdown" — is a structural property, not a textual one, so it's fronted by a real CommonMark/GFM engine ([`goldmark`](https://github.com/yuin/goldmark), the same engine under Hugo) rather than a hand-rolled line scanner. Verification is intentionally **asymmetric** across targets:

- **Gauge fidelity is not equally verifiable.** Gauge's Go parser/validation packages are internal to the `gauge` CLI binary, not a stable public library, and pull in gRPC runner infrastructure. Lexicon does not depend on them. Gauge output is verified against hand-authored expected fixtures for v1; a real round-trip through the actual `gauge` binary is a deferred, tracked improvement (see below), not a silent gap.
- **Gherkin fidelity is a real, cheap CI gate**: every emitted `.feature` golden fixture is fed through the real [`cucumber/gherkin`](https://github.com/cucumber/gherkin) library (`ParseGherkinDocument`) and asserted to parse with no error — see [`internal/adapters/emitter/golden_test.go`](internal/adapters/emitter/golden_test.go).
- **Robot Framework fidelity is not equally verifiable**, for the same reason as Gauge: a real parse-and-validate round trip would mean adding the `robotframework` Python package as a test-time dependency to an otherwise pure-Go CI pipeline. `.robot` output is verified against hand-authored expected fixtures for v1, checked by hand against the [Robot Framework User Guide](https://robotframework.org/robotframework/latest/RobotFrameworkUserGuide.html)'s documented syntax — not a silent gap, see the roadmap below.

## Roadmap / explicitly deferred

Named here on purpose, rather than discovered as a surprise:

- Gherkin's `Rule` keyword (no Gauge equivalent at all).
- Gauge Concepts (`.cpt` reusable step groups — Gauge-only, no Gherkin equivalent).
- Gauge teardown (`___`) blocks.
- Step-level data tables (Gherkin's *other* table use, distinct from Examples).
- Gherkin's `# language:` dialect header — English only for now.
- A real round-trip smoke test against the actual `gauge` binary.
- A real parse-and-validate round trip for Robot Framework output via the `robotframework` Python package.
- A real `lexicon` LSP server. In the meantime, see [Editor integration](docs/editor-integration.md) for wiring `lexicon check` into Neovim and VSCode via existing generic-linter tooling.

## Development

```bash
make build   # build ./lexicon
make test    # go test ./...
make lint    # golangci-lint run ./...
make run     # build + lexicon compile --list-targets
```

## Contributing

Issues and PRs welcome. Adding a new translation target is the easiest way to contribute meaningfully: implement `ports.Emitter` (`Format() string`, `Emit(w io.Writer, doc *domain.Document) error`) in a new file under `internal/adapters/emitter/`, register it in an `init()`, and add a `testdata/<case>/expected.<ext>` golden fixture.

## License

[MIT](LICENSE)
