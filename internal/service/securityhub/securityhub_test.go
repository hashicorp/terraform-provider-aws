package securityhub_test

import (
	"testing"
)

func TestAccSecurityHub_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic": testAccAccount_basic,
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
			"basic": testAccOrganizationConfiguration_basic,
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
