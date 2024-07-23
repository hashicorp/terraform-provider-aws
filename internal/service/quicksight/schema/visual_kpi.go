// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func kpiVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
										"target_values":  measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
										"trend_groups":   dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
										names.AttrValues: measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
										"sparkline": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPISparklineOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"color":              stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
													"tooltip_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
													names.AttrType:       stringSchema(true, validation.StringInSlice(quicksight.KPISparklineType_Values(), false)),
													"visibility":         stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
												},
											},
										},
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
										"visual_layout_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIVisualLayoutOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"standard_layout": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIVisualStandardLayout.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrType: stringSchema(true, validation.StringInSlice(quicksight.KPIVisualStandardLayoutType_Values(), false)),
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPISortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
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
										"actual_value": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIActualValueConditionalFormatting.html
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
										"comparison_value": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_KPIComparisonValueConditionalFormatting.html
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
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
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
		config.TargetValues = expandMeasureFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["sparkline"].([]interface{}); ok && len(v) > 0 {
		options.Sparkline = expandKPISparklineOptions(v)
	}
	if v, ok := tfMap["trend_arrows"].([]interface{}); ok && len(v) > 0 {
		options.TrendArrows = expandTrendArrowOptions(v)
	}
	if v, ok := tfMap["visual_layout_options"].([]interface{}); ok && len(v) > 0 {
		options.VisualLayoutOptions = expandKPIVisualLayoutOptions(v)
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

func expandKPISparklineOptions(tfList []interface{}) *quicksight.KPISparklineOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.KPISparklineOptions{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		options.Color = aws.String(v)
	}
	if v, ok := tfMap["tooltip_visibility"].(string); ok && v != "" {
		options.TooltipVisibility = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		options.Type = aws.String(v)
	}
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

func expandKPIVisualLayoutOptions(tfList []interface{}) *quicksight.KPIVisualLayoutOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.KPIVisualLayoutOptions{}

	if v, ok := tfMap["standard_layout"].([]interface{}); ok && len(v) > 0 {
		options.StandardLayout = expandKPIVisualStandardLayout(v)
	}

	return options
}

func expandKPIVisualStandardLayout(tfList []interface{}) *quicksight.KPIVisualStandardLayout {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.KPIVisualStandardLayout{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		options.Type = aws.String(v)
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

	if v, ok := tfMap["actual_value"].([]interface{}); ok && len(v) > 0 {
		options.ActualValue = expandKPIActualValueConditionalFormatting(v)
	}
	if v, ok := tfMap["comparison_value"].([]interface{}); ok && len(v) > 0 {
		options.ComparisonValue = expandKPIComparisonValueConditionalFormatting(v)
	}
	if v, ok := tfMap["primary_value"].([]interface{}); ok && len(v) > 0 {
		options.PrimaryValue = expandKPIPrimaryValueConditionalFormatting(v)
	}
	if v, ok := tfMap["progress_bar"].([]interface{}); ok && len(v) > 0 {
		options.ProgressBar = expandKPIProgressBarConditionalFormatting(v)
	}

	return options
}

func expandKPIActualValueConditionalFormatting(tfList []interface{}) *quicksight.KPIActualValueConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.KPIActualValueConditionalFormatting{}

	if v, ok := tfMap["icon"].([]interface{}); ok && len(v) > 0 {
		options.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		options.TextColor = expandConditionalFormattingColor(v)
	}

	return options
}

func expandKPIComparisonValueConditionalFormatting(tfList []interface{}) *quicksight.KPIComparisonValueConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.KPIComparisonValueConditionalFormatting{}

	if v, ok := tfMap["icon"].([]interface{}); ok && len(v) > 0 {
		options.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		options.TextColor = expandConditionalFormattingColor(v)
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

func flattenKPIVisual(apiObject *quicksight.KPIVisual) []interface{} {
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
		tfMap["chart_configuration"] = flattenKPIConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.ConditionalFormatting != nil {
		tfMap["conditional_formatting"] = flattenKPIConditionalFormatting(apiObject.ConditionalFormatting)
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

func flattenKPIConfiguration(apiObject *quicksight.KPIConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenKPIFieldWells(apiObject.FieldWells)
	}
	if apiObject.KPIOptions != nil {
		tfMap["kpi_options"] = flattenKPIOptions(apiObject.KPIOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenKPISortConfiguration(apiObject.SortConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenKPIFieldWells(apiObject *quicksight.KPIFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TargetValues != nil {
		tfMap["target_values"] = flattenMeasureFields(apiObject.TargetValues)
	}
	if apiObject.TrendGroups != nil {
		tfMap["trend_groups"] = flattenDimensionFields(apiObject.TrendGroups)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenKPIOptions(apiObject *quicksight.KPIOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Comparison != nil {
		tfMap["comparison"] = flattenComparisonConfiguration(apiObject.Comparison)
	}
	if apiObject.PrimaryValueDisplayType != nil {
		tfMap["primary_value_display_type"] = aws.StringValue(apiObject.PrimaryValueDisplayType)
	}
	if apiObject.PrimaryValueFontConfiguration != nil {
		tfMap["primary_value_font_configuration"] = flattenFontConfiguration(apiObject.PrimaryValueFontConfiguration)
	}
	if apiObject.ProgressBar != nil {
		tfMap["progress_bar"] = flattenProgressBarOptions(apiObject.ProgressBar)
	}
	if apiObject.SecondaryValue != nil {
		tfMap["secondary_value"] = flattenSecondaryValueOptions(apiObject.SecondaryValue)
	}
	if apiObject.SecondaryValueFontConfiguration != nil {
		tfMap["secondary_value_font_configuration"] = flattenFontConfiguration(apiObject.SecondaryValueFontConfiguration)
	}
	if apiObject.Sparkline != nil {
		tfMap["sparkline"] = flattenKPISparklineOptions(apiObject.Sparkline)
	}
	if apiObject.TrendArrows != nil {
		tfMap["trend_arrows"] = flattenTrendArrowOptions(apiObject.TrendArrows)
	}
	if apiObject.VisualLayoutOptions != nil {
		tfMap["visual_layout_options"] = flattenKPIVisualLayoutOptions(apiObject.VisualLayoutOptions)
	}

	return []interface{}{tfMap}
}

func flattenProgressBarOptions(apiObject *quicksight.ProgressBarOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenSecondaryValueOptions(apiObject *quicksight.SecondaryValueOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenKPISparklineOptions(apiObject *quicksight.KPISparklineOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.StringValue(apiObject.Color)
	}
	if apiObject.TooltipVisibility != nil {
		tfMap["tooltip_visibility"] = aws.StringValue(apiObject.TooltipVisibility)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenTrendArrowOptions(apiObject *quicksight.TrendArrowOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenKPIVisualLayoutOptions(apiObject *quicksight.KPIVisualLayoutOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.StandardLayout != nil {
		tfMap["standard_layout"] = flattenKPIVisualStandardLayout(apiObject.StandardLayout)
	}

	return []interface{}{tfMap}
}

func flattenKPIVisualStandardLayout(apiObject *quicksight.KPIVisualStandardLayout) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}

	return []interface{}{tfMap}
}

func flattenKPISortConfiguration(apiObject *quicksight.KPISortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TrendGroupSort != nil {
		tfMap["trend_group_sort"] = flattenFieldSortOptions(apiObject.TrendGroupSort)
	}

	return []interface{}{tfMap}
}

func flattenKPIConditionalFormatting(apiObject *quicksight.KPIConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenKPIConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []interface{}{tfMap}
}

func flattenKPIConditionalFormattingOption(apiObject []*quicksight.KPIConditionalFormattingOption) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.ActualValue != nil {
			tfMap["actual_value"] = flattenKPIActualValueConditionalFormatting(config.ActualValue)
		}
		if config.ComparisonValue != nil {
			tfMap["comparison_value"] = flattenKPIComparisonValueConditionalFormatting(config.ComparisonValue)
		}
		if config.PrimaryValue != nil {
			tfMap["primary_value"] = flattenKPIPrimaryValueConditionalFormatting(config.PrimaryValue)
		}
		if config.ProgressBar != nil {
			tfMap["progress_bar"] = flattenKPIProgressBarConditionalFormatting(config.ProgressBar)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenKPIActualValueConditionalFormatting(apiObject *quicksight.KPIActualValueConditionalFormatting) []interface{} {
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

func flattenKPIComparisonValueConditionalFormatting(apiObject *quicksight.KPIComparisonValueConditionalFormatting) []interface{} {
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

func flattenKPIPrimaryValueConditionalFormatting(apiObject *quicksight.KPIPrimaryValueConditionalFormatting) []interface{} {
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

func flattenKPIProgressBarConditionalFormatting(apiObject *quicksight.KPIProgressBarConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ForegroundColor != nil {
		tfMap["foreground_color"] = flattenConditionalFormattingColor(apiObject.ForegroundColor)
	}

	return []interface{}{tfMap}
}
