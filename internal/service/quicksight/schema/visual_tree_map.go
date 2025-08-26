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

func treeMapVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TreeMapVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
										"tree_map_sort": fieldSortOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
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

func expandTreeMapVisual(tfList []any) *awstypes.TreeMapVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TreeMapVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandTreeMapConfiguration(v)
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

func expandTreeMapConfiguration(tfList []any) *awstypes.TreeMapConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TreeMapConfiguration{}

	if v, ok := tfMap["color_label_options"].([]any); ok && len(v) > 0 {
		apiObject.ColorLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["color_scale"].([]any); ok && len(v) > 0 {
		apiObject.ColorScale = expandColorScale(v)
	}
	if v, ok := tfMap["data_labels"].([]any); ok && len(v) > 0 {
		apiObject.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandTreeMapFieldWells(v)
	}
	if v, ok := tfMap["group_label_options"].([]any); ok && len(v) > 0 {
		apiObject.GroupLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["size_label_options"].([]any); ok && len(v) > 0 {
		apiObject.SizeLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandTreeMapSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}

	return apiObject
}

func expandTreeMapFieldWells(tfList []any) *awstypes.TreeMapFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TreeMapFieldWells{}

	if v, ok := tfMap["tree_map_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.TreeMapAggregatedFieldWells = expandTreeMapAggregatedFieldWells(v)
	}

	return apiObject
}

func expandTreeMapAggregatedFieldWells(tfList []any) *awstypes.TreeMapAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TreeMapAggregatedFieldWells{}

	if v, ok := tfMap["colors"].([]any); ok && len(v) > 0 {
		apiObject.Colors = expandMeasureFields(v)
	}
	if v, ok := tfMap["groups"].([]any); ok && len(v) > 0 {
		apiObject.Groups = expandDimensionFields(v)
	}
	if v, ok := tfMap["sizes"].([]any); ok && len(v) > 0 {
		apiObject.Sizes = expandMeasureFields(v)
	}

	return apiObject
}

func expandTreeMapSortConfiguration(tfList []any) *awstypes.TreeMapSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TreeMapSortConfiguration{}

	if v, ok := tfMap["tree_map_group_items_limit_configuration"].([]any); ok && len(v) > 0 {
		apiObject.TreeMapGroupItemsLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["tree_map_sort"].([]any); ok && len(v) > 0 {
		apiObject.TreeMapSort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func flattenTreeMapVisual(apiObject *awstypes.TreeMapVisual) []any {
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

	return []any{tfMap}
}

func flattenTreeMapConfiguration(apiObject *awstypes.TreeMapConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}

func flattenTreeMapFieldWells(apiObject *awstypes.TreeMapFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.TreeMapAggregatedFieldWells != nil {
		tfMap["tree_map_aggregated_field_wells"] = flattenTreeMapAggregatedFieldWells(apiObject.TreeMapAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenTreeMapAggregatedFieldWells(apiObject *awstypes.TreeMapAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Colors != nil {
		tfMap["colors"] = flattenMeasureFields(apiObject.Colors)
	}
	if apiObject.Groups != nil {
		tfMap["groups"] = flattenDimensionFields(apiObject.Groups)
	}
	if apiObject.Sizes != nil {
		tfMap["sizes"] = flattenMeasureFields(apiObject.Sizes)
	}

	return []any{tfMap}
}

func flattenTreeMapSortConfiguration(apiObject *awstypes.TreeMapSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.TreeMapGroupItemsLimitConfiguration != nil {
		tfMap["tree_map_group_items_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.TreeMapGroupItemsLimitConfiguration)
	}
	if apiObject.TreeMapSort != nil {
		tfMap["tree_map_sort"] = flattenFieldSortOptions(apiObject.TreeMapSort)
	}

	return []any{tfMap}
}
