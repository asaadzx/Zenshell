package shell

import (
	"os"
	"reflect"
	"testing"
)

func eq(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("  got: %#v\n want: %#v", got, want)
	}
}

func TestTokenizeSimple(t *testing.T) {
	eq(t, tokenize("echo hello world"), []string{"echo", "hello", "world"})
	eq(t, tokenize("ls   -la"), []string{"ls", "-la"})
	eq(t, tokenize(""), nil)
	eq(t, tokenize("   "), nil)
}

func TestTokenizeQuotes(t *testing.T) {
	eq(t, tokenize(`echo 'hello world'`), []string{"echo", "hello world"})
	eq(t, tokenize(`echo "hello world"`), []string{"echo", "hello world"})
	eq(t, tokenize(`echo "it's fine"`), []string{"echo", "it's fine"})
}

func TestTokenizeMixedQuotes(t *testing.T) {
	eq(t, tokenize(`echo "foo"'bar'`), []string{"echo", "foobar"})
}

func TestTokenizeEscape(t *testing.T) {
	eq(t, tokenize(`echo hello\ world`), []string{"echo", "hello world"})
	eq(t, tokenize(`echo \\`), []string{"echo", "\\"})
}

func TestTokenizeMeta(t *testing.T) {
	eq(t, tokenize("a;b"), []string{"a", ";", "b"})
	eq(t, tokenize("a | b"), []string{"a", "|", "b"})
	eq(t, tokenize("a&&b"), []string{"a", "&&", "b"})
	eq(t, tokenize("a||b"), []string{"a", "||", "b"})
	eq(t, tokenize("a &"), []string{"a", "&"})
	eq(t, tokenize("a > out"), []string{"a", ">", "out"})
	eq(t, tokenize("a >> out"), []string{"a", ">>", "out"})
	eq(t, tokenize("a &> out"), []string{"a", "&>", "out"})
	eq(t, tokenize("a &>> out"), []string{"a", "&>>", "out"})
	eq(t, tokenize("a < in"), []string{"a", "<", "in"})
}

func TestTokenizeMetaAdjacent(t *testing.T) {
	eq(t, tokenize("a&&b||c"), []string{"a", "&&", "b", "||", "c"})
	eq(t, tokenize("a;b|c"), []string{"a", ";", "b", "|", "c"})
}

func TestTokenizeNewline(t *testing.T) {
	eq(t, tokenize("echo\nhello"), []string{"echo", "hello"})
	eq(t, tokenize("a\nb\nc"), []string{"a", "b", "c"})
}

func TestTokenizeTilde(t *testing.T) {
	home := os.Getenv("HOME")
	result := tokenize("ls ~")
	if len(result) < 2 {
		t.Fatalf("expected at least 2 tokens, got %d", len(result))
	}
	if result[1] != home {
		t.Errorf("expected ~ to expand to %q, got %q", home, result[1])
	}
}

func TestTokenizeTildePath(t *testing.T) {
	home := os.Getenv("HOME")
	result := tokenize("ls ~/foo")
	if len(result) < 2 {
		t.Fatalf("expected at least 2 tokens, got %d", len(result))
	}
	want := home + "/foo"
	if result[1] != want {
		t.Errorf("expected ~/foo to expand to %q, got %q", want, result[1])
	}
}

func TestNeedsContinuation(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"echo hello", false},
		{"echo 'unclosed", true},
		{`echo "unclosed`, true},
		{"echo \\", true},
		{`echo 'closed'`, false},
		{`echo "closed"`, false},
		{`echo 'single" inside'`, false},
	}
	for _, tt := range tests {
		got := needsContinuation(tt.input)
		if got != tt.want {
			t.Errorf("needsContinuation(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestExpandVar(t *testing.T) {
	t.Setenv("TEST_VAR", "hello")
	i := 1
	got := expandVar("$TEST_VAR", &i)
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
	if i != 9 {
		t.Errorf("i = %d, want 9", i)
	}
}

func TestExpandVarBrace(t *testing.T) {
	t.Setenv("TEST_VAR", "world")
	i := 2
	got := expandVar("${TEST_VAR}", &i)
	if got != "world" {
		t.Errorf("got %q, want %q", got, "world")
	}
}

func TestExpandVarUnset(t *testing.T) {
	i := 1
	got := expandVar("$NONEXISTENT_VAR_XYZ", &i)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestExpandVarDollarOnly(t *testing.T) {
	i := 0
	input := "$"
	got := expandVar(input, &i)
	if got != "$" {
		t.Errorf("expected literal $, got %q", got)
	}
}

func TestTokenizeVarExpansion(t *testing.T) {
	t.Setenv("MYVAR", "world")
	eq(t, tokenize("echo hello $MYVAR"), []string{"echo", "hello", "world"})
}

func TestTokenizeVarExpansionUnset(t *testing.T) {
	// Unset variables expand to nothing (dropped, like POSIX shell)
	eq(t, tokenize("echo $NOPE"), []string{"echo"})
}

func TestTokenizeVarExpansionBrace(t *testing.T) {
	t.Setenv("HOME", "/home/user")
	eq(t, tokenize("echo ${HOME}/docs"), []string{"echo", "/home/user/docs"})
}

func TestTokenizeDoubleQuoteEscape(t *testing.T) {
	eq(t, tokenize(`echo "hello\"world"`), []string{"echo", `hello"world`})
}

func TestTokenizeUnsetVarInMiddle(t *testing.T) {
	eq(t, tokenize("echo $NOPE bar"), []string{"echo", "bar"})
}

func TestTokenizePipesAndRedirects(t *testing.T) {
	eq(t, tokenize("cat foo | grep bar > out"), []string{"cat", "foo", "|", "grep", "bar", ">", "out"})
	eq(t, tokenize("cmd 2>&1"), []string{"cmd", "2", ">", "&", "1"})
}
