// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var validPolicyPathFramework = stringvalidator.All([]validator.String{
	stringvalidator.LengthBetween(1, 512),
	pathBeginsAndEndsWithSlashValidator{},
	pathContainsNoConsecutiveSlashesValidator{},
	stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9\.,\+@=_/-]*$`), "value must contain uppercase or lowercase alphanumeric characters or any of the following: / , . + @ = _ -"),
}...)

var _ validator.String = pathBeginsAndEndsWithSlashValidator{}

type pathBeginsAndEndsWithSlashValidator struct{}

func (v pathBeginsAndEndsWithSlashValidator) Description(_ context.Context) string {
	return "value must begin and end with a slash (/)"
}

func (v pathBeginsAndEndsWithSlashValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v pathBeginsAndEndsWithSlashValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	// Handled by LengthBetween validator
	if value == "" {
		return
	}

	if !strings.HasPrefix(value, "/") || !strings.HasSuffix(value, "/") {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}

var _ validator.String = pathContainsNoConsecutiveSlashesValidator{}

type pathContainsNoConsecutiveSlashesValidator struct{}

func (v pathContainsNoConsecutiveSlashesValidator) Description(_ context.Context) string {
	return "value must not contain consecutive slashes (//)"
}

func (v pathContainsNoConsecutiveSlashesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v pathContainsNoConsecutiveSlashesValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if strings.Contains(value, "//") {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}
