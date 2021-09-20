package aws

import (
	"testing"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAWSSecurityHub_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic": testAccAWSSecurityHubAccount_basic,
		},
		"Member": {
			"basic":  testAccAWSSecurityHubMember_basic,
			"invite": testAccAWSSecurityHubMember_invite,
		},
		"ActionTarget": {
			"basic":       testAccAwsSecurityHubActionTarget_basic,
			"disappears":  testAccAwsSecurityHubActionTarget_disappears,
			"Description": testAccAwsSecurityHubActionTarget_Description,
			"Name":        testAccAwsSecurityHubActionTarget_Name,
		},
		"Insight": {
			"basic":            testAccAwsSecurityHubInsight_basic,
			"disappears":       testAccAwsSecurityHubInsight_disappears,
			"DateFilters":      testAccAwsSecurityHubInsight_DateFilters,
			"GroupByAttribute": testAccAwsSecurityHubInsight_GroupByAttribute,
			"IpFilters":        testAccAwsSecurityHubInsight_IpFilters,
			"KeywordFilters":   testAccAwsSecurityHubInsight_KeywordFilters,
			"MapFilters":       testAccAwsSecurityHubInsight_MapFilters,
			"MultipleFilters":  testAccAwsSecurityHubInsight_MultipleFilters,
			"Name":             testAccAwsSecurityHubInsight_Name,
			"NumberFilters":    testAccAwsSecurityHubInsight_NumberFilters,
			"WorkflowStatus":   testAccAwsSecurityHubInsight_WorkflowStatus,
		},
		"InviteAccepter": {
			"basic": testAccAWSSecurityHubInviteAccepter_basic,
		},
		"OrganizationAdminAccount": {
			"basic":       testAccAwsSecurityHubOrganizationAdminAccount_basic,
			"disappears":  testAccAwsSecurityHubOrganizationAdminAccount_disappears,
			"MultiRegion": testAccAwsSecurityHubOrganizationAdminAccount_MultiRegion,
		},
		"OrganizationConfiguration": {
			"basic": testAccAwsSecurityHubOrganizationConfiguration_basic,
		},
		"ProductSubscription": {
			"basic": testAccAWSSecurityHubProductSubscription_basic,
		},
		"StandardsControl": {
			"basic":                                 testAccAWSSecurityHubStandardsControl_basic,
			"DisabledControlStatus":                 testAccAWSSecurityHubStandardsControl_disabledControlStatus,
			"EnabledControlStatusAndDisabledReason": testAccAWSSecurityHubStandardsControl_enabledControlStatusAndDisabledReason,
		},
		"StandardsSubscription": {
			"basic":      testAccAWSSecurityHubStandardsSubscription_basic,
			"disappears": testAccAWSSecurityHubStandardsSubscription_disappears,
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
