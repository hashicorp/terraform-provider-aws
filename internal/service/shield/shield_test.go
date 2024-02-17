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
			"basic":       testDRTAccessLogBucketAssociation_basic,
			"multibucket": testDRTAccessLogBucketAssociation_multibucket,
			"disappears":  testDRTAccessLogBucketAssociation_disappears,
		},
		"DRTAccessRoleARNAssociation": {
			"basic":      testDRTAccessRoleARNAssociation_basic,
			"disappears": testDRTAccessRoleARNAssociation_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
