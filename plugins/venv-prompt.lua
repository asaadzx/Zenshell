-- Virtualenv / Conda prompt indicator
-- Shows active Python virtualenv or conda environment in the prompt

local ESC = "\27["
local RESET = ESC .. "0m"
local CYAN = ESC .. "38;2;0;188;212m"
local GREEN = ESC .. "38;2;0;200;83m"
local YELLOW = ESC .. "38;2;255;202;40m"

local active_env = ""
local last_pwd = ""
local cache_ts = 0
local CACHE_TTL = 2

local function detect_env()
  local now = os.time()
  if now - cache_ts < CACHE_TTL and last_pwd ~= "" then
    return
  end

  last_pwd = io.popen("pwd"):read("*l") or ""
  cache_ts = now

  -- Check VIRTUAL_ENV (most common)
  local venv = os.getenv("VIRTUAL_ENV")
  if venv then
    local name = venv:match("/([^/]+)$")
    if name then
      active_env = name
      return
    end
  end

  -- Check CONDA_DEFAULT_ENV
  local conda = os.getenv("CONDA_DEFAULT_ENV")
  if conda then
    active_env = conda
    return
  end

  -- Check .venv directory in current or parent dirs
  local pwd = last_pwd
  while pwd and #pwd > 0 do
    local handle = io.open(pwd .. "/.venv/pyvenv.cfg")
    if handle then
      handle:close()
      active_env = pwd:match("/([^/]+)$") or "venv"
      return
    end
    local parent = pwd:match("^(.+)/[^/]+$")
    pwd = parent
  end

  active_env = ""
end

function get_prompt_suffix()
  detect_env()
  if active_env == "" then return "" end
  return CYAN .. "(" .. GREEN .. active_env .. CYAN .. ")" .. RESET
end

function execute_command(args)
  return false
end

print("Virtualenv prompt plugin loaded")
