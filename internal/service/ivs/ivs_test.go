// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIVS_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"PlaybackKeyPair": {
			"basic":        testAccPlaybackKeyPair_basic,
			"update":       testAccPlaybackKeyPair_update,
			names.AttrTags: testAccPlaybackKeyPair_tags,
			"disappears":   testAccPlaybackKeyPair_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
