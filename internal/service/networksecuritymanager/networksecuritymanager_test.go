// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// An AWS account can be set as a Network Security Manager administrator only
// once, and the acceptance tests all use the current (Organizations
// management) account, so they must run serially.
func TestAccNetworkSecurityManager_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AdminAccount": {
			acctest.CtBasic:      testAccAdminAccount_basic,
			acctest.CtDisappears: testAccAdminAccount_disappears,
			"full":               testAccAdminAccount_full,
			"priority":           testAccAdminAccount_priority,
			"scopeFilter":        testAccAdminAccount_scopeFilter,
			"firewallTypeScope":  testAccAdminAccount_firewallTypeScope,
			"recreate":           testAccAdminAccount_recreate,
			"validation":         testAccAdminAccount_validation,
			"Identity":           testAccNetworkSecurityManagerAdminAccount_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
