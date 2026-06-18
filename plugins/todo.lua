-- Simple todo list plugin
-- Usage:
--   todo                  - list all todos
--   todo add <text>       - add a new todo
--   todo done <n>         - mark todo #n as done
--   todo rm <n>           - remove todo #n
--   todo clear            - remove all completed

local DATA_FILE = os.getenv("HOME") .. "/.zencr/todos.json"

local todos = {}

local function load()
  local f = io.open(DATA_FILE, "r")
  if not f then return end
  local raw = f:read("*a")
  f:close()
  if raw and raw ~= "" then
    local ok, data = pcall(function()
      local t = {}
      for line in raw:gmatch('[^\n]+') do
        local text, done = line:match("^(.-)\t(done)$")
        if text then
          table.insert(t, { text = text, done = true })
        else
          table.insert(t, { text = line, done = false })
        end
      end
      return t
    end)
    if ok then todos = data end
  end
end

local function save()
  local lines = {}
  for _, todo in ipairs(todos) do
    if todo.done then
      table.insert(lines, todo.text .. "\tdone")
    else
      table.insert(lines, todo.text)
    end
  end
  local f = io.open(DATA_FILE, "w")
  if f then
    f:write(table.concat(lines, "\n") .. "\n")
    f:close()
  end
end

local function list_todos()
  if #todos == 0 then
    print("No todos. Use 'todo add <text>' to add one.")
    return
  end
  for i, todo in ipairs(todos) do
    local mark = todo.done and "✓" or " "
    local color = todo.done and "\27[90m" or "\27[0m"
    local reset = "\27[0m"
    print(string.format("%s%3d. [%s] %s%s", color, i, mark, todo.text, reset))
  end
end

load()

function execute_command(args)
  if #args == 0 then return false end

  if args[1] == "todo" then
    if #args == 1 then
      list_todos()
      return true
    end

    local sub = args[2]

    if sub == "add" and #args >= 3 then
      local text = table.concat(args, " ", 3)
      table.insert(todos, { text = text, done = false })
      save()
      print(string.format("Added todo: %s", text))
      return true
    end

    if sub == "done" and #args >= 3 then
      local n = tonumber(args[3])
      if n and n >= 1 and n <= #todos then
        todos[n].done = true
        save()
        print(string.format("Marked #%d as done: %s", n, todos[n].text))
      else
        print("todo: invalid number")
      end
      return true
    end

    if sub == "rm" and #args >= 3 then
      local n = tonumber(args[3])
      if n and n >= 1 and n <= #todos then
        local removed = table.remove(todos, n)
        save()
        print(string.format("Removed: %s", removed.text))
      else
        print("todo: invalid number")
      end
      return true
    end

    if sub == "clear" then
      local count = 0
      local i = 1
      while i <= #todos do
        if todos[i].done then
          table.remove(todos, i)
          count = count + 1
        else
          i = i + 1
        end
      end
      save()
      print(string.format("Cleared %d completed todos", count))
      return true
    end

    print("Usage: todo [add <text>|done <n>|rm <n>|clear]")
    return true
  end

  return false
end

print("Todo plugin loaded (" .. #todos .. " items)")
