package aws

import (
	"testing"
)

func TestAccAWSWorkspaces_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Directory": {
			"basic":                       testAccAwsWorkspacesDirectory_basic,
			"disappears":                  testAccAwsWorkspacesDirectory_disappears,
			"ipGroupIds":                  testAccAwsWorkspacesDirectory_ipGroupIds,
			"selfServicePermissions":      testAccAwsWorkspacesDirectory_selfServicePermissions,
			"subnetIDs":                   testAccAwsWorkspacesDirectory_subnetIds,
			"tags":                        testAccAwsWorkspacesDirectory_tags,
			"workspaceAccessProperties":   testAccAwsWorkspacesDirectory_workspaceAccessProperties,
			"workspaceCreationProperties": testAccAwsWorkspacesDirectory_workspaceCreationProperties,
			"workspaceCreationProperties_customSecurityGroupId_defaultOu": testAccAwsWorkspacesDirectory_workspaceCreationProperties_customSecurityGroupId_defaultOu,
		},
		"IpGroup":   {},
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
