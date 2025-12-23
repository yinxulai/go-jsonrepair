# go-jsonrepair

A Go implementation of [@josdejong/jsonrepair](https://github.com/josdejong/jsonrepair) - repair invalid JSON documents.

## Features

- Add missing quotes around keys
- Convert single quotes to double quotes
- Add missing commas between array/object elements
- Remove trailing commas
- Strip JavaScript comments (// and /* */)
- Convert Python constants (True/False/None) to JSON equivalents
- Repair truncated JSON
- Handle special quote characters
- Concatenate broken strings
- Remove MongoDB types and JSONP wrappers
- And more...

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
    fmt.Println(repaired) // Output: {"name": "John"}
}
```

## License

MIT