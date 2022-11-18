package validators

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validateInt64String ensures that the request contains a String value which represents a 64-bit integer.
func validateInt64String(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) (int64, bool) {
	s, ok := validateString(ctx, request, response)

	if !ok {
		return 0, false
	}

	i, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.AttributePath,
			"Invalid Attribute Value Format",
			err.Error()))
		return 0, false
	}

	return i, true
}

// validateString ensures that the request contains a String value.
func validateString(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) (string, bool) {
	t := request.AttributeConfig.Type(ctx)
	if t != types.StringType {
		response.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			request.AttributePath,
			"expected value of type string",
			t.String(),
		))
		return "", false
	}

	s := request.AttributeConfig.(types.String)

	if s.IsUnknown() || s.IsNull() {
		return "", false
	}

	return s.ValueString(), true
}
