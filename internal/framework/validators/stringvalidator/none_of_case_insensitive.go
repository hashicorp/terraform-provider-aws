// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

var _ validator.String = noneOfCaseInsensitiveValidator{}

type noneOfCaseInsensitiveValidator struct {
	values []string
}

func (v noneOfCaseInsensitiveValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v noneOfCaseInsensitiveValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value must be none of (case-insensitive): %s", tfslices.ApplyToAll(v.values, func(v string) string {
		return `"` + v + `"`
	}))
}

func (v noneOfCaseInsensitiveValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	for _, otherValue := range v.values {
		if !strings.EqualFold(value, otherValue) {
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

// NoneOfCaseInsensitive checks that the String held in the attribute
// is none of the given `values` using case-insensitive comparison.
func NoneOfCaseInsensitive(values ...string) noneOfCaseInsensitiveValidator {
	return noneOfCaseInsensitiveValidator{
		values: slices.Clone(values),
	}
}
