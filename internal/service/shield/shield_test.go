// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccShield_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DRTAccessLogBucketAssociation": {
			"basic":       testAccDRTAccessLogBucketAssociation_basic,
			"multibucket": testAccDRTAccessLogBucketAssociation_multiBucket,
			"disappears":  testAccDRTAccessLogBucketAssociation_disappears,
		},
		"DRTAccessRoleARNAssociation": {
			"basic":      testAccDRTAccessRoleARNAssociation_basic,
			"disappears": testAccDRTAccessRoleARNAssociation_disappears,
		},
		"ProactiveEngagement": {
			"basic":      testAccProactiveEngagement_basic,
			"disabled":   testAccProactiveEngagement_disabled,
			"disappears": testAccProactiveEngagement_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
