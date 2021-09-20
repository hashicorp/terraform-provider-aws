package iam

import (
	"testing"
)

func TestValidRoleProfileName(t *testing.T) {
	validNames := []string{
		"tf-test-role-profile-1",
	}

	for _, s := range validNames {
		_, errors := validRolePolicyName(s, "name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid IAM role policy name: %v", s, errors)
		}
	}

	invalidNames := []string{
		"invalid#name",
		"this-is-a-very-long-role-policy-name-this-is-a-very-long-role-policy-name-this-is-a-very-long-role-policy-name-this-is-a-very-long",
	}

	for _, s := range invalidNames {
		_, errors := validRolePolicyName(s, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid IAM role policy name: %v", s, errors)
		}
	}
}

func TestValidRoleProfileNamePrefix(t *testing.T) {
	validNamePrefixes := []string{
		"tf-test-role-profile-",
	}

	for _, s := range validNamePrefixes {
		_, errors := validRolePolicyNamePrefix(s, "name_prefix")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid IAM role policy name prefix: %v", s, errors)
		}
	}

	invalidNamePrefixes := []string{
		"invalid#name_prefix",
		"this-is-a-very-long-role-policy-name-prefix-this-is-a-very-long-role-policy-name-prefix-this-is-a-very-",
	}

	for _, s := range invalidNamePrefixes {
		_, errors := validRolePolicyNamePrefix(s, "name_prefix")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid IAM role policy name prefix: %v", s, errors)
		}
	}
}

func TestValidAccountAlias(t *testing.T) {
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
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "http://wrong.scheme.com", // nosemgrep: domain-names
			ErrCount: 1,
		},
		{
			Value:    "ftp://wrong.scheme.co.uk", // nosemgrep: domain-names
			ErrCount: 1,
		},
		{
			Value:    "%@invalidUrl",
			ErrCount: 1,
		},
		{
			Value:    "https://example.com/?query=param",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validOpenIDURL(tc.Value, "url")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d of OpenID URL validation errors, got %d", tc.ErrCount, len(errors))
		}
	}
}
