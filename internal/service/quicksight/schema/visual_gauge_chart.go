// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
													"arc_thickness": stringSchema(false, validation.StringInSlice(quicksight.ArcThicknessOptions_Values(), false)),
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
										"primary_value_display_type":       stringSchema(false, validation.StringInSlice(quicksight.PrimaryValueDisplayType_Values(), false)),
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

func expandGaugeChartVisual(tfList []interface{}) *quicksight.GaugeChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.GaugeChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandGaugeChartConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]interface{}); ok && len(v) > 0 {
		visual.ConditionalFormatting = expandGaugeChartConditionalFormatting(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		visual.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		visual.Title = expandVisualTitleLabelOptions(v)
	}

	return visual
}

func expandGaugeChartConfiguration(tfList []interface{}) *quicksight.GaugeChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.GaugeChartConfiguration{}

	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		config.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandGaugeChartFieldWells(v)
	}
	if v, ok := tfMap["gauge_chart_options"].([]interface{}); ok && len(v) > 0 {
		config.GaugeChartOptions = expandGaugeChartOptions(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.TooltipOptions = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}

	return config
}

func expandGaugeChartFieldWells(tfList []interface{}) *quicksight.GaugeChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.GaugeChartFieldWells{}

	if v, ok := tfMap["target_values"].([]interface{}); ok && len(v) > 0 {
		config.TargetValues = expandMeasureFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandGaugeChartOptions(tfList []interface{}) *quicksight.GaugeChartOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.GaugeChartOptions{}

	if v, ok := tfMap["primary_value_display_type"].(string); ok && v != "" {
		options.PrimaryValueDisplayType = aws.String(v)
	}
	if v, ok := tfMap["primary_value_font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.PrimaryValueFontConfiguration = expandFontConfiguration(v)
	}
	if v, ok := tfMap["arc"].([]interface{}); ok && len(v) > 0 {
		options.Arc = expandArcConfiguration(v)
	}
	if v, ok := tfMap["arc_axis"].([]interface{}); ok && len(v) > 0 {
		options.ArcAxis = expandArcAxisConfiguration(v)
	}
	if v, ok := tfMap["comparison"].([]interface{}); ok && len(v) > 0 {
		options.Comparison = expandComparisonConfiguration(v)
	}

	return options
}

func expandGaugeChartConditionalFormatting(tfList []interface{}) *quicksight.GaugeChartConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.GaugeChartConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]interface{}); ok && len(v) > 0 {
		config.ConditionalFormattingOptions = expandGaugeChartConditionalFormattingOptions(v)
	}

	return config
}

func expandGaugeChartConditionalFormattingOptions(tfList []interface{}) []*quicksight.GaugeChartConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.GaugeChartConditionalFormattingOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandGaugeChartConditionalFormattingOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandGaugeChartConditionalFormattingOption(tfMap map[string]interface{}) *quicksight.GaugeChartConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.GaugeChartConditionalFormattingOption{}

	if v, ok := tfMap["arc"].([]interface{}); ok && len(v) > 0 {
		options.Arc = expandGaugeChartArcConditionalFormatting(v)
	}
	if v, ok := tfMap["primary_value"].([]interface{}); ok && len(v) > 0 {
		options.PrimaryValue = expandGaugeChartPrimaryValueConditionalFormatting(v)
	}

	return options
}

func expandGaugeChartArcConditionalFormatting(tfList []interface{}) *quicksight.GaugeChartArcConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.GaugeChartArcConditionalFormatting{}

	if v, ok := tfMap["foreground_color"].([]interface{}); ok && len(v) > 0 {
		config.ForegroundColor = expandConditionalFormattingColor(v)
	}

	return config
}

func expandGaugeChartPrimaryValueConditionalFormatting(tfList []interface{}) *quicksight.GaugeChartPrimaryValueConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.GaugeChartPrimaryValueConditionalFormatting{}

	if v, ok := tfMap["icon"].([]interface{}); ok && len(v) > 0 {
		config.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		config.TextColor = expandConditionalFormattingColor(v)
	}

	return config
}

func expandArcConfiguration(tfList []interface{}) *quicksight.ArcConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ArcConfiguration{}

	if v, ok := tfMap["arc_angle"].(float64); ok {
		config.ArcAngle = aws.Float64(v)
	}
	if v, ok := tfMap["arc_thickness"].(string); ok && v != "" {
		config.ArcThickness = aws.String(v)
	}

	return config
}

func expandArcAxisConfiguration(tfList []interface{}) *quicksight.ArcAxisConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ArcAxisConfiguration{}

	if v, ok := tfMap["range"].([]interface{}); ok && len(v) > 0 {
		config.Range = expandArcAxisDisplayRange(v)
	}
	if v, ok := tfMap["reserve_range"].(int); ok {
		config.ReserveRange = aws.Int64(int64(v))
	}

	return config
}

func expandArcAxisDisplayRange(tfList []interface{}) *quicksight.ArcAxisDisplayRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ArcAxisDisplayRange{}

	if v, ok := tfMap[names.AttrMax].(float64); ok {
		config.Max = aws.Float64(v)
	}
	if v, ok := tfMap[names.AttrMin].(float64); ok {
		config.Min = aws.Float64(v)
	}

	return config
}

func flattenGaugeChartVisual(apiObject *quicksight.GaugeChartVisual) []interface{} {
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

	return []interface{}{tfMap}
}

func flattenGaugeChartConfiguration(apiObject *quicksight.GaugeChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenGaugeChartFieldWells(apiObject *quicksight.GaugeChartFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TargetValues != nil {
		tfMap["target_values"] = flattenMeasureFields(apiObject.TargetValues)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenGaugeChartOptions(apiObject *quicksight.GaugeChartOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Arc != nil {
		tfMap["arc"] = flattenArcConfiguration(apiObject.Arc)
	}
	if apiObject.ArcAxis != nil {
		tfMap["arc_axis"] = flattenArcAxisConfiguration(apiObject.ArcAxis)
	}
	if apiObject.Comparison != nil {
		tfMap["comparison"] = flattenComparisonConfiguration(apiObject.Comparison)
	}
	if apiObject.PrimaryValueDisplayType != nil {
		tfMap["primary_value_display_type"] = aws.StringValue(apiObject.PrimaryValueDisplayType)
	}
	if apiObject.PrimaryValueFontConfiguration != nil {
		tfMap["primary_value_font_configuration"] = flattenFontConfiguration(apiObject.PrimaryValueFontConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenArcConfiguration(apiObject *quicksight.ArcConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ArcAngle != nil {
		tfMap["arc_angle"] = aws.Float64Value(apiObject.ArcAngle)
	}
	if apiObject.ArcThickness != nil {
		tfMap["arc_thickness"] = aws.StringValue(apiObject.ArcThickness)
	}

	return []interface{}{tfMap}
}

func flattenArcAxisConfiguration(apiObject *quicksight.ArcAxisConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Range != nil {
		tfMap["range"] = flattenArcAxisDisplayRange(apiObject.Range)
	}
	if apiObject.ReserveRange != nil {
		tfMap["reserve_range"] = aws.Int64Value(apiObject.ReserveRange)
	}

	return []interface{}{tfMap}
}

func flattenArcAxisDisplayRange(apiObject *quicksight.ArcAxisDisplayRange) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Max != nil {
		tfMap[names.AttrMax] = aws.Float64Value(apiObject.Max)
	}
	if apiObject.Min != nil {
		tfMap[names.AttrMin] = aws.Float64Value(apiObject.Min)
	}

	return []interface{}{tfMap}
}

func flattenComparisonConfiguration(apiObject *quicksight.ComparisonConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComparisonFormat != nil {
		tfMap["comparison_format"] = flattenComparisonFormatConfiguration(apiObject.ComparisonFormat)
	}
	if apiObject.ComparisonMethod != nil {
		tfMap["comparison_method"] = aws.StringValue(apiObject.ComparisonMethod)
	}

	return []interface{}{tfMap}
}

func flattenComparisonFormatConfiguration(apiObject *quicksight.ComparisonFormatConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.NumberDisplayFormatConfiguration != nil {
		tfMap["number_display_format_configuration"] = flattenNumberDisplayFormatConfiguration(apiObject.NumberDisplayFormatConfiguration)
	}
	if apiObject.PercentageDisplayFormatConfiguration != nil {
		tfMap["percentage_display_format_configuration"] = flattenPercentageDisplayFormatConfiguration(apiObject.PercentageDisplayFormatConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenGaugeChartConditionalFormatting(apiObject *quicksight.GaugeChartConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenGaugeChartConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []interface{}{tfMap}
}

func flattenGaugeChartConditionalFormattingOption(apiObject []*quicksight.GaugeChartConditionalFormattingOption) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.Arc != nil {
			tfMap["arc"] = flattenGaugeChartArcConditionalFormatting(config.Arc)
		}
		if config.PrimaryValue != nil {
			tfMap["primary_value"] = flattenGaugeChartPrimaryValueConditionalFormatting(config.PrimaryValue)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenGaugeChartArcConditionalFormatting(apiObject *quicksight.GaugeChartArcConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ForegroundColor != nil {
		tfMap["foreground_color"] = flattenConditionalFormattingColor(apiObject.ForegroundColor)
	}

	return []interface{}{tfMap}
}

func flattenGaugeChartPrimaryValueConditionalFormatting(apiObject *quicksight.GaugeChartPrimaryValueConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Icon != nil {
		tfMap["icon"] = flattenConditionalFormattingIcon(apiObject.Icon)
	}
	if apiObject.TextColor != nil {
		tfMap["text_color"] = flattenConditionalFormattingColor(apiObject.TextColor)
	}

	return []interface{}{tfMap}
}
