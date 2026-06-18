package data

import (
	"strings"
	"testing"
)

func makeJSONTable() TableValue {
	return TableValue{
		Columns: []string{"name", "age", "active"},
		Rows: []RecordValue{
			{Fields: map[string]Value{"name": StringValue{"alice"}, "age": IntValue{30}, "active": BoolValue{true}}},
			{Fields: map[string]Value{"name": StringValue{"bob"}, "age": IntValue{25}, "active": BoolValue{false}}},
		},
	}
}

func TestToJSON(t *testing.T) {
	tbl := makeJSONTable()
	json, err := tbl.ToJSON()
	if err != nil {
		t.Fatalf("to-json error: %v", err)
	}
	if !strings.Contains(json, "alice") || !strings.Contains(json, "bob") {
		t.Errorf("expected alice and bob in json, got:\n%s", json)
	}
	if !strings.Contains(json, "30") || !strings.Contains(json, "25") {
		t.Errorf("expected ages in json")
	}
}

func TestFromJSON(t *testing.T) {
	input := `[
		{"name": "alice", "age": 30, "active": true},
		{"name": "bob", "age": 25, "active": false}
	]`
	tbl, err := FromJSON(input)
	if err != nil {
		t.Fatalf("from-json error: %v", err)
	}
	if len(tbl.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tbl.Rows))
	}
	if len(tbl.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(tbl.Columns))
	}
}

func TestJSONRoundTrip(t *testing.T) {
	original := makeJSONTable()
	json, err := original.ToJSON()
	if err != nil {
		t.Fatalf("to-json error: %v", err)
	}
	restored, err := FromJSON(json)
	if err != nil {
		t.Fatalf("from-json error: %v", err)
	}
	if len(restored.Rows) != len(original.Rows) {
		t.Errorf("row count mismatch: %d vs %d", len(restored.Rows), len(original.Rows))
	}
}

func TestToJSONEmpty(t *testing.T) {
	tbl := TableValue{Columns: []string{"a", "b"}, Rows: nil}
	json, err := tbl.ToJSON()
	if err != nil {
		t.Fatalf("to-json error: %v", err)
	}
	if json != "[]" {
		t.Errorf("expected '[]', got %q", json)
	}
}

func TestToCSV(t *testing.T) {
	tbl := makeJSONTable()
	csv, err := tbl.ToCSV()
	if err != nil {
		t.Fatalf("to-csv error: %v", err)
	}
	if !strings.Contains(csv, "name,age,active") {
		t.Errorf("expected header, got:\n%s", csv)
	}
}

func TestFromCSV(t *testing.T) {
	input := "name,age\nfoo,10\nbar,20\n"
	tbl, err := FromCSV(input)
	if err != nil {
		t.Fatalf("from-csv error: %v", err)
	}
	if len(tbl.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tbl.Rows))
	}
}

func TestCSVRoundTrip(t *testing.T) {
	original := makeJSONTable()
	csv, err := original.ToCSV()
	if err != nil {
		t.Fatalf("to-csv error: %v", err)
	}
	restored, err := FromCSV(csv)
	if err != nil {
		t.Fatalf("from-csv error: %v", err)
	}
	if len(restored.Rows) != len(original.Rows) {
		t.Errorf("row count mismatch: %d vs %d", len(restored.Rows), len(original.Rows))
	}
}

func TestFromJSONEmpty(t *testing.T) {
	tbl, err := FromJSON("[]")
	if err != nil {
		t.Fatalf("from-json error: %v", err)
	}
	if len(tbl.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(tbl.Rows))
	}
}

func TestFromJSONNested(t *testing.T) {
	input := `[{"name": "alice", "tags": ["dev", "ops"]}]`
	tbl, err := FromJSON(input)
	if err != nil {
		t.Fatalf("from-json error: %v", err)
	}
	if len(tbl.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(tbl.Rows))
	}
	tags, ok := tbl.Rows[0].Fields["tags"]
	if !ok {
		t.Fatal("expected tags field")
	}
	if tags.Kind() != KindList {
		t.Errorf("expected list kind, got %v", tags.Kind())
	}
}
