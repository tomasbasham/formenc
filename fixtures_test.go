package formenc_test

import (
	"time"

	"github.com/google/go-cmp/cmp"
)

// Comparer for MyDate type.
var MyDateComparer = cmp.Comparer(func(x, y MyDate) bool {
	return time.Time(x).Equal(time.Time(y))
})

type Person struct {
	Name     string   `form:"name"`
	Age      int      `form:"age,omitempty"`
	Pronouns []string `form:"pronouns"`
}

type ComplexPerson struct {
	ID        int      `form:"id"`
	Name      string   `form:"name"`
	Age       int      `form:"age,omitempty"`
	Pronouns  []string `form:"pronouns,omitempty"`
	CreatedAt MyDate   `form:"created_at"`
	Private   string   `form:"-"`
	Optional  *string  `form:"optional,omitempty"`
}

type IgnoredFieldsForm struct {
	Public  string `form:"public"`
	Private string `form:"-"`
	Ignored string `form:",ignore"`
	NoTag   string
	Empty   string `form:""`
	Omitted string `form:",omitempty"`
	Complex MyDate `form:"complex,omitempty"`
}

type User struct {
	Name    string  `form:"name"`
	Age     int     `form:"age,omitempty"`
	Address Address `form:"address"`
}

type Address struct {
	Street string `form:"street"`
	City   string `form:"city"`
	State  string `form:"state"`
	Zip    string `form:"zip"`
}

type MyDate time.Time

func (d MyDate) MarshalForm() (string, error) {
	return time.Time(d).Format("2006.01.02"), nil
}

func (d *MyDate) UnmarshalForm(b string) error {
	t, err := time.Parse("2006.01.02", b)
	if err != nil {
		return err
	}
	*d = MyDate(t)
	return nil
}
