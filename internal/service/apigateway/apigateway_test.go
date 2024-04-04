// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.APIGatewayServiceID, testAccErrorCheckSkip)
}

// skips tests that have error messages indicating unsupported features
func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"no matching Route53Zone found",
	)
}

func TestAccAPIGateway_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic": testAccAccount_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
