// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type alsoRequiresWhenEqualsValidator[T ~string] struct {
	value           T
	pathExpressions path.Expressions
}

func (v alsoRequiresWhenEqualsValidator[T]) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v alsoRequiresWhenEqualsValidator[T]) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Ensure that when this attribute equals %[1]q, the following are also configured: %[2]q", v.value, v.pathExpressions)
}

func (v alsoRequiresWhenEqualsValidator[T]) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if T(request.ConfigValue.ValueString()) != v.value {
		return
	}

	var responseDiags diag.Diagnostics

	for _, expression := range request.PathExpression.MergeExpressions(v.pathExpressions...) {
		matchedPaths, diags := request.Config.PathMatches(ctx, expression)
		response.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// Skip self.
			if mp.Equal(request.Path) {
				continue
			}

			var mpVal attr.Value
			response.Diagnostics.Append(request.Config.GetAttribute(ctx, mp, &mpVal)...)
			if response.Diagnostics.HasError() {
				return
			}

			// Defer if any target is unknown; we cannot decide yet.
			if mpVal.IsUnknown() {
				return
			}

			if mpVal.IsNull() {
				// Collect all errors.
				responseDiags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
					request.Path,
					fmt.Sprintf("Attribute %[1]q must be specified when %[2]q is %[3]q", mp, request.Path, v.value),
				))
			}
		}
	}

	response.Diagnostics.Append(responseDiags...)
}

func AlsoRequiresWhenEquals[T ~string](value T, expressions ...path.Expression) validator.String {
	return alsoRequiresWhenEqualsValidator[T]{
		value:           value,
		pathExpressions: expressions,
	}
}
