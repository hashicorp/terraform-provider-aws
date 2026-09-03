// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// serializeDelay is applied between serialized subtests. Account Access allows
// only one Application per IAM Identity Center instance, and a test account has
// a single instance, so every Application-creating test contends for it.
const serializeDelay = 5 * time.Second

// TestAccAccountAccess_serial runs every Application-related acceptance group
// sequentially. AWS Account Access enforces a 1:1 Application-to-Identity-
// Center-instance constraint, so concurrent CreateApplication calls against the
// shared organization instance can fail with AlreadyCreatedException. Each
// group is independently runnable and its CheckDestroy verifies cleanup before
// the next group begins.
func TestAccAccountAccess_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Application": {
			acctest.CtBasic:        testAccApplication_basic,
			acctest.CtDisappears:   testAccApplication_disappears,
			"tags":                 testAccAccountAccessApplication_tagsSerial,
			"Identity":             testAccAccountAccessApplication_identitySerial,
			"List_basic":           testAccAccountAccessApplication_List_basic,
			"List_includeResource": testAccAccountAccessApplication_List_includeResource,
		},
		"ApplicationDataSource": {
			acctest.CtBasic:             testAccApplicationDataSource_basic,
			"IdentityCenterInstanceARN": testAccApplicationDataSource_identityCenterInstanceARN,
			"tags":                      testAccAccountAccessApplicationDataSource_tagsSerial,
		},
		"Entitlement": {
			acctest.CtBasic:        testAccAccountAccessEntitlement_basic,
			acctest.CtDisappears:   testAccAccountAccessEntitlement_disappears,
			"group":                testAccAccountAccessEntitlement_group,
			"Identity":             testAccAccountAccessEntitlement_identitySerial,
			"List_basic":           testAccAccountAccessEntitlement_List_basic,
			"List_includeResource": testAccAccountAccessEntitlement_List_includeResource,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, serializeDelay)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckSSOAdminInstances(ctx, t)
	acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "account-access.amazonaws.com")
}
