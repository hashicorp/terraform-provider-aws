// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdktypes

import (
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

const (
	TypeDuration = schema.TypeString
)

type Duration string

func (d Duration) IsNull() bool {
	return d == ""
}

func (d Duration) Value() (time.Duration, bool, error) {
	if d.IsNull() {
		return 0, true, nil
	}

	value, err := time.ParseDuration(string(d))
	if err != nil {
		return 0, false, err
	}
	return value, false, nil
}

func ValidateDuration(i any, path cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Diagnostics{errs.NewIncorrectValueTypeAttributeError(path, "string")}
	}

	duration, _, err := Duration(v).Value()
	if err != nil {
		return diag.Diagnostics{errs.NewInvalidValueAttributeErrorf(path, "Cannot be parsed as duration: %s", err)}
	}

	if duration < 0 {
		return diag.Diagnostics{errs.NewInvalidValueAttributeError(path, "Must be greater than zero")}
	}

	return nil
}

func ValidateDurationBetween(min, max time.Duration) schema.SchemaValidateDiagFunc {
	return func(i any, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Diagnostics{errs.NewIncorrectValueTypeAttributeError(path, "string")}
		}

		duration, _, err := Duration(v).Value()
		if err != nil {
			return diag.Diagnostics{errs.NewInvalidValueAttributeErrorf(path, "Cannot be parsed as duration: %s", err)}
		}

		if duration < min || duration > max {
			return diag.Diagnostics{errs.NewInvalidValueAttributeErrorf(path, "Expected to be in the range (%d - %d)", min, max)}
		}

		return nil
	}
}
