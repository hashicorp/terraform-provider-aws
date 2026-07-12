// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("hello")
		time.Sleep(5 * time.Second)
	}
}
