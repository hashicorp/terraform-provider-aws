// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package objectvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// WarnAtMostOneOfChildren is equivalent to AtMostOneOfChildren, but returns a Warning instead of an error
func WarnAtMostOneOfChildren(expressions ...path.Expression) validator.Object {
	return warnAtMostOneOfChildrenValidator{
		pathExpressions: expressions,
	}
}

type warnAtMostOneOfChildrenValidator struct {
	pathExpressions path.Expressions
}

func (av warnAtMostOneOfChildrenValidator) Description(ctx context.Context) string {
	return av.MarkdownDescription(ctx)
}

func (av warnAtMostOneOfChildrenValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that one and only one attribute from this collection is set: %q", av.pathExpressions)
}

func (av warnAtMostOneOfChildrenValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := internal.ExactlyOneOfValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	var validateResp internal.ExactlyOneOfValidatorResponse

	av.validate(ctx, validateReq, &validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (av warnAtMostOneOfChildrenValidator) validate(ctx context.Context, req internal.ExactlyOneOfValidatorRequest, res *internal.ExactlyOneOfValidatorResponse) {
	count := 0
	expressions := req.PathExpression.MergeExpressions(av.pathExpressions...)

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

	if count > 1 {
		res.Diagnostics.Append(fwdiag.WarningInvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("%d attributes specified when at most one of %s may be set", count, expressions),
		))
	}
}
