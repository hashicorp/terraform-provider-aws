// Copyright IBM Corp. 2014, 2026
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
		"AccountV2": {
			acctest.CtBasic:      testAccAccountV2_basic,
			acctest.CtDisappears: testAccAccountV2_disappears,
			"tags":               testAccAccountV2_tags,
			"Identity":           testAccSecurityHubAccountV2_identitySerial,
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
			"Identity":           testAccSecurityHubAutomationRule_identitySerial,
		},
		"ActionTarget": {
			acctest.CtBasic:      testAccActionTarget_basic,
			acctest.CtDisappears: testAccActionTarget_disappears,
			"Description":        testAccActionTarget_Description,
			"Name":               testAccActionTarget_Name,
			"Identity":           testAccSecurityHubActionTarget_identitySerial,
		},
		"ConfigurationPolicy": {
			acctest.CtBasic:      testAccConfigurationPolicy_basic,
			acctest.CtDisappears: testAccConfigurationPolicy_disappears,
			"CustomParameters":   testAccConfigurationPolicy_controlCustomParameters,
			"ControlIdentifiers": testAccConfigurationPolicy_specificControlIdentifiers,
		},
		"ConfigurationPolicyAssociation": {
			acctest.CtBasic:          testAccConfigurationPolicyAssociation_basic,
			acctest.CtDisappears:     testAccConfigurationPolicyAssociation_disappears,
			"SelfManagedSecurityHub": testAccConfigurationPolicyAssociation_selfManagedSecurityHub,
		},
		"EnabledStandards": {
			acctest.CtBasic:            testAccEnabledStandardsDataSource_basic,
			"StandardsSubscriptionARN": testAccEnabledStandardsDataSource_standardsSubscriptionARN,
		},
		"FindingAggregator": {
			acctest.CtBasic:      testAccFindingAggregator_basic,
			acctest.CtDisappears: testAccFindingAggregator_disappears,
			"Identity":           testAccSecurityHubFindingAggregator_identitySerial,
		},
		"Insight": {
			acctest.CtBasic:       testAccInsight_basic,
			acctest.CtDisappears:  testAccInsight_disappears,
			"DateFilters":         testAccInsight_DateFilters,
			"GroupByAttribute":    testAccInsight_GroupByAttribute,
			"IpFilters":           testAccInsight_IPFilters,
			"KeywordFilters":      testAccInsight_KeywordFilters,
			"MapFilters":          testAccInsight_MapFilters,
			"MultipleFilters":     testAccInsight_MultipleFilters,
			"Name":                testAccInsight_Name,
			"NumberFilters":       testAccInsight_NumberFilters,
			"WorkflowStatus":      testAccInsight_WorkflowStatus,
			"StringFilters":       testAccInsight_StringFilters,
			"Identity":            testAccSecurityHubInsight_identitySerial,
			"ListBasic":           testAccInsight_List_basic,
			"ListIncludeResource": testAccInsight_List_includeResource,
			"ListRegionOverride":  testAccInsight_List_regionOverride,
		},
		"InviteAccepter": {
			acctest.CtBasic: testAccInviteAccepter_basic,
		},
		"Member": {
			acctest.CtBasic:            testAccMember_basic,
			acctest.CtDisappears:       testAccMember_disappears,
			"inviteTrue":               testAccMember_inviteTrue,
			"inviteFalse":              testAccMember_inviteFalse,
			"inviteOrganizationMember": testAccMember_inviteOrganizationMember,
			"Identity":                 testAccSecurityHubMember_identitySerial,
		},
		"OrganizationAdminAccount": {
			acctest.CtBasic:      testAccOrganizationAdminAccount_basic,
			acctest.CtDisappears: testAccOrganizationAdminAccount_disappears,
			"MultiRegion":        testAccOrganizationAdminAccount_MultiRegion,
			"Identity":           testAccSecurityHubOrganizationAdminAccount_identitySerial,
		},
		"OrganizationConfiguration": {
			acctest.CtBasic:        testAccOrganizationConfiguration_basic,
			"AutoEnableStandards":  testAccOrganizationConfiguration_autoEnableStandards,
			"CentralConfiguration": testAccOrganizationConfiguration_centralConfiguration,
		},
		"ProductSubscription": {
			acctest.CtBasic:      testAccProductSubscription_basic,
			acctest.CtDisappears: testAccProductSubscription_disappears,
			"Identity":           testAccSecurityHubProductSubscription_identitySerial,
		},
		"StandardsControl": {
			acctest.CtBasic:                         testAccStandardsControl_basic,
			"DisabledControlStatus":                 testAccStandardsControl_disabledControlStatus,
			"EnabledControlStatusAndDisabledReason": testAccStandardsControl_enabledControlStatusAndDisabledReason,
			"Identity":                              testAccSecurityHubStandardsControl_identitySerial,
		},
		"StandardsControlAssociation": {
			acctest.CtBasic: testAccStandardsControlAssociation_basic,
			"Identity":      testAccSecurityHubStandardsControlAssociation_identitySerial,
		},
		"StandardsControlAssociationsDataSource": {
			acctest.CtBasic: testAccStandardsControlAssociationsDataSource_basic,
		},
		"StandardsSubscription": {
			acctest.CtBasic:      testAccStandardsSubscription_basic,
			acctest.CtDisappears: testAccStandardsSubscription_disappears,
			"Identity":           testAccSecurityHubStandardsSubscription_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
