package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"bakshell/internal/data"
)

type redirect struct {
	fd       int // 0=stdin, 1=stdout, 2=stderr, -1=both(&)
	filename string
	append   bool // only for file redirects
	dup      int  // fd to dup to, -1 if file redirect
}

type command struct {
	args     []string
	redirs   []redirect
	isData   bool
	dataFn   func(args []string, stdin io.Reader) (string, int) // returns output, exit code
}

// isDataBuiltin returns true if name is a structured data command.
func isDataBuiltin(name string) bool {
	switch name {
	case "where", "sort-by", "select", "from-json", "from-csv", "to-json", "to-csv",
		"first", "last", "count", "uniq", "sum", "avg", "min", "max":
		return true
	}
	return false
}

// dataCommands maps data builtin names to their handler functions.
func (s *Shell) getDataFn(name string) func(args []string, stdin io.Reader) (string, int) {
	switch name {
	case "from-json":
		return s.handleFromJSON
	case "from-csv":
		return s.handleFromCSV
	case "to-json":
		return s.handleToJSON
	case "to-csv":
		return s.handleToCSV
	case "where":
		return s.handleWhere
	case "sort-by":
		return s.handleSortBy
	case "select":
		return s.handleSelect
	case "first":
		return s.handleFirst
	case "last":
		return s.handleLast
	case "count":
		return s.handleCount
	case "uniq":
		return s.handleUniq
	}
	return nil
}

// --- Data command handlers ---
// Each reads TableValue from stdin (JSON), transforms, returns JSON string.

func (s *Shell) handleFromJSON(args []string, stdin io.Reader) (string, int) {
	var input string
	if len(args) > 0 {
		// Read from file
		b, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "from-json: %v\n", err)
			return "", 1
		}
		input = string(b)
	} else {
		b, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "from-json: %v\n", err)
			return "", 1
		}
		input = string(b)
	}

	tbl, err := data.FromJSON(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "from-json: %v\n", err)
		return "", 1
	}
	json, err := tbl.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "from-json: %v\n", err)
		return "", 1
	}
	return json, 0
}

func (s *Shell) handleFromCSV(args []string, stdin io.Reader) (string, int) {
	var input string
	if len(args) > 0 {
		b, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "from-csv: %v\n", err)
			return "", 1
		}
		input = string(b)
	} else {
		b, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "from-csv: %v\n", err)
			return "", 1
		}
		input = string(b)
	}

	tbl, err := data.FromCSV(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "from-csv: %v\n", err)
		return "", 1
	}
	json, err := tbl.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "from-csv: %v\n", err)
		return "", 1
	}
	return json, 0
}

func (s *Shell) handleToJSON(args []string, stdin io.Reader) (string, int) {
	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "to-json: %v\n", err)
		return "", 1
	}
	// Validate by parsing
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "to-json: %v\n", err)
		return "", 1
	}
	json, err := tbl.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "to-json: %v\n", err)
		return "", 1
	}
	return json, 0
}

func (s *Shell) handleToCSV(args []string, stdin io.Reader) (string, int) {
	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "to-csv: %v\n", err)
		return "", 1
	}
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "to-csv: %v\n", err)
		return "", 1
	}
	csv, err := tbl.ToCSV()
	if err != nil {
		fmt.Fprintf(os.Stderr, "to-csv: %v\n", err)
		return "", 1
	}
	return csv, 0
}

func (s *Shell) handleWhere(args []string, stdin io.Reader) (string, int) {
	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "where: %v\n", err)
		return "", 1
	}
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "where: %v\n", err)
		return "", 1
	}

	var conds []data.Condition
	for _, arg := range args {
		cond, err := data.ParseCondition(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "where: %v\n", err)
			return "", 1
		}
		conds = append(conds, cond)
	}

	result, err := tbl.Filter(conds)
	if err != nil {
		fmt.Fprintf(os.Stderr, "where: %v\n", err)
		return "", 1
	}
	json, err := result.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "where: %v\n", err)
		return "", 1
	}
	return json, 0
}

func (s *Shell) handleSortBy(args []string, stdin io.Reader) (string, int) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "sort-by: expected field name\n")
		return "", 1
	}

	field := args[0]
	desc := false
	for _, a := range args[1:] {
		if a == "--desc" || a == "-d" {
			desc = true
		}
	}

	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sort-by: %v\n", err)
		return "", 1
	}
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sort-by: %v\n", err)
		return "", 1
	}

	result, err := tbl.SortBy(field, desc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sort-by: %v\n", err)
		return "", 1
	}
	json, err := result.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sort-by: %v\n", err)
		return "", 1
	}
	return json, 0
}

func (s *Shell) handleSelect(args []string, stdin io.Reader) (string, int) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "select: expected column names\n")
		return "", 1
	}

	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "select: %v\n", err)
		return "", 1
	}
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "select: %v\n", err)
		return "", 1
	}

	result, err := tbl.Select(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "select: %v\n", err)
		return "", 1
	}
	json, err := result.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "select: %v\n", err)
		return "", 1
	}
	return json, 0
}

func (s *Shell) handleFirst(args []string, stdin io.Reader) (string, int) {
	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "first: %v\n", err)
		return "", 1
	}
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "first: %v\n", err)
		return "", 1
	}

	n := 10 // default
	if len(args) > 0 {
		if v, err := strconv.Atoi(args[0]); err == nil && v > 0 {
			n = v
		}
	}
	result := tbl.First(n)
	json, err := result.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "first: %v\n", err)
		return "", 1
	}
	return json, 0
}

func (s *Shell) handleLast(args []string, stdin io.Reader) (string, int) {
	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "last: %v\n", err)
		return "", 1
	}
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "last: %v\n", err)
		return "", 1
	}

	n := 10
	if len(args) > 0 {
		if v, err := strconv.Atoi(args[0]); err == nil && v > 0 {
			n = v
		}
	}
	result := tbl.Last(n)
	json, err := result.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "last: %v\n", err)
		return "", 1
	}
	return json, 0
}

func (s *Shell) handleCount(args []string, stdin io.Reader) (string, int) {
	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "count: %v\n", err)
		return "", 1
	}
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "count: %v\n", err)
		return "", 1
	}

	// If field args given, do group-by count
	if len(args) > 0 {
		result := tbl.GroupBy(args)
		json, err := result.ToJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "count: %v\n", err)
			return "", 1
		}
		return json, 0
	}

	// Simple count
	out := fmt.Sprintf(`[{"count": %d}]`, len(tbl.Rows))
	return out, 0
}

func (s *Shell) handleUniq(args []string, stdin io.Reader) (string, int) {
	b, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
		return "", 1
	}
	tbl, err := data.FromJSON(string(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
		return "", 1
	}

	result := tbl.Unique(args)
	json, err := result.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "uniq: %v\n", err)
		return "", 1
	}
	return json, 0
}

// parseSegment splits tokens at | to build a pipeline,
// extracting redirections from each command.
func parseSegment(tokens []string) []command {
	var cmds []command
	start := 0
	for i, tok := range tokens {
		if tok == "|" {
			cmds = append(cmds, parseCommand(tokens[start:i]))
			start = i + 1
		}
	}
	cmds = append(cmds, parseCommand(tokens[start:]))
	return cmds
}

func parseCommand(tokens []string) command {
	var args []string
	var redirs []redirect

	peek := func(i int) string {
		if i < len(tokens) {
			return tokens[i]
		}
		return ""
	}

	isDigit := func(s string) bool {
		return len(s) == 1 && s[0] >= '0' && s[0] <= '9'
	}

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		// Handle numeric fd prefix before redirection ops
		if isDigit(tok) {
			next := peek(i + 1)

			if next == ">" || next == ">>" {
				fd := int(tok[0] - '0')
				isAppend := next == ">>"
				i++
				nnext := peek(i + 1)

				// Check for >& (dup)
				if nnext == "&" && peek(i+2) != "" {
					dupFd := peek(i + 2)
					if dupFd == "1" || dupFd == "2" {
						redirs = append(redirs, redirect{fd: fd, dup: int(dupFd[0] - '0')})
						i += 2
						continue
					}
				}

				// File redirect
				if nnext != "" && nnext != "|" && nnext != ";" && nnext != "&" && nnext != "&&" && nnext != "||" {
					redirs = append(redirs, redirect{fd: fd, filename: nnext, append: isAppend})
					i++
				}
				continue
			}
		}

		switch {
		case tok == ">":
			next := peek(i + 1)
			if next == "&" && peek(i+2) != "" && (peek(i+2) == "1" || peek(i+2) == "2") {
				// 1>&2 form (>&)
				redirs = append(redirs, redirect{fd: 1, dup: int(peek(i+2)[0] - '0')})
				i += 2
			} else if next != "" && next != "|" && next != ";" && next != "&" && next != "&&" && next != "||" {
				redirs = append(redirs, redirect{fd: 1, filename: next})
				i++
			}
		case tok == ">>":
			if n := peek(i + 1); n != "" && n != "|" && n != ";" && n != "&" && n != "&&" && n != "||" {
				redirs = append(redirs, redirect{fd: 1, filename: n, append: true})
				i++
			}
		case tok == "<":
			if n := peek(i + 1); n != "" && n != "|" && n != ";" && n != "&" && n != "&&" && n != "||" {
				redirs = append(redirs, redirect{fd: 0, filename: n})
				i++
			}
		case tok == "&>":
			if n := peek(i + 1); n != "" && n != "|" && n != ";" && n != "&" && n != "&&" && n != "||" {
				redirs = append(redirs, redirect{fd: -1, filename: n})
				i++
			}
		case tok == "&>>":
			if n := peek(i + 1); n != "" && n != "|" && n != ";" && n != "&" && n != "&&" && n != "||" {
				redirs = append(redirs, redirect{fd: -1, filename: n, append: true})
				i++
			}
		default:
			args = append(args, tok)
		}
	}
	cmd := command{args: args, redirs: redirs}
	if len(args) > 0 && isDataBuiltin(args[0]) {
		cmd.isData = true
	}
	return cmd
}

// execPipeline runs a pipeline (cmds[0] | cmds[1] | ...) and returns the exit code
// of the last command.
func (s *Shell) execPipeline(cmds []command) int {
	if len(cmds) == 0 {
		return 0
	}

	// Single command — no pipe needed
	if len(cmds) == 1 {
		return s.execCommand(cmds[0])
	}

	// Check if any command is a data command
	hasData := false
	for _, c := range cmds {
		if c.isData {
			hasData = true
			break
		}
	}

	if hasData {
		return s.execDataPipeline(cmds)
	}

	return s.execFDPipeline(cmds)
}

// execFDPipeline runs an fd-based concurrent pipeline (existing behavior).
func (s *Shell) execFDPipeline(cmds []command) int {
	last := len(cmds) - 1
	procs := make([]*exec.Cmd, len(cmds))
	pipes := make([]*os.File, 0, last*2)

	for i := 0; i < last; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "pipe:", err)
			return 1
		}
		pipes = append(pipes, r, w)
	}

	for i, cmd := range cmds {
		c := s.buildCmd(cmd)

		if i == 0 {
			c.Stdin = os.Stdin
		} else {
			c.Stdin = pipes[(i-1)*2]
		}

		if i < last {
			c.Stdout = pipes[i*2+1]
		} else {
			c.Stdout = os.Stdout
		}

		c.Stderr = os.Stderr
		applyRedirs(c, cmd.redirs)
		procs[i] = c
	}

	for _, c := range procs {
		if err := c.Start(); err != nil {
			if isCommandNotFound(err) {
				fmt.Fprintf(os.Stderr, "%s: command not found\n", c.Args[0])
			} else {
				fmt.Fprintf(os.Stderr, "%s: %v\n", c.Args[0], err)
			}
			return 127
		}
	}

	for i := 0; i < last; i++ {
		pipes[i*2+1].Close()
	}

	lastExit := 0
	for _, c := range procs {
		if err := c.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				lastExit = exitErr.ExitCode()
			} else {
				lastExit = 127
			}
		} else {
			lastExit = 0
		}
	}

	return lastExit
}

// execDataPipeline runs a pipeline that contains data commands.
// Commands run sequentially, passing JSON through pipes or directly.
func (s *Shell) execDataPipeline(cmds []command) int {
	last := len(cmds) - 1

	// We'll set up fd pipes between segments. For external→data or data→external
	// transitions, data is serialized as JSON through the fd pipe.
	// For data→data transitions, we pass the string directly.

	type pipeEnd struct {
		r *os.File
		w *os.File
	}

	pipes := make([]pipeEnd, last)
	for i := 0; i < last; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "pipe:", err)
			return 1
		}
		pipes[i] = pipeEnd{r, w}
	}

	var lastJSON string // holds JSON output from previous data command
	var lastExit int

	for i, cmd := range cmds {
		if cmd.isData {
			// Save undo state before transformation
			if lastJSON != "" {
				if tbl, err := data.FromJSON(lastJSON); err == nil {
					s.undoTable = &tbl
				}
			}

			// Determine stdin source
			var stdinR io.Reader
			if i == 0 {
				stdinR = os.Stdin
			} else if lastJSON != "" {
				stdinR = strings.NewReader(lastJSON)
			} else {
				stdinR = pipes[i-1].r
			}

			fn := s.getDataFn(cmd.args[0])
			if fn == nil {
				fmt.Fprintf(os.Stderr, "%s: internal error: no data handler\n", cmd.args[0])
				return 1
			}

			out, code := fn(cmd.args[1:], stdinR)
			lastExit = code
			if code != 0 {
				break
			}

			if i < last {
				// Check if next cmd is also a data cmd
				if cmds[i+1].isData {
					lastJSON = out
				} else {
					// Write to pipe for external command
					pipes[i].w.Write([]byte(out))
					pipes[i].w.Close()
					lastJSON = ""
				}
			} else {
				// Last command in pipeline — display output
				if out == "" {
					continue
				}
				// For format commands, print raw. For transform commands, show table
				if cmd.args[0] == "to-json" || cmd.args[0] == "to-csv" {
					fmt.Println(out)
				} else {
					if tbl, err := data.FromJSON(out); err == nil {
						fmt.Print(colorTable(tbl.String()))
					} else {
						fmt.Println(out)
					}
				}
			}
		} else {
			// External command
			c := s.buildCmd(cmd)

			if i == 0 {
				c.Stdin = os.Stdin
			} else if lastJSON != "" {
				r, w, err := os.Pipe()
				if err != nil {
					fmt.Fprintln(os.Stderr, "pipe:", err)
					return 1
				}
				w.Write([]byte(lastJSON))
				w.Close()
				c.Stdin = r
				lastJSON = ""
			} else {
				c.Stdin = pipes[i-1].r
			}

			if i < last {
				// If next cmd is data, capture stdout
				if cmds[i+1].isData {
					r, w, err := os.Pipe()
					if err != nil {
						fmt.Fprintln(os.Stderr, "pipe:", err)
						return 1
					}
					pipes[i] = pipeEnd{r, w}
					c.Stdout = pipes[i].w
				} else {
					c.Stdout = pipes[i].w
				}
			} else {
				c.Stdout = os.Stdout
			}

			c.Stderr = os.Stderr
			applyRedirs(c, cmd.redirs)

			if err := c.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					lastExit = exitErr.ExitCode()
				} else {
					if isCommandNotFound(err) {
						fmt.Fprintf(os.Stderr, "%s: command not found\n", c.Args[0])
					} else {
						fmt.Fprintf(os.Stderr, "%s: %v\n", c.Args[0], err)
					}
					lastExit = 127
				}
			} else {
				lastExit = 0
			}

			if i < last {
				pipes[i].w.Close()
				// If next cmd is a data cmd, read pipe output now
				if cmds[i+1].isData {
					b, err := io.ReadAll(pipes[i].r)
					if err == nil {
						lastJSON = string(b)
					}
				}
			}
		}
	}

	return lastExit
}

// execCommand runs a single command with redirections (no pipe).
func (s *Shell) execCommand(cmd command) int {
	if len(cmd.args) == 0 {
		return 0
	}

	switch cmd.args[0] {
	case "cd":
		return s.execCD(cmd.args[1:])
	case "exit", "quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
		return 0
	case "echo":
		return s.execEcho(cmd.args[1:])
	case "pwd":
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "pwd: %v\n", err)
			return 1
		}
		fmt.Println(pwd)
		return 0
	case "type":
		return s.execType(cmd.args[1:])
	case "export":
		return s.execExport(cmd.args[1:])
	case "unset":
		return s.execUnset(cmd.args[1:])
	case "history":
		return s.execHistory(cmd.args[1:])
	case "help":
		return s.execHelp(cmd.args[1:])
	case "alias":
		return s.execAlias(cmd.args[1:])
	case "unalias":
		return s.execUnalias(cmd.args[1:])
	case "trash":
		return s.execTrash(cmd.args[1:])
	case "undo":
		return s.execUndo(cmd.args[1:])
	case "confirm":
		return s.execConfirm(cmd.args[1:])
	case "from-json", "from-csv", "to-json", "to-csv", "where", "sort-by", "select", "first", "last", "count", "uniq":
		fn := s.getDataFn(cmd.args[0])
		if fn == nil {
			fmt.Fprintf(os.Stderr, "%s: internal error\n", cmd.args[0])
			return 1
		}
		out, code := fn(cmd.args[1:], os.Stdin)
		if out != "" {
			fmt.Println(out)
		}
		return code
	}

	if s.plugins.ExecuteCommand(cmd.args) {
		return 0
	}

	return s.execExternalCmd(cmd)
}

var builtins = map[string]string{
	"cd":        "change the current directory",
	"exit":      "exit the shell",
	"quit":      "exit the shell",
	"echo":      "display a line of text",
	"pwd":       "print the current working directory",
	"type":      "display information about command type",
	"export":    "set an environment variable",
	"unset":     "unset an environment variable",
	"history":   "display or clear the command history",
	"help":      "display information about built-in commands",
	"alias":     "define or display aliases",
	"unalias":   "remove alias definitions",
	"confirm":   "prompt for confirmation, exits 0 for yes, 1 otherwise",
	"trash":     "move files to trash (~/.zencr/trash/)",
	"undo":      "restore the previous table state",
	"from-json": "parse JSON into structured data",
	"from-csv":  "parse CSV into structured data",
	"to-json":   "convert structured data to JSON",
	"to-csv":    "convert structured data to CSV",
	"where":     "filter structured data rows by conditions",
	"sort-by":   "sort structured data by a field",
	"select":    "select specific columns from structured data",
	"first":     "show first N rows (default 10)",
	"last":      "show last N rows (default 10)",
	"count":     "count rows, or count by groups with `count field`",
	"uniq":      "show unique rows",
}

func (s *Shell) execCD(args []string) int {
	target := s.home
	if len(args) > 0 {
		target = strings.Join(args, " ")
	}
	if err := os.Chdir(target); err != nil {
		fmt.Fprintf(os.Stderr, "cd: %v\n", err)
		return 1
	}
	return 0
}

func (s *Shell) execEcho(args []string) int {
	noNewline := false
	parts := args
	if len(parts) > 0 && parts[0] == "-n" {
		noNewline = true
		parts = parts[1:]
	}
	out := strings.Join(parts, " ")
	if noNewline {
		fmt.Print(out)
	} else {
		fmt.Println(out)
	}
	return 0
}

func (s *Shell) execType(args []string) int {
	if len(args) == 0 {
		return 0
	}
	for _, name := range args {
		if desc, ok := builtins[name]; ok {
			fmt.Printf("%s is a shell built-in (%s)\n", name, desc)
			continue
		}
		path, err := exec.LookPath(name)
		if err != nil {
			fmt.Printf("%s: not found\n", name)
		} else {
			fmt.Printf("%s is %s\n", name, path)
		}
	}
	return 0
}

func (s *Shell) execExport(args []string) int {
	if len(args) == 0 {
		for _, e := range os.Environ() {
			fmt.Println(e)
		}
		return 0
	}
	for _, arg := range args {
		k, v, ok := strings.Cut(arg, "=")
		if !ok {
			// export NAME (mark for child processes — already in env, no-op)
			continue
		}
		if err := os.Setenv(k, v); err != nil {
			fmt.Fprintf(os.Stderr, "export: %v\n", err)
			return 1
		}
	}
	return 0
}

func (s *Shell) execUnset(args []string) int {
	for _, name := range args {
		if err := os.Unsetenv(name); err != nil {
			fmt.Fprintf(os.Stderr, "unset: %v\n", err)
			return 1
		}
	}
	return 0
}

func (s *Shell) execHistory(args []string) int {
	// Handle -c flag (clear)
	for _, a := range args {
		if a == "-c" {
			histPath := s.home + "/.zencr/history"
			if err := os.WriteFile(histPath, []byte{}, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "history: %v\n", err)
				return 1
			}
			return 0
		}
	}

	// Handle numeric arg (show last N)
	var limit int
	for _, a := range args {
		if n, err := strconv.Atoi(a); err == nil {
			limit = n
			break
		}
	}

	histPath := s.home + "/.zencr/history"
	data, err := os.ReadFile(histPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "history: %v\n", err)
		return 1
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	start := 0
	if limit > 0 && limit < len(lines) {
		start = len(lines) - limit
	}
	for i, line := range lines[start:] {
		fmt.Printf("%5d  %s\n", start+i+1, line)
	}
	return 0
}

func (s *Shell) execHelp(args []string) int {
	if len(args) > 0 {
		for _, name := range args {
			if desc, ok := builtins[name]; ok {
				fmt.Printf("%s - %s\n", name, desc)
			} else {
				fmt.Printf("help: no such built-in: %s\n", name)
			}
		}
		return 0
	}

	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("BakShell built-in commands:")
	for _, name := range names {
		fmt.Printf("  %-10s %s\n", name, builtins[name])
	}
	return 0
}

func (s *Shell) execAlias(args []string) int {
	if len(args) == 0 {
		// List all aliases
		names := make([]string, 0, len(s.aliases))
		for name := range s.aliases {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Printf("alias %s='%s'\n", name, s.aliases[name])
		}
		return 0
	}

	for _, arg := range args {
		k, v, ok := strings.Cut(arg, "=")
		if !ok {
			// Show single alias
			if val, exists := s.aliases[arg]; exists {
				fmt.Printf("alias %s='%s'\n", arg, val)
			}
			continue
		}
		// Strip surrounding quotes if present
		if len(v) >= 2 && ((v[0] == '\'' && v[len(v)-1] == '\'') || (v[0] == '"' && v[len(v)-1] == '"')) {
			v = v[1 : len(v)-1]
		}
		s.aliases[k] = v
	}
	return 0
}

func (s *Shell) execUnalias(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "unalias: usage: unalias [-a] name [name ...]\n")
		return 1
	}
	for _, a := range args {
		if a == "-a" {
			s.aliases = make(map[string]string)
			return 0
		}
		delete(s.aliases, a)
	}
	return 0
}

func (s *Shell) execConfirm(args []string) int {
	msg := "Are you sure?"
	if len(args) > 0 {
		msg = strings.Join(args, " ")
	}
	fmt.Printf("%s [y/N] ", msg)
	var answer string
	fmt.Scanln(&answer)
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "y" || answer == "yes" {
		return 0
	}
	return 1
}

func (s *Shell) execTrash(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "trash: usage: trash <file> [file...]\n")
		return 1
	}
	trashDir := s.home + "/.zencr/trash"
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "trash: %v\n", err)
		return 1
	}
	for _, arg := range args {
		base := arg
		if idx := strings.LastIndex(base, "/"); idx >= 0 {
			base = base[idx+1:]
		}
		target := trashDir + "/" + base
		if _, err := os.Stat(target); err == nil {
			for i := 1; ; i++ {
				target = fmt.Sprintf("%s/%s.%d", trashDir, base, i)
				if _, err := os.Stat(target); os.IsNotExist(err) {
					break
				}
			}
		}
		if err := os.Rename(arg, target); err != nil {
			fmt.Fprintf(os.Stderr, "trash: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "trash: moved %s to trash\n", arg)
	}
	return 0
}

func (s *Shell) execUndo(args []string) int {
	if s.undoTable == nil {
		fmt.Fprintf(os.Stderr, "undo: nothing to undo\n")
		return 1
	}
	json, err := s.undoTable.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "undo: %v\n", err)
		return 1
	}
	fmt.Println(json)
	s.undoTable = nil
	return 0
}

func colorTable(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return s
	}
	// Bold + underline header
	header := "\033[1;37m" + lines[0] + "\033[0m"
	// Dim separator
	sep := "\033[2m" + lines[1] + "\033[0m"
	rows := strings.Join(lines[2:], "\n")
	return header + "\n" + sep + "\n" + rows + "\n"
}

func (s *Shell) buildCmd(cmd command) *exec.Cmd {
	return exec.Command(cmd.args[0], cmd.args[1:]...)
}

func isCommandNotFound(err error) bool {
	if err == exec.ErrNotFound {
		return true
	}
	if e, ok := err.(*exec.Error); ok && e.Err == exec.ErrNotFound {
		return true
	}
	return false
}

func (s *Shell) execExternalCmd(cmd command) int {
	if _, err := exec.LookPath(cmd.args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "%s: command not found\n", cmd.args[0])
		return 127
	}

	c := s.buildCmd(cmd)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	applyRedirs(c, cmd.redirs)

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		if isCommandNotFound(err) {
			fmt.Fprintf(os.Stderr, "%s: command not found\n", cmd.args[0])
		} else {
			fmt.Fprintf(os.Stderr, "%s: %v\n", cmd.args[0], err)
		}
		return 127
	}
	return 0
}

func applyRedirs(c *exec.Cmd, redirs []redirect) {
	for _, r := range redirs {
		switch {
		case r.dup == 1:
			if r.fd == 2 {
				c.Stderr = c.Stdout
			}
		case r.dup == 2:
			if r.fd == 1 {
				c.Stdout = c.Stderr
			}

		case r.fd == -1:
			// &> or &>> — both stdout and stderr
			var flags int
			if r.append {
				flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
			} else {
				flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			}
			f, err := os.OpenFile(r.filename, flags, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "redirection error: %v\n", err)
				continue
			}
			c.Stdout = f
			c.Stderr = f

		default:
			// File redirection for a specific fd
			var flags int
			if r.append {
				flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
			} else if r.fd == 0 {
				flags = os.O_RDONLY
			} else {
				flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			}
			f, err := os.OpenFile(r.filename, flags, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "redirection error: %v\n", err)
				continue
			}
			switch r.fd {
			case 0:
				c.Stdin = f
			case 1:
				c.Stdout = f
			case 2:
				c.Stderr = f
			}
		}
	}
}
