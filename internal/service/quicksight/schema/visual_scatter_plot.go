// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func scatterPlotVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScatterPlotVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScatterPlotConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_labels": dataLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScatterPlotFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"scatter_plot_categorically_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScatterPlotCategoricallyAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"category":     dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrSize: measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
													"x_axis":       measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
													"y_axis":       measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"scatter_plot_unaggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScatterPlotUnaggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrSize: measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
													"x_axis":       dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"y_axis":       dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
												},
											},
										},
									},
								},
							},
							"legend":                 legendOptionsSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"tooltip":                tooltipOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"visual_palette":         visualPaletteSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
							"x_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"x_axis_label_options":   chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"y_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"y_axis_label_options":   chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
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

func expandScatterPlotVisual(tfList []any) *awstypes.ScatterPlotVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ScatterPlotVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandScatterPlotConfiguration(v)
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

func expandScatterPlotConfiguration(tfList []any) *awstypes.ScatterPlotConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ScatterPlotConfiguration{}

	if v, ok := tfMap["data_labels"].([]any); ok && len(v) > 0 {
		apiObject.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandScatterPlotFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]any); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]any); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}
	if v, ok := tfMap["x_axis_display_options"].([]any); ok && len(v) > 0 {
		apiObject.XAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["x_axis_label_options"].([]any); ok && len(v) > 0 {
		apiObject.XAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["y_axis_display_options"].([]any); ok && len(v) > 0 {
		apiObject.YAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["y_axis_label_options"].([]any); ok && len(v) > 0 {
		apiObject.YAxisLabelOptions = expandChartAxisLabelOptions(v)
	}

	return apiObject
}

func expandScatterPlotFieldWells(tfList []any) *awstypes.ScatterPlotFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ScatterPlotFieldWells{}

	if v, ok := tfMap["scatter_plot_categorically_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.ScatterPlotCategoricallyAggregatedFieldWells = expandScatterPlotCategoricallyAggregatedFieldWells(v)
	}
	if v, ok := tfMap["scatter_plot_unaggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.ScatterPlotUnaggregatedFieldWells = expandScatterPlotUnaggregatedFieldWells(v)
	}

	return apiObject
}

func expandScatterPlotCategoricallyAggregatedFieldWells(tfList []any) *awstypes.ScatterPlotCategoricallyAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ScatterPlotCategoricallyAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrSize].([]any); ok && len(v) > 0 {
		apiObject.Size = expandMeasureFields(v)
	}
	if v, ok := tfMap["x_axis"].([]any); ok && len(v) > 0 {
		apiObject.XAxis = expandMeasureFields(v)
	}
	if v, ok := tfMap["y_axis"].([]any); ok && len(v) > 0 {
		apiObject.YAxis = expandMeasureFields(v)
	}

	return apiObject
}

func expandScatterPlotUnaggregatedFieldWells(tfList []any) *awstypes.ScatterPlotUnaggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ScatterPlotUnaggregatedFieldWells{}

	if v, ok := tfMap[names.AttrSize].([]any); ok && len(v) > 0 {
		apiObject.Size = expandMeasureFields(v)
	}
	if v, ok := tfMap["x_axis"].([]any); ok && len(v) > 0 {
		apiObject.XAxis = expandDimensionFields(v)
	}
	if v, ok := tfMap["y_axis"].([]any); ok && len(v) > 0 {
		apiObject.YAxis = expandDimensionFields(v)
	}

	return apiObject
}

func flattenScatterPlotVisual(apiObject *awstypes.ScatterPlotVisual) []any {
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
		tfMap["chart_configuration"] = flattenScatterPlotConfiguration(apiObject.ChartConfiguration)
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

func flattenScatterPlotConfiguration(apiObject *awstypes.ScatterPlotConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenScatterPlotFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}
	if apiObject.XAxisDisplayOptions != nil {
		tfMap["x_axis_display_options"] = flattenAxisDisplayOptions(apiObject.XAxisDisplayOptions)
	}
	if apiObject.XAxisLabelOptions != nil {
		tfMap["x_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.XAxisLabelOptions)
	}
	if apiObject.YAxisDisplayOptions != nil {
		tfMap["y_axis_display_options"] = flattenAxisDisplayOptions(apiObject.YAxisDisplayOptions)
	}
	if apiObject.YAxisLabelOptions != nil {
		tfMap["y_axis_label_options"] = flattenChartAxisLabelOptions(apiObject.YAxisLabelOptions)
	}

	return []any{tfMap}
}

func flattenScatterPlotFieldWells(apiObject *awstypes.ScatterPlotFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}
	if apiObject.ScatterPlotCategoricallyAggregatedFieldWells != nil {
		tfMap["scatter_plot_categorically_aggregated_field_wells"] = flattenScatterPlotCategoricallyAggregatedFieldWells(apiObject.ScatterPlotCategoricallyAggregatedFieldWells)
	}
	if apiObject.ScatterPlotUnaggregatedFieldWells != nil {
		tfMap["scatter_plot_unaggregated_field_wells"] = flattenScatterPlotUnaggregatedFieldWells(apiObject.ScatterPlotUnaggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenScatterPlotCategoricallyAggregatedFieldWells(apiObject *awstypes.ScatterPlotCategoricallyAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionFields(apiObject.Category)
	}
	if apiObject.Size != nil {
		tfMap[names.AttrSize] = flattenMeasureFields(apiObject.Size)
	}
	if apiObject.XAxis != nil {
		tfMap["x_axis"] = flattenMeasureFields(apiObject.XAxis)
	}
	if apiObject.YAxis != nil {
		tfMap["y_axis"] = flattenMeasureFields(apiObject.YAxis)
	}

	return []any{tfMap}
}

func flattenScatterPlotUnaggregatedFieldWells(apiObject *awstypes.ScatterPlotUnaggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Size != nil {
		tfMap[names.AttrSize] = flattenMeasureFields(apiObject.Size)
	}
	if apiObject.XAxis != nil {
		tfMap["x_axis"] = flattenDimensionFields(apiObject.XAxis)
	}
	if apiObject.YAxis != nil {
		tfMap["y_axis"] = flattenDimensionFields(apiObject.YAxis)
	}

	return []any{tfMap}
}
