// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMediaLive_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Multiplex": {
			acctest.CtBasic:      testAccMultiplex_basic,
			acctest.CtDisappears: testAccMultiplex_disappears,
			"update":             testAccMultiplex_update,
			"updateTags":         testAccMultiplex_updateTags,
			"start":              testAccMultiplex_start,
		},
		"MultiplexProgram": {
			acctest.CtBasic:      testAccMultiplexProgram_basic,
			"update":             testAccMultiplexProgram_update,
			acctest.CtDisappears: testAccMultiplexProgram_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
