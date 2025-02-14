// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	switch validator.algorithm {
	case "md5":
		hashPattern = "^[a-fA-F0-9]{32}$"
	case "sha256":
		hashPattern = "^[a-fA-F0-9]{64}$"
	case "sha512":
		hashPattern = "^[a-fA-F0-9]{128}$"
	default:
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			validator.Description(ctx),
			"unsupported hash algorithm",
		))
		return
	}

	matched, err := regexp.MatchString(hashPattern, request.ConfigValue.ValueString())
	if err != nil {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			validator.Description(ctx),
			"error matching hash pattern: "+err.Error(),
		))
	}
	if !matched {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			validator.Description(ctx),
			"value must be a valid hash",
		))
	}
}

func Hash(algorithm string) validator.String {
	return hashValidator{
		algorithm: algorithm,
	}
}
