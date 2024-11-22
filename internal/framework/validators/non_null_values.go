// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Set = nonNullValuesValidator{}
var _ function.SetParameterValidator = nonNullValuesValidator{}

type nonNullValuesValidator struct{}

func (v nonNullValuesValidator) Description(_ context.Context) string {
	return "null values are not permitted"
}

func (v nonNullValuesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v nonNullValuesValidator) ValidateSet(_ context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := req.ConfigValue.Elements()

	for _, e := range elements {
		// Only evaluate known values for null
		if e.IsUnknown() {
			continue
		}

		if e.IsNull() {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Null Set Value",
				"This attribute contains a null value.",
			)
		}
	}
}

func (v nonNullValuesValidator) ValidateParameterSet(ctx context.Context, req function.SetParameterValidatorRequest, resp *function.SetParameterValidatorResponse) {
	if req.Value.IsNull() || req.Value.IsUnknown() {
		return
	}

	elements := req.Value.Elements()

	for _, e := range elements {
		// Only evaluate known values for null
		if e.IsUnknown() {
			continue
		}

		if e.IsNull() {
			resp.Error = function.ConcatFuncErrors(
				resp.Error,
				function.NewArgumentFuncError(
					req.ArgumentPosition,
					"Null Set Value: This attribute contains a null value.",
				),
			)
		}
	}
}

// NonNullValues returns a validator which ensures that any configured set
// only contains non-null values.
func NonNullValues() nonNullValuesValidator {
	return nonNullValuesValidator{}
}
