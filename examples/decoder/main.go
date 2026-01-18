package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tomasbasham/formenc"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		panic("Usage: go run main.go [input]")
	}

	var decoded map[string]interface{}
	if err := formenc.Unmarshal([]byte(strings.Join(args, " ")), &decoded); err != nil {
		panic(err)
	}
	fmt.Println(decoded)
}
