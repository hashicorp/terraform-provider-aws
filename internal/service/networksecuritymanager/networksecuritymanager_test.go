// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// An AWS account can be set as a Network Security Manager administrator only
// once, and the acceptance tests all use the current (Organizations
// management) account, so they must run serially. Each test case creates and
// removes the administrator account, and the service rejects setting it again
// for a while after a removal, so related checks are grouped into as few test
// cases as possible.
func TestAccNetworkSecurityManager_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AdminAccount": {
			acctest.CtBasic:      testAccAdminAccount_basic,
			acctest.CtDisappears: testAccAdminAccount_disappears,
			"full":               testAccAdminAccount_full,
			"adminScope":         testAccAdminAccount_adminScope,
			"Identity":           testAccNetworkSecurityManagerAdminAccount_identitySerial,
			"List_basic":         testAccAdminAccount_List_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
