// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBedrockAgent_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"KnowledgeBase": {
			"basicRDS":         testAccKnowledgeBase_basicRDS,
			"disappears":       testAccKnowledgeBase_disappears,
			"tags":             testAccKnowledgeBase_tags,
			"basicOpenSearch":  testAccKnowledgeBase_basicOpenSearch,
			"updateOpenSearch": testAccKnowledgeBase_updateOpenSearch,
		},
		"DataSource": {
			"basic":      testAccDataSource_basic,
			"disappears": testAccDataSource_disappears,
			"full":       testAccDataSource_full,
			"update":     testAccDataSource_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
