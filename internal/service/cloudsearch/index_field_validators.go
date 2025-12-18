// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"context"
	"fmt"
	"slices"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
	// Must use fwtypes.StringEnum to match the schema's CustomType
	var fieldType fwtypes.StringEnum[awstypes.IndexFieldType]
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("type"), &fieldType)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() || fieldType.IsNull() || fieldType.IsUnknown() {
		return
	}

	// Check if this field type is invalid for this attribute
	// Convert enum to string for comparison with invalidFieldTypes
	if slices.Contains(v.invalidFieldTypes, string(fieldType.ValueEnum())) {
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

// analysisSchemeDefaultPlanModifier sets the default analysis_scheme for text fields
// AWS CloudSearch returns "_en_default_" for text/text-array fields when not configured,
// so we need to set this in the plan to avoid "does not correlate" errors.
type analysisSchemeDefaultPlanModifier struct{}

func (m analysisSchemeDefaultPlanModifier) Description(ctx context.Context) string {
	return "Sets default analysis_scheme for text fields"
}

func (m analysisSchemeDefaultPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m analysisSchemeDefaultPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the config value is set (not null), use it as-is
	if !req.ConfigValue.IsNull() {
		return
	}

	// Get the "type" attribute from the same index_field block
	var fieldType fwtypes.StringEnum[awstypes.IndexFieldType]
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("type"), &fieldType)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() || fieldType.IsNull() || fieldType.IsUnknown() {
		return
	}

	// For text and text-array fields with no config, use "_en_default_"
	// This ensures the plan matches what AWS returns, preventing "does not correlate" errors
	fieldTypeStr := string(fieldType.ValueEnum())
	if fieldTypeStr == "text" || fieldTypeStr == "text-array" {
		// Use state if available, otherwise set default
		if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
			resp.PlanValue = req.StateValue
		} else {
			resp.PlanValue = types.StringValue("_en_default_")
		}
		return
	}

	// For non-text fields, use state if available (UseStateForUnknown behavior)
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}

// AnalysisSchemeDefault returns a plan modifier that sets the default analysis_scheme for text fields
func AnalysisSchemeDefault() planmodifier.String {
	return analysisSchemeDefaultPlanModifier{}
}
