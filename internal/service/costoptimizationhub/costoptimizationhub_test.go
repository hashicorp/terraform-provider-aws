// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCostOptimizationHub_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"EnrollmentStatus": {
			acctest.CtBasic:         testAccEnrollmentStatus_basic,
			acctest.CtDisappears:    testAccEnrollmentStatus_disappears,
			"includeMemberAccounts": testAccEnrollmentStatus_includeMemberAccounts,
		},
		"Preferences": {
			acctest.CtBasic:                    testAccPreferences_basic,
			acctest.CtDisappears:               testAccPreferences_disappears,
			"memberAccountsDiscountVisibility": testAccPreferences_memberAccountsDiscountVisibility,
			"savingsEstimationMode":            testAccPreferences_savingsEstimationMode,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
