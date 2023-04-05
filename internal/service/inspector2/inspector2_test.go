package inspector2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccInspector2_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Enabler": {
			"basic":      testAccEnabler_basic,
			"accountID":  testAccEnabler_accountID,
			"disappears": testAccEnabler_disappears,
		},
		"DelegatedAdminAccount": {
			"basic":      testAccDelegatedAdminAccount_basic,
			"disappears": testAccDelegatedAdminAccount_disappears,
		},
		"MemberAssociation": {
			"basic":      testAccMemberAssociation_basic,
			"disappears": testAccMemberAssociation_disappears,
		},
		"OrganizationConfiguration": {
			"basic":      testAccOrganizationConfiguration_basic,
			"disappears": testAccOrganizationConfiguration_disappears,
			"ec2ECR":     testAccOrganizationConfiguration_ec2ECR,
			"lambda":     testAccOrganizationConfiguration_lambda,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
