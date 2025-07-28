// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package names

import (
	"strings"
)

// toCamelCase converts a string to CamelCase.
func toCamelCase(in string, initialCap bool) string {
	out := strings.Builder{}

	nextIsCap := initialCap
	prevIsCap := false
	for i, ch := range []byte(in) {
		isCap := isCapitalLetter(ch)
		isLow := isLowercaseLetter(ch)
		isDig := isNumeric(ch)

		if nextIsCap {
			if isLow {
				ch = toUppercaseLetter(ch)
			}
		} else if i == 0 {
			if isCap {
				ch = toLowercaseLetter(ch)
			}
		} else if prevIsCap && isCap {
			ch = toLowercaseLetter(ch)
		}

		prevIsCap = isCap

		if isCap || isLow {
			out.WriteByte(ch)
			nextIsCap = false
		} else if isDig {
			out.WriteByte(ch)
			nextIsCap = true
		} else {
			nextIsCap = ch == '_' || ch == ' ' || ch == '-' || ch == '.'
		}
	}

	return out.String()
}

func toUppercaseLetter(ch byte) byte {
	ch += 'A'
	ch -= 'a'
	return ch
}

// ToCamelCase converts a string to CamelCase.
func ToCamelCase(in string) string {
	return toCamelCase(in, true)
}

// ToLowerCamelCase converts a string to camelCase.
func ToLowerCamelCase(in string) string {
	return toCamelCase(in, false)
}
