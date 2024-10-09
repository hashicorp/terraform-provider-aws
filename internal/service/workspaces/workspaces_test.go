// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccWorkSpaces_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Directory": {
			acctest.CtBasic:               testAccDirectory_basic,
			acctest.CtDisappears:          testAccDirectory_disappears,
			"ipGroupIds":                  testAccDirectory_ipGroupIDs,
			"selfServicePermissions":      testAccDirectory_selfServicePermissions,
			"subnetIDs":                   testAccDirectory_subnetIDs,
			"tags":                        testAccDirectory_tags,
			"workspaceAccessProperties":   testAccDirectory_workspaceAccessProperties,
			"workspaceCreationProperties": testAccDirectory_workspaceCreationProperties,
			"workspaceCreationProperties_customSecurityGroupId_defaultOu": testAccDirectory_workspaceCreationProperties_customSecurityGroupId_defaultOu,
		},
		"IpGroup": {
			acctest.CtBasic:       testAccIPGroup_basic,
			acctest.CtDisappears:  testAccIPGroup_disappears,
			"multipleDirectories": testAccIPGroup_MultipleDirectories,
			"tags":                testAccIPGroup_tags,
		},
		"Workspace": {
			acctest.CtBasic:          testAccWorkspace_basic,
			"recreate":               testAccWorkspace_recreate,
			"tags":                   testAccWorkspace_tags,
			"timeout":                testAccWorkspace_timeout,
			"validateRootVolumeSize": testAccWorkspace_validateRootVolumeSize,
			"validateUserVolumeSize": testAccWorkspace_validateUserVolumeSize,
			"workspaceProperties":    testAccWorkspace_workspaceProperties,
			"workspaceProperties_runningModeAlwaysOn": testAccWorkspace_workspaceProperties_runningModeAlwaysOn,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
