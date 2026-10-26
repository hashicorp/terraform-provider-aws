// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"time"
)

func main() {
	sleepDuration := time.Second * 5
	for {
		fmt.Println("hello") // nosemgrep:ci.calling-fmt.Print-and-variants
		time.Sleep(sleepDuration)
	}
}
