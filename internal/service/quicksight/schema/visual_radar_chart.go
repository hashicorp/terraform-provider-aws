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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func radarChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"alternate_band_colors_visibility": stringSchema(false, enum.Validate[types.Visibility]()),
							"alternate_band_even_color":        stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
							"alternate_band_odd_color":         stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
							"base_series_settings": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartSeriesSettings.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"area_style_settings": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartAreaStyleSettings.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
												},
											},
										},
									},
								},
							},
							"category_axis":          axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"category_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"color_axis":             axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"color_label_options":    chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"radar_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"category": dimensionFieldSchema(1),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"color":    dimensionFieldSchema(1),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"values":   measureFieldSchema(measureFieldsMaxItems20), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend": legendOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"shape":  stringSchema(false, enum.Validate[types.RadarChartShape]()),
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":        fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"color_items_limit":    itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"color_sort":           fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"start_angle":    floatSchema(false, validation.FloatBetween(-360, 360)),
							"visual_palette": visualPaletteSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
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

func expandRadarChartVisual(tfList []interface{}) *types.RadarChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.RadarChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandRadarChartConfiguration(v)
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

func expandRadarChartConfiguration(tfList []interface{}) *types.RadarChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.RadarChartConfiguration{}

	if v, ok := tfMap["alternate_band_colors_visibility"].(string); ok && v != "" {
		config.AlternateBandColorsVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["alternate_band_even_color"].(string); ok && v != "" {
		config.AlternateBandEvenColor = aws.String(v)
	}
	if v, ok := tfMap["alternate_band_odd_color"].(string); ok && v != "" {
		config.AlternateBandOddColor = aws.String(v)
	}
	if v, ok := tfMap["shape"].(string); ok && v != "" {
		config.Shape = types.RadarChartShape(v)
	}
	if v, ok := tfMap["start_angle"].(float64); ok {
		config.StartAngle = aws.Float64(v)
	}
	if v, ok := tfMap["base_series_settings"].([]interface{}); ok && len(v) > 0 {
		config.BaseSeriesSettings = expandRadarChartSeriesSettings(v)
	}
	if v, ok := tfMap["category_axis"].([]interface{}); ok && len(v) > 0 {
		config.CategoryAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["category_label_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["color_axis"].([]interface{}); ok && len(v) > 0 {
		config.ColorAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["color_label_options"].([]interface{}); ok && len(v) > 0 {
		config.ColorLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandRadarChartFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandRadarChartSortConfiguration(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}

	return config
}

func expandRadarChartFieldWells(tfList []interface{}) *types.RadarChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.RadarChartFieldWells{}

	if v, ok := tfMap["radar_chart_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.RadarChartAggregatedFieldWells = expandRadarChartAggregatedFieldWells(v)
	}

	return config
}

func expandRadarChartAggregatedFieldWells(tfList []interface{}) *types.RadarChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.RadarChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		config.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["colors"].([]interface{}); ok && len(v) > 0 {
		config.Color = expandDimensionFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandRadarChartSortConfiguration(tfList []interface{}) *types.RadarChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.RadarChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["color_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.ColorItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["color_sort"].([]interface{}); ok && len(v) > 0 {
		config.ColorSort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandRadarChartSeriesSettings(tfList []interface{}) *types.RadarChartSeriesSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.RadarChartSeriesSettings{}

	if v, ok := tfMap["area_style_settings"].([]interface{}); ok && len(v) > 0 {
		config.AreaStyleSettings = expandRadarChartAreaStyleSettings(v)
	}

	return config
}

func expandRadarChartAreaStyleSettings(tfList []interface{}) *types.RadarChartAreaStyleSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.RadarChartAreaStyleSettings{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = types.Visibility(v)
	}

	return config
}

func flattenRadarChartVisual(apiObject *types.RadarChartVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id": aws.ToString(apiObject.VisualId),
	}
	if apiObject.Actions != nil {
		tfMap["actions"] = flattenVisualCustomAction(apiObject.Actions)
	}
	if apiObject.ChartConfiguration != nil {
		tfMap["chart_configuration"] = flattenRadarChartConfiguration(apiObject.ChartConfiguration)
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

func flattenRadarChartConfiguration(apiObject *types.RadarChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["alternate_band_colors_visibility"] = types.Visibility(apiObject.AlternateBandColorsVisibility)

	if apiObject.AlternateBandEvenColor != nil {
		tfMap["alternate_band_even_color"] = aws.ToString(apiObject.AlternateBandEvenColor)
	}
	if apiObject.AlternateBandOddColor != nil {
		tfMap["alternate_band_odd_color"] = aws.ToString(apiObject.AlternateBandOddColor)
	}
	if apiObject.BaseSeriesSettings != nil {
		tfMap["base_series_settings"] = flattenRadarChartSeriesSettings(apiObject.BaseSeriesSettings)
	}
	if apiObject.CategoryAxis != nil {
		tfMap["category_axis"] = flattenAxisDisplayOptions(apiObject.CategoryAxis)
	}
	if apiObject.CategoryLabelOptions != nil {
		tfMap["category_label_options"] = flattenChartAxisLabelOptions(apiObject.CategoryLabelOptions)
	}
	if apiObject.ColorAxis != nil {
		tfMap["color_axis"] = flattenAxisDisplayOptions(apiObject.ColorAxis)
	}
	if apiObject.ColorLabelOptions != nil {
		tfMap["color_label_options"] = flattenChartAxisLabelOptions(apiObject.ColorLabelOptions)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenRadarChartFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}

	tfMap["shape"] = types.RadarChartShape(apiObject.Shape)

	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenRadarChartSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.StartAngle != nil {
		tfMap["start_angle"] = aws.ToFloat64(apiObject.StartAngle)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}

	return []interface{}{tfMap}
}

func flattenRadarChartSeriesSettings(apiObject *types.RadarChartSeriesSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AreaStyleSettings != nil {
		tfMap["area_style_settings"] = flattenRadarChartAreaStyleSettings(apiObject.AreaStyleSettings)
	}

	return []interface{}{tfMap}
}

func flattenRadarChartAreaStyleSettings(apiObject *types.RadarChartAreaStyleSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenRadarChartFieldWells(apiObject *types.RadarChartFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.RadarChartAggregatedFieldWells != nil {
		tfMap["radar_chart_aggregated_field_wells"] = flattenRadarChartAggregatedFieldWells(apiObject.RadarChartAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenRadarChartAggregatedFieldWells(apiObject *types.RadarChartAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.Color != nil {
		tfMap["color"] = flattenDimensionFields(apiObject.Color)
	}
	if apiObject.Values != nil {
		tfMap["values"] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenRadarChartSortConfiguration(apiObject *types.RadarChartSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryItemsLimit != nil {
		tfMap["category_items_limit"] = flattenItemsLimitConfiguration(apiObject.CategoryItemsLimit)
	}
	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}
	if apiObject.ColorItemsLimit != nil {
		tfMap["color_items_limit"] = flattenItemsLimitConfiguration(apiObject.ColorItemsLimit)
	}
	if apiObject.ColorSort != nil {
		tfMap["color_sort"] = flattenFieldSortOptions(apiObject.ColorSort)
	}

	return []interface{}{tfMap}
}
