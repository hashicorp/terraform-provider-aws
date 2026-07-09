// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type AlsoRequiresWhenEqualsValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type AlsoRequiresWhenEqualsValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

type AlsoRequiresWhenEqualsValidator struct {
	Value           attr.Value
	PathExpressions path.Expressions
}

func (v AlsoRequiresWhenEqualsValidator) Validate(ctx context.Context, request AlsoRequiresWhenEqualsValidatorRequest, response *AlsoRequiresWhenEqualsValidatorResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if !request.ConfigValue.Equal(v.Value) {
		return
	}

	var responseDiags diag.Diagnostics

	for _, expression := range request.PathExpression.MergeExpressions(v.PathExpressions...) {
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
					fmt.Sprintf("Attribute %[1]q must be specified when %[2]q is %[3]q", mp, request.Path, v.Value),
				))
			}
		}
	}

	response.Diagnostics.Append(responseDiags...)
}
