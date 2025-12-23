package main

import (
	"fmt"
	"log"

	"github.com/yinxulai/go-jsonrepair/jsonrepair"
)

func main() {
	examples := []struct {
		name  string
		input string
	}{
		{
			name:  "Basic - Unquoted Keys",
			input: `{name: "John", age: 30}`,
		},
		{
			name:  "Basic - Single Quotes",
			input: `{'name': 'John', 'age': 30}`,
		},
		{
			name:  "Basic - Trailing Commas",
			input: `{"items": [1, 2, 3,], "total": 3,}`,
		},
		{
			name:  "Comments - Single Line",
			input: "{\n  \"name\": \"John\", // This is a comment\n  \"age\": 30\n}",
		},
		{
			name:  "Comments - Multi Line",
			input: `{"name": "John", /* Multi-line comment */ "age": 30}`,
		},
		{
			name:  "Python - Constants",
			input: `{"active": True, "deleted": False, "value": None}`,
		},
		{
			name:  "MongoDB - Types",
			input: `{"id": ObjectId("507f1f77bcf86cd799439011"), "count": NumberLong("123")}`,
		},
		{
			name:  "Truncated - Missing Bracket",
			input: `{"name": "John", "age": 30`,
		},
		{
			name:  "Truncated - Incomplete Value",
			input: `{"name": "John", "data":`,
		},
		{
			name:  "String Concatenation",
			input: `{"message": "Hello " + "World"}`,
		},
		{
			name:  "JSONP - Wrapper",
			input: `callback({"success": true})`,
		},
		{
			name:  "Code Fence",
			input: "```json\n{\"data\": \"value\"}\n```",
		},
		{
			name:  "Array - Ellipsis",
			input: `[1, 2, 3, ...]`,
		},
		{
			name:  "Complex - Nested",
			input: `{name: 'John', address: {city: 'NYC', zip: '10001'}, tags: [1, 2, 3,]}`,
		},
	}

	fmt.Println("JSON Repair Examples")
	fmt.Println("====================\n")

	for _, example := range examples {
		fmt.Printf("Example: %s\n", example.name)
		fmt.Printf("Input:   %s\n", example.input)

		repaired, err := jsonrepair.Repair(example.input)
		if err != nil {
			log.Printf("Error: %v\n\n", err)
			continue
		}

		fmt.Printf("Output:  %s\n\n", repaired)
	}
}
