// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3ControlAccessGrants_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Instance": {
			acctest.CtBasic:      testAccAccessGrantsInstance_basic,
			acctest.CtDisappears: testAccAccessGrantsInstance_disappears,
			"tags":               testAccAccessGrantsInstance_tags,
			"identityCenter":     testAccAccessGrantsInstance_identityCenter,
		},
		"Location": {
			acctest.CtBasic:      testAccAccessGrantsLocation_basic,
			acctest.CtDisappears: testAccAccessGrantsLocation_disappears,
			"tags":               testAccAccessGrantsLocation_tags,
			"update":             testAccAccessGrantsLocation_update,
		},
		"Grant": {
			acctest.CtBasic:         testAccAccessGrant_basic,
			acctest.CtDisappears:    testAccAccessGrant_disappears,
			"tags":                  testAccAccessGrant_tags,
			"locationConfiguration": testAccAccessGrant_locationConfiguration,
		},
		"InstanceResourcePolicy": {
			acctest.CtBasic:      testAccAccessGrantsInstanceResourcePolicy_basic,
			acctest.CtDisappears: testAccAccessGrantsInstanceResourcePolicy_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
