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

var _ validator.String = prefixNoneOfCaseInsensitiveValidator{}

type prefixNoneOfCaseInsensitiveValidator struct {
	values     []string
	exceptions []string
}

func (v prefixNoneOfCaseInsensitiveValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v prefixNoneOfCaseInsensitiveValidator) MarkdownDescription(_ context.Context) string {
	desc := fmt.Sprintf("value must not begin with (case-insensitive): %s", tfslices.ApplyToAll(v.values, func(v string) string {
		return `"` + v + `"`
	}))
	if len(v.exceptions) > 0 {
		desc += fmt.Sprintf(", except: %s", tfslices.ApplyToAll(v.exceptions, func(v string) string {
			return `"` + v + `"`
		}))
	}
	return desc
}

func (v prefixNoneOfCaseInsensitiveValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := strings.ToLower(request.ConfigValue.ValueString())

	for _, exception := range v.exceptions {
		if strings.HasPrefix(value, strings.ToLower(exception)) {
			return
		}
	}

	for _, prefix := range v.values {
		if !strings.HasPrefix(value, strings.ToLower(prefix)) {
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

// PrefixNoneOfCaseInsensitive checks that the String held in the attribute
// begins with none of the given `values` using case-insensitive comparison,
// with optional exceptions for specific prefixes that should be allowed.
func PrefixNoneOfCaseInsensitive(values []string, exceptions []string) prefixNoneOfCaseInsensitiveValidator {
	return prefixNoneOfCaseInsensitiveValidator{
		values:     slices.Clone(values),
		exceptions: slices.Clone(exceptions),
	}
}
