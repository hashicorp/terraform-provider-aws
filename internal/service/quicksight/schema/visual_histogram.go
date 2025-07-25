// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func histogramVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HistogramVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
													names.AttrValue: {
														Type:         schema.TypeInt,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(0),
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
													"bin_count_limit": intBetweenSchema(attrOptional, 0, 1000),
													names.AttrValue: {
														Type:         schema.TypeFloat,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(0),
													},
												},
											},
										},
										"selected_bin_type": stringEnumSchema[awstypes.HistogramBinType](attrOptional),
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
													names.AttrValues: measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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

func expandHistogramVisual(tfList []any) *awstypes.HistogramVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HistogramVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandHistogramConfiguration(v)
	}
	if v, ok := tfMap["subtitle"].([]any); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandHistogramConfiguration(tfList []any) *awstypes.HistogramConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HistogramConfiguration{}

	if v, ok := tfMap["bin_options"].([]any); ok && len(v) > 0 {
		apiObject.BinOptions = expandHistogramBinOptions(v)
	}
	if v, ok := tfMap["data_labels"].([]any); ok && len(v) > 0 {
		apiObject.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandHistogramFieldWells(v)
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

	return apiObject
}

func expandHistogramFieldWells(tfList []any) *awstypes.HistogramFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HistogramFieldWells{}

	if v, ok := tfMap["histogram_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.HistogramAggregatedFieldWells = expandHistogramAggregatedFieldWells(v)
	}

	return apiObject
}

func expandHistogramAggregatedFieldWells(tfList []any) *awstypes.HistogramAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HistogramAggregatedFieldWells{}

	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandHistogramBinOptions(tfList []any) *awstypes.HistogramBinOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HistogramBinOptions{}

	if v, ok := tfMap["selected_bin_type"].(string); ok && v != "" {
		apiObject.SelectedBinType = awstypes.HistogramBinType(v)
	}
	if v, ok := tfMap["start_value"].(float64); ok {
		apiObject.StartValue = aws.Float64(v)
	}
	if v, ok := tfMap["bin_count"].([]any); ok && len(v) > 0 {
		apiObject.BinCount = expandBinCountOptions(v)
	}
	if v, ok := tfMap["bin_width"].([]any); ok && len(v) > 0 {
		apiObject.BinWidth = expandBinWidthOptions(v)
	}

	return apiObject
}

func expandBinCountOptions(tfList []any) *awstypes.BinCountOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BinCountOptions{}

	if v, ok := tfMap[names.AttrValue].(int); ok {
		apiObject.Value = aws.Int32(int32(v))
	}

	return apiObject
}

func expandBinWidthOptions(tfList []any) *awstypes.BinWidthOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BinWidthOptions{}

	if v, ok := tfMap["bin_count_limit"].(int); ok {
		apiObject.BinCountLimit = aws.Int64(int64(v))
	}
	if v, ok := tfMap[names.AttrValue].(float64); ok {
		apiObject.Value = aws.Float64(v)
	}

	return apiObject
}

func flattenHistogramVisual(apiObject *awstypes.HistogramVisual) []any {
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
		tfMap["chart_configuration"] = flattenHistogramConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []any{tfMap}
}

func flattenHistogramConfiguration(apiObject *awstypes.HistogramConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}

func flattenHistogramBinOptions(apiObject *awstypes.HistogramBinOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BinCount != nil {
		tfMap["bin_count"] = flattenBinCountOptions(apiObject.BinCount)
	}
	if apiObject.BinWidth != nil {
		tfMap["bin_width"] = flattenBinWidthOptions(apiObject.BinWidth)
	}
	tfMap["selected_bin_type"] = apiObject.SelectedBinType
	if apiObject.StartValue != nil {
		tfMap["start_value"] = aws.ToFloat64(apiObject.StartValue)
	}

	return []any{tfMap}
}

func flattenBinCountOptions(apiObject *awstypes.BinCountOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Value != nil {
		tfMap[names.AttrValue] = aws.ToInt32(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenBinWidthOptions(apiObject *awstypes.BinWidthOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BinCountLimit != nil {
		tfMap["bin_count_limit"] = aws.ToInt64(apiObject.BinCountLimit)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = aws.ToFloat64(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenHistogramFieldWells(apiObject *awstypes.HistogramFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.HistogramAggregatedFieldWells != nil {
		tfMap["histogram_aggregated_field_wells"] = flattenHistogramAggregatedFieldWells(apiObject.HistogramAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenHistogramAggregatedFieldWells(apiObject *awstypes.HistogramAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}
