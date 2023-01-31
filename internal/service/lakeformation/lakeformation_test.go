package lakeformation_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLakeFormation_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DataLakeSettings": {
			"basic":            testAccDataLakeSettings_basic,
			"dataSource":       testAccDataLakeSettingsDataSource_basic,
			"disappears":       testAccDataLakeSettings_disappears,
			"withoutCatalogId": testAccDataLakeSettings_withoutCatalogID,
		},
		"PermissionsBasic": {
			"basic":              testAccPermissions_basic,
			"database":           testAccPermissions_database,
			"databaseIAMAllowed": testAccPermissions_databaseIAMAllowed,
			"databaseMultiple":   testAccPermissions_databaseMultiple,
			"dataLocation":       testAccPermissions_dataLocation,
			"disappears":         testAccPermissions_disappears,
			"lfTag":              testAccPermissions_lfTag,
			"lfTagPolicy":        testAccPermissions_lfTagPolicy,
		},
		"PermissionsDataSource": {
			"basic":            testAccPermissionsDataSource_basic,
			"database":         testAccPermissionsDataSource_database,
			"dataLocation":     testAccPermissionsDataSource_dataLocation,
			"lfTag":            testAccPermissionsDataSource_lfTag,
			"lfTagPolicy":      testAccPermissionsDataSource_lfTagPolicy,
			"table":            testAccPermissionsDataSource_table,
			"tableWithColumns": testAccPermissionsDataSource_tableWithColumns,
		},
		"PermissionsTable": {
			"basic":              testAccPermissions_tableBasic,
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
			"basic":                   testAccPermissions_twcBasic,
			"implicit":                testAccPermissions_twcImplicit,
			"wildcardExcludedColumns": testAccPermissions_twcWildcardExcludedColumns,
			"wildcardSelectOnly":      testAccPermissions_twcWildcardSelectOnly,
			"wildcardSelectPlus":      testAccPermissions_twcWildcardSelectPlus,
		},
		"LFTags": {
			"basic":      testAccLFTag_basic,
			"disappears": testAccLFTag_disappears,
			"values":     testAccLFTag_values,
		},
		"ResourceLFTags": {
			"basic":            testAccResourceLFTags_basic,
			"database":         testAccResourceLFTags_database,
			"databaseMultiple": testAccResourceLFTags_databaseMultiple,
			"table":            testAccResourceLFTags_table,
			"tableWithColumns": testAccResourceLFTags_tableWithColumns,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
