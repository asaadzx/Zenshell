package data

import (
	"testing"
)

func makeTestTable() TableValue {
	return TableValue{
		Columns: []string{"name", "age", "city"},
		Rows: []RecordValue{
			{Fields: map[string]Value{"name": StringValue{"alice"}, "age": IntValue{30}, "city": StringValue{"nyc"}}},
			{Fields: map[string]Value{"name": StringValue{"bob"}, "age": IntValue{25}, "city": StringValue{"sf"}}},
			{Fields: map[string]Value{"name": StringValue{"charlie"}, "age": IntValue{35}, "city": StringValue{"nyc"}}},
		},
	}
}

func TestFilterEqual(t *testing.T) {
	tbl := makeTestTable()
	conds := []Condition{{Field: "city", Op: "==", Value: StringValue{"nyc"}}}
	result, err := tbl.Filter(conds)
	if err != nil {
		t.Fatalf("filter error: %v", err)
	}
	if len(result.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestFilterGreater(t *testing.T) {
	tbl := makeTestTable()
	conds := []Condition{{Field: "age", Op: ">", Value: IntValue{28}}}
	result, err := tbl.Filter(conds)
	if err != nil {
		t.Fatalf("filter error: %v", err)
	}
	if len(result.Rows) != 2 {
		t.Errorf("expected 2 rows (alice, charlie), got %d", len(result.Rows))
	}
}

func TestFilterMultiple(t *testing.T) {
	tbl := makeTestTable()
	conds := []Condition{
		{Field: "city", Op: "==", Value: StringValue{"nyc"}},
		{Field: "age", Op: ">", Value: IntValue{28}},
	}
	result, err := tbl.Filter(conds)
	if err != nil {
		t.Fatalf("filter error: %v", err)
	}
	if len(result.Rows) != 2 {
		t.Errorf("expected 2 rows (alice, charlie), got %d", len(result.Rows))
	}
}

func TestFilterAgeUnder30(t *testing.T) {
	tbl := makeTestTable()
	conds := []Condition{
		{Field: "city", Op: "==", Value: StringValue{"nyc"}},
		{Field: "age", Op: "<", Value: IntValue{30}},
	}
	result, err := tbl.Filter(conds)
	if err != nil {
		t.Fatalf("filter error: %v", err)
	}
	if len(result.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(result.Rows))
	}
}

func TestSortAsc(t *testing.T) {
	tbl := makeTestTable()
	result, err := tbl.SortBy("age", false)
	if err != nil {
		t.Fatalf("sort error: %v", err)
	}
	if len(result.Rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(result.Rows))
	}
	if result.Rows[0].Fields["name"].String() != "bob" {
		t.Errorf("expected first row 'bob', got %s", result.Rows[0].Fields["name"].String())
	}
	if result.Rows[2].Fields["name"].String() != "charlie" {
		t.Errorf("expected last row 'charlie', got %s", result.Rows[2].Fields["name"].String())
	}
}

func TestSortDesc(t *testing.T) {
	tbl := makeTestTable()
	result, err := tbl.SortBy("age", true)
	if err != nil {
		t.Fatalf("sort error: %v", err)
	}
	if result.Rows[0].Fields["name"].String() != "charlie" {
		t.Errorf("expected first row 'charlie', got %s", result.Rows[0].Fields["name"].String())
	}
}

func TestSortString(t *testing.T) {
	tbl := makeTestTable()
	result, err := tbl.SortBy("name", false)
	if err != nil {
		t.Fatalf("sort error: %v", err)
	}
	if result.Rows[0].Fields["name"].String() != "alice" {
		t.Errorf("expected first 'alice', got %s", result.Rows[0].Fields["name"].String())
	}
}

func TestSortInvalidField(t *testing.T) {
	tbl := makeTestTable()
	_, err := tbl.SortBy("nonexistent", false)
	if err == nil {
		t.Errorf("expected error for invalid field")
	}
}

func TestSelect(t *testing.T) {
	tbl := makeTestTable()
	result, err := tbl.Select([]string{"name", "age"})
	if err != nil {
		t.Fatalf("select error: %v", err)
	}
	if len(result.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(result.Columns))
	}
	if result.Columns[0] != "name" || result.Columns[1] != "age" {
		t.Errorf("expected [name age], got %v", result.Columns)
	}
}

func TestSelectInvalidField(t *testing.T) {
	tbl := makeTestTable()
	_, err := tbl.Select([]string{"nope"})
	if err == nil {
		t.Errorf("expected error for invalid column")
	}
}

func TestFilterEmpty(t *testing.T) {
	tbl := makeTestTable()
	result, err := tbl.Filter(nil)
	if err != nil {
		t.Fatalf("filter error: %v", err)
	}
	if len(result.Rows) != 3 {
		t.Errorf("expected 3 rows with no filter, got %d", len(result.Rows))
	}
}

func TestFilterNoMatch(t *testing.T) {
	tbl := makeTestTable()
	conds := []Condition{{Field: "age", Op: ">", Value: IntValue{100}}}
	result, err := tbl.Filter(conds)
	if err != nil {
		t.Fatalf("filter error: %v", err)
	}
	if len(result.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(result.Rows))
	}
}

func TestParseCondition(t *testing.T) {
	cond, err := ParseCondition("age > 20")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if cond.Field != "age" || cond.Op != ">" {
		t.Errorf("expected (age, >), got (%s, %s)", cond.Field, cond.Op)
	}
	if v, ok := cond.Value.(IntValue); !ok || v.Value != 20 {
		t.Errorf("expected IntValue(20), got %#v", cond.Value)
	}
}

func TestParseConditionString(t *testing.T) {
	cond, err := ParseCondition(`name == "alice"`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if cond.Field != "name" || cond.Op != "==" {
		t.Errorf("expected (name, ==), got (%s, %s)", cond.Field, cond.Op)
	}
}

func TestParseConditionInvalid(t *testing.T) {
	_, err := ParseCondition("not a condition")
	if err == nil {
		t.Errorf("expected error for invalid condition")
	}
}

func TestParseConditionRegex(t *testing.T) {
	cond, err := ParseCondition("name ~= foo")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if cond.Op != "~=" {
		t.Errorf("expected ~=, got %s", cond.Op)
	}
}

func TestTableDisplay(t *testing.T) {
	tbl := makeTestTable()
	s := tbl.String()
	if s == "" {
		t.Errorf("expected non-empty display")
	}
	if len(tbl.Rows) != 3 {
		t.Errorf("expected 3 rows")
	}
}

func TestFirst(t *testing.T) {
	tbl := makeTestTable()
	r := tbl.First(2)
	if len(r.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(r.Rows))
	}
	if name, ok := r.Rows[0].Fields["name"]; !ok || name.String() != "alice" {
		t.Errorf("expected first row alice, got %v", name)
	}
}

func TestFirstZero(t *testing.T) {
	tbl := makeTestTable()
	r := tbl.First(0)
	if len(r.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(r.Rows))
	}
}

func TestFirstExceedsLength(t *testing.T) {
	tbl := makeTestTable()
	r := tbl.First(100)
	if len(r.Rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(r.Rows))
	}
}

func TestLast(t *testing.T) {
	tbl := makeTestTable()
	r := tbl.Last(2)
	if len(r.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(r.Rows))
	}
	if name, ok := r.Rows[0].Fields["name"]; !ok || name.String() != "bob" {
		t.Errorf("expected first row bob, got %v", name)
	}
}

func TestLastZero(t *testing.T) {
	tbl := makeTestTable()
	r := tbl.Last(0)
	if len(r.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(r.Rows))
	}
}

func TestUnique(t *testing.T) {
	tbl := TableValue{
		Columns: []string{"name", "city"},
		Rows: []RecordValue{
			{Fields: map[string]Value{"name": StringValue{"alice"}, "city": StringValue{"nyc"}}},
			{Fields: map[string]Value{"name": StringValue{"bob"}, "city": StringValue{"sf"}}},
			{Fields: map[string]Value{"name": StringValue{"alice"}, "city": StringValue{"nyc"}}},
		},
	}
	r := tbl.Unique([]string{"name", "city"})
	if len(r.Rows) != 2 {
		t.Errorf("expected 2 unique rows, got %d", len(r.Rows))
	}
}

func TestUniqueField(t *testing.T) {
	tbl := TableValue{
		Columns: []string{"name", "city"},
		Rows: []RecordValue{
			{Fields: map[string]Value{"name": StringValue{"alice"}, "city": StringValue{"nyc"}}},
			{Fields: map[string]Value{"name": StringValue{"bob"}, "city": StringValue{"nyc"}}},
			{Fields: map[string]Value{"name": StringValue{"charlie"}, "city": StringValue{"nyc"}}},
		},
	}
	r := tbl.Unique([]string{"city"})
	if len(r.Rows) != 1 {
		t.Errorf("expected 1 unique city, got %d", len(r.Rows))
	}
}

func TestGroupBy(t *testing.T) {
	tbl := makeTestTable()
	r := tbl.GroupBy([]string{"city"})
	if len(r.Rows) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(r.Rows))
	}
	foundNYC := false
	foundSF := false
	for _, row := range r.Rows {
		city, _ := row.Fields["city"]
		count, _ := row.Fields["count"]
		if city.String() == "nyc" {
			foundNYC = true
			if count.String() != "2" {
				t.Errorf("expected nyc count 2, got %s", count.String())
			}
		}
		if city.String() == "sf" {
			foundSF = true
			if count.String() != "1" {
				t.Errorf("expected sf count 1, got %s", count.String())
			}
		}
	}
	if !foundNYC || !foundSF {
		t.Errorf("missing groups: nyc=%v sf=%v", foundNYC, foundSF)
	}
}

func TestAggregateSum(t *testing.T) {
	tbl := makeTestTable()
	val, err := tbl.Aggregate("age", "sum")
	if err != nil {
		t.Fatalf("sum error: %v", err)
	}
	if val.String() != "90" {
		t.Errorf("expected sum 90, got %s", val.String())
	}
}

func TestAggregateAvg(t *testing.T) {
	tbl := makeTestTable()
	val, err := tbl.Aggregate("age", "avg")
	if err != nil {
		t.Fatalf("avg error: %v", err)
	}
	if val.String() != "30" {
		t.Errorf("expected avg 30, got %s", val.String())
	}
}

func TestAggregateMin(t *testing.T) {
	tbl := makeTestTable()
	val, err := tbl.Aggregate("age", "min")
	if err != nil {
		t.Fatalf("min error: %v", err)
	}
	if val.String() != "25" {
		t.Errorf("expected min 25, got %s", val.String())
	}
}

func TestAggregateMax(t *testing.T) {
	tbl := makeTestTable()
	val, err := tbl.Aggregate("age", "max")
	if err != nil {
		t.Fatalf("max error: %v", err)
	}
	if val.String() != "35" {
		t.Errorf("expected max 35, got %s", val.String())
	}
}

func TestAggregateEmpty(t *testing.T) {
	tbl := TableValue{Columns: []string{"x"}}
	_, err := tbl.Aggregate("x", "sum")
	if err == nil {
		t.Errorf("expected error on empty table")
	}
}

func TestFirstEmptyTable(t *testing.T) {
	tbl := TableValue{Columns: []string{"a"}}
	r := tbl.First(5)
	if len(r.Rows) != 0 {
		t.Errorf("expected 0 rows from empty table")
	}
}

func TestUniqueEmptyTable(t *testing.T) {
	tbl := TableValue{Columns: []string{"a"}}
	r := tbl.Unique(nil)
	if len(r.Rows) != 0 {
		t.Errorf("expected 0 rows")
	}
}

func TestGroupByEmptyTable(t *testing.T) {
	tbl := TableValue{Columns: []string{"city"}}
	r := tbl.GroupBy([]string{"city"})
	if len(r.Rows) != 0 {
		t.Errorf("expected 0 groups")
	}
}

func TestAggregateCount(t *testing.T) {
	tbl := makeTestTable()
	val, err := tbl.Aggregate("age", "count")
	if err != nil {
		t.Fatalf("count error: %v", err)
	}
	if val.String() != "3" {
		t.Errorf("expected count 3, got %s", val.String())
	}
}
