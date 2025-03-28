// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// This is inspired by `stringvalidator.ExactlyOneOf` and `schemavalidator.ExactlyOneOfValidator`
// It returns a warning instead of an error.
// It could likely be moved to an internal validators package if useful elsewhere.

func warnAtMostOneOf(expressions ...path.Expression) validator.String {
	return AtMostOneOfValidator{
		PathExpressions: expressions,
	}
}

type AtMostOneOfValidator struct {
	PathExpressions path.Expressions
}

type AtMostOneOfValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type AtMostOneOfValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (av AtMostOneOfValidator) Description(ctx context.Context) string {
	return av.MarkdownDescription(ctx)
}

func (av AtMostOneOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that at most one attribute from this collection is set: %q", av.PathExpressions)
}

func (av AtMostOneOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := AtMostOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AtMostOneOfValidatorResponse{}

	av.Validate(ctx, validateReq, validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (av AtMostOneOfValidator) Validate(ctx context.Context, req AtMostOneOfValidatorRequest, res *AtMostOneOfValidatorResponse) {
	count := 0
	expressions := req.PathExpression.MergeExpressions(av.PathExpressions...)

	// If current attribute is unknown, delay validation
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Now that we know the current attribute is known, check whether it is
	// null to determine if it should contribute to the count. Later logic
	// will remove a duplicate matching path, should it be included in the
	// given expressions.
	if !req.ConfigValue.IsNull() {
		count++
	}

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		res.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(req.Path) {
				continue
			}

			var mpVal attr.Value
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			res.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if !mpVal.IsNull() {
				count++
			}
		}
	}

	if count > 1 {
		var attributeNames []string
		for _, expr := range expressions {
			attributeName, _ := expr.Steps().LastStep()
			attributeNames = append(attributeNames, attributeName.String())
		}

		res.Diagnostics.Append(warnInvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("%d attributes specified when at most one of [%s] is allowed", count, strings.Join(attributeNames, ", ")),
		))
	}
}

// warnInvalidAttributeCombinationDiagnostic returns a warning Diagnostic to be used when a schemavalidator of attributes is invalid.
func warnInvalidAttributeCombinationDiagnostic(path path.Path, description string) diag.Diagnostic {
	return diag.NewAttributeWarningDiagnostic(
		path,
		"Invalid Attribute Combination",
		description+"\n\nThis will be an error in a future version of the provider",
	)
}
