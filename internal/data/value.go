package data

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Kind int

const (
	KindString Kind = iota
	KindInt
	KindFloat
	KindBool
	KindList
	KindRecord
	KindTable
)

func (k Kind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindBool:
		return "bool"
	case KindList:
		return "list"
	case KindRecord:
		return "record"
	case KindTable:
		return "table"
	default:
		return "unknown"
	}
}

type Value interface {
	Kind() Kind
	String() string
	Display() string
	Compare(op string, other Value) (bool, error)
}

type StringValue struct{ Value string }

func (v StringValue) Kind() Kind          { return KindString }
func (v StringValue) String() string       { return v.Value }
func (v StringValue) Display() string      { return v.Value }
func (v StringValue) GoString() string     { return fmt.Sprintf("%q", v.Value) }

func (v StringValue) Compare(op string, other Value) (bool, error) {
	s2, ok := other.(StringValue)
	if !ok {
		return false, fmt.Errorf("cannot compare string with %s", other.Kind())
	}
	switch op {
	case "==":
		return v.Value == s2.Value, nil
	case "!=":
		return v.Value != s2.Value, nil
	case "<":
		return v.Value < s2.Value, nil
	case "<=":
		return v.Value <= s2.Value, nil
	case ">":
		return v.Value > s2.Value, nil
	case ">=":
		return v.Value >= s2.Value, nil
	case "~=":
		matched, err := regexp.MatchString(s2.Value, v.Value)
		if err != nil {
			return false, fmt.Errorf("invalid regex %q: %w", s2.Value, err)
		}
		return matched, nil
	case "in":
		return false, fmt.Errorf("'in' requires list operand")
	default:
		return false, fmt.Errorf("unknown operator %q", op)
	}
}

type IntValue struct{ Value int64 }

func (v IntValue) Kind() Kind          { return KindInt }
func (v IntValue) String() string       { return strconv.FormatInt(v.Value, 10) }
func (v IntValue) Display() string      { return strconv.FormatInt(v.Value, 10) }

func (v IntValue) Compare(op string, other Value) (bool, error) {
	switch o := other.(type) {
	case IntValue:
		switch op {
		case "==":
			return v.Value == o.Value, nil
		case "!=":
			return v.Value != o.Value, nil
		case "<":
			return v.Value < o.Value, nil
		case "<=":
			return v.Value <= o.Value, nil
		case ">":
			return v.Value > o.Value, nil
		case ">=":
			return v.Value >= o.Value, nil
		default:
			return false, fmt.Errorf("unknown operator %q for int", op)
		}
	case FloatValue:
		a := float64(v.Value)
		switch op {
		case "==":
			return a == o.Value, nil
		case "!=":
			return a != o.Value, nil
		case "<":
			return a < o.Value, nil
		case "<=":
			return a <= o.Value, nil
		case ">":
			return a > o.Value, nil
		case ">=":
			return a >= o.Value, nil
		default:
			return false, fmt.Errorf("unknown operator %q for int/float", op)
		}
	default:
		return false, fmt.Errorf("cannot compare int with %s", other.Kind())
	}
}

type FloatValue struct{ Value float64 }

func (v FloatValue) Kind() Kind          { return KindFloat }
func (v FloatValue) String() string       { return strconv.FormatFloat(v.Value, 'f', -1, 64) }
func (v FloatValue) Display() string      { return strconv.FormatFloat(v.Value, 'f', 2, 64) }

func (v FloatValue) Compare(op string, other Value) (bool, error) {
	switch o := other.(type) {
	case FloatValue:
		switch op {
		case "==":
			return v.Value == o.Value, nil
		case "!=":
			return v.Value != o.Value, nil
		case "<":
			return v.Value < o.Value, nil
		case "<=":
			return v.Value <= o.Value, nil
		case ">":
			return v.Value > o.Value, nil
		case ">=":
			return v.Value >= o.Value, nil
		default:
			return false, fmt.Errorf("unknown operator %q for float", op)
		}
	case IntValue:
		a := v.Value
		b := float64(o.Value)
		switch op {
		case "==":
			return a == b, nil
		case "!=":
			return a != b, nil
		case "<":
			return a < b, nil
		case "<=":
			return a <= b, nil
		case ">":
			return a > b, nil
		case ">=":
			return a >= b, nil
		default:
			return false, fmt.Errorf("unknown operator %q for float/int", op)
		}
	default:
		return false, fmt.Errorf("cannot compare float with %s", other.Kind())
	}
}

type BoolValue struct{ Value bool }

func (v BoolValue) Kind() Kind          { return KindBool }
func (v BoolValue) String() string       { return strconv.FormatBool(v.Value) }
func (v BoolValue) Display() string      { return strconv.FormatBool(v.Value) }

func (v BoolValue) Compare(op string, other Value) (bool, error) {
	o, ok := other.(BoolValue)
	if !ok {
		return false, fmt.Errorf("cannot compare bool with %s", other.Kind())
	}
	switch op {
	case "==":
		return v.Value == o.Value, nil
	case "!=":
		return v.Value != o.Value, nil
	default:
		return false, fmt.Errorf("operator %q not supported for bool", op)
	}
}

type ListValue struct{ Values []Value }

func (v ListValue) Kind() Kind  { return KindList }
func (v ListValue) String() string {
	parts := make([]string, len(v.Values))
	for i, val := range v.Values {
		parts[i] = val.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
func (v ListValue) Display() string { return v.String() }

func (v ListValue) Compare(op string, other Value) (bool, error) {
	switch op {
	case "in":
		for _, item := range v.Values {
			eq, err := item.Compare("==", other)
			if err != nil {
				return false, err
			}
			if eq {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("operator %q not supported for list", op)
	}
}

type RecordValue struct {
	Fields map[string]Value
}

func (v RecordValue) Kind() Kind  { return KindRecord }
func (v RecordValue) String() string {
	return v.Display()
}
func (v RecordValue) Display() string {
	parts := make([]string, 0, len(v.Fields))
	for k, val := range v.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", k, val.Display()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
func (v RecordValue) Compare(op string, other Value) (bool, error) {
	return false, fmt.Errorf("cannot compare records directly")
}

type TableValue struct {
	Columns []string
	Rows    []RecordValue
}

func (v TableValue) Kind() Kind  { return KindTable }
func (v TableValue) String() string {
	if len(v.Rows) == 0 {
		return "(empty table)"
	}
	var b strings.Builder
	// Header
	for i, col := range v.Columns {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(col)
	}
	b.WriteString("\n")
	// Separator
	for i := range v.Columns {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(strings.Repeat("-", len(v.Columns[i])))
	}
	b.WriteString("\n")
	// Rows
	for _, row := range v.Rows {
		for i, col := range v.Columns {
			if i > 0 {
				b.WriteString("  ")
			}
			if val, ok := row.Fields[col]; ok {
				b.WriteString(val.Display())
			} else {
				b.WriteString("-")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}
func (v TableValue) Display() string { return v.String() }
func (v TableValue) Compare(op string, other Value) (bool, error) {
	return false, fmt.Errorf("cannot compare tables directly")
}

// ParseValue attempts to parse a string into the best matching Value type.
func ParseValue(s string) Value {
	if s == "" {
		return StringValue{Value: s}
	}
	if s == "true" {
		return BoolValue{Value: true}
	}
	if s == "false" {
		return BoolValue{Value: false}
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return IntValue{Value: n}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return FloatValue{Value: f}
	}
	return StringValue{Value: s}
}
