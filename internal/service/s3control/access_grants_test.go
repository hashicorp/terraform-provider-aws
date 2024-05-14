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
			"basic":          testAccAccessGrantsInstance_basic,
			"disappears":     testAccAccessGrantsInstance_disappears,
			names.AttrTags:   testAccAccessGrantsInstance_tags,
			"identityCenter": testAccAccessGrantsInstance_identityCenter,
		},
		"Location": {
			"basic":        testAccAccessGrantsLocation_basic,
			"disappears":   testAccAccessGrantsLocation_disappears,
			names.AttrTags: testAccAccessGrantsLocation_tags,
			"update":       testAccAccessGrantsLocation_update,
		},
		"Grant": {
			"basic":                 testAccAccessGrant_basic,
			"disappears":            testAccAccessGrant_disappears,
			names.AttrTags:          testAccAccessGrant_tags,
			"locationConfiguration": testAccAccessGrant_locationConfiguration,
		},
		"InstanceResourcePolicy": {
			"basic":      testAccAccessGrantsInstanceResourcePolicy_basic,
			"disappears": testAccAccessGrantsInstanceResourcePolicy_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
