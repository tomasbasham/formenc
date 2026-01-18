package formenc_test

import (
	"fmt"
	"net/url"
	"os"

	"github.com/tomasbasham/formenc"
)

type Animal int

const (
	Unknown Animal = iota
	Gopher
	Zebra
)

func (a Animal) MarshalForm() (string, error) {
	switch a {
	case Gopher:
		return "gopher", nil
	case Zebra:
		return "zebra", nil
	default:
		return "unknown", nil
	}
}

func (a *Animal) UnmarshalForm(value string) error {
	switch value {
	case "gopher":
		*a = Gopher
	case "zebra":
		*a = Zebra
	default:
		*a = Unknown
	}
	return nil
}

func Example_customMarshal() {
	type PetOwner struct {
		OwnerName string `form:"owner_name"`
		PetType   Animal `form:"pet_type"`
	}

	owner := PetOwner{
		OwnerName: "Alice",
		PetType:   Gopher,
	}

	data, err := formenc.Marshal(owner)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	encoded, _ := url.PathUnescape(string(data))
	fmt.Println(encoded)
	// Output:
	// owner_name=Alice&pet_type=gopher
}

func ExampleMarshal() {
	user := User{
		Name: "Jane Doe",
		Age:  28,
		Address: Address{
			Street: "456 Oak St",
			City:   "Othertown",
			State:  "CA",
			Zip:    "67890",
		},
	}

	data, err := formenc.Marshal(user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	encoded, _ := url.PathUnescape(string(data))
	fmt.Println(encoded)
	// Output:
	// address[city]=Othertown&address[state]=CA&address[street]=456+Oak+St&address[zip]=67890&age=28&name=Jane+Doe
}

func ExampleUnmarshal() {
	data := []byte("name=John+Doe&age=30&address[street]=123+Main+St&address[city]=Anytown&address[state]=NY&address[zip]=12345")

	var user User
	if err := formenc.Unmarshal(data, &user); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Printf("%#v\n", user)
	// Output:
	// formenc_test.User{Name:"John Doe", Age:30, Address:formenc_test.Address{Street:"123 Main St", City:"Anytown", State:"NY", Zip:"12345"}}
}
