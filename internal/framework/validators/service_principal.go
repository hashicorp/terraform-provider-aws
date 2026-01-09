// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// servicePrincipalValidator validates that a string Attribute's value is a valid AWS service principal.
type servicePrincipalValidator struct{}

// Description describes the validation in plain text formatting.
func (validator servicePrincipalValidator) Description(_ context.Context) string {
	return "value must be a valid AWS service principal (e.g., ec2.amazonaws.com)"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator servicePrincipalValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// ValidateString performs the validation.
func (validator servicePrincipalValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if !inttypes.IsServicePrincipal(request.ConfigValue.ValueString()) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

// ServicePrincipal returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid AWS service principal (e.g., ec2.amazonaws.com).
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func ServicePrincipal() validator.String {
	return servicePrincipalValidator{}
}
