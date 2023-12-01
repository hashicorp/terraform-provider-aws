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
			"basic":                       TestAccSecurityHubAccount_basic,
			"disappears":                  TestAccSecurityHubAccount_disappears,
			"EnableDefaultStandardsFalse": TestAccSecurityHubAccount_enableDefaultStandardsFalse,
			"MigrateV0":                   TestAccSecurityHubAccount_migrateV0,
			"Full":                        TestAccSecurityHubAccount_full,
			"RemoveControlFindingGeneratorDefaultValue": TestAccSecurityHubAccount_removeControlFindingGeneratorDefaultValue,
		},
		"Member": {
			"basic":  TestAccSecurityHubMember_basic,
			"invite": TestAccSecurityHubMember_invite,
		},
		"ActionTarget": {
			"basic":       TestAccSecurityHubActionTarget_basic,
			"disappears":  TestAccSecurityHubActionTarget_disappears,
			"Description": TestAccSecurityHubActionTarget_Description,
			"Name":        TestAccSecurityHubActionTarget_Name,
		},
		"Insight": {
			"basic":            TestAccSecurityHubInsight_basic,
			"disappears":       TestAccSecurityHubInsight_disappears,
			"DateFilters":      TestAccSecurityHubInsight_DateFilters,
			"GroupByAttribute": TestAccSecurityHubInsight_GroupByAttribute,
			"IpFilters":        TestAccSecurityHubInsight_IPFilters,
			"KeywordFilters":   TestAccSecurityHubInsight_KeywordFilters,
			"MapFilters":       TestAccSecurityHubInsight_MapFilters,
			"MultipleFilters":  TestAccSecurityHubInsight_MultipleFilters,
			"Name":             TestAccSecurityHubInsight_Name,
			"NumberFilters":    TestAccSecurityHubInsight_NumberFilters,
			"WorkflowStatus":   TestAccSecurityHubInsight_WorkflowStatus,
		},
		"InviteAccepter": {
			"basic": TestAccSecurityHubInviteAccepter_basic,
		},
		"OrganizationAdminAccount": {
			"basic":       TestAccSecurityHubOrganizationAdminAccount_basic,
			"disappears":  TestAccSecurityHubOrganizationAdminAccount_disappears,
			"MultiRegion": TestAccSecurityHubOrganizationAdminAccount_MultiRegion,
		},
		"OrganizationConfiguration": {
			"basic":               TestAccSecurityHubOrganizationConfiguration_basic,
			"AutoEnableStandards": TestAccSecurityHubOrganizationConfiguration_autoEnableStandards,
		},
		"ProductSubscription": {
			"basic": TestAccSecurityHubProductSubscription_basic,
		},
		"StandardsControl": {
			"basic":                                 TestAccSecurityHubStandardsControl_basic,
			"DisabledControlStatus":                 TestAccSecurityHubStandardsControl_disabledControlStatus,
			"EnabledControlStatusAndDisabledReason": TestAccSecurityHubStandardsControl_enabledControlStatusAndDisabledReason,
		},
		"StandardsSubscription": {
			"basic":      TestAccSecurityHubStandardsSubscription_basic,
			"disappears": TestAccSecurityHubStandardsSubscription_disappears,
		},
		"FindingAggregator": {
			"basic":      TestAccSecurityHubFindingAggregator_basic,
			"disappears": TestAccSecurityHubFindingAggregator_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
