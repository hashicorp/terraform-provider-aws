// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func treeMapVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TreeMapVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TreeMapConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"color_scale":         colorScaleSchema(),            // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColorScale.html
							"data_labels":         dataLabelOptionsSchema(),      // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TreeMapFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"tree_map_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TreeMapAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"colors": measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
													"groups": dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"sizes":  measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"group_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"legend":              legendOptionsSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"size_label_options":  chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TreeMapSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"tree_map_group_items_limit_configuration": itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"tree_map_sort": fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
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

func expandTreeMapVisual(tfList []interface{}) *types.TreeMapVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.TreeMapVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandTreeMapConfiguration(v)
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

func expandTreeMapConfiguration(tfList []interface{}) *types.TreeMapConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TreeMapConfiguration{}

	if v, ok := tfMap["color_label_options"].([]interface{}); ok && len(v) > 0 {
		config.ColorLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["color_scale"].([]interface{}); ok && len(v) > 0 {
		config.ColorScale = expandColorScale(v)
	}
	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		config.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandTreeMapFieldWells(v)
	}
	if v, ok := tfMap["group_label_options"].([]interface{}); ok && len(v) > 0 {
		config.GroupLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["size_label_options"].([]interface{}); ok && len(v) > 0 {
		config.SizeLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandTreeMapSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}

	return config
}

func expandTreeMapFieldWells(tfList []interface{}) *types.TreeMapFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TreeMapFieldWells{}

	if v, ok := tfMap["tree_map_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.TreeMapAggregatedFieldWells = expandTreeMapAggregatedFieldWells(v)
	}

	return config
}

func expandTreeMapAggregatedFieldWells(tfList []interface{}) *types.TreeMapAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TreeMapAggregatedFieldWells{}

	if v, ok := tfMap["colors"].([]interface{}); ok && len(v) > 0 {
		config.Colors = expandMeasureFields(v)
	}
	if v, ok := tfMap["groups"].([]interface{}); ok && len(v) > 0 {
		config.Groups = expandDimensionFields(v)
	}
	if v, ok := tfMap["sizes"].([]interface{}); ok && len(v) > 0 {
		config.Sizes = expandMeasureFields(v)
	}

	return config
}

func expandTreeMapSortConfiguration(tfList []interface{}) *types.TreeMapSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TreeMapSortConfiguration{}

	if v, ok := tfMap["tree_map_group_items_limit_configuration"].([]interface{}); ok && len(v) > 0 {
		config.TreeMapGroupItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["tree_map_sort"].([]interface{}); ok && len(v) > 0 {
		config.TreeMapSort = expandFieldSortOptionsList(v)
	}

	return config
}

func flattenTreeMapVisual(apiObject *types.TreeMapVisual) []interface{} {
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
		tfMap["chart_configuration"] = flattenTreeMapConfiguration(apiObject.ChartConfiguration)
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

func flattenTreeMapConfiguration(apiObject *types.TreeMapConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColorLabelOptions != nil {
		tfMap["color_label_options"] = flattenChartAxisLabelOptions(apiObject.ColorLabelOptions)
	}
	if apiObject.ColorScale != nil {
		tfMap["color_scale"] = flattenColorScale(apiObject.ColorScale)
	}
	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenTreeMapFieldWells(apiObject.FieldWells)
	}
	if apiObject.GroupLabelOptions != nil {
		tfMap["group_label_options"] = flattenChartAxisLabelOptions(apiObject.GroupLabelOptions)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.SizeLabelOptions != nil {
		tfMap["size_label_options"] = flattenChartAxisLabelOptions(apiObject.SizeLabelOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenTreeMapSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}

	return []interface{}{tfMap}
}

func flattenTreeMapFieldWells(apiObject *types.TreeMapFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TreeMapAggregatedFieldWells != nil {
		tfMap["tree_map_aggregated_field_wells"] = flattenTreeMapAggregatedFieldWells(apiObject.TreeMapAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenTreeMapAggregatedFieldWells(apiObject *types.TreeMapAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Colors != nil {
		tfMap["colors"] = flattenMeasureFields(apiObject.Colors)
	}
	if apiObject.Groups != nil {
		tfMap["groups"] = flattenDimensionFields(apiObject.Groups)
	}
	if apiObject.Sizes != nil {
		tfMap["sizes"] = flattenMeasureFields(apiObject.Sizes)
	}

	return []interface{}{tfMap}
}

func flattenTreeMapSortConfiguration(apiObject *types.TreeMapSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TreeMapGroupItemsLimitConfiguration != nil {
		tfMap["tree_map_group_items_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.TreeMapGroupItemsLimitConfiguration)
	}
	if apiObject.TreeMapSort != nil {
		tfMap["tree_map_sort"] = flattenFieldSortOptions(apiObject.TreeMapSort)
	}

	return []interface{}{tfMap}
}
