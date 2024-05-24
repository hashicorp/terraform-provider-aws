// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccResourceExplorer2_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Index": {
			acctest.CtBasic:      testAccIndex_basic,
			acctest.CtDisappears: testAccIndex_disappears,
			"tags":               testAccIndex_tags,
			"type":               testAccIndex_type,
		},
		"View": {
			acctest.CtBasic:      testAccView_basic,
			"defaultView":        testAccView_defaultView,
			acctest.CtDisappears: testAccView_disappears,
			"filter":             testAccView_filter,
			"tags":               testAccView_tags,
		},
		"SearchDataSource": {
			acctest.CtBasic: testAccSearchDataSource_basic,
			"indexType":     testAccSearchDataSource_IndexType,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
