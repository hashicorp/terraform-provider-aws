package aws

import (
	"testing"
)

func TestAccDataSourceAwsWorkspaces_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Bundle": {
			"basic":                   testAccDataSourceAwsWorkspaceBundle_basic,
			"bundleIDAndNameConflict": testAccDataSourceAwsWorkspaceBundle_bundleIDAndNameConflict,
			"byOwnerName":             testAccDataSourceAwsWorkspaceBundle_byOwnerName,
			"privateOwner":            testAccDataSourceAwsWorkspaceBundle_privateOwner,
		},
		"Directory": {
			"basic": testAccDataSourceAwsWorkspacesDirectory_basic,
		},
		"Image":     {},
		"Workspace": {},
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
