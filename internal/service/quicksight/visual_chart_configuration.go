package quicksight

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func axisDisplayOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"axis_line_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
				"axis_offset": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"data_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDataOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"date_axis_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateAxisOptions.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"missing_date_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
									},
								},
							},
							"numeric_axis_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericAxisOptions.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"range": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayRange.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"data_driven": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayDataDrivenRange.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{}, // For future extensions
														},
													},
													"min_max": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayMinMaxRange.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"maximum": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
																"minimum": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
										"scale": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisScale.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"linear": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisLinearScale.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"step_count": {
																	Type:     schema.TypeInt,
																	Optional: true,
																},
																"step_size": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
															},
														},
													},
													"logarithmic": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisLogarithmicScale.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"base": {
																	Type:     schema.TypeFloat,
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"grid_line_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
				"scrollbar_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ScrollBarOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
							"visible_range": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisibleRangeOptions.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"percent_range": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PercentVisibleRange.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"from": floatSchema(false, validation.FloatBetween(0, 100)),
													"to":   floatSchema(false, validation.FloatBetween(0, 100)),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"tick_label_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisTickLabelOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"label_options": labelOptionsSchema(),
							"rotation_angle": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func chartAxisLabelOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"axis_label_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisLabelOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"apply_to": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisLabelReferenceOptions.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column":   columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
										"field_id": stringSchema(true, validation.StringLenBetween(1, 512)),
									},
								},
							},
							"custom_label": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
						},
					},
				},
				"sort_icon_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
				"visibility":           stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
			},
		},
	}
}

func itemsLimitConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"items_limit": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"other_categories": stringSchema(true, validation.StringInSlice(quicksight.OtherCategories_Values(), false)),
			},
		},
	}
}

func contributionAnalysisDefaultsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ContributionAnalysisDefault.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 200,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"contributor_dimensions": {
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 4,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"column_name":         stringSchema(true, validation.StringLenBetween(1, 128)),
							"data_set_identifier": stringSchema(true, validation.StringLenBetween(1, 2048)),
						},
					},
				},
				"measure_field_id": stringSchema(true, validation.StringLenBetween(1, 512)),
			},
		},
	}
}

func referenceLineSchema(maxItems int) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineDataConfiguration.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"axis_binding": stringSchema(false, validation.StringInSlice(quicksight.AxisBinding_Values(), false)),
							"dynamic_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineDynamicDataConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"calculation":                  numericalAggregationFunctionSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericalAggregationFunction.html
										"column":                       columnSchema(),                           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
										"measure_aggregation_function": aggregationFunctionSchema(true),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
									},
								},
							},
							"static_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineStaticDataConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"value": {
											Type:     schema.TypeFloat,
											Required: true,
										},
									},
								},
							},
						},
					},
				},
				"label_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineLabelConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"custom_label_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineCustomLabelConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"custom_label": stringSchema(true, validation.StringMatch(regexp.MustCompile(`.*\S.*`), "")),
									},
								},
							},
							"font_color":          stringSchema(false, validation.StringMatch(regexp.MustCompile(`^#[A-F0-9]{6}$`), "")),
							"font_configuration":  fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
							"horizontal_position": stringSchema(false, validation.StringInSlice(quicksight.ReferenceLineLabelHorizontalPosition_Values(), false)),
							"value_label_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineValueLabelConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"format_configuration": numericFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericFormatConfiguration.html
										"relative_position":    stringSchema(false, validation.StringInSlice(quicksight.ReferenceLineValueLabelRelativePosition_Values(), false)),
									},
								},
							},
							"vertical_position": stringSchema(false, validation.StringInSlice(quicksight.ReferenceLineLabelVerticalPosition_Values(), false)),
						},
					},
				},
				"status": stringSchema(false, validation.StringInSlice(quicksight.Status_Values(), false)),
				"style_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLineStyleConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":   stringSchema(false, validation.StringMatch(regexp.MustCompile(`^#[A-F0-9]{6}$`), "")),
							"pattern": stringSchema(false, validation.StringInSlice(quicksight.ReferenceLinePatternType_Values(), false)),
						},
					},
				},
			},
		},
	}
}

func smallMultiplesOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SmallMultiplesOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"max_visible_columns": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntBetween(1, 10),
				},
				"max_visible_rows": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntBetween(1, 10),
				},
				"panel_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PanelConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"background_color":      stringSchema(false, validation.StringMatch(regexp.MustCompile(`^#[A-F0-9]{6}(?:[A-F0-9]{2})?$`), "")),
							"background_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
							"border_color":          stringSchema(false, validation.StringMatch(regexp.MustCompile(`^#[A-F0-9]{6}(?:[A-F0-9]{2})?$`), "")),
							"border_style":          stringSchema(false, validation.StringInSlice(quicksight.PanelBorderStyle_Values(), false)),
							"border_thickness": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"border_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
							"gutter_spacing": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"gutter_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
							"title": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PanelTitleOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"font_configuration":        fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
										"horizontal_text_alignment": stringSchema(false, validation.StringInSlice(quicksight.HorizontalTextAlignment_Values(), false)),
										"visibility":                stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func expandAxisDisplayOptions(tfList []interface{}) *quicksight.AxisDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisDisplayOptions{}

	if v, ok := tfMap["axis_line_visibility"].(string); ok && v != "" {
		options.AxisLineVisibility = aws.String(v)
	}
	if v, ok := tfMap["axis_offset"].(string); ok && v != "" {
		options.AxisOffset = aws.String(v)
	}
	if v, ok := tfMap["grid_line_visibility"].(string); ok && v != "" {
		options.GridLineVisibility = aws.String(v)
	}
	if v, ok := tfMap["data_options"].([]interface{}); ok && len(v) > 0 {
		options.DataOptions = expandAxisDataOptions(v)
	}
	if v, ok := tfMap["scrollbar_options"].([]interface{}); ok && len(v) > 0 {
		options.ScrollbarOptions = expandScrollBarOptions(v)
	}
	if v, ok := tfMap["tick_label_options"].([]interface{}); ok && len(v) > 0 {
		options.TickLabelOptions = expandAxisTickLabelOptions(v)
	}

	return options
}

func expandAxisDataOptions(tfList []interface{}) *quicksight.AxisDataOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisDataOptions{}

	if v, ok := tfMap["date_axis_options"].([]interface{}); ok && len(v) > 0 {
		options.DateAxisOptions = expandDateAxisOptions(v)
	}
	if v, ok := tfMap["numeric_axis_options"].([]interface{}); ok && len(v) > 0 {
		options.NumericAxisOptions = expandNumericAxisOptions(v)
	}

	return options
}

func expandDateAxisOptions(tfList []interface{}) *quicksight.DateAxisOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.DateAxisOptions{}

	if v, ok := tfMap["missing_date_visibility"].(string); ok && v != "" {
		options.MissingDateVisibility = aws.String(v)
	}

	return options
}

func expandNumericAxisOptions(tfList []interface{}) *quicksight.NumericAxisOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.NumericAxisOptions{}

	if v, ok := tfMap["range"].([]interface{}); ok && len(v) > 0 {
		options.Range = expandAxisDisplayRange(v)
	}
	if v, ok := tfMap["scale"].([]interface{}); ok && len(v) > 0 {
		options.Scale = expandAxisScale(v)
	}

	return options
}

func expandAxisDisplayRange(tfList []interface{}) *quicksight.AxisDisplayRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisDisplayRange{}

	if v, ok := tfMap["data_driven"].([]interface{}); ok && len(v) > 0 {
		options.DataDriven = expandAxisDisplayDataDrivenRange(v)
	}
	if v, ok := tfMap["min_max"].([]interface{}); ok && len(v) > 0 {
		options.MinMax = expandAxisDisplayMinMaxRange(v)
	}

	return options
}

func expandAxisDisplayDataDrivenRange(tfList []interface{}) *quicksight.AxisDisplayDataDrivenRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	//tfMap, ok := tfList[0].(map[string]interface{})
	//if !ok {
	//	return nil
	//}

	options := &quicksight.AxisDisplayDataDrivenRange{}

	return options
}

func expandAxisDisplayMinMaxRange(tfList []interface{}) *quicksight.AxisDisplayMinMaxRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisDisplayMinMaxRange{}

	if v, ok := tfMap["maximum"].(float64); ok {
		options.Maximum = aws.Float64(v)
	}
	if v, ok := tfMap["minimum"].(float64); ok {
		options.Minimum = aws.Float64(v)
	}

	return options
}

func expandAxisScale(tfList []interface{}) *quicksight.AxisScale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisScale{}

	if v, ok := tfMap["linear"].([]interface{}); ok && len(v) > 0 {
		options.Linear = expandAxisLinearScale(v)
	}
	if v, ok := tfMap["logarithmic"].([]interface{}); ok && len(v) > 0 {
		options.Logarithmic = expandAxisLogarithmicScale(v)
	}

	return options
}

func expandAxisLinearScale(tfList []interface{}) *quicksight.AxisLinearScale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisLinearScale{}

	if v, ok := tfMap["step_count"].(int64); ok {
		options.StepCount = aws.Int64(v)
	}
	if v, ok := tfMap["step_size"].(float64); ok {
		options.StepSize = aws.Float64(v)
	}

	return options
}

func expandAxisLogarithmicScale(tfList []interface{}) *quicksight.AxisLogarithmicScale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisLogarithmicScale{}

	if v, ok := tfMap["base"].(float64); ok {
		options.Base = aws.Float64(v)
	}

	return options
}

func expandScrollBarOptions(tfList []interface{}) *quicksight.ScrollBarOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ScrollBarOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}
	if v, ok := tfMap["visible_range"].([]interface{}); ok && len(v) > 0 {
		options.VisibleRange = expandVisibleRangeOptions(v)
	}

	return options
}

func expandVisibleRangeOptions(tfList []interface{}) *quicksight.VisibleRangeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.VisibleRangeOptions{}

	if v, ok := tfMap["percent_range"].([]interface{}); ok && len(v) > 0 {
		options.PercentRange = expandPercentVisibleRange(v)
	}

	return options
}

func expandPercentVisibleRange(tfList []interface{}) *quicksight.PercentVisibleRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PercentVisibleRange{}

	if v, ok := tfMap["from"].(float64); ok {
		options.From = aws.Float64(v)
	}
	if v, ok := tfMap["to"].(float64); ok {
		options.To = aws.Float64(v)
	}

	return options
}

func expandAxisTickLabelOptions(tfList []interface{}) *quicksight.AxisTickLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisTickLabelOptions{}

	if v, ok := tfMap["rotation_angle"].(float64); ok {
		options.RotationAngle = aws.Float64(v)
	}
	if v, ok := tfMap["label_options"].([]interface{}); ok && len(v) > 0 {
		options.LabelOptions = expandLabelOptions(v)
	}

	return options
}

func expandChartAxisLabelOptions(tfList []interface{}) *quicksight.ChartAxisLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ChartAxisLabelOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}
	if v, ok := tfMap["sort_icon_visibility"].(string); ok && v != "" {
		options.SortIconVisibility = aws.String(v)
	}
	if v, ok := tfMap["axis_label_options"].([]interface{}); ok && len(v) > 0 {
		options.AxisLabelOptions = expandAxisLabelOptionsList(v)
	}

	return options
}

func expandAxisLabelOptionsList(tfList []interface{}) []*quicksight.AxisLabelOptions {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.AxisLabelOptions
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandAxisLabelOptions(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandAxisLabelOptions(tfMap map[string]interface{}) *quicksight.AxisLabelOptions {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.AxisLabelOptions{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		options.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["apply_to"].([]interface{}); ok && len(v) > 0 {
		options.ApplyTo = expandAxisLabelReferenceOptions(v)
	}
	if v, ok := tfMap["font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.FontConfiguration = expandFontConfiguration(v)
	}

	return options
}

func expandAxisLabelReferenceOptions(tfList []interface{}) *quicksight.AxisLabelReferenceOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.AxisLabelReferenceOptions{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		options.Column = expandColumnIdentifier(v)
	}

	return options
}

func expandContributionAnalysisDefaults(tfList []interface{}) []*quicksight.ContributionAnalysisDefault {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.ContributionAnalysisDefault
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandContributionAnalysisDefault(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandContributionAnalysisDefault(tfMap map[string]interface{}) *quicksight.ContributionAnalysisDefault {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.ContributionAnalysisDefault{}

	if v, ok := tfMap["measure_field_id"].(string); ok && v != "" {
		options.MeasureFieldId = aws.String(v)
	}
	if v, ok := tfMap["contributor_dimensions"].([]interface{}); ok && len(v) > 0 {
		options.ContributorDimensions = expandColumnIdentifiers(v)
	}

	return options
}

func expandReferenceLines(tfList []interface{}) []*quicksight.ReferenceLine {
	if len(tfList) == 0 {
		return nil
	}

	var lines []*quicksight.ReferenceLine
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		line := expandReferenceLine(tfMap)
		if line == nil {
			continue
		}

		lines = append(lines, line)
	}

	return lines
}

func expandReferenceLine(tfMap map[string]interface{}) *quicksight.ReferenceLine {
	if tfMap == nil {
		return nil
	}

	line := &quicksight.ReferenceLine{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		line.Status = aws.String(v)
	}
	if v, ok := tfMap["data_configuration"].([]interface{}); ok && len(v) > 0 {
		line.DataConfiguration = expandReferenceLineDataConfiguration(v)
	}
	if v, ok := tfMap["label_configuration"].([]interface{}); ok && len(v) > 0 {
		line.LabelConfiguration = expandReferenceLineLabelConfiguration(v)
	}
	if v, ok := tfMap["style_configuration"].([]interface{}); ok && len(v) > 0 {
		line.StyleConfiguration = expandReferenceLineStyleConfiguration(v)
	}

	return line
}

func expandReferenceLineDataConfiguration(tfList []interface{}) *quicksight.ReferenceLineDataConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ReferenceLineDataConfiguration{}

	if v, ok := tfMap["axis_binding"].(string); ok && v != "" {
		config.AxisBinding = aws.String(v)
	}
	if v, ok := tfMap["dynamic_configuration"].([]interface{}); ok && len(v) > 0 {
		config.DynamicConfiguration = expandReferenceLineDynamicDataConfiguration(v)
	}
	if v, ok := tfMap["static_configuration"].([]interface{}); ok && len(v) > 0 {
		config.StaticConfiguration = expandReferenceLineStaticDataConfiguration(v)
	}

	return config
}

func expandReferenceLineDynamicDataConfiguration(tfList []interface{}) *quicksight.ReferenceLineDynamicDataConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ReferenceLineDynamicDataConfiguration{}

	if v, ok := tfMap["calculation"].([]interface{}); ok && len(v) > 0 {
		config.Calculation = expandNumericalAggregationFunction(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		config.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["measure_aggregation_function"].([]interface{}); ok && len(v) > 0 {
		config.MeasureAggregationFunction = expandAggregationFunction(v)
	}

	return config
}

func expandReferenceLineStaticDataConfiguration(tfList []interface{}) *quicksight.ReferenceLineStaticDataConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ReferenceLineStaticDataConfiguration{}

	if v, ok := tfMap["value"].(float64); ok {
		config.Value = aws.Float64(v)
	}

	return config
}

func expandReferenceLineLabelConfiguration(tfList []interface{}) *quicksight.ReferenceLineLabelConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ReferenceLineLabelConfiguration{}

	if v, ok := tfMap["font_color"].(string); ok && v != "" {
		config.FontColor = aws.String(v)
	}
	if v, ok := tfMap["horizontal_position"].(string); ok && v != "" {
		config.HorizontalPosition = aws.String(v)
	}
	if v, ok := tfMap["vertical_position"].(string); ok && v != "" {
		config.VerticalPosition = aws.String(v)
	}
	if v, ok := tfMap["custom_label_configuration"].([]interface{}); ok && len(v) > 0 {
		config.CustomLabelConfiguration = expandReferenceLineCustomLabelConfiguration(v)
	}
	if v, ok := tfMap["font_configuration"].([]interface{}); ok && len(v) > 0 {
		config.FontConfiguration = expandFontConfiguration(v)
	}
	if v, ok := tfMap["value_label_configuration"].([]interface{}); ok && len(v) > 0 {
		config.ValueLabelConfiguration = expandReferenceLineValueLabelConfiguration(v)
	}

	return config
}

func expandReferenceLineCustomLabelConfiguration(tfList []interface{}) *quicksight.ReferenceLineCustomLabelConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ReferenceLineCustomLabelConfiguration{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		config.CustomLabel = aws.String(v)
	}

	return config
}

func expandReferenceLineValueLabelConfiguration(tfList []interface{}) *quicksight.ReferenceLineValueLabelConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ReferenceLineValueLabelConfiguration{}

	if v, ok := tfMap["relative_position"].(string); ok && v != "" {
		config.RelativePosition = aws.String(v)
	}
	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.FormatConfiguration = expandNumericFormatConfiguration(v)
	}

	return config
}

func expandReferenceLineStyleConfiguration(tfList []interface{}) *quicksight.ReferenceLineStyleConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ReferenceLineStyleConfiguration{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		config.Color = aws.String(v)
	}
	if v, ok := tfMap["pattern"].(string); ok && v != "" {
		config.Pattern = aws.String(v)
	}

	return config
}

func expandSmallMultiplesOptions(tfList []interface{}) *quicksight.SmallMultiplesOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.SmallMultiplesOptions{}

	if v, ok := tfMap["max_visible_columns"].(int64); ok {
		options.MaxVisibleColumns = aws.Int64(v)
	}
	if v, ok := tfMap["max_visible_rows"].(int64); ok {
		options.MaxVisibleRows = aws.Int64(v)
	}
	if v, ok := tfMap["panel_configuration"].([]interface{}); ok && len(v) > 0 {
		options.PanelConfiguration = expandPanelConfiguration(v)
	}

	return options
}

func expandPanelConfiguration(tfList []interface{}) *quicksight.PanelConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PanelConfiguration{}

	if v, ok := tfMap["background_color"].(string); ok && v != "" {
		config.BackgroundColor = aws.String(v)
	}
	if v, ok := tfMap["background_visibility"].(string); ok && v != "" {
		config.BackgroundVisibility = aws.String(v)
	}
	if v, ok := tfMap["border_color"].(string); ok && v != "" {
		config.BorderColor = aws.String(v)
	}
	if v, ok := tfMap["border_style"].(string); ok && v != "" {
		config.BorderStyle = aws.String(v)
	}
	if v, ok := tfMap["border_thickness"].(string); ok && v != "" {
		config.BorderThickness = aws.String(v)
	}
	if v, ok := tfMap["border_visibility"].(string); ok && v != "" {
		config.BorderVisibility = aws.String(v)
	}
	if v, ok := tfMap["gutter_spacing"].(string); ok && v != "" {
		config.GutterSpacing = aws.String(v)
	}
	if v, ok := tfMap["gutter_visibility"].(string); ok && v != "" {
		config.GutterVisibility = aws.String(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		config.Title = expandPanelTitleOptions(v)
	}

	return config
}

func expandPanelTitleOptions(tfList []interface{}) *quicksight.PanelTitleOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PanelTitleOptions{}

	if v, ok := tfMap["horizontal_text_alignment"].(string); ok && v != "" {
		options.HorizontalTextAlignment = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}
	if v, ok := tfMap["font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.FontConfiguration = expandFontConfiguration(v)
	}

	return options
}

func expandItemsLimitConfiguration(tfList []interface{}) *quicksight.ItemsLimitConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ItemsLimitConfiguration{}

	if v, ok := tfMap["items_limit"].(int64); ok {
		config.ItemsLimit = aws.Int64(v)
	}
	if v, ok := tfMap["other_categories"].(string); ok && v != "" {
		config.OtherCategories = aws.String(v)
	}

	return config
}
