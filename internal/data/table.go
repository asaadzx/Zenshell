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
