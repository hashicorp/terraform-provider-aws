package hil_test

import (
	"fmt"
	"log"

	"github.com/hashicorp/hil"
)

func Example_basic() {
	input := "${6 + 2}"

	tree, err := hil.Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	result, err := hil.Eval(tree, &hil.EvalConfig{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Type: %s\n", result.Type)
	fmt.Printf("Value: %s\n", result.Value)
	// Output:
	// Type: TypeString
	// Value: 8
}
