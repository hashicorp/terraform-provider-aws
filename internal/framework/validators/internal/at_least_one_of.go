// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Object = (*atLeastOneOfWhenValidator)(nil)
)

func AtLeastOneOfValidator(diagFactory invalidAttributeCombinationDiagnosticFunc, expressions ...path.Expression) atLeastOneOfWhenValidator {
	return atLeastOneOfWhenValidator{
		conditionalMatchedPathCountValidator{
			when:            always{},
			pathExpressions: expressions,
		},
		diagFactory,
	}
}

type atLeastOneOfWhenValidator struct {
	conditionalMatchedPathCountValidator
	invalidAttributeCombinationDiagnosticFunc
}

func (v atLeastOneOfWhenValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v atLeastOneOfWhenValidator) MarkdownDescription(context.Context) string {
	if v.when.String() == "" {
		return fmt.Sprintf("Ensure that at most least attribute from this collection is configured: %[1]q", v.pathExpressions)
	}
	return fmt.Sprintf("Ensure that when this attribute value matches the condition, at least one attribute from this collection is configured: %[1]q", v.pathExpressions)
}

func (v atLeastOneOfWhenValidator) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
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

func (v atLeastOneOfWhenValidator) validate(ctx context.Context, request validatorRequest, response *validatorResponse) {
	v.conditionalMatchedPathCountValidator.validate(ctx, request, response, v.eval)
}

func (v atLeastOneOfWhenValidator) eval(_ context.Context, requestPath path.Path, expressions path.Expressions, count int) diag.Diagnostics {
	var (
		diags       diag.Diagnostics
		description string
	)

	switch when := v.when.String(); {
	case count == 0:
		if when == "" {
			description = fmt.Sprintf("At least one of %[1]s must be configured. No attribute configured.", expressions)
		} else {
			description = fmt.Sprintf("At least one of %[1]s must be configured when %[2]s %[3]s. No attribute configured.", expressions, requestPath, when)
		}
	}

	if description != "" {
		diags.Append(v.invalidAttributeCombinationDiagnosticFunc(requestPath, description))
	}

	return diags
}
