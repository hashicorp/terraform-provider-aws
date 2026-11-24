// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"regexp"

	"github.com/YakDriver/regexache"
)

// Package-level var declarations compile at init() time: not allowed.

// ruleid: regexp-no-compile-at-init
var pattern1 = regexp.MustCompile(`^[a-z]+$`)

// ruleid: regexp-no-compile-at-init
var pattern2 = regexache.MustCompile(`^[a-z]+$`)

// ruleid: regexp-no-compile-at-init
var pattern3 = regexp.Compile(`^[a-z]+$`)

// ruleid: regexp-no-compile-at-init
var pattern4 *regexp.Regexp = regexp.MustCompile(`^[a-z]+$`)

var (
	// ruleid: regexp-no-compile-at-init
	pattern5 = regexp.MustCompile(`^[a-z]+$`)
	// ruleid: regexp-no-compile-at-init
	pattern6 = regexache.MustCompilePOSIX(`^[a-z]+$`)
)

// Compiling inside func init() also happens at init() time: not allowed.
func init() {
	// ruleid: regexp-no-compile-at-init
	p := regexp.MustCompile(`^[a-z]+$`)
	_ = p
}

// Compiling inside a regular function is lazy: allowed.
func validate(s string) bool {
	// ok: regexp-no-compile-at-init
	return regexache.MustCompile(`^[a-z]+$`).MatchString(s)
}

// Assigning to a local var inside a regular function is allowed.
func validate2(s string) bool {
	// ok: regexp-no-compile-at-init
	p := regexp.MustCompile(`^[a-z]+$`)
	return p.MatchString(s)
}

// A full var declaration inside a regular function is also lazy: allowed.
func validate3(s string) bool {
	// ok: regexp-no-compile-at-init
	var p = regexp.MustCompile(`^[a-z]+$`)
	return p.MatchString(s)
}

// Non-regexp package-level vars are allowed.
// ok: regexp-no-compile-at-init
var notARegexp = compute()

func compute() string {
	return "x"
}
