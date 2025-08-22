// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"errors"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ipv4AddressValidator validates that a string Attribute's value is a valid IPv4 address.
type ipv4AddressValidator struct{}

// Description describes the validation in plain text formatting.
func (validator ipv4AddressValidator) Description(_ context.Context) string {
	return "value must be a valid IPv4 address"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator ipv4AddressValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// Validate performs the validation.
func (validator ipv4AddressValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := validateIPv4Address(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

// IPv4Address returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid IPv4 address.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IPv4Address() validator.String {
	return ipv4AddressValidator{}
}

func validateIPv4Address(value string) error {
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return err
	}

	if !addr.Is4() {
		return errors.New("invalid IPv4 address")
	}
	return nil
}

// ipv6AddressValidator validates that a string Attribute's value is a valid IPv6 address.
type ipv6AddressValidator struct{}

// Description describes the validation in plain text formatting.
func (validator ipv6AddressValidator) Description(_ context.Context) string {
	return "value must be a valid IPv6 address"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator ipv6AddressValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// Validate performs the validation.
func (validator ipv6AddressValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := validateIPv6Address(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))

		return
	}
}

// IPv6Address returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid IPv6 address.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IPv6Address() validator.String {
	return ipv6AddressValidator{}
}

func validateIPv6Address(value string) error {
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return err
	}

	if !addr.Is6() {
		return errors.New("invalid IPv6 address")
	}
	return nil
}
