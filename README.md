# JsonMapper_v2

The JsonMapper_v2 is an enhanced version of the original JSON manipulation library hosted at [https://github.com/skkim-01/json-mapper](https://github.com/skkim-01/json-mapper). This iteration has been improved using insights from ChatGPT to provide a more robust and feature-rich interface for working with JSON data in Go.

## Features

- **Initialization**: Create a new `JsonMapper` instance from a JSON string, file, or byte slice, allowing for flexible data sources.
- **Find**: Retrieve values from the JSON structure using a dot-separated key path. Supports array indexing with both `.index` and `[index]` notations.
- **Add**: Insert or update values at a specified key path. The function intelligently handles missing intermediate maps or slices, creating them as needed. Supports appending to slices using `-1` index.
- **Remove**: Remove values at a specified key path, including elements from arrays, shifting subsequent elements as needed.
- **Type-specific Finders**: Retrieve values of specific types (e.g., bool, string, int) from the JSON structure, simplifying type assertions and error handling.
- **WriteFile**: Save the current JSON structure to a file, with an option to format the output with indentation for readability.
- **Logical and Comparison Conditions**: Perform advanced queries within the JSON structure using a combination of logical (AND, OR, XOR, NOR) and comparison (equal, not equal, greater than, etc.) operators.

## Usage

Here's a brief example of how to use jsonmapper_v2:

```go
package main

import (
    "fmt"
    "github.com/skkim-01/jsonmapper_v2"
)

func main() {
    jsonStr := `{"name": "John", "age": 30, "friends": [{"name": "Doe", "age": 25}, {"name": "Jane", "age": 28}]}`

    jm, err := jsonmapper_v2.NewJsonMapStr(jsonStr)
    if err != nil {
        fmt.Println("Error initializing JsonMapper:", err)
        return
    }

    // Find a value
    age, _ := jm.Find("age")
    fmt.Println("Age:", age)

    // Add a new friend
    jm.Add("friends[-1]", map[string]interface{}{"name": "Alice", "age": 24})

    // Remove an element
    jm.Remove("friends[0]")

    // Write changes to a file
    jm.WriteFile("updated.json", true)
}
```