// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// objectWarnExactlyOneOfChildren acts similarly to `objectvalidator.ExactlyOneOf` except that it returns a Warning and
// that it doesn't include the Object in the count of matched attributes.
func objectWarnExactlyOneOfChildren(expressions ...path.Expression) validator.Object {
	return ExactlyOneOfChildrenValidator{
		PathExpressions: expressions,
	}
}

type ExactlyOneOfChildrenValidator struct {
	PathExpressions path.Expressions
}

func (av ExactlyOneOfChildrenValidator) Description(ctx context.Context) string {
	return av.MarkdownDescription(ctx)
}

func (av ExactlyOneOfChildrenValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that one and only one attribute from this collection is set: %q", av.PathExpressions)
}

func (av ExactlyOneOfChildrenValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := ExactlyOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	var validateResp ExactlyOneOfValidatorResponse

	av.Validate(ctx, validateReq, &validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (av ExactlyOneOfChildrenValidator) Validate(ctx context.Context, req ExactlyOneOfValidatorRequest, res *ExactlyOneOfValidatorResponse) {
	count := 0
	expressions := req.PathExpression.MergeExpressions(av.PathExpressions...)

	// If current attribute is unknown, delay validation
	if req.ConfigValue.IsUnknown() {
		return
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
			// TODO: Technically,this would be a misconfguration, so we should probably return an internal error
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

	if count == 0 {
		res.Diagnostics.Append(fwdiag.WarningInvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("No attribute specified when one (and only one) of %s is required", expressions),
		))
	}

	if count > 1 {
		res.Diagnostics.Append(fwdiag.WarningInvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("%d attributes specified when one (and only one) of %s is required", count, expressions),
		))
	}
}
