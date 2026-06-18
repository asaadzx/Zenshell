package shell

import (
	"os"
	"strings"
)

func tokenize(input string) []string {
	var tokens []string
	var cur strings.Builder
	i := 0

	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}

	isMeta := func(ch byte) bool {
		return ch == ';' || ch == '|' || ch == '&' || ch == '>' || ch == '<'
	}

	emitOp := func() {
		ch := input[i]
		switch {
		case ch == '&' && i+1 < len(input) && input[i+1] == '&':
			tokens = append(tokens, "&&")
			i += 2
		case ch == '|' && i+1 < len(input) && input[i+1] == '|':
			tokens = append(tokens, "||")
			i += 2
		case ch == '>' && i+1 < len(input) && input[i+1] == '>':
			tokens = append(tokens, ">>")
			i += 2
		case ch == '&' && i+1 < len(input) && input[i+1] == '>':
			if i+2 < len(input) && input[i+2] == '>' {
				tokens = append(tokens, "&>>")
				i += 3
			} else {
				tokens = append(tokens, "&>")
				i += 2
			}
		default:
			tokens = append(tokens, string(ch))
			i++
		}
	}

	for i < len(input) {
		ch := input[i]

		if ch == ' ' || ch == '\t' || ch == '\n' {
			flush()
			i++
			continue
		}

		if isMeta(ch) {
			flush()
			emitOp()
			continue
		}

		switch {
		case ch == '\\':
			if i+1 < len(input) {
				cur.WriteByte(input[i+1])
				i += 2
			} else {
				i++
			}

		case ch == '\'':
			i++
			for i < len(input) && input[i] != '\'' {
				cur.WriteByte(input[i])
				i++
			}
			if i < len(input) {
				i++
			}

		case ch == '"':
			i++
			for i < len(input) && input[i] != '"' {
				if input[i] == '\\' && i+1 < len(input) {
					cur.WriteByte(input[i+1])
					i += 2
					continue
				}
				if input[i] == '$' {
					i++
					cur.WriteString(expandVar(input, &i))
					continue
				}
				cur.WriteByte(input[i])
				i++
			}
			if i < len(input) {
				i++
			}

		case ch == '$':
			i++
			cur.WriteString(expandVar(input, &i))

		case ch == '~' && cur.Len() == 0 && (i+1 >= len(input) || input[i+1] == '/' || input[i+1] == ' ' || input[i+1] == '\t' || isMeta(input[i+1])):
			cur.WriteString(os.Getenv("HOME"))
			i++

		default:
			cur.WriteByte(input[i])
			i++
		}
	}

	flush()
	return tokens
}

func expandVar(input string, i *int) string {
	start := *i
	if *i >= len(input) {
		return "$"
	}

	if input[*i] == '{' {
		*i++
		close := strings.IndexByte(input[*i:], '}')
		if close < 0 {
			return "${" + input[*i:]
		}
		name := input[*i : *i+close]
		*i += close + 1
		if val := os.Getenv(name); val != "" {
			return val
		}
		return ""
	}

	for *i < len(input) && (isAlphaNum(input[*i]) || input[*i] == '_') {
		*i++
	}
	name := input[start:*i]
	if name == "" {
		return "$"
	}
	if val := os.Getenv(name); val != "" {
		return val
	}
	return ""
}

func isAlphaNum(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

// needsContinuation checks if the line has unclosed quotes or a continuation char.
func needsContinuation(line string) bool {
	// Check for trailing unescaped backslash
	if strings.HasSuffix(line, "\\") && !strings.HasSuffix(line, "\\\\") {
		return true
	}

	inSingle := false
	inDouble := false
	escaped := false
	for _, ch := range line {
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inDouble {
			escaped = true
			continue
		}
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}
	}
	return inSingle || inDouble
}
