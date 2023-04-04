package quicksight

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func gaugeChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GaugeChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
										"target_values": measureFieldSchema(measureFieldsMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
										"values":        measureFieldSchema(measureFieldsMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
																"max": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
																"min": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
															},
														},
													},
													"reserve_angle": {
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
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["reserve_angle"].(int64); ok {
		config.ReserveRange = aws.Int64(v)
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

	if v, ok := tfMap["max"].(float64); ok {
		config.Max = aws.Float64(v)
	}
	if v, ok := tfMap["min"].(float64); ok {
		config.Min = aws.Float64(v)
	}

	return config
}
