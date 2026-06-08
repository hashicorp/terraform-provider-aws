// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func CheckResourceAttrFormat(ctx context.Context, resourceName, attributeName, format string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue, err := populateFromResourceState(s, resourceName, format)
		if err != nil {
			return err
		}

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}
