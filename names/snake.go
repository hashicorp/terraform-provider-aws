// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package names

import (
	"strings"
)

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(in string) string {
	out := strings.Builder{}

	for i, ch := range []byte(in) {
		isCap := isCapitalLetter(ch)
		isLow := isLowercaseLetter(ch)
		isDig := isNumeric(ch)

		if isCap {
			ch = toLowercaseLetter(ch)
		}

		if i < len(in)-1 {
			nextCh := in[i+1]
			nextIsCap := isCapitalLetter(nextCh)
			nextIsLow := isLowercaseLetter(nextCh)
			nextIsDig := isNumeric(nextCh)

			// Append underscore if case changes.
			if (isCap && nextIsLow) || (isLow && (nextIsCap || nextIsDig) || (isDig && (nextIsCap || nextIsLow))) {
				if isCap && nextIsLow {
					if prevIsCap := i > 0 && isCapitalLetter(in[i-1]); prevIsCap {
						out.WriteByte('_')
					}
				}
				out.WriteByte(ch)
				if isLow || isDig {
					out.WriteByte('_')
				}

				continue
			}
		}

		if isCap || isLow || isDig {
			out.WriteByte(ch)
		} else {
			out.WriteByte('_')
		}
	}

	return out.String()
}

func isCapitalLetter(ch byte) bool {
	return ch >= 'A' && ch <= 'Z'
}

func isLowercaseLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z'
}

func isNumeric(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func toLowercaseLetter(ch byte) byte {
	ch += 'a'
	ch -= 'A'
	return ch
}
