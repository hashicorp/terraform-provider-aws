// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// indexFieldAttributeValidator validates that an attribute is only set for compatible field types
type indexFieldAttributeValidator struct {
	attributeName     string   // Name of attribute being validated
	invalidFieldTypes []string // Field types where this attribute is invalid
}

func (v indexFieldAttributeValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("%s cannot be set for certain field types", v.attributeName)
}

func (v indexFieldAttributeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateBool implements validator.Bool
func (v indexFieldAttributeValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If attribute is null/unknown, nothing to validate
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Get the "type" attribute from the same index_field block
	var fieldType types.String
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("type"), &fieldType)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() || fieldType.IsNull() || fieldType.IsUnknown() {
		return
	}

	// Check if this field type is invalid for this attribute
	if slices.Contains(v.invalidFieldTypes, fieldType.ValueString()) {
		// Create attribute-specific reason message
		var reason string
		switch v.attributeName {
		case "facet":
			reason = "Text fields in AWS CloudSearch are always searchable and cannot be faceted."
		case "search":
			reason = "Text fields in AWS CloudSearch are always searchable."
		case "sort":
			reason = "AWS CloudSearch does not support sorting on array fields."
		case "highlight":
			reason = "AWS CloudSearch only supports highlighting for text and text-array fields."
		default:
			reason = "This attribute is not supported by AWS CloudSearch for this field type."
		}

		resp.Diagnostics.AddAttributeError(
			req.Path,
			fmt.Sprintf("Invalid Attribute for Field Type '%s'", fieldType.ValueString()),
			fmt.Sprintf(
				"The '%s' attribute cannot be set for '%s' field types, even when set to false. "+
					"%s Remove this attribute from the configuration.",
				v.attributeName,
				fieldType.ValueString(),
				reason,
			),
		)
	}
}

// invalidForFieldTypes creates a validator that errors if the attribute is set for the specified field types
func invalidForFieldTypes(attributeName string, fieldTypes ...string) indexFieldAttributeValidator {
	return indexFieldAttributeValidator{
		attributeName:     attributeName,
		invalidFieldTypes: fieldTypes,
	}
}
