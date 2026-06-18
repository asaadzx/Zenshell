-- Node.js version prompt indicator
-- Shows the active Node.js version from .nvmrc or .node-version

local ESC = "\27["
local RESET = ESC .. "0m"
local GREEN = ESC .. "38;2;0;200;83m"
local YELLOW = ESC .. "38;2;255;202;40m"

local node_version = ""
local last_pwd = ""
local cache_ts = 0
local CACHE_TTL = 5

local function read_file(path)
  local f = io.open(path, "r")
  if not f then return nil end
  local content = f:read("*l")
  f:close()
  return content
end

local function trim(s)
  return (s:gsub("^%s+", ""):gsub("%s+$", ""))
end

local function find_version_file()
  local pwd = io.popen("pwd"):read("*l") or ""
  if pwd == "" then return nil end

  local search = pwd
  while search and #search > 0 do
    local nvmrc = read_file(search .. "/.nvmrc")
    if nvmrc then return trim(nvmrc) end

    local nodever = read_file(search .. "/.node-version")
    if nodever then return trim(nodever) end

    local parent = search:match("^(.+)/[^/]+$")
    search = parent
  end
  return nil
end

local function detect_node()
  local now = os.time()
  local pwd = io.popen("pwd"):read("*l") or ""
  if now - cache_ts < CACHE_TTL and pwd == last_pwd then
    return
  end
  last_pwd = pwd
  cache_ts = now

  node_version = find_version_file() or ""
end

function get_prompt_suffix()
  detect_node()
  if node_version == "" then return "" end
  return GREEN .. "  " .. node_version .. RESET
end

function execute_command(args)
  return false
end

print("Node version prompt plugin loaded")
