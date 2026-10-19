// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Object = (*requiredTogetherValidator)(nil)
)

func RequiredTogetherValidator(expressions ...path.Expression) requiredTogetherValidator {
	return requiredTogetherValidator{
		when:            always{},
		pathExpressions: expressions,
	}
}

type requiredTogetherValidator struct {
	when            when
	pathExpressions path.Expressions
}

func (v requiredTogetherValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v requiredTogetherValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("These attributes must be configured together: %[1]q", v.pathExpressions)
}

func (v requiredTogetherValidator) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	validateRequest := validatorRequest{
		Config:         request.Config,
		ConfigValue:    request.ConfigValue,
		Path:           request.Path,
		PathExpression: request.PathExpression,
	}
	var validateResponse validatorResponse

	v.validate(ctx, validateRequest, &validateResponse)

	response.Diagnostics.Append(validateResponse.Diagnostics...)
}

func (v requiredTogetherValidator) validate(ctx context.Context, request validatorRequest, response *validatorResponse) {
	if request.ConfigValue.IsUnknown() {
		return
	}

	if !v.when.Eval(ctx, request.ConfigValue) {
		return
	}

	var configuredPaths, foundPaths, unknownPaths path.Paths
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

			foundPaths.Append(mp)

			var mpVal attr.Value
			response.Diagnostics.Append(request.Config.GetAttribute(ctx, mp, &mpVal)...)
			if response.Diagnostics.HasError() {
				continue
			}

			switch {
			case mpVal.IsUnknown():
				// If value is unknown, it may be null or a value, so we cannot
				// know if the validator should succeed or not. Collect the path
				// so we use it to skip the validation later and continue to
				// collect all path matching diagnostics.
				unknownPaths.Append(mp)
				continue
			case mpVal.IsNull():
				// If value is null, move onto the next one.
				continue
			default:
				// Value is known and not null, it is configured.
				configuredPaths.Append(mp)
			}
		}
	}

	// Return early if all paths were null.
	if len(configuredPaths) == 0 {
		return
	}

	// If there are unknown values, we cannot know if the validator should
	// succeed or not.
	if len(unknownPaths) > 0 {
		return
	}

	// If configured paths does not equal all matched paths, then something
	// was missing. We compare the number of matched paths instead of path
	// expressions to prevent false negatives with path expressions that match
	// more than one path.
	if len(configuredPaths) != len(foundPaths) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			configuredPaths[0],
			v.Description(ctx),
		))
	}
}
