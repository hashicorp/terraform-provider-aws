// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// ipv4CIDRNetworkAddressValidator validates that a string Attribute's value is a valid IPv4 CIDR that represents a network address.
type ipv4CIDRNetworkAddressValidator struct{}

// Description describes the validation in plain text formatting.
func (validator ipv4CIDRNetworkAddressValidator) Description(_ context.Context) string {
	return "value must be a valid IPv4 CIDR that represents a network address"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator ipv4CIDRNetworkAddressValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// Validate performs the validation.
func (validator ipv4CIDRNetworkAddressValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := verify.ValidateIPv4CIDRBlock(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

// IPv4CIDRNetworkAddress returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid IPv4 CIDR network address.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IPv4CIDRNetworkAddress() validator.String {
	return ipv4CIDRNetworkAddressValidator{}
}

// ipv6CIDRNetworkAddressValidator validates that a string Attribute's value is a valid IPv6 CIDR that represents a network address.
type ipv6CIDRNetworkAddressValidator struct{}

// Description describes the validation in plain text formatting.
func (validator ipv6CIDRNetworkAddressValidator) Description(_ context.Context) string {
	return "value must be a valid IPv6 CIDR that represents a network address"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator ipv6CIDRNetworkAddressValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// Validate performs the validation.
func (validator ipv6CIDRNetworkAddressValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := verify.ValidateIPv6CIDRBlock(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))

		return
	}
}

// IPv6CIDRNetworkAddress returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid IPv6 CIDR network address.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IPv6CIDRNetworkAddress() validator.String {
	return ipv6CIDRNetworkAddressValidator{}
}
