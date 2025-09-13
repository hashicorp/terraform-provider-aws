// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func gaugeChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_labels": dataLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"target_values":  measureFieldSchema(measureFieldsMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
										names.AttrValues: measureFieldSchema(measureFieldsMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
									},
								},
							},
							"gauge_chart_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"arc": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ArcConfiguration.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"arc_angle": {
														Type:     schema.TypeFloat,
														Optional: true,
													},
													"arc_thickness": stringEnumSchema[awstypes.ArcThicknessOptions](attrOptional),
												},
											},
										},
										"arc_axis": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ArcAxisConfiguration.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"range": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ArcAxisDisplayRange.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrMax: {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
																names.AttrMin: {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
															},
														},
													},
													"reserve_range": {
														Type:     schema.TypeInt,
														Optional: true,
													},
												},
											},
										},
										"comparison":                       comparisonConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComparisonConfiguration.html
										"primary_value_display_type":       stringEnumSchema[awstypes.PrimaryValueDisplayType](attrOptional),
										"primary_value_font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
									},
								},
							},
							"tooltip":        tooltipOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"visual_palette": visualPaletteSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
						},
					},
				},
				"conditional_formatting": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartConditionalFormatting.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"conditional_formatting_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartConditionalFormattingOption.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 100,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"arc": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartArcConditionalFormatting.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"foreground_color": conditionalFormattingColorSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingColor.html
												},
											},
										},
										"primary_value": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartPrimaryValueConditionalFormatting.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"icon":       conditionalFormattingIconSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingIcon.html
													"text_color": conditionalFormattingColorSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingColor.html
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

func expandGaugeChartVisual(tfList []any) *awstypes.GaugeChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GaugeChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandGaugeChartConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]any); ok && len(v) > 0 {
		apiObject.ConditionalFormatting = expandGaugeChartConditionalFormatting(v)
	}
	if v, ok := tfMap["subtitle"].([]any); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandGaugeChartConfiguration(tfList []any) *awstypes.GaugeChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GaugeChartConfiguration{}

	if v, ok := tfMap["data_labels"].([]any); ok && len(v) > 0 {
		apiObject.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandGaugeChartFieldWells(v)
	}
	if v, ok := tfMap["gauge_chart_options"].([]any); ok && len(v) > 0 {
		apiObject.GaugeChartOptions = expandGaugeChartOptions(v)
	}
	if v, ok := tfMap["tooltip"].([]any); ok && len(v) > 0 {
		apiObject.TooltipOptions = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]any); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}

	return apiObject
}

func expandGaugeChartFieldWells(tfList []any) *awstypes.GaugeChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GaugeChartFieldWells{}

	if v, ok := tfMap["target_values"].([]any); ok && len(v) > 0 {
		apiObject.TargetValues = expandMeasureFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandGaugeChartOptions(tfList []any) *awstypes.GaugeChartOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GaugeChartOptions{}

	if v, ok := tfMap["primary_value_display_type"].(string); ok && v != "" {
		apiObject.PrimaryValueDisplayType = awstypes.PrimaryValueDisplayType(v)
	}
	if v, ok := tfMap["primary_value_font_configuration"].([]any); ok && len(v) > 0 {
		apiObject.PrimaryValueFontConfiguration = expandFontConfiguration(v)
	}
	if v, ok := tfMap["arc"].([]any); ok && len(v) > 0 {
		apiObject.Arc = expandArcConfiguration(v)
	}
	if v, ok := tfMap["arc_axis"].([]any); ok && len(v) > 0 {
		apiObject.ArcAxis = expandArcAxisConfiguration(v)
	}
	if v, ok := tfMap["comparison"].([]any); ok && len(v) > 0 {
		apiObject.Comparison = expandComparisonConfiguration(v)
	}

	return apiObject
}

func expandGaugeChartConditionalFormatting(tfList []any) *awstypes.GaugeChartConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GaugeChartConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]any); ok && len(v) > 0 {
		apiObject.ConditionalFormattingOptions = expandGaugeChartConditionalFormattingOptions(v)
	}

	return apiObject
}

func expandGaugeChartConditionalFormattingOptions(tfList []any) []awstypes.GaugeChartConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.GaugeChartConditionalFormattingOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandGaugeChartConditionalFormattingOption(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandGaugeChartConditionalFormattingOption(tfMap map[string]any) *awstypes.GaugeChartConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GaugeChartConditionalFormattingOption{}

	if v, ok := tfMap["arc"].([]any); ok && len(v) > 0 {
		apiObject.Arc = expandGaugeChartArcConditionalFormatting(v)
	}
	if v, ok := tfMap["primary_value"].([]any); ok && len(v) > 0 {
		apiObject.PrimaryValue = expandGaugeChartPrimaryValueConditionalFormatting(v)
	}

	return apiObject
}

func expandGaugeChartArcConditionalFormatting(tfList []any) *awstypes.GaugeChartArcConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GaugeChartArcConditionalFormatting{}

	if v, ok := tfMap["foreground_color"].([]any); ok && len(v) > 0 {
		apiObject.ForegroundColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func expandGaugeChartPrimaryValueConditionalFormatting(tfList []any) *awstypes.GaugeChartPrimaryValueConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GaugeChartPrimaryValueConditionalFormatting{}

	if v, ok := tfMap["icon"].([]any); ok && len(v) > 0 {
		apiObject.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]any); ok && len(v) > 0 {
		apiObject.TextColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func expandArcConfiguration(tfList []any) *awstypes.ArcConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ArcConfiguration{}

	if v, ok := tfMap["arc_angle"].(float64); ok {
		apiObject.ArcAngle = aws.Float64(v)
	}
	if v, ok := tfMap["arc_thickness"].(string); ok && v != "" {
		apiObject.ArcThickness = awstypes.ArcThicknessOptions(v)
	}

	return apiObject
}

func expandArcAxisConfiguration(tfList []any) *awstypes.ArcAxisConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ArcAxisConfiguration{}

	if v, ok := tfMap["range"].([]any); ok && len(v) > 0 {
		apiObject.Range = expandArcAxisDisplayRange(v)
	}
	if v, ok := tfMap["reserve_range"].(int); ok {
		apiObject.ReserveRange = int32(v)
	}

	return apiObject
}

func expandArcAxisDisplayRange(tfList []any) *awstypes.ArcAxisDisplayRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ArcAxisDisplayRange{}

	if v, ok := tfMap[names.AttrMax].(float64); ok {
		apiObject.Max = aws.Float64(v)
	}
	if v, ok := tfMap[names.AttrMin].(float64); ok {
		apiObject.Min = aws.Float64(v)
	}

	return apiObject
}

func flattenGaugeChartVisual(apiObject *awstypes.GaugeChartVisual) []any {
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
		tfMap["chart_configuration"] = flattenGaugeChartConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.ConditionalFormatting != nil {
		tfMap["conditional_formatting"] = flattenGaugeChartConditionalFormatting(apiObject.ConditionalFormatting)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []any{tfMap}
}

func flattenGaugeChartConfiguration(apiObject *awstypes.GaugeChartConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenGaugeChartFieldWells(apiObject.FieldWells)
	}
	if apiObject.GaugeChartOptions != nil {
		tfMap["gauge_chart_options"] = flattenGaugeChartOptions(apiObject.GaugeChartOptions)
	}
	if apiObject.TooltipOptions != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.TooltipOptions)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}

	return []any{tfMap}
}

func flattenGaugeChartFieldWells(apiObject *awstypes.GaugeChartFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.TargetValues != nil {
		tfMap["target_values"] = flattenMeasureFields(apiObject.TargetValues)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenGaugeChartOptions(apiObject *awstypes.GaugeChartOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Arc != nil {
		tfMap["arc"] = flattenArcConfiguration(apiObject.Arc)
	}
	if apiObject.ArcAxis != nil {
		tfMap["arc_axis"] = flattenArcAxisConfiguration(apiObject.ArcAxis)
	}
	if apiObject.Comparison != nil {
		tfMap["comparison"] = flattenComparisonConfiguration(apiObject.Comparison)
	}
	tfMap["primary_value_display_type"] = apiObject.PrimaryValueDisplayType
	if apiObject.PrimaryValueFontConfiguration != nil {
		tfMap["primary_value_font_configuration"] = flattenFontConfiguration(apiObject.PrimaryValueFontConfiguration)
	}

	return []any{tfMap}
}

func flattenArcConfiguration(apiObject *awstypes.ArcConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ArcAngle != nil {
		tfMap["arc_angle"] = aws.ToFloat64(apiObject.ArcAngle)
	}
	tfMap["arc_thickness"] = apiObject.ArcThickness

	return []any{tfMap}
}

func flattenArcAxisConfiguration(apiObject *awstypes.ArcAxisConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Range != nil {
		tfMap["range"] = flattenArcAxisDisplayRange(apiObject.Range)
	}
	tfMap["reserve_range"] = apiObject.ReserveRange

	return []any{tfMap}
}

func flattenArcAxisDisplayRange(apiObject *awstypes.ArcAxisDisplayRange) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Max != nil {
		tfMap[names.AttrMax] = aws.ToFloat64(apiObject.Max)
	}
	if apiObject.Min != nil {
		tfMap[names.AttrMin] = aws.ToFloat64(apiObject.Min)
	}

	return []any{tfMap}
}

func flattenComparisonConfiguration(apiObject *awstypes.ComparisonConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComparisonFormat != nil {
		tfMap["comparison_format"] = flattenComparisonFormatConfiguration(apiObject.ComparisonFormat)
	}
	tfMap["comparison_method"] = apiObject.ComparisonMethod

	return []any{tfMap}
}

func flattenComparisonFormatConfiguration(apiObject *awstypes.ComparisonFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.NumberDisplayFormatConfiguration != nil {
		tfMap["number_display_format_configuration"] = flattenNumberDisplayFormatConfiguration(apiObject.NumberDisplayFormatConfiguration)
	}
	if apiObject.PercentageDisplayFormatConfiguration != nil {
		tfMap["percentage_display_format_configuration"] = flattenPercentageDisplayFormatConfiguration(apiObject.PercentageDisplayFormatConfiguration)
	}

	return []any{tfMap}
}

func flattenGaugeChartConditionalFormatting(apiObject *awstypes.GaugeChartConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenGaugeChartConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []any{tfMap}
}

func flattenGaugeChartConditionalFormattingOption(apiObjects []awstypes.GaugeChartConditionalFormattingOption) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Arc != nil {
			tfMap["arc"] = flattenGaugeChartArcConditionalFormatting(apiObject.Arc)
		}
		if apiObject.PrimaryValue != nil {
			tfMap["primary_value"] = flattenGaugeChartPrimaryValueConditionalFormatting(apiObject.PrimaryValue)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenGaugeChartArcConditionalFormatting(apiObject *awstypes.GaugeChartArcConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ForegroundColor != nil {
		tfMap["foreground_color"] = flattenConditionalFormattingColor(apiObject.ForegroundColor)
	}

	return []any{tfMap}
}

func flattenGaugeChartPrimaryValueConditionalFormatting(apiObject *awstypes.GaugeChartPrimaryValueConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Icon != nil {
		tfMap["icon"] = flattenConditionalFormattingIcon(apiObject.Icon)
	}
	if apiObject.TextColor != nil {
		tfMap["text_color"] = flattenConditionalFormattingColor(apiObject.TextColor)
	}

	return []any{tfMap}
}
