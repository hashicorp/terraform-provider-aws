// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

var (
	// _ validator.Bool    = (*ExactlyOneOfValidator)(nil)
	// _ validator.Float32 = (*ExactlyOneOfValidator)(nil)
	// _ validator.Float64 = (*ExactlyOneOfValidator)(nil)
	// _ validator.Int32   = (*ExactlyOneOfValidator)(nil)
	// _ validator.Int64   = (*ExactlyOneOfValidator)(nil)
	// _ validator.List    = (*ExactlyOneOfValidator)(nil)
	// _ validator.Map     = (*ExactlyOneOfValidator)(nil)
	// _ validator.Number  = (*ExactlyOneOfValidator)(nil)
	// _ validator.Object  = (*ExactlyOneOfValidator)(nil)
	// _ validator.Set     = (*ExactlyOneOfValidator)(nil)
	_ validator.String = (*ExactlyOneOfValidator)(nil)
	// _ validator.Dynamic = (*ExactlyOneOfValidator)(nil)
)

type ExactlyOneOfValidator struct {
	PathExpressions path.Expressions
}

type ValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type ValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (av ExactlyOneOfValidator) Description(ctx context.Context) string {
	return av.MarkdownDescription(ctx)
}

func (av ExactlyOneOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that one and only one attribute from this collection is set: %q", av.PathExpressions)
}

func (av ExactlyOneOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := ValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	var validateResp ValidatorResponse

	av.Validate(ctx, validateReq, &validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (av ExactlyOneOfValidator) Validate(ctx context.Context, req ValidatorRequest, res *ValidatorResponse) {
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
