// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/YakDriver/regexache"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type hashValidator struct {
	algorithm string
}

func (validator hashValidator) Description(_ context.Context) string {
	return "value must be a valid hash"
}

func (validator hashValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator hashValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	var hashPattern string
	switch algo := validator.algorithm; algo {
	case "md5":
		hashPattern = `^[a-fA-F0-9]{32}$`
	case "sha256":
		hashPattern = `^[a-fA-F0-9]{64}$`
	case "sha512":
		hashPattern = `^[a-fA-F0-9]{128}$`
	default:
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			"unsupported hash algorithm",
			algo,
		))
		return
	}

	if !regexache.MustCompile(hashPattern).MatchString(request.ConfigValue.ValueString()) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

func Hash(algorithm string) validator.String {
	return hashValidator{
		algorithm: algorithm,
	}
}
