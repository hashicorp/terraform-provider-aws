// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

// principalARNValidator validates ARNs specifically for RAM principals.
// Only allows: organizations (organization/, ou/) and iam (role/, user/) ARNs.
type principalARNValidator struct{}

func (v principalARNValidator) Description(_ context.Context) string {
	return "value must be a valid RAM principal ARN (organization, OU, IAM role, or IAM user)"
}

func (v principalARNValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v principalARNValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if !arn.IsARN(value) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
		return
	}

	parsedARN, err := arn.Parse(value)
	if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
		return
	}

	// Validate it's an allowed ARN type for RAM principals
	switch parsedARN.Service {
	case "organizations":
		if strings.HasPrefix(parsedARN.Resource, "organization/") ||
			strings.HasPrefix(parsedARN.Resource, "ou/") {
			return
		}
	case "iam":
		if strings.HasPrefix(parsedARN.Resource, "role/") ||
			strings.HasPrefix(parsedARN.Resource, "user/") {
			return
		}
	}

	response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
		request.Path,
		v.Description(ctx),
		value,
	))
}

// principalValidator returns a string validator which ensures that any configured
// attribute value is a valid RAM principal:
//   - AWS account ID (exactly 12 digits) - uses fwvalidators.AWSAccountID()
//   - Organization ARN
//   - Organizational unit (OU) ARN
//   - IAM role ARN
//   - IAM user ARN
//   - Service principal name - uses fwvalidators.ServicePrincipal()
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func principalValidator() validator.String {
	return stringvalidator.Any(
		fwvalidators.AWSAccountID(),
		principalARNValidator{},
		fwvalidators.ServicePrincipal(),
	)
}
