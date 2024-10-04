// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main_test

import (
	"fmt"
	"testing"
)

const (
	stateFarm = "Jake"
)

// Example is a test function that is analyzed when testing the main function
func Example() {
	fmt.Printf("%s\n", stateFarm)

	a := "Hello, World!"
	fmt.Printf("%s\n", a)

	b := `Hello, World!`
	fmt.Printf("%s\n", b)

	c := `
	just a
	newline`
	fmt.Printf("%s\n", c)

	d := `
# This is a comment
# This is another comment
# This is a third comment
`
	fmt.Printf("%s\n", d)

	e := "\n"
	fmt.Printf("%s\n", e)
}

func TestMain(t *testing.T) {
	Example()
}
