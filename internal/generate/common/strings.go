// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// FirstLower returns a string with the first character as lower case.
func FirstLower(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

// FirstUpper returns a string with the first character as upper case.
func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// Title returns a string with the first character of each word as upper case.
func Title(s string) string {
	return cases.Title(language.Und, cases.NoLower).String(s)
}
