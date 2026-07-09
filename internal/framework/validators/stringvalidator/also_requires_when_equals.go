// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

type alsoRequiresWhenEqualsValidator struct {
	internal.AlsoRequiresWhenEqualsValidator
}

func (v alsoRequiresWhenEqualsValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v alsoRequiresWhenEqualsValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Ensure that when this attribute equals %[1]q, the following are also configured: %[2]q", v.Value, v.PathExpressions)
}

func (v alsoRequiresWhenEqualsValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	validateRequest := internal.AlsoRequiresWhenEqualsValidatorRequest{
		Config:         request.Config,
		ConfigValue:    request.ConfigValue,
		Path:           request.Path,
		PathExpression: request.PathExpression,
	}
	var validateResponse internal.AlsoRequiresWhenEqualsValidatorResponse

	v.Validate(ctx, validateRequest, &validateResponse)

	response.Diagnostics.Append(validateResponse.Diagnostics...)
}

func AlsoRequiresWhenEquals[T ~string](value T, expressions ...path.Expression) validator.String {
	return alsoRequiresWhenEqualsValidator{
		internal.AlsoRequiresWhenEqualsValidator{
			Value:           types.StringValue(string(value)),
			PathExpressions: expressions,
		},
	}
}
