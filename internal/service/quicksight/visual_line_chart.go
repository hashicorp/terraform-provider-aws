package quicksight

import (
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func lineChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LineChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
													"values":          measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
																"value": {
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
																"value": {
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
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
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
							"type":                   stringSchema(false, validation.StringInSlice(quicksight.LineChartType_Values(), false)),
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
				"marker_color": stringSchema(false, validation.StringMatch(regexp.MustCompile(`^#[A-F0-9]{6}$`), "")),
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
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
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

	if v, ok := tfMap["type"].(string); ok && v != "" {
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
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["periods_backward"].(int64); ok {
		options.PeriodsBackward = aws.Int64(v)
	}
	if v, ok := tfMap["periods_forward"].(int64); ok {
		options.PeriodsForward = aws.Int64(v)
	}
	if v, ok := tfMap["prediction_interval"].(int64); ok {
		options.PredictionInterval = aws.Int64(v)
	}
	if v, ok := tfMap["seasonality"].(int64); ok {
		options.Seasonality = aws.Int64(v)
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

	if v, ok := tfMap["value"].(float64); ok {
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
	if v, ok := tfMap["value"].(float64); ok {
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
