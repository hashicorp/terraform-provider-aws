package aws

import (
	"testing"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
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
		"IpGroup": {
			"basic":               testAccAwsWorkspacesIpGroup_basic,
			"disappears":          testAccAwsWorkspacesIpGroup_disappears,
			"multipleDirectories": testAccAwsWorkspacesIpGroup_MultipleDirectories,
			"tags":                testAccAwsWorkspacesIpGroup_tags,
		},
		"Workspace": {
			"basic":                  testAccAwsWorkspacesWorkspace_basic,
			"recreate":               testAccAwsWorkspacesWorkspace_recreate,
			"tags":                   testAccAwsWorkspacesWorkspace_tags,
			"timeout":                testAccAwsWorkspacesWorkspace_timeout,
			"validateRootVolumeSize": testAccAwsWorkspacesWorkspace_validateRootVolumeSize,
			"validateUserVolumeSize": testAccAwsWorkspacesWorkspace_validateUserVolumeSize,
			"workspaceProperties":    testAccAwsWorkspacesWorkspace_workspaceProperties,
			"workspaceProperties_runningModeAlwaysOn": testAccAwsWorkspacesWorkspace_workspaceProperties_runningModeAlwaysOn,
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
