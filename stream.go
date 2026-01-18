package formenc

import (
	"fmt"
	"io"
)

// Decoder reads form-urlencoded data from an [io.Reader] and decodes it into a
// Go value.
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a new [Decoder] that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the form-urlencoded data from the underlying [io.Reader] and
// decodes it into v.
func (d *Decoder) Decode(v interface{}) error {
	body, err := io.ReadAll(d.r)
	if err != nil {
		return fmt.Errorf("form: failed to read body: %w", err)
	}

	return Unmarshal(body, v)
}

// Encoder writes form-urlencoded data to an [io.Writer].
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a new [Encoder] that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode encodes v as form-urlencoded data and writes it to the underlying
// [io.Writer].
func (e *Encoder) Encode(v interface{}) error {
	data, err := Marshal(v)
	if err != nil {
		return err
	}

	_, err = e.w.Write(data)
	return err
}
