// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func boxPlotVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
										"all_data_points_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
										"outlier_visibility":         stringEnumSchema[awstypes.Visibility](attrOptional),
										"style_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotStyleOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"fill_style": stringEnumSchema[awstypes.BoxPlotFillStyle](attrOptional),
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
													"group_by":       dimensionFieldSchema(1),                    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(measureFieldsMaxItems5), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":                         legendOptionsSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"primary_y_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"primary_y_axis_label_options":   chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"reference_lines":                referenceLineSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BoxPlotSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_sort":            fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"pagination_configuration": paginationConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PaginationConfiguration.html
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
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntAtLeast(0),
				},
				"page_size": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		},
	}
}

func expandBoxPlotVisual(tfList []any) *awstypes.BoxPlotVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BoxPlotVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandBoxPlotChartConfiguration(v)
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

func expandBoxPlotChartConfiguration(tfList []any) *awstypes.BoxPlotChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BoxPlotChartConfiguration{}

	if v, ok := tfMap["box_plot_options"].([]any); ok && len(v) > 0 {
		apiObject.BoxPlotOptions = expandBoxPlotOptions(v)
	}
	if v, ok := tfMap["category_axis"].([]any); ok && len(v) > 0 {
		apiObject.CategoryAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["category_label_options"].([]any); ok && len(v) > 0 {
		apiObject.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandBoxPlotFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
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
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandBoxPlotSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]any); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}

	return apiObject
}

func expandBoxPlotOptions(tfList []any) *awstypes.BoxPlotOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BoxPlotOptions{}

	if v, ok := tfMap["all_data_points_visibility"].(string); ok && v != "" {
		apiObject.AllDataPointsVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["outlier_visibility"].(string); ok && v != "" {
		apiObject.OutlierVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["style_options"].([]any); ok && len(v) > 0 {
		apiObject.StyleOptions = expandBoxPlotStyleOptions(v)
	}

	return apiObject
}

func expandBoxPlotStyleOptions(tfList []any) *awstypes.BoxPlotStyleOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BoxPlotStyleOptions{}

	if v, ok := tfMap["fill_style"].(string); ok && v != "" {
		apiObject.FillStyle = awstypes.BoxPlotFillStyle(v)
	}

	return apiObject
}

func expandBoxPlotFieldWells(tfList []any) *awstypes.BoxPlotFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BoxPlotFieldWells{}

	if v, ok := tfMap["box_plot_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.BoxPlotAggregatedFieldWells = expandBoxPlotAggregatedFieldWells(v)
	}

	return apiObject
}

func expandBoxPlotAggregatedFieldWells(tfList []any) *awstypes.BoxPlotAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BoxPlotAggregatedFieldWells{}

	if v, ok := tfMap["group_by"].([]any); ok && len(v) > 0 {
		apiObject.GroupBy = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandBoxPlotSortConfiguration(tfList []any) *awstypes.BoxPlotSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BoxPlotSortConfiguration{}

	if v, ok := tfMap["category_sort"].([]any); ok && len(v) > 0 {
		apiObject.CategorySort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["pagination_configuration"].([]any); ok && len(v) > 0 {
		apiObject.PaginationConfiguration = expandPaginationConfiguration(v)
	}

	return apiObject
}

func expandPaginationConfiguration(tfList []any) *awstypes.PaginationConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PaginationConfiguration{}

	if v, ok := tfMap["page_number"].(int); ok {
		apiObject.PageNumber = aws.Int64(int64(v))
	}
	if v, ok := tfMap["page_size"].(int); ok {
		apiObject.PageSize = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenBoxPlotVisual(apiObject *awstypes.BoxPlotVisual) []any {
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

	return []any{tfMap}
}

func flattenBoxPlotChartConfiguration(apiObject *awstypes.BoxPlotChartConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}

func flattenBoxPlotOptions(apiObject *awstypes.BoxPlotOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["all_data_points_visibility"] = apiObject.AllDataPointsVisibility
	tfMap["outlier_visibility"] = apiObject.OutlierVisibility
	if apiObject.StyleOptions != nil {
		tfMap["style_options"] = flattenBoxPlotStyleOptions(apiObject.StyleOptions)
	}

	return []any{tfMap}
}

func flattenBoxPlotStyleOptions(apiObject *awstypes.BoxPlotStyleOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"fill_style": apiObject.FillStyle,
	}

	return []any{tfMap}
}

func flattenBoxPlotFieldWells(apiObject *awstypes.BoxPlotFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BoxPlotAggregatedFieldWells != nil {
		tfMap["box_plot_aggregated_field_wells"] = flattenBoxPlotAggregatedFieldWells(apiObject.BoxPlotAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenBoxPlotAggregatedFieldWells(apiObject *awstypes.BoxPlotAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.GroupBy != nil {
		tfMap["group_by"] = flattenDimensionFields(apiObject.GroupBy)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenBoxPlotSortConfiguration(apiObject *awstypes.BoxPlotSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}
	if apiObject.PaginationConfiguration != nil {
		tfMap["pagination_configuration"] = flattenPaginationConfiguration(apiObject.PaginationConfiguration)
	}

	return []any{tfMap}
}

func flattenPaginationConfiguration(apiObject *awstypes.PaginationConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.PageNumber != nil {
		tfMap["page_number"] = aws.ToInt64(apiObject.PageNumber)
	}
	if apiObject.PageSize != nil {
		tfMap["page_size"] = aws.ToInt64(apiObject.PageSize)
	}

	return []any{tfMap}
}
