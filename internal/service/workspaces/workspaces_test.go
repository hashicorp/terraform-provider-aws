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
			"tenancy":                     TestAccWorkspacesDirectory_dedicatedTenancy,
			"workspaceAccessProperties":   testAccDirectory_workspaceAccessProperties,
			"workspaceCreationProperties": testAccDirectory_workspaceCreationProperties,
			"workspaceCreationProperties_customSecurityGroupId_defaultOu": testAccDirectory_workspaceCreationProperties_customSecurityGroupId_defaultOu,
			"workspaceCertificateBasedAuthProperties":                     testAccDirectory_CertificateBasedAuthProperties,
			"workspaceSamlProperties":                                     testAccDirectory_SamlProperties,
			"workspacePoolsBasic":                                         testAccDirectory_poolsBasic,
			"workspacePoolsADConfig":                                      testAccDirectory_poolsADConfig,
			"workspacePoolsWorkspaceCreation":                             testAccDirectory_poolsWorkspaceCreation,
			"workspacePoolsWorkspaceCreationAD":                           testAccDirectory_poolsWorkspaceCreationAD,
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
			"workspaceProperties_runningModeAlwaysOn":                 testAccWorkspace_workspaceProperties_runningModeAlwaysOn,
			"workspaceProperties_runningModeAutoStopTimeoutInMinutes": testAccWorkspace_workspaceProperties_runningModeAutoStopTimeoutInMinutes,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
