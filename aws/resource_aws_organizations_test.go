package aws

import (
	"testing"
)

func TestAccAWSOrganizations(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Organization": {
			"basic":                      testAccAwsOrganizationsOrganization_basic,
			"AwsServiceAccessPrincipals": testAccAwsOrganizationsOrganization_AwsServiceAccessPrincipals,
			"EnabledPolicyTypes":         testAccAwsOrganizationsOrganization_EnabledPolicyTypes,
			"FeatureSet":                 testAccAwsOrganizationsOrganization_FeatureSet,
		},
		"Account": {
			"basic": testAccAwsOrganizationsAccount_basic,
		},
		"OrganizationalUnit": {
			"basic": testAccAwsOrganizationsOrganizationalUnit_basic,
			"Name":  testAccAwsOrganizationsOrganizationalUnit_Name,
		},
		"PolicyAttachment": {
			"Account":            testAccAwsOrganizationsPolicyAttachment_Account,
			"OrganizationalUnit": testAccAwsOrganizationsPolicyAttachment_OrganizationalUnit,
			"Root":               testAccAwsOrganizationsPolicyAttachment_Root,
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
