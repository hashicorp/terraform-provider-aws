// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNetworkFlowMonitor_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Monitor": {
			acctest.CtBasic:      testAccMonitor_basic,
			acctest.CtDisappears: testAccMonitor_disappears,
			"tags":               testAccMonitor_tags,
			"update":             testAccMonitor_update,
		},
		"Scope": {
			acctest.CtBasic:      testAccScope_basic,
			acctest.CtDisappears: testAccScope_disappears,
			"tags":               testAccScope_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
