# Editor integration

`lexicon check` already produces exactly the diagnostics an editor wants — file, severity, and message — one line per finding. This wires that output into your editor's inline diagnostics without Lexicon shipping a bespoke LSP server: both setups below run `lexicon check` as an external tool and translate its output into diagnostics.

This is deliberately the cheapest way to get editor feedback, not the final one — see the "LSP/editor tooling" line in the [README's Roadmap](../README.md#roadmap--explicitly-deferred). If it turns out people actually want this, a real `lexicon` LSP server (most likely built on `golang.org/x/tools`' `jsonrpc2`/`protocol` packages) is the natural next step.

## Known limitations

Read this before wiring anything up — it explains why diagnostics won't always land on the exact offending line:

- **No line/column information.** `check`'s findings ([`internal/ports/linter.go`](../internal/ports/linter.go)) carry only `Severity`, `Message`, and an optional `Scenario` name — no source position. Both setups below anchor diagnostics to line 1 of the file (or, for the Neovim setup, whatever line nvim-lint defaults to when no line is captured) rather than the actual offending line.
- **Scenario-qualified findings don't resolve for `%f`-based tools.** When a finding includes a scenario, `check.go` prints the path as `<path> (<Scenario Name>)` before the colon. `efm-langserver`'s `%f` pattern expects a bare file path there, so those specific diagnostics won't show up under the VSCode/efm setup — only file-level findings (e.g. "no Feature heading found") are guaranteed to surface. The Neovim setup is unaffected, since it never tries to parse a file path out of the output.

Fixing either limitation for real means `check` emitting actual line numbers, which needs source-position tracking added to the parser and linter — a separate piece of work.

## Neovim — `nvim-lint`

Requires [`nvim-lint`](https://github.com/mfussenegger/nvim-lint) and a `lexicon` binary on your `$PATH`.

```lua
-- Give *.lex.md files a distinct filetype so the lexicon linter runs only
-- on them, not on every markdown file.
vim.filetype.add({
  pattern = {
    ['.*%.lex%.md'] = 'markdown.lexicon',
  },
})

local severities = {
  error = vim.diagnostic.severity.ERROR,
  warning = vim.diagnostic.severity.WARN,
}

require('lint').linters.lexicon = {
  cmd = 'lexicon',
  args = { 'check' },
  -- stdin defaults to false, so nvim-lint automatically appends the
  -- current buffer's filename as the last argument.
  stream = 'stdout',
  ignore_exitcode = true, -- `check` exits 1 when it reports an error-severity finding
  parser = require('lint.parser').from_pattern(
    '^[^:]+:%s*(%a+):%s*(.+)$',
    { 'severity', 'message' },
    severities,
    { severity = vim.diagnostic.severity.WARN }
  ),
}

require('lint').linters_by_ft.lexicon = { 'lexicon' }

vim.api.nvim_create_autocmd({ 'BufWritePost' }, {
  callback = function() require('lint').try_lint() end,
})
```

Save a `.lex.md` file and diagnostics should appear via `vim.diagnostic`.

## VSCode — `efm-langserver` + Generic LSP Client

VSCode has no built-in generic LSP client, so this setup has two parts: [`efm-langserver`](https://github.com/mattn/efm-langserver) does the actual work of wrapping `lexicon check` as a real LSP server, and [`Generic LSP Client`](https://marketplace.visualstudio.com/items?itemName=llllvvuu.llllvvuu-glspc) is the thin VSCode extension that lets VSCode talk to it.

1. **Install `efm-langserver`:**

   ```bash
   go install github.com/mattn/efm-langserver@latest
   # or: brew install efm-langserver
   ```

2. **Install the [`Generic LSP Client`](https://marketplace.visualstudio.com/items?itemName=llllvvuu.llllvvuu-glspc) extension** from the VSCode Marketplace.

3. **Create `efm-langserver`'s config** at `$HOME/.config/efm-langserver/config.yaml` (or `%APPDATA%\efm-langserver\config.yaml` on Windows) — this is normalizing `error`/`warning` to the single-character codes vim errorformat's `%t` expects:

   ```yaml
   version: 2
   languages:
     lexicon:
       lint-command: "lexicon check ${INPUT} | sed -e 's/: error: /: E: /' -e 's/: warning: /: W: /'"
       lint-formats:
         - '%f: %t: %m'
       lint-ignore-exit-code: true
       lint-source: lexicon
   ```

4. **In VSCode's `settings.json`**, tag `.lex.md` files with a language id and point the Generic LSP Client at `efm-langserver`:

   ```json
   {
     "files.associations": { "*.lex.md": "lexicon" },
     "glspc.languageId": "lexicon",
     "glspc.serverCommand": "efm-langserver"
   }
   ```

   If your `config.yaml` isn't at the default location, add `"glspc.serverCommandArguments": ["-c", "/absolute/path/to/config.yaml"]`.

Reload VSCode, open a `.lex.md` file, and diagnostics should appear in the Problems panel.

## Other editors

`efm-langserver`'s own README documents client setup for [vim-lsp](https://github.com/mattn/efm-langserver#configuration-for-vim-lsp), [coc.nvim](https://github.com/mattn/efm-langserver#configuration-for-cocnvim), [Eglot](https://github.com/mattn/efm-langserver#configuration-for-eglot-emacs), [Helix](https://github.com/mattn/efm-langserver#configuration-for-helix), and [Sublime Text LSP](https://github.com/mattn/efm-langserver#configuration-for-sublimetext-lsp) — all of them reuse the same `config.yaml` `languages.lexicon` block from the VSCode setup above; only the client-side wiring differs.
