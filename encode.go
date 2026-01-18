package formenc

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Marshaler is the interface implemented by types that can marshal themselves
// into a form description.
type Marshaler interface {
	MarshalForm() (string, error)
}

// EncodeToString is a convenience function that returns the form encoding of v
// as a string.
func EncodeToString(v interface{}) (string, error) {
	b, err := Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Marshal returns the form encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte{}, nil
	}

	// Dereference pointer if needed.
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return []byte{}, nil
		}
		rv = rv.Elem()
	}

	// Ensure the top-level value is a struct or map.
	if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Map {
		return nil, fmt.Errorf("form: top-level value must be struct or map")
	}

	// Ensure map keys are strings.
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("form: map keys must be strings")
	}

	values := url.Values{}
	if err := marshalValue(values, nil, rv); err != nil {
		return nil, err
	}

	return []byte(values.Encode()), nil
}

func marshalValue(out url.Values, path []string, v reflect.Value) error {
	// Handle nill pointers early to avoid dereferencing them.
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return nil
	}

	// Only deref if we can actually modify the value (it's addressable) or if
	// it's not nil
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	// Handle custom Marshaler first.
	if m, ok := asMarshaler(v); ok {
		return marshaler(out, path, m)
	}

	// Dispatch based on the kind of the value.
	switch v.Kind() {
	case reflect.Struct:
		return marshalStruct(out, path, v)
	case reflect.Map:
		return marshalMap(out, path, v)
	case reflect.Slice, reflect.Array:
		return marshalSlice(out, path, v)
	case reflect.Interface:
		if !v.IsNil() {
			return marshalValue(out, path, v.Elem())
		}
		return nil
	default:
		return marshalScalar(out, path, v)
	}
}

func marshaler(out url.Values, path []string, m Marshaler) error {
	s, err := m.MarshalForm()
	if err != nil {
		return err
	}
	out.Add(renderPath(path), s)
	return nil
}

func marshalStruct(out url.Values, path []string, v reflect.Value) error {
	tags := tags(v)
	for i := 0; i < v.NumField(); i++ {
		tag := tags[i]
		if tag.Ignore {
			continue
		}
		fv := v.Field(i)
		if tag.Omit && isEmptyValue(fv) {
			continue
		}
		if tag.Name == "" {
			continue
		}
		if err := marshalValue(out, append(path, tag.Name), fv); err != nil {
			return err
		}
	}
	return nil
}

func marshalMap(out url.Values, path []string, v reflect.Value) error {
	for _, k := range v.MapKeys() {
		mv := v.MapIndex(k)
		if !mv.IsValid() || (mv.Kind() == reflect.Interface && mv.IsNil()) {
			continue
		}
		if err := marshalValue(out, append(path, k.String()), mv); err != nil {
			return err
		}
	}
	return nil
}

func marshalSlice(out url.Values, path []string, v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if !elem.IsValid() || (elem.Kind() == reflect.Interface && elem.IsNil()) {
			continue
		}
		if err := marshalValue(out, append(path, ""), elem); err != nil {
			return err
		}
	}
	return nil
}

func marshalScalar(out url.Values, path []string, v reflect.Value) error {
	out.Add(renderPath(path), getScalar(v))
	return nil
}

func asMarshaler(v reflect.Value) (Marshaler, bool) {
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(Marshaler); ok {
			return m, true
		}
	}
	if m, ok := v.Interface().(Marshaler); ok {
		return m, true
	}
	return nil, false
}

func renderPath(path []string) string {
	var b strings.Builder
	b.WriteString(path[0])
	for _, p := range path[1:] {
		if p == "" {
			b.WriteString("[]")
		} else {
			b.WriteString("[")
			b.WriteString(p)
			b.WriteString("]")
		}
	}
	return b.String()
}

func getScalar(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits())
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		panic("form: unsupported type: " + v.Type().String())
	}
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Interface, reflect.Pointer:
		return v.IsZero()
	}
	return false
}
