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
			acctest.CtBasic:      testAccDRTAccessLogBucketAssociation_basic,
			"multibucket":        testAccDRTAccessLogBucketAssociation_multiBucket,
			acctest.CtDisappears: testAccDRTAccessLogBucketAssociation_disappears,
		},
		"DRTAccessRoleARNAssociation": {
			acctest.CtBasic:      testAccDRTAccessRoleARNAssociation_basic,
			acctest.CtDisappears: testAccDRTAccessRoleARNAssociation_disappears,
		},
		"ProactiveEngagement": {
			acctest.CtBasic:      testAccProactiveEngagement_basic,
			"disabled":           testAccProactiveEngagement_disabled,
			acctest.CtDisappears: testAccProactiveEngagement_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
