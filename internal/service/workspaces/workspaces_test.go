package workspaces_test

import (
	"testing"
)

func TestAccWorkSpaces_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Directory": {
			"basic":                       testAccDirectory_basic,
			"disappears":                  testAccDirectory_disappears,
			"ipGroupIds":                  testAccDirectory_ipGroupIDs,
			"selfServicePermissions":      testAccDirectory_selfServicePermissions,
			"subnetIDs":                   testAccDirectory_subnetIDs,
			"tags":                        testAccDirectory_tags,
			"workspaceAccessProperties":   testAccDirectory_workspaceAccessProperties,
			"workspaceCreationProperties": testAccDirectory_workspaceCreationProperties,
			"workspaceCreationProperties_customSecurityGroupId_defaultOu": testAccDirectory_workspaceCreationProperties_customSecurityGroupId_defaultOu,
		},
		"IpGroup": {
			"basic":               testAccIPGroup_basic,
			"disappears":          testAccIPGroup_disappears,
			"multipleDirectories": testAccIPGroup_MultipleDirectories,
			"tags":                testAccIPGroup_tags,
		},
		"Workspace": {
			"basic":                  testAccWorkspace_basic,
			"recreate":               testAccWorkspace_recreate,
			"tags":                   testAccWorkspace_tags,
			"timeout":                testAccWorkspace_timeout,
			"validateRootVolumeSize": testAccWorkspace_validateRootVolumeSize,
			"validateUserVolumeSize": testAccWorkspace_validateUserVolumeSize,
			"workspaceProperties":    testAccWorkspace_workspaceProperties,
			"workspaceProperties_runningModeAlwaysOn": testAccWorkspace_workspaceProperties_runningModeAlwaysOn,
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
