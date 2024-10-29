// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ComposeAggregateImportStateCheckFunc lets you compose multiple ImportStateCheckFunc into
// a single ImportStateCheckFunc.
func ComposeAggregateImportStateCheckFunc(fs ...resource.ImportStateCheckFunc) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		var result []error

		for i, f := range fs {
			if err := f(is); err != nil {
				result = append(result, fmt.Errorf("Import check %d/%d error: %w", i+1, len(fs), err))
			}
		}

		return errors.Join(result...)
	}
}

func ImportCheckResourceAttr(key, expected string) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("Attribute '%s' expected 1 instance state, got %d", key, len(is))
		}

		rs := is[0]
		if rs.Attributes[key] != expected {
			return fmt.Errorf("Attribute '%s' expected %s, got %s", key, expected, rs.Attributes[key])
		}
		return nil
	}
}

func ImportCheckResourceAttrSet(key string, set bool) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("Attribute '%s' expected 1 instance state, got %d", key, len(is))
		}

		rs := is[0]
		if set && rs.Attributes[key] == "" {
			return fmt.Errorf("Attribute '%s' expected to be set, got not set", key)
		}

		if !set && rs.Attributes[key] != "" {
			return fmt.Errorf("Attribute '%s' expected to be not set, got set (%s)", key, rs.Attributes[key])
		}

		return nil
	}
}
