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

func (v ExactlyOneOfValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v ExactlyOneOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that one and only one attribute from this collection is set: %q", v.PathExpressions)
}

func (v ExactlyOneOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := ValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	var validateResp ValidatorResponse

	v.validate(ctx, validateReq, &validateResp)

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

func (v ExactlyOneOfValidator) validate(ctx context.Context, req ValidatorRequest, resp *ValidatorResponse) {
	count := 0

	// If current attribute is unknown, delay validation.
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

	expressions := req.PathExpression.MergeExpressions(v.PathExpressions...)
	for _, expression := range req.PathExpression.MergeExpressions(v.PathExpressions...) {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// Skip self.
			if mp.Equal(req.Path) {
				continue
			}

			var mpVal attr.Value
			resp.Diagnostics.Append(req.Config.GetAttribute(ctx, mp, &mpVal)...)
			if resp.Diagnostics.HasError() {
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

	switch {
	case count == 0:
		resp.Diagnostics.Append(fwdiag.WarningInvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("No attribute specified when one (and only one) of %s is required", expressions),
		))
	case count > 1:
		resp.Diagnostics.Append(fwdiag.WarningInvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("%d attributes specified when one (and only one) of %s is required", count, expressions),
		))
	}
}
