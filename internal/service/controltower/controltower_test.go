// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package controltower_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccControlTower_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"LandingZone": {
			acctest.CtBasic:      testAccLandingZone_basic,
			acctest.CtDisappears: testAccLandingZone_disappears,
			"tags":               testAccLandingZone_tags,
		},
		"Control": {
			acctest.CtBasic:      testAccControl_basic,
			acctest.CtDisappears: testAccControl_disappears,
			"parameters":         testAccControl_parameters,
		},
		"Baseline": {
			acctest.CtBasic:      testAccBaseline_basic,
			acctest.CtDisappears: testAccBaseline_disappears,
			"tags":               testAccBaseline_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
