package formenc_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/tomasbasham/formenc"
)

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   []byte
		target  interface{}
		want    interface{}
		wantErr bool
	}{
		"empty input": {
			input:   []byte(""),
			target:  &Person{},
			wantErr: true,
		},
		"whitespace only": {
			input:  []byte("   "),
			target: &Person{},
			want:   &Person{},
		},
		"malformed query string with empty value": {
			input:  []byte("name=john&age="),
			target: &Person{},
			want: &Person{
				Name: "john",
				Age:  0,
			},
		},
		"duplicate keys with different values": {
			input:  []byte("name=john&name=jane"),
			target: &Person{},
			want: &Person{
				Name: "jane",
			},
		},
		"special characters in values": {
			input:  []byte("name=john+doe&age=20"),
			target: &Person{},
			want: &Person{
				Name: "john doe",
				Age:  20,
			},
		},
		"url encoded special characters": {
			input:  []byte("name=john%40example.com"),
			target: &Person{},
			want: &Person{
				Name: "john@example.com",
			},
		},
		"nested struct with missing fields": {
			input:  []byte("name=john"),
			target: &User{},
			want: &User{
				Name:    "john",
				Age:     0,
				Address: Address{},
			},
		},
		"map with interface values - all strings": {
			input:  []byte("string=hello&number=42&bool=true"),
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"string": "hello",
				"number": "42",
				"bool":   "true",
			},
		},
		"deeply nested map structure": {
			input:  []byte("data[level1][level2][level3]=value"),
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"data": map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"level3": "value",
						},
					},
				},
			},
		},
		"array within map": {
			input:  []byte("data[items][]=a&data[items][]=b&data[items][]=c"),
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"data": map[string]interface{}{
					"items": []interface{}{"a", "b", "c"},
				},
			},
		},
		"mixed nested arrays and maps": {
			input:  []byte("matrix[0][]=1&matrix[0][]=2&matrix[1][]=3&matrix[1][]=4"),
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"matrix": map[string]interface{}{
					"0": []interface{}{"1", "2"},
					"1": []interface{}{"3", "4"},
				},
			},
		},
		"nested maps": {
			input:  []byte("users[0][name]=john&users[1][name]=jane&users[0][age]=20&users[1][age]=25"),
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"users": map[string]interface{}{
					"0": map[string]interface{}{"name": "john", "age": "20"},
					"1": map[string]interface{}{"name": "jane", "age": "25"},
				},
			},
		},
		"zero values": {
			input:  []byte("name="),
			target: &Person{},
			want: &Person{
				Name: "",
				Age:  0,
			},
		},
		"boolean edge cases": {
			input:  []byte("t=t&f=f&one=1&zero=0&yes=true&no=false"),
			target: new(map[string]bool),
			want: &map[string]bool{
				"t":    true,
				"f":    false,
				"one":  true,
				"zero": false,
				"yes":  true,
				"no":   false,
			},
		},
		"large integer values": {
			input:  []byte("max_int64=9223372036854775807&min_int64=-9223372036854775808"),
			target: new(map[string]int64),
			want: &map[string]int64{
				"max_int64": 9223372036854775807,
				"min_int64": -9223372036854775808,
			},
		},
		"float precision": {
			input:  []byte("pi=3.14159265359&e=2.71828182846"),
			target: new(map[string]float64),
			want: &map[string]float64{
				"pi": 3.14159265359,
				"e":  2.71828182846,
			},
		},
		"scientific notation": {
			input:  []byte("sci=1.23e10"),
			target: new(map[string]float64),
			want: &map[string]float64{
				"sci": 1.23e10,
			},
		},
		"negative zero": {
			input:  []byte("val=-0"),
			target: new(map[string]int),
			want: &map[string]int{
				"val": 0,
			},
		},
		"slice with single element": {
			input:  []byte("items[]=single"),
			target: new(map[string][]string),
			want: &map[string][]string{
				"items": {"single"},
			},
		},
		"empty slice notation": {
			input:  []byte("items[]="),
			target: new(map[string][]string),
			want: &map[string][]string{
				"items": {""},
			},
		},
		"repeated scalar converts to slice in map": {
			input:  []byte("tags=go&tags=golang&tags=programming"),
			target: new(map[string][]string),
			want: &map[string][]string{
				"tags": {"go", "golang", "programming"},
			},
		},
		"unicode in keys and values": {
			input:  []byte("名前=太郎&city=東京"),
			target: new(map[string]string),
			want: &map[string]string{
				"名前":   "太郎",
				"city": "東京",
			},
		},
		"percent-encoded unicode": {
			input:  []byte("name=%E5%A4%AA%E9%83%8E"),
			target: new(map[string]string),
			want: &map[string]string{
				"name": "太郎",
			},
		},
		"complex nested structure with arrays": {
			input:  []byte("config[database][connections][][host]=db1&config[database][connections][][host]=db2"),
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"config": map[string]interface{}{
					"database": map[string]interface{}{
						"connections": []interface{}{
							map[string]interface{}{"host": "db1"},
							map[string]interface{}{"host": "db2"},
						},
					},
				},
			},
		},
		"numeric indices in maps": {
			input:  []byte("matrix[0][0]=a&matrix[0][1]=b&matrix[1][0]=c&matrix[1][1]=d"),
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"matrix": map[string]interface{}{
					"0": map[string]interface{}{
						"0": "a",
						"1": "b",
					},
					"1": map[string]interface{}{
						"0": "c",
						"1": "d",
					},
				},
			},
		},
		"consecutive empty brackets": {
			input:  []byte("grid[][]=1&grid[][]=2"),
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"grid": []interface{}{
					[]interface{}{"1"},
					[]interface{}{"2"},
				},
			},
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := formenc.Unmarshal(tt.input, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.target, tt.want, MyDateComparer); diff != "" {
					t.Errorf("mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func TestDecodeString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		target  interface{}
		want    interface{}
		wantErr bool
	}{
		"valid string": {
			input:  "name=john&age=20",
			target: &Person{},
			want: &Person{
				Name: "john",
				Age:  20,
			},
		},
		"empty string": {
			input:   "",
			target:  &Person{},
			wantErr: true,
		},
		"nil target": {
			input:   "name=john",
			target:  nil,
			wantErr: true,
		},
		"non-pointer target": {
			input:   "name=john",
			target:  Person{},
			wantErr: true,
		},
		"complex nested structure": {
			input:  "user[profile][name]=jane&user[profile][age]=25",
			target: new(map[string]interface{}),
			want: &map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"name": "jane",
						"age":  "25",
					},
				},
			},
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := formenc.DecodeString(tt.input, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.target, tt.want, MyDateComparer); diff != "" {
					t.Errorf("mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func TestUnmarshal_InvalidTypes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   []byte
		target  interface{}
		wantErr bool
	}{
		"slice target": {
			input:   []byte("name=john"),
			target:  new([]string),
			wantErr: true,
		},
		"string target": {
			input:   []byte("name=john"),
			target:  new(string),
			wantErr: true,
		},
		"int target": {
			input:   []byte("value=42"),
			target:  new(int),
			wantErr: true,
		},
		"float target": {
			input:   []byte("value=3.14"),
			target:  new(float64),
			wantErr: true,
		},
		"bool target": {
			input:   []byte("value=true"),
			target:  new(bool),
			wantErr: true,
		},
		"channel target": {
			input:   []byte("value=data"),
			target:  new(chan int),
			wantErr: true,
		},
		"function target": {
			input:   []byte("value=data"),
			target:  new(func()),
			wantErr: true,
		},
		"map with int keys": {
			input:   []byte("1=value"),
			target:  new(map[int]string),
			wantErr: true,
		},
		"map with float keys": {
			input:   []byte("3.14=value"),
			target:  new(map[float64]string),
			wantErr: true,
		},
		"map with bool keys": {
			input:   []byte("true=value"),
			target:  new(map[bool]string),
			wantErr: true,
		},
		"struct with complex64 field": {
			input:   []byte("value=1+2i"),
			target:  &struct{ Value complex64 }{},
			wantErr: true,
		},
		"struct with complex128 field": {
			input:   []byte("value=1+2i"),
			target:  &struct{ Value complex128 }{},
			wantErr: true,
		},
		"struct with channel field": {
			input:   []byte("ch=data"),
			target:  &struct{ Ch chan int }{},
			wantErr: true,
		},
		"struct with function field": {
			input:   []byte("fn=data"),
			target:  &struct{ Fn func() }{},
			wantErr: true,
		},
		"map with complex64 values": {
			input:   []byte("value=1+2i"),
			target:  new(map[string]complex64),
			wantErr: true,
		},
		"map with complex128 values": {
			input:   []byte("value=1+2i"),
			target:  new(map[string]complex128),
			wantErr: true,
		},
		"map with channel values": {
			input:   []byte("ch=data"),
			target:  new(map[string]chan int),
			wantErr: true,
		},
		"map with function values": {
			input:   []byte("fn=data"),
			target:  new(map[string]func()),
			wantErr: true,
		},
		"nested struct with complex field": {
			input:   []byte("nested[value]=1+2i"),
			target:  &struct{ Nested struct{ Value complex64 } }{},
			wantErr: true,
		},
		"nested map with invalid value type": {
			input:   []byte("outer[inner]=value"),
			target:  new(map[string]map[string]complex128),
			wantErr: true,
		},
		"slice of complex numbers in struct": {
			input:   []byte("values[]=1+2i"),
			target:  &struct{ Values []complex64 }{},
			wantErr: true,
		},
		"slice of channels in map": {
			input:   []byte("channels[]=data"),
			target:  new(map[string][]chan int),
			wantErr: true,
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := formenc.Unmarshal(tt.input, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestInvalidUnmarshalError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		target  interface{}
		wantErr bool
		wantMsg string
	}{
		"nil target": {
			target:  nil,
			wantMsg: "form: Unmarshal(nil)",
			wantErr: true,
		},
		"non-pointer target": {
			target:  Person{},
			wantMsg: "form: Unmarshal(non-pointer formenc_test.Person)",
			wantErr: true,
		},
		"nil pointer target": {
			target:  (*Person)(nil),
			wantMsg: "form: Unmarshal(nil *formenc_test.Person)",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := formenc.Unmarshal([]byte("name=test"), tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				got := err.Error()
				if got != tt.wantMsg {
					t.Errorf("mismatch:\n  got:  %q\n  want: %q", got, tt.wantMsg)
				}
			}
		})
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	benchmarks := map[string]struct {
		input  []byte
		target func() interface{}
	}{
		"basic form": {
			input:  []byte("name=john&age=20&pronouns[]=he&pronouns[]=him"),
			target: func() interface{} { return &Person{} },
		},
		"complex form": {
			input:  []byte("id=1&name=jane&age=25&pronouns[]=she&pronouns[]=her&created_at=2025.02.08&optional=optional_value"),
			target: func() interface{} { return &ComplexPerson{} },
		},
		"nested form": {
			input:  []byte("name=john&age=30&address[street]=123+Main+St&address[city]=Anytown&address[state]=CA&address[zip]=12345"),
			target: func() interface{} { return &User{} },
		},
		"small map": {
			input:  []byte("a=1&b=2&c=3"),
			target: func() interface{} { return new(map[string]string) },
		},
		"medium map": {
			input:  generateEncodedMap(50),
			target: func() interface{} { return new(map[string]string) },
		},
		"large map": {
			input:  generateEncodedMap(500),
			target: func() interface{} { return new(map[string]string) },
		},
		"map with typed slices": {
			input:  []byte("tags[]=go&tags[]=golang&tags[]=programming&ids[]=1&ids[]=2&ids[]=3"),
			target: func() interface{} { return new(map[string][]string) },
		},
		"map with interface slices": {
			input:  []byte("items[]=a&items[]=b&items[]=c&items[]=d&items[]=e"),
			target: func() interface{} { return new(map[string]interface{}) },
		},
		"deeply nested map": {
			input:  []byte("level1[level2][level3][level4]=deep&level1[level2][level3][data][]=a&level1[level2][level3][data][]=b"),
			target: func() interface{} { return new(map[string]interface{}) },
		},
		"mixed types map": {
			input:  []byte("string=text&int=42&float=3.14159&bool=true"),
			target: func() interface{} { return new(map[string]interface{}) },
		},
		"url encoded content": {
			input:  []byte("email=user%40example.com&url=https%3A%2F%2Fexample.com%2Fpath&name=john+doe"),
			target: func() interface{} { return new(map[string]string) },
		},
		"unicode content": {
			input:  []byte("name=%E5%A4%AA%E9%83%8E&city=%E6%9D%B1%E4%BA%AC"),
			target: func() interface{} { return new(map[string]string) },
		},
	}
	for name, bm := range benchmarks {
		bm := bm
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				target := bm.target()
				if err := formenc.Unmarshal(bm.input, target); err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func generateEncodedMap(size int) []byte {
	var parts []string
	for i := 0; i < size; i++ {
		parts = append(parts, fmt.Sprintf("key_%d=value_%d", i, i))
	}
	return []byte(strings.Join(parts, "&"))
}
