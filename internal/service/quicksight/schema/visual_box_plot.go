// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func boxPlotVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"box_plot_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"all_data_points_visibility": stringSchema(false, enum.Validate[types.Visibility]()),
										"outlier_visibility":         stringSchema(false, enum.Validate[types.Visibility]()),
										"style_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotStyleOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"fill_style": stringSchema(false, enum.Validate[types.BoxPlotFillStyle]()),
												},
											},
										},
									},
								},
							},
							"category_axis":          axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"category_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"box_plot_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"group_by": dimensionFieldSchema(1),                    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"values":   measureFieldSchema(measureFieldsMaxItems5), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":                         legendOptionsSchema(),                       // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"primary_y_axis_display_options": axisDisplayOptionsSchema(),                  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"primary_y_axis_label_options":   chartAxisLabelOptionsSchema(),               // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"reference_lines":                referenceLineSchema(referenceLinesMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_sort":            fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"pagination_configuration": paginationConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PaginationConfiguration.html
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

func paginationConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PaginationConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"page_number": {
					Type:             schema.TypeInt,
					Required:         true,
					ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
				},
				"page_size": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		},
	}
}

func expandBoxPlotVisual(tfList []interface{}) *types.BoxPlotVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.BoxPlotVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandBoxPlotChartConfiguration(v)
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

func expandBoxPlotChartConfiguration(tfList []interface{}) *types.BoxPlotChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.BoxPlotChartConfiguration{}

	if v, ok := tfMap["box_plot_options"].([]interface{}); ok && len(v) > 0 {
		config.BoxPlotOptions = expandBoxPlotOptions(v)
	}
	if v, ok := tfMap["category_axis"].([]interface{}); ok && len(v) > 0 {
		config.CategoryAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["category_label_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandBoxPlotFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.PrimaryYAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.PrimaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["reference_lines"].([]interface{}); ok && len(v) > 0 {
		config.ReferenceLines = expandReferenceLines(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandBoxPlotSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}

	return config
}

func expandBoxPlotOptions(tfList []interface{}) *types.BoxPlotOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.BoxPlotOptions{}

	if v, ok := tfMap["all_data_points_visibility"].(string); ok && v != "" {
		options.AllDataPointsVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["outlier_visibility"].(string); ok && v != "" {
		options.OutlierVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["style_options"].([]interface{}); ok && len(v) > 0 {
		options.StyleOptions = expandBoxPlotStyleOptions(v)
	}

	return options
}

func expandBoxPlotStyleOptions(tfList []interface{}) *types.BoxPlotStyleOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.BoxPlotStyleOptions{}

	if v, ok := tfMap["fill_style"].(string); ok && v != "" {
		options.FillStyle = types.BoxPlotFillStyle(v)
	}

	return options
}

func expandBoxPlotFieldWells(tfList []interface{}) *types.BoxPlotFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.BoxPlotFieldWells{}

	if v, ok := tfMap["box_plot_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.BoxPlotAggregatedFieldWells = expandBoxPlotAggregatedFieldWells(v)
	}

	return config
}

func expandBoxPlotAggregatedFieldWells(tfList []interface{}) *types.BoxPlotAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.BoxPlotAggregatedFieldWells{}

	if v, ok := tfMap["group_by"].([]interface{}); ok && len(v) > 0 {
		config.GroupBy = expandDimensionFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandBoxPlotSortConfiguration(tfList []interface{}) *types.BoxPlotSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.BoxPlotSortConfiguration{}

	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["pagination_configuration"].([]interface{}); ok && len(v) > 0 {
		config.PaginationConfiguration = expandPaginationConfiguration(v)
	}

	return config
}

func expandPaginationConfiguration(tfList []interface{}) *types.PaginationConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.PaginationConfiguration{}

	if v, ok := tfMap["page_number"].(int); ok {
		config.PageNumber = aws.Int64(int64(v))
	}
	if v, ok := tfMap["page_size"].(int); ok {
		config.PageSize = aws.Int64(int64(v))
	}

	return config
}

func flattenBoxPlotVisual(apiObject *types.BoxPlotVisual) []interface{} {
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
		tfMap["chart_configuration"] = flattenBoxPlotChartConfiguration(apiObject.ChartConfiguration)
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

func flattenBoxPlotChartConfiguration(apiObject *types.BoxPlotChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BoxPlotOptions != nil {
		tfMap["box_plot_options"] = flattenBoxPlotOptions(apiObject.BoxPlotOptions)
	}
	if apiObject.CategoryAxis != nil {
		tfMap["category_axis"] = flattenAxisDisplayOptions(apiObject.CategoryAxis)
	}
	if apiObject.CategoryLabelOptions != nil {
		tfMap["category_label_options"] = flattenChartAxisLabelOptions(apiObject.CategoryLabelOptions)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenBoxPlotFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
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
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenBoxPlotSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}

	return []interface{}{tfMap}
}

func flattenBoxPlotOptions(apiObject *types.BoxPlotOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["all_data_points_visibility"] = types.Visibility(apiObject.AllDataPointsVisibility)

	tfMap["outlier_visibility"] = types.Visibility(apiObject.OutlierVisibility)

	if apiObject.StyleOptions != nil {
		tfMap["style_options"] = flattenBoxPlotStyleOptions(apiObject.StyleOptions)
	}

	return []interface{}{tfMap}
}

func flattenBoxPlotStyleOptions(apiObject *types.BoxPlotStyleOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["fill_style"] = types.BoxPlotFillStyle(apiObject.FillStyle)

	return []interface{}{tfMap}
}

func flattenBoxPlotFieldWells(apiObject *types.BoxPlotFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BoxPlotAggregatedFieldWells != nil {
		tfMap["box_plot_aggregated_field_wells"] = flattenBoxPlotAggregatedFieldWells(apiObject.BoxPlotAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenBoxPlotAggregatedFieldWells(apiObject *types.BoxPlotAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.GroupBy != nil {
		tfMap["group_by"] = flattenDimensionFields(apiObject.GroupBy)
	}
	if apiObject.Values != nil {
		tfMap["values"] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenBoxPlotSortConfiguration(apiObject *types.BoxPlotSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}
	if apiObject.PaginationConfiguration != nil {
		tfMap["pagination_configuration"] = flattenPaginationConfiguration(apiObject.PaginationConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenPaginationConfiguration(apiObject *types.PaginationConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PageNumber != nil {
		tfMap["page_number"] = aws.ToInt64(apiObject.PageNumber)
	}
	if apiObject.PageSize != nil {
		tfMap["page_size"] = aws.ToInt64(apiObject.PageSize)
	}

	return []interface{}{tfMap}
}
