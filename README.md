# go-jsonrepair

A Go implementation of [@josdejong/jsonrepair](https://github.com/josdejong/jsonrepair) - repair invalid JSON documents.

## Features

This library can fix many types of malformed JSON:

- ✅ **Add missing quotes** around keys and values
- ✅ **Convert single quotes** to double quotes
- ✅ **Add missing commas** between array/object elements
- ✅ **Remove trailing commas**
- ✅ **Strip JavaScript comments** (both `//` and `/* */`)
- ✅ **Convert Python constants** (`True`/`False`/`None` to `true`/`false`/`null`)
- ✅ **Repair truncated JSON** by adding missing closing brackets
- ✅ **Handle special quote characters** (like Unicode quotes)
- ✅ **Concatenate broken strings** (strings split with `+`)
- ✅ **Remove MongoDB types** (`NumberLong`, `ISODate`, `ObjectId`, etc.)
- ✅ **Strip JSONP wrappers** (like `callback({...})`)
- ✅ **Remove code fences** (like ` ```json ... ``` `)
- ✅ **Handle ellipsis** in arrays (like `[1, 2, ...]`)

## Installation

```bash
go get github.com/yinxulai/go-jsonrepair
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/yinxulai/go-jsonrepair/jsonrepair"
)

func main() {
    broken := "{name: 'John'}"
    repaired, err := jsonrepair.Repair(broken)
    if err != nil {
        panic(err)
    }
    fmt.Println(repaired) // Output: {"name":"John"}
}
```

## Examples

### Basic Repairs

```go
// Unquoted keys
jsonrepair.Repair(`{name: "John"}`)
// → {"name":"John"}

// Single quotes
jsonrepair.Repair(`{'name': 'John'}`)
// → {"name":"John"}

// Trailing commas
jsonrepair.Repair(`{"items": [1, 2, 3,]}`)
// → {"items":[1,2,3]}
```

### Comments

```go
// Single-line comments
jsonrepair.Repair(`{"a": 1, // comment
"b": 2}`)
// → {"a":1,"b":2}

// Multi-line comments
jsonrepair.Repair(`{"a": 1, /* comment */ "b": 2}`)
// → {"a":1,"b":2}
```

### Python Constants

```go
jsonrepair.Repair(`{"active": True, "deleted": False, "data": None}`)
// → {"active":true,"deleted":false,"data":null}
```

### MongoDB Types

```go
jsonrepair.Repair(`{"id": ObjectId("507f..."), "count": NumberLong("123")}`)
// → {"id":"507f...","count":"123"}
```

### Truncated JSON

```go
jsonrepair.Repair(`{"name": "John", "age": 30`)
// → {"name":"John","age":30}

jsonrepair.Repair(`{"name": "John", "data":`)
// → {"name":"John","data":null}
```

### String Concatenation

```go
jsonrepair.Repair(`{"message": "Hello " + "World"}`)
// → {"message":"Hello World"}
```

### JSONP and Code Fences

```go
jsonrepair.Repair(`callback({"success": true})`)
// → {"success":true}

jsonrepair.Repair("```json\n{\"data\": \"value\"}\n```")
// → {"data":"value"}
```

## Running Examples

See the `examples` directory for more examples:

```bash
cd examples
go run examples.go
```

Or run the main demo:

```bash
go run main.go
```

## Testing

Run the test suite:

```bash
make test
```

Build the project:

```bash
make build
```

## Coverage

The library has comprehensive test coverage (77.5%) covering all major repair scenarios.

## License

MIT