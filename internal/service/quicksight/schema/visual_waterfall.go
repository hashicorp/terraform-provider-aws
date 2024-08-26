// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func waterfallVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WaterfallVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WaterfallChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"category_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"category_axis_label_options":   chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"data_labels":                   dataLabelOptionsSchema(),      // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WaterfallChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"waterfall_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WaterfallChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"breakdowns":     dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"categories":     dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":                         legendOptionsSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"primary_y_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"primary_y_axis_label_options":   chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WaterfallChartSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"breakdown_items_limit": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":         fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"visual_palette": visualPaletteSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
							"waterfall_chart_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WaterfallChartOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"total_bar_label": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
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

func expandWaterfallVisual(tfList []interface{}) *quicksight.WaterfallVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.WaterfallVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandWaterfallChartConfiguration(v)
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

func expandWaterfallChartConfiguration(tfList []interface{}) *quicksight.WaterfallChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.WaterfallChartConfiguration{}

	if v, ok := tfMap["category_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["category_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		config.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandWaterfallChartFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.PrimaryYAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.PrimaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandWaterfallChartSortConfiguration(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}
	if v, ok := tfMap["waterfall_chart_options"].([]interface{}); ok && len(v) > 0 {
		config.WaterfallChartOptions = expandWaterfallChartOptions(v)
	}

	return config
}

func expandWaterfallChartFieldWells(tfList []interface{}) *quicksight.WaterfallChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.WaterfallChartFieldWells{}

	if v, ok := tfMap["waterfall_chart_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.WaterfallChartAggregatedFieldWells = expandWaterfallChartAggregatedFieldWells(v)
	}

	return config
}

func expandWaterfallChartAggregatedFieldWells(tfList []interface{}) *quicksight.WaterfallChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.WaterfallChartAggregatedFieldWells{}

	if v, ok := tfMap["breakdowns"].([]interface{}); ok && len(v) > 0 {
		config.Breakdowns = expandDimensionFields(v)
	}
	if v, ok := tfMap["categories"].([]interface{}); ok && len(v) > 0 {
		config.Categories = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandWaterfallChartSortConfiguration(tfList []interface{}) *quicksight.WaterfallChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.WaterfallChartSortConfiguration{}

	if v, ok := tfMap["breakdown_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.BreakdownItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandWaterfallChartOptions(tfList []interface{}) *quicksight.WaterfallChartOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.WaterfallChartOptions{}

	if v, ok := tfMap["total_bar_label"].(string); ok && v != "" {
		options.TotalBarLabel = aws.String(v)
	}

	return options
}

func flattenWaterfallVisual(apiObject *quicksight.WaterfallVisual) []interface{} {
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
		tfMap["chart_configuration"] = flattenWaterfallChartConfiguration(apiObject.ChartConfiguration)
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

func flattenWaterfallChartConfiguration(apiObject *quicksight.WaterfallChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryAxisDisplayOptions != nil {
		tfMap["category_axis_display_options"] = flattenAxisDisplayOptions(apiObject.CategoryAxisDisplayOptions)
	}
	if apiObject.CategoryAxisLabelOptions != nil {
		tfMap["category_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.CategoryAxisLabelOptions)
	}
	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenWaterfallChartFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.PrimaryYAxisDisplayOptions != nil {
		tfMap["primary_y_axis_display_options"] = flattenAxisDisplayOptions(apiObject.PrimaryYAxisDisplayOptions)
	}
	if apiObject.PrimaryYAxisLabelOptions != nil {
		tfMap["primary_y_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.PrimaryYAxisLabelOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenWaterfallChartSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}
	if apiObject.WaterfallChartOptions != nil {
		tfMap["waterfall_chart_options"] = flattenWaterfallChartOptions(apiObject.WaterfallChartOptions)
	}

	return []interface{}{tfMap}
}

func flattenWaterfallChartFieldWells(apiObject *quicksight.WaterfallChartFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.WaterfallChartAggregatedFieldWells != nil {
		tfMap["waterfall_chart_aggregated_field_wells"] = flattenWaterfallChartAggregatedFieldWells(apiObject.WaterfallChartAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenWaterfallChartAggregatedFieldWells(apiObject *quicksight.WaterfallChartAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Breakdowns != nil {
		tfMap["breakdowns"] = flattenDimensionFields(apiObject.Breakdowns)
	}
	if apiObject.Categories != nil {
		tfMap["categories"] = flattenDimensionFields(apiObject.Categories)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenWaterfallChartSortConfiguration(apiObject *quicksight.WaterfallChartSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BreakdownItemsLimit != nil {
		tfMap["breakdown_items_limit"] = flattenItemsLimitConfiguration(apiObject.BreakdownItemsLimit)
	}
	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}

	return []interface{}{tfMap}
}

func flattenWaterfallChartOptions(apiObject *quicksight.WaterfallChartOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TotalBarLabel != nil {
		tfMap["total_bar_label"] = aws.StringValue(apiObject.TotalBarLabel)
	}

	return []interface{}{tfMap}
}
