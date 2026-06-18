-- Random developer quote plugin
-- Shows a random programming quote in the prompt

local ESC = "\27["
local RESET = ESC .. "0m"
local DIM = ESC .. "2m"
local ITALIC = ESC .. "3m"

local quotes = {
  "Any fool can write code that a computer can understand. Good programmers write code that humans can understand. — M. Fowler",
  "First solve the problem, then write the code. — J. Johnson",
  "Simplicity is the ultimate sophistication. — L. da Vinci",
  "Make it work, make it right, make it fast. — K. Beck",
  "Talk is cheap. Show me the code. — L. Torvalds",
  "Programs must be written for people to read. — H. Abelson",
  "Premature optimization is the root of all evil. — D. Knuth",
  "Debugging is twice as hard as writing the code. — B. Kernighan",
  "The best code is no code at all. — J. Raskin",
  "Code is like humor. When you have to explain it, it's bad. — C. Hunt",
  "Software entropy never decreases. — W. Humphrey",
  "There are only two hard things: cache invalidation and naming things. — P. Karlton",
  "It works on my machine. — Every developer",
  "A language that doesn't affect the way you think about programming is not worth knowing. — A. Perlis",
  "The best programs are written so that computing machines can perform them quickly. — A. Turing",
  "Don't comment bad code — rewrite it. — B. Kernighan",
  "The function of good software is to make the complex appear to be simple. — G. Booch",
  "Before software can be reusable it first has to be usable. — R. Johnson",
  "The only way to learn a new programming language is by writing programs in it. — D. Ritchie",
  "Measuring programming progress by lines of code is like measuring aircraft building progress by weight. — B. Gates",
  "Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live. — J. Woods",
  "Without requirements or design, programming is the art of adding bugs to an empty text file. — L. S. Louis",
  "If debugging is the process of removing bugs, then programming must be the process of putting them in. — E. Dijkstra",
  "The most important property of a program is whether it accomplishes the intention of its user. — C.A.R. Hoare",
  "The trouble with programmers is that you can never tell what a programmer is doing until it's too late. — S. Weinberg",
}

local rng = math.random
local last_idx = 0
local quote_text = ""
local rotate_every = 3  -- show same quote for N prompts

local counter = 0

function get_prompt_suffix()
  counter = counter + 1
  if counter % rotate_every == 1 or quote_text == "" then
    local idx = rng(#quotes)
    if idx == last_idx then
      idx = idx % #quotes + 1
    end
    last_idx = idx
    quote_text = quotes[idx]
  end
  return DIM .. ITALIC .. quote_text .. RESET
end

function execute_command(args)
  return false
end

print("Quote plugin loaded (" .. #quotes .. " quotes)")
