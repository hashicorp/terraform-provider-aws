// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccGlue_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"CatalogTableOptimizer": {
			acctest.CtBasic:                                   testAccCatalogTableOptimizer_basic,
			"deleteOrphanFileConfiguration":                   testAccCatalogTableOptimizer_DeleteOrphanFileConfiguration,
			"deleteOrphanFileConfigurationWithRunRateInHours": testAccCatalogTableOptimizer_DeleteOrphanFileConfigurationWithRunRateInHours,
			acctest.CtDisappears:                              testAccCatalogTableOptimizer_disappears,
			"retentionConfiguration":                          testAccCatalogTableOptimizer_RetentionConfiguration,
			"retentionConfigurationWithRunRateInHours":        testAccCatalogTableOptimizer_RetentionConfigurationWithRunRateInHours,
			"update": testAccCatalogTableOptimizer_update,
		},
		"DataCatalogEncryptionSettings": {
			acctest.CtBasic: testAccDataCatalogEncryptionSettings_basic,
			"dataSource":    testAccDataCatalogEncryptionSettingsDataSource_basic,
		},
		"ResourcePolicy": {
			acctest.CtBasic:      testAccResourcePolicy_basic,
			"update":             testAccResourcePolicy_update,
			"hybrid":             testAccResourcePolicy_hybrid,
			acctest.CtDisappears: testAccResourcePolicy_disappears,
			"equivalent":         testAccResourcePolicy_ignoreEquivalent,
			"Identity":           testAccGlueResourcePolicy_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
