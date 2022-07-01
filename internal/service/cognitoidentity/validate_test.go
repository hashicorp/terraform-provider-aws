package cognitoidentity

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentity"
)

func TestValidIdentityPoolName(t *testing.T) {
	validValues := []string{
		"123",
		"1 2 3",
		"foo",
		"foo bar",
		"foo_bar",
		"1foo 2bar 3",
		"foo-bar_123",
		"foo-bar",
	}

	for _, s := range validValues {
		_, errors := validIdentityPoolName(s, "identity_pool_name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Identity Pool Name: %v", s, errors)
		}
	}

	invalidValues := []string{
		"foo*",
		"foo:bar",
		"foo&bar",
		"foo1^bar2",
	}

	for _, s := range invalidValues {
		_, errors := validIdentityPoolName(s, "identity_pool_name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Identity Pool Name: %v", s, errors)
		}
	}
}

func TestValidIdentityProvidersClientID(t *testing.T) {
	validValues := []string{
		"7lhlkkfbfb4q5kpp90urffao",
		"12345678",
		"foo_123",
		strings.Repeat("W", 128),
	}

	for _, s := range validValues {
		_, errors := validIdentityProvidersClientID(s, "client_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Identity Provider Client ID: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		strings.Repeat("W", 129), // > 128
		"foo-bar",
		"foo:bar",
		"foo;bar",
	}

	for _, s := range invalidValues {
		_, errors := validIdentityProvidersClientID(s, "client_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Identity Provider Client ID: %v", s, errors)
		}
	}
}

func TestValidIdentityProvidersProviderName(t *testing.T) {
	validValues := []string{
		"foo",
		"7346241598935552",
		"foo_bar",
		"foo:bar",
		"foo/bar",
		"foo-bar",
		"cognito-idp.us-east-1.amazonaws.com/us-east-1_Zr231apJu", //lintignore:AWSAT003
		strings.Repeat("W", 128),
	}

	for _, s := range validValues {
		_, errors := validIdentityProvidersProviderName(s, "provider_name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Identity Provider Name: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		strings.Repeat("W", 129), // > 128
		"foo;bar_baz",
		"foobar,foobaz",
		"foobar=foobaz",
	}

	for _, s := range invalidValues {
		_, errors := validIdentityProvidersProviderName(s, "provider_name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Identity Provider Name: %v", s, errors)
		}
	}
}

func TestValidProviderDeveloperName(t *testing.T) {
	validValues := []string{
		"1",
		"foo",
		"1.2",
		"foo1-bar2-baz3",
		"foo_bar",
	}

	for _, s := range validValues {
		_, errors := validProviderDeveloperName(s, "developer_provider_name")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Provider Developer Name: %v", s, errors)
		}
	}

	invalidValues := []string{
		"foo!",
		"foo:bar",
		"foo/bar",
		"foo;bar",
	}

	for _, s := range invalidValues {
		_, errors := validProviderDeveloperName(s, "developer_provider_name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Provider Developer Name: %v", s, errors)
		}
	}
}

func TestValidRoleMappingsAmbiguousRoleResolutionAgainstType(t *testing.T) {
	cases := []struct {
		AmbiguousRoleResolution interface{}
		Type                    string
		ErrCount                int
	}{
		{
			AmbiguousRoleResolution: nil,
			Type:                    cognitoidentity.RoleMappingTypeToken,
			ErrCount:                1,
		},
		{
			AmbiguousRoleResolution: "foo",
			Type:                    cognitoidentity.RoleMappingTypeToken,
			ErrCount:                0, // 0 as it should be defined, the value isn't validated here
		},
		{
			AmbiguousRoleResolution: cognitoidentity.AmbiguousRoleResolutionTypeAuthenticatedRole,
			Type:                    cognitoidentity.RoleMappingTypeToken,
			ErrCount:                0,
		},
		{
			AmbiguousRoleResolution: cognitoidentity.AmbiguousRoleResolutionTypeDeny,
			Type:                    cognitoidentity.RoleMappingTypeToken,
			ErrCount:                0,
		},
	}

	for _, tc := range cases {
		m := make(map[string]interface{})
		// Reproducing the undefined ambiguous_role_resolution
		if tc.AmbiguousRoleResolution != nil {
			m["ambiguous_role_resolution"] = tc.AmbiguousRoleResolution
		}
		m["type"] = tc.Type

		errors := validRoleMappingsAmbiguousRoleResolutionAgainstType(m)
		if len(errors) != tc.ErrCount {
			t.Fatalf("Cognito Role Mappings validation failed: %v, expected err count %d, got %d, for config %#v", errors, tc.ErrCount, len(errors), m)
		}
	}
}

func TestValidRoleMappingsRulesConfiguration(t *testing.T) {
	cases := []struct {
		MappingRule []interface{}
		Type        string
		ErrCount    int
	}{
		{
			MappingRule: nil,
			Type:        cognitoidentity.RoleMappingTypeRules,
			ErrCount:    1,
		},
		{
			MappingRule: []interface{}{
				map[string]interface{}{
					"Claim":     "isAdmin",
					"MatchType": "Equals",
					"RoleARN":   "arn:foo",
					"Value":     "paid",
				},
			},
			Type:     cognitoidentity.RoleMappingTypeRules,
			ErrCount: 0,
		},
		{
			MappingRule: []interface{}{
				map[string]interface{}{
					"Claim":     "isAdmin",
					"MatchType": "Equals",
					"RoleARN":   "arn:foo",
					"Value":     "paid",
				},
			},
			Type:     cognitoidentity.RoleMappingTypeToken,
			ErrCount: 1,
		},
		{
			MappingRule: nil,
			Type:        cognitoidentity.RoleMappingTypeToken,
			ErrCount:    0,
		},
	}

	for _, tc := range cases {
		m := make(map[string]interface{})
		// Reproducing the undefined mapping_rule
		if tc.MappingRule != nil {
			m["mapping_rule"] = tc.MappingRule
		}
		m["type"] = tc.Type

		errors := validRoleMappingsRulesConfiguration(m)
		if len(errors) != tc.ErrCount {
			t.Fatalf("Cognito Role Mappings validation failed: %v, expected err count %d, got %d, for config %#v", errors, tc.ErrCount, len(errors), m)
		}
	}
}

func TestValidRoles(t *testing.T) {
	validValues := []map[string]interface{}{
		{"authenticated": "hoge"},
		{"unauthenticated": "hoge"},
		{"authenticated": "hoge", "unauthenticated": "hoge"},
	}

	for _, s := range validValues {
		errors := validRoles(s)
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Roles: %v", s, errors)
		}
	}

	invalidValues := []map[string]interface{}{
		{},
		{"invalid": "hoge"},
	}

	for _, s := range invalidValues {
		errors := validRoles(s)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Roles: %v", s, errors)
		}
	}
}

func TestValidSupportedLoginProviders(t *testing.T) {
	validValues := []string{
		"foo",
		"7346241598935552",
		"123456789012.apps.googleusercontent.com", // nosemgrep: domain-names
		"foo_bar",
		"foo;bar",
		"foo/bar",
		"foo-bar",
		"xvz1evFS4wEEPTGEFPHBog;kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw",
		strings.Repeat("W", 128),
	}

	for _, s := range validValues {
		_, errors := validSupportedLoginProviders(s, "supported_login_providers")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito Supported Login Providers: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		strings.Repeat("W", 129), // > 128
		"foo:bar_baz",
		"foobar,foobaz",
		"foobar=foobaz",
	}

	for _, s := range invalidValues {
		_, errors := validSupportedLoginProviders(s, "supported_login_providers")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito Supported Login Providers: %v", s, errors)
		}
	}
}
