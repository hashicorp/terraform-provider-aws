// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func tableVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"field_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"order": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100,
											Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringLenBetween(1, 512)},
										},
										"selected_field_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldOption.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id":     stringLenBetweenSchema(attrRequired, 1, 512),
													"custom_label": stringLenBetweenSchema(attrOptional, 1, 2048),
													"url_styling": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldURLConfiguration.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"image_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldImageConfiguration.html
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"sizing_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellImageSizingConfiguration.html
																				Type:     schema.TypeList,
																				Optional: true,
																				MinItems: 1,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"table_cell_image_scaling_configuration": stringEnumSchema[awstypes.TableCellImageScalingConfiguration](attrOptional),
																					},
																				},
																			},
																		},
																	},
																},
																"link_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldLinkConfiguration.html
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			names.AttrContent: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldLinkContentConfiguration.html
																				Type:     schema.TypeList,
																				Optional: true,
																				MinItems: 1,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"custom_icon_content": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldCustomIconContent.html
																							Type:     schema.TypeList,
																							Optional: true,
																							MinItems: 1,
																							MaxItems: 1,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"icon": stringEnumSchema[awstypes.TableFieldIconSetType](attrOptional),
																								},
																							},
																						},
																						"custom_text_content": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldCustomTextContent.html
																							Type:     schema.TypeList,
																							Optional: true,
																							MinItems: 1,
																							MaxItems: 1,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
																									names.AttrValue: {
																										Type:     schema.TypeString,
																										Optional: true,
																									},
																								},
																							},
																						},
																					},
																				},
																			},
																			names.AttrTarget: stringEnumSchema[awstypes.URLTargetConfiguration](attrOptional),
																		},
																	},
																},
															},
														},
													},
													"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
													"width": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"table_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"group_by":       dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"table_unaggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableUnaggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrValues: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_UnaggregatedField.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 200,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
																"field_id":             stringLenBetweenSchema(attrRequired, 1, 512),
																"format_configuration": formatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FormatConfiguration.html
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"paginated_report_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TablePaginatedReportOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"overflow_column_header_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
										"vertical_overflow_visibility":      stringEnumSchema[awstypes.Visibility](attrOptional),
									},
								},
							},
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"pagination_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PaginationConfiguration.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"page_number": intAtLeastSchema(attrRequired, 1),
													"page_size": {
														Type:     schema.TypeInt,
														Required: true,
													},
												},
											},
										},
										"row_sort": fieldSortOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"table_inline_visualizations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableInlineVisualization.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 200,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"data_bars": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataBarsOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id":       stringLenBetweenSchema(attrRequired, 1, 512),
													"negative_color": hexColorSchema(attrOptional),
													"positive_color": hexColorSchema(attrOptional),
												},
											},
										},
									},
								},
							},
							"table_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"cell_style":                  tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"header_style":                tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"orientation":                 stringEnumSchema[awstypes.TableOrientation](attrOptional),
										"row_alternate_color_options": rowAlternateColorOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RowAlternateColorOptions.html
									},
								},
							},
							"total_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TotalOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"custom_label": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"placement":         stringEnumSchema[awstypes.TableTotalsPlacement](attrOptional),
										"scroll_status":     stringEnumSchema[awstypes.TableTotalsScrollStatus](attrOptional),
										"total_cell_style":  tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"totals_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
									},
								},
							},
						},
					},
				},
				"conditional_formatting": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableConditionalFormatting.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"conditional_formatting_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableConditionalFormattingOption.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 100,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"cell": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellConditionalFormatting.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id":    stringLenBetweenSchema(attrRequired, 1, 512),
													"text_format": textConditionalFormatSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextConditionalFormat.html
												},
											},
										},
										"row": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableRowConditionalFormatting.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"background_color": conditionalFormattingColorSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingColor.html
													"text_color":       conditionalFormattingColorSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingColor.html
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

func expandTableVisual(tfList []any) *awstypes.TableVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandTableConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]any); ok && len(v) > 0 {
		apiObject.ConditionalFormatting = expandTableConditionalFormatting(v)
	}
	if v, ok := tfMap["subtitle"].([]any); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandTableConfiguration(tfList []any) *awstypes.TableConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableConfiguration{}

	if v, ok := tfMap["field_options"].([]any); ok && len(v) > 0 {
		apiObject.FieldOptions = expandTableFieldOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandTableFieldWells(v)
	}
	if v, ok := tfMap["paginated_report_options"].([]any); ok && len(v) > 0 {
		apiObject.PaginatedReportOptions = expandTablePaginatedReportOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandTableSortConfiguration(v)
	}
	if v, ok := tfMap["table_inline_visualizations"].([]any); ok && len(v) > 0 {
		apiObject.TableInlineVisualizations = expandTableInlineVisualizations(v)
	}
	if v, ok := tfMap["table_options"].([]any); ok && len(v) > 0 {
		apiObject.TableOptions = expandTableOptions(v)
	}
	if v, ok := tfMap["total_options"].([]any); ok && len(v) > 0 {
		apiObject.TotalOptions = expandTableTotalOptions(v)
	}

	return apiObject
}

func expandTableFieldWells(tfList []any) *awstypes.TableFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableFieldWells{}

	if v, ok := tfMap["table_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.TableAggregatedFieldWells = expandTableAggregatedFieldWells(v)
	}
	if v, ok := tfMap["table_unaggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.TableUnaggregatedFieldWells = expandTableUnaggregatedFieldWells(v)
	}

	return apiObject
}

func expandTableAggregatedFieldWells(tfList []any) *awstypes.TableAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableAggregatedFieldWells{}

	if v, ok := tfMap["group_by"].([]any); ok && len(v) > 0 {
		apiObject.GroupBy = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandTableUnaggregatedFieldWells(tfList []any) *awstypes.TableUnaggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableUnaggregatedFieldWells{}

	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandUnaggregatedFields(v)
	}

	return apiObject
}

func expandUnaggregatedFields(tfList []any) []awstypes.UnaggregatedField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.UnaggregatedField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandUnaggregatedField(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandUnaggregatedField(tfMap map[string]any) *awstypes.UnaggregatedField {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.UnaggregatedField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandFormatConfiguration(v)
	}

	return apiObject
}

func expandTableSortConfiguration(tfList []any) *awstypes.TableSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableSortConfiguration{}

	if v, ok := tfMap["pagination_configuration"].([]any); ok && len(v) > 0 {
		apiObject.PaginationConfiguration = expandPaginationConfiguration(v)
	}
	if v, ok := tfMap["row_sort"].([]any); ok && len(v) > 0 {
		apiObject.RowSort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func expandTableFieldOptions(tfList []any) *awstypes.TableFieldOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableFieldOptions{}

	if v, ok := tfMap["order"].([]any); ok && len(v) > 0 {
		apiObject.Order = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["selected_field_options"].([]any); ok && len(v) > 0 {
		apiObject.SelectedFieldOptions = expandTableFieldOptionsList(v)
	}

	return apiObject
}

func expandTableFieldOptionsList(tfList []any) []awstypes.TableFieldOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.TableFieldOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandTableFieldOption(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandTableFieldOption(tfMap map[string]any) *awstypes.TableFieldOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TableFieldOption{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		apiObject.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["width"].(string); ok && v != "" {
		apiObject.Width = aws.String(v)
	}
	if v, ok := tfMap["url_styling"].([]any); ok && len(v) > 0 {
		apiObject.URLStyling = expandTableFieldURLConfiguration(v)
	}

	return apiObject
}

func expandTableFieldURLConfiguration(tfList []any) *awstypes.TableFieldURLConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableFieldURLConfiguration{}

	if v, ok := tfMap["image_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ImageConfiguration = expandTableFieldImageConfiguration(v)
	}
	if v, ok := tfMap["link_configuration"].([]any); ok && len(v) > 0 {
		apiObject.LinkConfiguration = expandTableFieldLinkConfiguration(v)
	}

	return apiObject
}

func expandTableFieldImageConfiguration(tfList []any) *awstypes.TableFieldImageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableFieldImageConfiguration{}

	if v, ok := tfMap["sizing_options"].([]any); ok && len(v) > 0 {
		apiObject.SizingOptions = expandTableCellImageSizingConfiguration(v)
	}

	return apiObject
}

func expandTableCellImageSizingConfiguration(tfList []any) *awstypes.TableCellImageSizingConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableCellImageSizingConfiguration{}

	if v, ok := tfMap["table_cell_image_scaling_configuration"].(string); ok && v != "" {
		apiObject.TableCellImageScalingConfiguration = awstypes.TableCellImageScalingConfiguration(v)
	}

	return apiObject
}

func expandTableFieldLinkConfiguration(tfList []any) *awstypes.TableFieldLinkConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableFieldLinkConfiguration{}

	if v, ok := tfMap[names.AttrTarget].(string); ok && v != "" {
		apiObject.Target = awstypes.URLTargetConfiguration(v)
	}
	if v, ok := tfMap[names.AttrContent].([]any); ok && len(v) > 0 {
		apiObject.Content = expandTableFieldLinkContentConfiguration(v)
	}

	return apiObject
}

func expandTableFieldLinkContentConfiguration(tfList []any) *awstypes.TableFieldLinkContentConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableFieldLinkContentConfiguration{}

	if v, ok := tfMap["custom_icon_content"].([]any); ok && len(v) > 0 {
		apiObject.CustomIconContent = expandTableFieldCustomIconContent(v)
	}
	if v, ok := tfMap["custom_text_content"].([]any); ok && len(v) > 0 {
		apiObject.CustomTextContent = expandTableFieldCustomTextContent(v)
	}

	return apiObject
}

func expandTableFieldCustomIconContent(tfList []any) *awstypes.TableFieldCustomIconContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableFieldCustomIconContent{}

	if v, ok := tfMap["icon"].(string); ok && v != "" {
		apiObject.Icon = awstypes.TableFieldIconSetType(v)
	}

	return apiObject
}

func expandTableFieldCustomTextContent(tfList []any) *awstypes.TableFieldCustomTextContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableFieldCustomTextContent{}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}
	if v, ok := tfMap["custom_text_content"].([]any); ok && len(v) > 0 {
		apiObject.FontConfiguration = expandFontConfiguration(v)
	}

	return apiObject
}

func expandTablePaginatedReportOptions(tfList []any) *awstypes.TablePaginatedReportOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TablePaginatedReportOptions{}

	if v, ok := tfMap["overflow_column_header_visibility"].(string); ok && v != "" {
		apiObject.OverflowColumnHeaderVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["vertical_overflow_visibility"].(string); ok && v != "" {
		apiObject.VerticalOverflowVisibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandTableOptions(tfList []any) *awstypes.TableOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableOptions{}

	if v, ok := tfMap["orientation"].(string); ok && v != "" {
		apiObject.Orientation = awstypes.TableOrientation(v)
	}
	if v, ok := tfMap["cell_style"].([]any); ok && len(v) > 0 {
		apiObject.CellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["header_style"].([]any); ok && len(v) > 0 {
		apiObject.HeaderStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["row_alternate_color_options"].([]any); ok && len(v) > 0 {
		apiObject.RowAlternateColorOptions = expandRowAlternateColorOptions(v)
	}

	return apiObject
}

func expandTableTotalOptions(tfList []any) *awstypes.TotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TotalOptions{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		apiObject.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["placement"].(string); ok && v != "" {
		apiObject.Placement = awstypes.TableTotalsPlacement(v)
	}
	if v, ok := tfMap["scroll_status"].(string); ok && v != "" {
		apiObject.ScrollStatus = awstypes.TableTotalsScrollStatus(v)
	}
	if v, ok := tfMap["totals_visibility"].(string); ok && v != "" {
		apiObject.TotalsVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["total_cell_style"].([]any); ok && len(v) > 0 {
		apiObject.TotalCellStyle = expandTableCellStyle(v)
	}

	return apiObject
}

func expandTableConditionalFormatting(tfList []any) *awstypes.TableConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]any); ok && len(v) > 0 {
		apiObject.ConditionalFormattingOptions = expandTableConditionalFormattingOptions(v)
	}

	return apiObject
}

func expandTableConditionalFormattingOptions(tfList []any) []awstypes.TableConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.TableConditionalFormattingOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandTableConditionalFormattingOption(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandTableConditionalFormattingOption(tfMap map[string]any) *awstypes.TableConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TableConditionalFormattingOption{}

	if v, ok := tfMap["cell"].([]any); ok && len(v) > 0 {
		apiObject.Cell = expandTableCellConditionalFormatting(v)
	}
	if v, ok := tfMap["row"].([]any); ok && len(v) > 0 {
		apiObject.Row = expandTableRowConditionalFormatting(v)
	}

	return apiObject
}

func expandTableCellConditionalFormatting(tfList []any) *awstypes.TableCellConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableCellConditionalFormatting{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["text_format"].([]any); ok && len(v) > 0 {
		apiObject.TextFormat = expandTextConditionalFormat(v)
	}

	return apiObject
}

func expandTableRowConditionalFormatting(tfList []any) *awstypes.TableRowConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableRowConditionalFormatting{}

	if v, ok := tfMap["background_color"].([]any); ok && len(v) > 0 {
		apiObject.BackgroundColor = expandConditionalFormattingColor(v)
	}
	if v, ok := tfMap["text_color"].([]any); ok && len(v) > 0 {
		apiObject.TextColor = expandConditionalFormattingColor(v)
	}

	return apiObject
}

func expandTableInlineVisualizations(tfList []any) []awstypes.TableInlineVisualization {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.TableInlineVisualization

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandTableInlineVisualization(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandTableInlineVisualization(tfMap map[string]any) *awstypes.TableInlineVisualization {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TableInlineVisualization{}

	if v, ok := tfMap["data_bars"].([]any); ok && len(v) > 0 {
		apiObject.DataBars = expandDataBarsOptions(v)
	}

	return apiObject
}

func expandDataBarsOptions(tfList []any) *awstypes.DataBarsOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataBarsOptions{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["negative_color"].(string); ok && v != "" {
		apiObject.NegativeColor = aws.String(v)
	}
	if v, ok := tfMap["positive_color"].(string); ok && v != "" {
		apiObject.PositiveColor = aws.String(v)
	}

	return apiObject
}

func flattenTableVisual(apiObject *awstypes.TableVisual) []any {
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
		tfMap["chart_configuration"] = flattenTableConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.ConditionalFormatting != nil {
		tfMap["conditional_formatting"] = flattenTableConditionalFormatting(apiObject.ConditionalFormatting)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []any{tfMap}
}

func flattenTableConfiguration(apiObject *awstypes.TableConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldOptions != nil {
		tfMap["field_options"] = flattenTableFieldOptions(apiObject.FieldOptions)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenTableFieldWells(apiObject.FieldWells)
	}
	if apiObject.PaginatedReportOptions != nil {
		tfMap["paginated_report_options"] = flattenTablePaginatedReportOptions(apiObject.PaginatedReportOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenTableSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.TableInlineVisualizations != nil {
		tfMap["table_inline_visualizations"] = flattenTableInlineVisualization(apiObject.TableInlineVisualizations)
	}
	if apiObject.TableOptions != nil {
		tfMap["table_options"] = flattenTableOptions(apiObject.TableOptions)
	}
	if apiObject.TotalOptions != nil {
		tfMap["total_options"] = flattenTotalOptions(apiObject.TotalOptions)
	}

	return []any{tfMap}
}

func flattenTableFieldOptions(apiObject *awstypes.TableFieldOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Order != nil {
		tfMap["order"] = apiObject.Order
	}
	if apiObject.SelectedFieldOptions != nil {
		tfMap["selected_field_options"] = flattenTableFieldOption(apiObject.SelectedFieldOptions)
	}

	return []any{tfMap}
}

func flattenTableFieldOption(apiObjects []awstypes.TableFieldOption) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.FieldId != nil {
			tfMap["field_id"] = aws.ToString(apiObject.FieldId)
		}
		if apiObject.CustomLabel != nil {
			tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
		}
		if apiObject.URLStyling != nil {
			tfMap["url_styling"] = flattenTableFieldURLConfiguration(apiObject.URLStyling)
		}
		tfMap["visibility"] = apiObject.Visibility
		if apiObject.Width != nil {
			tfMap["width"] = aws.ToString(apiObject.Width)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTableFieldURLConfiguration(apiObject *awstypes.TableFieldURLConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ImageConfiguration != nil {
		tfMap["image_configuration"] = flattenTableFieldImageConfiguration(apiObject.ImageConfiguration)
	}
	if apiObject.LinkConfiguration != nil {
		tfMap["link_configuration"] = flattenTableFieldLinkConfiguration(apiObject.LinkConfiguration)
	}

	return []any{tfMap}
}

func flattenTableFieldImageConfiguration(apiObject *awstypes.TableFieldImageConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.SizingOptions != nil {
		tfMap["sizing_options"] = flattenTableCellImageSizingConfiguration(apiObject.SizingOptions)
	}

	return []any{tfMap}
}

func flattenTableCellImageSizingConfiguration(apiObject *awstypes.TableCellImageSizingConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"table_cell_image_scaling_configuration": apiObject.TableCellImageScalingConfiguration,
	}

	return []any{tfMap}
}

func flattenTableFieldLinkConfiguration(apiObject *awstypes.TableFieldLinkConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Content != nil {
		tfMap[names.AttrContent] = flattenTableFieldLinkContentConfiguration(apiObject.Content)
	}
	tfMap[names.AttrTarget] = apiObject.Target

	return []any{tfMap}
}

func flattenTableFieldLinkContentConfiguration(apiObject *awstypes.TableFieldLinkContentConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CustomIconContent != nil {
		tfMap["custom_icon_content"] = flattenTableFieldCustomIconContent(apiObject.CustomIconContent)
	}
	if apiObject.CustomTextContent != nil {
		tfMap["custom_text_content"] = flattenTableFieldCustomTextContent(apiObject.CustomTextContent)
	}

	return []any{tfMap}
}

func flattenTableFieldCustomIconContent(apiObject *awstypes.TableFieldCustomIconContent) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"icon": apiObject.Icon,
	}

	return []any{tfMap}
}

func flattenTableFieldCustomTextContent(apiObject *awstypes.TableFieldCustomTextContent) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = aws.ToString(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenTableFieldWells(apiObject *awstypes.TableFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.TableAggregatedFieldWells != nil {
		tfMap["table_aggregated_field_wells"] = flattenTableAggregatedFieldWells(apiObject.TableAggregatedFieldWells)
	}
	if apiObject.TableUnaggregatedFieldWells != nil {
		tfMap["table_unaggregated_field_wells"] = flattenTableUnaggregatedFieldWells(apiObject.TableUnaggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenTableAggregatedFieldWells(apiObject *awstypes.TableAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.GroupBy != nil {
		tfMap["group_by"] = flattenDimensionFields(apiObject.GroupBy)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenTableUnaggregatedFieldWells(apiObject *awstypes.TableUnaggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenUnaggregatedField(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenUnaggregatedField(apiObjects []awstypes.UnaggregatedField) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Column != nil {
			tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
		}
		if apiObject.FieldId != nil {
			tfMap["field_id"] = aws.ToString(apiObject.FieldId)
		}
		if apiObject.FormatConfiguration != nil {
			tfMap["format_configuration"] = flattenFormatConfiguration(apiObject.FormatConfiguration)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTablePaginatedReportOptions(apiObject *awstypes.TablePaginatedReportOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"overflow_column_header_visibility": apiObject.OverflowColumnHeaderVisibility,
		"vertical_overflow_visibility":      apiObject.VerticalOverflowVisibility,
	}

	return []any{tfMap}
}

func flattenTableSortConfiguration(apiObject *awstypes.TableSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.PaginationConfiguration != nil {
		tfMap["pagination_configuration"] = flattenPaginationConfiguration(apiObject.PaginationConfiguration)
	}
	if apiObject.RowSort != nil {
		tfMap["row_sort"] = flattenFieldSortOptions(apiObject.RowSort)
	}

	return []any{tfMap}
}

func flattenTableInlineVisualization(apiObjects []awstypes.TableInlineVisualization) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.DataBars != nil {
			tfMap["data_bars"] = flattenDataBarsOptions(apiObject.DataBars)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataBarsOptions(apiObject *awstypes.DataBarsOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.NegativeColor != nil {
		tfMap["negative_color"] = aws.ToString(apiObject.NegativeColor)
	}
	if apiObject.PositiveColor != nil {
		tfMap["positive_color"] = aws.ToString(apiObject.PositiveColor)
	}

	return []any{tfMap}
}

func flattenTableOptions(apiObject *awstypes.TableOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CellStyle != nil {
		tfMap["cell_style"] = flattenTableCellStyle(apiObject.CellStyle)
	}
	if apiObject.HeaderStyle != nil {
		tfMap["header_style"] = flattenTableCellStyle(apiObject.HeaderStyle)
	}
	tfMap["orientation"] = apiObject.Orientation
	if apiObject.RowAlternateColorOptions != nil {
		tfMap["row_alternate_color_options"] = flattenRowAlternateColorOptions(apiObject.RowAlternateColorOptions)
	}

	return []any{tfMap}
}

func flattenTotalOptions(apiObject *awstypes.TotalOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}
	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
	}
	tfMap["placement"] = apiObject.Placement
	tfMap["scroll_status"] = apiObject.ScrollStatus
	if apiObject.TotalCellStyle != nil {
		tfMap["total_cell_style"] = flattenTableCellStyle(apiObject.TotalCellStyle)
	}
	tfMap["totals_visibility"] = apiObject.TotalsVisibility

	return []any{tfMap}
}

func flattenTableConditionalFormatting(apiObject *awstypes.TableConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenTableConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []any{tfMap}
}

func flattenTableConditionalFormattingOption(apiObjects []awstypes.TableConditionalFormattingOption) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Cell != nil {
			tfMap["cell"] = flattenTableCellConditionalFormatting(apiObject.Cell)
		}
		if apiObject.Row != nil {
			tfMap["row"] = flattenTableRowConditionalFormatting(apiObject.Row)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTableCellConditionalFormatting(apiObject *awstypes.TableCellConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.TextFormat != nil {
		tfMap["text_format"] = flattenTextConditionalFormat(apiObject.TextFormat)
	}

	return []any{tfMap}
}

func flattenTableRowConditionalFormatting(apiObject *awstypes.TableRowConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = flattenConditionalFormattingColor(apiObject.BackgroundColor)
	}
	if apiObject.TextColor != nil {
		tfMap["text_color"] = flattenConditionalFormattingColor(apiObject.TextColor)
	}

	return []any{tfMap}
}
