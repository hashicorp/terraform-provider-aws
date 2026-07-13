// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func DiscriminatorRequires[T ~string](mapping map[T]path.Expression) validator.String {
	return discriminatorRequires[T]{mapping: mapping}
}

type discriminatorRequires[T ~string] struct {
	mapping map[T]path.Expression
}

func (v discriminatorRequires[T]) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v discriminatorRequires[T]) MarkdownDescription(ctx context.Context) string {
	return "Ensure that when this attribute value matches the condition, the corresponding mapped attribute is also configured"
}

func (v discriminatorRequires[T]) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	vDiscriminating := T(request.ConfigValue.ValueString())
	expression, ok := v.mapping[vDiscriminating]
	if !ok {
		return
	}

	matchedPaths, diags := request.Config.PathMatches(ctx, expression)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	for _, mp := range matchedPaths {
		var mpVal attr.Value
		response.Diagnostics.Append(request.Config.GetAttribute(ctx, mp, &mpVal)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Defer if any target is unknown; we cannot decide yet.
		if mpVal.IsUnknown() {
			return
		}

		if mpVal.IsNull() {
			response.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
				request.Path,
				fmt.Sprintf("Attribute %[1]q must be configured when %[2]q equals %[3]q", mp, request.Path, vDiscriminating),
			))
		}
	}
}
