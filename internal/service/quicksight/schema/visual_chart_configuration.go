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

var axisDisplayOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"axis_line_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
				"axis_offset": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"data_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDataOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"date_axis_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateAxisOptions.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"missing_date_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
									},
								},
							},
							"numeric_axis_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericAxisOptions.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"range": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayRange.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"data_driven": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayDataDrivenRange.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{}, // For future extensions
														},
													},
													"min_max": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayMinMaxRange.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"maximum": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
																"minimum": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
										"scale": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisScale.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"linear": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisLinearScale.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"step_count": {
																	Type:     schema.TypeInt,
																	Optional: true,
																},
																"step_size": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
															},
														},
													},
													"logarithmic": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisLogarithmicScale.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"base": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"grid_line_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
				"scrollbar_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScrollBarOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
							"visible_range": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisibleRangeOptions.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"percent_range": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PercentVisibleRange.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"from": floatBetweenSchema(attrOptional, 0, 100),
													"to":   floatBetweenSchema(attrOptional, 0, 100),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"tick_label_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisTickLabelOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"label_options": labelOptionsSchema(),
							"rotation_angle": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
})

var chartAxisLabelOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"axis_label_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisLabelOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"apply_to": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisLabelReferenceOptions.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column":   columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
										"field_id": stringLenBetweenSchema(attrRequired, 1, 512),
									},
								},
							},
							"custom_label": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
						},
					},
				},
				"sort_icon_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
				"visibility":           stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

var itemsLimitConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"items_limit": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"other_categories": stringEnumSchema[awstypes.OtherCategories](attrRequired),
			},
		},
	}
})

var contributionAnalysisDefaultsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ContributionAnalysisDefault.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 200,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"contributor_dimensions": {
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 4,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"column_name":         stringLenBetweenSchema(attrRequired, 1, 128),
							"data_set_identifier": stringLenBetweenSchema(attrRequired, 1, 2048),
						},
					},
				},
				"measure_field_id": stringLenBetweenSchema(attrRequired, 1, 512),
			},
		},
	}
})

var referenceLineSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: referenceLinesMaxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineDataConfiguration.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"axis_binding": stringEnumSchema[awstypes.AxisBinding](attrOptional),
							"dynamic_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineDynamicDataConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"calculation":                  numericalAggregationFunctionSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericalAggregationFunction.html
										"column":                       columnSchema(true),                       // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
										"measure_aggregation_function": aggregationFunctionSchema(true),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
									},
								},
							},
							"static_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineStaticDataConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrValue: {
											Type:     schema.TypeFloat,
											Required: true,
										},
									},
								},
							},
						},
					},
				},
				"label_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineLabelConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"custom_label_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineCustomLabelConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"custom_label": stringMatchSchema(attrRequired, `.*\S.*`, ""),
									},
								},
							},
							"font_color":          hexColorSchema(attrOptional),
							"font_configuration":  fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
							"horizontal_position": stringEnumSchema[awstypes.ReferenceLineLabelHorizontalPosition](attrOptional),
							"value_label_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineValueLabelConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"format_configuration": numericFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericFormatConfiguration.html
										"relative_position":    stringEnumSchema[awstypes.ReferenceLineValueLabelRelativePosition](attrOptional),
									},
								},
							},
							"vertical_position": stringEnumSchema[awstypes.ReferenceLineLabelVerticalPosition](attrOptional),
						},
					},
				},
				names.AttrStatus: stringEnumSchema[awstypes.Status](attrOptional),
				"style_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineStyleConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":   hexColorSchema(attrOptional),
							"pattern": stringEnumSchema[awstypes.ReferenceLinePatternType](attrOptional),
						},
					},
				},
			},
		},
	}
})

var smallMultiplesOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SmallMultiplesOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"max_visible_columns": intBetweenSchema(attrOptional, 1, 10),
				"max_visible_rows":    intBetweenSchema(attrOptional, 1, 10),
				"panel_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PanelConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"background_color":      stringMatchSchema(attrOptional, `^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`, ""),
							"background_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
							"border_color":          stringMatchSchema(attrOptional, `^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`, ""),
							"border_style":          stringEnumSchema[awstypes.PanelBorderStyle](attrOptional),
							"border_thickness": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"border_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
							"gutter_spacing": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"gutter_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
							"title": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PanelTitleOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"font_configuration":        fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
										"horizontal_text_alignment": stringEnumSchema[awstypes.HorizontalTextAlignment](attrOptional),
										"visibility":                stringEnumSchema[awstypes.Visibility](attrOptional),
									},
								},
							},
						},
					},
				},
			},
		},
	}
})

func expandAxisDisplayOptions(tfList []any) *awstypes.AxisDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisDisplayOptions{}

	if v, ok := tfMap["axis_line_visibility"].(string); ok && v != "" {
		apiObject.AxisLineVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["axis_offset"].(string); ok && v != "" {
		apiObject.AxisOffset = aws.String(v)
	}
	if v, ok := tfMap["grid_line_visibility"].(string); ok && v != "" {
		apiObject.GridLineVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["data_options"].([]any); ok && len(v) > 0 {
		apiObject.DataOptions = expandAxisDataOptions(v)
	}
	if v, ok := tfMap["scrollbar_options"].([]any); ok && len(v) > 0 {
		apiObject.ScrollbarOptions = expandScrollBarOptions(v)
	}
	if v, ok := tfMap["tick_label_options"].([]any); ok && len(v) > 0 {
		apiObject.TickLabelOptions = expandAxisTickLabelOptions(v)
	}

	return apiObject
}

func expandAxisDataOptions(tfList []any) *awstypes.AxisDataOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisDataOptions{}

	if v, ok := tfMap["date_axis_options"].([]any); ok && len(v) > 0 {
		apiObject.DateAxisOptions = expandDateAxisOptions(v)
	}
	if v, ok := tfMap["numeric_axis_options"].([]any); ok && len(v) > 0 {
		apiObject.NumericAxisOptions = expandNumericAxisOptions(v)
	}

	return apiObject
}

func expandDateAxisOptions(tfList []any) *awstypes.DateAxisOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateAxisOptions{}

	if v, ok := tfMap["missing_date_visibility"].(string); ok && v != "" {
		apiObject.MissingDateVisibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandNumericAxisOptions(tfList []any) *awstypes.NumericAxisOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericAxisOptions{}

	if v, ok := tfMap["range"].([]any); ok && len(v) > 0 {
		apiObject.Range = expandAxisDisplayRange(v)
	}
	if v, ok := tfMap["scale"].([]any); ok && len(v) > 0 {
		apiObject.Scale = expandAxisScale(v)
	}

	return apiObject
}

func expandAxisDisplayRange(tfList []any) *awstypes.AxisDisplayRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisDisplayRange{}

	if v, ok := tfMap["data_driven"].([]any); ok && len(v) > 0 {
		apiObject.DataDriven = expandAxisDisplayDataDrivenRange(v)
	}
	if v, ok := tfMap["min_max"].([]any); ok && len(v) > 0 {
		apiObject.MinMax = expandAxisDisplayMinMaxRange(v)
	}

	return apiObject
}

func expandAxisDisplayDataDrivenRange(tfList []any) *awstypes.AxisDisplayDataDrivenRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	//tfMap, ok := tfList[0].(map[string]interface{})
	//if !ok {
	//	return nil
	//}

	apiObject := &awstypes.AxisDisplayDataDrivenRange{}

	return apiObject
}

func expandAxisDisplayMinMaxRange(tfList []any) *awstypes.AxisDisplayMinMaxRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisDisplayMinMaxRange{}

	if v, ok := tfMap["maximum"].(float64); ok {
		apiObject.Maximum = aws.Float64(v)
	}
	if v, ok := tfMap["minimum"].(float64); ok {
		apiObject.Minimum = aws.Float64(v)
	}

	return apiObject
}

func expandAxisScale(tfList []any) *awstypes.AxisScale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisScale{}

	if v, ok := tfMap["linear"].([]any); ok && len(v) > 0 {
		apiObject.Linear = expandAxisLinearScale(v)
	}
	if v, ok := tfMap["logarithmic"].([]any); ok && len(v) > 0 {
		apiObject.Logarithmic = expandAxisLogarithmicScale(v)
	}

	return apiObject
}

func expandAxisLinearScale(tfList []any) *awstypes.AxisLinearScale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisLinearScale{}

	if v, ok := tfMap["step_count"].(int); ok {
		apiObject.StepCount = aws.Int32(int32(v))
	}
	if v, ok := tfMap["step_size"].(float64); ok {
		apiObject.StepSize = aws.Float64(v)
	}

	return apiObject
}

func expandAxisLogarithmicScale(tfList []any) *awstypes.AxisLogarithmicScale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisLogarithmicScale{}

	if v, ok := tfMap["base"].(float64); ok {
		apiObject.Base = aws.Float64(v)
	}

	return apiObject
}

func expandScrollBarOptions(tfList []any) *awstypes.ScrollBarOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ScrollBarOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["visible_range"].([]any); ok && len(v) > 0 {
		apiObject.VisibleRange = expandVisibleRangeOptions(v)
	}

	return apiObject
}

func expandVisibleRangeOptions(tfList []any) *awstypes.VisibleRangeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.VisibleRangeOptions{}

	if v, ok := tfMap["percent_range"].([]any); ok && len(v) > 0 {
		apiObject.PercentRange = expandPercentVisibleRange(v)
	}

	return apiObject
}

func expandPercentVisibleRange(tfList []any) *awstypes.PercentVisibleRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PercentVisibleRange{}

	if v, ok := tfMap["from"].(float64); ok {
		apiObject.From = aws.Float64(v)
	}
	if v, ok := tfMap["to"].(float64); ok {
		apiObject.To = aws.Float64(v)
	}

	return apiObject
}

func expandAxisTickLabelOptions(tfList []any) *awstypes.AxisTickLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisTickLabelOptions{}

	if v, ok := tfMap["rotation_angle"].(float64); ok {
		apiObject.RotationAngle = aws.Float64(v)
	}
	if v, ok := tfMap["label_options"].([]any); ok && len(v) > 0 {
		apiObject.LabelOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandChartAxisLabelOptions(tfList []any) *awstypes.ChartAxisLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ChartAxisLabelOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["sort_icon_visibility"].(string); ok && v != "" {
		apiObject.SortIconVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["axis_label_options"].([]any); ok && len(v) > 0 {
		apiObject.AxisLabelOptions = expandAxisLabelOptionsList(v)
	}

	return apiObject
}

func expandAxisLabelOptionsList(tfList []any) []awstypes.AxisLabelOptions {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.AxisLabelOptions

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandAxisLabelOptions(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandAxisLabelOptions(tfMap map[string]any) *awstypes.AxisLabelOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AxisLabelOptions{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		apiObject.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["apply_to"].([]any); ok && len(v) > 0 {
		apiObject.ApplyTo = expandAxisLabelReferenceOptions(v)
	}
	if v, ok := tfMap["font_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FontConfiguration = expandFontConfiguration(v)
	}

	return apiObject
}

func expandAxisLabelReferenceOptions(tfList []any) *awstypes.AxisLabelReferenceOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.AxisLabelReferenceOptions{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}

	return apiObject
}

func expandContributionAnalysisDefaults(tfList []any) []awstypes.ContributionAnalysisDefault {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ContributionAnalysisDefault

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandContributionAnalysisDefault(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandContributionAnalysisDefault(tfMap map[string]any) *awstypes.ContributionAnalysisDefault {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ContributionAnalysisDefault{}

	if v, ok := tfMap["measure_field_id"].(string); ok && v != "" {
		apiObject.MeasureFieldId = aws.String(v)
	}
	if v, ok := tfMap["contributor_dimensions"].([]any); ok && len(v) > 0 {
		apiObject.ContributorDimensions = expandColumnIdentifiers(v)
	}

	return apiObject
}

func expandReferenceLines(tfList []any) []awstypes.ReferenceLine {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ReferenceLine

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandReferenceLine(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandReferenceLine(tfMap map[string]any) *awstypes.ReferenceLine {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ReferenceLine{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.WidgetStatus(v)
	}
	if v, ok := tfMap["data_configuration"].([]any); ok && len(v) > 0 {
		apiObject.DataConfiguration = expandReferenceLineDataConfiguration(v)
	}
	if v, ok := tfMap["label_configuration"].([]any); ok && len(v) > 0 {
		apiObject.LabelConfiguration = expandReferenceLineLabelConfiguration(v)
	}
	if v, ok := tfMap["style_configuration"].([]any); ok && len(v) > 0 {
		apiObject.StyleConfiguration = expandReferenceLineStyleConfiguration(v)
	}

	return apiObject
}

func expandReferenceLineDataConfiguration(tfList []any) *awstypes.ReferenceLineDataConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReferenceLineDataConfiguration{}

	if v, ok := tfMap["axis_binding"].(string); ok && v != "" {
		apiObject.AxisBinding = awstypes.AxisBinding(v)
	}
	if v, ok := tfMap["dynamic_configuration"].([]any); ok && len(v) > 0 {
		apiObject.DynamicConfiguration = expandReferenceLineDynamicDataConfiguration(v)
	}
	if v, ok := tfMap["static_configuration"].([]any); ok && len(v) > 0 {
		apiObject.StaticConfiguration = expandReferenceLineStaticDataConfiguration(v)
	}

	return apiObject
}

func expandReferenceLineDynamicDataConfiguration(tfList []any) *awstypes.ReferenceLineDynamicDataConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReferenceLineDynamicDataConfiguration{}

	if v, ok := tfMap["calculation"].([]any); ok && len(v) > 0 {
		apiObject.Calculation = expandNumericalAggregationFunction(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["measure_aggregation_function"].([]any); ok && len(v) > 0 {
		apiObject.MeasureAggregationFunction = expandAggregationFunction(v)
	}

	return apiObject
}

func expandReferenceLineStaticDataConfiguration(tfList []any) *awstypes.ReferenceLineStaticDataConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReferenceLineStaticDataConfiguration{}

	if v, ok := tfMap[names.AttrValue].(float64); ok {
		apiObject.Value = v
	}

	return apiObject
}

func expandReferenceLineLabelConfiguration(tfList []any) *awstypes.ReferenceLineLabelConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReferenceLineLabelConfiguration{}

	if v, ok := tfMap["font_color"].(string); ok && v != "" {
		apiObject.FontColor = aws.String(v)
	}
	if v, ok := tfMap["horizontal_position"].(string); ok && v != "" {
		apiObject.HorizontalPosition = awstypes.ReferenceLineLabelHorizontalPosition(v)
	}
	if v, ok := tfMap["vertical_position"].(string); ok && v != "" {
		apiObject.VerticalPosition = awstypes.ReferenceLineLabelVerticalPosition(v)
	}
	if v, ok := tfMap["custom_label_configuration"].([]any); ok && len(v) > 0 {
		apiObject.CustomLabelConfiguration = expandReferenceLineCustomLabelConfiguration(v)
	}
	if v, ok := tfMap["font_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FontConfiguration = expandFontConfiguration(v)
	}
	if v, ok := tfMap["value_label_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ValueLabelConfiguration = expandReferenceLineValueLabelConfiguration(v)
	}

	return apiObject
}

func expandReferenceLineCustomLabelConfiguration(tfList []any) *awstypes.ReferenceLineCustomLabelConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReferenceLineCustomLabelConfiguration{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		apiObject.CustomLabel = aws.String(v)
	}

	return apiObject
}

func expandReferenceLineValueLabelConfiguration(tfList []any) *awstypes.ReferenceLineValueLabelConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReferenceLineValueLabelConfiguration{}

	if v, ok := tfMap["relative_position"].(string); ok && v != "" {
		apiObject.RelativePosition = awstypes.ReferenceLineValueLabelRelativePosition(v)
	}
	if v, ok := tfMap["format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandNumericFormatConfiguration(v)
	}

	return apiObject
}

func expandReferenceLineStyleConfiguration(tfList []any) *awstypes.ReferenceLineStyleConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReferenceLineStyleConfiguration{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap["pattern"].(string); ok && v != "" {
		apiObject.Pattern = awstypes.ReferenceLinePatternType(v)
	}

	return apiObject
}

func expandSmallMultiplesOptions(tfList []any) *awstypes.SmallMultiplesOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.SmallMultiplesOptions{}

	if v, ok := tfMap["max_visible_columns"].(int); ok {
		apiObject.MaxVisibleColumns = aws.Int64(int64(v))
	}
	if v, ok := tfMap["max_visible_rows"].(int); ok {
		apiObject.MaxVisibleRows = aws.Int64(int64(v))
	}
	if v, ok := tfMap["panel_configuration"].([]any); ok && len(v) > 0 {
		apiObject.PanelConfiguration = expandPanelConfiguration(v)
	}

	return apiObject
}

func expandPanelConfiguration(tfList []any) *awstypes.PanelConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PanelConfiguration{}

	if v, ok := tfMap["background_color"].(string); ok && v != "" {
		apiObject.BackgroundColor = aws.String(v)
	}
	if v, ok := tfMap["background_visibility"].(string); ok && v != "" {
		apiObject.BackgroundVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["border_color"].(string); ok && v != "" {
		apiObject.BorderColor = aws.String(v)
	}
	if v, ok := tfMap["border_style"].(string); ok && v != "" {
		apiObject.BorderStyle = awstypes.PanelBorderStyle(v)
	}
	if v, ok := tfMap["border_thickness"].(string); ok && v != "" {
		apiObject.BorderThickness = aws.String(v)
	}
	if v, ok := tfMap["border_visibility"].(string); ok && v != "" {
		apiObject.BorderVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["gutter_spacing"].(string); ok && v != "" {
		apiObject.GutterSpacing = aws.String(v)
	}
	if v, ok := tfMap["gutter_visibility"].(string); ok && v != "" {
		apiObject.GutterVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandPanelTitleOptions(v)
	}

	return apiObject
}

func expandPanelTitleOptions(tfList []any) *awstypes.PanelTitleOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PanelTitleOptions{}

	if v, ok := tfMap["horizontal_text_alignment"].(string); ok && v != "" {
		apiObject.HorizontalTextAlignment = awstypes.HorizontalTextAlignment(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["font_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FontConfiguration = expandFontConfiguration(v)
	}

	return apiObject
}

func expandItemsLimitConfiguration(tfList []any) *awstypes.ItemsLimitConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ItemsLimitConfiguration{}

	if v, ok := tfMap["items_limit"].(int); ok {
		apiObject.ItemsLimit = aws.Int64(int64(v))
	}
	if v, ok := tfMap["other_categories"].(string); ok && v != "" {
		apiObject.OtherCategories = awstypes.OtherCategories(v)
	}

	return apiObject
}

func flattenAxisDisplayOptions(apiObject *awstypes.AxisDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["axis_line_visibility"] = apiObject.AxisLineVisibility
	if apiObject.AxisOffset != nil {
		tfMap["axis_offset"] = aws.ToString(apiObject.AxisOffset)
	}
	if apiObject.DataOptions != nil {
		tfMap["data_options"] = flattenAxisDataOptions(apiObject.DataOptions)
	}
	tfMap["grid_line_visibility"] = apiObject.GridLineVisibility
	if apiObject.ScrollbarOptions != nil {
		tfMap["scrollbar_options"] = flattenScrollBarOptions(apiObject.ScrollbarOptions)
	}
	if apiObject.TickLabelOptions != nil {
		tfMap["tick_label_options"] = flattenAxisTickLabelOptions(apiObject.TickLabelOptions)
	}

	return []any{tfMap}
}

func flattenAxisDataOptions(apiObject *awstypes.AxisDataOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DateAxisOptions != nil {
		tfMap["date_axis_options"] = flattenDateAxisOptions(apiObject.DateAxisOptions)
	}
	if apiObject.NumericAxisOptions != nil {
		tfMap["numeric_axis_options"] = flattenNumericAxisOptions(apiObject.NumericAxisOptions)
	}

	return []any{tfMap}
}

func flattenDateAxisOptions(apiObject *awstypes.DateAxisOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"missing_date_visibility": apiObject.MissingDateVisibility,
	}

	return []any{tfMap}
}

func flattenNumericAxisOptions(apiObject *awstypes.NumericAxisOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Range != nil {
		tfMap["range"] = flattenAxisDisplayRange(apiObject.Range)
	}
	if apiObject.Scale != nil {
		tfMap["scale"] = flattenAxisScale(apiObject.Scale)
	}

	return []any{tfMap}
}

func flattenAxisDisplayRange(apiObject *awstypes.AxisDisplayRange) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DataDriven != nil {
		tfMap["data_driven"] = flattenAxisDisplayDataDrivenRange(apiObject.DataDriven)
	}
	if apiObject.MinMax != nil {
		tfMap["min_max"] = flattenAxisDisplayMinMaxRange(apiObject.MinMax)
	}

	return []any{tfMap}
}

func flattenAxisDisplayDataDrivenRange(apiObject *awstypes.AxisDisplayDataDrivenRange) []any {
	if apiObject == nil {
		return nil
	}

	// For future extensions
	tfMap := map[string]any{}

	return []any{tfMap}
}

func flattenAxisDisplayMinMaxRange(apiObject *awstypes.AxisDisplayMinMaxRange) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Maximum != nil {
		tfMap["maximum"] = aws.ToFloat64(apiObject.Maximum)
	}
	if apiObject.Minimum != nil {
		tfMap["minimum"] = aws.ToFloat64(apiObject.Minimum)
	}

	return []any{tfMap}
}

func flattenAxisScale(apiObject *awstypes.AxisScale) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Linear != nil {
		tfMap["linear"] = flattenAxisLinearScale(apiObject.Linear)
	}
	if apiObject.Logarithmic != nil {
		tfMap["logarithmic"] = flattenAxisLogarithmicScale(apiObject.Logarithmic)
	}

	return []any{tfMap}
}

func flattenAxisLinearScale(apiObject *awstypes.AxisLinearScale) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.StepCount != nil {
		tfMap["step_count"] = aws.ToInt32(apiObject.StepCount)
	}
	if apiObject.StepSize != nil {
		tfMap["step_size"] = aws.ToFloat64(apiObject.StepSize)
	}

	return []any{tfMap}
}

func flattenAxisLogarithmicScale(apiObject *awstypes.AxisLogarithmicScale) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Base != nil {
		tfMap["base"] = aws.ToFloat64(apiObject.Base)
	}

	return []any{tfMap}
}

func flattenScrollBarOptions(apiObject *awstypes.ScrollBarOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["visibility"] = apiObject.Visibility
	if apiObject.VisibleRange != nil {
		tfMap["visible_range"] = flattenVisibleRangeOptions(apiObject.VisibleRange)
	}

	return []any{tfMap}
}

func flattenVisibleRangeOptions(apiObject *awstypes.VisibleRangeOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.PercentRange != nil {
		tfMap["percent_range"] = flattenPercentVisibleRange(apiObject.PercentRange)
	}

	return []any{tfMap}
}

func flattenPercentVisibleRange(apiObject *awstypes.PercentVisibleRange) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.From != nil {
		tfMap["from"] = aws.ToFloat64(apiObject.From)
	}
	if apiObject.To != nil {
		tfMap["to"] = aws.ToFloat64(apiObject.To)
	}

	return []any{tfMap}
}

func flattenAxisTickLabelOptions(apiObject *awstypes.AxisTickLabelOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.LabelOptions != nil {
		tfMap["label_options"] = flattenLabelOptions(apiObject.LabelOptions)
	}
	if apiObject.RotationAngle != nil {
		tfMap["rotation_angle"] = aws.ToFloat64(apiObject.RotationAngle)
	}

	return []any{tfMap}
}

func flattenChartAxisLabelOptions(apiObject *awstypes.ChartAxisLabelOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.AxisLabelOptions != nil {
		tfMap["axis_label_options"] = flattenAxisLabelOptions(apiObject.AxisLabelOptions)
	}
	tfMap["sort_icon_visibility"] = apiObject.SortIconVisibility
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenAxisLabelOptions(apiObjects []awstypes.AxisLabelOptions) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.ApplyTo != nil {
			tfMap["apply_to"] = flattenAxisLabelReferenceOptions(apiObject.ApplyTo)
		}
		if apiObject.CustomLabel != nil {
			tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
		}
		if apiObject.FontConfiguration != nil {
			tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAxisLabelReferenceOptions(apiObject *awstypes.AxisLabelReferenceOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}

	return []any{tfMap}
}

func flattenContributionAnalysisDefault(apiObjects []awstypes.ContributionAnalysisDefault) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"measure_field_id": aws.ToString(apiObject.MeasureFieldId),
		}

		if apiObject.ContributorDimensions != nil {
			tfMap["contribution_dimensions"] = flattenColumnIdentifiers(apiObject.ContributorDimensions)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnIdentifiers(apiObjects []awstypes.ColumnIdentifier) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"column_name":         aws.ToString(apiObject.ColumnName),
			"data_set_identifier": aws.ToString(apiObject.DataSetIdentifier),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenReferenceLine(apiObjects []awstypes.ReferenceLine) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.DataConfiguration != nil {
			tfMap["data_configuration"] = flattenReferenceLineDataConfiguration(apiObject.DataConfiguration)
		}
		if apiObject.LabelConfiguration != nil {
			tfMap["label_configuration"] = flattenReferenceLineLabelConfiguration(apiObject.LabelConfiguration)
		}
		tfMap[names.AttrStatus] = apiObject.Status
		if apiObject.StyleConfiguration != nil {
			tfMap["style_configuration"] = flattenReferenceLineStyleConfiguration(apiObject.StyleConfiguration)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenReferenceLineDataConfiguration(apiObject *awstypes.ReferenceLineDataConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["axis_binding"] = apiObject.AxisBinding
	if apiObject.DynamicConfiguration != nil {
		tfMap["dynamic_configuration"] = flattenReferenceLineDynamicDataConfiguration(apiObject.DynamicConfiguration)
	}
	if apiObject.StaticConfiguration != nil {
		tfMap["static_configuration"] = flattenReferenceLineStaticDataConfiguration(apiObject.StaticConfiguration)
	}

	return []any{tfMap}
}

func flattenReferenceLineDynamicDataConfiguration(apiObject *awstypes.ReferenceLineDynamicDataConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Calculation != nil {
		tfMap["calculation"] = flattenNumericalAggregationFunction(apiObject.Calculation)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.MeasureAggregationFunction != nil {
		tfMap["measure_aggregation_function"] = flattenAggregationFunction(apiObject.MeasureAggregationFunction)
	}

	return []any{tfMap}
}

func flattenReferenceLineStaticDataConfiguration(apiObject *awstypes.ReferenceLineStaticDataConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrValue: apiObject.Value,
	}

	return []any{tfMap}
}

func flattenReferenceLineLabelConfiguration(apiObject *awstypes.ReferenceLineLabelConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CustomLabelConfiguration != nil {
		tfMap["custom_label_configuration"] = flattenReferenceLineCustomLabelConfiguration(apiObject.CustomLabelConfiguration)
	}
	if apiObject.FontColor != nil {
		tfMap["font_color"] = aws.ToString(apiObject.FontColor)
	}
	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	tfMap["horizontal_position"] = apiObject.HorizontalPosition
	if apiObject.ValueLabelConfiguration != nil {
		tfMap["value_label_configuration"] = flattenReferenceLineValueLabelConfiguration(apiObject.ValueLabelConfiguration)
	}
	tfMap["vertical_position"] = apiObject.VerticalPosition

	return []any{tfMap}
}
func flattenReferenceLineCustomLabelConfiguration(apiObject *awstypes.ReferenceLineCustomLabelConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"custom_label": aws.ToString(apiObject.CustomLabel),
	}

	return []any{tfMap}
}

func flattenReferenceLineValueLabelConfiguration(apiObject *awstypes.ReferenceLineValueLabelConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenNumericFormatConfiguration(apiObject.FormatConfiguration)
	}
	tfMap["relative_position"] = apiObject.RelativePosition

	return []any{tfMap}
}

func flattenReferenceLineStyleConfiguration(apiObject *awstypes.ReferenceLineStyleConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	tfMap["pattern"] = apiObject.Pattern

	return []any{tfMap}
}

func flattenSmallMultiplesOptions(apiObject *awstypes.SmallMultiplesOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.MaxVisibleColumns != nil {
		tfMap["max_visible_columns"] = aws.ToInt64(apiObject.MaxVisibleColumns)
	}
	if apiObject.MaxVisibleRows != nil {
		tfMap["max_visible_rows"] = aws.ToInt64(apiObject.MaxVisibleRows)
	}
	if apiObject.PanelConfiguration != nil {
		tfMap["panel_configuration"] = flattenPanelConfiguration(apiObject.PanelConfiguration)
	}

	return []any{tfMap}
}

func flattenPanelConfiguration(apiObject *awstypes.PanelConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = aws.ToString(apiObject.BackgroundColor)
	}
	tfMap["background_visibility"] = apiObject.BackgroundVisibility
	if apiObject.BorderColor != nil {
		tfMap["border_color"] = aws.ToString(apiObject.BorderColor)
	}
	tfMap["border_style"] = apiObject.BorderStyle
	if apiObject.BorderThickness != nil {
		tfMap["border_thickness"] = aws.ToString(apiObject.BorderThickness)
	}
	tfMap["border_visibility"] = apiObject.BorderVisibility
	if apiObject.GutterSpacing != nil {
		tfMap["gutter_spacing"] = aws.ToString(apiObject.GutterSpacing)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenPanelTitleOptions(apiObject.Title)
	}

	return []any{tfMap}
}

func flattenPanelTitleOptions(apiObject *awstypes.PanelTitleOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	tfMap["horizontal_text_alignment"] = apiObject.HorizontalTextAlignment
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}
