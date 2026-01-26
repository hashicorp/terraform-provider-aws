// Copyright IBM Corp. 2014, 2026
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
			acctest.CtDisappears:                     testAccKnowledgeBase_disappears,
			"tags":                                   testAccKnowledgeBase_tags,
			"update":                                 testAccKnowledgeBase_update,
			"OpenSearchServerlessBasic":              testAccKnowledgeBase_OpenSearchServerless_basic,
			"Kendra":                                 testAccKnowledgeBase_Kendra_basic,
			"NeptuneAnalytics":                       testAccKnowledgeBase_NeptuneAnalytics_basic,
			"OpenSearchManagedClusterBasic":          testAccKnowledgeBase_OpenSearchManagedCluster_basic,
			"S3Vectors":                              testAccKnowledgeBase_S3Vectors_update,
			"StructuredDataStoreRedshiftProvisioned": testAccKnowledgeBase_StructuredDataStore_redshiftProvisioned,
			"StructuredDataStoreRedshiftServerless":  testAccKnowledgeBase_StructuredDataStore_redshiftServerless,
			"RDS":                                    testAccKnowledgeBase_RDS_basic,
			"RDSSupplementalDataStorage":             testAccKnowledgeBase_RDS_supplementalDataStorage,
		},
		"DataSource": {
			acctest.CtBasic:         testAccDataSource_basic,
			acctest.CtDisappears:    testAccDataSource_disappears,
			"full":                  testAccDataSource_full,
			"update":                testAccDataSource_update,
			"semantic":              testAccDataSource_fullSemantic,
			"hierarchical":          testAccDataSource_fullHierarchical,
			"parsing":               testAccDataSource_parsing,
			"parsingModality":       testAccDataSource_parsingModality,
			"bedrockDataAutomation": testAccDataSource_bedrockDataAutomation,
			"customTransformation":  testAccDataSource_fullCustomTranformation,
			"webConfiguration":      testAccDataSource_webConfiguration,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
