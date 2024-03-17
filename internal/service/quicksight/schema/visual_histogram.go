// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

func histogramVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HistogramVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HistogramConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"bin_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HistogramBinOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"bin_count": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BinCountOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"value": {
														Type:             schema.TypeInt,
														Optional:         true,
														ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
													},
												},
											},
										},
										"bin_width": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BinWidthOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"bin_count_limit": {
														Type:             schema.TypeInt,
														Optional:         true,
														ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 1000)),
													},
													"value": {
														Type:             schema.TypeFloat,
														Optional:         true,
														ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
													},
												},
											},
										},
										"selected_bin_type": stringSchema(false, enum.Validate[types.HistogramBinType]()),
										"start_value": {
											Type:     schema.TypeFloat,
											Optional: true,
										},
									},
								},
							},
							"data_labels": dataLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HistogramFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"histogram_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HistogramAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"values": measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"tooltip":                tooltipOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"visual_palette":         visualPaletteSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
							"x_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"x_axis_label_options":   chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"y_axis_display_options": axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
						},
					},
				},
				"subtitle": visualSubtitleLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
				"title":    visualTitleLabelOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
			},
		},
	}
}

func expandHistogramVisual(tfList []interface{}) *types.HistogramVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.HistogramVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandHistogramConfiguration(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		visual.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		visual.Title = expandVisualTitleLabelOptions(v)
	}

	return visual
}

func expandHistogramConfiguration(tfList []interface{}) *types.HistogramConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.HistogramConfiguration{}

	if v, ok := tfMap["bin_options"].([]interface{}); ok && len(v) > 0 {
		config.BinOptions = expandHistogramBinOptions(v)
	}
	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		config.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandHistogramFieldWells(v)
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

	return config
}

func expandHistogramFieldWells(tfList []interface{}) *types.HistogramFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.HistogramFieldWells{}

	if v, ok := tfMap["histogram_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.HistogramAggregatedFieldWells = expandHistogramAggregatedFieldWells(v)
	}

	return config
}

func expandHistogramAggregatedFieldWells(tfList []interface{}) *types.HistogramAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.HistogramAggregatedFieldWells{}

	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandHistogramBinOptions(tfList []interface{}) *types.HistogramBinOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.HistogramBinOptions{}

	if v, ok := tfMap["selected_bin_type"].(string); ok && v != "" {
		options.SelectedBinType = types.HistogramBinType(v)
	}
	if v, ok := tfMap["start_value"].(float64); ok {
		options.StartValue = aws.Float64(v)
	}
	if v, ok := tfMap["bin_count"].([]interface{}); ok && len(v) > 0 {
		options.BinCount = expandBinCountOptions(v)
	}
	if v, ok := tfMap["bin_width"].([]interface{}); ok && len(v) > 0 {
		options.BinWidth = expandBinWidthOptions(v)
	}

	return options
}

func expandBinCountOptions(tfList []interface{}) *types.BinCountOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.BinCountOptions{}

	if v, ok := tfMap["value"].(int); ok {
		options.Value = aws.Int32(int32(v))
	}

	return options
}

func expandBinWidthOptions(tfList []interface{}) *types.BinWidthOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.BinWidthOptions{}

	if v, ok := tfMap["bin_count_limit"].(int); ok {
		options.BinCountLimit = aws.Int64(int64(v))
	}
	if v, ok := tfMap["value"].(float64); ok {
		options.Value = aws.Float64(v)
	}

	return options
}

func flattenHistogramVisual(apiObject *types.HistogramVisual) []interface{} {
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
		tfMap["chart_configuration"] = flattenHistogramConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []interface{}{tfMap}
}

func flattenHistogramConfiguration(apiObject *types.HistogramConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BinOptions != nil {
		tfMap["bin_options"] = flattenHistogramBinOptions(apiObject.BinOptions)
	}
	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenHistogramFieldWells(apiObject.FieldWells)
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

	return []interface{}{tfMap}
}

func flattenHistogramBinOptions(apiObject *types.HistogramBinOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BinCount != nil {
		tfMap["bin_count"] = flattenBinCountOptions(apiObject.BinCount)
	}
	if apiObject.BinWidth != nil {
		tfMap["bin_width"] = flattenBinWidthOptions(apiObject.BinWidth)
	}

	tfMap["selected_bin_type"] = types.HistogramBinType(apiObject.SelectedBinType)

	if apiObject.StartValue != nil {
		tfMap["start_value"] = aws.ToFloat64(apiObject.StartValue)
	}

	return []interface{}{tfMap}
}

func flattenBinCountOptions(apiObject *types.BinCountOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Value != nil {
		tfMap["value"] = aws.ToInt32(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenBinWidthOptions(apiObject *types.BinWidthOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BinCountLimit != nil {
		tfMap["bin_count_limit"] = aws.ToInt64(apiObject.BinCountLimit)
	}
	if apiObject.Value != nil {
		tfMap["value"] = aws.ToFloat64(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenHistogramFieldWells(apiObject *types.HistogramFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.HistogramAggregatedFieldWells != nil {
		tfMap["histogram_aggregated_field_wells"] = flattenHistogramAggregatedFieldWells(apiObject.HistogramAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenHistogramAggregatedFieldWells(apiObject *types.HistogramAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Values != nil {
		tfMap["values"] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}
