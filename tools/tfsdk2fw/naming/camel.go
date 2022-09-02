package naming

import (
	"strings"
)

// ToCamelCase converts a string to CamelCase.
func ToCamelCase(s string) string {
	c := strings.Builder{}

	capitalizeNext := true
	for _, ch := range []byte(strings.TrimSpace(s)) {
		isCap := isCapitalLetter(ch)
		isLow := isLowercaseLetter(ch)
		isDig := isNumeric(ch)

		if capitalizeNext && isLow {
			ch = toCapitalLetter(ch)
		}

		if isCap || isLow {
			c.WriteByte(ch)
			capitalizeNext = false
		} else if isDig {
			c.WriteByte(ch)
			capitalizeNext = true
		} else {
			capitalizeNext = ch == '_' || ch == ' ' || ch == '-' || ch == '.'
		}
	}

	return c.String()
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

func toCapitalLetter(ch byte) byte {
	ch += 'A'
	ch -= 'a'
	return ch
}
