package data

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Condition represents a filter condition: field op value.
type Condition struct {
	Field string
	Op    string // ==, !=, <, <=, >, >=, ~=, in
	Value Value
}

// Filter returns rows where all conditions match.
func (t TableValue) Filter(conds []Condition) (TableValue, error) {
	if len(conds) == 0 {
		return t, nil
	}

	var filtered []RecordValue
	for _, row := range t.Rows {
		matches := true
		for _, c := range conds {
			val, ok := row.Fields[c.Field]
			if !ok && c.Op != "!=" {
				matches = false
				break
			}
			if !ok {
				// For !=, missing field means it's not equal
				if c.Op == "!=" {
					continue
				}
				matches = false
				break
			}
			result, err := val.Compare(c.Op, c.Value)
			if err != nil {
				return TableValue{}, fmt.Errorf("filter on %s: %w", c.Field, err)
			}
			if !result {
				matches = false
				break
			}
		}
		if matches {
			filtered = append(filtered, row)
		}
	}
	return TableValue{Columns: t.Columns, Rows: filtered}, nil
}

// LessFunc compares two rows for sorting. Returns true if a < b.
type LessFunc func(a, b RecordValue) bool

// SortBy returns rows sorted by the given field and direction.
func (t TableValue) SortBy(field string, desc bool) (TableValue, error) {
	if len(t.Rows) == 0 {
		return t, nil
	}

	// Verify field exists
	found := false
	for _, col := range t.Columns {
		if col == field {
			found = true
			break
		}
	}
	if !found {
		return TableValue{}, fmt.Errorf("no such column: %s", field)
	}

	sorted := make([]RecordValue, len(t.Rows))
	copy(sorted, t.Rows)

	sort.SliceStable(sorted, func(i, j int) bool {
		a := sorted[i].Fields[field]
		b := sorted[j].Fields[field]

		// Compare as values
		less, err := lessThan(a, b)
		if err != nil {
			return false
		}
		if desc {
			return !less && !valuesEqual(a, b)
		}
		return less
	})

	return TableValue{Columns: t.Columns, Rows: sorted}, nil
}

func lessThan(a, b Value) (bool, error) {
	if a == nil || b == nil {
		return false, nil
	}

	switch va := a.(type) {
	case IntValue:
		switch vb := b.(type) {
		case IntValue:
			return va.Value < vb.Value, nil
		case FloatValue:
			return float64(va.Value) < vb.Value, nil
		}
	case FloatValue:
		switch vb := b.(type) {
		case FloatValue:
			return va.Value < vb.Value, nil
		case IntValue:
			return va.Value < float64(vb.Value), nil
		}
	case StringValue:
		if vb, ok := b.(StringValue); ok {
			return va.Value < vb.Value, nil
		}
	}
	return false, fmt.Errorf("cannot compare %s with %s", a.Kind(), b.Kind())
}

func valuesEqual(a, b Value) bool {
	r, err := a.Compare("==", b)
	return err == nil && r
}

// Select returns a new table with only the given columns.
func (t TableValue) Select(cols []string) (TableValue, error) {
	if len(cols) == 0 {
		return t, nil
	}

	// Verify all columns exist
	for _, col := range cols {
		found := false
		for _, c := range t.Columns {
			if c == col {
				found = true
				break
			}
		}
		if !found {
			return TableValue{}, fmt.Errorf("no such column: %s", col)
		}
	}

	rows := make([]RecordValue, len(t.Rows))
	for i, row := range t.Rows {
		fields := make(map[string]Value, len(cols))
		for _, col := range cols {
			if val, ok := row.Fields[col]; ok {
				fields[col] = val
			}
		}
		rows[i] = RecordValue{Fields: fields}
	}
	return TableValue{Columns: cols, Rows: rows}, nil
}

// First returns the first n rows.
func (t TableValue) First(n int) TableValue {
	if n <= 0 {
		return TableValue{Columns: t.Columns}
	}
	if n >= len(t.Rows) {
		return t
	}
	rows := make([]RecordValue, n)
	copy(rows, t.Rows[:n])
	return TableValue{Columns: t.Columns, Rows: rows}
}

// Last returns the last n rows.
func (t TableValue) Last(n int) TableValue {
	if n <= 0 {
		return TableValue{Columns: t.Columns}
	}
	if n >= len(t.Rows) {
		return t
	}
	rows := make([]RecordValue, n)
	copy(rows, t.Rows[len(t.Rows)-n:])
	return TableValue{Columns: t.Columns, Rows: rows}
}

// Unique returns rows with unique values for the given set of fields.
// If no fields are given, uses all columns.
func (t TableValue) Unique(fields []string) TableValue {
	if len(fields) == 0 {
		fields = t.Columns
	}

	seen := make(map[string]bool)
	var rows []RecordValue

	for _, row := range t.Rows {
		var keyParts []string
		for _, f := range fields {
			if val, ok := row.Fields[f]; ok {
				keyParts = append(keyParts, val.String())
			}
		}
		key := strings.Join(keyParts, "\x00")
		if !seen[key] {
			seen[key] = true
			rows = append(rows, row)
		}
	}

	return TableValue{Columns: t.Columns, Rows: rows}
}

// GroupBy returns the table grouped by the given field(s),
// with each group's row count as a new column.
// Returns a table with group columns + count column.
func (t TableValue) GroupBy(fields []string) TableValue {
	if len(fields) == 0 {
		fields = t.Columns
	}
	if len(t.Rows) == 0 {
		cols := append([]string{}, fields...)
		cols = append(cols, "count")
		return TableValue{Columns: cols}
	}

	type group struct {
		fields map[string]Value
		count  int
	}
	groups := make(map[string]*group)
	order := make([]string, 0)

	for _, row := range t.Rows {
		var keyParts []string
		entry := make(map[string]Value)
		for _, f := range fields {
			if val, ok := row.Fields[f]; ok {
				keyParts = append(keyParts, val.String())
				entry[f] = val
			}
		}
		key := strings.Join(keyParts, "\x00")
		if g, ok := groups[key]; ok {
			g.count++
		} else {
			groups[key] = &group{fields: entry, count: 1}
			order = append(order, key)
		}
	}

	cols := append([]string{}, fields...)
	cols = append(cols, "count")
	rows := make([]RecordValue, len(order))
	for i, key := range order {
		g := groups[key]
		rowFields := make(map[string]Value, len(cols))
		for k, v := range g.fields {
			rowFields[k] = v
		}
		rowFields["count"] = IntValue{Value: int64(g.count)}
		rows[i] = RecordValue{Fields: rowFields}
	}

	return TableValue{Columns: cols, Rows: rows}
}

// Aggregate performs an aggregation on a field.
// Valid ops: sum, avg, min, max.
func (t TableValue) Aggregate(field, op string) (Value, error) {
	if len(t.Rows) == 0 {
		return nil, fmt.Errorf("aggregate on empty table")
	}

	// Collect numeric values
	var nums []float64
	for _, row := range t.Rows {
		val, ok := row.Fields[field]
		if !ok {
			continue
		}
		switch v := val.(type) {
		case IntValue:
			nums = append(nums, float64(v.Value))
		case FloatValue:
			nums = append(nums, v.Value)
		}
	}

	if len(nums) == 0 {
		return nil, fmt.Errorf("no numeric values in column %q", field)
	}

	switch op {
	case "count":
		return IntValue{Value: int64(len(nums))}, nil
	case "sum":
		var total float64
		for _, n := range nums {
			total += n
		}
		if allInts(nums) {
			return IntValue{Value: int64(total)}, nil
		}
		return FloatValue{Value: total}, nil
	case "avg":
		var total float64
		for _, n := range nums {
			total += n
		}
		return FloatValue{Value: total / float64(len(nums))}, nil
	case "min":
		min := nums[0]
		for _, n := range nums[1:] {
			if n < min {
				min = n
			}
		}
		if allInts(nums) {
			return IntValue{Value: int64(min)}, nil
		}
		return FloatValue{Value: min}, nil
	case "max":
		max := nums[0]
		for _, n := range nums[1:] {
			if n > max {
				max = n
			}
		}
		if allInts(nums) {
			return IntValue{Value: int64(max)}, nil
		}
		return FloatValue{Value: max}, nil
	default:
		return nil, fmt.Errorf("unknown aggregate op %q", op)
	}
}

func allInts(nums []float64) bool {
	for _, n := range nums {
		if n != float64(int64(n)) {
			return false
		}
	}
	return true
}

// ParseCondition parses a condition string like "cpu > 20" or "name == foo".
func ParseCondition(s string) (Condition, error) {
	re := regexp.MustCompile(`^(\w+)\s*(==|!=|<=|>=|<|>|~=|in)\s*(.+)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return Condition{}, fmt.Errorf("invalid condition: %q", s)
	}
	field := matches[1]
	op := matches[2]
	raw := strings.TrimSpace(matches[3])
	return Condition{Field: field, Op: op, Value: ParseValue(raw)}, nil
}
