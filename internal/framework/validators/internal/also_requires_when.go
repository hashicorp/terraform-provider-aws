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
	_ validator.Bool   = (*alsoRequiresWhenValidator)(nil)
	_ validator.String = (*alsoRequiresWhenValidator)(nil)
)

type When interface {
	// Eval returns true if the condition is met for the given attribute value.
	Eval(context.Context, attr.Value) bool
	// String returns a string representation of the condition, usable in an error message.
	String() string
}

func AlsoRequiresWhenValidator(when When, expressions ...path.Expression) alsoRequiresWhenValidator {
	return alsoRequiresWhenValidator{whenValidator{
		when:            when,
		pathExpressions: expressions,
	}}
}

type alsoRequiresWhenValidator struct {
	whenValidator
}

func (v alsoRequiresWhenValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v alsoRequiresWhenValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Ensure that when this attribute value matches the condition, the following are also configured: %[1]q", v.pathExpressions)
}

func (v alsoRequiresWhenValidator) ValidateBool(ctx context.Context, request validator.BoolRequest, response *validator.BoolResponse) {
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

func (v alsoRequiresWhenValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
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

func (v alsoRequiresWhenValidator) validate(ctx context.Context, request ValidatorRequest, response *ValidatorResponse) {
	v.whenValidator.validate(ctx, request, response, v.eval)
}

func (v alsoRequiresWhenValidator) eval(_ context.Context, requestPath path.Path, matchedPath path.Path, matchedValue attr.Value) diag.Diagnostics {
	var diags diag.Diagnostics
	if matchedValue.IsNull() {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			requestPath,
			fmt.Sprintf("Attribute %[1]q must be configured when %[2]q %[3]s", matchedPath, requestPath, v.when.String()),
		))
	}
	return diags
}

type whenValidator struct {
	when            When
	pathExpressions path.Expressions
}

func (v whenValidator) validate(ctx context.Context, request ValidatorRequest, response *ValidatorResponse, cb func(context.Context, path.Path, path.Path, attr.Value) diag.Diagnostics) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if !v.when.Eval(ctx, request.ConfigValue) {
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

			// Collect all errors.
			responseDiags.Append(cb(ctx, request.Path, mp, mpVal)...)
		}
	}

	response.Diagnostics.Append(responseDiags...)
}
