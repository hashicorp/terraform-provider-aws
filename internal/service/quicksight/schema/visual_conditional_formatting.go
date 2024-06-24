// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/names"
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
													"color": stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
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
							names.AttrExpression: stringSchema(true, validation.StringLenBetween(1, 4096)),
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
							"color":              stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
							names.AttrExpression: stringSchema(true, validation.StringLenBetween(1, 4096)),
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
							"color":              stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
							names.AttrExpression: stringSchema(true, validation.StringLenBetween(1, 4096)),
							"icon_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingCustomIconOptions.html
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"icon":         stringSchema(false, validation.StringInSlice(quicksight.Icon_Values(), false)),
										"unicode_icon": stringSchema(false, validation.StringMatch(regexache.MustCompile(`^[^\\u0000-\\u00FF]$`), "")),
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
										"icon_display_option": stringSchema(false, validation.StringInSlice(quicksight.ConditionalFormattingIconDisplayOption_Values(), false))},
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
							names.AttrExpression: stringSchema(true, validation.StringLenBetween(1, 4096)),
							"icon_set_type":      stringSchema(false, validation.StringInSlice(quicksight.ConditionalFormattingIconSetType_Values(), false)),
						},
					},
				},
			},
		},
	}
}

func expandConditionalFormattingColor(tfList []interface{}) *quicksight.ConditionalFormattingColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	color := &quicksight.ConditionalFormattingColor{}

	if v, ok := tfMap["gradient"].([]interface{}); ok && len(v) > 0 {
		color.Gradient = expandConditionalFormattingGradientColor(v)
	}
	if v, ok := tfMap["solid"].([]interface{}); ok && len(v) > 0 {
		color.Solid = expandConditionalFormattingSolidColor(v)
	}

	return color
}

func expandConditionalFormattingGradientColor(tfList []interface{}) *quicksight.ConditionalFormattingGradientColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	color := &quicksight.ConditionalFormattingGradientColor{}

	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		color.Expression = aws.String(v)
	}
	if v, ok := tfMap["color"].([]interface{}); ok && len(v) > 0 {
		color.Color = expandGradientColor(v)
	}

	return color
}

func expandGradientColor(tfList []interface{}) *quicksight.GradientColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	color := &quicksight.GradientColor{}

	if v, ok := tfMap["stops"].([]interface{}); ok && len(v) > 0 {
		color.Stops = expandGradientStops(v)
	}

	return color
}

func expandGradientStops(tfList []interface{}) []*quicksight.GradientStop {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.GradientStop
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandGradientStop(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandGradientStop(tfMap map[string]interface{}) *quicksight.GradientStop {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.GradientStop{}

	if v, ok := tfMap["gradient_offset"].(float64); ok {
		options.GradientOffset = aws.Float64(v)
	}
	if v, ok := tfMap["color"].(string); ok && v != "" {
		options.Color = aws.String(v)
	}
	if v, ok := tfMap["data_value"].(float64); ok {
		options.DataValue = aws.Float64(v)
	}

	return options
}

func expandConditionalFormattingSolidColor(tfList []interface{}) *quicksight.ConditionalFormattingSolidColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	color := &quicksight.ConditionalFormattingSolidColor{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		color.Color = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		color.Expression = aws.String(v)
	}

	return color
}

func expandConditionalFormattingIcon(tfList []interface{}) *quicksight.ConditionalFormattingIcon {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	icon := &quicksight.ConditionalFormattingIcon{}

	if v, ok := tfMap["custom_condition"].([]interface{}); ok && len(v) > 0 {
		icon.CustomCondition = expandConditionalFormattingCustomIconCondition(v)
	}
	if v, ok := tfMap["icon_set"].([]interface{}); ok && len(v) > 0 {
		icon.IconSet = expandConditionalFormattingIconSet(v)
	}

	return icon
}

func expandConditionalFormattingCustomIconCondition(tfList []interface{}) *quicksight.ConditionalFormattingCustomIconCondition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	icon := &quicksight.ConditionalFormattingCustomIconCondition{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		icon.Color = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
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

func expandConditionalFormattingCustomIconOptions(tfList []interface{}) *quicksight.ConditionalFormattingCustomIconOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ConditionalFormattingCustomIconOptions{}

	if v, ok := tfMap["icon"].(string); ok && v != "" {
		options.Icon = aws.String(v)
	}
	if v, ok := tfMap["unicode_icon"].(string); ok && v != "" {
		options.UnicodeIcon = aws.String(v)
	}

	return options
}

func expandConditionalFormattingIconDisplayConfiguration(tfList []interface{}) *quicksight.ConditionalFormattingIconDisplayConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ConditionalFormattingIconDisplayConfiguration{}

	if v, ok := tfMap["icon_display_option"].(string); ok && v != "" {
		config.IconDisplayOption = aws.String(v)
	}

	return config
}

func expandConditionalFormattingIconSet(tfList []interface{}) *quicksight.ConditionalFormattingIconSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ConditionalFormattingIconSet{}

	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		options.Expression = aws.String(v)
	}
	if v, ok := tfMap["icon_set_type"].(string); ok && v != "" {
		options.IconSetType = aws.String(v)
	}

	return options
}

func expandTextConditionalFormat(tfList []interface{}) *quicksight.TextConditionalFormat {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TextConditionalFormat{}

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

func flattenConditionalFormattingColor(apiObject *quicksight.ConditionalFormattingColor) []interface{} {
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

func flattenConditionalFormattingGradientColor(apiObject *quicksight.ConditionalFormattingGradientColor) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = flattenGradientColor(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.StringValue(apiObject.Expression)
	}

	return []interface{}{tfMap}
}

func flattenGradientColor(apiObject *quicksight.GradientColor) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Stops != nil {
		tfMap["stops"] = flattenGradientStop(apiObject.Stops)
	}

	return []interface{}{tfMap}
}

func flattenGradientStop(apiObject []*quicksight.GradientStop) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.GradientOffset != nil {
			tfMap["gradient_offset"] = aws.Float64Value(config.GradientOffset)
		}
		if config.Color != nil {
			tfMap["color"] = aws.StringValue(config.Color)
		}
		if config.DataValue != nil {
			tfMap["data_value"] = aws.Float64Value(config.DataValue)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenConditionalFormattingSolidColor(apiObject *quicksight.ConditionalFormattingSolidColor) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.StringValue(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.StringValue(apiObject.Expression)
	}

	return []interface{}{tfMap}
}

func flattenConditionalFormattingIcon(apiObject *quicksight.ConditionalFormattingIcon) []interface{} {
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

func flattenConditionalFormattingCustomIconCondition(apiObject *quicksight.ConditionalFormattingCustomIconCondition) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.StringValue(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.StringValue(apiObject.Expression)
	}
	if apiObject.IconOptions != nil {
		tfMap["icon_options"] = flattenConditionalFormattingCustomIconOptions(apiObject.IconOptions)
	}
	if apiObject.DisplayConfiguration != nil {
		tfMap["display_configuration"] = flattenConditionalFormattingIconDisplayConfiguration(apiObject.DisplayConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenConditionalFormattingCustomIconOptions(apiObject *quicksight.ConditionalFormattingCustomIconOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Icon != nil {
		tfMap["icon"] = aws.StringValue(apiObject.Icon)
	}
	if apiObject.UnicodeIcon != nil {
		tfMap["unicode_icon"] = aws.StringValue(apiObject.UnicodeIcon)
	}

	return []interface{}{tfMap}
}

func flattenConditionalFormattingIconDisplayConfiguration(apiObject *quicksight.ConditionalFormattingIconDisplayConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.IconDisplayOption != nil {
		tfMap["icon_display_option"] = aws.StringValue(apiObject.IconDisplayOption)
	}

	return []interface{}{tfMap}
}

func flattenConditionalFormattingIconSet(apiObject *quicksight.ConditionalFormattingIconSet) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.StringValue(apiObject.Expression)
	}
	if apiObject.IconSetType != nil {
		tfMap["icon_set_type"] = aws.StringValue(apiObject.IconSetType)
	}

	return []interface{}{tfMap}
}
