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
	_ validator.String = (*conflictsWithWhenValidator)(nil)
)

func ConflictsWithWhenValidator(when When, expressions ...path.Expression) conflictsWithWhenValidator {
	return conflictsWithWhenValidator{whenValidator{
		when:            when,
		pathExpressions: expressions,
	}}
}

type conflictsWithWhenValidator struct {
	whenValidator
}

func (v conflictsWithWhenValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v conflictsWithWhenValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Ensure that when this attribute value matches the condition, the following are not also configured: %[1]q", v.pathExpressions)
}

func (v conflictsWithWhenValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
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

func (v conflictsWithWhenValidator) validate(ctx context.Context, request ValidatorRequest, response *ValidatorResponse) {
	v.whenValidator.validate(ctx, request, response, v.eval)
}

func (v conflictsWithWhenValidator) eval(_ context.Context, requestPath path.Path, matchedPath path.Path, matchedValue attr.Value) diag.Diagnostics {
	var diags diag.Diagnostics
	if !matchedValue.IsNull() {
		diags.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			requestPath,
			fmt.Sprintf("Attribute %[1]q must not be configured when %[2]q %[3]s", matchedPath, requestPath, v.when.String()),
		))
	}
	return diags
}
