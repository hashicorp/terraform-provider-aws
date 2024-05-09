// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"ValidSageMakerName",
		"Valid-5a63Mak3r-Name",
		"123-456-789",
		"1234",
		strings.Repeat("W", 63),
	}
	for _, v := range validNames {
		_, errors := validName(v, names.AttrName)
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
		_, errors := validName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid SageMaker name", v)
		}
	}
}

func TestValidPrefix(t *testing.T) {
	t.Parallel()

	maxLength := 37
	validPrefixes := []string{
		"ValidSageMakerName",
		"Valid-5a63Mak3r-Name",
		"123-456-789",
		"1234",
		strings.Repeat("W", maxLength),
	}
	for _, v := range validPrefixes {
		_, errors := validPrefix(v, names.AttrNamePrefix)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid SageMaker prefix with maximum length %d chars: %q", v, maxLength, errors)
		}
	}

	invalidPrefixes := []string{
		"Invalid prefix",                 // blanks are not allowed
		"1#{}nook",                       // other non-alphanumeric chars
		"-nook",                          // cannot start with hyphen
		strings.Repeat("W", maxLength+1), // length > maxLength
	}
	for _, v := range invalidPrefixes {
		_, errors := validPrefix(v, names.AttrNamePrefix)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid SageMaker prefix", v)
		}
	}
}
