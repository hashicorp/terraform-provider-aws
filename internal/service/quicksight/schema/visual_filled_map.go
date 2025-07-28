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

func filledMapVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilledMapVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
													"geospatial":     dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
										"category_sort": fieldSortOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
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
													"field_id": stringLenBetweenSchema(attrRequired, 1, 512),
													names.AttrFormat: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ShapeConditionalFormat.html
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

func expandFilledMapVisual(tfList []any) *awstypes.FilledMapVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilledMapVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandFilledMapConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]any); ok && len(v) > 0 {
		apiObject.ConditionalFormatting = expandFilledMapConditionalFormatting(v)
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

func expandFilledMapConfiguration(tfList []any) *awstypes.FilledMapConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilledMapConfiguration{}

	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandFilledMapFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["map_style_options"].([]any); ok && len(v) > 0 {
		apiObject.MapStyleOptions = expandGeospatialMapStyleOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandFilledMapSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["value_axis"].([]any); ok && len(v) > 0 {
		apiObject.WindowOptions = expandGeospatialWindowOptions(v)
	}

	return apiObject
}

func expandFilledMapFieldWells(tfList []any) *awstypes.FilledMapFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilledMapFieldWells{}

	if v, ok := tfMap["filled_map_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FilledMapAggregatedFieldWells = expandFilledMapAggregatedFieldWells(v)
	}

	return apiObject
}

func expandFilledMapAggregatedFieldWells(tfList []any) *awstypes.FilledMapAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilledMapAggregatedFieldWells{}

	if v, ok := tfMap["geospatial"].([]any); ok && len(v) > 0 {
		apiObject.Geospatial = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandFilledMapSortConfiguration(tfList []any) *awstypes.FilledMapSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilledMapSortConfiguration{}

	if v, ok := tfMap["category_sort"].([]any); ok && len(v) > 0 {
		apiObject.CategorySort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func expandFilledMapConditionalFormatting(tfList []any) *awstypes.FilledMapConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilledMapConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]any); ok && len(v) > 0 {
		apiObject.ConditionalFormattingOptions = expandFilledMapConditionalFormattingOptions(v)
	}

	return apiObject
}

func expandFilledMapConditionalFormattingOptions(tfList []any) []awstypes.FilledMapConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FilledMapConditionalFormattingOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandFilledMapConditionalFormattingOption(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFilledMapConditionalFormattingOption(tfMap map[string]any) *awstypes.FilledMapConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FilledMapConditionalFormattingOption{}

	if v, ok := tfMap["shape"].([]any); ok && len(v) > 0 {
		apiObject.Shape = expandFilledMapShapeConditionalFormatting(v)
	}

	return apiObject
}

func expandFilledMapShapeConditionalFormatting(tfList []any) *awstypes.FilledMapShapeConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilledMapShapeConditionalFormatting{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrFormat].([]any); ok && len(v) > 0 {
		apiObject.Format = expandShapeConditionalFormat(v)
	}

	return apiObject
}

func expandShapeConditionalFormat(tfList []any) *awstypes.ShapeConditionalFormat {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ShapeConditionalFormat{}

	if v, ok := tfMap["background_color"].([]any); ok && len(v) > 0 {
		apiObject.BackgroundColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func flattenFilledMapVisual(apiObject *awstypes.FilledMapVisual) []any {
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

	return []any{tfMap}
}

func flattenFilledMapConfiguration(apiObject *awstypes.FilledMapConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}

func flattenFilledMapFieldWells(apiObject *awstypes.FilledMapFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FilledMapAggregatedFieldWells != nil {
		tfMap["filled_map_aggregated_field_wells"] = flattenFilledMapAggregatedFieldWells(apiObject.FilledMapAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenFilledMapAggregatedFieldWells(apiObject *awstypes.FilledMapAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Geospatial != nil {
		tfMap["geospatial"] = flattenDimensionFields(apiObject.Geospatial)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenFilledMapSortConfiguration(apiObject *awstypes.FilledMapSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}

	return []any{tfMap}
}

func flattenFilledMapConditionalFormatting(apiObject *awstypes.FilledMapConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenFilledMapConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []any{tfMap}
}

func flattenFilledMapConditionalFormattingOption(apiObjects []awstypes.FilledMapConditionalFormattingOption) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Shape != nil {
			tfMap["shape"] = flattenFilledMapShapeConditionalFormatting(apiObject.Shape)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilledMapShapeConditionalFormatting(apiObject *awstypes.FilledMapShapeConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.Format != nil {
		tfMap[names.AttrFormat] = flattenShapeConditionalFormat(apiObject.Format)
	}

	return []any{tfMap}
}

func flattenShapeConditionalFormat(apiObject *awstypes.ShapeConditionalFormat) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = flattenConditionalFormattingColor(apiObject.BackgroundColor)
	}

	return []any{tfMap}
}
