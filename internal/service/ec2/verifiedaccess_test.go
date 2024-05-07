// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccess_serial(t *testing.T) {
	t.Parallel()

	semaphore := tfsync.GetSemaphore("VerifiedAccess", "AWS_EC2_VERIFIED_ACCESS_INSTANCE_LIMIT", 5)
	testCases := map[string]map[string]func(*testing.T, tfsync.Semaphore){
		"Endpoint": {
			"basic":            testAccVerifiedAccessEndpoint_basic,
			"networkInterface": testAccVerifiedAccessEndpoint_networkInterface,
			names.AttrTags:     testAccVerifiedAccessEndpoint_tags,
			"disappears":       testAccVerifiedAccessEndpoint_disappears,
			"policyDocument":   testAccVerifiedAccessEndpoint_policyDocument,
		},
		"Group": {
			"basic":          testAccVerifiedAccessGroup_basic,
			"kms":            testAccVerifiedAccessGroup_kms,
			"updateKMS":      testAccVerifiedAccessGroup_updateKMS,
			"disappears":     testAccVerifiedAccessGroup_disappears,
			names.AttrTags:   testAccVerifiedAccessGroup_tags,
			names.AttrPolicy: testAccVerifiedAccessGroup_policy,
			"updatePolicy":   testAccVerifiedAccessGroup_updatePolicy,
			"setPolicy":      testAccVerifiedAccessGroup_setPolicy,
		},
		"Instance": {
			"basic":               testAccVerifiedAccessInstance_basic,
			names.AttrDescription: testAccVerifiedAccessInstance_description,
			"fipsEnabled":         testAccVerifiedAccessInstance_fipsEnabled,
			"disappears":          testAccVerifiedAccessInstance_disappears,
			names.AttrTags:        testAccVerifiedAccessInstance_tags,
		},
		"InstanceLoggingConfiguration": {
			"accessLogsIncludeTrustContext":                 testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsIncludeTrustContext,
			"accessLogsLogVersion":                          testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsLogVersion,
			"accessLogsCloudWatchLogs":                      testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsCloudWatchLogs,
			"accessLogsKinesisDataFirehose":                 testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsKinesisDataFirehose,
			"accessLogsS3":                                  testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsS3,
			"accessLogsCloudWatchLogsKinesisDataFirehoseS3": testAccVerifiedAccessInstanceLoggingConfiguration_accessLogsCloudWatchLogsKinesisDataFirehoseS3,
			"disappears":                                    testAccVerifiedAccessInstanceLoggingConfiguration_disappears,
		},
		"InstanceTrustProviderAttachment": {
			"basic":      testAccVerifiedAccessInstanceTrustProviderAttachment_basic,
			"disappears": testAccVerifiedAccessInstanceTrustProviderAttachment_disappears,
		},
	}

	acctest.RunLimitedConcurrencyTests2Levels(t, semaphore, testCases)
}

func testAccPreCheckVerifiedAccessSynchronize(t *testing.T, semaphore tfsync.Semaphore) {
	tfsync.TestAccPreCheckSyncronize(t, semaphore, "Verified Access")
}
