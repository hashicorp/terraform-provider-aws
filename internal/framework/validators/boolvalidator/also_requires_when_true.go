// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package boolvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// AlsoRequiresWhenTrue checks that each path.Expression has a non-null
// configuration value when the bool attribute being validated has a known
// value of true.
//
// Unlike boolvalidator.AlsoRequires in
// github.com/hashicorp/terraform-plugin-framework-validators, which fires
// whenever the attribute is non-null regardless of value, this validator
// fires only when the attribute is explicitly true. Use it for asymmetric
// constraints such as "when `foo_enabled` is true, `foo_arn` must be set."
//
// Relative path.Expressions are resolved using the attribute being
// validated.
func AlsoRequiresWhenTrue(expressions ...path.Expression) validator.Bool {
	return alsoRequiresWhenTrueValidator{
		pathExpressions: expressions,
	}
}

type alsoRequiresWhenTrueValidator struct {
	pathExpressions path.Expressions
}

func (v alsoRequiresWhenTrueValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v alsoRequiresWhenTrueValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Ensure that when this attribute is true, the following are also configured: %q", v.pathExpressions)
}

func (v alsoRequiresWhenTrueValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If the gating attribute is not configured, the condition cannot be
	// "true," so there is nothing to enforce. If it is unknown, defer
	// validation until it becomes known.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Only fire when the gating attribute is true.
	if !req.ConfigValue.ValueBool() {
		return
	}

	expressions := req.PathExpression.MergeExpressions(v.pathExpressions...)

	for _, expression := range expressions {
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
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			// Defer if any target is unknown; we cannot decide yet.
			if mpVal.IsUnknown() {
				return
			}

			if mpVal.IsNull() {
				resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
					req.Path,
					fmt.Sprintf("Attribute %q must be specified when %q is true", mp, req.Path),
				))
			}
		}
	}
}
