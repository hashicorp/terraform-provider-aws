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
	_ validator.String = (*exactlyOneOfWhenValidator)(nil)
)

func ExactlyOneOfWhenValidator(when When, expressions ...path.Expression) exactlyOneOfWhenValidator {
	return exactlyOneOfWhenValidator{oneOfWhenValidator{
		when:            when,
		pathExpressions: expressions,
	}}
}

type exactlyOneOfWhenValidator struct {
	oneOfWhenValidator
}

func (v exactlyOneOfWhenValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v exactlyOneOfWhenValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("Ensure that when this attribute value matches the condition, one and only one attribute from this collection is configured: %[1]q", v.pathExpressions)
}

func (v exactlyOneOfWhenValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
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

func (v exactlyOneOfWhenValidator) validate(ctx context.Context, request ValidatorRequest, response *ValidatorResponse) {
	v.oneOfWhenValidator.validate(ctx, request, response, v.eval)
}

func (v exactlyOneOfWhenValidator) eval(_ context.Context, requestPath path.Path, expressions path.Expressions, count int) diag.Diagnostics {
	var diags diag.Diagnostics
	switch {
	case count == 0:
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			requestPath,
			fmt.Sprintf("No attribute configured when one (and only one) of %[1]s is required", expressions),
		))
	case count > 1:
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			requestPath,
			fmt.Sprintf("%[1]d attributes configured when one (and only one) of %[2]s is required", count, expressions),
		))
	}
	return diags
}

type oneOfWhenValidator struct {
	when            When
	pathExpressions path.Expressions
}

func (v oneOfWhenValidator) validate(ctx context.Context, request ValidatorRequest, response *ValidatorResponse, cb func(context.Context, path.Path, path.Expressions, int) diag.Diagnostics) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if !v.when.Eval(ctx, request.ConfigValue) {
		return
	}

	count := 0
	expressions := request.PathExpression.MergeExpressions(v.pathExpressions...)
	for _, expression := range expressions {
		matchedPaths, diags := request.Config.PathMatches(ctx, expression)
		response.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
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
				count++
			}
		}
	}

	// Collect all diagnositics.
	response.Diagnostics.Append(cb(ctx, request.Path, expressions, count)...)
}
