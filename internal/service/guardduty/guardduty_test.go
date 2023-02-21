package guardduty_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccGuardDuty_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Detector": {
			"basic":                             testAccDetector_basic,
			"datasources_s3logs":                testAccDetector_datasources_s3logs,
			"datasources_kubernetes_audit_logs": testAccDetector_datasources_kubernetes_audit_logs,
			"datasources_malware_protection":    testAccDetector_datasources_malware_protection,
			"datasources_all":                   testAccDetector_datasources_all,
			"tags":                              testAccDetector_tags,
			"datasource_basic":                  testAccDetectorDataSource_basic,
			"datasource_id":                     testAccDetectorDataSource_ID,
		},
		"Filter": {
			"basic":      testAccFilter_basic,
			"update":     testAccFilter_update,
			"tags":       testAccFilter_tags,
			"disappears": testAccFilter_disappears,
		},
		"InviteAccepter": {
			"basic": testAccInviteAccepter_basic,
		},
		"IPSet": {
			"basic": testAccIPSet_basic,
			"tags":  testAccIPSet_tags,
		},
		"OrganizationAdminAccount": {
			"basic": testAccOrganizationAdminAccount_basic,
		},
		"OrganizationConfiguration": {
			"basic":             testAccOrganizationConfiguration_basic,
			"s3Logs":            testAccOrganizationConfiguration_s3logs,
			"kubernetes":        testAccOrganizationConfiguration_kubernetes,
			"malwareProtection": testAccOrganizationConfiguration_malwareprotection,
		},
		"ThreatIntelSet": {
			"basic": testAccThreatIntelSet_basic,
			"tags":  testAccThreatIntelSet_tags,
		},
		"Member": {
			"basic":              testAccMember_basic,
			"inviteOnUpdate":     testAccMember_invite_onUpdate,
			"inviteDisassociate": testAccMember_invite_disassociate,
			"invitationMessage":  testAccMember_invitationMessage,
		},
		"PublishingDestination": {
			"basic":      testAccPublishingDestination_basic,
			"disappears": testAccPublishingDestination_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccMemberFromEnv(t *testing.T) (string, string) {
	accountID := os.Getenv("AWS_GUARDDUTY_MEMBER_ACCOUNT_ID")
	if accountID == "" {
		t.Skip(
			"Environment variable AWS_GUARDDUTY_MEMBER_ACCOUNT_ID is not set. " +
				"To properly test inviting GuardDuty member accounts, " +
				"a valid AWS account ID must be provided.")
	}
	email := os.Getenv("AWS_GUARDDUTY_MEMBER_EMAIL")
	if email == "" {
		t.Skip(
			"Environment variable AWS_GUARDDUTY_MEMBER_EMAIL is not set. " +
				"To properly test inviting GuardDuty member accounts, " +
				"a valid email associated with the AWS_GUARDDUTY_MEMBER_ACCOUNT_ID must be provided.")
	}
	return accountID, email
}
