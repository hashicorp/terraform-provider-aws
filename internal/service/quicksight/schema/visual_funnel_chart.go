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

func funnelChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"category_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"data_label_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartDataLabelOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_label_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
										"label_color":               hexColorSchema(attrOptional),
										"label_font_configuration":  fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
										"measure_data_label_style":  stringEnumSchema[awstypes.FunnelChartMeasureDataLabelStyle](attrOptional),
										"measure_label_visibility":  stringEnumSchema[awstypes.Visibility](attrOptional),
										"position":                  stringEnumSchema[awstypes.DataLabelPosition](attrOptional),
										"visibility":                stringEnumSchema[awstypes.Visibility](attrOptional),
									},
								},
							},
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"funnel_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"category":       dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit": itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":        fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
									},
								},
							},
							"tooltip":             tooltipOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"value_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"visual_palette":      visualPaletteSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
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

func expandFunnelChartVisual(tfList []any) *awstypes.FunnelChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FunnelChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandFunnelChartConfiguration(v)
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

func expandFunnelChartConfiguration(tfList []any) *awstypes.FunnelChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FunnelChartConfiguration{}

	if v, ok := tfMap["category_label_options"].([]any); ok && len(v) > 0 {
		apiObject.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["data_label_options"].([]any); ok && len(v) > 0 {
		apiObject.DataLabelOptions = expandFunnelChartDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandFunnelChartFieldWells(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandFunnelChartSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["value_label_options"].([]any); ok && len(v) > 0 {
		apiObject.ValueLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]any); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}

	return apiObject
}

func expandFunnelChartFieldWells(tfList []any) *awstypes.FunnelChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FunnelChartFieldWells{}

	if v, ok := tfMap["funnel_chart_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FunnelChartAggregatedFieldWells = expandFunnelChartAggregatedFieldWells(v)
	}

	return apiObject
}

func expandFunnelChartAggregatedFieldWells(tfList []any) *awstypes.FunnelChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FunnelChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandFunnelChartSortConfiguration(tfList []any) *awstypes.FunnelChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FunnelChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]any); ok && len(v) > 0 {
		apiObject.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]any); ok && len(v) > 0 {
		apiObject.CategorySort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func expandFunnelChartDataLabelOptions(tfList []any) *awstypes.FunnelChartDataLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FunnelChartDataLabelOptions{}

	if v, ok := tfMap["category_label_visibility"].(string); ok && v != "" {
		apiObject.CategoryLabelVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["label_color"].(string); ok && v != "" {
		apiObject.LabelColor = aws.String(v)
	}
	if v, ok := tfMap["measure_data_label_style"].(string); ok && v != "" {
		apiObject.MeasureDataLabelStyle = awstypes.FunnelChartMeasureDataLabelStyle(v)
	}
	if v, ok := tfMap["measure_label_visibility"].(string); ok && v != "" {
		apiObject.MeasureLabelVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["position"].(string); ok && v != "" {
		apiObject.Position = awstypes.DataLabelPosition(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["label_font_configuration"].([]any); ok && len(v) > 0 {
		apiObject.LabelFontConfiguration = expandFontConfiguration(v)
	}

	return apiObject
}

func flattenFunnelChartVisual(apiObject *awstypes.FunnelChartVisual) []any {
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
		tfMap["chart_configuration"] = flattenFunnelChartConfiguration(apiObject.ChartConfiguration)
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

func flattenFunnelChartConfiguration(apiObject *awstypes.FunnelChartConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CategoryLabelOptions != nil {
		tfMap["category_label_options"] = flattenChartAxisLabelOptions(apiObject.CategoryLabelOptions)
	}
	if apiObject.DataLabelOptions != nil {
		tfMap["data_label_options"] = flattenFunnelChartDataLabelOptions(apiObject.DataLabelOptions)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenFunnelChartFieldWells(apiObject.FieldWells)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenFunnelChartSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}
	if apiObject.ValueLabelOptions != nil {
		tfMap["value_label_options"] = flattenChartAxisLabelOptions(apiObject.ValueLabelOptions)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}

	return []any{tfMap}
}

func flattenFunnelChartDataLabelOptions(apiObject *awstypes.FunnelChartDataLabelOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["category_label_visibility"] = apiObject.CategoryLabelVisibility
	if apiObject.LabelColor != nil {
		tfMap["label_color"] = aws.ToString(apiObject.LabelColor)
	}
	if apiObject.LabelFontConfiguration != nil {
		tfMap["label_font_configuration"] = flattenFontConfiguration(apiObject.LabelFontConfiguration)
	}
	tfMap["measure_data_label_style"] = apiObject.MeasureDataLabelStyle
	tfMap["measure_label_visibility"] = apiObject.MeasureLabelVisibility
	tfMap["position"] = apiObject.Position
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenFunnelChartFieldWells(apiObject *awstypes.FunnelChartFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FunnelChartAggregatedFieldWells != nil {
		tfMap["funnel_chart_aggregated_field_wells"] = flattenFunnelChartAggregatedFieldWells(apiObject.FunnelChartAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenFunnelChartAggregatedFieldWells(apiObject *awstypes.FunnelChartAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenFunnelChartSortConfiguration(apiObject *awstypes.FunnelChartSortConfiguration) []any {
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

	return []any{tfMap}
}
