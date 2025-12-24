package jsonrepair

import (
	"encoding/json"
	"testing"
)

func TestRepairBasicCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already valid JSON",
			input:    `{"name": "John"}`,
			expected: `{"name": "John"}`,
		},
		{
			name:     "unquoted keys",
			input:    `{name: "John"}`,
			expected: `{"name": "John"}`,
		},
		{
			name:     "single quotes",
			input:    `{'name': 'John'}`,
			expected: `{"name": "John"}`,
		},
		{
			name:     "mixed quotes",
			input:    `{name: 'John'}`,
			expected: `{"name": "John"}`,
		},
		{
			name:     "trailing comma in object",
			input:    `{"a": 1,}`,
			expected: `{"a": 1}`,
		},
		{
			name:     "trailing comma in array",
			input:    `[1, 2, 3,]`,
			expected: `[1, 2, 3]`,
		},
		{
			name:     "single line comment",
			input:    "{\"a\": 1, // comment\n\"b\": 2}",
			expected: `{"a": 1, "b": 2}`,
		},
		{
			name:     "multi-line comment",
			input:    `{"a": 1, /* comment */ "b": 2}`,
			expected: `{"a": 1, "b": 2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			// Validate that result is valid JSON
			var v interface{}
			if err := json.Unmarshal([]byte(result), &v); err != nil {
				t.Errorf("Result is not valid JSON: %v\nResult: %s", err, result)
			}

			// Compare normalized JSON
			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairPythonConstants(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Python True",
			input:    `{"value": True}`,
			expected: `{"value": true}`,
		},
		{
			name:     "Python False",
			input:    `{"value": False}`,
			expected: `{"value": false}`,
		},
		{
			name:     "Python None",
			input:    `{"value": None}`,
			expected: `{"value": null}`,
		},
		{
			name:     "Mixed Python constants",
			input:    `{"a": True, "b": False, "c": None}`,
			expected: `{"a": true, "b": false, "c": null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairTruncatedJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "truncated object",
			input:    `{"a": 1`,
			expected: `{"a": 1}`,
		},
		{
			name:     "truncated array",
			input:    `[1, 2, 3`,
			expected: `[1, 2, 3]`,
		},
		{
			name:     "truncated string",
			input:    `{"name": "John`,
			expected: `{"name": "John"}`,
		},
		{
			name:     "truncated nested",
			input:    `{"a": {"b": 1`,
			expected: `{"a": {"b": 1}}`,
		},
		{
			name:     "truncated after colon",
			input:    `{"a": 1, "b":`,
			expected: `{"a": 1, "b": null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairStringConcatenation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "concatenate strings with +",
			input:    `"hello" + "world"`,
			expected: `"helloworld"`,
		},
		{
			name:     "concatenate in object",
			input:    `{"text": "hello" + "world"}`,
			expected: `{"text": "helloworld"}`,
		},
		{
			name:     "concatenate single quotes",
			input:    `'hello' + 'world'`,
			expected: `"helloworld"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairMongoDBTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "NumberLong",
			input:    `{"value": NumberLong("123")}`,
			expected: `{"value": "123"}`,
		},
		{
			name:     "NumberInt",
			input:    `{"value": NumberInt("456")}`,
			expected: `{"value": "456"}`,
		},
		{
			name:     "ISODate",
			input:    `{"date": ISODate("2021-01-01")}`,
			expected: `{"date": "2021-01-01"}`,
		},
		{
			name:     "ObjectId",
			input:    `{"_id": ObjectId("507f1f77bcf86cd799439011")}`,
			expected: `{"_id": "507f1f77bcf86cd799439011"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairJSONP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSONP wrapper",
			input:    `callback({"a": 1})`,
			expected: `{"a": 1}`,
		},
		{
			name:     "JSONP with complex name",
			input:    `myCallback123({"name": "John"})`,
			expected: `{"name": "John"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairCodeFence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "code fence with json",
			input:    "```json\n{\"a\": 1}\n```",
			expected: `{"a": 1}`,
		},
		{
			name:     "code fence without language",
			input:    "```\n{\"name\": \"John\"}\n```",
			expected: `{"name": "John"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairEllipsis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ellipsis in array",
			input:    `[1, 2, 3, ...]`,
			expected: `[1, 2, 3]`,
		},
		{
			name:     "ellipsis at start",
			input:    `[..., 1, 2]`,
			expected: `[1, 2]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "integer",
			input:    `{"value": 123}`,
			expected: `{"value": 123}`,
		},
		{
			name:     "negative",
			input:    `{"value": -456}`,
			expected: `{"value": -456}`,
		},
		{
			name:     "decimal",
			input:    `{"value": 3.14}`,
			expected: `{"value": 3.14}`,
		},
		{
			name:     "exponent",
			input:    `{"value": 1.23e10}`,
			expected: `{"value": 1.23e10}`,
		},
		{
			name:     "exponent with plus",
			input:    `{"value": 1.23e+10}`,
			expected: `{"value": 1.23e+10}`,
		},
		{
			name:     "exponent with minus",
			input:    `{"value": 1.23e-10}`,
			expected: `{"value": 1.23e-10}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRepairComplexCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "complex nested object",
			input:    `{name: 'John', age: 30, address: {city: 'NYC', zip: '10001'}}`,
			expected: `{"name": "John", "age": 30, "address": {"city": "NYC", "zip": "10001"}}`,
		},
		{
			name:     "array of objects",
			input:    `[{name: 'John'}, {name: 'Jane'}]`,
			expected: `[{"name": "John"}, {"name": "Jane"}]`,
		},
		{
			name:     "mixed issues",
			input:    "{name: 'John', // comment\nage: 30,}",
			expected: `{"name": "John", "age": 30}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Repair(tt.input)
			if err != nil {
				t.Fatalf("Repair() error = %v", err)
			}

			if !jsonEqual(result, tt.expected) {
				t.Errorf("Repair() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Helper function to compare JSON strings by parsing and comparing
func jsonEqual(a, b string) bool {
	var va, vb interface{}
	if err := json.Unmarshal([]byte(a), &va); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &vb); err != nil {
		return false
	}

	// Marshal both back to get normalized JSON
	aa, err := json.Marshal(va)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(vb)
	if err != nil {
		return false
	}

	return string(aa) == string(bb)
}
