// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccFMS_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AdminAccount": {
			"basic": testAccAdminAccount_basic,
		},
		"Policy": {
			"basic":                  testAccPolicy_basic,
			"cloudfrontDistribution": testAccPolicy_cloudFrontDistribution,
			"includeMap":             testAccPolicy_includeMap,
			"update":                 testAccPolicy_update,
			"policyOption":           testAccPolicy_policyOption,
			"resourceTags":           testAccPolicy_resourceTags,
			"tags":                   testAccPolicy_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
