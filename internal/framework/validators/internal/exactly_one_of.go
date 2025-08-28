// Copyright (c) HashiCorp, Inc.
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
	// _ validator.Bool    = ExactlyOneOfValidator{}
	// _ validator.Float32 = ExactlyOneOfValidator{}
	// _ validator.Float64 = ExactlyOneOfValidator{}
	// _ validator.Int32   = ExactlyOneOfValidator{}
	// _ validator.Int64   = ExactlyOneOfValidator{}
	// _ validator.List    = ExactlyOneOfValidator{}
	// _ validator.Map     = ExactlyOneOfValidator{}
	// _ validator.Number  = ExactlyOneOfValidator{}
	// _ validator.Object  = ExactlyOneOfValidator{}
	// _ validator.Set     = ExactlyOneOfValidator{}
	_ validator.String = ExactlyOneOfValidator{}
	// _ validator.Dynamic = ExactlyOneOfValidator{}
)

type ExactlyOneOfValidator struct {
	PathExpressions path.Expressions
}

type ExactlyOneOfValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type ExactlyOneOfValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

func (av ExactlyOneOfValidator) Description(ctx context.Context) string {
	return av.MarkdownDescription(ctx)
}

func (av ExactlyOneOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that one and only one attribute from this collection is set: %q", av.PathExpressions)
}

func (av ExactlyOneOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
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

func (av ExactlyOneOfValidator) Validate(ctx context.Context, req ExactlyOneOfValidatorRequest, res *ExactlyOneOfValidatorResponse) {
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
