// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func pieChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PieChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PieChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"category_label_options":         chartAxisLabelOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"contribution_analysis_defaults": contributionAnalysisDefaultsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ContributionAnalysisDefault.html
							"data_labels":                    dataLabelOptionsSchema(),             // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"donut_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DonutOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"arc_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ArcOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"arc_thickness": stringSchema(false, validation.StringInSlice(quicksight.ArcThicknessOptions_Values(), false)),
												},
											},
										},
										"donut_center_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DonutCenterOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"label_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
												},
											},
										},
									},
								},
							},
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PieChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"pie_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PieChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"category":        dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"small_multiples": dimensionFieldSchema(1),                          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues:  measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":                  legendOptionsSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"small_multiples_options": smallMultiplesOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SmallMultiplesOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PieChartSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit":                itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":                       fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"small_multiples_limit_configuration": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"small_multiples_sort":                fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
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

func expandPieChartVisual(tfList []interface{}) *quicksight.PieChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.PieChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandPieChartConfiguration(v)
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

func expandPieChartConfiguration(tfList []interface{}) *quicksight.PieChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PieChartConfiguration{}

	if v, ok := tfMap["category_label_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["contribution_analysis_defaults"].([]interface{}); ok && len(v) > 0 {
		config.ContributionAnalysisDefaults = expandContributionAnalysisDefaults(v)
	}
	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		config.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["donut_options"].([]interface{}); ok && len(v) > 0 {
		config.DonutOptions = expandDonutOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandPieChartFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["small_multiples_options"].([]interface{}); ok && len(v) > 0 {
		config.SmallMultiplesOptions = expandSmallMultiplesOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandPieChartSortConfiguration(v)
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

func expandPieChartFieldWells(tfList []interface{}) *quicksight.PieChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PieChartFieldWells{}

	if v, ok := tfMap["pie_chart_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.PieChartAggregatedFieldWells = expandPieChartAggregatedFieldWells(v)
	}

	return config
}

func expandPieChartAggregatedFieldWells(tfList []interface{}) *quicksight.PieChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PieChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		config.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["small_multiples"].([]interface{}); ok && len(v) > 0 {
		config.SmallMultiples = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandPieChartSortConfiguration(tfList []interface{}) *quicksight.PieChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PieChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["small_multiples_limit_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SmallMultiplesLimitConfiguration = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["small_multiples_sort"].([]interface{}); ok && len(v) > 0 {
		config.SmallMultiplesSort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandDonutOptions(tfList []interface{}) *quicksight.DonutOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.DonutOptions{}

	if v, ok := tfMap["arc_options"].([]interface{}); ok && len(v) > 0 {
		options.ArcOptions = expandArcOptions(v)
	}
	if v, ok := tfMap["donut_center_options"].([]interface{}); ok && len(v) > 0 {
		options.DonutCenterOptions = expandDonutCenterOptions(v)
	}

	return options
}

func expandArcOptions(tfList []interface{}) *quicksight.ArcOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ArcOptions{}

	if v, ok := tfMap["arc_thickness"].(string); ok && v != "" {
		options.ArcThickness = aws.String(v)
	}

	return options
}

func expandDonutCenterOptions(tfList []interface{}) *quicksight.DonutCenterOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.DonutCenterOptions{}

	if v, ok := tfMap["label_visibility"].(string); ok && v != "" {
		options.LabelVisibility = aws.String(v)
	}

	return options
}

func flattenPieChartVisual(apiObject *quicksight.PieChartVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id": aws.StringValue(apiObject.VisualId),
	}
	if apiObject.Actions != nil {
		tfMap[names.AttrActions] = flattenVisualCustomAction(apiObject.Actions)
	}
	if apiObject.ChartConfiguration != nil {
		tfMap["chart_configuration"] = flattenPieChartConfiguration(apiObject.ChartConfiguration)
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

func flattenPieChartConfiguration(apiObject *quicksight.PieChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryLabelOptions != nil {
		tfMap["category_label_options"] = flattenChartAxisLabelOptions(apiObject.CategoryLabelOptions)
	}
	if apiObject.ContributionAnalysisDefaults != nil {
		tfMap["contribution_analysis_defaults"] = flattenContributionAnalysisDefault(apiObject.ContributionAnalysisDefaults)
	}
	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.DonutOptions != nil {
		tfMap["donut_options"] = flattenDonutOptions(apiObject.DonutOptions)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenPieChartFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.SmallMultiplesOptions != nil {
		tfMap["small_multiples_options"] = flattenSmallMultiplesOptions(apiObject.SmallMultiplesOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenPieChartSortConfiguration(apiObject.SortConfiguration)
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

func flattenDonutOptions(apiObject *quicksight.DonutOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ArcOptions != nil {
		tfMap["arc_options"] = flattenArcOptions(apiObject.ArcOptions)
	}
	if apiObject.DonutCenterOptions != nil {
		tfMap["donut_center_options"] = flattenDonutCenterOptions(apiObject.DonutCenterOptions)
	}

	return []interface{}{tfMap}
}

func flattenArcOptions(apiObject *quicksight.ArcOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ArcThickness != nil {
		tfMap["arc_thickness"] = aws.StringValue(apiObject.ArcThickness)
	}

	return []interface{}{tfMap}
}

func flattenDonutCenterOptions(apiObject *quicksight.DonutCenterOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LabelVisibility != nil {
		tfMap["label_visibility"] = aws.StringValue(apiObject.LabelVisibility)
	}

	return []interface{}{tfMap}
}

func flattenPieChartFieldWells(apiObject *quicksight.PieChartFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PieChartAggregatedFieldWells != nil {
		tfMap["pie_chart_aggregated_field_wells"] = flattenPieChartAggregatedFieldWells(apiObject.PieChartAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenPieChartAggregatedFieldWells(apiObject *quicksight.PieChartAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.SmallMultiples != nil {
		tfMap["small_multiples"] = flattenDimensionFields(apiObject.SmallMultiples)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenPieChartSortConfiguration(apiObject *quicksight.PieChartSortConfiguration) []interface{} {
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
	if apiObject.SmallMultiplesLimitConfiguration != nil {
		tfMap["small_multiples_limit_configuration"] = flattenItemsLimitConfiguration(apiObject.SmallMultiplesLimitConfiguration)
	}
	if apiObject.SmallMultiplesSort != nil {
		tfMap["small_multiples_sort"] = flattenFieldSortOptions(apiObject.SmallMultiplesSort)
	}

	return []interface{}{tfMap}
}
