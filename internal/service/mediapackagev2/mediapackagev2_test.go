// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackagev2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// The Default AWS Quota for how many MediaPackage V2 Channel Groups you can have is 3
// We'll serialize the tests to prevent hitting that quota
func TestAccMediaPackageV2_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"ChannelGroup": {
			acctest.CtBasic:      testAccMediaPackageV2ChannelGroup_basic,
			"description":        testAccMediaPackageV2ChannelGroup_description,
			acctest.CtDisappears: testAccMediaPackageV2ChannelGroup_disappears,
			"tags":               testAccMediaPackageV2ChannelGroup_tagsSerial,
		},
		"Channel": {
			acctest.CtBasic:      testAccMediaPackageV2Channel_basic,
			"cmaf":               testAccMediaPackageV2Channel_cmaf,
			"description":        testAccMediaPackageV2Channel_description,
			acctest.CtDisappears: testAccMediaPackageV2Channel_disappears,
			"tags":               testAccMediaPackageV2Channel_tagsSerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
