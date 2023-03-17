package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

type arnValidator struct{}

func (validator arnValidator) Description(_ context.Context) string {
	return "value must be a valid ARN"
}

func (validator arnValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator arnValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if errs := verify.ValidateARN(request.ConfigValue.ValueString()); errs != nil {
		for _, v := range errs {
			response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				request.Path,
				validator.Description(ctx),
				v.Error(),
			))
		}
		return
	}
}

func ARN() validator.String {
	return arnValidator{}
}
