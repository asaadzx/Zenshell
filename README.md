# BakShell

A blazing-fast, customizable shell with Lua plugin support — rewritten in Go.

## Quickstart

```sh
go build -ldflags="-s -w" -o bsh ./cmd/bsh
./bsh
```

Single static binary, ~3MB stripped. No runtime deps.

## Features

- **Lua config** — theme colors, prompt format, plugin selection at `~/.zencr/config.lua`
- **Lua plugins** — overload `execute_command` and `get_prompt` from Lua scripts
- **Aquia theme** — beautiful two-line prompt with git status, exit code, and Aquia color palette
- **Readline input** — arrow-key history, line editing, history persistence
- **No libreadline / liblua** — fully static, no system library dependencies

## Configuration

```lua
-- ~/.zencr/config.lua
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

`%u` → user, `%h` → hostname, `%d` → cwd (with `~` for home).

## Plugins

Lua scripts in `~/.zencr/plugins/`. Each plugin can define:

```lua
function execute_command(args)
    if args[1] == "hello" then
        print("Hello, World!")
        return true  -- command handled
    end
    return false     -- pass to shell
end

function get_prompt()
    return "❯ "      -- custom prompt (overrides theme prompt_format)
end

function set_exit_code(code)
    -- called after every command with the exit code
end
```

## Development

```sh
go build ./cmd/bsh && go vet ./...
```

Push a `v*` tag to trigger CI — builds `.tar.gz`, `.deb`, `.rpm`, `.tar.zst` for Linux and macOS via GoReleaser.

## License

MIT
