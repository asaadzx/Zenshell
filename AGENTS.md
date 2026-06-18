# BakShell — AGENTS.md

## Project status

Baklava Shell (`bsh`) is a Go shell being rewritten from the original C++ Zen Shell. The Go rewrite is **functional** — it handles the main loop, Lua config, Lua plugins, prompt theming, history persistence (via `chzyer/readline`), built-in commands (`cd`, `exit`/`quit`), external command execution via `os/exec`, and SIGINT handling.

The old C++ code has been removed. This is a pure Go project.

## Architecture (Go)

- **Module**: `bakshell` (root package `main`)
- **Entrypoint**: `main.go` → `Shell.Run()` in `shell.go`
- **Config**: `~/.zencr/config.lua` — parsed via `gopher-lua` into `Config` struct
- **Plugins**: Lua scripts in `~/.zencr/plugins/` — `execute_command(args)` and `get_prompt()` entrypoints cached at load time
- **Prompt theming**: `prompt.go` — `hex_to_ansi` + `strings.ReplaceAll` for `%u`, `%h`, `%d` (no regex)
- **History**: automatically managed by `chzyer/readline` → `~/.zencr/history`
- **Input**: `chzyer/readline` — arrow key history, line editing, EOF/^C handling

## Build & run

```sh
go build -ldflags="-s -w" -o bsh ./cmd/bsh
./bsh
```

Single static binary, ~3MB stripped. No runtime deps (no libreadline, no liblua).

## Dependencies (Go)

- `github.com/yuin/gopher-lua` — pure-Go Lua VM (no CGo, no liblua)
- `github.com/chzyer/readline` — line editing with history

## Key optimizations over C++ version

1. **No regex in hot path** — prompt formatting uses `strings.ReplaceAll`
2. **Cached Lua functions** — `execute_command` and `get_prompt` are looked up once after plugin loading, not every iteration
3. **No libreadline/liblu5.3** — fully static, no system library deps
4. **History auto-persisted** — `chzyer/readline` writes to file on each line

## Source layout

```
cmd/bsh/main.go              — entry point
internal/shell/shell.go      — Shell struct, main loop, group-based chaining (&&/||/;/&), timing
internal/shell/completer.go  — Tab completion (PATH cache + file paths)
internal/shell/tokenize.go   — Tokenizer: quoting, $VAR, ~, metachar splitting (;/|/&/></)
internal/shell/pipeline.go   — Pipeline execution (|), redirections (>/>
/<</2>/2>&1/&>), builtins
internal/shell/tokenize_test.go  — Tokenizer tests
internal/shell/pipeline_test.go  — Pipeline/builtin tests
internal/config/config.go    — Lua config parser (Config, ThemeConfig, SettingsConfig)
internal/plugins/plugins.go  — Lua plugin loader + cached call wrappers
internal/prompt/prompt.go    — hex_to_ansi, format_prompt with %u/%h/%d/%t/%T/%?/%$
internal/prompt/prompt_test.go — Prompt tests
internal/data/value.go       — Structured data type system (String/Int/Float/Bool/List/Record/Table)
internal/data/table.go       — Table operations (filter, sort, select, parse condition)
internal/data/value_test.go  — Value type tests
internal/data/table_test.go  — Table operation tests
plugins/                     — Lua plugin scripts
.goreleaser.yaml             — GoReleaser config for cross-platform releases
.github/workflows/release.yml — CI: builds deb/rpm/tar.gz/tar.zst for linux + macos
```

## Feature details

### Tokenizer (`tokenize.go`)
- Splits input into tokens at whitespace AND shell metacharacters (`;`, `|`, `&`, `>`, `<`)
- Newlines (`\n`) treated as whitespace for multi-line input
- Single quotes `'...'` — literal, no expansion
- Double quotes `"..."` — `$VAR` expansion, `\\` escape
- `$VAR` / `${VAR}` expansion via `os.Getenv`
- `~` at token start expands to `$HOME`
- Multi-char operators produced as single tokens: `&&`, `||`, `>>`, `&>`, `&>>`
- `needsContinuation()` detects unclosed quotes/trailing `\` for multi-line input

### Pipeline & redirections (`pipeline.go`)
- `cmd1 | cmd2` — stdout of cmd1 piped to stdin of cmd2
- `cmd > file` / `cmd >> file` — stdout redirect (truncate/append)
- `cmd < file` — stdin redirect
- `cmd 2> file` / `cmd 2>> file` — stderr redirect
- `cmd 2>&1` — stderr to stdout
- `cmd 1>&2` — stdout to stderr
- `cmd &> file` / `cmd &>> file` — both stdout+stderr to file
- Builtins (`cd`, `exit`/`quit`) handled directly; everything else through `os/exec`
- Lua plugin `execute_command` hook checked before external execution

### Chaining (`shell.go`: `parseGroups` + `execute`)
- `A ; B` — run A, then B unconditionally
- `A && B` — run B only if A exits 0
- `A || B` — run B only if A exits non-zero
- `A &` — run A in background (via `connBg` skip logic)
- Proper short-circuit semantics: `A && B || C` works like bash
- Pipelines work within groups: `A && B | C || D`

### Tab completion (`completer.go`)
- Implements `readline.AutoCompleter` interface
- Command completion from a cached scan of `$PATH` (built once at first tab press, **refreshed when `$PATH` changes**)
- File path completion with `~` expansion
- Directories get a trailing `/`

### Command timing
- Commands taking >100ms print a dimmed duration to stderr

### Built-in commands
| Command | Description |
|---------|-------------|
| `cd` | Change directory (defaults to `$HOME`) |
| `exit` / `quit` | Exit the shell |
| `echo` | Display a line of text (`-n` suppresses newline) |
| `pwd` | Print working directory |
| `type` | Show whether a command is builtin, external, or not found |
| `export` | Set environment variable (`export NAME=value`) |
| `unset` | Unset environment variable |
| `history` | Show command history (`history -c` to clear, `history N` for last N) |
| `alias` | Define or display aliases |
| `unalias` | Remove alias definitions (`unalias -a` removes all) |
| `help` | List built-in commands with descriptions |

### Alias expansion
- Aliases defined via `alias name=value` are expanded **once** at the start of each command segment (after `|`, `;`, `&&`, `||`, `&`)
- Expanded by re-tokenizing the alias value

### Multi-line input
- Lines ending with unescaped `\` or containing unclosed quotes prompt for continuation with `> `
- Newlines in continuation lines are treated as whitespace by the tokenizer

### Prompt specifiers
| Specifier | Expands to |
|-----------|------------|
| `%u` | Username |
| `%h` | Hostname |
| `%d` | Current directory (with `~` for home) |
| `%t` | Current time (HH:MM) |
| `%T` | Current time (HH:MM:SS) |
| `%?` | Last exit code |
| `%$` | `#` for root, `$` otherwise |

## Feature details (new)

### Structured data pipeline
- **Value types**: `StringValue`, `IntValue`, `FloatValue`, `BoolValue`, `ListValue`, `RecordValue`, `TableValue`
- **Comparisons**: `==`, `!=`, `<`, `<=`, `>`, `>=`, `~=` (regex), `in` (list membership)
- **Table operations**: `Filter(conditions)`, `SortBy(field, desc)`, `Select(columns)`
- **Condition parsing**: `ParseCondition("field op value")` for `where`-style filters
- `ParseValue(s)` auto-detects int/float/bool/string

### Data files
- `internal/data/value.go` — Value interface + all implementations
- `internal/data/table.go` — Table with filter/sort/select + condition parser
- `internal/data/format.go` — JSON/CSV serialization (ToJSON, FromJSON, ToCSV, FromCSV)

### Data pipeline (shell integration)
- Data commands: `from-json`, `from-csv`, `to-json`, `to-csv`, `where`, `sort-by`, `select`
- JSON is the internal interchange format between data commands in a pipeline
- Data commands can be chained: `from-csv data.csv | where age > 30 | sort-by name | to-json`
- Mixed pipelines: `external-cmd | from-json | where ...` or `from-json ... | grep ...`
- Data commands run sequentially within a pipeline (not concurrently like external processes)
- Pipeline detection: `execPipeline` checks for `isData` flag on commands and routes to `execDataPipeline`

### Data command reference
| Command | Description |
|---------|-------------|
| `from-json [file]` | Parse JSON from file or stdin into data table |
| `from-csv [file]` | Parse CSV from file or stdin into data table |
| `to-json` | Convert data table to JSON output |
| `to-csv` | Convert data table to CSV output |
| `where <condition...>` | Filter rows by conditions (`"age > 30"`) |
| `sort-by <field> [--desc]` | Sort rows by a field |
| `select <column...>` | Keep only specified columns |
| `first [n]` | Show first n rows (default 10) |
| `last [n]` | Show last n rows (default 10) |
| `count [field...]` | Count rows, or count by groups |
| `uniq [field...]` | Show unique rows |
| `confirm [message]` | Prompt yes/no, exits 0 for yes |
| `trash <file...>` | Move files to ~/.zencr/trash/ |
| `undo` | Restore previous table state |

### Pipeline files
- `internal/shell/pipeline.go` — `command` struct with `isData` flag, data handlers, `execDataPipeline`

## CI / Release

Push a `v*` tag to trigger `.github/workflows/release.yml`. GoReleaser cross-builds for:
- **Linux**: `amd64`, `arm64` → `.tar.gz`, `.deb`, `.rpm`, `.tar.zst`
- **macOS**: `amd64`, `arm64` → `.tar.gz`

Creates a draft GitHub Release with all artifacts.

## Verification

```sh
go build ./cmd/bsh
go vet ./...
go test ./...
```

Test files exist for `internal/shell` (tokenizer, pipeline) and `internal/prompt`.
