package workspaces_test

import (
	"testing"
)

func TestAccWorkSpacesDataSource_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Bundle": {
			"basic":                   testAccWorkspaceBundleDataSource_basic,
			"bundleIDAndNameConflict": testAccWorkspaceBundleDataSource_bundleIDAndNameConflict,
			"byOwnerName":             testAccWorkspaceBundleDataSource_byOwnerName,
			"privateOwner":            testAccWorkspaceBundleDataSource_privateOwner,
		},
		"Directory": {
			"basic": testAccDirectoryDataSource_basic,
		},
		"Image": {
			"basic": testAccImageDataSource_basic,
		},
		"Workspace": {
			"byWorkspaceID":                     testAccWorkspaceDataSource_byWorkspaceID,
			"byDirectoryID_userName":            testAccWorkspaceDataSource_byDirectoryID_userName,
			"workspaceIDAndDirectoryIDConflict": testAccWorkspaceDataSource_workspaceIDAndDirectoryIDConflict,
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
