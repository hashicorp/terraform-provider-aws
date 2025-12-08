// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"errors"
	"fmt"
	"regexp"

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

func ImportCheckNoResourceAttr(key string) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("Attribute '%s' expected 1 instance state, got %d", key, len(is))
		}

		rs := is[0]
		if v, ok := rs.Attributes[key]; ok {
			return fmt.Errorf("Attribute '%s' expected no value, got %q", key, v)
		}
		return nil
	}
}

func ImportCheckResourceAttr(key, expected string) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("Attribute '%s' expected 1 instance state, got %d", key, len(is))
		}

		rs := is[0]
		if rs.Attributes[key] != expected {
			return fmt.Errorf("Attribute '%s' expected %q, got %q", key, expected, rs.Attributes[key])
		}
		return nil
	}
}

func ImportMatchResourceAttr(key string, r *regexp.Regexp) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("Attribute '%s' expected 1 instance state, got %d", key, len(is))
		}

		rs := is[0]
		if !r.MatchString(rs.Attributes[key]) {
			return fmt.Errorf("Attribute '%s' didn't match %q, got %#v", key, r.String(), rs.Attributes[key])
		}
		return nil
	}
}

func ImportCheckResourceAttrSet(key string) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("Attribute '%s' expected 1 instance state, got %d", key, len(is))
		}

		rs := is[0]
		if rs.Attributes[key] == "" {
			return fmt.Errorf("Attribute '%s' expected to be set, had no value", key)
		}

		return nil
	}
}

func ImportStateIDAccountID(ctx context.Context) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return AccountID(ctx), nil
	}
}
