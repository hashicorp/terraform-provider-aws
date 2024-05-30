// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidRoleProfileName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"tf-test-role-profile-1",
	}

	for _, s := range validNames {
		_, errors := validRolePolicyName(s, names.AttrName)
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid IAM role policy name: %v", s, errors)
		}
	}

	invalidNames := []string{
		"invalid#name",
		"this-is-a-very-long-role-policy-name-this-is-a-very-long-role-policy-name-this-is-a-very-long-role-policy-name-this-is-a-very-long",
	}

	for _, s := range invalidNames {
		_, errors := validRolePolicyName(s, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid IAM role policy name: %v", s, errors)
		}
	}
}

func TestValidAccountAlias(t *testing.T) {
	t.Parallel()

	validAliases := []string{
		"tf-alias",
		"0tf-alias1",
	}

	for _, s := range validAliases {
		_, errors := validAccountAlias(s, "account_alias")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid account alias: %v", s, errors)
		}
	}

	invalidAliases := []string{
		"tf",
		"-tf",
		"tf-",
		"TF-Alias",
		"tf-alias-tf-alias-tf-alias-tf-alias-tf-alias-tf-alias-tf-alias-tf-alias",
	}

	for _, s := range invalidAliases {
		_, errors := validAccountAlias(s, "account_alias")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid account alias: %v", s, errors)
		}
	}
}

func TestValidOpenIDURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value: "https://good.test",
		},
		{
			Value:    "http://wrong.scheme.test",
			ErrCount: 1,
		},
		{
			Value:    "ftp://wrong.scheme.test",
			ErrCount: 1,
		},
		{
			Value:    "%@invalidUrl",
			ErrCount: 1,
		},
		{
			Value:    "https://no-queries.test/?query=param",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validOpenIDURL(tc.Value, names.AttrURL)

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d of OpenID URL validation errors, got %d", tc.ErrCount, len(errors))
		}
	}
}

func TestValidRolePolicyRoleName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value: "S3Access",
		},
		{
			Value: "role/S3Access",
		},
		{
			Value:    "arn:aws:iam::123456789012:role/S3Access", // lintignore:AWSAT005
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validRolePolicyRole(tc.Value, names.AttrRole)

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d Role Policy role name validation errors, got %d", tc.ErrCount, len(errors))
		}
	}
}
