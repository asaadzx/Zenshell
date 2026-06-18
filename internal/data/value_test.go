package data

import (
	"testing"
)

func TestStringValue(t *testing.T) {
	v := StringValue{Value: "hello"}
	if v.Kind() != KindString {
		t.Errorf("expected KindString, got %v", v.Kind())
	}
	if v.String() != "hello" {
		t.Errorf("expected 'hello', got %q", v.String())
	}
}

func TestIntValue(t *testing.T) {
	v := IntValue{Value: 42}
	if v.Kind() != KindInt {
		t.Errorf("expected KindInt, got %v", v.Kind())
	}
	if v.String() != "42" {
		t.Errorf("expected '42', got %q", v.String())
	}
}

func TestFloatValue(t *testing.T) {
	v := FloatValue{Value: 3.14}
	if v.Kind() != KindFloat {
		t.Errorf("expected KindFloat, got %v", v.Kind())
	}
}

func TestBoolValue(t *testing.T) {
	v := BoolValue{Value: true}
	if v.Kind() != KindBool {
		t.Errorf("expected KindBool, got %v", v.Kind())
	}
	if v.String() != "true" {
		t.Errorf("expected 'true', got %q", v.String())
	}
}

func TestListValue(t *testing.T) {
	v := ListValue{Values: []Value{StringValue{"a"}, StringValue{"b"}}}
	if v.Kind() != KindList {
		t.Errorf("expected KindList, got %v", v.Kind())
	}
}

func TestRecordValue(t *testing.T) {
	v := RecordValue{Fields: map[string]Value{"name": StringValue{"foo"}}}
	if v.Kind() != KindRecord {
		t.Errorf("expected KindRecord, got %v", v.Kind())
	}
}

func TestTableValue(t *testing.T) {
	v := TableValue{
		Columns: []string{"name", "age"},
		Rows: []RecordValue{
			{Fields: map[string]Value{"name": StringValue{"alice"}, "age": IntValue{30}}},
		},
	}
	if v.Kind() != KindTable {
		t.Errorf("expected KindTable, got %v", v.Kind())
	}
}

func TestCompareString(t *testing.T) {
	a := StringValue{"apple"}
	b := StringValue{"banana"}
	eq, _ := a.Compare("==", StringValue{"apple"})
	if !eq {
		t.Errorf("expected apple == apple")
	}
	lt, _ := a.Compare("<", b)
	if !lt {
		t.Errorf("expected apple < banana")
	}
}

func TestCompareInt(t *testing.T) {
	a := IntValue{10}
	b := IntValue{20}
	eq, _ := a.Compare("==", IntValue{10})
	if !eq {
		t.Errorf("expected 10 == 10")
	}
	lt, _ := a.Compare("<", b)
	if !lt {
		t.Errorf("expected 10 < 20")
	}
	gt, _ := b.Compare(">", a)
	if !gt {
		t.Errorf("expected 20 > 10")
	}
}

func TestCompareFloat(t *testing.T) {
	a := FloatValue{1.5}
	b := FloatValue{2.5}
	lt, _ := a.Compare("<", b)
	if !lt {
		t.Errorf("expected 1.5 < 2.5")
	}
}

func TestCompareIntFloat(t *testing.T) {
	a := IntValue{5}
	b := FloatValue{5.0}
	eq, _ := a.Compare("==", b)
	if !eq {
		t.Errorf("expected 5 == 5.0")
	}
}

func TestCompareRegex(t *testing.T) {
	a := StringValue{"hello world"}
	matched, err := a.Compare("~=", StringValue{"world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matched {
		t.Errorf("expected 'hello world' ~= 'world'")
	}
}

func TestCompareListIn(t *testing.T) {
	list := ListValue{Values: []Value{StringValue{"a"}, StringValue{"b"}, StringValue{"c"}}}
	found, err := list.Compare("in", StringValue{"b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Errorf("expected 'b' in [a, b, c]")
	}
	notFound, _ := list.Compare("in", StringValue{"z"})
	if notFound {
		t.Errorf("expected 'z' not in [a, b, c]")
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		input string
		kind  Kind
	}{
		{"hello", KindString},
		{"42", KindInt},
		{"3.14", KindFloat},
		{"true", KindBool},
		{"false", KindBool},
		{"", KindString},
	}
	for _, tt := range tests {
		v := ParseValue(tt.input)
		if v.Kind() != tt.kind {
			t.Errorf("ParseValue(%q) = %v, want %v", tt.input, v.Kind(), tt.kind)
		}
	}
}
