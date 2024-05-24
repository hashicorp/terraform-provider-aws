// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGuardDuty_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Detector": {
			acctest.CtBasic:                     testAccDetector_basic,
			"datasources_s3logs":                testAccDetector_datasources_s3logs,
			"datasources_kubernetes_audit_logs": testAccDetector_datasources_kubernetes_audit_logs,
			"datasources_malware_protection":    testAccDetector_datasources_malware_protection,
			"datasources_all":                   testAccDetector_datasources_all,
			"tags":                              testAccDetector_tags,
			"datasource_basic":                  testAccDetectorDataSource_basic,
			"datasource_id":                     testAccDetectorDataSource_ID,
		},
		"DetectorFeature": {
			acctest.CtBasic:            testAccDetectorFeature_basic,
			"additional_configuration": testAccDetectorFeature_additionalConfiguration,
			"multiple":                 testAccDetectorFeature_multiple,
		},
		"Filter": {
			acctest.CtBasic:      testAccFilter_basic,
			"update":             testAccFilter_update,
			"tags":               testAccFilter_tags,
			acctest.CtDisappears: testAccFilter_disappears,
		},
		"FindingIDs": {
			"datasource_basic": testAccFindingIDsDataSource_basic,
		},
		"InviteAccepter": {
			acctest.CtBasic: testAccInviteAccepter_basic,
		},
		"IPSet": {
			acctest.CtBasic: testAccIPSet_basic,
			"tags":          testAccIPSet_tags,
		},
		"OrganizationAdminAccount": {
			acctest.CtBasic: testAccOrganizationAdminAccount_basic,
		},
		"OrganizationConfiguration": {
			acctest.CtBasic:                 testAccOrganizationConfiguration_basic,
			"autoEnableOrganizationMembers": testAccOrganizationConfiguration_autoEnableOrganizationMembers,
			"s3Logs":                        testAccOrganizationConfiguration_s3logs,
			"kubernetes":                    testAccOrganizationConfiguration_kubernetes,
			"malwareProtection":             testAccOrganizationConfiguration_malwareprotection,
		},
		"OrganizationConfigurationFeature": {
			acctest.CtBasic:            testAccOrganizationConfigurationFeature_basic,
			"additional_configuration": testAccOrganizationConfigurationFeature_additionalConfiguration,
			"multiple":                 testAccOrganizationConfigurationFeature_multiple,
		},
		"ThreatIntelSet": {
			acctest.CtBasic: testAccThreatIntelSet_basic,
			"tags":          testAccThreatIntelSet_tags,
		},
		"Member": {
			acctest.CtBasic:      testAccMember_basic,
			"inviteOnUpdate":     testAccMember_invite_onUpdate,
			"inviteDisassociate": testAccMember_invite_disassociate,
			"invitationMessage":  testAccMember_invitationMessage,
		},
		"PublishingDestination": {
			acctest.CtBasic:      testAccPublishingDestination_basic,
			acctest.CtDisappears: testAccPublishingDestination_disappears,
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

// testAccPreCheckDetectorExists verifies the current account has a single active GuardDuty detector configured.
func testAccPreCheckDetectorExists(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

	_, err := tfguardduty.FindDetector(ctx, conn)

	if tfresource.NotFound(err) {
		t.Skipf("reading this AWS account's single GuardDuty Detector: %s", err)
	}

	if err != nil {
		t.Fatalf("listing GuardDuty Detectors: %s", err)
	}
}

// testAccPreCheckDetectorNotExists verifies the current account has no active GuardDuty detector configured.
func testAccPreCheckDetectorNotExists(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

	_, err := tfguardduty.FindDetector(ctx, conn)

	if tfresource.NotFound(err) {
		return
	}

	if err != nil {
		t.Fatalf("listing GuardDuty Detectors: %s", err)
	}

	t.Skip("this AWS account has a GuardDuty Detector")
}
