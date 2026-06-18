package data

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ToJSON serializes a TableValue to a JSON array of objects.
func (t TableValue) ToJSON() (string, error) {
	if len(t.Rows) == 0 {
		return "[]", nil
	}

	rows := make([]map[string]interface{}, len(t.Rows))
	for i, row := range t.Rows {
		m := make(map[string]interface{}, len(t.Columns))
		for _, col := range t.Columns {
			if val, ok := row.Fields[col]; ok {
				m[col] = valueToJSON(val)
			}
		}
		rows[i] = m
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return "", fmt.Errorf("to-json: %w", err)
	}
	return string(data), nil
}

func valueToJSON(v Value) interface{} {
	switch val := v.(type) {
	case StringValue:
		return val.Value
	case IntValue:
		return val.Value
	case FloatValue:
		return val.Value
	case BoolValue:
		return val.Value
	case ListValue:
		items := make([]interface{}, len(val.Values))
		for i, item := range val.Values {
			items[i] = valueToJSON(item)
		}
		return items
	default:
		return v.String()
	}
}

// FromJSON parses a JSON array of objects into a TableValue.
func FromJSON(s string) (TableValue, error) {
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(s), &rows); err != nil {
		return TableValue{}, fmt.Errorf("from-json: %w", err)
	}

	if len(rows) == 0 {
		return TableValue{}, nil
	}

	// Collect all columns preserving order from first row
	colSet := make(map[string]bool)
	var columns []string
	for _, row := range rows {
		for k := range row {
			if !colSet[k] {
				colSet[k] = true
				columns = append(columns, k)
			}
		}
	}

	result := make([]RecordValue, len(rows))
	for i, row := range rows {
		fields := make(map[string]Value, len(columns))
		for _, col := range columns {
			if raw, ok := row[col]; ok {
				fields[col] = jsonToValue(raw)
			}
		}
		result[i] = RecordValue{Fields: fields}
	}

	return TableValue{Columns: columns, Rows: result}, nil
}

func jsonToValue(raw interface{}) Value {
	switch v := raw.(type) {
	case string:
		return StringValue{Value: v}
	case float64:
		// Check if it's actually an integer
		if v == float64(int64(v)) {
			return IntValue{Value: int64(v)}
		}
		return FloatValue{Value: v}
	case bool:
		return BoolValue{Value: v}
	case nil:
		return StringValue{Value: ""}
	case []interface{}:
		items := make([]Value, len(v))
		for i, item := range v {
			items[i] = jsonToValue(item)
		}
		return ListValue{Values: items}
	case map[string]interface{}:
		fields := make(map[string]Value, len(v))
		for k, val := range v {
			fields[k] = jsonToValue(val)
		}
		return RecordValue{Fields: fields}
	default:
		return StringValue{Value: fmt.Sprintf("%v", v)}
	}
}

// --- CSV ---

// ToCSV serializes a TableValue to CSV text.
func (t TableValue) ToCSV() (string, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)

	if len(t.Columns) > 0 {
		if err := w.Write(t.Columns); err != nil {
			return "", fmt.Errorf("to-csv: %w", err)
		}
	}

	for _, row := range t.Rows {
		record := make([]string, len(t.Columns))
		for i, col := range t.Columns {
			if val, ok := row.Fields[col]; ok {
				record[i] = val.String()
			}
		}
		if err := w.Write(record); err != nil {
			return "", fmt.Errorf("to-csv: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("to-csv: %w", err)
	}
	return b.String(), nil
}

// FromCSV parses CSV text into a TableValue.
func FromCSV(s string) (TableValue, error) {
	r := csv.NewReader(strings.NewReader(s))
	records, err := r.ReadAll()
	if err != nil {
		return TableValue{}, fmt.Errorf("from-csv: %w", err)
	}

	if len(records) == 0 {
		return TableValue{}, nil
	}

	columns := records[0]
	rows := make([]RecordValue, len(records)-1)
	for i := 1; i < len(records); i++ {
		fields := make(map[string]Value, len(columns))
		for j, col := range columns {
			if j < len(records[i]) {
				fields[col] = ParseValue(records[i][j])
			}
		}
		rows[i-1] = RecordValue{Fields: fields}
	}

	return TableValue{Columns: columns, Rows: rows}, nil
}

// parseNumberOrString attempts to parse a string as int or float, falling back to string.
func parseNumberOrString(s string) Value {
	if s == "" {
		return StringValue{Value: s}
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return IntValue{Value: n}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return FloatValue{Value: f}
	}
	return StringValue{Value: s}
}
