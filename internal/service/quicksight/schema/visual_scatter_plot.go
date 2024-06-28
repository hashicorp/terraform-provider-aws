// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
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

func expandScatterPlotVisual(tfList []interface{}) *quicksight.ScatterPlotVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.ScatterPlotVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandScatterPlotConfiguration(v)
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

func expandScatterPlotConfiguration(tfList []interface{}) *quicksight.ScatterPlotConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ScatterPlotConfiguration{}

	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		config.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandScatterPlotFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}
	if v, ok := tfMap["x_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.XAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["x_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.XAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["y_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.YAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["y_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.YAxisLabelOptions = expandChartAxisLabelOptions(v)
	}

	return config
}

func expandScatterPlotFieldWells(tfList []interface{}) *quicksight.ScatterPlotFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ScatterPlotFieldWells{}

	if v, ok := tfMap["scatter_plot_categorically_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.ScatterPlotCategoricallyAggregatedFieldWells = expandScatterPlotCategoricallyAggregatedFieldWells(v)
	}
	if v, ok := tfMap["scatter_plot_unaggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.ScatterPlotUnaggregatedFieldWells = expandScatterPlotUnaggregatedFieldWells(v)
	}

	return config
}

func expandScatterPlotCategoricallyAggregatedFieldWells(tfList []interface{}) *quicksight.ScatterPlotCategoricallyAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ScatterPlotCategoricallyAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		config.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrSize].([]interface{}); ok && len(v) > 0 {
		config.Size = expandMeasureFields(v)
	}
	if v, ok := tfMap["x_axis"].([]interface{}); ok && len(v) > 0 {
		config.XAxis = expandMeasureFields(v)
	}
	if v, ok := tfMap["y_axis"].([]interface{}); ok && len(v) > 0 {
		config.YAxis = expandMeasureFields(v)
	}

	return config
}

func expandScatterPlotUnaggregatedFieldWells(tfList []interface{}) *quicksight.ScatterPlotUnaggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ScatterPlotUnaggregatedFieldWells{}

	if v, ok := tfMap[names.AttrSize].([]interface{}); ok && len(v) > 0 {
		config.Size = expandMeasureFields(v)
	}
	if v, ok := tfMap["x_axis"].([]interface{}); ok && len(v) > 0 {
		config.XAxis = expandDimensionFields(v)
	}
	if v, ok := tfMap["y_axis"].([]interface{}); ok && len(v) > 0 {
		config.YAxis = expandDimensionFields(v)
	}

	return config
}

func flattenScatterPlotVisual(apiObject *quicksight.ScatterPlotVisual) []interface{} {
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

	return []interface{}{tfMap}
}

func flattenScatterPlotConfiguration(apiObject *quicksight.ScatterPlotConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenScatterPlotFieldWells(apiObject *quicksight.ScatterPlotFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ScatterPlotCategoricallyAggregatedFieldWells != nil {
		tfMap["scatter_plot_categorically_aggregated_field_wells"] = flattenScatterPlotCategoricallyAggregatedFieldWells(apiObject.ScatterPlotCategoricallyAggregatedFieldWells)
	}
	if apiObject.ScatterPlotUnaggregatedFieldWells != nil {
		tfMap["scatter_plot_unaggregated_field_wells"] = flattenScatterPlotUnaggregatedFieldWells(apiObject.ScatterPlotUnaggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenScatterPlotCategoricallyAggregatedFieldWells(apiObject *quicksight.ScatterPlotCategoricallyAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenScatterPlotUnaggregatedFieldWells(apiObject *quicksight.ScatterPlotUnaggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Size != nil {
		tfMap[names.AttrSize] = flattenMeasureFields(apiObject.Size)
	}
	if apiObject.XAxis != nil {
		tfMap["x_axis"] = flattenDimensionFields(apiObject.XAxis)
	}
	if apiObject.YAxis != nil {
		tfMap["y_axis"] = flattenDimensionFields(apiObject.YAxis)
	}

	return []interface{}{tfMap}
}
