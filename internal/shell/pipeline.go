package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type redirect struct {
	fd       int // 0=stdin, 1=stdout, 2=stderr, -1=both(&)
	filename string
	append   bool // only for file redirects
	dup      int  // fd to dup to, -1 if file redirect
}

type command struct {
	args   []string
	redirs []redirect
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
	return command{args: args, redirs: redirs}
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
			c.Stdin = pipes[(i-1)*2] // read end of prev pipe
		}

		if i < last {
			c.Stdout = pipes[i*2+1] // write end of this pipe
		} else {
			c.Stdout = os.Stdout
		}

		c.Stderr = os.Stderr
		applyRedirs(c, cmd.redirs)
		procs[i] = c
	}

	// Start all processes
	for _, c := range procs {
		if err := c.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 127
		}
	}

	// Close write ends in parent so readers see EOF
	for i := 0; i < last; i++ {
		pipes[i*2+1].Close()
	}

	// Wait for all processes
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

// execCommand runs a single command with redirections (no pipe).
func (s *Shell) execCommand(cmd command) int {
	if len(cmd.args) == 0 {
		return 0
	}

	switch cmd.args[0] {
	case "cd":
		target := s.home
		if len(cmd.args) > 1 {
			target = strings.Join(cmd.args[1:], " ")
		}
		if err := os.Chdir(target); err != nil {
			fmt.Fprintf(os.Stderr, "cd: %v\n", err)
			return 1
		}
		return 0
	case "exit", "quit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	}

	if s.plugins.ExecuteCommand(cmd.args) {
		return 0
	}

	return s.execExternalCmd(cmd)
}

func (s *Shell) buildCmd(cmd command) *exec.Cmd {
	return exec.Command(cmd.args[0], cmd.args[1:]...)
}

func (s *Shell) execExternalCmd(cmd command) int {
	c := s.buildCmd(cmd)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	applyRedirs(c, cmd.redirs)

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmd.args[0], err)
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
