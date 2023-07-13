// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccServiceCatalog_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"OrganizationsAccess": {
			"basic": testAccOrganizationsAccess_basic,
		},
		"PortfolioShare": {
			"basic":              testAccPortfolioShare_basic,
			"sharePrincipals":    testAccPortfolioShare_sharePrincipals,
			"organizationalUnit": testAccPortfolioShare_organizationalUnit,
			"disappears":         testAccPortfolioShare_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
