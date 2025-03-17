// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

var _ validator.String = suffixNoneOfValidator{}

type suffixNoneOfValidator struct {
	values []string
}

func (v suffixNoneOfValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v suffixNoneOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value must end with none of: %s", tfslices.ApplyToAll(v.values, func(v string) string {
		return `"` + v + `"`
	}))
}

func (v suffixNoneOfValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	for _, otherValue := range v.values {
		if !strings.HasSuffix(value, otherValue) {
			continue
		}

		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			request.ConfigValue.String(),
		))

		break
	}
}

// SuffixNoneOf checks that the String held in the attribute
// ends with none of the given `values`.
func SuffixNoneOf(values ...string) suffixNoneOfValidator {
	return suffixNoneOfValidator{
		values: slices.Clone(values),
	}
}
