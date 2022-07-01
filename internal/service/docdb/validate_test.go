package docdb

import (
	"strings"
	"testing"
)

func TestValidIdentifier(t *testing.T) {
	validNames := []string{
		"a",
		"hello-world",
		"hello-world-0123456789",
		strings.Repeat("w", 63),
	}
	for _, v := range validNames {
		_, errors := validIdentifier(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid DocDB Identifier: %q", v, errors)
		}
	}

	invalidNames := []string{
		"",
		"special@character",
		"slash/in-the-middle",
		"dot.in-the-middle",
		"two-hyphen--in-the-middle",
		"0-first-numeric",
		"-first-hyphen",
		"end-hyphen-",
		strings.Repeat("W", 64),
	}
	for _, v := range invalidNames {
		_, errors := validIdentifier(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid DocDB Identifier", v)
		}
	}
}
