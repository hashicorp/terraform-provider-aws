// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAccount_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AlternateContact": {
			acctest.CtBasic:      testAccAlternateContact_basic,
			acctest.CtDisappears: testAccAlternateContact_disappears,
			"AccountID":          testAccAlternateContact_accountID,
		},
		"PrimaryContact": {
			acctest.CtBasic: testAccPrimaryContact_basic,
		},
		"Region": {
			acctest.CtBasic: testAccRegion_basic,
			"AccountID":     testAccRegion_accountID,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
