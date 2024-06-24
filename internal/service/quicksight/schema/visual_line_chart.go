// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
										"axis_binding":          stringSchema(false, validation.StringInSlice(quicksight.AxisBinding_Values(), false)),
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
													"periods_backward":    intSchema(false, validation.IntBetween(0, 1000)),
													"periods_forward":     intSchema(false, validation.IntBetween(1, 1000)),
													"prediction_interval": intSchema(false, validation.IntBetween(50, 95)),
													"seasonality":         intSchema(false, validation.IntBetween(1, 180)),
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
																"date": stringSchema(true, verify.ValidUTCTimestamp),
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
																"end_date":   stringSchema(true, verify.ValidUTCTimestamp),
																"start_date": stringSchema(true, verify.ValidUTCTimestamp),
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
													"treatment_option": stringSchema(false, validation.StringInSlice(quicksight.MissingDataTreatmentOption_Values(), false)),
												},
											},
										},
									},
								},
							},
							"primary_y_axis_label_options": chartAxisLabelOptionsSchema(),               // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"reference_lines":              referenceLineSchema(referenceLinesMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
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
													"treatment_option": stringSchema(false, validation.StringInSlice(quicksight.MissingDataTreatmentOption_Values(), false)),
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
													"axis_binding": stringSchema(true, validation.StringInSlice(quicksight.AxisBinding_Values(), false)),
													"field_id":     stringSchema(true, validation.StringLenBetween(1, 512)),
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
													"axis_binding": stringSchema(true, validation.StringInSlice(quicksight.AxisBinding_Values(), false)),
													"field_id":     stringSchema(true, validation.StringLenBetween(1, 512)),
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
										"category_items_limit_configuration":  itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":                       fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"color_items_limit_configuration":     itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"small_multiples_limit_configuration": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"small_multiples_sort":                fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"tooltip":                tooltipOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							names.AttrType:           stringOptionalComputedSchema(validation.StringInSlice(quicksight.LineChartType_Values(), false)),
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

func lineChartLineStyleSettingsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartLineStyleSettings.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"line_interpolation": stringSchema(false, validation.StringInSlice(quicksight.LineInterpolation_Values(), false)),
				"line_style":         stringSchema(false, validation.StringInSlice(quicksight.LineChartLineStyle_Values(), false)),
				"line_visibility":    stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
				"line_width": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func lineChartMarkerStyleSettingsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartMarkerStyleSettings.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"marker_color": stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
				"marker_shape": stringSchema(false, validation.StringInSlice(quicksight.LineChartMarkerShape_Values(), false)),
				"marker_size": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"marker_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
			},
		},
	}
}

func expandLineChartVisual(tfList []interface{}) *quicksight.LineChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.LineChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandLineChartConfiguration(v)
	}
	if v, ok := tfMap["column_hierarchies"].([]interface{}); ok && len(v) > 0 {
		visual.ColumnHierarchies = expandColumnHierarchies(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		visual.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		visual.Title = expandVisualTitleLabelOptions(v)
	}

	return visual
}

func expandLineChartConfiguration(tfList []interface{}) *quicksight.LineChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.LineChartConfiguration{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		config.Type = aws.String(v)
	}
	if v, ok := tfMap["contribution_analysis_defaults"].([]interface{}); ok && len(v) > 0 {
		config.ContributionAnalysisDefaults = expandContributionAnalysisDefaults(v)
	}
	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		config.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["default_series_settings"].([]interface{}); ok && len(v) > 0 {
		config.DefaultSeriesSettings = expandLineChartDefaultSeriesSettings(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandLineChartFieldWells(v)
	}
	if v, ok := tfMap["forecast_configurations"].([]interface{}); ok && len(v) > 0 {
		config.ForecastConfigurations = expandForecastConfigurations(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.PrimaryYAxisDisplayOptions = expandLineSeriesAxisDisplayOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.PrimaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["reference_lines"].([]interface{}); ok && len(v) > 0 {
		config.ReferenceLines = expandReferenceLines(v)
	}
	if v, ok := tfMap["secondary_y_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.SecondaryYAxisDisplayOptions = expandLineSeriesAxisDisplayOptions(v)
	}
	if v, ok := tfMap["secondary_y_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.SecondaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["series"].([]interface{}); ok && len(v) > 0 {
		config.Series = expandSeriesItems(v)
	}
	if v, ok := tfMap["small_multiples_options"].([]interface{}); ok && len(v) > 0 {
		config.SmallMultiplesOptions = expandSmallMultiplesOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandLineChartSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}
	if v, ok := tfMap["x_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.XAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["x_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.XAxisLabelOptions = expandChartAxisLabelOptions(v)
	}

	return config
}

func expandLineChartFieldWells(tfList []interface{}) *quicksight.LineChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.LineChartFieldWells{}

	if v, ok := tfMap["line_chart_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.LineChartAggregatedFieldWells = expandLineChartAggregatedFieldWells(v)
	}

	return config
}

func expandLineChartAggregatedFieldWells(tfList []interface{}) *quicksight.LineChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.LineChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		config.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["colors"].([]interface{}); ok && len(v) > 0 {
		config.Colors = expandDimensionFields(v)
	}
	if v, ok := tfMap["small_multiples"].([]interface{}); ok && len(v) > 0 {
		config.SmallMultiples = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandLineChartSortConfiguration(tfList []interface{}) *quicksight.LineChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.LineChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.CategoryItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["color_items_limit_configuration"].([]interface{}); ok && len(v) > 0 {
		config.ColorItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["small_multiples_limit_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SmallMultiplesLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["small_multiples_sort"].([]interface{}); ok && len(v) > 0 {
		config.SmallMultiplesSort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandLineChartDefaultSeriesSettings(tfList []interface{}) *quicksight.LineChartDefaultSeriesSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.LineChartDefaultSeriesSettings{}

	if v, ok := tfMap["axis_binding"].(string); ok && v != "" {
		options.AxisBinding = aws.String(v)
	}
	if v, ok := tfMap["line_style_settings"].([]interface{}); ok && len(v) > 0 {
		options.LineStyleSettings = expandLineChartLineStyleSettings(v)
	}
	if v, ok := tfMap["marker_style_settings"].([]interface{}); ok && len(v) > 0 {
		options.MarkerStyleSettings = expandLineChartMarkerStyleSettings(v)
	}

	return options
}

func expandLineChartLineStyleSettings(tfList []interface{}) *quicksight.LineChartLineStyleSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.LineChartLineStyleSettings{}

	if v, ok := tfMap["line_interpolation"].(string); ok && v != "" {
		options.LineInterpolation = aws.String(v)
	}
	if v, ok := tfMap["line_style"].(string); ok && v != "" {
		options.LineStyle = aws.String(v)
	}
	if v, ok := tfMap["line_visibility"].(string); ok && v != "" {
		options.LineVisibility = aws.String(v)
	}
	if v, ok := tfMap["line_width"].(string); ok && v != "" {
		options.LineWidth = aws.String(v)
	}

	return options
}

func expandLineChartMarkerStyleSettings(tfList []interface{}) *quicksight.LineChartMarkerStyleSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.LineChartMarkerStyleSettings{}

	if v, ok := tfMap["marker_color"].(string); ok && v != "" {
		options.MarkerColor = aws.String(v)
	}
	if v, ok := tfMap["marker_shape"].(string); ok && v != "" {
		options.MarkerShape = aws.String(v)
	}
	if v, ok := tfMap["marker_size"].(string); ok && v != "" {
		options.MarkerSize = aws.String(v)
	}
	if v, ok := tfMap["marker_visibility"].(string); ok && v != "" {
		options.MarkerVisibility = aws.String(v)
	}

	return options
}

func expandForecastConfigurations(tfList []interface{}) []*quicksight.ForecastConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.ForecastConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandForecastConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, config)
	}

	return configs
}

func expandForecastConfiguration(tfMap map[string]interface{}) *quicksight.ForecastConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &quicksight.ForecastConfiguration{}

	if v, ok := tfMap["forecast_properties"].([]interface{}); ok && len(v) > 0 {
		config.ForecastProperties = expandTimeBasedForecastProperties(v)
	}
	if v, ok := tfMap["scenario"].([]interface{}); ok && len(v) > 0 {
		config.Scenario = expandForecastScenario(v)
	}

	return config
}

func expandTimeBasedForecastProperties(tfList []interface{}) *quicksight.TimeBasedForecastProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TimeBasedForecastProperties{}

	if v, ok := tfMap["lower_boundary"].(float64); ok {
		options.LowerBoundary = aws.Float64(v)
	}
	if v, ok := tfMap["periods_backward"].(int); ok {
		options.PeriodsBackward = aws.Int64(int64(v))
	}
	if v, ok := tfMap["periods_forward"].(int); ok {
		options.PeriodsForward = aws.Int64(int64(v))
	}
	if v, ok := tfMap["prediction_interval"].(int); ok {
		options.PredictionInterval = aws.Int64(int64(v))
	}
	if v, ok := tfMap["seasonality"].(int); ok {
		options.Seasonality = aws.Int64(int64(v))
	}
	if v, ok := tfMap["upper_boundary"].(float64); ok {
		options.UpperBoundary = aws.Float64(v)
	}
	return options
}

func expandForecastScenario(tfList []interface{}) *quicksight.ForecastScenario {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	scenario := &quicksight.ForecastScenario{}

	if v, ok := tfMap["what_if_point_scenario"].([]interface{}); ok && len(v) > 0 {
		scenario.WhatIfPointScenario = expandWhatIfPointScenario(v)
	}
	if v, ok := tfMap["what_if_range_scenario"].([]interface{}); ok && len(v) > 0 {
		scenario.WhatIfRangeScenario = expandWhatIfRangeScenario(v)
	}

	return scenario
}

func expandWhatIfPointScenario(tfList []interface{}) *quicksight.WhatIfPointScenario {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	scenario := &quicksight.WhatIfPointScenario{}

	if v, ok := tfMap["date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		scenario.Date = aws.Time(t)
	}

	if v, ok := tfMap[names.AttrValue].(float64); ok {
		scenario.Value = aws.Float64(v)
	}

	return scenario
}

func expandWhatIfRangeScenario(tfList []interface{}) *quicksight.WhatIfRangeScenario {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	scenario := &quicksight.WhatIfRangeScenario{}

	if v, ok := tfMap["end_date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		scenario.EndDate = aws.Time(t)
	}
	if v, ok := tfMap["start_date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		scenario.StartDate = aws.Time(t)
	}
	if v, ok := tfMap[names.AttrValue].(float64); ok {
		scenario.Value = aws.Float64(v)
	}

	return scenario
}

func expandLineSeriesAxisDisplayOptions(tfList []interface{}) *quicksight.LineSeriesAxisDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.LineSeriesAxisDisplayOptions{}

	if v, ok := tfMap["axis_options"].([]interface{}); ok && len(v) > 0 {
		options.AxisOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["missing_data_configuration"].([]interface{}); ok && len(v) > 0 {
		options.MissingDataConfigurations = expandMissingDataConfigurations(v)
	}

	return options
}

func expandMissingDataConfigurations(tfList []interface{}) []*quicksight.MissingDataConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.MissingDataConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandMissingDataConfiguration(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandMissingDataConfiguration(tfMap map[string]interface{}) *quicksight.MissingDataConfiguration {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.MissingDataConfiguration{}

	if v, ok := tfMap["treatment_option"].(string); ok && v != "" {
		options.TreatmentOption = aws.String(v)
	}

	return options
}

func expandSeriesItems(tfList []interface{}) []*quicksight.SeriesItem {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.SeriesItem
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandSeriesItem(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandSeriesItem(tfMap map[string]interface{}) *quicksight.SeriesItem {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.SeriesItem{}

	if v, ok := tfMap["data_field_series_item"].([]interface{}); ok && len(v) > 0 {
		options.DataFieldSeriesItem = expandDataFieldSeriesItem(v)
	}
	if v, ok := tfMap["field_series_item"].([]interface{}); ok && len(v) > 0 {
		options.FieldSeriesItem = expandFieldSeriesItem(v)
	}

	return options
}

func expandDataFieldSeriesItem(tfList []interface{}) *quicksight.DataFieldSeriesItem {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.DataFieldSeriesItem{}

	if v, ok := tfMap["axis_binding"].(string); ok && v != "" {
		options.AxisBinding = aws.String(v)
	}
	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["field_value"].(string); ok && v != "" {
		options.FieldValue = aws.String(v)
	}

	if v, ok := tfMap["settings"].([]interface{}); ok && len(v) > 0 {
		options.Settings = expandLineChartSeriesSettings(v)
	}

	return options
}

func expandFieldSeriesItem(tfList []interface{}) *quicksight.FieldSeriesItem {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.FieldSeriesItem{}

	if v, ok := tfMap["axis_binding"].(string); ok && v != "" {
		options.AxisBinding = aws.String(v)
	}
	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["settings"].([]interface{}); ok && len(v) > 0 {
		options.Settings = expandLineChartSeriesSettings(v)
	}

	return options
}

func expandLineChartSeriesSettings(tfList []interface{}) *quicksight.LineChartSeriesSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.LineChartSeriesSettings{}

	if v, ok := tfMap["line_style_settings"].([]interface{}); ok && len(v) > 0 {
		options.LineStyleSettings = expandLineChartLineStyleSettings(v)
	}
	if v, ok := tfMap["marker_style_settings"].([]interface{}); ok && len(v) > 0 {
		options.MarkerStyleSettings = expandLineChartMarkerStyleSettings(v)
	}

	return options
}

func flattenLineChartVisual(apiObject *quicksight.LineChartVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id": aws.StringValue(apiObject.VisualId),
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

	return []interface{}{tfMap}
}

func flattenLineChartConfiguration(apiObject *quicksight.LineChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}
	if apiObject.XAxisDisplayOptions != nil {
		tfMap["x_axis_display_options"] = flattenAxisDisplayOptions(apiObject.XAxisDisplayOptions)
	}
	if apiObject.XAxisLabelOptions != nil {
		tfMap["x_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.XAxisLabelOptions)
	}

	return []interface{}{tfMap}
}

func flattenLineChartDefaultSeriesSettings(apiObject *quicksight.LineChartDefaultSeriesSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AxisBinding != nil {
		tfMap["axis_binding"] = aws.StringValue(apiObject.AxisBinding)
	}
	if apiObject.LineStyleSettings != nil {
		tfMap["line_style_settings"] = flattenLineChartLineStyleSettings(apiObject.LineStyleSettings)
	}
	if apiObject.MarkerStyleSettings != nil {
		tfMap["marker_style_settings"] = flattenLineChartMarkerStyleSettings(apiObject.MarkerStyleSettings)
	}

	return []interface{}{tfMap}
}

func flattenLineChartLineStyleSettings(apiObject *quicksight.LineChartLineStyleSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LineInterpolation != nil {
		tfMap["line_interpolation"] = aws.StringValue(apiObject.LineInterpolation)
	}
	if apiObject.LineStyle != nil {
		tfMap["line_style"] = aws.StringValue(apiObject.LineStyle)
	}
	if apiObject.LineVisibility != nil {
		tfMap["line_visibility"] = aws.StringValue(apiObject.LineVisibility)
	}
	if apiObject.LineWidth != nil {
		tfMap["line_width"] = aws.StringValue(apiObject.LineWidth)
	}

	return []interface{}{tfMap}
}

func flattenLineChartMarkerStyleSettings(apiObject *quicksight.LineChartMarkerStyleSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.MarkerColor != nil {
		tfMap["marker_color"] = aws.StringValue(apiObject.MarkerColor)
	}
	if apiObject.MarkerShape != nil {
		tfMap["marker_shape"] = aws.StringValue(apiObject.MarkerShape)
	}
	if apiObject.MarkerSize != nil {
		tfMap["marker_size"] = aws.StringValue(apiObject.MarkerSize)
	}
	if apiObject.MarkerVisibility != nil {
		tfMap["marker_visibility"] = aws.StringValue(apiObject.MarkerVisibility)
	}

	return []interface{}{tfMap}
}

func flattenLineChartFieldWells(apiObject *quicksight.LineChartFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LineChartAggregatedFieldWells != nil {
		tfMap["line_chart_aggregated_field_wells"] = flattenLineChartAggregatedFieldWells(apiObject.LineChartAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenLineChartAggregatedFieldWells(apiObject *quicksight.LineChartAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenForecastConfiguration(apiObject []*quicksight.ForecastConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.ForecastProperties != nil {
			tfMap["forecast_properties"] = flattenTimeBasedForecastProperties(config.ForecastProperties)
		}
		if config.Scenario != nil {
			tfMap["scenario"] = flattenForecastScenario(config.Scenario)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTimeBasedForecastProperties(apiObject *quicksight.TimeBasedForecastProperties) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LowerBoundary != nil {
		tfMap["lower_boundary"] = aws.Float64Value(apiObject.LowerBoundary)
	}
	if apiObject.PeriodsBackward != nil {
		tfMap["periods_backward"] = aws.Int64Value(apiObject.PeriodsBackward)
	}
	if apiObject.PeriodsForward != nil {
		tfMap["periods_forward"] = aws.Int64Value(apiObject.PeriodsForward)
	}
	if apiObject.PredictionInterval != nil {
		tfMap["prediction_interval"] = aws.Int64Value(apiObject.PredictionInterval)
	}
	if apiObject.Seasonality != nil {
		tfMap["seasonality"] = aws.Int64Value(apiObject.Seasonality)
	}
	if apiObject.UpperBoundary != nil {
		tfMap["upper_boundary"] = aws.Float64Value(apiObject.UpperBoundary)
	}

	return []interface{}{tfMap}
}

func flattenForecastScenario(apiObject *quicksight.ForecastScenario) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.WhatIfPointScenario != nil {
		tfMap["what_if_point_scenario"] = flattenWhatIfPointScenario(apiObject.WhatIfPointScenario)
	}
	if apiObject.WhatIfRangeScenario != nil {
		tfMap["what_if_range_scenario"] = flattenWhatIfRangeScenario(apiObject.WhatIfRangeScenario)
	}

	return []interface{}{tfMap}
}

func flattenWhatIfPointScenario(apiObject *quicksight.WhatIfPointScenario) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Date != nil {
		tfMap["date"] = aws.TimeValue(apiObject.Date).Format(time.RFC3339)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = aws.Float64Value(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenWhatIfRangeScenario(apiObject *quicksight.WhatIfRangeScenario) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.EndDate != nil {
		tfMap["end_date"] = aws.TimeValue(apiObject.EndDate).Format(time.RFC3339)
	}
	if apiObject.StartDate != nil {
		tfMap["start_date"] = aws.TimeValue(apiObject.StartDate).Format(time.RFC3339)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = aws.Float64Value(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenLineSeriesAxisDisplayOptions(apiObject *quicksight.LineSeriesAxisDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AxisOptions != nil {
		tfMap["axis_options"] = flattenAxisDisplayOptions(apiObject.AxisOptions)
	}
	if apiObject.MissingDataConfigurations != nil {
		tfMap["missing_data_configurations"] = flattenMissingDataConfiguration(apiObject.MissingDataConfigurations)
	}

	return []interface{}{tfMap}
}

func flattenMissingDataConfiguration(apiObject []*quicksight.MissingDataConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.TreatmentOption != nil {
			tfMap["treatment_option"] = aws.StringValue(config.TreatmentOption)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSeriesItem(apiObject []*quicksight.SeriesItem) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.DataFieldSeriesItem != nil {
			tfMap["data_field_series_item"] = flattenDataFieldSeriesItem(config.DataFieldSeriesItem)
		}
		if config.FieldSeriesItem != nil {
			tfMap["field_series_item"] = flattenFieldSeriesItem(config.FieldSeriesItem)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataFieldSeriesItem(apiObject *quicksight.DataFieldSeriesItem) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AxisBinding != nil {
		tfMap["axis_binding"] = aws.StringValue(apiObject.AxisBinding)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.FieldValue != nil {
		tfMap["field_value"] = aws.StringValue(apiObject.FieldValue)
	}
	if apiObject.Settings != nil {
		tfMap["settings"] = flattenLineChartSeriesSettings(apiObject.Settings)
	}

	return []interface{}{tfMap}
}

func flattenLineChartSeriesSettings(apiObject *quicksight.LineChartSeriesSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LineStyleSettings != nil {
		tfMap["line_style_settings"] = flattenLineChartLineStyleSettings(apiObject.LineStyleSettings)
	}
	if apiObject.MarkerStyleSettings != nil {
		tfMap["marker_style_settings"] = flattenLineChartMarkerStyleSettings(apiObject.MarkerStyleSettings)
	}

	return []interface{}{tfMap}
}

func flattenFieldSeriesItem(apiObject *quicksight.FieldSeriesItem) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AxisBinding != nil {
		tfMap["axis_binding"] = aws.StringValue(apiObject.AxisBinding)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.Settings != nil {
		tfMap["settings"] = flattenLineChartSeriesSettings(apiObject.Settings)
	}

	return []interface{}{tfMap}
}

func flattenLineChartSortConfiguration(apiObject *quicksight.LineChartSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}
