package main

import (
	"fmt"
	"github.com/yinxulai/go-jsonrepair/jsonrepair"
)

func main() {
	// Example 1: Unquoted keys
	broken1 := `{name: "John", age: 30}`
	repaired1, err := jsonrepair.Repair(broken1)
	if err != nil {
		fmt.Printf("Error repairing example 1: %v\n", err)
		return
	}
	fmt.Printf("Input:    %s\n", broken1)
	fmt.Printf("Repaired: %s\n\n", repaired1)

	// Example 2: Single quotes
	broken2 := `{'name': 'Jane', 'city': 'NYC'}`
	repaired2, err := jsonrepair.Repair(broken2)
	if err != nil {
		fmt.Printf("Error repairing example 2: %v\n", err)
		return
	}
	fmt.Printf("Input:    %s\n", broken2)
	fmt.Printf("Repaired: %s\n\n", repaired2)

	// Example 3: Trailing commas
	broken3 := `{"items": [1, 2, 3,], "count": 3,}`
	repaired3, err := jsonrepair.Repair(broken3)
	if err != nil {
		fmt.Printf("Error repairing example 3: %v\n", err)
		return
	}
	fmt.Printf("Input:    %s\n", broken3)
	fmt.Printf("Repaired: %s\n\n", repaired3)

	// Example 4: Comments
	broken4 := `{"a": 1, /* comment */ "b": 2}`
	repaired4, err := jsonrepair.Repair(broken4)
	if err != nil {
		fmt.Printf("Error repairing example 4: %v\n", err)
		return
	}
	fmt.Printf("Input:    %s\n", broken4)
	fmt.Printf("Repaired: %s\n\n", repaired4)

	// Example 5: Python constants
	broken5 := `{"active": True, "deleted": False, "data": None}`
	repaired5, err := jsonrepair.Repair(broken5)
	if err != nil {
		fmt.Printf("Error repairing example 5: %v\n", err)
		return
	}
	fmt.Printf("Input:    %s\n", broken5)
	fmt.Printf("Repaired: %s\n\n", repaired5)

	// Example 6: Truncated JSON
	broken6 := `{"name": "John", "age": 30`
	repaired6, err := jsonrepair.Repair(broken6)
	if err != nil {
		fmt.Printf("Error repairing example 6: %v\n", err)
		return
	}
	fmt.Printf("Input:    %s\n", broken6)
	fmt.Printf("Repaired: %s\n\n", repaired6)

	fmt.Println("All examples repaired successfully!")
}
