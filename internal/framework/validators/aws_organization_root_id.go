// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// awsOrganizationRootIDValidator validates that a string
type awsOrganizationRootIDValidator struct{}

// Description describes the validation in plain text formatting.
func (validator awsOrganizationRootIDValidator) Description(_ context.Context) string {
	return "value must be a valid AWS organization root ID"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator awsOrganizationRootIDValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// ValidateString performs the validation
func (validator awsOrganizationRootIDValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_Root.html
	if !itypes.IsAWSOrganizationRootID(request.ConfigValue.ValueString()) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

// AWSOrganizationRootID returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid AWS Organization Root ID.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func AWSOrganizationRootID() validator.String { // nosemgrep:ci.aws-in-func-name
	return awsOrganizationRootIDValidator{}
}
