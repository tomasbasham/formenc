# formenc

A Go module providing encoding and decoding of form data (e.g.,
`application/x-www-form-urlencoded` and `multipart/form-data`) into Go types,
with support for nested structures, slices, and maps. Built on reflection,
`formenc` provides type-safe encoding and decoding whilst preserving the
structure of complex data types.

## Prerequisites

You will need the following things properly installed on your computer:

- [Go](https://golang.org/): any one of the **three latest major**
  [releases](https://golang.org/doc/devel/release.html)

## Installation

With [Go module](https://go.dev/wiki/Modules) support (Go 1.11+), simply add the
following import

```go
import "github.com/tomasbasham/formenc"
```

to your code, and then `go [build|run|test]` will automatically fetch the
necessary dependencies.

Otherwise, to install the `formenc` module, run the following command:

```bash
go get -u github.com/tomasbasham/formenc
```

## Usage

To use this module, import it into your Go code and call `Marshal` to encode
structs or maps into form data, or `Unmarshal` to decode form data back into Go
values.

### Encoding

To encode a Go struct or map into form data, use the `formenc.Marshal` function.

```go
type Person struct {
    Name  string   `form:"name"`
    Age   int      `form:"age"`
    Tags  []string `form:"tags"`
}

p := Person{
    Name: "Alice",
    Age:  30,
    Tags: []string{"admin", "user"},
}

data, err := formenc.Marshal(p)
// data: "age=30&name=Alice&tags[]=admin&tags[]=user"
```

You can also encode maps:

```go
m := map[string]any{
    "user": map[string]any{
        "name": "Bob",
        "age":  25,
    },
}

data, err := formenc.Marshal(m)
// data: "user[age]=25&user[name]=Bob"
```

### Decoding

To decode form data into a Go struct or map, use the `formenc.Unmarshal`
function.

```go
data := "name=Charlie&age=35&tags[]=developer&tags[]=reviewer"

var p Person
err := formenc.Unmarshal([]byte(data), &p)
// p.Name: "Charlie"
// p.Age: 35
// p.Tags: []string{"developer", "reviewer"}
```

For dynamic structures, decode into map[string]any:

```go
data := "user[name]=Diana&user[permissions][]=read&user[permissions][]=write"

var m map[string]any
err := formenc.Unmarshal([]byte(data), &m)
// m: map[string]any{
//     "user": map[string]any{
//         "name": "Diana",
//         "permissions": []any{"read", "write"},
//     },
// }
```

### Streaming

Use `Encoder` and `Decoder` for working with `io.Reader` and `io.Writer`:

```go
// Encoding
var buf bytes.Buffer
enc := formenc.NewEncoder(&buf)
err := enc.Encode(person)

// Decoding
dec := formenc.NewDecoder(request.Body)
var person Person
err := dec.Decode(&person)
```

### Struct Tags

Control field behaviour using struct tags:

```go
type Config struct {
    APIKey    string `form:"api_key"`          // Custom field name
    Debug     bool   `form:"debug,omitempty"`  // Omit if zero value
    Internal  string `form:"-"`                // Always ignore
}
```

### Custom Marshalling

Implement `Marshaler` or `Unmarshaler` for custom encoding logic:

```go
type Timestamp time.Time

func (t Timestamp) MarshalForm() (string, error) {
    return time.Time(t).Format(time.RFC3339), nil
}

func (t *Timestamp) UnmarshalForm(s string) error {
    parsed, err := time.Parse(time.RFC3339, s)
    if err != nil {
        return err
    }
    *t = Timestamp(parsed)
    return nil
}
```

### Type Guarantees

| Target type       | Guarantee           |
| ----------------- | ------------------- |
| Structs           | Typed, validated    |
| Typed maps/slices | Typed, validated    |
| `map[string]any`  | Perfect round-trip  |
| `[]any`           | Perfect round-trip  |
| Mixed nesting     | Structure preserved |

## License

This project is licensed under the [MIT License](LICENSE).
