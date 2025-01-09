// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func CheckResourceAttrGlobalARNFormat(ctx context.Context, resourceName, attributeName, arnService, arnFormat string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := PrimaryInstanceState(s, resourceName)
		if err != nil {
			return err
		}

		var foo strings.Builder
		str := arnFormat
		for str != "" {
			var (
				stuff string
				found bool
			)
			stuff, str, found = strings.Cut(str, "{")
			foo.WriteString(stuff)
			if found {
				var param string
				param, str, found = strings.Cut(str, "}")
				if !found {
					return fmt.Errorf("missing closing '}' in ARN format %q", arnFormat)
				}

				attr, ok := is.Attributes[param]
				if !ok {
					return fmt.Errorf("attribute %q not found in resource %q, referenced in ARN format %q", param, resourceName, arnFormat)
				}
				foo.WriteString(attr)
			}
		}
		resource := foo.String()

		return CheckResourceAttrGlobalARN(ctx, resourceName, attributeName, arnService, resource)(s)
	}
}
