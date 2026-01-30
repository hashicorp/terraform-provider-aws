// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.QuickSightServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Account is already subscribed to Quicksight",
	)
}

func TestAccQuickSight_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AccountSettings": {
			acctest.CtBasic: testAccAccountSettings_basic,
		},
		"AccountSubscription": {
			acctest.CtBasic:      testAccAccountSubscription_basic,
			acctest.CtDisappears: testAccAccountSubscription_disappears,
		},
		"IPRestriction": {
			acctest.CtBasic:      testAccIPRestriction_basic,
			acctest.CtDisappears: testAccIPRestriction_disappears,
			"update":             testAccIPRestriction_update,
		},
		"KeyRegistration": {
			acctest.CtBasic:      testAccKeyRegistration_basic,
			acctest.CtDisappears: testAccKeyRegistration_disappears,
		},
		"RoleCustomPermission": {
			acctest.CtBasic:      testAccRoleCustomPermission_basic,
			acctest.CtDisappears: testAccRoleCustomPermission_disappears,
			"update":             testAccRoleCustomPermission_update,
		},
		"RoleMembership": {
			acctest.CtBasic:      testAccRoleMembership_basic,
			acctest.CtDisappears: testAccRoleMembership_disappears,
			"role":               testAccRoleMembership_role,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
