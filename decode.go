package formenc

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// InvalidUnmarshalError describes an invalid argument passed to [Unmarshal].
// (The argument to [Unmarshal] must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "form: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "form: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "form: Unmarshal(nil " + e.Type.String() + ")"
}

// Unmarshaler is the interface implemented by types that can unmarshal a form
// description of themselves. The input can be assumed to be a valid encoding of
// a form value. [Unmarshaler.UnmarshalForm] must copy the form data if it
// wishes to retain the data after returning.
type Unmarshaler interface {
	UnmarshalForm(string) error
}

// DecodeString is a convenience function that parses the form data in the
// string and stores the result in the value pointed to by v. If v is nil or not
// a pointer, DecodeString returns an [InvalidValueError].
func DecodeString(data string, v interface{}) error {
	return Unmarshal([]byte(data), v)
}

// Unmarshal parses the form data and stores the result in the value pointed to
// by v. If v is nil or not a pointer, Unmarshal returns an [InvalidValueError].
func Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("form: empty input")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Map {
		return fmt.Errorf("form: top-level value must be struct or map")
	}

	// Ensure map keys are strings.
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("form: map keys must be strings")
	}

	// Make sure to trim spaces to avoid future parse errors. url.ParseQuery does
	// not do this automatically and can produce keys containing only spaces.
	values, err := url.ParseQuery(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("form: invalid form data: %w", err)
	}

	return unmarshalForm(values, rv)
}

func unmarshalForm(values url.Values, v reflect.Value) error {
	for rawKey, vals := range values {
		path, err := parseKey(rawKey)
		if err != nil {
			return err
		}
		for _, val := range vals {
			if err := assign(v, path, val); err != nil {
				return fmt.Errorf("form: %w", err)
			}
		}
	}
	return nil
}

func assign(v reflect.Value, path []pathSegment, val string) error {
	v = deref(v)

	// If the path is empty, we are at a leaf node.
	if len(path) == 0 {
		return assignLeaf(v, val)
	}

	// Get the next segment of the path.
	seg := path[0]

	// Dispatch based on the kind of the value.
	switch v.Kind() {
	case reflect.Struct:
		return assignStructField(v, seg.Key, path[1:], val)
	case reflect.Map:
		return assignMapValue(v, seg, path[1:], val)
	case reflect.Slice:
		return assignSliceValue(v, seg, path[1:], val)
	case reflect.Interface:
		return assignInterfaceValue(v, path, val)
	default:
		return fmt.Errorf("cannot assign to %v", v.Kind())
	}
}

// dereference a pointer value, allocating a new value if needed.
func deref(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v.Elem()
	}
	return v
}

// assign a leaf value (string) to v. If v implements [Unmarshaler], use that.
func assignLeaf(v reflect.Value, val string) error {
	if u, ok := asUnmarshaler(v); ok {
		return u.UnmarshalForm(val)
	}
	return setScalar(v, val)
}

// assign a struct field identified by key.
func assignStructField(v reflect.Value, key string, path []pathSegment, val string) error {
	field := findStructField(v, key)
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("unknown field %q in struct %v", key, v.Type())
	}
	return assign(field, path, val)
}

// assign a map value identified by a path segment.
func assignMapValue(v reflect.Value, seg pathSegment, path []pathSegment, val string) error {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	key := reflect.ValueOf(seg.Key)
	elem := v.MapIndex(key)
	elemType := v.Type().Elem()

	switch elemType.Kind() {
	case reflect.Interface:
		newVal, err := inferInterfaceValue(elem, path, val)
		if err != nil {
			return err
		}
		v.SetMapIndex(key, newVal)
		return nil

	// Typed slice: get existing slice or make a new one
	case reflect.Slice:
		var slice reflect.Value
		if elem.IsValid() {
			slice = elem
		} else {
			slice = reflect.MakeSlice(elemType, 0, 1)
		}

		// New element
		newElem := reflect.New(elemType.Elem()).Elem()
		if err := assignLeaf(newElem, val); err != nil {
			return err
		}

		slice = reflect.Append(slice, newElem)
		v.SetMapIndex(key, slice)
		return nil

	// Single value
	default:
		if !elem.IsValid() {
			elem = reflect.New(elemType).Elem()
		}
		if err := assign(deref(elem), path, val); err != nil {
			return err
		}
		v.SetMapIndex(key, elem)
		return nil
	}
}

// assign a slice value identified by a path segment.
func assignSliceValue(v reflect.Value, seg pathSegment, path []pathSegment, val string) error {
	if !seg.Index {
		return fmt.Errorf("form: expected slice index")
	}
	elemType := v.Type().Elem()

	var newElem reflect.Value
	if elemType.Kind() == reflect.Interface {
		var err error
		newElem, err = inferInterfaceValue(reflect.Value{}, path, val)
		if err != nil {
			return err
		}
	} else {
		newElem = reflect.New(elemType).Elem()
		if len(path) == 0 {
			// Leaf element
			if err := assignLeaf(newElem, val); err != nil {
				return err
			}
		} else {
			// Nested struct/map
			if err := assign(newElem, path, val); err != nil {
				return err
			}
		}
	}
	v.Set(reflect.Append(v, newElem))
	return nil
}

func assignInterfaceValue(v reflect.Value, path []pathSegment, val string) error {
	if !v.IsValid() || v.IsNil() {
		newVal, err := inferInterfaceValue(v, path, val)
		if err != nil {
			return err
		}
		v.Set(newVal)
		return nil
	}
	return assign(v.Elem(), path, val)
}

// infer the value for an interface type based on the path segments.
func inferInterfaceValue(v reflect.Value, path []pathSegment, val string) (reflect.Value, error) {
	// Leaf node. When no type information is available, default to string. This
	// is consistent with form value semantics, and guarantees round-trip safety.
	if len(path) == 0 {
		return reflect.ValueOf(val), nil
	}

	// However we do want to infer the structure of nested values, so we can build
	// maps and slices as needed. Get the next path segment.
	seg := path[0]

	// If the next segment has an index, it's a slice element.
	if seg.Index {
		return inferSliceValue(v, path, val)
	}

	// Otherwise it's a map element.
	return inferMapValue(v, seg, path, val)
}

// infer a slice value for the given path segment.
func inferSliceValue(v reflect.Value, path []pathSegment, val string) (reflect.Value, error) {
	var slice []interface{}
	if v.IsValid() && !v.IsNil() {
		slice = v.Interface().([]interface{})
	}

	elem, err := inferInterfaceValue(reflect.Value{}, path[1:], val)
	if err != nil {
		return reflect.Value{}, err
	}

	slice = append(slice, elem.Interface())
	return reflect.ValueOf(slice), nil
}

// infer a map value for the given path segment. Unlike slices, we need to
// explicitly instantiate the map if it doesn't exist, as it is not possible to
// insert into a nil map.
func inferMapValue(v reflect.Value, seg pathSegment, path []pathSegment, val string) (reflect.Value, error) {
	m := make(map[string]interface{})
	if v.IsValid() && !v.IsNil() {
		m = v.Interface().(map[string]interface{})
	}

	elem, err := inferInterfaceValue(reflect.ValueOf(m[seg.Key]), path[1:], val)
	if err != nil {
		return reflect.Value{}, err
	}

	m[seg.Key] = elem.Interface()
	return reflect.ValueOf(m), nil
}

func asUnmarshaler(v reflect.Value) (Unmarshaler, bool) {
	if v.CanAddr() {
		if u, ok := v.Addr().Interface().(Unmarshaler); ok {
			return u, true
		}
	}
	if u, ok := v.Interface().(Unmarshaler); ok {
		return u, true
	}
	return nil, false
}

func findStructField(v reflect.Value, key string) reflect.Value {
	tags := tags(v)
	for i := 0; i < v.NumField(); i++ {
		if tags[i].Ignore {
			continue
		}
		if tags[i].Name == key {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

func setScalar(v reflect.Value, val string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(v, val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setUint(v, val)
	case reflect.Float32, reflect.Float64:
		return setFloat(v, val)
	case reflect.Bool:
		return parseBool(v, val)
	default:
		return fmt.Errorf("unsupported type: %v", v.Type())
	}
	return nil
}

func setInt(v reflect.Value, s string) error {
	if s == "" {
		v.SetInt(0)
		return nil
	}
	i, err := strconv.ParseInt(s, 10, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("setInt: %w", err)
	}
	v.SetInt(i)
	return nil
}

func setUint(v reflect.Value, s string) error {
	if s == "" {
		v.SetUint(0)
		return nil
	}
	i, err := strconv.ParseUint(s, 10, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("parseUint: %w", err)
	}
	v.SetUint(i)
	return nil
}

func setFloat(v reflect.Value, s string) error {
	if s == "" {
		v.SetFloat(0)
		return nil
	}
	f, err := strconv.ParseFloat(s, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("parseFloat: %w", err)
	}
	v.SetFloat(f)
	return nil
}

func parseBool(v reflect.Value, s string) error {
	if s == "" {
		v.SetBool(false)
		return nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("parseBool: %w", err)
	}
	v.SetBool(b)
	return nil
}
