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
internal/shell/config.go     — Lua config parser (Config, ThemeConfig, SettingsConfig)
internal/shell/plugins.go    — Lua plugin loader + cached call wrappers
internal/shell/prompt.go     — hex_to_ansi, format_prompt with %u/%h/%d
plugins/                     — Lua plugin scripts
.goreleaser.yaml             — GoReleaser config for cross-platform releases
.github/workflows/release.yml — CI: builds deb/rpm/tar.gz/tar.zst for linux + macos
```

## Feature details

### Tokenizer (`tokenize.go`)
- Splits input into tokens at whitespace AND shell metacharacters (`;`, `|`, `&`, `>`, `<`)
- Single quotes `'...'` — literal, no expansion
- Double quotes `"..."` — `$VAR` expansion, `\\` escape
- `$VAR` / `${VAR}` expansion via `os.Getenv`
- `~` at token start expands to `$HOME`
- Multi-char operators produced as single tokens: `&&`, `||`, `>>`, `&>`, `&>>`

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
- Command completion from a cached scan of `$PATH` (built once at first tab press)
- File path completion with `~` expansion
- Directories get a trailing `/`

### Command timing
- Commands taking >100ms print a dimmed duration to stderr

## CI / Release

Push a `v*` tag to trigger `.github/workflows/release.yml`. GoReleaser cross-builds for:
- **Linux**: `amd64`, `arm64` → `.tar.gz`, `.deb`, `.rpm`, `.tar.zst`
- **macOS**: `amd64`, `arm64` → `.tar.gz`

Creates a draft GitHub Release with all artifacts.

## Verification

```sh
go build ./cmd/bsh
go vet ./...
```

No tests exist yet.
