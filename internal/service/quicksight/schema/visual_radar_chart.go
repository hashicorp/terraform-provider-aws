// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func radarChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"alternate_band_colors_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
							"alternate_band_even_color":        hexColorSchema(attrOptional),
							"alternate_band_odd_color":         hexColorSchema(attrOptional),
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
													"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
													"category":       dimensionFieldSchema(1),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"color":          dimensionFieldSchema(1),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(measureFieldsMaxItems20), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend": legendOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"shape":  stringEnumSchema[awstypes.RadarChartShape](attrOptional),
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RadarChartSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit": itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":        fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"color_items_limit":    itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"color_sort":           fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"start_angle":    floatBetweenSchema(attrOptional, -360, 360),
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

func expandRadarChartVisual(tfList []any) *awstypes.RadarChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RadarChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandRadarChartConfiguration(v)
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

func expandRadarChartConfiguration(tfList []any) *awstypes.RadarChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RadarChartConfiguration{}

	if v, ok := tfMap["alternate_band_colors_visibility"].(string); ok && v != "" {
		apiObject.AlternateBandColorsVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["alternate_band_even_color"].(string); ok && v != "" {
		apiObject.AlternateBandEvenColor = aws.String(v)
	}
	if v, ok := tfMap["alternate_band_odd_color"].(string); ok && v != "" {
		apiObject.AlternateBandOddColor = aws.String(v)
	}
	if v, ok := tfMap["shape"].(string); ok && v != "" {
		apiObject.Shape = awstypes.RadarChartShape(v)
	}
	if v, ok := tfMap["start_angle"].(float64); ok {
		apiObject.StartAngle = aws.Float64(v)
	}
	if v, ok := tfMap["base_series_settings"].([]any); ok && len(v) > 0 {
		apiObject.BaseSeriesSettings = expandRadarChartSeriesSettings(v)
	}
	if v, ok := tfMap["category_axis"].([]any); ok && len(v) > 0 {
		apiObject.CategoryAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["category_label_options"].([]any); ok && len(v) > 0 {
		apiObject.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["color_axis"].([]any); ok && len(v) > 0 {
		apiObject.ColorAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["color_label_options"].([]any); ok && len(v) > 0 {
		apiObject.ColorLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandRadarChartFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandRadarChartSortConfiguration(v)
	}
	if v, ok := tfMap["visual_palette"].([]any); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}

	return apiObject
}

func expandRadarChartFieldWells(tfList []any) *awstypes.RadarChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RadarChartFieldWells{}

	if v, ok := tfMap["radar_chart_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.RadarChartAggregatedFieldWells = expandRadarChartAggregatedFieldWells(v)
	}

	return apiObject
}

func expandRadarChartAggregatedFieldWells(tfList []any) *awstypes.RadarChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RadarChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["colors"].([]any); ok && len(v) > 0 {
		apiObject.Color = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandRadarChartSortConfiguration(tfList []any) *awstypes.RadarChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RadarChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]any); ok && len(v) > 0 {
		apiObject.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]any); ok && len(v) > 0 {
		apiObject.CategorySort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["color_items_limit"].([]any); ok && len(v) > 0 {
		apiObject.ColorItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["color_sort"].([]any); ok && len(v) > 0 {
		apiObject.ColorSort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func expandRadarChartSeriesSettings(tfList []any) *awstypes.RadarChartSeriesSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RadarChartSeriesSettings{}

	if v, ok := tfMap["area_style_settings"].([]any); ok && len(v) > 0 {
		apiObject.AreaStyleSettings = expandRadarChartAreaStyleSettings(v)
	}

	return apiObject
}

func expandRadarChartAreaStyleSettings(tfList []any) *awstypes.RadarChartAreaStyleSettings {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RadarChartAreaStyleSettings{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func flattenRadarChartVisual(apiObject *awstypes.RadarChartVisual) []any {
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

	return []any{tfMap}
}

func flattenRadarChartConfiguration(apiObject *awstypes.RadarChartConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"alternate_band_colors_visibility": apiObject.AlternateBandColorsVisibility,
	}

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
	tfMap["shape"] = apiObject.Shape
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenRadarChartSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.StartAngle != nil {
		tfMap["start_angle"] = aws.ToFloat64(apiObject.StartAngle)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}

	return []any{tfMap}
}

func flattenRadarChartSeriesSettings(apiObject *awstypes.RadarChartSeriesSettings) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.AreaStyleSettings != nil {
		tfMap["area_style_settings"] = flattenRadarChartAreaStyleSettings(apiObject.AreaStyleSettings)
	}

	return []any{tfMap}
}

func flattenRadarChartAreaStyleSettings(apiObject *awstypes.RadarChartAreaStyleSettings) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"visibility": apiObject.Visibility,
	}

	return []any{tfMap}
}

func flattenRadarChartFieldWells(apiObject *awstypes.RadarChartFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.RadarChartAggregatedFieldWells != nil {
		tfMap["radar_chart_aggregated_field_wells"] = flattenRadarChartAggregatedFieldWells(apiObject.RadarChartAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenRadarChartAggregatedFieldWells(apiObject *awstypes.RadarChartAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.Color != nil {
		tfMap["color"] = flattenDimensionFields(apiObject.Color)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenRadarChartSortConfiguration(apiObject *awstypes.RadarChartSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}
