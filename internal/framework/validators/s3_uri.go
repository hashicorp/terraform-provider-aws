// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// s3URIValidator validates that a string Attribute's value is a valid S3 URI.
type s3URIValidator struct{}

func (validator s3URIValidator) Description(_ context.Context) string {
	return "value must be a valid S3 URI"
}

func (validator s3URIValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (validator s3URIValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if !regexache.MustCompile(`^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$`).MatchString(request.ConfigValue.ValueString()) {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			validator.Description(ctx),
			request.ConfigValue.ValueString(),
		))
		return
	}
}

// S3URI returns a string validator which ensures that any configured
// attribute value:
//
//   - Is a string, which represents a valid S3 URI (s3://bucket[/key]).
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func S3URI() validator.String {
	return s3URIValidator{}
}
