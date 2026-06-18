-- Aquia theme prompt for BakShell
-- Optimized: cached lookups, single git call, table.concat builder

local ESC = "\27["
local RESET = ESC .. "0m"
local BOLD = ESC .. "1m"

-- Aquia color palette (pre-computed ANSI)
local C = {
  aqua    = ESC .. "38;2;0;188;212m",
  aqua_bg = ESC .. "48;2;0;188;212m",
  teal    = ESC .. "38;2;0;150;136m",
  cyan    = ESC .. "38;2;128;222;234m",
  white   = ESC .. "38;2;238;238;238m",
  gray    = ESC .. "38;2;120;120;120m",
  yellow  = ESC .. "38;2;255;202;40m",
  green   = ESC .. "38;2;0;200;83m",
  red     = ESC .. "38;2;255;82;82m",
  coral   = ESC .. "38;2;255;138;101m",
}

-- Cached at init (never change during session)
local USER, HOST, HOME, CACHED_USER_HOST

-- Git cache
local git_cache = {}
local git_cache_ts = 0
local GIT_TTL = 2

-- Last exit code tracking
local last_exit = 0

-- ── helpers ────────────────────────────────────────

local function read_cmd(cmd)
  local h = io.popen(cmd)
  if not h then return "" end
  local r = h:read("*a")
  h:close()
  return r:gsub("%s+$", "")
end

local function dir_short(pwd)
  if not pwd or pwd == "" then pwd = read_cmd("pwd") end
  if pwd:sub(1, #HOME) == HOME then
    pwd = "~" .. pwd:sub(#HOME + 1)
  end
  return pwd
end

-- Single git call: branch + porcelain in one
local function git_status()
  local now = os.time()
  if now - git_cache_ts < GIT_TTL then
    return git_cache
  end

  local handle = io.popen("git status --porcelain --branch 2>/dev/null")
  if not handle then
    git_cache = nil
    return nil
  end

  local output = handle:read("*a")
  handle:close()
  if output == "" then
    git_cache = nil
    return nil
  end

  local lines = {}
  for line in output:gmatch("[^\n]+") do
    table.insert(lines, line)
  end

  -- First line is branch info: "## branch...origin/main"
  local branch_line = lines[1] or ""
  local branch = branch_line:match("^## ([^%.]+)") or branch_line:match("^## (.+)%.%.%.") or ""
  -- Remove leading "HEAD (no branch)" or similar
  if branch == "" then
    branch = branch_line:match("^## (.+)$") or "detached"
  end

  local dirty = false
  local staged = false
  local untracked = false
  local ahead = 0
  local behind = 0

  -- Parse ahead/behind from branch line: "## main...origin/main [ahead 1, behind 2]"
  local ahead_s, behind_s = branch_line:match("ahead (%d+)"), branch_line:match("behind (%d+)")
  if ahead_s then ahead = tonumber(ahead_s) end
  if behind_s then behind = tonumber(behind_s) end

  -- Check remaining lines for file status
  for i = 2, #lines do
    local line = lines[i]
    if line:match("^[MARCD]") or line:match("^.[MARCD]") then
      staged = true
      dirty = true
    elseif line:match("^.[MT]") or line:match("^[MT]") then
      -- modified but not staged
      dirty = true
    elseif line:match("^%?%?") or line:match("^!!") then
      untracked = true
      dirty = true
    end
  end

  if dirty == false and staged == false and untracked == false then
    dirty = false
  end

  git_cache = { branch = branch, dirty = dirty, staged = staged, untracked = untracked, ahead = ahead, behind = behind }
  git_cache_ts = now
  return git_cache
end

local function git_str()
  local g = git_status()
  if not g or g.branch == "" then return "" end

  local parts = { C.cyan .. "(", C.yellow .. g.branch }
  if g.dirty then
    table.insert(parts, C.red .. " ●")
  else
    table.insert(parts, C.green .. " ✓")
  end
  if g.ahead > 0 then
    table.insert(parts, C.teal .. " ↑" .. g.ahead)
  end
  if g.behind > 0 then
    table.insert(parts, C.coral .. " ↓" .. g.behind)
  end
  table.insert(parts, C.cyan .. ")" .. RESET)
  return table.concat(parts)
end

local function exit_str()
  if last_exit == 0 then return "" end
  return C.gray .. " [" .. last_exit .. "]" .. RESET
end

-- ── prompt builder ─────────────────────────────────

function get_prompt()
  local pwd = dir_short(nil)
  local gs = git_str()
  local es = exit_str()

  -- Line 1: ╭─ user@host ~/path (git) [exit]
  local line1 = {}
  table.insert(line1, C.aqua .. "╭─" .. RESET)
  table.insert(line1, " ")
  table.insert(line1, BOLD .. C.white .. CACHED_USER_HOST .. RESET)
  table.insert(line1, " ")
  table.insert(line1, C.cyan .. pwd .. RESET)
  if gs ~= "" then
    table.insert(line1, " ")
    table.insert(line1, gs)
  end
  if es ~= "" then
    table.insert(line1, es)
  end

  -- Line 2: ╰─❯ 
  local line2 = C.aqua .. "╰─❯ " .. RESET

  return table.concat(line1) .. "\n" .. line2
end

-- ── execute hook ───────────────────────────────────

function execute_command(args)
  if args[1] == "exit" or args[1] == "quit" then
    return false
  end
  return false
end

-- Expose for shell to set exit code after each command
function set_exit_code(code)
  last_exit = code
end

-- ── init ───────────────────────────────────────────

USER = os.getenv("USER") or "user"
HOST = read_cmd("hostname")
if HOST == "" then HOST = "localhost" end
HOME = os.getenv("HOME") or "/home/" .. USER
CACHED_USER_HOST = USER .. "@" .. HOST

print("Aquia theme loaded")
