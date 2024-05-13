// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppFabric_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AppAuthorization": {
			"basic":        testAccAppAuthorization_basic,
			"disappears":   testAccAppAuthorization_disappears,
			"apiKeyUpdate": testAccAppAuthorization_apiKeyUpdate,
			"oath2Update":  testAccAppAuthorization_oath2Update,
			names.AttrTags: testAccAppAuthorization_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
