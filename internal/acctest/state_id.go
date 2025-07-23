// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// AttrImportStateIdFunc is a resource.ImportStateIdFunc that returns the value of the specified attribute
func AttrImportStateIdFunc(resourceName, attrName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[attrName], nil
	}
}

// CrossRegionAttrImportStateIdFunc is a resource.ImportStateIdFunc that returns the value
// of the specified attribute and appends the region
func CrossRegionAttrImportStateIdFunc(resourceName, attrName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		id := rs.Primary.Attributes[attrName]
		region, ok := rs.Primary.Attributes[names.AttrRegion]
		if !ok {
			return "", fmt.Errorf("Attribute \"region\" not found in %s", resourceName)
		}

		return id + "@" + region, nil
	}
}

// CrossRegionImportStateIdFunc is a resource.ImportStateIdFunc that appends the region
func CrossRegionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		id := rs.Primary.ID
		region, ok := rs.Primary.Attributes[names.AttrRegion]
		if !ok {
			return "", fmt.Errorf("Attribute \"region\" not found in %s", resourceName)
		}

		return id + "@" + region, nil
	}
}

// CrossRegionImportStateIdFuncAdapter adapts an ImportStateIdFunc by appending the region
func CrossRegionImportStateIdFuncAdapter(resourceName string, f Func) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		id, err := f(resourceName)(s)
		if err != nil {
			return "", err
		}

		region, ok := rs.Primary.Attributes[names.AttrRegion]
		if !ok {
			return "", fmt.Errorf("Attribute \"region\" not found in %s", resourceName)
		}

		return id + "@" + region, nil
	}
}

type Func func(resourceName string) resource.ImportStateIdFunc
