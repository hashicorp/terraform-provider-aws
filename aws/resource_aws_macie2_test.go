package aws

import (
	"os"
	"testing"
)

func TestAccAWSMacie2_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic":                        testAccAwsMacie2Account_basic,
			"finding_publishing_frequency": testAccAwsMacie2Account_FindingPublishingFrequency,
			"status":                       testAccAwsMacie2Account_WithStatus,
			"finding_and_status":           testAccAwsMacie2Account_WithFindingAndStatus,
			"disappears":                   testAccAwsMacie2Account_disappears,
		},
		"ClassificationJob": {
			"basic":          testAccAwsMacie2ClassificationJob_basic,
			"name_generated": testAccAwsMacie2ClassificationJob_Name_Generated,
			"name_prefix":    testAccAwsMacie2ClassificationJob_NamePrefix,
			"disappears":     testAccAwsMacie2ClassificationJob_disappears,
			"status":         testAccAwsMacie2ClassificationJob_Status,
			"complete":       testAccAwsMacie2ClassificationJob_complete,
			"tags":           testAccAwsMacie2ClassificationJob_WithTags,
		},
		"CustomDataIdentifier": {
			"basic":              testAccAwsMacie2CustomDataIdentifier_basic,
			"name_generated":     testAccAwsMacie2CustomDataIdentifier_Name_Generated,
			"name_prefix":        testAccAwsMacie2CustomDataIdentifier_disappears,
			"disappears":         testAccAwsMacie2CustomDataIdentifier_NamePrefix,
			"classification_job": testAccAwsMacie2CustomDataIdentifier_WithClassificationJob,
			"tags":               testAccAwsMacie2CustomDataIdentifier_WithTags,
		},
		"FindingsFilter": {
			"basic":          testAccAwsMacie2FindingsFilter_basic,
			"name_generated": testAccAwsMacie2FindingsFilter_Name_Generated,
			"name_prefix":    testAccAwsMacie2FindingsFilter_NamePrefix,
			"disappears":     testAccAwsMacie2FindingsFilter_disappears,
			"complete":       testAccAwsMacie2FindingsFilter_complete,
			"date":           testAccAwsMacie2FindingsFilter_WithDate,
			"number":         testAccAwsMacie2FindingsFilter_WithNumber,
			"tags":           testAccAwsMacie2FindingsFilter_withTags,
		},
		"OrganizationAdminAccount": {
			"basic":      testAccAwsMacie2OrganizationAdminAccount_basic,
			"disappears": testAccAwsMacie2OrganizationAdminAccount_disappears,
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

func testAccAWSMacie2MemberFromEnv(t *testing.T) (string, string) {
	accountID := os.Getenv("AWS_MACIE_MEMBER_ACCOUNT_ID")
	if accountID == "" {
		t.Skip(
			"Environment variable AWS_MACIE_MEMBER_ACCOUNT_ID is not set. " +
				"To properly test inviting MACIE member accounts, " +
				"a valid AWS account ID must be provided.")
	}
	email := os.Getenv("AWS_MACIE_MEMBER_EMAIL")
	if email == "" {
		t.Skip(
			"Environment variable AWS_MACIE_MEMBER_EMAIL is not set. " +
				"To properly test inviting MACIE member accounts, " +
				"a valid email associated with the AWS_MACIE_MEMBER_ACCOUNT_ID must be provided.")
	}
	return accountID, email
}
