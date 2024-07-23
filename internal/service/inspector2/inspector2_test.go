// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccInspector2_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Enabler": {
			acctest.CtBasic:                      testAccEnabler_basic,
			"accountID":                          testAccEnabler_accountID,
			acctest.CtDisappears:                 testAccEnabler_disappears,
			"lambda":                             testAccEnabler_lambda,
			"updateResourceTypes":                testAccEnabler_updateResourceTypes,
			"updateResourceTypes_disjoint":       testAccEnabler_updateResourceTypes_disjoint,
			"memberAccount_basic":                testAccEnabler_memberAccount_basic,
			"memberAccount_multiple":             testAccEnabler_memberAccount_multiple,
			"memberAccount_updateMemberAccounts": testAccEnabler_memberAccount_updateMemberAccounts,
			"memberAccount_updateMemberAccountsAndScanTypes": testAccEnabler_memberAccount_updateMemberAccountsAndScanTypes,
			"memberAccount_disappearsMemberAssociation":      testAccEnabler_memberAccount_disappearsMemberAssociation,
		},
		"DelegatedAdminAccount": {
			acctest.CtBasic:      testAccDelegatedAdminAccount_basic,
			acctest.CtDisappears: testAccDelegatedAdminAccount_disappears,
		},
		"MemberAssociation": {
			acctest.CtBasic:      testAccMemberAssociation_basic,
			acctest.CtDisappears: testAccMemberAssociation_disappears,
		},
		"OrganizationConfiguration": {
			acctest.CtBasic:      testAccOrganizationConfiguration_basic,
			acctest.CtDisappears: testAccOrganizationConfiguration_disappears,
			"ec2ECR":             testAccOrganizationConfiguration_ec2ECR,
			"lambda":             testAccOrganizationConfiguration_lambda,
			"lambdaCode":         testAccOrganizationConfiguration_lambdaCode,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
