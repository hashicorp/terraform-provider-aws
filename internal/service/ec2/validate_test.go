// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"testing"

	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
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
		_, errors := validSecurityGroupRuleDescription(v, names.AttrDescription)
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
		_, errors := validSecurityGroupRuleDescription(v, names.AttrDescription)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid security group rule description", v)
		}
	}
}

func TestValidLaunchTemplateUserData(t *testing.T) {
	t.Parallel()

	b64 := func(n int) string {
		return inttypes.Base64Encode(bytes.Repeat([]byte("u"), n))
	}
	validUserData := map[string]string{
		"empty":        "",
		"one byte":     b64(1),
		"at the limit": b64(maxUserDataLength),
	}
	for name, v := range validUserData {
		_, errors := validLaunchTemplateUserData(v, "user_data")
		if len(errors) != 0 {
			t.Fatalf("%s: should be valid launch template user data: %q", name, errors)
		}
	}

	invalidUserData := map[string]string{
		"not base64":    "not base64!",
		"one byte over": b64(maxUserDataLength + 1),
	}
	for name, v := range invalidUserData {
		_, errors := validLaunchTemplateUserData(v, "user_data")
		if len(errors) == 0 {
			t.Fatalf("%s: should be invalid launch template user data", name)
		}
	}
}
