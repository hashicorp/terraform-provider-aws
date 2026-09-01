// Copyright IBM Corp. 2014, 2026
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
		resource, err := populateFromResourceState(s, resourceName, arnFormat)
		if err != nil {
			return err
		}

		return CheckResourceAttrGlobalARN(ctx, resourceName, attributeName, arnService, resource)(s)
	}
}

func CheckResourceAttrGlobalARNNoAccountFormat(resourceName, attributeName, arnService, arnFormat string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, err := populateFromResourceState(s, resourceName, arnFormat)
		if err != nil {
			return err
		}

		return CheckResourceAttrGlobalARNNoAccount(resourceName, attributeName, arnService, resource)(s)
	}
}

func CheckResourceAttrRegionalARNFormat(ctx context.Context, resourceName, attributeName, arnService, arnFormat string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, err := populateFromResourceState(s, resourceName, arnFormat)
		if err != nil {
			return err
		}

		return CheckResourceAttrRegionalARN(ctx, resourceName, attributeName, arnService, resource)(s)
	}
}

func populateFromResourceState(s *terraform.State, resourceName, format string) (string, error) {
	is, err := PrimaryInstanceState(s, resourceName)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	str := format
	for str != "" {
		var (
			stuff string
			found bool
		)
		stuff, str, found = strings.Cut(str, "{")
		buf.WriteString(stuff)
		if found {
			var param string
			param, str, found = strings.Cut(str, "}")
			if !found {
				return "", fmt.Errorf("missing closing '}' in format %q", format)
			}

			attr, ok := is.Attributes[param]
			if !ok {
				return "", fmt.Errorf("attribute %q not found in resource %q, referenced in format %q", param, resourceName, format)
			}
			buf.WriteString(attr)
		}
	}

	return buf.String(), nil
}
