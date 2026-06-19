<p align="center">
  <img src="docs/logo/BaklavaShellLogo.png" alt="Baklava Shell" width="300"/>
</p>

<p align="center">
  <em>Ba</em>klava <em>Sh</em>ell — like layered phyllo, every command is a layer of perfection.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go" alt="Go version"/>
  <img src="https://img.shields.io/badge/license-GPLv3-blue" alt="License"/>
  <img src="https://img.shields.io/badge/build-static-brightgreen" alt="Build"/>
  <img src="https://img.shields.io/badge/baklava-sweet-ff69b4" alt="Baklava"/>
</p>

<p align="center">
  A blazing-fast, customizable shell with Lua plugin support — rewritten in Go from the original C++ Zen Shell.<br/>
  Single static binary, ~3 MB stripped. No libreadline, no liblua, no CGo.
</p>

<p align="center">
  <img src="docs/logo/screenshot.png" alt="BakShell screenshot" width="80%"/>
</p>

---

## Quickstart

```sh
go build -ldflags="-s -w" -o bsh ./cmd/bsh
./bsh
```

## Features

| Category | What you get |
|----------|-------------|
| **Lua config** | Theme colors, prompt format, plugin selection at `~/.bshc/config.lua` |
| **Lua plugins** | Hook into `execute_command`, `get_prompt`, `set_exit_code` from Lua |
| **Prompt theming** | Aquia theme with git status, exit code, rich color palette — or build your own |
| **Readline input** | Arrow-key history, line editing, persistent `~/.bshc/history` |
| **Fully static** | Zero runtime deps. No libreadline, no liblua, no CGo. Just a binary. |
| **Builtins** | `cd`, `exit`, `echo`, `pwd`, `type`, `export`, `unset`, `history`, `alias`, `unalias`, `help`, `banner` |
| **Data pipeline** | `from-json`, `from-csv`, `to-json`, `to-csv`, `where`, `sort-by`, `select`, `first`, `last`, `count`, `uniq`, `confirm`, `trash`, `undo` |
| **Scripting** | `if`/`else`/`end`, `for`/`end`, `while`/`end`, `source`, `[ cond ]` tests, variables |
| **Tab completion** | PATH-aware command completion + filesystem path completion |
| **Command timing** | Auto-timed commands (dimmed duration for anything >100 ms) |
| **Aliases** | `alias name=value` with recursive expansion |
| **Multi-line input** | Continuation prompts for unclosed quotes and `\` continuations |

## Configuration

```lua
-- ~/.bshc/config.lua
plugins = {
    "aquia-prompt.lua",
    "autosuggest.lua",
}

theme = {
    prompt_color = "#4287f5",
    background   = "#000000",
    prompt_format = "[%u@%h %d]$ "
}

settings = {
    history_size = 1000,
    auto_complete = true
}
```

| Specifier | Expands to |
|-----------|------------|
| `%u` | Username |
| `%h` | Hostname |
| `%d` | Current directory (with `~` for home) |
| `%t` / `%T` | Time (HH:MM / HH:MM:SS) |
| `%?` | Last exit code |
| `%$` | `#` for root, `$` otherwise |

## Plugins

Lua scripts in `~/.bshc/plugins/` can hook into four entrypoints:

```lua
function execute_command(args)
    if args[1] == "hello" then
        print("Hello, World!")
        return true  -- handled, don't pass to shell
    end
    return false     -- not handled, pass to shell
end

function get_prompt()
    return "❯ "      -- custom prompt (overrides theme prompt_format)
end

function set_exit_code(code)
    -- called after every command with the exit code
end

function get_suggestion(line)
    -- called on every keystroke; return a completion suffix
    -- from history for inline autosuggest (shown dimmed on screen)
    for _, cmd in ipairs(history) do
        if cmd:sub(1, #line) == line then
            return cmd:sub(#line + 1)
        end
    end
    return ""
end
```

### Included plugins

| Plugin | Description |
|--------|-------------|
| `aquia-prompt.lua` | Two-line prompt with git, exit code, Aquia palette |
| `git-prompt.lua` | Git branch + status in prompt |
| `autosuggest.lua` | History-based inline suggestions via `get_suggestion` hook |
| `powerlevel10k.lua` | Full-featured p10k-style prompt theme |
| `venv-prompt.lua` | Python virtualenv/conda indicator |
| `node-version.lua` | Node.js version from `.nvmrc`/`.node-version` |
| `command-timer.lua` | Elapsed time for slow commands |
| `todo.lua` | Simple todo list manager |
| `jump.lua` | Frecency-based directory jumping |
| `quote.lua` | Random developer quotes in prompt |
| `proxy.lua` | Auto proxy based on network patterns |
| `syntax-highlighting.lua` | Command syntax highlighting |

## Directory layout

```
~/.bshc/
├── config.lua         -- Shell configuration (theme, plugins, settings)
├── plugins/           -- Lua plugin scripts
├── history            -- Command history (auto-managed)
├── todos.json         -- Todo plugin data
└── jump.db            -- Jump plugin frecency database
```

## Roadmap / TODO

### High priority
- [x] **Branding**: logo assets and screenshot added to README
- [x] **Suggestion plugin**: wire up `get_suggestion` hook in Go for inline suggestions

### Medium priority
- [x] **Cleanup**: removed duplicate `ghost-prompt.lua` (identical to `powerlevel10k.lua`)
- [x] **Document `.bshc`**: directory layout documented above
- [x] **Plugin dev guide**: included in README
- [x] **Refactor plugins**: removed dead hooks not called by the shell

### Low priority
- [ ] **New plugins**: fzf integration, zoxide-style dir nav, weather, motd
- [x] **Test coverage**: tests for config, plugins, and cmd/bsh packages

## Development

```sh
go build ./cmd/bsh && go vet ./...
go test ./...
```

Push a `v*` tag to trigger CI — builds `.tar.gz`, `.deb`, `.rpm`, `.tar.zst` for Linux and macOS via GoReleaser.

## License

This project is licensed under the **GNU General Public License v3.0** — see the [LICENSE](LICENSE) file for details.

GNU GPLv3 means:
- ✅ Free to use, modify, and distribute
- ✅ Must preserve license and copyright notices
- ✅ Derivative works must also be GPLv3
- ✅ Source code must be made available to users

For more information, see [https://www.gnu.org/licenses/gpl-3.0.html](https://www.gnu.org/licenses/gpl-3.0.html)
