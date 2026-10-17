// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Network Security Manager resources are account-level policy objects, and
// the acceptance tests all run in the current account, so they run serially.
func TestAccNetworkSecurityManager_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Scope": {
			acctest.CtBasic:      testAccScope_basic,
			acctest.CtDisappears: testAccScope_disappears,
			"full":               testAccScope_full,
			"resourceTypes":      testAccScope_resourceTypes,
			"selection":          testAccScope_selection,
			"explicitARNs":       testAccScope_explicitARNs,
			"expressions":        testAccScope_expressions,
			"albConfig":          testAccScope_albConfig,
			"publish":            testAccScope_publish,
			"description":        testAccScope_description,
			"updateToken":        testAccScope_updateToken,
			"replace":            testAccScope_replace,
			"recreate":           testAccScope_recreate,
			"validation":         testAccScope_validation,
			"accountFilter":      testAccScope_accountFilter,
			"Identity":           testAccNetworkSecurityManagerScope_identitySerial,
			"tags":               testAccNetworkSecurityManagerScope_tagsSerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
