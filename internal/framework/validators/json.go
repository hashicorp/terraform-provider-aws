// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// jsonValidator validates that a string Attribute's value is valid JSON.
type jsonValidator struct{}

// Description describes the validation in plain text formatting.
func (validator jsonValidator) Description(_ context.Context) string {
	return "value must be valid JSON"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator jsonValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// Validate performs the validation.
func (validator jsonValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	configValue := request.ConfigValue

	if configValue.IsNull() || configValue.IsUnknown() {
		return
	}

	// https://datatracker.ietf.org/doc/html/rfc7159.
	if valueString := configValue.ValueString(); !json.Valid([]byte(valueString)) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			valueString,
		))
		return
	}
}

// JSON returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents valid JSON.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func JSON() validator.String {
	return jsonValidator{}
}
