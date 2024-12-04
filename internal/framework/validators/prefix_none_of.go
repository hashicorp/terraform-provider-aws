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

var _ validator.String = prefixNoneOfValidator{}

type prefixNoneOfValidator struct {
	values []string
}

func (v prefixNoneOfValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v prefixNoneOfValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value must begin with none of: %s", tfslices.ApplyToAll(v.values, func(v string) string {
		return `"` + v + `"`
	}))
}

func (v prefixNoneOfValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	for _, otherValue := range v.values {
		if !strings.HasPrefix(value, otherValue) {
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

// PrefixNoneOf checks that the String held in the attribute
// begins with none of the given `values`.
func PrefixNoneOf(values ...string) prefixNoneOfValidator {
	return prefixNoneOfValidator{
		values: slices.Clone(values),
	}
}
