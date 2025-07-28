// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var conditionalFormattingColorSchema = sync.OnceValue(func() *schema.Schema {
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
													"color": hexColorSchema(attrOptional),
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
							names.AttrExpression: stringLenBetweenSchema(attrRequired, 1, 4096),
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
							"color":              hexColorSchema(attrOptional),
							names.AttrExpression: stringLenBetweenSchema(attrRequired, 1, 4096),
						},
					},
				},
			},
		},
	}
})

var conditionalFormattingIconSchema = sync.OnceValue(func() *schema.Schema {
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
							"color":              hexColorSchema(attrOptional),
							names.AttrExpression: stringLenBetweenSchema(attrRequired, 1, 4096),
							"icon_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingCustomIconOptions.html
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"icon":         stringEnumSchema[awstypes.Icon](attrOptional),
										"unicode_icon": stringMatchSchema(attrOptional, `^[^\\u0000-\\u00FF]$`, ""),
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
										"icon_display_option": stringEnumSchema[awstypes.ConditionalFormattingIconDisplayOption](attrOptional)},
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
							names.AttrExpression: stringLenBetweenSchema(attrRequired, 1, 4096),
							"icon_set_type":      stringEnumSchema[awstypes.ConditionalFormattingIconSetType](attrOptional),
						},
					},
				},
			},
		},
	}
})

func expandConditionalFormattingColor(tfList []any) *awstypes.ConditionalFormattingColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ConditionalFormattingColor{}

	if v, ok := tfMap["gradient"].([]any); ok && len(v) > 0 {
		apiObject.Gradient = expandConditionalFormattingGradientColor(v)
	}
	if v, ok := tfMap["solid"].([]any); ok && len(v) > 0 {
		apiObject.Solid = expandConditionalFormattingSolidColor(v)
	}

	return apiObject
}

func expandConditionalFormattingGradientColor(tfList []any) *awstypes.ConditionalFormattingGradientColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ConditionalFormattingGradientColor{}

	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}
	if v, ok := tfMap["color"].([]any); ok && len(v) > 0 {
		apiObject.Color = expandGradientColor(v)
	}

	return apiObject
}

func expandGradientColor(tfList []any) *awstypes.GradientColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GradientColor{}

	if v, ok := tfMap["stops"].([]any); ok && len(v) > 0 {
		apiObject.Stops = expandGradientStops(v)
	}

	return apiObject
}

func expandGradientStops(tfList []any) []awstypes.GradientStop {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.GradientStop

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandGradientStop(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandGradientStop(tfMap map[string]any) *awstypes.GradientStop {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GradientStop{}

	if v, ok := tfMap["gradient_offset"].(float64); ok {
		apiObject.GradientOffset = v
	}
	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap["data_value"].(float64); ok {
		apiObject.DataValue = aws.Float64(v)
	}

	return apiObject
}

func expandConditionalFormattingSolidColor(tfList []any) *awstypes.ConditionalFormattingSolidColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ConditionalFormattingSolidColor{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}

	return apiObject
}

func expandConditionalFormattingIcon(tfList []any) *awstypes.ConditionalFormattingIcon {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ConditionalFormattingIcon{}

	if v, ok := tfMap["custom_condition"].([]any); ok && len(v) > 0 {
		apiObject.CustomCondition = expandConditionalFormattingCustomIconCondition(v)
	}
	if v, ok := tfMap["icon_set"].([]any); ok && len(v) > 0 {
		apiObject.IconSet = expandConditionalFormattingIconSet(v)
	}

	return apiObject
}

func expandConditionalFormattingCustomIconCondition(tfList []any) *awstypes.ConditionalFormattingCustomIconCondition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ConditionalFormattingCustomIconCondition{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}
	if v, ok := tfMap["icon_options"].([]any); ok && len(v) > 0 {
		apiObject.IconOptions = expandConditionalFormattingCustomIconOptions(v)
	}
	if v, ok := tfMap["display_configuration"].([]any); ok && len(v) > 0 {
		apiObject.DisplayConfiguration = expandConditionalFormattingIconDisplayConfiguration(v)
	}

	return apiObject
}

func expandConditionalFormattingCustomIconOptions(tfList []any) *awstypes.ConditionalFormattingCustomIconOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ConditionalFormattingCustomIconOptions{}

	if v, ok := tfMap["icon"].(string); ok && v != "" {
		apiObject.Icon = awstypes.Icon(v)
	}
	if v, ok := tfMap["unicode_icon"].(string); ok && v != "" {
		apiObject.UnicodeIcon = aws.String(v)
	}

	return apiObject
}

func expandConditionalFormattingIconDisplayConfiguration(tfList []any) *awstypes.ConditionalFormattingIconDisplayConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ConditionalFormattingIconDisplayConfiguration{}

	if v, ok := tfMap["icon_display_option"].(string); ok && v != "" {
		apiObject.IconDisplayOption = awstypes.ConditionalFormattingIconDisplayOption(v)
	}

	return apiObject
}

func expandConditionalFormattingIconSet(tfList []any) *awstypes.ConditionalFormattingIconSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ConditionalFormattingIconSet{}

	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}
	if v, ok := tfMap["icon_set_type"].(string); ok && v != "" {
		apiObject.IconSetType = awstypes.ConditionalFormattingIconSetType(v)
	}

	return apiObject
}

func expandTextConditionalFormat(tfList []any) *awstypes.TextConditionalFormat {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TextConditionalFormat{}

	if v, ok := tfMap["background_color"].([]any); ok && len(v) > 0 {
		apiObject.BackgroundColor = expandConditionalFormattingColor(v)
	}
	if v, ok := tfMap["icon"].([]any); ok && len(v) > 0 {
		apiObject.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]any); ok && len(v) > 0 {
		apiObject.TextColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func flattenConditionalFormattingColor(apiObject *awstypes.ConditionalFormattingColor) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Gradient != nil {
		tfMap["gradient"] = flattenConditionalFormattingGradientColor(apiObject.Gradient)
	}
	if apiObject.Solid != nil {
		tfMap["solid"] = flattenConditionalFormattingSolidColor(apiObject.Solid)
	}

	return []any{tfMap}
}

func flattenConditionalFormattingGradientColor(apiObject *awstypes.ConditionalFormattingGradientColor) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Color != nil {
		tfMap["color"] = flattenGradientColor(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
	}

	return []any{tfMap}
}

func flattenGradientColor(apiObject *awstypes.GradientColor) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Stops != nil {
		tfMap["stops"] = flattenGradientStop(apiObject.Stops)
	}

	return []any{tfMap}
}

func flattenGradientStop(apiObjects []awstypes.GradientStop) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		tfMap["gradient_offset"] = apiObject.GradientOffset
		if apiObject.Color != nil {
			tfMap["color"] = aws.ToString(apiObject.Color)
		}
		if apiObject.DataValue != nil {
			tfMap["data_value"] = aws.ToFloat64(apiObject.DataValue)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenConditionalFormattingSolidColor(apiObject *awstypes.ConditionalFormattingSolidColor) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
	}

	return []any{tfMap}
}

func flattenConditionalFormattingIcon(apiObject *awstypes.ConditionalFormattingIcon) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CustomCondition != nil {
		tfMap["custom_condition"] = flattenConditionalFormattingCustomIconCondition(apiObject.CustomCondition)
	}
	if apiObject.IconSet != nil {
		tfMap["icon_set"] = flattenConditionalFormattingIconSet(apiObject.IconSet)
	}

	return []any{tfMap}
}

func flattenConditionalFormattingCustomIconCondition(apiObject *awstypes.ConditionalFormattingCustomIconCondition) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
	}
	if apiObject.IconOptions != nil {
		tfMap["icon_options"] = flattenConditionalFormattingCustomIconOptions(apiObject.IconOptions)
	}
	if apiObject.DisplayConfiguration != nil {
		tfMap["display_configuration"] = flattenConditionalFormattingIconDisplayConfiguration(apiObject.DisplayConfiguration)
	}

	return []any{tfMap}
}

func flattenConditionalFormattingCustomIconOptions(apiObject *awstypes.ConditionalFormattingCustomIconOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["icon"] = apiObject.Icon
	if apiObject.UnicodeIcon != nil {
		tfMap["unicode_icon"] = aws.ToString(apiObject.UnicodeIcon)
	}

	return []any{tfMap}
}

func flattenConditionalFormattingIconDisplayConfiguration(apiObject *awstypes.ConditionalFormattingIconDisplayConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"icon_display_option": apiObject.IconDisplayOption,
	}

	return []any{tfMap}
}

func flattenConditionalFormattingIconSet(apiObject *awstypes.ConditionalFormattingIconSet) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
	}
	tfMap["icon_set_type"] = apiObject.IconSetType

	return []any{tfMap}
}
