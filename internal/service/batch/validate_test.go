// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		strings.Repeat("W", 128), // <= 128
	}
	for _, v := range validNames {
		_, errors := validName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Batch name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"s@mple",
		strings.Repeat("W", 129), // >= 129
	}
	for _, v := range invalidNames {
		_, errors := validName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid Batch name: %q", v, errors)
		}
	}
}

func TestValidPrefix(t *testing.T) {
	t.Parallel()

	validPrefixes := []string{
		strings.Repeat("W", 102), // <= 102
	}
	for _, v := range validPrefixes {
		_, errors := validPrefix(v, names.AttrPrefix)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Batch prefix: %q", v, errors)
		}
	}

	invalidPrefixes := []string{
		"s@mple",
		strings.Repeat("W", 103), // >= 103
	}
	for _, v := range invalidPrefixes {
		_, errors := validPrefix(v, names.AttrPrefix)
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid Batch prefix: %q", v, errors)
		}
	}
}

func TestValidShareIdentifier(t *testing.T) {
	t.Parallel()

	validShareIdentifiers := []string{
		"sample*",
		"sample1",
		strings.Repeat("W", 255),       // <= 255,
		strings.Repeat("W", 254) + "*", // optional asterisk
	}
	for _, v := range validShareIdentifiers {
		_, errors := validShareIdentifier(v, "share_identifier")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid share identifier: %q", v, errors)
		}
	}

	invalidShareIdentifiers := []string{
		"s@mple",
		strings.Repeat("W", 256), // > 255
	}
	for _, v := range invalidShareIdentifiers {
		_, errors := validShareIdentifier(v, "share_identifier")
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid share identifier: %q", v, errors)
		}
	}
}
