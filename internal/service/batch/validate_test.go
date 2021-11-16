package batch

import (
	"strings"
	"testing"
)

func TestValidName(t *testing.T) {
	validNames := []string{
		strings.Repeat("W", 128), // <= 128
	}
	for _, v := range validNames {
		_, errors := validName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Batch name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"s@mple",
		strings.Repeat("W", 129), // >= 129
	}
	for _, v := range invalidNames {
		_, errors := validName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid Batch name: %q", v, errors)
		}
	}
}

func TestValidPrefix(t *testing.T) {
	validPrefixes := []string{
		strings.Repeat("W", 102), // <= 102
	}
	for _, v := range validPrefixes {
		_, errors := validPrefix(v, "prefix")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Batch prefix: %q", v, errors)
		}
	}

	invalidPrefixes := []string{
		"s@mple",
		strings.Repeat("W", 103), // >= 103
	}
	for _, v := range invalidPrefixes {
		_, errors := validPrefix(v, "prefix")
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid Batch prefix: %q", v, errors)
		}
	}
}
