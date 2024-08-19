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
			acctest.CtBasic:      testAccAdminAccount_basic,
			acctest.CtDisappears: testAccAdminAccount_disappears,
		},
		"Policy": {
			"alb":                    testAccPolicy_alb,
			acctest.CtBasic:          testAccPolicy_basic,
			"cloudfrontDistribution": testAccPolicy_cloudFrontDistribution,
			acctest.CtDisappears:     testAccPolicy_disappears,
			"includeMap":             testAccPolicy_includeMap,
			"policyOption":           testAccPolicy_policyOption,
			"resourceTags":           testAccPolicy_resourceTags,
			"securityGroup":          testAccPolicy_securityGroup,
			"tags":                   testAccPolicy_tags,
			"update":                 testAccPolicy_update,
			"rscSet":                 testAccPolicy_rscSet,
		},
		"ResourceSet": {
			acctest.CtBasic:      testAccFMSResourceSet_basic,
			acctest.CtDisappears: testAccFMSResourceSet_disappears,
			"tags":               testAccFMSResourceSet_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
