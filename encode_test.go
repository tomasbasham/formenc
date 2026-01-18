package formenc_test

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/tomasbasham/formenc"
)

var (
	baseTime    = time.Date(2025, 2, 8, 0, 0, 0, 0, time.UTC)
	optionalVal = "optional_value"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   interface{}
		want    []byte
		wantErr bool
	}{
		"nil value": {
			input: nil,
			want:  []byte{},
		},
		"nil pointer": {
			input: (*Person)(nil),
			want:  []byte{},
		},
		"zero values in struct": {
			input: &Person{},
			want:  pathEscape("name="),
		},
		"struct with all values": {
			input: &Person{
				Name:     "john",
				Age:      30,
				Pronouns: []string{"he", "him"},
			},
			want: pathEscape("age=30&name=john&pronouns[]=he&pronouns[]=him"),
		},
		"struct with omitempty and zero values": {
			input: &ComplexPerson{},
			want:  pathEscape("created_at=0001.01.01&id=0&name="),
		},
		"struct with omitempty and non-zero values": {
			input: &ComplexPerson{
				Name:     "jane",
				Age:      25,
				Pronouns: []string{"she", "her"},
			},
			want: pathEscape("age=25&created_at=0001.01.01&id=0&name=jane&pronouns[]=she&pronouns[]=her"),
		},
		"struct with custom type": {
			input: ComplexPerson{
				ID:        1,
				Name:      "jane",
				Pronouns:  []string{"she", "her"},
				Age:       25,
				CreatedAt: MyDate(baseTime),
				Private:   "hidden",
				Optional:  &optionalVal,
			},
			want: pathEscape("age=25&created_at=2025.02.08&id=1&name=jane&optional=optional_value&pronouns[]=she&pronouns[]=her"),
		},
		"empty slice": {
			input: &Person{
				Name:     "john",
				Pronouns: []string{},
			},
			want: pathEscape("name=john"),
		},
		"nil slice": {
			input: &Person{
				Name:     "john",
				Pronouns: nil,
			},
			want: pathEscape("name=john"),
		},
		"slice with empty strings": {
			input: &Person{
				Name:     "john",
				Pronouns: []string{"", ""},
			},
			want: pathEscape("name=john&pronouns[]=&pronouns[]="),
		},
		"deeply nested empty structs": {
			input: &User{},
			want:  pathEscape("address[city]=&address[state]=&address[street]=&address[zip]=&name="),
		},
		"deeply nested structs": {
			input: User{
				Name: "john",
				Age:  20,
				Address: Address{
					Street: "123 Main St",
					City:   "Anytown",
					State:  "CA",
					Zip:    "12345",
				},
			},
			want: pathEscape("address[city]=Anytown&address[state]=CA&address[street]=123+Main+St&address[zip]=12345&age=20&name=john"),
		},
		"ignroed fields": {
			input: IgnoredFieldsForm{
				Public:  "visible",
				Private: "hidden",
				Ignored: "skip",
				NoTag:   "value",
				Empty:   "value",
				Omitted: "",
				Complex: MyDate(baseTime),
			},
			want: pathEscape("Empty=value&NoTag=value&complex=2025.02.08&public=visible"),
		},
		"map with nil interface values": {
			input: map[string]interface{}{
				"key1": "value",
				"key2": nil,
			},
			want: pathEscape("key1=value"),
		},
		"map with empty string keys": {
			input: map[string]string{
				"":    "empty-key",
				"key": "value",
			},
			want: pathEscape("=empty-key&key=value"),
		},
		"map with special characters in values": {
			input: map[string]string{
				"url":   "https://example.com/path?query=value",
				"email": "user@example.com",
			},
			want: []byte("email=user%40example.com&url=https%3A%2F%2Fexample.com%2Fpath%3Fquery%3Dvalue"),
		},
		"nested maps": {
			input: map[string]interface{}{
				"outer": map[string]string{
					"inner": "value",
				},
			},
			want: pathEscape("outer[inner]=value"),
		},
		"map with slice values": {
			input: map[string]interface{}{
				"items": []string{"a", "b", "c"},
			},
			want: pathEscape("items[]=a&items[]=b&items[]=c"),
		},
		"map with mixed value types": {
			input: map[string]interface{}{
				"string": "text",
				"int":    42,
				"float":  3.14,
				"bool":   true,
			},
			want: pathEscape("bool=true&float=3.14&int=42&string=text"),
		},
		"unicode in struct fields": {
			input: &Person{
				Name: "太郎",
				Age:  25,
			},
			want: pathEscape("age=25&name=太郎"),
		},
		"large numbers": {
			input: map[string]int64{
				"max": 9223372036854775807,
				"min": -9223372036854775808,
			},
			want: pathEscape("max=9223372036854775807&min=-9223372036854775808"),
		},
		"float precision": {
			input: map[string]float64{
				"pi": 3.141592653589793,
				"e":  2.718281828459045,
			},
			want: pathEscape("e=2.718281828459045&pi=3.141592653589793"),
		},
		"boolean values": {
			input: map[string]bool{
				"yes": true,
				"no":  false,
			},
			want: pathEscape("no=false&yes=true"),
		},
		"pointer to primitive": {
			input: map[string]*int{
				"value": intPointer(42),
			},
			want: pathEscape("value=42"),
		},
		"nil pointer in map": {
			input: map[string]*int{
				"value": nil,
			},
			want: []byte(""),
		},
		"deeply nested structure": {
			input: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "deep",
					},
				},
			},
			want: pathEscape("level1[level2][level3]=deep"),
		},
		"all scalar types in map": {
			input: map[string]interface{}{
				"int":     int(1),
				"int8":    int8(2),
				"int16":   int16(3),
				"int32":   int32(4),
				"int64":   int64(5),
				"uint":    uint(6),
				"uint8":   uint8(7),
				"uint16":  uint16(8),
				"uint32":  uint32(9),
				"uint64":  uint64(10),
				"float32": float32(11.1),
				"float64": float64(12.2),
				"bool":    true,
				"string":  "text",
			},
			want: pathEscape("bool=true&float32=11.1&float64=12.2&int=1&int16=3&int32=4&int64=5&int8=2&string=text&uint=6&uint16=8&uint32=9&uint64=10&uint8=7"),
		},
		"nested array in map": {
			input: map[string]interface{}{
				"matrix": [][]int{
					{1, 2, 3},
					{4, 5, 6},
				},
			},
			want: pathEscape("matrix[][]=1&matrix[][]=2&matrix[][]=3&matrix[][]=4&matrix[][]=5&matrix[][]=6"),
		},
		"empty map": {
			input: map[string]interface{}{},
			want:  []byte(""),
		},
		"nil map": {
			input: map[string]interface{}(nil),
			want:  []byte(""),
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := formenc.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(got, tt.want, MyDateComparer); diff != "" {
					t.Errorf("mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func TestEncodeToString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   interface{}
		want    string
		wantErr bool
	}{
		"basic form": {
			input: &Person{
				Name: "john",
				Age:  20,
			},
			want: "age=20&name=john",
		},
		"empty struct": {
			input: &Person{},
			want:  "name=",
		},
		"nil pointer": {
			input: (*Person)(nil),
			want:  "",
		},
		"simple map": {
			input: map[string]string{"key": "value"},
			want:  "key=value",
		},
		"invalid input - string": {
			input:   "string",
			wantErr: true,
		},
		"invalid input - int": {
			input:   42,
			wantErr: true,
		},
		"nested structure": {
			input: map[string]interface{}{
				"user": map[string]string{
					"name": "john",
					"role": "admin",
				},
			},
			want: pathEscapeString("user[name]=john&user[role]=admin"),
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := formenc.EncodeToString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(got, tt.want); diff != "" {
					t.Errorf("mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func TestMarshal_UnsupportedTypes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   interface{}
		wantErr bool
	}{
		"channel": {
			input:   make(chan int),
			wantErr: true,
		},
		"function": {
			input:   func() {},
			wantErr: true,
		},
		"complex64": {
			input:   complex64(1 + 2i),
			wantErr: true,
		},
		"complex128": {
			input:   complex128(1 + 2i),
			wantErr: true,
		},
		"map with non-string keys": {
			input:   map[int]string{1: "value"},
			wantErr: true,
		},
		"string scalar": {
			input:   "hello",
			wantErr: true,
		},
		"int scalar": {
			input:   42,
			wantErr: true,
		},
		"float scalar": {
			input:   3.14,
			wantErr: true,
		},
		"bool scalar": {
			input:   true,
			wantErr: true,
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := formenc.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestMarshal_CustomMarshaler(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   interface{}
		want    []byte
		wantErr bool
	}{
		"custom date type in struct": {
			input: &ComplexPerson{
				CreatedAt: MyDate(baseTime),
			},
			want: pathEscape("created_at=2025.02.08&id=0&name="),
		},
		"custom date in map": {
			input: map[string]interface{}{
				"date": MyDate(baseTime),
			},
			want: pathEscape("date=2025.02.08"),
		},
		"custom date in slice": {
			input: map[string]interface{}{
				"dates": []MyDate{MyDate(baseTime)},
			},
			want: pathEscape("dates[]=2025.02.08"),
		},
		"nested custom types": {
			input: map[string]interface{}{
				"event": map[string]interface{}{
					"scheduled": MyDate(baseTime),
				},
			},
			want: pathEscape("event[scheduled]=2025.02.08"),
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := formenc.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(got, tt.want, MyDateComparer); diff != "" {
					t.Errorf("mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func TestMarshal_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   interface{}
		target  interface{}
		wantErr bool
	}{
		"basic form": {
			input: &Person{
				Name:     "john",
				Age:      30,
				Pronouns: []string{"he", "him"},
			},
			target: &Person{},
		},
		"complex form": {
			input: &ComplexPerson{
				ID:        1,
				Name:      "jane",
				Age:       25,
				Pronouns:  []string{"she", "her"},
				CreatedAt: MyDate(baseTime),
				Optional:  &optionalVal,
			},
			target: &ComplexPerson{},
		},
		"nested form": {
			input: &User{
				Name: "john",
				Age:  30,
				Address: Address{
					Street: "123 Main St",
					City:   "Anytown",
					State:  "CA",
					Zip:    "12345",
				},
			},
			target: &User{},
		},
		"simple map": {
			input: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			target: new(map[string]string),
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			encoded, err := formenc.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}

			err = formenc.Unmarshal(encoded, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.target, ref(tt.input), MyDateComparer); diff != "" {
					t.Errorf("mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	benchmarks := map[string]struct {
		input interface{}
	}{
		"basic form": {
			input: &Person{
				Name:     "john",
				Age:      20,
				Pronouns: []string{"he", "him"},
			},
		},
		"complex form with custom type": {
			input: &ComplexPerson{
				ID:        1,
				Name:      "jane",
				Age:       25,
				Pronouns:  []string{"she", "her"},
				CreatedAt: MyDate(baseTime),
				Optional:  &optionalVal,
			},
		},
		"nested form": {
			input: &User{
				Name: "john",
				Age:  30,
				Address: Address{
					Street: "123 Main St",
					City:   "Anytown",
					State:  "CA",
					Zip:    "12345",
				},
			},
		},
		"small map": {
			input: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
		},
		"medium map": {
			input: generateMap(50),
		},
		"large map": {
			input: generateMap(500),
		},
		"map with slices": {
			input: map[string]interface{}{
				"tags":  []string{"go", "golang", "programming", "web"},
				"ids":   []int{1, 2, 3, 4, 5},
				"flags": []bool{true, false, true},
			},
		},
		"deeply nested map": {
			input: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"level4": "deep",
							"data":   []string{"a", "b", "c"},
						},
					},
				},
			},
		},
		"mixed types map": {
			input: map[string]interface{}{
				"string": "text",
				"int":    42,
				"float":  3.14159,
				"bool":   true,
				"slice":  []int{1, 2, 3},
				"nested": map[string]string{"key": "value"},
			},
		},
	}
	for name, bm := range benchmarks {
		bm := bm
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if _, err := formenc.Marshal(bm.input); err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func intPointer(i int) *int {
	return &i
}

func pathEscape(s string) []byte {
	return []byte(url.PathEscape(s))
}

func pathEscapeString(s string) string {
	return url.PathEscape(s)
}

func ref(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		ptr := reflect.New(reflect.TypeOf(v))
		ptr.Elem().Set(rv)
		return ptr.Interface()
	}
	return v
}

func generateMap(size int) map[string]interface{} {
	m := make(map[string]interface{}, size)
	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key_%d", i)
		m[key] = fmt.Sprintf("value_%d", i)
	}
	return m
}
