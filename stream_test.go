package formenc_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tomasbasham/formenc"
)

func TestDecoder_BasicForm(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		want    interface{}
		wantErr bool
	}{
		"valid query string": {
			input: "name=john&age=20&aliases[]=johnny&aliases[]=jonny",
			want: BasicForm{
				Name:    "john",
				Age:     20,
				Aliases: []string{"johnny", "jonny"},
			},
		},
		"invalid query string": {
			input:   "%%%",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got BasicForm
			decoder := formenc.NewDecoder(strings.NewReader(tt.input))
			err := decoder.Decode(&got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("(-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestEncoder(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   interface{}
		want    []byte
		wantErr bool
	}{
		"basic form": {
			input: &BasicForm{
				Name:    "john",
				Age:     20,
				Aliases: []string{"johnny", "jonny"},
			},
			want: pathEscape("age=20&aliases[]=johnny&aliases[]=jonny&name=john"),
		},
		"invalid target": {
			input:   map[int]interface{}{},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var b bytes.Buffer
			encoder := formenc.NewEncoder(&b)
			err := encoder.Encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, b.Bytes()); diff != "" {
					t.Errorf("(-want +got):\n%s", diff)
				}
			}
		})
	}
}
