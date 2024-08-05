// Copyright (c) HashiCorp, Inc.
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
		},
		"DataCellsFilter": {
			acctest.CtBasic:      testAccDataCellsFilter_basic,
			"columnWildcard":     testAccDataCellsFilter_columnWildcard,
			acctest.CtDisappears: testAccDataCellsFilter_disappears,
			"rowFilter":          testAccDataCellsFilter_rowFilter,
		},
		"DataLakeSettingsDataSource": {
			acctest.CtBasic:  testAccDataLakeSettingsDataSource_basic,
			"readOnlyAdmins": testAccDataLakeSettingsDataSource_readOnlyAdmins,
		},
		"PermissionsBasic": {
			acctest.CtBasic:       testAccPermissions_basic,
			"database":            testAccPermissions_database,
			"databaseIAMAllowed":  testAccPermissions_databaseIAMAllowed,
			"databaseMultiple":    testAccPermissions_databaseMultiple,
			"dataCellsFilter":     testAccPermissions_dataCellsFilter,
			"dataLocation":        testAccPermissions_dataLocation,
			acctest.CtDisappears:  testAccPermissions_disappears,
			"lfTag":               testAccPermissions_lfTag,
			"lfTagPolicy":         testAccPermissions_lfTagPolicy,
			"lfTagPolicyMultiple": testAccPermissions_lfTagPolicyMultiple,
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
		},
		"PermissionsTable": {
			acctest.CtBasic:      testAccPermissions_tableBasic,
			"iamAllowed":         testAccPermissions_tableIAMAllowed,
			"implicit":           testAccPermissions_tableImplicit,
			"multipleRoles":      testAccPermissions_tableMultipleRoles,
			"selectOnly":         testAccPermissions_tableSelectOnly,
			"selectPlus":         testAccPermissions_tableSelectPlus,
			"wildcardNoSelect":   testAccPermissions_tableWildcardNoSelect,
			"wildcardSelectOnly": testAccPermissions_tableWildcardSelectOnly,
			"wildcardSelectPlus": testAccPermissions_tableWildcardSelectPlus,
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
