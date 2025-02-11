// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
										"primary_value_display_type":       stringEnumSchema[awstypes.PrimaryValueDisplayType](attrOptional),
										"primary_value_font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
										"progress_bar": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ProgressBarOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
													"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
													"color":              hexColorSchema(attrOptional),
													"tooltip_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
													names.AttrType:       stringEnumSchema[awstypes.KPISparklineType](attrRequired),
													"visibility":         stringEnumSchema[awstypes.Visibility](attrOptional),
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
													"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
																names.AttrType: stringEnumSchema[awstypes.KPIVisualStandardLayoutType](attrRequired),
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
										"trend_group_sort": fieldSortOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
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

func expandKPIVisual(tfList []interface{}) *awstypes.KPIVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandKPIConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]interface{}); ok && len(v) > 0 {
		apiObject.ConditionalFormatting = expandKPIConditionalFormatting(v)
	}
	if v, ok := tfMap["column_hierarchies"].([]interface{}); ok && len(v) > 0 {
		apiObject.ColumnHierarchies = expandColumnHierarchies(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandKPIConfiguration(tfList []interface{}) *awstypes.KPIConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIConfiguration{}

	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		apiObject.FieldWells = expandKPIFieldWells(v)
	}
	if v, ok := tfMap["kpi_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.KPIOptions = expandKPIOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandKPISortConfiguration(v)
	}

	return apiObject
}

func expandKPIFieldWells(tfList []interface{}) *awstypes.KPIFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIFieldWells{}

	if v, ok := tfMap["trend_groups"].([]interface{}); ok && len(v) > 0 {
		apiObject.TrendGroups = expandDimensionFields(v)
	}
	if v, ok := tfMap["target_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.TargetValues = expandMeasureFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}
	return apiObject
}

func expandKPIOptions(tfList []interface{}) *awstypes.KPIOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIOptions{}

	if v, ok := tfMap["primary_value_display_type"].(string); ok && v != "" {
		apiObject.PrimaryValueDisplayType = awstypes.PrimaryValueDisplayType(v)
	}
	if v, ok := tfMap["comparison"].([]interface{}); ok && len(v) > 0 {
		apiObject.Comparison = expandComparisonConfiguration(v)
	}
	if v, ok := tfMap["primary_value_font_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.PrimaryValueFontConfiguration = expandFontConfiguration(v)
	}
	if v, ok := tfMap["progress_bar"].([]interface{}); ok && len(v) > 0 {
		apiObject.ProgressBar = expandProgressBarOptions(v)
	}
	if v, ok := tfMap["secondary_value"].([]interface{}); ok && len(v) > 0 {
		apiObject.SecondaryValue = expandSecondaryValueOptions(v)
	}
	if v, ok := tfMap["secondary_value_font_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.SecondaryValueFontConfiguration = expandFontConfiguration(v)
	}
	if v, ok := tfMap["sparkline"].([]interface{}); ok && len(v) > 0 {
		apiObject.Sparkline = expandKPISparklineOptions(v)
	}
	if v, ok := tfMap["trend_arrows"].([]interface{}); ok && len(v) > 0 {
		apiObject.TrendArrows = expandTrendArrowOptions(v)
	}
	if v, ok := tfMap["visual_layout_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.VisualLayoutOptions = expandKPIVisualLayoutOptions(v)
	}

	return apiObject
}

func expandProgressBarOptions(tfList []interface{}) *awstypes.ProgressBarOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ProgressBarOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandSecondaryValueOptions(tfList []interface{}) *awstypes.SecondaryValueOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SecondaryValueOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandKPISparklineOptions(tfList []interface{}) *awstypes.KPISparklineOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPISparklineOptions{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap["tooltip_visibility"].(string); ok && v != "" {
		apiObject.TooltipVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.KPISparklineType(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandTrendArrowOptions(tfList []interface{}) *awstypes.TrendArrowOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TrendArrowOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandKPIVisualLayoutOptions(tfList []interface{}) *awstypes.KPIVisualLayoutOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIVisualLayoutOptions{}

	if v, ok := tfMap["standard_layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.StandardLayout = expandKPIVisualStandardLayout(v)
	}

	return apiObject
}

func expandKPIVisualStandardLayout(tfList []interface{}) *awstypes.KPIVisualStandardLayout {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIVisualStandardLayout{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.KPIVisualStandardLayoutType(v)
	}

	return apiObject
}

func expandKPISortConfiguration(tfList []interface{}) *awstypes.KPISortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPISortConfiguration{}

	if v, ok := tfMap["trend_group_sort"].([]interface{}); ok && len(v) > 0 {
		apiObject.TrendGroupSort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func expandKPIConditionalFormatting(tfList []interface{}) *awstypes.KPIConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.ConditionalFormattingOptions = expandKPIConditionalFormattingOptions(v)
	}

	return apiObject
}

func expandKPIConditionalFormattingOptions(tfList []interface{}) []awstypes.KPIConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.KPIConditionalFormattingOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandKPIConditionalFormattingOption(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandKPIConditionalFormattingOption(tfMap map[string]interface{}) *awstypes.KPIConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.KPIConditionalFormattingOption{}

	if v, ok := tfMap["actual_value"].([]interface{}); ok && len(v) > 0 {
		apiObject.ActualValue = expandKPIActualValueConditionalFormatting(v)
	}
	if v, ok := tfMap["comparison_value"].([]interface{}); ok && len(v) > 0 {
		apiObject.ComparisonValue = expandKPIComparisonValueConditionalFormatting(v)
	}
	if v, ok := tfMap["primary_value"].([]interface{}); ok && len(v) > 0 {
		apiObject.PrimaryValue = expandKPIPrimaryValueConditionalFormatting(v)
	}
	if v, ok := tfMap["progress_bar"].([]interface{}); ok && len(v) > 0 {
		apiObject.ProgressBar = expandKPIProgressBarConditionalFormatting(v)
	}

	return apiObject
}

func expandKPIActualValueConditionalFormatting(tfList []interface{}) *awstypes.KPIActualValueConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIActualValueConditionalFormatting{}

	if v, ok := tfMap["icon"].([]interface{}); ok && len(v) > 0 {
		apiObject.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		apiObject.TextColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func expandKPIComparisonValueConditionalFormatting(tfList []interface{}) *awstypes.KPIComparisonValueConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIComparisonValueConditionalFormatting{}

	if v, ok := tfMap["icon"].([]interface{}); ok && len(v) > 0 {
		apiObject.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		apiObject.TextColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func expandKPIPrimaryValueConditionalFormatting(tfList []interface{}) *awstypes.KPIPrimaryValueConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIPrimaryValueConditionalFormatting{}

	if v, ok := tfMap["icon"].([]interface{}); ok && len(v) > 0 {
		apiObject.Icon = expandConditionalFormattingIcon(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		apiObject.TextColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func expandKPIProgressBarConditionalFormatting(tfList []interface{}) *awstypes.KPIProgressBarConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.KPIProgressBarConditionalFormatting{}

	if v, ok := tfMap["foreground_color"].([]interface{}); ok && len(v) > 0 {
		apiObject.ForegroundColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func flattenKPIVisual(apiObject *awstypes.KPIVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id": aws.ToString(apiObject.VisualId),
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

func flattenKPIConfiguration(apiObject *awstypes.KPIConfiguration) []interface{} {
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

func flattenKPIFieldWells(apiObject *awstypes.KPIFieldWells) []interface{} {
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

func flattenKPIOptions(apiObject *awstypes.KPIOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Comparison != nil {
		tfMap["comparison"] = flattenComparisonConfiguration(apiObject.Comparison)
	}
	tfMap["primary_value_display_type"] = apiObject.PrimaryValueDisplayType
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

func flattenProgressBarOptions(apiObject *awstypes.ProgressBarOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visibility": apiObject.Visibility,
	}

	return []interface{}{tfMap}
}

func flattenSecondaryValueOptions(apiObject *awstypes.SecondaryValueOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visibility": apiObject.Visibility,
	}

	return []interface{}{tfMap}
}

func flattenKPISparklineOptions(apiObject *awstypes.KPISparklineOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	tfMap["tooltip_visibility"] = apiObject.TooltipVisibility
	tfMap[names.AttrType] = apiObject.Type
	tfMap["visibility"] = apiObject.Visibility

	return []interface{}{tfMap}
}

func flattenTrendArrowOptions(apiObject *awstypes.TrendArrowOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visibility": apiObject.Visibility,
	}

	return []interface{}{tfMap}
}

func flattenKPIVisualLayoutOptions(apiObject *awstypes.KPIVisualLayoutOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.StandardLayout != nil {
		tfMap["standard_layout"] = flattenKPIVisualStandardLayout(apiObject.StandardLayout)
	}

	return []interface{}{tfMap}
}

func flattenKPIVisualStandardLayout(apiObject *awstypes.KPIVisualStandardLayout) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrType: apiObject.Type,
	}

	return []interface{}{tfMap}
}

func flattenKPISortConfiguration(apiObject *awstypes.KPISortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.TrendGroupSort != nil {
		tfMap["trend_group_sort"] = flattenFieldSortOptions(apiObject.TrendGroupSort)
	}

	return []interface{}{tfMap}
}

func flattenKPIConditionalFormatting(apiObject *awstypes.KPIConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenKPIConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []interface{}{tfMap}
}

func flattenKPIConditionalFormattingOption(apiObjects []awstypes.KPIConditionalFormattingOption) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.ActualValue != nil {
			tfMap["actual_value"] = flattenKPIActualValueConditionalFormatting(apiObject.ActualValue)
		}
		if apiObject.ComparisonValue != nil {
			tfMap["comparison_value"] = flattenKPIComparisonValueConditionalFormatting(apiObject.ComparisonValue)
		}
		if apiObject.PrimaryValue != nil {
			tfMap["primary_value"] = flattenKPIPrimaryValueConditionalFormatting(apiObject.PrimaryValue)
		}
		if apiObject.ProgressBar != nil {
			tfMap["progress_bar"] = flattenKPIProgressBarConditionalFormatting(apiObject.ProgressBar)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenKPIActualValueConditionalFormatting(apiObject *awstypes.KPIActualValueConditionalFormatting) []interface{} {
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

func flattenKPIComparisonValueConditionalFormatting(apiObject *awstypes.KPIComparisonValueConditionalFormatting) []interface{} {
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

func flattenKPIPrimaryValueConditionalFormatting(apiObject *awstypes.KPIPrimaryValueConditionalFormatting) []interface{} {
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

func flattenKPIProgressBarConditionalFormatting(apiObject *awstypes.KPIProgressBarConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ForegroundColor != nil {
		tfMap["foreground_color"] = flattenConditionalFormattingColor(apiObject.ForegroundColor)
	}

	return []interface{}{tfMap}
}
