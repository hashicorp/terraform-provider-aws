package validators

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

	if err := validateUTCTimestamp(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			validator.Description(ctx),
			err.Error(),
		))
		return
	}
}

func UTCTimestamp() validator.String {
	return utcTimestampValidator{}
}

func validateUTCTimestamp(value string) error {
	_, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return fmt.Errorf("must be in RFC3339 time format %q. Example: %s", time.RFC3339, err)
	}

	return nil
}
