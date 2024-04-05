// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMacie2_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic":                        testAccAccount_basic,
			"finding_publishing_frequency": testAccAccount_FindingPublishingFrequency,
			"status":                       testAccAccount_WithStatus,
			"finding_and_status":           testAccAccount_WithFindingAndStatus,
			"disappears":                   testAccAccount_disappears,
		},
		"ClassificationExportConfiguration": {
			"basic": testAccClassificationExportConfiguration_basic,
		},
		"ClassificationJob": {
			"basic":           testAccClassificationJob_basic,
			"name_generated":  testAccClassificationJob_Name_Generated,
			"name_prefix":     testAccClassificationJob_NamePrefix,
			"disappears":      testAccClassificationJob_disappears,
			"status":          testAccClassificationJob_Status,
			"complete":        testAccClassificationJob_complete,
			"tags":            testAccClassificationJob_WithTags,
			"bucket_criteria": testAccClassificationJob_BucketCriteria,
		},
		"CustomDataIdentifier": {
			"basic":              testAccCustomDataIdentifier_basic,
			"name_generated":     testAccCustomDataIdentifier_Name_Generated,
			"name_prefix":        testAccCustomDataIdentifier_disappears,
			"disappears":         testAccCustomDataIdentifier_NamePrefix,
			"classification_job": testAccCustomDataIdentifier_WithClassificationJob,
			"tags":               testAccCustomDataIdentifier_WithTags,
		},
		"FindingsFilter": {
			"basic":          testAccFindingsFilter_basic,
			"name_generated": testAccFindingsFilter_Name_Generated,
			"name_prefix":    testAccFindingsFilter_NamePrefix,
			"disappears":     testAccFindingsFilter_disappears,
			"complete":       testAccFindingsFilter_complete,
			"date":           testAccFindingsFilter_WithDate,
			"number":         testAccFindingsFilter_WithNumber,
			"tags":           testAccFindingsFilter_withTags,
		},
		"OrganizationAdminAccount": {
			"basic":      testAccOrganizationAdminAccount_basic,
			"disappears": testAccOrganizationAdminAccount_disappears,
		},
		"Member": {
			"basic":                                 testAccMember_basic,
			"disappears":                            testAccMember_disappears,
			"tags":                                  testAccMember_withTags,
			"invitation_disable_email_notification": testAccMember_invitationDisableEmailNotification,
			"invite":                                testAccMember_invite,
			"invite_removed":                        testAccMember_inviteRemoved,
			"status":                                testAccMember_status,
		},
		"InvitationAccepter": {
			"basic": testAccInvitationAccepter_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
