// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"testing"
)

func TestValidSecurityGroupRuleDescription(t *testing.T) {
	t.Parallel()

	validDescriptions := []string{
		"testrule",
		"testRule",
		"testRule 123",
		`testRule 123 ._-:/()#,@[]+=&;{}!$*`,
	}
	for _, v := range validDescriptions {
		_, errors := validSecurityGroupRuleDescription(v, "description")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid security group rule description: %q", v, errors)
		}
	}

	invalidDescriptions := []string{
		"`",
		"%%",
		`\`,
	}
	for _, v := range invalidDescriptions {
		_, errors := validSecurityGroupRuleDescription(v, "description")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid security group rule description", v)
		}
	}
}
