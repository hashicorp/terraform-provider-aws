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

func funnelChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
										"category_label_visibility": stringSchema(false, enum.Validate[types.Visibility]()),
										"label_color":               stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
										"label_font_configuration":  fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
										"measure_data_label_style":  stringSchema(false, enum.Validate[types.FunnelChartMeasureDataLabelStyle]()),
										"measure_label_visibility":  stringSchema(false, enum.Validate[types.Visibility]()),
										"position":                  stringSchema(false, enum.Validate[types.DataLabelPosition]()),
										"visibility":                stringSchema(false, enum.Validate[types.Visibility]()),
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
													"category": dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"values":   measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
										"category_items_limit": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":        fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
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

func expandFunnelChartVisual(tfList []interface{}) *types.FunnelChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.FunnelChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandFunnelChartConfiguration(v)
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

func expandFunnelChartConfiguration(tfList []interface{}) *types.FunnelChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FunnelChartConfiguration{}

	if v, ok := tfMap["category_label_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["data_label_options"].([]interface{}); ok && len(v) > 0 {
		config.DataLabelOptions = expandFunnelChartDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandFunnelChartFieldWells(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandFunnelChartSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["value_label_options"].([]interface{}); ok && len(v) > 0 {
		config.ValueLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}

	return config
}

func expandFunnelChartFieldWells(tfList []interface{}) *types.FunnelChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FunnelChartFieldWells{}

	if v, ok := tfMap["funnel_chart_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FunnelChartAggregatedFieldWells = expandFunnelChartAggregatedFieldWells(v)
	}

	return config
}

func expandFunnelChartAggregatedFieldWells(tfList []interface{}) *types.FunnelChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FunnelChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		config.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandFunnelChartSortConfiguration(tfList []interface{}) *types.FunnelChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FunnelChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandFunnelChartDataLabelOptions(tfList []interface{}) *types.FunnelChartDataLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.FunnelChartDataLabelOptions{}

	if v, ok := tfMap["category_label_visibility"].(string); ok && v != "" {
		options.CategoryLabelVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["label_color"].(string); ok && v != "" {
		options.LabelColor = aws.String(v)
	}
	if v, ok := tfMap["measure_data_label_style"].(string); ok && v != "" {
		options.MeasureDataLabelStyle = types.FunnelChartMeasureDataLabelStyle(v)
	}
	if v, ok := tfMap["measure_label_visibility"].(string); ok && v != "" {
		options.MeasureLabelVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["position"].(string); ok && v != "" {
		options.Position = types.DataLabelPosition(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}
	if v, ok := tfMap["label_font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.LabelFontConfiguration = expandFontConfiguration(v)
	}

	return options
}

func flattenFunnelChartVisual(apiObject *types.FunnelChartVisual) []interface{} {
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

	return []interface{}{tfMap}
}

func flattenFunnelChartConfiguration(apiObject *types.FunnelChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenFunnelChartDataLabelOptions(apiObject *types.FunnelChartDataLabelOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["category_label_visibility"] = types.Visibility(apiObject.CategoryLabelVisibility)

	if apiObject.LabelColor != nil {
		tfMap["label_color"] = aws.ToString(apiObject.LabelColor)
	}
	if apiObject.LabelFontConfiguration != nil {
		tfMap["label_font_configuration"] = flattenFontConfiguration(apiObject.LabelFontConfiguration)
	}

	tfMap["measure_data_label_style"] = types.FunnelChartMeasureDataLabelStyle(apiObject.MeasureDataLabelStyle)

	tfMap["measure_label_visibility"] = types.Visibility(apiObject.MeasureLabelVisibility)

	tfMap["position"] = types.DataLabelPosition(apiObject.Position)

	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenFunnelChartFieldWells(apiObject *types.FunnelChartFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FunnelChartAggregatedFieldWells != nil {
		tfMap["funnel_chart_aggregated_field_wells"] = flattenFunnelChartAggregatedFieldWells(apiObject.FunnelChartAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenFunnelChartAggregatedFieldWells(apiObject *types.FunnelChartAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.Values != nil {
		tfMap["values"] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenFunnelChartSortConfiguration(apiObject *types.FunnelChartSortConfiguration) []interface{} {
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

	return []interface{}{tfMap}
}
