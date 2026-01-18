package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/tomasbasham/formenc"
)

// FormRequest represents the expected structure of the form data.
type FormRequest struct {
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
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader(strings.Join(input, "&")))

	// Set the content type to application/x-www-form-urlencoded.
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Handle the request.
	handleRequest(w, r)
	fmt.Println(w.Body.String())
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var req FormRequest

	dec := formenc.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Name: %s\n", req.Name)
	fmt.Fprintf(w, "Age: %d\n", req.Age)
	fmt.Fprintf(w, "Address: %#v\n", req.Address)

	if req.Pronouns != nil {
		fmt.Fprintf(w, "Pronouns: %#v", req.Pronouns)
	}
}
