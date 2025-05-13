// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccResourceExplorer2_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Index": {
			acctest.CtBasic:           testAccIndex_basic,
			acctest.CtDisappears:      testAccIndex_disappears,
			"tags":                    testAccIndex_tags,
			"type":                    testAccIndex_type,
			"Identity_Basic":          testAccIndex_Identity_Basic,
			"Identity_RegionOverride": testAccIndex_Identity_RegionOverride,
		},
		"View": {
			acctest.CtBasic:           testAccView_basic,
			"defaultView":             testAccView_defaultView,
			acctest.CtDisappears:      testAccView_disappears,
			"filter":                  testAccView_filter,
			"scope":                   testAccView_scope,
			"tags":                    testAccView_tags,
			"Identity_Basic":          testAccView_Identity_Basic,
			"Identity_RegionOverride": testAccView_Identity_RegionOverride,
		},
		"SearchDataSource": {
			acctest.CtBasic: testAccSearchDataSource_basic,
			"indexType":     testAccSearchDataSource_IndexType,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
