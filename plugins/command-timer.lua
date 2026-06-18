-- Command timer plugin
-- Shows elapsed time for each command with configurable threshold

local ESC = "\27["
local RESET = ESC .. "0m"
local DIM = ESC .. "2m"

local threshold_ms = 50    -- minimum time to show (milliseconds)
local start_time = 0

function set_exit_code(code)
  -- no-op, we use execute_command to measure time
end

function execute_command(args)
  if #args == 0 then return false end
  start_time = os.clock()
  return false
end

function on_command_complete()
  if start_time == 0 then return end
  local elapsed = (os.clock() - start_time) * 1000
  start_time = 0

  if elapsed >= threshold_ms then
    local unit = "ms"
    local val = elapsed
    if val >= 1000 then
      val = val / 1000
      unit = "s"
    end
    io.stderr:write(string.format("%s(%.1f %s)%s\n", DIM, val, unit, RESET))
  end
end

print("Command timer plugin loaded (threshold: " .. threshold_ms .. "ms)")
