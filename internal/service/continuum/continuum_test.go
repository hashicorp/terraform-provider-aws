// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Continuum allows one collector per account and Region, and creating an environment starts a scan,
// so every acceptance test runs serially.
func TestAccContinuum_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Integration": {
			acctest.CtBasic:        testAccIntegration_basic,
			acctest.CtDisappears:   testAccIntegration_disappears,
			"updateExternalID":     testAccIntegration_updateExternalID,
			"updateRole":           testAccIntegration_updateRole,
			"identity":             testAccContinuumIntegration_identitySerial,
			"List_basic":           testAccIntegration_List_basic,
			"List_includeResource": testAccIntegration_List_includeResource,
			"List_regionOverride":  testAccIntegration_List_regionOverride,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
