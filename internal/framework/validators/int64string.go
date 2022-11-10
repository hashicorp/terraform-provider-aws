package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ tfsdk.AttributeValidator = int64StringBetweenValidator{}

// int64StringBetweenValidator validates that a string Attribute's value is an integer in a range.
type int64StringBetweenValidator struct {
	min, max int64
}

// Description describes the validation in plain text formatting.
func (validator int64StringBetweenValidator) Description(_ context.Context) string {
	return fmt.Sprintf("integer value must be between %d and %d", validator.min, validator.max)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator int64StringBetweenValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// Validate performs the validation.
func (validator int64StringBetweenValidator) Validate(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) {
	i, ok := validateInt64String(ctx, request, response)

	if !ok {
		return
	}

	if i < validator.min || i > validator.max {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.AttributePath,
			validator.Description(ctx),
			fmt.Sprintf("%d", i),
		))

		return
	}
}

// Int64StringBetween returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a 64-bit integer.
//   - Integer value is greater than or equal to the given minimum and less than or equal to the given maximum.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func Int64StringBetween(min, max int64) tfsdk.AttributeValidator {
	if min > max {
		return nil
	}

	return int64StringBetweenValidator{
		min: min,
		max: max,
	}
}
