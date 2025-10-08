// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func lineChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"contribution_analysis_defaults": contributionAnalysisDefaultsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ContributionAnalysisDefault.html
							"data_labels":                    dataLabelOptionsSchema(),             // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"default_series_settings": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartDefaultSeriesSettings.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"axis_binding":          stringEnumSchema[awstypes.AxisBinding](attrOptional),
										"line_style_settings":   lineChartLineStyleSettingsSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartLineStyleSettings.html
										"marker_style_settings": lineChartMarkerStyleSettingsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartMarkerStyleSettings.html
									},
								},
							},
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"line_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"category":        dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"colors":          dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"small_multiples": dimensionFieldSchema(1),                          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues:  measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"forecast_configurations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ForecastConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"forecast_properties": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeBasedForecastProperties.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"lower_boundary": {
														Type:     schema.TypeFloat,
														Optional: true,
													},
													"periods_backward":    intBetweenSchema(attrOptional, 0, 1000),
													"periods_forward":     intBetweenSchema(attrOptional, 1, 1000),
													"prediction_interval": intBetweenSchema(attrOptional, 50, 95),
													"seasonality":         intBetweenSchema(attrOptional, 1, 180),
													"upper_boundary": {
														Type:     schema.TypeFloat,
														Optional: true,
													},
												},
											},
										},
										"scenario": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ForecastScenario.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"what_if_point_scenario": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WhatIfPointScenario.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"date": utcTimestampStringSchema(attrRequired),
																names.AttrValue: {
																	Type:     schema.TypeFloat,
																	Required: true,
																},
															},
														},
													},
													"what_if_range_scenario": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WhatIfRangeScenario.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"end_date":   utcTimestampStringSchema(attrRequired),
																"start_date": utcTimestampStringSchema(attrRequired),
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
									},
								},
							},
							"legend": legendOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"primary_y_axis_display_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineSeriesAxisDisplayOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"axis_options": axisDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
										"missing_data_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MissingDataConfiguration.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"treatment_option": stringEnumSchema[awstypes.MissingDataTreatmentOption](attrOptional),
												},
											},
										},
									},
								},
							},
							"primary_y_axis_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"reference_lines":              referenceLineSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
							"secondary_y_axis_display_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineSeriesAxisDisplayOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"axis_options": axisDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
										"missing_data_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MissingDataConfiguration.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"treatment_option": stringEnumSchema[awstypes.MissingDataTreatmentOption](attrOptional),
												},
											},
										},
									},
								},
							},
							"secondary_y_axis_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"series": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SeriesItem.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"data_field_series_item": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataFieldSeriesItem.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"axis_binding": stringEnumSchema[awstypes.AxisBinding](attrRequired),
													"field_id":     stringLenBetweenSchema(attrRequired, 1, 512),
													"field_value": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"settings": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartSeriesSettings.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"line_style_settings":   lineChartLineStyleSettingsSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartLineStyleSettings.html
																"marker_style_settings": lineChartMarkerStyleSettingsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartMarkerStyleSettings.html
															},
														},
													},
												},
											},
										},
										"field_series_item": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSeriesItem.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"axis_binding": stringEnumSchema[awstypes.AxisBinding](attrRequired),
													"field_id":     stringLenBetweenSchema(attrRequired, 1, 512),
													"settings": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartSeriesSettings.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"line_style_settings":   lineChartLineStyleSettingsSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartLineStyleSettings.html
																"marker_style_settings": lineChartMarkerStyleSettingsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartMarkerStyleSettings.html
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"small_multiples_options": smallMultiplesOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SmallMultiplesOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit_configuration":  itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":                       fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"color_items_limit_configuration":     itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"small_multiples_limit_configuration": itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"small_multiples_sort":                fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"tooltip":                tooltipOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							names.AttrType:           stringEnumSchema[awstypes.LineChartType](attrOptionalComputed),
							"visual_palette":         visualPaletteSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
							"x_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"x_axis_label_options":   chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
						},
					},
				},
				"column_hierarchies": columnHierarchiesSchema(),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnHierarchy.html
				"subtitle":           visualSubtitleLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
				"title":              visualTitleLabelOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
			},
		},
	}
}

var lineChartLineStyleSettingsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartLineStyleSettings.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"line_interpolation": stringEnumSchema[awstypes.LineInterpolation](attrOptional),
				"line_style":         stringEnumSchema[awstypes.LineChartLineStyle](attrOptional),
				"line_visibility":    stringEnumSchema[awstypes.Visibility](attrOptional),
				"line_width": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
})

var lineChartMarkerStyleSettingsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartMarkerStyleSettings.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"marker_color": hexColorSchema(attrOptional),
				"marker_shape": stringEnumSchema[awstypes.LineChartMarkerShape](attrOptional),
				"marker_size": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"marker_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

func expandLineChartVisual(tfList []any) *awstypes.LineChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandLineChartConfiguration(v)
	}
	if v, ok := tfMap["column_hierarchies"].([]any); ok && len(v) > 0 {
		apiObject.ColumnHierarchies = expandColumnHierarchies(v)
	}
	if v, ok := tfMap["subtitle"].([]any); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandLineChartConfiguration(tfList []any) *awstypes.LineChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartConfiguration{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.LineChartType(v)
	}
	if v, ok := tfMap["contribution_analysis_defaults"].([]any); ok && len(v) > 0 {
		apiObject.ContributionAnalysisDefaults = expandContributionAnalysisDefaults(v)
	}
	if v, ok := tfMap["data_labels"].([]any); ok && len(v) > 0 {
		apiObject.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["default_series_settings"].([]any); ok && len(v) > 0 {
		apiObject.DefaultSeriesSettings = expandLineChartDefaultSeriesSettings(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandLineChartFieldWells(v)
	}
	if v, ok := tfMap["forecast_configurations"].([]any); ok && len(v) > 0 {
		apiObject.ForecastConfigurations = expandForecastConfigurations(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_display_options"].([]any); ok && len(v) > 0 {
		apiObject.PrimaryYAxisDisplayOptions = expandLineSeriesAxisDisplayOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_label_options"].([]any); ok && len(v) > 0 {
		apiObject.PrimaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["reference_lines"].([]any); ok && len(v) > 0 {
		apiObject.ReferenceLines = expandReferenceLines(v)
	}
	if v, ok := tfMap["secondary_y_axis_display_options"].([]any); ok && len(v) > 0 {
		apiObject.SecondaryYAxisDisplayOptions = expandLineSeriesAxisDisplayOptions(v)
	}
	if v, ok := tfMap["secondary_y_axis_label_options"].([]any); ok && len(v) > 0 {
		apiObject.SecondaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["series"].([]any); ok && len(v) > 0 {
		apiObject.Series = expandSeriesItems(v)
	}
	if v, ok := tfMap["small_multiples_options"].([]any); ok && len(v) > 0 {
		apiObject.SmallMultiplesOptions = expandSmallMultiplesOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandLineChartSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]any); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}
	if v, ok := tfMap["x_axis_display_options"].([]any); ok && len(v) > 0 {
		apiObject.XAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["x_axis_label_options"].([]any); ok && len(v) > 0 {
		apiObject.XAxisLabelOptions = expandChartAxisLabelOptions(v)
	}

	return apiObject
}

func expandLineChartFieldWells(tfList []any) *awstypes.LineChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartFieldWells{}

	if v, ok := tfMap["line_chart_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.LineChartAggregatedFieldWells = expandLineChartAggregatedFieldWells(v)
	}

	return apiObject
}

func expandLineChartAggregatedFieldWells(tfList []any) *awstypes.LineChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["colors"].([]any); ok && len(v) > 0 {
		apiObject.Colors = expandDimensionFields(v)
	}
	if v, ok := tfMap["small_multiples"].([]any); ok && len(v) > 0 {
		apiObject.SmallMultiples = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandLineChartSortConfiguration(tfList []any) *awstypes.LineChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]any); ok && len(v) > 0 {
		apiObject.CategoryItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]any); ok && len(v) > 0 {
		apiObject.CategorySort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["color_items_limit_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ColorItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["small_multiples_limit_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SmallMultiplesLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["small_multiples_sort"].([]any); ok && len(v) > 0 {
		apiObject.SmallMultiplesSort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func expandLineChartDefaultSeriesSettings(tfList []any) *awstypes.LineChartDefaultSeriesSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartDefaultSeriesSettings{}

	if v, ok := tfMap["axis_binding"].(string); ok && v != "" {
		apiObject.AxisBinding = awstypes.AxisBinding(v)
	}
	if v, ok := tfMap["line_style_settings"].([]any); ok && len(v) > 0 {
		apiObject.LineStyleSettings = expandLineChartLineStyleSettings(v)
	}
	if v, ok := tfMap["marker_style_settings"].([]any); ok && len(v) > 0 {
		apiObject.MarkerStyleSettings = expandLineChartMarkerStyleSettings(v)
	}

	return apiObject
}

func expandLineChartLineStyleSettings(tfList []any) *awstypes.LineChartLineStyleSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartLineStyleSettings{}

	if v, ok := tfMap["line_interpolation"].(string); ok && v != "" {
		apiObject.LineInterpolation = awstypes.LineInterpolation(v)
	}
	if v, ok := tfMap["line_style"].(string); ok && v != "" {
		apiObject.LineStyle = awstypes.LineChartLineStyle(v)
	}
	if v, ok := tfMap["line_visibility"].(string); ok && v != "" {
		apiObject.LineVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["line_width"].(string); ok && v != "" {
		apiObject.LineWidth = aws.String(v)
	}

	return apiObject
}

func expandLineChartMarkerStyleSettings(tfList []any) *awstypes.LineChartMarkerStyleSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartMarkerStyleSettings{}

	if v, ok := tfMap["marker_color"].(string); ok && v != "" {
		apiObject.MarkerColor = aws.String(v)
	}
	if v, ok := tfMap["marker_shape"].(string); ok && v != "" {
		apiObject.MarkerShape = awstypes.LineChartMarkerShape(v)
	}
	if v, ok := tfMap["marker_size"].(string); ok && v != "" {
		apiObject.MarkerSize = aws.String(v)
	}
	if v, ok := tfMap["marker_visibility"].(string); ok && v != "" {
		apiObject.MarkerVisibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandForecastConfigurations(tfList []any) []awstypes.ForecastConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ForecastConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandForecastConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandForecastConfiguration(tfMap map[string]any) *awstypes.ForecastConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ForecastConfiguration{}

	if v, ok := tfMap["forecast_properties"].([]any); ok && len(v) > 0 {
		apiObject.ForecastProperties = expandTimeBasedForecastProperties(v)
	}
	if v, ok := tfMap["scenario"].([]any); ok && len(v) > 0 {
		apiObject.Scenario = expandForecastScenario(v)
	}

	return apiObject
}

func expandTimeBasedForecastProperties(tfList []any) *awstypes.TimeBasedForecastProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TimeBasedForecastProperties{}

	if v, ok := tfMap["lower_boundary"].(float64); ok {
		apiObject.LowerBoundary = aws.Float64(v)
	}
	if v, ok := tfMap["periods_backward"].(int); ok {
		apiObject.PeriodsBackward = aws.Int32(int32(v))
	}
	if v, ok := tfMap["periods_forward"].(int); ok {
		apiObject.PeriodsForward = aws.Int32(int32(v))
	}
	if v, ok := tfMap["prediction_interval"].(int); ok {
		apiObject.PredictionInterval = aws.Int32(int32(v))
	}
	if v, ok := tfMap["seasonality"].(int); ok {
		apiObject.Seasonality = aws.Int32(int32(v))
	}
	if v, ok := tfMap["upper_boundary"].(float64); ok {
		apiObject.UpperBoundary = aws.Float64(v)
	}

	return apiObject
}

func expandForecastScenario(tfList []any) *awstypes.ForecastScenario {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ForecastScenario{}

	if v, ok := tfMap["what_if_point_scenario"].([]any); ok && len(v) > 0 {
		apiObject.WhatIfPointScenario = expandWhatIfPointScenario(v)
	}
	if v, ok := tfMap["what_if_range_scenario"].([]any); ok && len(v) > 0 {
		apiObject.WhatIfRangeScenario = expandWhatIfRangeScenario(v)
	}

	return apiObject
}

func expandWhatIfPointScenario(tfList []any) *awstypes.WhatIfPointScenario {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.WhatIfPointScenario{}

	if v, ok := tfMap["date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		apiObject.Date = aws.Time(t)
	}

	if v, ok := tfMap[names.AttrValue].(float64); ok {
		apiObject.Value = v
	}

	return apiObject
}

func expandWhatIfRangeScenario(tfList []any) *awstypes.WhatIfRangeScenario {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.WhatIfRangeScenario{}

	if v, ok := tfMap["end_date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		apiObject.EndDate = aws.Time(t)
	}
	if v, ok := tfMap["start_date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		apiObject.StartDate = aws.Time(t)
	}
	if v, ok := tfMap[names.AttrValue].(float64); ok {
		apiObject.Value = v
	}

	return apiObject
}

func expandLineSeriesAxisDisplayOptions(tfList []any) *awstypes.LineSeriesAxisDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineSeriesAxisDisplayOptions{}

	if v, ok := tfMap["axis_options"].([]any); ok && len(v) > 0 {
		apiObject.AxisOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["missing_data_configuration"].([]any); ok && len(v) > 0 {
		apiObject.MissingDataConfigurations = expandMissingDataConfigurations(v)
	}

	return apiObject
}

func expandMissingDataConfigurations(tfList []any) []awstypes.MissingDataConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.MissingDataConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandMissingDataConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandMissingDataConfiguration(tfMap map[string]any) *awstypes.MissingDataConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MissingDataConfiguration{}

	if v, ok := tfMap["treatment_option"].(string); ok && v != "" {
		apiObject.TreatmentOption = awstypes.MissingDataTreatmentOption(v)
	}

	return apiObject
}

func expandSeriesItems(tfList []any) []awstypes.SeriesItem {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SeriesItem

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandSeriesItem(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandSeriesItem(tfMap map[string]any) *awstypes.SeriesItem {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SeriesItem{}

	if v, ok := tfMap["data_field_series_item"].([]any); ok && len(v) > 0 {
		apiObject.DataFieldSeriesItem = expandDataFieldSeriesItem(v)
	}
	if v, ok := tfMap["field_series_item"].([]any); ok && len(v) > 0 {
		apiObject.FieldSeriesItem = expandFieldSeriesItem(v)
	}

	return apiObject
}

func expandDataFieldSeriesItem(tfList []any) *awstypes.DataFieldSeriesItem {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataFieldSeriesItem{}

	if v, ok := tfMap["axis_binding"].(string); ok && v != "" {
		apiObject.AxisBinding = awstypes.AxisBinding(v)
	}
	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["field_value"].(string); ok && v != "" {
		apiObject.FieldValue = aws.String(v)
	}
	if v, ok := tfMap["settings"].([]any); ok && len(v) > 0 {
		apiObject.Settings = expandLineChartSeriesSettings(v)
	}

	return apiObject
}

func expandFieldSeriesItem(tfList []any) *awstypes.FieldSeriesItem {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FieldSeriesItem{}

	if v, ok := tfMap["axis_binding"].(string); ok && v != "" {
		apiObject.AxisBinding = awstypes.AxisBinding(v)
	}
	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["settings"].([]any); ok && len(v) > 0 {
		apiObject.Settings = expandLineChartSeriesSettings(v)
	}

	return apiObject
}

func expandLineChartSeriesSettings(tfList []any) *awstypes.LineChartSeriesSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LineChartSeriesSettings{}

	if v, ok := tfMap["line_style_settings"].([]any); ok && len(v) > 0 {
		apiObject.LineStyleSettings = expandLineChartLineStyleSettings(v)
	}
	if v, ok := tfMap["marker_style_settings"].([]any); ok && len(v) > 0 {
		apiObject.MarkerStyleSettings = expandLineChartMarkerStyleSettings(v)
	}

	return apiObject
}

func flattenLineChartVisual(apiObject *awstypes.LineChartVisual) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"visual_id": aws.ToString(apiObject.VisualId),
	}

	if apiObject.Actions != nil {
		tfMap[names.AttrActions] = flattenVisualCustomAction(apiObject.Actions)
	}
	if apiObject.ChartConfiguration != nil {
		tfMap["chart_configuration"] = flattenLineChartConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.ColumnHierarchies != nil {
		tfMap["column_hierarchies"] = flattenColumnHierarchy(apiObject.ColumnHierarchies)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []any{tfMap}
}

func flattenLineChartConfiguration(apiObject *awstypes.LineChartConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ContributionAnalysisDefaults != nil {
		tfMap["contribution_analysis_defaults"] = flattenContributionAnalysisDefault(apiObject.ContributionAnalysisDefaults)
	}
	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.DefaultSeriesSettings != nil {
		tfMap["default_series_settings"] = flattenLineChartDefaultSeriesSettings(apiObject.DefaultSeriesSettings)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenLineChartFieldWells(apiObject.FieldWells)
	}
	if apiObject.ForecastConfigurations != nil {
		tfMap["forecast_configurations"] = flattenForecastConfiguration(apiObject.ForecastConfigurations)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.PrimaryYAxisDisplayOptions != nil {
		tfMap["primary_y_axis_display_options"] = flattenLineSeriesAxisDisplayOptions(apiObject.PrimaryYAxisDisplayOptions)
	}
	if apiObject.PrimaryYAxisLabelOptions != nil {
		tfMap["primary_y_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.PrimaryYAxisLabelOptions)
	}
	if apiObject.ReferenceLines != nil {
		tfMap["reference_lines"] = flattenReferenceLine(apiObject.ReferenceLines)
	}
	if apiObject.SecondaryYAxisDisplayOptions != nil {
		tfMap["secondary_y_axis_display_options"] = flattenLineSeriesAxisDisplayOptions(apiObject.SecondaryYAxisDisplayOptions)
	}
	if apiObject.SecondaryYAxisLabelOptions != nil {
		tfMap["secondary_y_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.SecondaryYAxisLabelOptions)
	}
	if apiObject.Series != nil {
		tfMap["series"] = flattenSeriesItem(apiObject.Series)
	}
	if apiObject.SmallMultiplesOptions != nil {
		tfMap["small_multiples_options"] = flattenSmallMultiplesOptions(apiObject.SmallMultiplesOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenLineChartSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}
	tfMap[names.AttrType] = apiObject.Type
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}
	if apiObject.XAxisDisplayOptions != nil {
		tfMap["x_axis_display_options"] = flattenAxisDisplayOptions(apiObject.XAxisDisplayOptions)
	}
	if apiObject.XAxisLabelOptions != nil {
		tfMap["x_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.XAxisLabelOptions)
	}

	return []any{tfMap}
}

func flattenLineChartDefaultSeriesSettings(apiObject *awstypes.LineChartDefaultSeriesSettings) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["axis_binding"] = apiObject.AxisBinding
	if apiObject.LineStyleSettings != nil {
		tfMap["line_style_settings"] = flattenLineChartLineStyleSettings(apiObject.LineStyleSettings)
	}
	if apiObject.MarkerStyleSettings != nil {
		tfMap["marker_style_settings"] = flattenLineChartMarkerStyleSettings(apiObject.MarkerStyleSettings)
	}

	return []any{tfMap}
}

func flattenLineChartLineStyleSettings(apiObject *awstypes.LineChartLineStyleSettings) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"line_interpolation": apiObject.LineInterpolation,
		"line_style":         apiObject.LineStyle,
		"line_visibility":    apiObject.LineVisibility,
	}

	if apiObject.LineWidth != nil {
		tfMap["line_width"] = aws.ToString(apiObject.LineWidth)
	}

	return []any{tfMap}
}

func flattenLineChartMarkerStyleSettings(apiObject *awstypes.LineChartMarkerStyleSettings) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"marker_shape":      apiObject.MarkerShape,
		"marker_visibility": apiObject.MarkerVisibility,
	}

	if apiObject.MarkerColor != nil {
		tfMap["marker_color"] = aws.ToString(apiObject.MarkerColor)
	}
	if apiObject.MarkerSize != nil {
		tfMap["marker_size"] = aws.ToString(apiObject.MarkerSize)
	}

	return []any{tfMap}
}

func flattenLineChartFieldWells(apiObject *awstypes.LineChartFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.LineChartAggregatedFieldWells != nil {
		tfMap["line_chart_aggregated_field_wells"] = flattenLineChartAggregatedFieldWells(apiObject.LineChartAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenLineChartAggregatedFieldWells(apiObject *awstypes.LineChartAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.Colors != nil {
		tfMap["colors"] = flattenDimensionFields(apiObject.Colors)
	}
	if apiObject.SmallMultiples != nil {
		tfMap["small_multiples"] = flattenDimensionFields(apiObject.SmallMultiples)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenForecastConfiguration(apiObjects []awstypes.ForecastConfiguration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.ForecastProperties != nil {
			tfMap["forecast_properties"] = flattenTimeBasedForecastProperties(apiObject.ForecastProperties)
		}
		if apiObject.Scenario != nil {
			tfMap["scenario"] = flattenForecastScenario(apiObject.Scenario)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTimeBasedForecastProperties(apiObject *awstypes.TimeBasedForecastProperties) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.LowerBoundary != nil {
		tfMap["lower_boundary"] = aws.ToFloat64(apiObject.LowerBoundary)
	}
	if apiObject.PeriodsBackward != nil {
		tfMap["periods_backward"] = aws.ToInt32(apiObject.PeriodsBackward)
	}
	if apiObject.PeriodsForward != nil {
		tfMap["periods_forward"] = aws.ToInt32(apiObject.PeriodsForward)
	}
	if apiObject.PredictionInterval != nil {
		tfMap["prediction_interval"] = aws.ToInt32(apiObject.PredictionInterval)
	}
	if apiObject.Seasonality != nil {
		tfMap["seasonality"] = aws.ToInt32(apiObject.Seasonality)
	}
	if apiObject.UpperBoundary != nil {
		tfMap["upper_boundary"] = aws.ToFloat64(apiObject.UpperBoundary)
	}

	return []any{tfMap}
}

func flattenForecastScenario(apiObject *awstypes.ForecastScenario) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.WhatIfPointScenario != nil {
		tfMap["what_if_point_scenario"] = flattenWhatIfPointScenario(apiObject.WhatIfPointScenario)
	}
	if apiObject.WhatIfRangeScenario != nil {
		tfMap["what_if_range_scenario"] = flattenWhatIfRangeScenario(apiObject.WhatIfRangeScenario)
	}

	return []any{tfMap}
}

func flattenWhatIfPointScenario(apiObject *awstypes.WhatIfPointScenario) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Date != nil {
		tfMap["date"] = aws.ToTime(apiObject.Date).Format(time.RFC3339)
	}
	tfMap[names.AttrValue] = apiObject.Value

	return []any{tfMap}
}

func flattenWhatIfRangeScenario(apiObject *awstypes.WhatIfRangeScenario) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.EndDate != nil {
		tfMap["end_date"] = aws.ToTime(apiObject.EndDate).Format(time.RFC3339)
	}
	if apiObject.StartDate != nil {
		tfMap["start_date"] = aws.ToTime(apiObject.StartDate).Format(time.RFC3339)
	}
	tfMap[names.AttrValue] = apiObject.Value

	return []any{tfMap}
}

func flattenLineSeriesAxisDisplayOptions(apiObject *awstypes.LineSeriesAxisDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.AxisOptions != nil {
		tfMap["axis_options"] = flattenAxisDisplayOptions(apiObject.AxisOptions)
	}
	if apiObject.MissingDataConfigurations != nil {
		tfMap["missing_data_configurations"] = flattenMissingDataConfiguration(apiObject.MissingDataConfigurations)
	}

	return []any{tfMap}
}

func flattenMissingDataConfiguration(apiObjects []awstypes.MissingDataConfiguration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"treatment_option": apiObject.TreatmentOption,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSeriesItem(apiObjects []awstypes.SeriesItem) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.DataFieldSeriesItem != nil {
			tfMap["data_field_series_item"] = flattenDataFieldSeriesItem(apiObject.DataFieldSeriesItem)
		}
		if apiObject.FieldSeriesItem != nil {
			tfMap["field_series_item"] = flattenFieldSeriesItem(apiObject.FieldSeriesItem)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataFieldSeriesItem(apiObject *awstypes.DataFieldSeriesItem) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"axis_binding": apiObject.AxisBinding,
	}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FieldValue != nil {
		tfMap["field_value"] = aws.ToString(apiObject.FieldValue)
	}
	if apiObject.Settings != nil {
		tfMap["settings"] = flattenLineChartSeriesSettings(apiObject.Settings)
	}

	return []any{tfMap}
}

func flattenLineChartSeriesSettings(apiObject *awstypes.LineChartSeriesSettings) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.LineStyleSettings != nil {
		tfMap["line_style_settings"] = flattenLineChartLineStyleSettings(apiObject.LineStyleSettings)
	}
	if apiObject.MarkerStyleSettings != nil {
		tfMap["marker_style_settings"] = flattenLineChartMarkerStyleSettings(apiObject.MarkerStyleSettings)
	}

	return []any{tfMap}
}

func flattenFieldSeriesItem(apiObject *awstypes.FieldSeriesItem) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"axis_binding": apiObject.AxisBinding,
	}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.Settings != nil {
		tfMap["settings"] = flattenLineChartSeriesSettings(apiObject.Settings)
	}

	return []any{tfMap}
}

func flattenLineChartSortConfiguration(apiObject *awstypes.LineChartSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CategoryItemsLimitConfiguration != nil {
		tfMap["category_items_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.CategoryItemsLimitConfiguration)
	}
	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}
	if apiObject.ColorItemsLimitConfiguration != nil {
		tfMap["color_items_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.ColorItemsLimitConfiguration)
	}
	if apiObject.SmallMultiplesLimitConfiguration != nil {
		tfMap["small_multiples_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.SmallMultiplesLimitConfiguration)
	}
	if apiObject.SmallMultiplesSort != nil {
		tfMap["small_multiples_sort"] = flattenFieldSortOptions(apiObject.SmallMultiplesSort)
	}

	return []any{tfMap}
}
