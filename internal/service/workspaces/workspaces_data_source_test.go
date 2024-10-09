// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccWorkSpacesDataSource_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Bundle": {
			acctest.CtBasic:           testAccWorkspaceBundleDataSource_basic,
			"bundleIDAndNameConflict": testAccWorkspaceBundleDataSource_bundleIDAndNameConflict,
			"byOwnerName":             testAccWorkspaceBundleDataSource_byOwnerName,
			"privateOwner":            testAccWorkspaceBundleDataSource_privateOwner,
		},
		"Directory": {
			acctest.CtBasic: testAccDirectoryDataSource_basic,
		},
		"Image": {
			acctest.CtBasic: testAccImageDataSource_basic,
		},
		"Workspace": {
			"byWorkspaceID":                     testAccWorkspaceDataSource_byWorkspaceID,
			"byDirectoryID_userName":            testAccWorkspaceDataSource_byDirectoryID_userName,
			"workspaceIDAndDirectoryIDConflict": testAccWorkspaceDataSource_workspaceIDAndDirectoryIDConflict,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
