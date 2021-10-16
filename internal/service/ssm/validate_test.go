package ssm

import (
	"strings"
	"testing"
)

func TestValidName(t *testing.T) {
	validNames := []string{
		".foo-bar_123",
		strings.Repeat("W", 128),
	}
	for _, v := range validNames {
		_, errors := validName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid SSM Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"foo+bar",
		"tf",
		strings.Repeat("W", 129), // > 128
	}
	for _, v := range invalidNames {
		_, errors := validName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid SSM Name: %q", v, errors)
		}
	}
}
