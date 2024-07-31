// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidDeskPhoneNumber(t *testing.T) {
	t.Parallel()

	validNumbers := []string{
		"+12345678912",
		"+6598765432",
	}
	for _, v := range validNumbers {
		_, errors := validDeskPhoneNumber(v, "desk_phone_number")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid desk phone number: %q", v, errors)
		}
	}

	invalidNumbers := []string{
		"1232",
		"invalid",
	}
	for _, v := range invalidNumbers {
		_, errors := validDeskPhoneNumber(v, "desk_phone_number")
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid desk phone number: %q", v, errors)
		}
	}
}

func TestValidPhoneNumberPrefix(t *testing.T) {
	t.Parallel()

	validPrefixes := []string{
		"+12345",
		"+659876",
		"+6598765432",
		"+61",
		"+1",
	}
	for _, v := range validPrefixes {
		_, errors := validPhoneNumberPrefix(v, names.AttrPrefix)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid phone number prefix: %q", v, errors)
		}
	}

	invalidPrefixes := []string{
		"1232",
		"98765432",
		"invalid",
	}
	for _, v := range invalidPrefixes {
		_, errors := validPhoneNumberPrefix(v, names.AttrPrefix)
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid phone number prefix: %q", v, errors)
		}
	}
}
