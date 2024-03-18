// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// awsAccountIDValidator validates that a string Attribute's value is a valid AWS account ID.
type awsAccountIDValidator struct{}

// Description describes the validation in plain text formatting.
func (validator awsAccountIDValidator) Description(_ context.Context) string {
	return "value must be a valid AWS account ID"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator awsAccountIDValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// ValidateString performs the validation.
func (validator awsAccountIDValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// https://docs.aws.amazon.com/accounts/latest/reference/manage-acct-identifiers.html.
	if !itypes.IsAWSAccountID(request.ConfigValue.ValueString()) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

// AWSAccountID returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid AWS account ID.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func AWSAccountID() validator.String { // nosemgrep:ci.aws-in-func-name
	return awsAccountIDValidator{}
}
