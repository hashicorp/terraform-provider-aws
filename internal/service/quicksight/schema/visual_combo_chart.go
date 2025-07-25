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

func comboChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"bar_data_labels":        dataLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"bars_arrangement":       stringEnumSchema[awstypes.BarsArrangement](attrOptional),
							"category_axis":          axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"category_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"color_label_options":    chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"combo_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"bar_values":  measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
													"category":    dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"colors":      dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"line_values": measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":                           legendOptionsSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"line_data_labels":                 dataLabelOptionsSchema(),      // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"primary_y_axis_display_options":   axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"primary_y_axis_label_options":     chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"reference_lines":                  referenceLineSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
							"secondary_y_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"secondary_y_axis_label_options":   chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartSortConfiguration.html
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
							"tooltip":        tooltipOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"visual_palette": visualPaletteSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
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

func expandComboChartVisual(tfList []any) *awstypes.ComboChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ComboChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandComboChartConfiguration(v)
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

func expandComboChartConfiguration(tfList []any) *awstypes.ComboChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ComboChartConfiguration{}

	if v, ok := tfMap["bars_arrangement"].(string); ok && v != "" {
		apiObject.BarsArrangement = awstypes.BarsArrangement(v)
	}
	if v, ok := tfMap["bar_data_labels"].([]any); ok && len(v) > 0 {
		apiObject.BarDataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["category_axis"].([]any); ok && len(v) > 0 {
		apiObject.CategoryAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["category_label_options"].([]any); ok && len(v) > 0 {
		apiObject.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["color_label_options"].([]any); ok && len(v) > 0 {
		apiObject.ColorLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandComboChartFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["line_data_labels"].([]any); ok && len(v) > 0 {
		apiObject.LineDataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_display_options"].([]any); ok && len(v) > 0 {
		apiObject.PrimaryYAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_label_options"].([]any); ok && len(v) > 0 {
		apiObject.PrimaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["reference_lines"].([]any); ok && len(v) > 0 {
		apiObject.ReferenceLines = expandReferenceLines(v)
	}
	if v, ok := tfMap["secondary_y_axis_display_options"].([]any); ok && len(v) > 0 {
		apiObject.SecondaryYAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["secondary_y_axis_label_options"].([]any); ok && len(v) > 0 {
		apiObject.SecondaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandComboChartSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]any); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}

	return apiObject
}

func expandComboChartFieldWells(tfList []any) *awstypes.ComboChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ComboChartFieldWells{}

	if v, ok := tfMap["combo_chart_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.ComboChartAggregatedFieldWells = expandComboChartAggregatedFieldWells(v)
	}

	return apiObject
}

func expandComboChartAggregatedFieldWells(tfList []any) *awstypes.ComboChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ComboChartAggregatedFieldWells{}

	if v, ok := tfMap["bar_values"].([]any); ok && len(v) > 0 {
		apiObject.BarValues = expandMeasureFields(v)
	}
	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["colors"].([]any); ok && len(v) > 0 {
		apiObject.Colors = expandDimensionFields(v)
	}
	if v, ok := tfMap["line_values"].([]any); ok && len(v) > 0 {
		apiObject.LineValues = expandMeasureFields(v)
	}

	return apiObject
}

func expandComboChartSortConfiguration(tfList []any) *awstypes.ComboChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ComboChartSortConfiguration{}

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

func flattenComboChartVisual(apiObject *awstypes.ComboChartVisual) []any {
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
		tfMap["chart_configuration"] = flattenComboChartConfiguration(apiObject.ChartConfiguration)
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

func flattenComboChartConfiguration(apiObject *awstypes.ComboChartConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BarDataLabels != nil {
		tfMap["bar_data_labels"] = flattenDataLabelOptions(apiObject.BarDataLabels)
	}
	tfMap["bars_arrangement"] = apiObject.BarsArrangement
	if apiObject.CategoryAxis != nil {
		tfMap["category_axis"] = flattenAxisDisplayOptions(apiObject.CategoryAxis)
	}
	if apiObject.CategoryLabelOptions != nil {
		tfMap["category_label_options"] = flattenChartAxisLabelOptions(apiObject.CategoryLabelOptions)
	}
	if apiObject.ColorLabelOptions != nil {
		tfMap["color_label_options"] = flattenChartAxisLabelOptions(apiObject.ColorLabelOptions)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenComboChartFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.LineDataLabels != nil {
		tfMap["line_data_labels"] = flattenDataLabelOptions(apiObject.LineDataLabels)
	}
	if apiObject.PrimaryYAxisDisplayOptions != nil {
		tfMap["primary_y_axis_display_options"] = flattenAxisDisplayOptions(apiObject.PrimaryYAxisDisplayOptions)
	}
	if apiObject.PrimaryYAxisLabelOptions != nil {
		tfMap["primary_y_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.PrimaryYAxisLabelOptions)
	}
	if apiObject.ReferenceLines != nil {
		tfMap["reference_lines"] = flattenReferenceLine(apiObject.ReferenceLines)
	}
	if apiObject.SecondaryYAxisDisplayOptions != nil {
		tfMap["secondary_y_axis_display_options"] = flattenAxisDisplayOptions(apiObject.SecondaryYAxisDisplayOptions)
	}
	if apiObject.SecondaryYAxisLabelOptions != nil {
		tfMap["secondary_y_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.SecondaryYAxisLabelOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenComboChartSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}

	return []any{tfMap}
}

func flattenComboChartFieldWells(apiObject *awstypes.ComboChartFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComboChartAggregatedFieldWells != nil {
		tfMap["combo_chart_aggregated_field_wells"] = flattenComboChartAggregatedFieldWells(apiObject.ComboChartAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenComboChartAggregatedFieldWells(apiObject *awstypes.ComboChartAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BarValues != nil {
		tfMap["bar_values"] = flattenMeasureFields(apiObject.BarValues)
	}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.Colors != nil {
		tfMap["colors"] = flattenDimensionFields(apiObject.Colors)
	}
	if apiObject.LineValues != nil {
		tfMap["line_values"] = flattenMeasureFields(apiObject.LineValues)
	}

	return []any{tfMap}
}

func flattenComboChartSortConfiguration(apiObject *awstypes.ComboChartSortConfiguration) []any {
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
