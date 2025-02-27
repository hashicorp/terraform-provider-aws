// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataZone_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Environment": {
			acctest.CtBasic:      testAccEnvironment_basic,
			acctest.CtDisappears: testAccEnvironment_disappears,
			"update":             testAccEnvironment_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
