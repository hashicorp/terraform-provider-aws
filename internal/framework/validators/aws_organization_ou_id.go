// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// awsOrganizationOUIDValidator validates that a string
type awsOrganizationOUIDValidator struct{}

// Description describes the validation in plain text formatting.
func (validator awsOrganizationOUIDValidator) Description(_ context.Context) string {
	return "value must be a valid AWS organizational unit ID"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator awsOrganizationOUIDValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// ValidateString performs the validation
func (validator awsOrganizationOUIDValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_OrganizationalUnit.html
	if !itypes.IsAWSOrganizationOUID(request.ConfigValue.ValueString()) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

// AWSOrganizationOUID returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid AWS Organizational Unit ID.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func AWSOrganizationOUID() validator.String { // nosemgrep:ci.aws-in-func-name
	return awsOrganizationOUIDValidator{}
}
