-- Default configuration for Zen Shell

-- List of active plugins
plugins = {
    "aquia-prompt.lua",
    "autosuggest.lua",
    "syntax-highlighting.lua",
    "git-prompt.lua"
}

-- Theme settings
theme = {
    prompt_color = "#4287f5",     -- Blue prompt (hex format)
    background = "#000000",       -- Black background (hex format)
    prompt_format = "[%u@%h %d]$ " -- Format: [username@hostname directory]$
}

-- Custom shell settings
settings = {
    history_size = 1000,
    auto_complete = true
} 