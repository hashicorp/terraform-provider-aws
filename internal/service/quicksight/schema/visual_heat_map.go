// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func heatMapVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeatMapVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeatMapConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color_scale":          colorScaleSchema(),            // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColorScale.html
							"column_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"data_labels":          dataLabelOptionsSchema(),      // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeatMapFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"heat_map_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeatMapAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"columns": dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"rows":    dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"values":  measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":            legendOptionsSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"row_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeatMapSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"heat_map_column_items_limit_configuration": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"heat_map_column_sort":                      fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"heat_map_row_items_limit_configuration":    itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"heat_map_row_sort":                         fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"tooltip": tooltipOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
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

func expandHeatMapVisual(tfList []interface{}) *types.HeatMapVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.HeatMapVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandHeatMapConfiguration(v)
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

func expandHeatMapConfiguration(tfList []interface{}) *types.HeatMapConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.HeatMapConfiguration{}

	if v, ok := tfMap["color_scale"].([]interface{}); ok && len(v) > 0 {
		config.ColorScale = expandColorScale(v)
	}
	if v, ok := tfMap["column_label_options"].([]interface{}); ok && len(v) > 0 {
		config.ColumnLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		config.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandHeatMapFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["row_label_options"].([]interface{}); ok && len(v) > 0 {
		config.RowLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandHeatMapSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}

	return config
}

func expandHeatMapFieldWells(tfList []interface{}) *types.HeatMapFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.HeatMapFieldWells{}

	if v, ok := tfMap["heat_map_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.HeatMapAggregatedFieldWells = expandHeatMapAggregatedFieldWells(v)
	}

	return config
}

func expandHeatMapAggregatedFieldWells(tfList []interface{}) *types.HeatMapAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.HeatMapAggregatedFieldWells{}

	if v, ok := tfMap["columns"].([]interface{}); ok && len(v) > 0 {
		config.Columns = expandDimensionFields(v)
	}
	if v, ok := tfMap["rows"].([]interface{}); ok && len(v) > 0 {
		config.Rows = expandDimensionFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandHeatMapSortConfiguration(tfList []interface{}) *types.HeatMapSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.HeatMapSortConfiguration{}

	if v, ok := tfMap["heat_map_column_items_limit_configuration"].([]interface{}); ok && len(v) > 0 {
		config.HeatMapColumnItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["heat_map_column_sort"].([]interface{}); ok && len(v) > 0 {
		config.HeatMapColumnSort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["heat_map_row_items_limit_configuration"].([]interface{}); ok && len(v) > 0 {
		config.HeatMapRowItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["heat_map_row_sort"].([]interface{}); ok && len(v) > 0 {
		config.HeatMapRowSort = expandFieldSortOptionsList(v)
	}

	return config
}

func flattenHeatMapVisual(apiObject *types.HeatMapVisual) []interface{} {
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
		tfMap["chart_configuration"] = flattenHeatMapConfiguration(apiObject.ChartConfiguration)
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

func flattenHeatMapConfiguration(apiObject *types.HeatMapConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColorScale != nil {
		tfMap["color_scale"] = flattenColorScale(apiObject.ColorScale)
	}
	if apiObject.ColumnLabelOptions != nil {
		tfMap["column_label_options"] = flattenChartAxisLabelOptions(apiObject.ColumnLabelOptions)
	}
	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenHeatMapFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.RowLabelOptions != nil {
		tfMap["row_label_options"] = flattenChartAxisLabelOptions(apiObject.RowLabelOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenHeatMapSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}

	return []interface{}{tfMap}
}

func flattenHeatMapFieldWells(apiObject *types.HeatMapFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.HeatMapAggregatedFieldWells != nil {
		tfMap["heat_map_aggregated_field_wells"] = flattenHeatMapAggregatedFieldWells(apiObject.HeatMapAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenHeatMapAggregatedFieldWells(apiObject *types.HeatMapAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Columns != nil {
		tfMap["columns"] = flattenDimensionFields(apiObject.Columns)
	}
	if apiObject.Rows != nil {
		tfMap["rows"] = flattenDimensionFields(apiObject.Rows)
	}
	if apiObject.Values != nil {
		tfMap["values"] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenHeatMapSortConfiguration(apiObject *types.HeatMapSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.HeatMapColumnItemsLimitConfiguration != nil {
		tfMap["heat_map_column_items_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.HeatMapColumnItemsLimitConfiguration)
	}
	if apiObject.HeatMapColumnSort != nil {
		tfMap["heat_map_column_sort"] = flattenFieldSortOptions(apiObject.HeatMapColumnSort)
	}
	if apiObject.HeatMapRowItemsLimitConfiguration != nil {
		tfMap["heat_map_row_items_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.HeatMapRowItemsLimitConfiguration)
	}
	if apiObject.HeatMapRowSort != nil {
		tfMap["heat_map_row_sort"] = flattenFieldSortOptions(apiObject.HeatMapRowSort)
	}

	return []interface{}{tfMap}
}
