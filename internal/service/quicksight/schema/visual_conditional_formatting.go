// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

func conditionalFormattingColorSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingColor.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"gradient": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingGradientColor.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GradientColor.html
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"stops": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GradientStop.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"gradient_offset": {
														Type:     schema.TypeFloat,
														Required: true,
													},
													"color": stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
													"data_value": {
														Type:     schema.TypeFloat,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"expression": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 4096))),
						},
					},
				},
				"solid": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingSolidColor.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":      stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
							"expression": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 4096))),
						},
					},
				},
			},
		},
	}
}

func conditionalFormattingIconSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingIcon.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_condition": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingCustomIconCondition.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":      stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
							"expression": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 4096))),
							"icon_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingCustomIconOptions.html
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"icon":         stringSchema(false, enum.Validate[types.Icon]()),
										"unicode_icon": stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[^\\u0000-\\u00FF]$`), ""))),
									},
								},
							},
							"display_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingIconDisplayConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"icon_display_option": stringSchema(false, enum.Validate[types.ConditionalFormattingIconDisplayOption]()),
									},
								},
							},
						},
					},
				},
				"icon_set": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingIconSet.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"expression":    stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 4096))),
							"icon_set_type": stringSchema(false, enum.Validate[types.ConditionalFormattingIconSetType]()),
						},
					},
				},
			},
		},
	}
}

func expandConditionalFormattingColor(tfList []interface{}) *types.ConditionalFormattingColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	color := &types.ConditionalFormattingColor{}

	if v, ok := tfMap["gradient"].([]interface{}); ok && len(v) > 0 {
		color.Gradient = expandConditionalFormattingGradientColor(v)
	}
	if v, ok := tfMap["solid"].([]interface{}); ok && len(v) > 0 {
		color.Solid = expandConditionalFormattingSolidColor(v)
	}

	return color
}

func expandConditionalFormattingGradientColor(tfList []interface{}) *types.ConditionalFormattingGradientColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	color := &types.ConditionalFormattingGradientColor{}

	if v, ok := tfMap["expression"].(string); ok && v != "" {
		color.Expression = aws.String(v)
	}
	if v, ok := tfMap["color"].([]interface{}); ok && len(v) > 0 {
		color.Color = expandGradientColor(v)
	}

	return color
}

func expandGradientColor(tfList []interface{}) *types.GradientColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	color := &types.GradientColor{}

	if v, ok := tfMap["stops"].([]interface{}); ok && len(v) > 0 {
		color.Stops = expandGradientStops(v)
	}

	return color
}

func expandGradientStops(tfList []interface{}) []types.GradientStop {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.GradientStop
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandGradientStop(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandGradientStop(tfMap map[string]interface{}) *types.GradientStop {
	if tfMap == nil {
		return nil
	}

	options := &types.GradientStop{}

	if v, ok := tfMap["gradient_offset"].(float64); ok {
		options.GradientOffset = v
	}
	if v, ok := tfMap["color"].(string); ok && v != "" {
		options.Color = aws.String(v)
	}
	if v, ok := tfMap["data_value"].(float64); ok {
		options.DataValue = aws.Float64(v)
	}

	return options
}

func expandConditionalFormattingSolidColor(tfList []interface{}) *types.ConditionalFormattingSolidColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	color := &types.ConditionalFormattingSolidColor{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		color.Color = aws.String(v)
	}
	if v, ok := tfMap["expression"].(string); ok && v != "" {
		color.Expression = aws.String(v)
	}

	return color
}

func expandConditionalFormattingIcon(tfList []interface{}) *types.ConditionalFormattingIcon {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	icon := &types.ConditionalFormattingIcon{}

	if v, ok := tfMap["custom_condition"].([]interface{}); ok && len(v) > 0 {
		icon.CustomCondition = expandConditionalFormattingCustomIconCondition(v)
	}
	if v, ok := tfMap["icon_set"].([]interface{}); ok && len(v) > 0 {
		icon.IconSet = expandConditionalFormattingIconSet(v)
	}

	return icon
}

func expandConditionalFormattingCustomIconCondition(tfList []interface{}) *types.ConditionalFormattingCustomIconCondition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	icon := &types.ConditionalFormattingCustomIconCondition{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		icon.Color = aws.String(v)
	}
	if v, ok := tfMap["expression"].(string); ok && v != "" {
		icon.Expression = aws.String(v)
	}
	if v, ok := tfMap["icon_options"].([]interface{}); ok && len(v) > 0 {
		icon.IconOptions = expandConditionalFormattingCustomIconOptions(v)
	}
	if v, ok := tfMap["display_configuration"].([]interface{}); ok && len(v) > 0 {
		icon.DisplayConfiguration = expandConditionalFormattingIconDisplayConfiguration(v)
	}

	return icon
}

func expandConditionalFormattingCustomIconOptions(tfList []interface{}) *types.ConditionalFormattingCustomIconOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.ConditionalFormattingCustomIconOptions{}

	if v, ok := tfMap["icon"].(string); ok && v != "" {
		options.Icon = types.Icon(v)
	}
	if v, ok := tfMap["unicode_icon"].(string); ok && v != "" {
		options.UnicodeIcon = aws.String(v)
	}

	return options
}

func expandConditionalFormattingIconDisplayConfiguration(tfList []interface{}) *types.ConditionalFormattingIconDisplayConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.ConditionalFormattingIconDisplayConfiguration{}

	if v, ok := tfMap["icon_display_option"].(string); ok && v != "" {
		config.IconDisplayOption = types.ConditionalFormattingIconDisplayOption(v)
	}

	return config
}

func expandConditionalFormattingIconSet(tfList []interface{}) *types.ConditionalFormattingIconSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.ConditionalFormattingIconSet{}

	if v, ok := tfMap["expression"].(string); ok && v != "" {
		options.Expression = aws.String(v)
	}
	if v, ok := tfMap["icon_set_type"].(string); ok && v != "" {
		options.IconSetType = types.ConditionalFormattingIconSetType(v)
	}

	return options
}

func expandTextConditionalFormat(tfList []interface{}) *types.TextConditionalFormat {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TextConditionalFormat{}

	if v, ok := tfMap["background_color"].([]interface{}); ok && len(v) > 0 {
		options.BackgroundColor = expandConditionalFormattingColor(v)
	}
	if v, ok := tfMap["icon"].([]interface{}); ok && len(v) > 0 {
		options.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		options.TextColor = expandConditionalFormattingColor(v)
	}

	return options
}

func flattenConditionalFormattingColor(apiObject *types.ConditionalFormattingColor) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Gradient != nil {
		tfMap["gradient"] = flattenConditionalFormattingGradientColor(apiObject.Gradient)
	}
	if apiObject.Solid != nil {
		tfMap["solid"] = flattenConditionalFormattingSolidColor(apiObject.Solid)
	}

	return []interface{}{tfMap}
}

func flattenConditionalFormattingGradientColor(apiObject *types.ConditionalFormattingGradientColor) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = flattenGradientColor(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap["expression"] = aws.ToString(apiObject.Expression)
	}

	return []interface{}{tfMap}
}

func flattenGradientColor(apiObject *types.GradientColor) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Stops != nil {
		tfMap["stops"] = flattenGradientStop(apiObject.Stops)
	}

	return []interface{}{tfMap}
}

func flattenGradientStop(apiObject []types.GradientStop) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}

		tfMap["gradient_offset"] = config.GradientOffset

		tfMap["color"] = aws.ToString(config.Color)

		tfMap["data_value"] = aws.ToFloat64(config.DataValue)

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenConditionalFormattingSolidColor(apiObject *types.ConditionalFormattingSolidColor) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap["expression"] = aws.ToString(apiObject.Expression)
	}

	return []interface{}{tfMap}
}

func flattenConditionalFormattingIcon(apiObject *types.ConditionalFormattingIcon) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomCondition != nil {
		tfMap["custom_condition"] = flattenConditionalFormattingCustomIconCondition(apiObject.CustomCondition)
	}
	if apiObject.IconSet != nil {
		tfMap["icon_set"] = flattenConditionalFormattingIconSet(apiObject.IconSet)
	}

	return []interface{}{tfMap}
}

func flattenConditionalFormattingCustomIconCondition(apiObject *types.ConditionalFormattingCustomIconCondition) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap["expression"] = aws.ToString(apiObject.Expression)
	}
	if apiObject.IconOptions != nil {
		tfMap["icon_options"] = flattenConditionalFormattingCustomIconOptions(apiObject.IconOptions)
	}
	if apiObject.DisplayConfiguration != nil {
		tfMap["display_configuration"] = flattenConditionalFormattingIconDisplayConfiguration(apiObject.DisplayConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenConditionalFormattingCustomIconOptions(apiObject *types.ConditionalFormattingCustomIconOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["icon"] = types.Icon(apiObject.Icon)

	tfMap["unicode_icon"] = aws.ToString(apiObject.UnicodeIcon)

	return []interface{}{tfMap}
}

func flattenConditionalFormattingIconDisplayConfiguration(apiObject *types.ConditionalFormattingIconDisplayConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["icon_display_option"] = types.ConditionalFormattingIconDisplayOption(apiObject.IconDisplayOption)

	return []interface{}{tfMap}
}

func flattenConditionalFormattingIconSet(apiObject *types.ConditionalFormattingIconSet) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Expression != nil {
		tfMap["expression"] = aws.ToString(apiObject.Expression)
	}

	tfMap["icon_set_type"] = types.ConditionalFormattingIconSetType(apiObject.IconSetType)

	return []interface{}{tfMap}
}
