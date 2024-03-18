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
		"AutomationRule": {
			"basic":         testAccAutomationRule_basic,
			"full":          testAccAutomationRule_full,
			"disappears":    testAccAutomationRule_disappears,
			"stringFilters": testAccAutomationRule_stringFilters,
			"numberFilters": testAccAutomationRule_numberFilters,
			"dateFilters":   testAccAutomationRule_dateFilters,
			"mapFilters":    testAccAutomationRule_mapFilters,
			"tags":          testAccAutomationRule_tags,
		},
		"ActionTarget": {
			"basic":       testAccActionTarget_basic,
			"disappears":  testAccActionTarget_disappears,
			"Description": testAccActionTarget_Description,
			"Name":        testAccActionTarget_Name,
		},
		"ConfigurationPolicy": {
			"basic":              testAccConfigurationPolicy_basic,
			"disappears":         testAccConfigurationPolicy_disappears,
			"CustomParameters":   testAccConfigurationPolicy_controlCustomParameters,
			"ControlIdentifiers": testAccConfigurationPolicy_specificControlIdentifiers,
		},
		"ConfigurationPolicyAssociation": {
			"basic":      testAccConfigurationPolicyAssociation_basic,
			"disappears": testAccConfigurationPolicyAssociation_disappears,
		},
		"FindingAggregator": {
			"basic":      testAccFindingAggregator_basic,
			"disappears": testAccFindingAggregator_disappears,
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
		"Member": {
			"basic":  testAccMember_basic,
			"invite": testAccMember_invite,
		},
		"OrganizationAdminAccount": {
			"basic":       testAccOrganizationAdminAccount_basic,
			"disappears":  testAccOrganizationAdminAccount_disappears,
			"MultiRegion": testAccOrganizationAdminAccount_MultiRegion,
		},
		"OrganizationConfiguration": {
			"basic":                testAccOrganizationConfiguration_basic,
			"AutoEnableStandards":  testAccOrganizationConfiguration_autoEnableStandards,
			"CentralConfiguration": testAccOrganizationConfiguration_centralConfiguration,
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
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
