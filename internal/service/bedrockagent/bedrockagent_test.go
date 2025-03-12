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
			acctest.CtBasic:                     testAccKnowledgeBase_basic,
			acctest.CtDisappears:                testAccKnowledgeBase_disappears,
			"tags":                              testAccKnowledgeBase_tags,
			"OpenSearchBasic":                   testAccKnowledgeBase_OpenSearch_basic,
			"OpenSearchUpdate":                  testAccKnowledgeBase_OpenSearch_update,
			"OpenSearchSupplementalDataStorage": testAccKnowledgeBase_OpenSearch_supplementalDataStorage,
		},
		"DataSource": {
			acctest.CtBasic:        testAccDataSource_basic,
			acctest.CtDisappears:   testAccDataSource_disappears,
			"full":                 testAccDataSource_full,
			"update":               testAccDataSource_update,
			"semantic":             testAccDataSource_fullSemantic,
			"hierarchical":         testAccDataSource_fullHierarchical,
			"parsing":              testAccDataSource_parsing,
			"customtransformation": testAccDataSource_fullCustomTranformation,
			"webconfiguration":     testAccDataSource_webConfiguration,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
