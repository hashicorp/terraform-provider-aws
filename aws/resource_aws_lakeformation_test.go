package aws

import (
	"testing"
)

func TestAccAWSLakeFormation_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"DataLakeSettings": {
			"basic":            testAccAWSLakeFormationDataLakeSettings_basic,
			"disappears":       testAccAWSLakeFormationDataLakeSettings_disappears,
			"withoutCatalogId": testAccAWSLakeFormationDataLakeSettings_withoutCatalogId,
			"dataSource":       testAccAWSLakeFormationDataLakeSettingsDataSource_basic,
		},
		"BasicPermissions": {
			"basic":        testAccAWSLakeFormationPermissions_basic,
			"dataLocation": testAccAWSLakeFormationPermissions_dataLocation,
			"database":     testAccAWSLakeFormationPermissions_database,
			"LFTag":        testAccAWSLakeFormationPermissions_lf_tag,
			"LFTagPolicy":  testAccAWSLakeFormationPermissions_lf_tag_policy,
		},
		"TablePermissions": {
			"columnWildcardPermissions":           testAccAWSLakeFormationPermissions_columnWildcardPermissions,
			"implicitTableWithColumnsPermissions": testAccAWSLakeFormationPermissions_implicitTableWithColumnsPermissions,
			"implicitTablePermissions":            testAccAWSLakeFormationPermissions_implicitTablePermissions,
			"selectPermissions":                   testAccAWSLakeFormationPermissions_selectPermissions,
			"tableName":                           testAccAWSLakeFormationPermissions_tableName,
			"tableWildcard":                       testAccAWSLakeFormationPermissions_tableWildcard,
			"tableWildcardPermissions":            testAccAWSLakeFormationPermissions_tableWildcardPermissions,
			"tableWithColumns":                    testAccAWSLakeFormationPermissions_tableWithColumns,
		},
		"DataSourcePermissions": {
			"basicDataSource":            testAccAWSLakeFormationPermissionsDataSource_basic,
			"dataLocationDataSource":     testAccAWSLakeFormationPermissionsDataSource_dataLocation,
			"databaseDataSource":         testAccAWSLakeFormationPermissionsDataSource_database,
			"LFTagDataSource":            testAccAWSLakeFormationPermissionsDataSource_lf_tag,
			"LFTagPolicyDataSource":      testAccAWSLakeFormationPermissionsDataSource_lf_tag_policy,
			"tableDataSource":            testAccAWSLakeFormationPermissionsDataSource_table,
			"tableWithColumnsDataSource": testAccAWSLakeFormationPermissionsDataSource_tableWithColumns,
		},
		"LFTags": {
			"basic":      testAccAWSLakeFormationLFTag_basic,
			"disappears": testAccAWSLakeFormationLFTag_disappears,
			"values":     testAccAWSLakeFormationLFTag_values,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
