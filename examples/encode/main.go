package main

import (
	"fmt"
	"strings"

	"github.com/tomasbasham/formenc"
)

// User represents the expected structure of the form data.
type User struct {
	Name     string   `form:"name"`
	Age      int      `form:"age"`
	Address  Address  `form:"address"`
	Pronouns []string `form:"pronouns,omitempty"`
}

// Address represents a nested structure within the form data.
type Address struct {
	Street string `form:"street"`
	City   string `form:"city"`
	Zip    string `form:"zip,omitempty"`
}

// Sample input data simulating a form submission.
var input = []string{
	"name=John Smith",
	"age=20",
	"address[street]=123+Main+St",
	"address[city]=Anytown",
	"address[zip]=12345",
	"pronouns[]=he",
	"pronouns[]=him",
}

func main() {
	data := strings.Join(input, "&")

	var user User
	if err := formenc.DecodeString(data, &user); err != nil {
		panic(err)
	}
	fmt.Printf("User: %#v\n", user)

	s, err := formenc.EncodeToString(user)
	if err != nil {
		panic(err)
	}
	fmt.Println("Encoded:", s)

	var userMap map[string]interface{}
	if err = formenc.DecodeString(data, &userMap); err != nil {
		panic(err)
	}
	fmt.Printf("User Map: %#v\n", userMap)

	s, err = formenc.EncodeToString(userMap)
	if err != nil {
		panic(err)
	}
	fmt.Println("Encoded Map:", s)
}
