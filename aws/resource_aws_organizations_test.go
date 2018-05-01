package aws

import (
	"testing"
)

func TestAccAWSOrganizations(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Organization": {
			"basic":                      testAccAwsOrganizationsOrganization_basic,
			"AwsServiceAccessPrincipals": testAccAwsOrganizationsOrganization_AwsServiceAccessPrincipals,
			"FeatureSet":                 testAccAwsOrganizationsOrganization_FeatureSet,
		},
		"Account": {
			"basic":      testAccAwsOrganizationsAccount_basic,
			"parentRoot": testAccAwsOrganizationsAccount_parentRoot,
			"parentOU":   testAccAwsOrganizationsAccount_parentOU,
		},
		"Unit": {
			"basic":       testAccAwsOrganizationsUnit_basic,
			"importBasic": testAccAwsOrganizationsUnit_importBasic,
			"update":      testAccAwsOrganizationsUnitUpdate,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
