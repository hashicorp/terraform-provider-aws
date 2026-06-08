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
		// Catalog tests are serialized due to aws_lakeformation_data_lake_settings, a per-account
		// singleton that is needed for permissions. Ideally, we would be able to do a setup for
		// all the tests and a clean up after, so they could run in parallel.
		"Catalog": {
			acctest.CtBasic:                     testAccCatalog_basic,
			acctest.CtDisappears:                testAccCatalog_disappears,
			"catalogPropertiesDataLakeAccess":   testAccCatalog_catalogPropertiesDataLakeAccess,
			"configurationError":                testAccCatalog_configurationError,
			"federatedCatalog_mySQL":            testAccCatalog_FederatedCatalog_mySQL,
			"federatedCatalog_s3Tables":         testAccCatalog_FederatedCatalog_s3Tables,
			"tags":                              testAccCatalog_tags,
			"targetRedshiftCatalog_serverless":  testAccCatalog_TargetRedshiftCatalog_serverless,
			"targetRedshiftCatalog_provisioned": testAccCatalog_TargetRedshiftCatalog_provisioned,
			"identity":                          testAccGlueCatalog_identitySerial,
		},
		// Catalog data source tests are serialized due to aws_lakeformation_data_lake_settings, a per-account
		// singleton that is needed for permissions. Ideally, we would be able to do a setup for
		// all the tests and a clean up after, so they could run in parallel.
		"CatalogDataSource": {
			"catalogPropertiesDataLakeAccess":  testAccCatalogDataSource_catalogPropertiesDataLakeAccess,
			"federatedCatalog_mySQL":           testAccCatalogDataSource_FederatedCatalog_mySQL,
			"federatedCatalog_s3Tables":        testAccCatalogDataSource_FederatedCatalog_s3Tables,
			"targetRedshiftCatalog_serverless": testAccCatalogDataSource_TargetRedshiftCatalog_serverless,
		},
		// Catalog list tests are serialized due to aws_lakeformation_data_lake_settings, a per-account
		// singleton that is needed for permissions. Ideally, we would be able to do a setup for
		// all the tests and a clean up after, so they could run in parallel.
		"CatalogList": {
			acctest.CtBasic:   testAccCatalog_List_basic,
			"regionOverride":  testAccCatalog_List_regionOverride,
			"includeResource": testAccCatalog_List_includeResource,
		},
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
