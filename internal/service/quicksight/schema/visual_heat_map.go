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

func heatMapVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeatMapVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
													"columns":        dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"rows":           dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
										"heat_map_column_items_limit_configuration": itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"heat_map_column_sort":                      fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"heat_map_row_items_limit_configuration":    itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"heat_map_row_sort":                         fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
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

func expandHeatMapVisual(tfList []any) *awstypes.HeatMapVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HeatMapVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandHeatMapConfiguration(v)
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

func expandHeatMapConfiguration(tfList []any) *awstypes.HeatMapConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HeatMapConfiguration{}

	if v, ok := tfMap["color_scale"].([]any); ok && len(v) > 0 {
		apiObject.ColorScale = expandColorScale(v)
	}
	if v, ok := tfMap["column_label_options"].([]any); ok && len(v) > 0 {
		apiObject.ColumnLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["data_labels"].([]any); ok && len(v) > 0 {
		apiObject.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandHeatMapFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["row_label_options"].([]any); ok && len(v) > 0 {
		apiObject.RowLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandHeatMapSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}

	return apiObject
}

func expandHeatMapFieldWells(tfList []any) *awstypes.HeatMapFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HeatMapFieldWells{}

	if v, ok := tfMap["heat_map_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.HeatMapAggregatedFieldWells = expandHeatMapAggregatedFieldWells(v)
	}

	return apiObject
}

func expandHeatMapAggregatedFieldWells(tfList []any) *awstypes.HeatMapAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HeatMapAggregatedFieldWells{}

	if v, ok := tfMap["columns"].([]any); ok && len(v) > 0 {
		apiObject.Columns = expandDimensionFields(v)
	}
	if v, ok := tfMap["rows"].([]any); ok && len(v) > 0 {
		apiObject.Rows = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandHeatMapSortConfiguration(tfList []any) *awstypes.HeatMapSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HeatMapSortConfiguration{}

	if v, ok := tfMap["heat_map_column_items_limit_configuration"].([]any); ok && len(v) > 0 {
		apiObject.HeatMapColumnItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["heat_map_column_sort"].([]any); ok && len(v) > 0 {
		apiObject.HeatMapColumnSort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["heat_map_row_items_limit_configuration"].([]any); ok && len(v) > 0 {
		apiObject.HeatMapRowItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["heat_map_row_sort"].([]any); ok && len(v) > 0 {
		apiObject.HeatMapRowSort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func flattenHeatMapVisual(apiObject *awstypes.HeatMapVisual) []any {
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

	return []any{tfMap}
}

func flattenHeatMapConfiguration(apiObject *awstypes.HeatMapConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}

func flattenHeatMapFieldWells(apiObject *awstypes.HeatMapFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.HeatMapAggregatedFieldWells != nil {
		tfMap["heat_map_aggregated_field_wells"] = flattenHeatMapAggregatedFieldWells(apiObject.HeatMapAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenHeatMapAggregatedFieldWells(apiObject *awstypes.HeatMapAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Columns != nil {
		tfMap["columns"] = flattenDimensionFields(apiObject.Columns)
	}
	if apiObject.Rows != nil {
		tfMap["rows"] = flattenDimensionFields(apiObject.Rows)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenHeatMapSortConfiguration(apiObject *awstypes.HeatMapSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}
