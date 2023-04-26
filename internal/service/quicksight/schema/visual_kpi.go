package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func kpiVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"target_values": measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
										"trend_groups":  dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
										"values":        measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
									},
								},
							},
							"kpi_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"comparison":                       comparisonConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComparisonConfiguration.html
										"primary_value_display_type":       stringSchema(false, validation.StringInSlice(quicksight.PrimaryValueDisplayType_Values(), false)),
										"primary_value_font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
										"progress_bar": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ProgressBarOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
												},
											},
										},
										"secondary_value": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SecondaryValueOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
												},
											},
										},
										"secondary_value_font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
										"trend_arrows": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TrendArrowOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
												},
											},
										},
									},
								},
							},
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPISortConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"trend_group_sort": fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
									},
								},
							},
						},
					},
				},
				"column_hierarchies": columnHierarchiesSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnHierarchy.html
				"conditional_formatting": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIConditionalFormatting.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"conditional_formatting_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIConditionalFormattingOption.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 100,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"primary_value": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIPrimaryValueConditionalFormatting.html
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
										"progress_bar": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIProgressBarConditionalFormatting.html
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

func expandKPIVisual(tfList []interface{}) *quicksight.KPIVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.KPIVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandKPIConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]interface{}); ok && len(v) > 0 {
		visual.ConditionalFormatting = expandKPIConditionalFormatting(v)
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

func expandKPIConfiguration(tfList []interface{}) *quicksight.KPIConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.KPIConfiguration{}

	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandKPIFieldWells(v)
	}
	if v, ok := tfMap["kpi_options"].([]interface{}); ok && len(v) > 0 {
		config.KPIOptions = expandKPIOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandKPISortConfiguration(v)
	}

	return config
}

func expandKPIFieldWells(tfList []interface{}) *quicksight.KPIFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.KPIFieldWells{}

	if v, ok := tfMap["trend_groups"].([]interface{}); ok && len(v) > 0 {
		config.TrendGroups = expandDimensionFields(v)
	}
	if v, ok := tfMap["target_values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}
	return config
}

func expandKPIOptions(tfList []interface{}) *quicksight.KPIOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.KPIOptions{}

	if v, ok := tfMap["primary_value_display_type"].(string); ok && v != "" {
		options.PrimaryValueDisplayType = aws.String(v)
	}
	if v, ok := tfMap["comparison"].([]interface{}); ok && len(v) > 0 {
		options.Comparison = expandComparisonConfiguration(v)
	}
	if v, ok := tfMap["primary_value_font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.PrimaryValueFontConfiguration = expandFontConfiguration(v)
	}
	if v, ok := tfMap["progress_bar"].([]interface{}); ok && len(v) > 0 {
		options.ProgressBar = expandProgressBarOptions(v)
	}
	if v, ok := tfMap["secondary_value"].([]interface{}); ok && len(v) > 0 {
		options.SecondaryValue = expandSecondaryValueOptions(v)
	}
	if v, ok := tfMap["secondary_value_font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.SecondaryValueFontConfiguration = expandFontConfiguration(v)
	}
	if v, ok := tfMap["trend_arrows"].([]interface{}); ok && len(v) > 0 {
		options.TrendArrows = expandTrendArrowOptions(v)
	}

	return options
}

func expandProgressBarOptions(tfList []interface{}) *quicksight.ProgressBarOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ProgressBarOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}

	return options
}

func expandSecondaryValueOptions(tfList []interface{}) *quicksight.SecondaryValueOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.SecondaryValueOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}

	return options
}

func expandTrendArrowOptions(tfList []interface{}) *quicksight.TrendArrowOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TrendArrowOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}

	return options
}

func expandKPISortConfiguration(tfList []interface{}) *quicksight.KPISortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.KPISortConfiguration{}

	if v, ok := tfMap["trend_group_sort"].([]interface{}); ok && len(v) > 0 {
		config.TrendGroupSort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandKPIConditionalFormatting(tfList []interface{}) *quicksight.KPIConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.KPIConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]interface{}); ok && len(v) > 0 {
		config.ConditionalFormattingOptions = expandKPIConditionalFormattingOptions(v)
	}

	return config
}

func expandKPIConditionalFormattingOptions(tfList []interface{}) []*quicksight.KPIConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.KPIConditionalFormattingOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandKPIConditionalFormattingOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandKPIConditionalFormattingOption(tfMap map[string]interface{}) *quicksight.KPIConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.KPIConditionalFormattingOption{}

	if v, ok := tfMap["primary_value"].([]interface{}); ok && len(v) > 0 {
		options.PrimaryValue = expandKPIPrimaryValueConditionalFormatting(v)
	}
	if v, ok := tfMap["progress_bar"].([]interface{}); ok && len(v) > 0 {
		options.ProgressBar = expandKPIProgressBarConditionalFormatting(v)
	}

	return options
}

func expandKPIPrimaryValueConditionalFormatting(tfList []interface{}) *quicksight.KPIPrimaryValueConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.KPIPrimaryValueConditionalFormatting{}

	if v, ok := tfMap["icon"].([]interface{}); ok && len(v) > 0 {
		options.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		options.TextColor = expandConditionalFormattingColor(v)
	}

	return options
}

func expandKPIProgressBarConditionalFormatting(tfList []interface{}) *quicksight.KPIProgressBarConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.KPIProgressBarConditionalFormatting{}

	if v, ok := tfMap["foreground_color"].([]interface{}); ok && len(v) > 0 {
		options.ForegroundColor = expandConditionalFormattingColor(v)
	}

	return options
}
