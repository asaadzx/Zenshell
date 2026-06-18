-- Auto proxy plugin
-- Automatically sets HTTP_PROXY/HTTPS_PROXY based on matching domain patterns
-- Configure: proxy_patterns = { ["*.internal.corp"] = "http://proxy:8080", ... }

local DEFAULT_PROXY = os.getenv("BSH_DEFAULT_PROXY") or "http://127.0.0.1:8080"

-- Patterns: host pattern -> proxy URL (supports * wildcard)
local patterns = {}
if type(proxy_patterns) == "table" then
  patterns = proxy_patterns
else
  -- Default patterns
  patterns = {
    ["*.local"] = DEFAULT_PROXY,
    ["*.internal"] = DEFAULT_PROXY,
    ["10.*"] = DEFAULT_PROXY,
    ["192.168.*"] = DEFAULT_PROXY,
  }
end

local active = false
local current_proxy = ""
local last_host = ""

local function matches_pattern(host, pattern)
  if not pattern or not host then return false end
  if pattern:sub(1, 1) == "*" then
    local suffix = pattern:sub(2)
    return host:sub(-#suffix) == suffix
  end
  return host == pattern
end

local function find_proxy(host)
  for pat, proxy in pairs(patterns) do
    if matches_pattern(host, pat) then
      return proxy
    end
  end
  return nil
end

local function detect_host()
  -- Use current domain from /etc/resolv.conf or hostname
  local handle = io.popen("hostname 2>/dev/null")
  if not handle then return "" end
  local name = handle:read("*l") or ""
  handle:close()
  return name
end

local function update_proxy()
  local host = detect_host()
  if host == last_host then return end
  last_host = host

  local proxy = find_proxy(host)

  if proxy and not active then
    os.setenv("HTTP_PROXY", proxy)
    os.setenv("HTTPS_PROXY", proxy)
    os.setenv("http_proxy", proxy)
    os.setenv("https_proxy", proxy)
    active = true
    current_proxy = proxy
    print("proxy: enabled " .. proxy .. " for " .. host)
  elseif not proxy and active then
    os.setenv("HTTP_PROXY", "")
    os.setenv("HTTPS_PROXY", "")
    os.setenv("http_proxy", "")
    os.setenv("https_proxy", "")
    active = false
    print("proxy: disabled for " .. host)
  end
end

update_proxy()

function get_prompt_suffix()
  if not active then return "" end
  return "\27[90m proxy\27[0m"
end

function execute_command(args)
  if #args == 0 then return false end

  if args[1] == "proxy" then
    if #args == 1 then
      if active then
        print("proxy: enabled (" .. current_proxy .. ")")
      else
        print("proxy: disabled")
      end
      return true
    end

    if args[2] == "on" then
      local proxy = args[3] or DEFAULT_PROXY
      os.setenv("HTTP_PROXY", proxy)
      os.setenv("HTTPS_PROXY", proxy)
      os.setenv("http_proxy", proxy)
      os.setenv("https_proxy", proxy)
      active = true
      current_proxy = proxy
      print("proxy: enabled " .. proxy)
      return true
    end

    if args[2] == "off" then
      os.setenv("HTTP_PROXY", "")
      os.setenv("HTTPS_PROXY", "")
      os.setenv("http_proxy", "")
      os.setenv("https_proxy", "")
      active = false
      print("proxy: disabled")
      return true
    end

    print("Usage: proxy [on [url]|off]")
    return true
  end

  return false
end

print("Proxy plugin loaded (" .. (active and "active" or "inactive") .. ")")
