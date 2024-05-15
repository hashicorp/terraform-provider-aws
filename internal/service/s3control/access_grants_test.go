// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlAccessGrants_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Instance": {
			acctest.CtBasic:  testAccAccessGrantsInstance_basic,
			"disappears":     testAccAccessGrantsInstance_disappears,
			names.AttrTags:   testAccAccessGrantsInstance_tags,
			"identityCenter": testAccAccessGrantsInstance_identityCenter,
		},
		"Location": {
			acctest.CtBasic: testAccAccessGrantsLocation_basic,
			"disappears":    testAccAccessGrantsLocation_disappears,
			names.AttrTags:  testAccAccessGrantsLocation_tags,
			"update":        testAccAccessGrantsLocation_update,
		},
		"Grant": {
			acctest.CtBasic:         testAccAccessGrant_basic,
			"disappears":            testAccAccessGrant_disappears,
			names.AttrTags:          testAccAccessGrant_tags,
			"locationConfiguration": testAccAccessGrant_locationConfiguration,
		},
		"InstanceResourcePolicy": {
			acctest.CtBasic: testAccAccessGrantsInstanceResourcePolicy_basic,
			"disappears":    testAccAccessGrantsInstanceResourcePolicy_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
