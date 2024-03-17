// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func filledMapVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"filled_map_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"geospatial": dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"values":     measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":            legendOptionsSchema(),             // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"map_style_options": geospatialMapStyleOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialMapStyleOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_sort": fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
									},
								},
							},
							"tooltip":        tooltipOptionsSchema(),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"window_options": geospatialWindowOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialWindowOptions.html
						},
					},
				},
				"column_hierarchies": columnHierarchiesSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnHierarchy.html
				"conditional_formatting": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapConditionalFormatting.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"conditional_formatting_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapConditionalFormattingOption.html
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 200,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"shape": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapShapeConditionalFormatting.html
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
													"format": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ShapeConditionalFormat.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"background_color": conditionalFormattingColorSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingColor.html
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
					},
				},
				"subtitle": visualSubtitleLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
				"title":    visualTitleLabelOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
			},
		},
	}
}

func expandFilledMapVisual(tfList []interface{}) *types.FilledMapVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.FilledMapVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandFilledMapConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]interface{}); ok && len(v) > 0 {
		visual.ConditionalFormatting = expandFilledMapConditionalFormatting(v)
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

func expandFilledMapConfiguration(tfList []interface{}) *types.FilledMapConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilledMapConfiguration{}

	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandFilledMapFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["map_style_options"].([]interface{}); ok && len(v) > 0 {
		config.MapStyleOptions = expandGeospatialMapStyleOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandFilledMapSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["value_axis"].([]interface{}); ok && len(v) > 0 {
		config.WindowOptions = expandGeospatialWindowOptions(v)
	}

	return config
}

func expandFilledMapFieldWells(tfList []interface{}) *types.FilledMapFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilledMapFieldWells{}

	if v, ok := tfMap["filled_map_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FilledMapAggregatedFieldWells = expandFilledMapAggregatedFieldWells(v)
	}

	return config
}

func expandFilledMapAggregatedFieldWells(tfList []interface{}) *types.FilledMapAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilledMapAggregatedFieldWells{}

	if v, ok := tfMap["geospatial"].([]interface{}); ok && len(v) > 0 {
		config.Geospatial = expandDimensionFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandFilledMapSortConfiguration(tfList []interface{}) *types.FilledMapSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilledMapSortConfiguration{}

	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandFilledMapConditionalFormatting(tfList []interface{}) *types.FilledMapConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilledMapConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]interface{}); ok && len(v) > 0 {
		config.ConditionalFormattingOptions = expandFilledMapConditionalFormattingOptions(v)
	}

	return config
}

func expandFilledMapConditionalFormattingOptions(tfList []interface{}) []types.FilledMapConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.FilledMapConditionalFormattingOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandFilledMapConditionalFormattingOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandFilledMapConditionalFormattingOption(tfMap map[string]interface{}) *types.FilledMapConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	options := &types.FilledMapConditionalFormattingOption{}

	if v, ok := tfMap["shape"].([]interface{}); ok && len(v) > 0 {
		options.Shape = expandFilledMapShapeConditionalFormatting(v)
	}

	return options
}

func expandFilledMapShapeConditionalFormatting(tfList []interface{}) *types.FilledMapShapeConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.FilledMapShapeConditionalFormatting{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["format"].([]interface{}); ok && len(v) > 0 {
		options.Format = expandShapeConditionalFormat(v)
	}

	return options
}

func expandShapeConditionalFormat(tfList []interface{}) *types.ShapeConditionalFormat {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.ShapeConditionalFormat{}

	if v, ok := tfMap["background_color"].([]interface{}); ok && len(v) > 0 {
		options.BackgroundColor = expandConditionalFormattingColor(v)
	}

	return options
}

func flattenFilledMapVisual(apiObject *types.FilledMapVisual) []interface{} {
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
		tfMap["chart_configuration"] = flattenFilledMapConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.ColumnHierarchies != nil {
		tfMap["column_hierarchies"] = flattenColumnHierarchy(apiObject.ColumnHierarchies)
	}
	if apiObject.ConditionalFormatting != nil {
		tfMap["conditional_formatting"] = flattenFilledMapConditionalFormatting(apiObject.ConditionalFormatting)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []interface{}{tfMap}
}

func flattenFilledMapConfiguration(apiObject *types.FilledMapConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenFilledMapFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.MapStyleOptions != nil {
		tfMap["map_style_options"] = flattenGeospatialMapStyleOptions(apiObject.MapStyleOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenFilledMapSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}
	if apiObject.WindowOptions != nil {
		tfMap["window_options"] = flattenGeospatialWindowOptions(apiObject.WindowOptions)
	}

	return []interface{}{tfMap}
}

func flattenFilledMapFieldWells(apiObject *types.FilledMapFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FilledMapAggregatedFieldWells != nil {
		tfMap["filled_map_aggregated_field_wells"] = flattenFilledMapAggregatedFieldWells(apiObject.FilledMapAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenFilledMapAggregatedFieldWells(apiObject *types.FilledMapAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Geospatial != nil {
		tfMap["geospatial"] = flattenDimensionFields(apiObject.Geospatial)
	}
	if apiObject.Values != nil {
		tfMap["values"] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenFilledMapSortConfiguration(apiObject *types.FilledMapSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}

	return []interface{}{tfMap}
}

func flattenFilledMapConditionalFormatting(apiObject *types.FilledMapConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenFilledMapConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []interface{}{tfMap}
}

func flattenFilledMapConditionalFormattingOption(apiObject []types.FilledMapConditionalFormattingOption) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.Shape != nil {
			tfMap["shape"] = flattenFilledMapShapeConditionalFormatting(config.Shape)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilledMapShapeConditionalFormatting(apiObject *types.FilledMapShapeConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.Format != nil {
		tfMap["format"] = flattenShapeConditionalFormat(apiObject.Format)
	}

	return []interface{}{tfMap}
}

func flattenShapeConditionalFormat(apiObject *types.ShapeConditionalFormat) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = flattenConditionalFormattingColor(apiObject.BackgroundColor)
	}

	return []interface{}{tfMap}
}
