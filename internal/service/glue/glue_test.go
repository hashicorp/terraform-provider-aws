package glue_test

import "testing"

func TestAccGlue_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"DataCatalogEncryptionSettings": {
			"basic":      testAccDataCatalogEncryptionSettings_basic,
			"dataSource": testAccDataCatalogEncryptionSettingsDataSource_basic,
		},
		"ResourcePolicy": {
			"basic":      testAccResourcePolicy_basic,
			"update":     testAccResourcePolicy_update,
			"hybrid":     testAccResourcePolicy_hybrid,
			"disappears": testAccResourcePolicy_disappears,
			"equivalent": testAccResourcePolicy_ignoreEquivalent,
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
