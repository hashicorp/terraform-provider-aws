// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLakeFormation_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DataLakeSettings": {
			acctest.CtBasic:      testAccDataLakeSettings_basic,
			acctest.CtDisappears: testAccDataLakeSettings_disappears,
			"withoutCatalogId":   testAccDataLakeSettings_withoutCatalogID,
			"readOnlyAdmins":     testAccDataLakeSettings_readOnlyAdmins,
			"parameters":         testAccDataLakeSettings_parameters,
		},
		"DataCellsFilter": {
			acctest.CtBasic:          testAccDataCellsFilter_basic,
			"columnWildcard":         testAccDataCellsFilter_columnWildcard,
			"columnWildcardMultiple": testAccDataCellsFilter_columnWildcardMultiple,
			acctest.CtDisappears:     testAccDataCellsFilter_disappears,
			"rowFilter":              testAccDataCellsFilter_rowFilter,
		},
		"DataLakeSettingsDataSource": {
			acctest.CtBasic:  testAccDataLakeSettingsDataSource_basic,
			"readOnlyAdmins": testAccDataLakeSettingsDataSource_readOnlyAdmins,
		},
		"IdentityCenterConfiguration": {
			acctest.CtBasic:      testAccLakeFormationIdentityCenterConfiguration_basic,
			acctest.CtDisappears: testAccLakeFormationIdentityCenterConfiguration_disappears,
			"Identity":           testAccLakeFormationIdentityCenterConfiguration_identitySerial,
		},
		"OptIn": {
			acctest.CtBasic:      testAccOptIn_basic,
			acctest.CtDisappears: testAccOptIn_disappears,
			"table":              testAccOptIn_table,
		},
		"PermissionsBasic": {
			acctest.CtBasic:         testAccPermissions_basic,
			"database":              testAccPermissions_database,
			"databaseIAMAllowed":    testAccPermissions_databaseIAMAllowed,
			"databaseIAMPrincipals": testAccPermissions_databaseIAMPrincipals,
			"databaseMultiple":      testAccPermissions_databaseMultiple,
			"dataCellsFilter":       testAccPermissions_dataCellsFilter,
			"dataLocation":          testAccPermissions_dataLocation,
			acctest.CtDisappears:    testAccPermissions_disappears,
			"lfTag":                 testAccPermissions_lfTag,
			"lfTagPolicy":           testAccPermissions_lfTagPolicy,
			"lfTagPolicyMultiple":   testAccPermissions_lfTagPolicyMultiple,
			"nonIAMPrincipals":      testAccPermissions_catalogResource_nonIAMPrincipals,
		},
		"PermissionsDataSource": {
			acctest.CtBasic:    testAccPermissionsDataSource_basic,
			"dataCellsFilter":  testAccPermissionsDataSource_dataCellsFilter,
			"database":         testAccPermissionsDataSource_database,
			"dataLocation":     testAccPermissionsDataSource_dataLocation,
			"lfTag":            testAccPermissionsDataSource_lfTag,
			"lfTagPolicy":      testAccPermissionsDataSource_lfTagPolicy,
			"table":            testAccPermissionsDataSource_table,
			"tableWithColumns": testAccPermissionsDataSource_tableWithColumns,
			"nonIAMPrincipals": testAccPermissionsDataSource_catalogResource_nonIAMPrincipals,
		},
		"PermissionsTable": {
			acctest.CtBasic:      testAccPermissions_tableBasic,
			"iamAllowed":         testAccPermissions_tableIAMAllowed,
			"iamPrincipals":      testAccPermissions_tableIAMPrincipals,
			"implicit":           testAccPermissions_tableImplicit,
			"multipleRoles":      testAccPermissions_tableMultipleRoles,
			"selectOnly":         testAccPermissions_tableSelectOnly,
			"selectPlus":         testAccPermissions_tableSelectPlus,
			"wildcardNoSelect":   testAccPermissions_tableWildcardNoSelect,
			"wildcardSelectOnly": testAccPermissions_tableWildcardSelectOnly,
			"wildcardSelectPlus": testAccPermissions_tableWildcardSelectPlus,
			"nonIAMPrincipals":   testAccPermissions_table_nonIAMPrincipals,
		},
		"PermissionsTableWithColumns": {
			acctest.CtBasic:           testAccPermissions_twcBasic,
			"implicit":                testAccPermissions_twcImplicit,
			"wildcardExcludedColumns": testAccPermissions_twcWildcardExcludedColumns,
			"wildcardSelectOnly":      testAccPermissions_twcWildcardSelectOnly,
			"wildcardSelectPlus":      testAccPermissions_twcWildcardSelectPlus,
		},
		"LFTags": {
			acctest.CtBasic:      testAccLFTag_basic,
			acctest.CtDisappears: testAccLFTag_disappears,
			"tagKeyComplex":      testAccLFTag_TagKey_complex,
			"values":             testAccLFTag_Values,
			"valuesOverFifty":    testAccLFTag_Values_overFifty,
		},
		"LFTagExpression": {
			acctest.CtBasic:      testAccLFTagExpression_basic,
			acctest.CtDisappears: testAccLFTagExpression_disappears,
			"update":             testAccLFTagExpression_update,
		},
		"ResourceLFTag": {
			acctest.CtBasic:      testAccResourceLFTag_basic,
			acctest.CtDisappears: testAccResourceLFTag_disappears,
			"table":              testAccResourceLFTag_table,
			"tableWithColumns":   testAccResourceLFTag_tableWithColumns,
		},
		"ResourceLFTags": {
			acctest.CtBasic:        testAccResourceLFTags_basic,
			"database":             testAccResourceLFTags_database,
			"databaseMultipleTags": testAccResourceLFTags_databaseMultipleTags,
			acctest.CtDisappears:   testAccResourceLFTags_disappears,
			"hierarchy":            testAccResourceLFTags_hierarchy,
			"table":                testAccResourceLFTags_table,
			"tableWithColumns":     testAccResourceLFTags_tableWithColumns,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
