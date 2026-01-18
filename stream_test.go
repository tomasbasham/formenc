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
			input: "name=john&age=20&pronouns[]=he&pronouns[]=him",
			want: Person{
				Name:     "john",
				Age:      20,
				Pronouns: []string{"he", "him"},
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

			var got Person
			decoder := formenc.NewDecoder(strings.NewReader(tt.input))
			err := decoder.Decode(&got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
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
			input: &Person{
				Name:     "john",
				Age:      20,
				Pronouns: []string{"he", "him"},
			},
			want: pathEscape("age=20&name=john&pronouns[]=he&pronouns[]=him"),
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
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, b.Bytes()); diff != "" {
					t.Errorf("(-want +got):\n%s", diff)
				}
			}
		})
	}
}
