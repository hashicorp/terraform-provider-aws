// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSecurityHub_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic":                       testAccAccount_basic,
			"disappears":                  testAccAccount_disappears,
			"EnableDefaultStandardsFalse": testAccAccount_enableDefaultStandardsFalse,
			"MigrateV0":                   testAccAccount_migrateV0,
			"Full":                        testAccAccount_full,
			"RemoveControlFindingGeneratorDefaultValue": testAccAccount_removeControlFindingGeneratorDefaultValue,
		},
		"Member": {
			"basic":  testAccMember_basic,
			"invite": testAccMember_invite,
		},
		"ActionTarget": {
			"basic":       testAccActionTarget_basic,
			"disappears":  testAccActionTarget_disappears,
			"Description": testAccActionTarget_Description,
			"Name":        testAccActionTarget_Name,
		},
		"Insight": {
			"basic":            testAccInsight_basic,
			"disappears":       testAccInsight_disappears,
			"DateFilters":      testAccInsight_DateFilters,
			"GroupByAttribute": testAccInsight_GroupByAttribute,
			"IpFilters":        testAccInsight_IPFilters,
			"KeywordFilters":   testAccInsight_KeywordFilters,
			"MapFilters":       testAccInsight_MapFilters,
			"MultipleFilters":  testAccInsight_MultipleFilters,
			"Name":             testAccInsight_Name,
			"NumberFilters":    testAccInsight_NumberFilters,
			"WorkflowStatus":   testAccInsight_WorkflowStatus,
		},
		"InviteAccepter": {
			"basic": testAccInviteAccepter_basic,
		},
		"OrganizationAdminAccount": {
			"basic":       testAccOrganizationAdminAccount_basic,
			"disappears":  testAccOrganizationAdminAccount_disappears,
			"MultiRegion": testAccOrganizationAdminAccount_MultiRegion,
		},
		"OrganizationConfiguration": {
			"basic":               testAccOrganizationConfiguration_basic,
			"AutoEnableStandards": testAccOrganizationConfiguration_autoEnableStandards,
		},
		"ProductSubscription": {
			"basic": testAccProductSubscription_basic,
		},
		"StandardsControl": {
			"basic":                                 testAccStandardsControl_basic,
			"DisabledControlStatus":                 testAccStandardsControl_disabledControlStatus,
			"EnabledControlStatusAndDisabledReason": testAccStandardsControl_enabledControlStatusAndDisabledReason,
		},
		"StandardsSubscription": {
			"basic":      testAccStandardsSubscription_basic,
			"disappears": testAccStandardsSubscription_disappears,
		},
		"FindingAggregator": {
			"basic":      testAccFindingAggregator_basic,
			"disappears": testAccFindingAggregator_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

// TestAccSecurityHub_centralConfiguration is a multi-account test stuite for central configuration features.
// Central configuration can only be enabled from a *member* delegated admin account.
// The primary provider is expected to be an organizations member account and the alternate provider is expected to be the organizations management account.
func TestAccSecurityHub_centralConfiguration(t *testing.T) {
	t.Parallel()
	testCases := map[string]map[string]func(t *testing.T){
		"OrganizationConfiguration": {
			"centralConfiguration": testAccOrganizationConfiguration_centralConfiguration,
		},
		"ConfigurationPolicy": {
			"basic":              testAccConfigurationPolicy_basic,
			"customParameters":   testAccConfigurationPolicy_controlCustomParameters,
			"controlIdentifiers": testAccConfigurationPolicy_specificControlIdentifiers,
		},
		"ConfigurationPolicyAssociation": {
			"basic": testAccConfigurationPolicyAssociation_basic,
		},
	}
	acctest.RunSerialTests2Levels(t, testCases, 0)
}

const testAccMemberAccountDelegatedAdminConfig_base = `
resource "aws_securityhub_account" "test" {}

data "aws_caller_identity" "member" {}

resource "aws_securityhub_organization_admin_account" "test" {
  provider = awsalternate

  admin_account_id = data.aws_caller_identity.member.account_id

  depends_on = [aws_securityhub_account.test]
}
`
