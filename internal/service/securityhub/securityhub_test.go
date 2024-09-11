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
			acctest.CtBasic:               testAccAccount_basic,
			acctest.CtDisappears:          testAccAccount_disappears,
			"EnableDefaultStandardsFalse": testAccAccount_enableDefaultStandardsFalse,
			"MigrateV0":                   testAccAccount_migrateV0,
			"Full":                        testAccAccount_full,
			"RemoveControlFindingGeneratorDefaultValue": testAccAccount_removeControlFindingGeneratorDefaultValue,
		},
		"AutomationRule": {
			acctest.CtBasic:      testAccAutomationRule_basic,
			"full":               testAccAutomationRule_full,
			acctest.CtDisappears: testAccAutomationRule_disappears,
			"stringFilters":      testAccAutomationRule_stringFilters,
			"numberFilters":      testAccAutomationRule_numberFilters,
			"dateFilters":        testAccAutomationRule_dateFilters,
			"mapFilters":         testAccAutomationRule_mapFilters,
			"tags":               testAccAutomationRule_tags,
		},
		"ActionTarget": {
			acctest.CtBasic:      testAccActionTarget_basic,
			acctest.CtDisappears: testAccActionTarget_disappears,
			"Description":        testAccActionTarget_Description,
			"Name":               testAccActionTarget_Name,
		},
		"ConfigurationPolicy": {
			acctest.CtBasic:      testAccConfigurationPolicy_basic,
			acctest.CtDisappears: testAccConfigurationPolicy_disappears,
			"CustomParameters":   testAccConfigurationPolicy_controlCustomParameters,
			"ControlIdentifiers": testAccConfigurationPolicy_specificControlIdentifiers,
		},
		"ConfigurationPolicyAssociation": {
			acctest.CtBasic:      testAccConfigurationPolicyAssociation_basic,
			acctest.CtDisappears: testAccConfigurationPolicyAssociation_disappears,
		},
		"FindingAggregator": {
			acctest.CtBasic:      testAccFindingAggregator_basic,
			acctest.CtDisappears: testAccFindingAggregator_disappears,
		},
		"Insight": {
			acctest.CtBasic:      testAccInsight_basic,
			acctest.CtDisappears: testAccInsight_disappears,
			"DateFilters":        testAccInsight_DateFilters,
			"GroupByAttribute":   testAccInsight_GroupByAttribute,
			"IpFilters":          testAccInsight_IPFilters,
			"KeywordFilters":     testAccInsight_KeywordFilters,
			"MapFilters":         testAccInsight_MapFilters,
			"MultipleFilters":    testAccInsight_MultipleFilters,
			"Name":               testAccInsight_Name,
			"NumberFilters":      testAccInsight_NumberFilters,
			"WorkflowStatus":     testAccInsight_WorkflowStatus,
		},
		"InviteAccepter": {
			acctest.CtBasic: testAccInviteAccepter_basic,
		},
		"Member": {
			acctest.CtBasic: testAccMember_basic,
			"invite":        testAccMember_invite,
		},
		"OrganizationAdminAccount": {
			acctest.CtBasic:      testAccOrganizationAdminAccount_basic,
			acctest.CtDisappears: testAccOrganizationAdminAccount_disappears,
			"MultiRegion":        testAccOrganizationAdminAccount_MultiRegion,
		},
		"OrganizationConfiguration": {
			acctest.CtBasic:        testAccOrganizationConfiguration_basic,
			"AutoEnableStandards":  testAccOrganizationConfiguration_autoEnableStandards,
			"CentralConfiguration": testAccOrganizationConfiguration_centralConfiguration,
		},
		"ProductSubscription": {
			acctest.CtBasic: testAccProductSubscription_basic,
		},
		"StandardsControl": {
			acctest.CtBasic:                         testAccStandardsControl_basic,
			"DisabledControlStatus":                 testAccStandardsControl_disabledControlStatus,
			"EnabledControlStatusAndDisabledReason": testAccStandardsControl_enabledControlStatusAndDisabledReason,
		},
		"StandardsSubscription": {
			acctest.CtBasic:      testAccStandardsSubscription_basic,
			acctest.CtDisappears: testAccStandardsSubscription_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
