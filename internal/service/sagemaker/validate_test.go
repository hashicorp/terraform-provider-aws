package sagemaker

import (
	"strings"
	"testing"
)

func TestValidName(t *testing.T) {
	validNames := []string{
		"ValidSageMakerName",
		"Valid-5a63Mak3r-Name",
		"123-456-789",
		"1234",
		strings.Repeat("W", 63),
	}
	for _, v := range validNames {
		_, errors := validName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid SageMaker name with maximum length 63 chars: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Invalid name",          // blanks are not allowed
		"1#{}nook",              // other non-alphanumeric chars
		"-nook",                 // cannot start with hyphen
		strings.Repeat("W", 64), // length > 63
	}
	for _, v := range invalidNames {
		_, errors := validName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid SageMaker name", v)
		}
	}
}
