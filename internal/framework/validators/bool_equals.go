// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Bool = boolEqualsValidator{}

type boolEqualsValidator struct {
	value types.Bool
}

func (v boolEqualsValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Value must be %q", v.value)
}

func (v boolEqualsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v boolEqualsValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	configValue := req.ConfigValue

	if !configValue.Equal(v.value) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			req.Path,
			v.Description(ctx),
			configValue.String(),
		))
	}
}

// BoolEquals checks that the Bool held in the attribute matches the
// given `value`
func BoolEquals(value bool) validator.Bool {
	return boolEqualsValidator{
		value: types.BoolValue(value),
	}
}
