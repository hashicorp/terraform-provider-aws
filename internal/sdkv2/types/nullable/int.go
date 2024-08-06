// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nullable

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	TypeNullableInt = schema.TypeString
)

type Int string

func (i Int) IsNull() bool {
	return i == ""
}

func (i Int) ValueInt32() (int32, bool, error) {
	if i.IsNull() {
		return 0, true, nil
	}

	value, err := strconv.ParseInt(string(i), 10, 32)
	if err != nil {
		return 0, false, err
	}
	return int32(value), false, nil
}

func (i Int) ValueInt64() (int64, bool, error) {
	if i.IsNull() {
		return 0, true, nil
	}

	value, err := strconv.ParseInt(string(i), 10, 64)
	if err != nil {
		return 0, false, err
	}
	return value, false, nil
}

// ValidateTypeStringNullableInt provides custom error messaging for TypeString ints
// Some arguments require an int value or unspecified, empty field.
func ValidateTypeStringNullableInt(v interface{}, k string) (ws []string, es []error) {
	value, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %w", k, value, err))
	}

	return
}

// ValidateTypeStringNullableIntAtLeast provides custom error messaging for TypeString ints
// Some arguments require an int value or unspecified, empty field.
func ValidateTypeStringNullableIntAtLeast(min int64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (ws []string, es []error) {
		value, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if value == "" {
			return
		}

		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %w", k, value, err))
			return
		}

		if v < min {
			es = append(es, fmt.Errorf("expected %s to be at least (%d), got %d", k, min, v))
		}

		return
	}
}

// ValidateTypeStringNullableIntBetween provides custom error messaging for TypeString ints
// Some arguments require an int value or unspecified, empty field.
func ValidateTypeStringNullableIntBetween(min, max int64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (ws []string, es []error) {
		value, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if value == "" {
			return
		}

		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			es = append(es, fmt.Errorf("%s: cannot parse '%s' as int: %w", k, value, err))
			return
		}

		if v < min || v > max {
			es = append(es, fmt.Errorf("expected %s to be at between (%d) and (%d), got %d", k, min, max, v))
		}

		return
	}
}
