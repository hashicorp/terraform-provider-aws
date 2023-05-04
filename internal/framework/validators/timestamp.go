package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/types/timestamp"
)

type utcTimestampValidator struct{}

func (validator utcTimestampValidator) Description(_ context.Context) string {
	return "value must be a valid UTC Timestamp"
}

func (validator utcTimestampValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator utcTimestampValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	t := timestamp.New(request.ConfigValue.ValueString())
	if err := t.ValidateUTCFormat(); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

func UTCTimestamp() validator.String {
	return utcTimestampValidator{}
}

type onceADayWindowFormatValidator struct{}

func (validator onceADayWindowFormatValidator) Description(_ context.Context) string {
	return `value must satisfy the format of "hh24:mi-hh24:mi"`
}

func (validator onceADayWindowFormatValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator onceADayWindowFormatValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	t := timestamp.New(request.ConfigValue.ValueString())
	if err := t.ValidateOnceADayWindowFormat(); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

func OnceADayWindowFormat() validator.String {
	return onceADayWindowFormatValidator{}
}

type onceAWeekWindowFormatValidator struct{}

func (validator onceAWeekWindowFormatValidator) Description(_ context.Context) string {
	return `value must satisfy the format of "ddd:hh24:mi-ddd:hh24:mi"`
}

func (validator onceAWeekWindowFormatValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator onceAWeekWindowFormatValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	t := timestamp.New(request.ConfigValue.ValueString())
	if err := t.ValidateOnceAWeekWindowFormat(); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

func OnceAWeekWindowFormat() validator.String {
	return onceAWeekWindowFormatValidator{}
}
