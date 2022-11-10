package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
func (validator ipv4CIDRNetworkAddressValidator) Validate(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) {
	s, ok := validateString(ctx, request, response)

	if !ok {
		return
	}

	if err := verify.ValidateIPv4CIDRBlock(s); err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.AttributePath,
			validator.Description(ctx),
			err.Error(),
		))

		return
	}
}

// IPv4CIDRNetworkAddress returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid IPv4 CIDR network address.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IPv4CIDRNetworkAddress() tfsdk.AttributeValidator {
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
func (validator ipv6CIDRNetworkAddressValidator) Validate(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) {
	s, ok := validateString(ctx, request, response)

	if !ok {
		return
	}

	if err := verify.ValidateIPv6CIDRBlock(s); err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.AttributePath,
			validator.Description(ctx),
			err.Error(),
		))

		return
	}
}

// IPv6CIDRNetworkAddress returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid IPv6 CIDR network address.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func IPv6CIDRNetworkAddress() tfsdk.AttributeValidator {
	return ipv6CIDRNetworkAddressValidator{}
}
