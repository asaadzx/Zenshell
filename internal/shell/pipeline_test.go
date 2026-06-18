package shell

import (
	"strings"
	"testing"
)

func TestParseCommandSimple(t *testing.T) {
	cmd := parseCommand([]string{"ls", "-la"})
	if len(cmd.args) != 2 {
		t.Fatalf("expected 2 args, got %d: %#v", len(cmd.args), cmd.args)
	}
	if cmd.args[0] != "ls" {
		t.Errorf("expected ls, got %s", cmd.args[0])
	}
	if len(cmd.redirs) != 0 {
		t.Errorf("expected 0 redirs, got %d", len(cmd.redirs))
	}
}

func TestParseCommandNoArgs(t *testing.T) {
	cmd := parseCommand(nil)
	if len(cmd.args) != 0 {
		t.Errorf("expected no args, got %#v", cmd.args)
	}
}

func TestParseCommandStdoutRedirect(t *testing.T) {
	cmd := parseCommand([]string{"echo", "hello", ">", "out.txt"})
	if len(cmd.args) != 2 {
		t.Fatalf("expected 2 args, got %d: %#v", len(cmd.args), cmd.args)
	}
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	r := cmd.redirs[0]
	if r.fd != 1 {
		t.Errorf("expected fd 1, got %d", r.fd)
	}
	if r.filename != "out.txt" {
		t.Errorf("expected out.txt, got %s", r.filename)
	}
	if r.append {
		t.Errorf("expected append=false")
	}
}

func TestParseCommandInputRedirect(t *testing.T) {
	cmd := parseCommand([]string{"cat", "<", "in.txt"})
	if len(cmd.args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %#v", len(cmd.args), cmd.args)
	}
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	if cmd.redirs[0].fd != 0 {
		t.Errorf("expected fd 0, got %d", cmd.redirs[0].fd)
	}
}

func TestParseCommandAppend(t *testing.T) {
	cmd := parseCommand([]string{"echo", "hello", ">>", "log.txt"})
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	if !cmd.redirs[0].append {
		t.Errorf("expected append=true")
	}
}

func TestParseCommandBothRedirect(t *testing.T) {
	cmd := parseCommand([]string{"cmd", "&>", "out.txt"})
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	if cmd.redirs[0].fd != -1 {
		t.Errorf("expected fd -1 (both), got %d", cmd.redirs[0].fd)
	}
}

func TestParseCommandBothAppend(t *testing.T) {
	cmd := parseCommand([]string{"cmd", "&>>", "out.txt"})
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	if cmd.redirs[0].fd != -1 {
		t.Errorf("expected fd -1 (both), got %d", cmd.redirs[0].fd)
	}
	if !cmd.redirs[0].append {
		t.Errorf("expected append=true")
	}
}

func TestParseCommandStderrRedirect(t *testing.T) {
	// Tokenizer splits "2>" into "2" and ">" tokens
	cmd := parseCommand([]string{"cmd", "2", ">", "err.txt"})
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	if cmd.redirs[0].fd != 2 {
		t.Errorf("expected fd 2, got %d", cmd.redirs[0].fd)
	}
	if cmd.redirs[0].filename != "err.txt" {
		t.Errorf("expected err.txt, got %s", cmd.redirs[0].filename)
	}
}

func TestParseCommandStderrAppend(t *testing.T) {
	cmd := parseCommand([]string{"cmd", "2", ">>", "err.log"})
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	if cmd.redirs[0].fd != 2 {
		t.Errorf("expected fd 2, got %d", cmd.redirs[0].fd)
	}
	if !cmd.redirs[0].append {
		t.Errorf("expected append=true")
	}
}

func TestParseCommandDupStderr(t *testing.T) {
	// Tokenizer splits "2>&1" into "2", ">", "&", "1"
	cmd := parseCommand([]string{"cmd", "2", ">", "&", "1"})
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	r := cmd.redirs[0]
	if r.fd != 2 {
		t.Errorf("expected fd 2, got %d", r.fd)
	}
	if r.dup != 1 {
		t.Errorf("expected dup to 1, got %d", r.dup)
	}
}

func TestParseCommandDupStdoutToStderr(t *testing.T) {
	// 1>&2
	cmd := parseCommand([]string{"cmd", "1", ">", "&", "2"})
	if len(cmd.redirs) != 1 {
		t.Fatalf("expected 1 redir, got %d", len(cmd.redirs))
	}
	r := cmd.redirs[0]
	if r.fd != 1 {
		t.Errorf("expected fd 1, got %d", r.fd)
	}
	if r.dup != 2 {
		t.Errorf("expected dup to 2, got %d", r.dup)
	}
}

func TestParseCommandMultipleRedirs(t *testing.T) {
	// Tokenizer splits each redir op into its own token
	cmd := parseCommand([]string{"cmd", ">", "out", "2", ">", "err", "<", "in"})
	if len(cmd.args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %#v", len(cmd.args), cmd.args)
	}
	if len(cmd.redirs) != 3 {
		t.Fatalf("expected 3 redirs, got %d", len(cmd.redirs))
	}
}

func TestParseSegment(t *testing.T) {
	cmds := parseSegment([]string{"ls", "-la", "|", "grep", "foo"})
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if len(cmds[0].args) != 2 || cmds[0].args[0] != "ls" {
		t.Errorf("expected ls -la, got %#v", cmds[0].args)
	}
	if len(cmds[1].args) != 2 || cmds[1].args[0] != "grep" {
		t.Errorf("expected grep foo, got %#v", cmds[1].args)
	}
}

func TestParseSegmentTriplePipe(t *testing.T) {
	cmds := parseSegment([]string{"a", "|", "b", "|", "c"})
	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cmds))
	}
	if cmds[0].args[0] != "a" || cmds[1].args[0] != "b" || cmds[2].args[0] != "c" {
		t.Errorf("unexpected command order: %#v", cmds)
	}
}

func TestParseSegmentNoPipe(t *testing.T) {
	cmds := parseSegment([]string{"echo", "hi"})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
}

func TestParseSegmentRedirectInMiddle(t *testing.T) {
	cmds := parseSegment([]string{"echo", "hello", ">", "file", "|", "wc"})
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if len(cmds[0].redirs) != 1 {
		t.Errorf("expected 1 redir on first cmd, got %d", len(cmds[0].redirs))
	}
}

func TestParseSegmentLeadingPipe(t *testing.T) {
	cmds := parseSegment([]string{"|", "wc"})
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if len(cmds[0].args) != 0 {
		t.Errorf("expected empty first command, got %#v", cmds[0].args)
	}
}

func TestParseGroups(t *testing.T) {
	grps := parseGroups([]string{"a", ";", "b"})
	if len(grps) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(grps))
	}
	if grps[0].next != connSemi {
		t.Errorf("expected connSemi, got %v", grps[0].next)
	}
}

func TestParseGroupsAndOr(t *testing.T) {
	grps := parseGroups([]string{"a", "&&", "b", "||", "c"})
	if len(grps) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(grps))
	}
	if grps[0].next != connAnd {
		t.Errorf("group 0 expected connAnd, got %v", grps[0].next)
	}
	if grps[1].next != connOr {
		t.Errorf("group 1 expected connOr, got %v", grps[1].next)
	}
	if grps[2].next != connEnd {
		t.Errorf("group 2 expected connEnd, got %v", grps[2].next)
	}
}

func TestParseGroupsBackground(t *testing.T) {
	grps := parseGroups([]string{"sleep", "10", "&"})
	if len(grps) != 1 {
		t.Fatalf("expected 1 group, got %d", len(grps))
	}
	if grps[0].next != connBg {
		t.Errorf("expected connBg, got %v", grps[0].next)
	}
	if len(grps[0].cmds) != 1 || grps[0].cmds[0].args[0] != "sleep" {
		t.Errorf("expected sleep command, got %#v", grps[0].cmds)
	}
}

func TestIsOperator(t *testing.T) {
	for _, op := range []string{";", "&&", "||", "&", "|"} {
		if !isOperator(op) {
			t.Errorf("isOperator(%q) = false, want true", op)
		}
	}
	if isOperator("echo") {
		t.Errorf("isOperator('echo') = true, want false")
	}
}

func TestBuiltinsMap(t *testing.T) {
	expected := []string{"cd", "exit", "quit", "echo", "pwd", "type", "export", "unset",
		"history", "help", "alias", "unalias", "confirm",
		"from-json", "from-csv", "to-json", "to-csv",
		"where", "sort-by", "select",
		"first", "last", "count", "uniq"}
	for _, name := range expected {
		if _, ok := builtins[name]; !ok {
			t.Errorf("builtins map missing %q", name)
		}
	}
}

func TestIsDataBuiltin(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"from-json", true},
		{"from-csv", true},
		{"to-json", true},
		{"to-csv", true},
		{"where", true},
		{"sort-by", true},
		{"select", true},
		{"ls", false},
		{"cd", false},
		{"echo", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isDataBuiltin(tt.name); got != tt.want {
			t.Errorf("isDataBuiltin(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParseCommandDataFlag(t *testing.T) {
	cmd := parseCommand([]string{"from-json", "file.json"})
	if !cmd.isData {
		t.Errorf("expected isData=true for from-json")
	}

	cmd = parseCommand([]string{"ls", "-la"})
	if cmd.isData {
		t.Errorf("expected isData=false for ls")
	}
}

func TestHandleFromJSON(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"alice","age":30}]`
	stdin := strings.NewReader(input)
	out, code := s.handleFromJSON(nil, stdin)
	if code != 0 {
		t.Fatalf("handleFromJSON exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleWhere(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"alice","age":30},{"name":"bob","age":25}]`
	stdin := strings.NewReader(input)
	out, code := s.handleWhere([]string{"age > 28"}, stdin)
	if code != 0 {
		t.Fatalf("handleWhere exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleFromCSV(t *testing.T) {
	s := &Shell{}
	input := "name,age\nalice,30\nbob,25\n"
	stdin := strings.NewReader(input)
	out, code := s.handleFromCSV(nil, stdin)
	if code != 0 {
		t.Fatalf("handleFromCSV exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleToJSON(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"alice","age":30}]`
	stdin := strings.NewReader(input)
	out, code := s.handleToJSON(nil, stdin)
	if code != 0 {
		t.Fatalf("handleToJSON exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleToCSV(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"alice","age":30}]`
	stdin := strings.NewReader(input)
	out, code := s.handleToCSV(nil, stdin)
	if code != 0 {
		t.Fatalf("handleToCSV exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleSortBy(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"bob","age":25},{"name":"alice","age":30}]`
	stdin := strings.NewReader(input)
	out, code := s.handleSortBy([]string{"name"}, stdin)
	if code != 0 {
		t.Fatalf("handleSortBy exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleSelect(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"alice","age":30,"city":"nyc"}]`
	stdin := strings.NewReader(input)
	out, code := s.handleSelect([]string{"name"}, stdin)
	if code != 0 {
		t.Fatalf("handleSelect exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleFromJSONNoArgs(t *testing.T) {
	s := &Shell{}
	// with no file arg, reads from stdin
	input := `[{"x":1}]`
	stdin := strings.NewReader(input)
	out, code := s.handleFromJSON(nil, stdin)
	if code != 0 {
		t.Fatalf("handleFromJSON exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleWhereNoMatch(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"alice","age":30}]`
	stdin := strings.NewReader(input)
	out, code := s.handleWhere([]string{"age > 50"}, stdin)
	if code != 0 {
		t.Fatalf("handleWhere exit code %d", code)
	}
	if out != "[]" {
		t.Errorf("expected empty array, got %q", out)
	}
}

func TestHandleSortByDesc(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"bob","age":25},{"name":"alice","age":30}]`
	stdin := strings.NewReader(input)
	out, code := s.handleSortBy([]string{"age", "--desc"}, stdin)
	if code != 0 {
		t.Fatalf("handleSortBy exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestParseCommandDataFlagCase(t *testing.T) {
	for _, name := range []string{"where", "sort-by", "select", "from-csv", "to-csv"} {
		cmd := parseCommand([]string{name, "arg"})
		if !cmd.isData {
			t.Errorf("expected isData=true for %s", name)
		}
	}
}

func TestHandleFirst(t *testing.T) {
	s := &Shell{}
	input := `[{"n":1},{"n":2},{"n":3}]`
	stdin := strings.NewReader(input)
	out, code := s.handleFirst([]string{"2"}, stdin)
	if code != 0 {
		t.Fatalf("handleFirst exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleLast(t *testing.T) {
	s := &Shell{}
	input := `[{"n":1},{"n":2},{"n":3}]`
	stdin := strings.NewReader(input)
	out, code := s.handleLast([]string{"2"}, stdin)
	if code != 0 {
		t.Fatalf("handleLast exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleCount(t *testing.T) {
	s := &Shell{}
	input := `[{"n":1},{"n":2},{"n":3}]`
	stdin := strings.NewReader(input)
	out, code := s.handleCount(nil, stdin)
	if code != 0 {
		t.Fatalf("handleCount exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleCountGroupBy(t *testing.T) {
	s := &Shell{}
	input := `[{"city":"nyc","name":"a"},{"city":"nyc","name":"b"},{"city":"sf","name":"c"}]`
	stdin := strings.NewReader(input)
	out, code := s.handleCount([]string{"city"}, stdin)
	if code != 0 {
		t.Fatalf("handleCount exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleUniq(t *testing.T) {
	s := &Shell{}
	input := `[{"name":"alice"},{"name":"bob"},{"name":"alice"}]`
	stdin := strings.NewReader(input)
	out, code := s.handleUniq([]string{"name"}, stdin)
	if code != 0 {
		t.Fatalf("handleUniq exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestHandleFirstDefault(t *testing.T) {
	s := &Shell{}
	input := `[{"n":1},{"n":2},{"n":3},{"n":4},{"n":5},{"n":6},{"n":7},{"n":8},{"n":9},{"n":10},{"n":11}]`
	stdin := strings.NewReader(input)
	out, code := s.handleFirst(nil, stdin)
	if code != 0 {
		t.Fatalf("handleFirst exit code %d", code)
	}
	if out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestIsDataBuiltinNew(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"first", true},
		{"last", true},
		{"count", true},
		{"uniq", true},
		{"confirm", false},
	}
	for _, tt := range tests {
		if got := isDataBuiltin(tt.name); got != tt.want {
			t.Errorf("isDataBuiltin(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
