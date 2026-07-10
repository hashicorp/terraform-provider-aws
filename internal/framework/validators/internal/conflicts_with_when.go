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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String = (*ConflictsWithWhenValidator)(nil)
)

type ConflictsWithWhenValidator struct {
	When            func(context.Context, attr.Value) bool
	PathExpressions path.Expressions
}

func (v ConflictsWithWhenValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v ConflictsWithWhenValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Ensure that when this attribute value matches the condition, the following are not also configured: %[1]q", v.PathExpressions)
}

func (v ConflictsWithWhenValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	validateRequest := ValidatorRequest{
		Config:         request.Config,
		ConfigValue:    request.ConfigValue,
		Path:           request.Path,
		PathExpression: request.PathExpression,
	}
	var validateResponse ValidatorResponse

	v.validate(ctx, validateRequest, &validateResponse)

	response.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (v ConflictsWithWhenValidator) validate(ctx context.Context, request ValidatorRequest, response *ValidatorResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if !v.When(ctx, request.ConfigValue) {
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

			if !mpVal.IsNull() {
				// Collect all errors.
				responseDiags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
					request.Path,
					fmt.Sprintf("Attribute %[1]q must not be configured", mp),
				))
			}
		}
	}

	response.Diagnostics.Append(responseDiags...)
}
