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

func barCharVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BarChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BarChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"bars_arrangement":               stringEnumSchema[awstypes.BarsArrangement](attrOptionalComputed),
							"category_axis":                  axisDisplayOptionsSchema(),           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"category_label_options":         chartAxisLabelOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"color_label_options":            chartAxisLabelOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"contribution_analysis_defaults": contributionAnalysisDefaultsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ContributionAnalysisDefault.html
							"data_labels":                    dataLabelOptionsSchema(),             // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BarChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"bar_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BarChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"category":        dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"colors":          dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"small_multiples": dimensionFieldSchema(1),                          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues:  measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":                  legendOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"orientation":             stringEnumSchema[awstypes.BarChartOrientation](attrOptionalComputed),
							"reference_lines":         referenceLineSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
							"small_multiples_options": smallMultiplesOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SmallMultiplesOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BarChartSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit":                itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":                       fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"color_items_limit":                   itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"color_sort":                          fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
										"small_multiples_limit_configuration": itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"small_multiples_sort":                fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"tooltip":             tooltipOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"value_axis":          axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
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

func expandBarChartVisual(tfList []any) *awstypes.BarChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BarChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandBarChartConfiguration(v)
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

func expandBarChartConfiguration(tfList []any) *awstypes.BarChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BarChartConfiguration{}

	if v, ok := tfMap["bars_arrangement"].(string); ok && v != "" {
		apiObject.BarsArrangement = awstypes.BarsArrangement(v)
	}
	if v, ok := tfMap["orientation"].(string); ok && v != "" {
		apiObject.Orientation = awstypes.BarChartOrientation(v)
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
	if v, ok := tfMap["contribution_analysis_defaults"].([]any); ok && len(v) > 0 {
		apiObject.ContributionAnalysisDefaults = expandContributionAnalysisDefaults(v)
	}
	if v, ok := tfMap["data_labels"].([]any); ok && len(v) > 0 {
		apiObject.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandBarChartFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["reference_lines"].([]any); ok && len(v) > 0 {
		apiObject.ReferenceLines = expandReferenceLines(v)
	}
	if v, ok := tfMap["small_multiples_options"].([]any); ok && len(v) > 0 {
		apiObject.SmallMultiplesOptions = expandSmallMultiplesOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandBarChartSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["value_axis"].([]any); ok && len(v) > 0 {
		apiObject.ValueAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["value_label_options"].([]any); ok && len(v) > 0 {
		apiObject.ValueLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]any); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}

	return apiObject
}

func expandBarChartFieldWells(tfList []any) *awstypes.BarChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BarChartFieldWells{}

	if v, ok := tfMap["bar_chart_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.BarChartAggregatedFieldWells = expandBarChartAggregatedFieldWells(v)
	}

	return apiObject
}

func expandBarChartAggregatedFieldWells(tfList []any) *awstypes.BarChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BarChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["colors"].([]any); ok && len(v) > 0 {
		apiObject.Colors = expandDimensionFields(v)
	}
	if v, ok := tfMap["small_multiples"].([]any); ok && len(v) > 0 {
		apiObject.SmallMultiples = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandBarChartSortConfiguration(tfList []any) *awstypes.BarChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BarChartSortConfiguration{}

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
	if v, ok := tfMap["small_multiples_limit_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SmallMultiplesLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["small_multiples_sort"].([]any); ok && len(v) > 0 {
		apiObject.SmallMultiplesSort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func flattenBarChartVisual(apiObject *awstypes.BarChartVisual) []any {
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
		tfMap["chart_configuration"] = flattenBarChartConfiguration(apiObject.ChartConfiguration)
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

func flattenBarChartConfiguration(apiObject *awstypes.BarChartConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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
	if apiObject.ContributionAnalysisDefaults != nil {
		tfMap["contribution_analysis_defaults"] = flattenContributionAnalysisDefault(apiObject.ContributionAnalysisDefaults)
	}
	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenBarChartFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	tfMap["orientation"] = apiObject.Orientation
	if apiObject.ReferenceLines != nil {
		tfMap["reference_lines"] = flattenReferenceLine(apiObject.ReferenceLines)
	}
	if apiObject.SmallMultiplesOptions != nil {
		tfMap["small_multiples_options"] = flattenSmallMultiplesOptions(apiObject.SmallMultiplesOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenBarChartSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}
	if apiObject.ValueAxis != nil {
		tfMap["value_axis"] = flattenAxisDisplayOptions(apiObject.ValueAxis)
	}
	if apiObject.ValueLabelOptions != nil {
		tfMap["value_label_options"] = flattenChartAxisLabelOptions(apiObject.ValueLabelOptions)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}

	return []any{tfMap}
}

func flattenBarChartFieldWells(apiObject *awstypes.BarChartFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BarChartAggregatedFieldWells != nil {
		tfMap["bar_chart_aggregated_field_wells"] = flattenBarChartAggregatedFieldWells(apiObject.BarChartAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenBarChartAggregatedFieldWells(apiObject *awstypes.BarChartAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.Colors != nil {
		tfMap["colors"] = flattenDimensionFields(apiObject.Colors)
	}
	if apiObject.SmallMultiples != nil {
		tfMap["small_multiples"] = flattenDimensionFields(apiObject.SmallMultiples)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenBarChartSortConfiguration(apiObject *awstypes.BarChartSortConfiguration) []any {
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
	if apiObject.SmallMultiplesLimitConfiguration != nil {
		tfMap["small_multiples_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.SmallMultiplesLimitConfiguration)
	}
	if apiObject.SmallMultiplesSort != nil {
		tfMap["small_multiples_sort"] = flattenFieldSortOptions(apiObject.SmallMultiplesSort)
	}

	return []any{tfMap}
}

func flattenItemsLimitConfiguration(apiObject *awstypes.ItemsLimitConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ItemsLimit != nil {
		tfMap["items_limit"] = aws.ToInt64(apiObject.ItemsLimit)
	}
	tfMap["other_categories"] = apiObject.OtherCategories

	return []any{tfMap}
}

func flattenFieldSortOptions(apiObjects []awstypes.FieldSortOptions) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.ColumnSort != nil {
			tfMap["column_sort"] = flattenColumnSort(apiObject.ColumnSort)
		}
		if apiObject.FieldSort != nil {
			tfMap["field_sort"] = flattenFieldSort(apiObject.FieldSort)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnSort(apiObject *awstypes.ColumnSort) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["direction"] = apiObject.Direction
	if apiObject.SortBy != nil {
		tfMap["sort_by"] = flattenColumnIdentifier(apiObject.SortBy)
	}
	if apiObject.AggregationFunction != nil {
		tfMap["aggregation_function"] = flattenAggregationFunction(apiObject.AggregationFunction)
	}

	return []any{tfMap}
}

func flattenFieldSort(apiObject *awstypes.FieldSort) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["direction"] = apiObject.Direction
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}

	return []any{tfMap}
}
