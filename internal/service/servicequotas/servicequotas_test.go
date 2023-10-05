// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccServiceQuotas_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Template": {
			"basic":      testAccTemplate_basic,
			"disappears": testAccTemplate_disappears,
			"value":      testAccTemplate_value,
		},
		"TemplateAssociation": {
			"basic":       testAccTemplateAssociation_basic,
			"disappears":  testAccTemplateAssociation_disappears,
			"skipDestroy": testAccTemplateAssociation_skipDestroy,
		},
		"TemplatesDataSource": {
			"basic": testAccTemplatesDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
