// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/YakDriver/regexache"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run sigint_helper.go <delay_seconds>") // nosemgrep:ci.calling-fmt.Print-and-variants
		os.Exit(1)
	}

	delay, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Invalid delay: %v\n", err) // nosemgrep:ci.calling-fmt.Print-and-variants
		os.Exit(1)
	}

	time.Sleep(time.Duration(delay) * time.Second)

	// Find terraform process doing apply
	cmd := exec.Command("ps", "aux") //lintignore:XR007
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running ps: %v\n", err) // nosemgrep:ci.calling-fmt.Print-and-variants
		os.Exit(1)
	}

	lines := strings.Split(string(output), "\n")
	re := regexache.MustCompile(`/opt/homebrew/bin/terraform apply.*-auto-approve`)

	for _, line := range lines {
		if re.MatchString(line) && !strings.Contains(line, "sigint_helper") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				pid, err := strconv.Atoi(fields[1])
				if err != nil {
					continue
				}

				fmt.Printf("Sending SIGINT to PID %d: %s\n", pid, line) // nosemgrep:ci.calling-fmt.Print-and-variants
				_ = syscall.Kill(pid, syscall.SIGINT)
				return
			}
		}
	}

	fmt.Println("No matching terraform process found") // nosemgrep:ci.calling-fmt.Print-and-variants
}
