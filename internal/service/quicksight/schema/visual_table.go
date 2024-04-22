// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func tableVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
													"field_id":     stringSchema(true, validation.StringLenBetween(1, 512)),
													"custom_label": stringSchema(false, validation.StringLenBetween(1, 2048)),
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
																						"table_cell_image_scaling_configuration": stringSchema(false, validation.StringInSlice(quicksight.TableCellImageScalingConfiguration_Values(), false)),
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
																			"content": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldLinkContentConfiguration.html
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
																									"icon": stringSchema(false, validation.StringInSlice(quicksight.TableFieldIconSetType_Values(), false)),
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
																									"value": {
																										Type:     schema.TypeString,
																										Optional: true,
																									},
																								},
																							},
																						},
																					},
																				},
																			},
																			"target": stringSchema(false, validation.StringInSlice(quicksight.URLTargetConfiguration_Values(), false)),
																		},
																	},
																},
															},
														},
													},
													"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
													"group_by": dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"values":   measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"values": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_UnaggregatedField.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 200,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
																"field_id":             stringSchema(true, validation.StringLenBetween(1, 512)),
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
										"overflow_column_header_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
										"vertical_overflow_visibility":      stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
													"page_number": intSchema(true, validation.IntAtLeast(1)),
													"page_size": {
														Type:     schema.TypeInt,
														Required: true,
													},
												},
											},
										},
										"row_sort": fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
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
													"field_id":       stringSchema(true, validation.StringLenBetween(1, 512)),
													"negative_color": stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
													"positive_color": stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
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
										"orientation":                 stringSchema(false, validation.StringInSlice(quicksight.TableOrientation_Values(), false)),
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
										"placement":         stringSchema(false, validation.StringInSlice(quicksight.TableTotalsPlacement_Values(), false)),
										"scroll_status":     stringSchema(false, validation.StringInSlice(quicksight.TableTotalsScrollStatus_Values(), false)),
										"total_cell_style":  tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"totals_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
													"field_id":    stringSchema(true, validation.StringLenBetween(1, 512)),
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

func expandTableVisual(tfList []interface{}) *quicksight.TableVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.TableVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandTableConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]interface{}); ok && len(v) > 0 {
		visual.ConditionalFormatting = expandTableConditionalFormatting(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		visual.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		visual.Title = expandVisualTitleLabelOptions(v)
	}

	return visual
}

func expandTableConfiguration(tfList []interface{}) *quicksight.TableConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.TableConfiguration{}

	if v, ok := tfMap["field_options"].([]interface{}); ok && len(v) > 0 {
		config.FieldOptions = expandTableFieldOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandTableFieldWells(v)
	}
	if v, ok := tfMap["paginated_report_options"].([]interface{}); ok && len(v) > 0 {
		config.PaginatedReportOptions = expandTablePaginatedReportOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandTableSortConfiguration(v)
	}
	if v, ok := tfMap["table_inline_visualizations"].([]interface{}); ok && len(v) > 0 {
		config.TableInlineVisualizations = expandTableInlineVisualizations(v)
	}
	if v, ok := tfMap["table_options"].([]interface{}); ok && len(v) > 0 {
		config.TableOptions = expandTableOptions(v)
	}
	if v, ok := tfMap["total_options"].([]interface{}); ok && len(v) > 0 {
		config.TotalOptions = expandTableTotalOptions(v)
	}

	return config
}

func expandTableFieldWells(tfList []interface{}) *quicksight.TableFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.TableFieldWells{}

	if v, ok := tfMap["table_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.TableAggregatedFieldWells = expandTableAggregatedFieldWells(v)
	}
	if v, ok := tfMap["table_unaggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.TableUnaggregatedFieldWells = expandTableUnaggregatedFieldWells(v)
	}

	return config
}

func expandTableAggregatedFieldWells(tfList []interface{}) *quicksight.TableAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.TableAggregatedFieldWells{}

	if v, ok := tfMap["group_by"].([]interface{}); ok && len(v) > 0 {
		config.GroupBy = expandDimensionFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandTableUnaggregatedFieldWells(tfList []interface{}) *quicksight.TableUnaggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.TableUnaggregatedFieldWells{}

	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandUnaggregatedFields(v)
	}

	return config
}

func expandUnaggregatedFields(tfList []interface{}) []*quicksight.UnaggregatedField {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.UnaggregatedField
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandUnaggregatedField(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandUnaggregatedField(tfMap map[string]interface{}) *quicksight.UnaggregatedField {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.UnaggregatedField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		options.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		options.FormatConfiguration = expandFormatConfiguration(v)
	}

	return options
}

func expandTableSortConfiguration(tfList []interface{}) *quicksight.TableSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.TableSortConfiguration{}

	if v, ok := tfMap["pagination_configuration"].([]interface{}); ok && len(v) > 0 {
		config.PaginationConfiguration = expandPaginationConfiguration(v)
	}
	if v, ok := tfMap["row_sort"].([]interface{}); ok && len(v) > 0 {
		config.RowSort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandTableFieldOptions(tfList []interface{}) *quicksight.TableFieldOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableFieldOptions{}

	if v, ok := tfMap["order"].([]interface{}); ok && len(v) > 0 {
		options.Order = flex.ExpandStringList(v)
	}
	if v, ok := tfMap["selected_field_options"].([]interface{}); ok && len(v) > 0 {
		options.SelectedFieldOptions = expandTableFieldOptionsList(v)
	}

	return options
}

func expandTableFieldOptionsList(tfList []interface{}) []*quicksight.TableFieldOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.TableFieldOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandTableFieldOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandTableFieldOption(tfMap map[string]interface{}) *quicksight.TableFieldOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.TableFieldOption{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		options.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}
	if v, ok := tfMap["width"].(string); ok && v != "" {
		options.Width = aws.String(v)
	}
	if v, ok := tfMap["url_styling"].([]interface{}); ok && len(v) > 0 {
		options.URLStyling = expandTableFieldURLConfiguration(v)
	}

	return options
}

func expandTableFieldURLConfiguration(tfList []interface{}) *quicksight.TableFieldURLConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableFieldURLConfiguration{}

	if v, ok := tfMap["image_configuration"].([]interface{}); ok && len(v) > 0 {
		options.ImageConfiguration = expandTableFieldImageConfiguration(v)
	}
	if v, ok := tfMap["link_configuration"].([]interface{}); ok && len(v) > 0 {
		options.LinkConfiguration = expandTableFieldLinkConfiguration(v)
	}

	return options
}

func expandTableFieldImageConfiguration(tfList []interface{}) *quicksight.TableFieldImageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableFieldImageConfiguration{}

	if v, ok := tfMap["sizing_options"].([]interface{}); ok && len(v) > 0 {
		options.SizingOptions = expandTableCellImageSizingConfiguration(v)
	}

	return options
}

func expandTableCellImageSizingConfiguration(tfList []interface{}) *quicksight.TableCellImageSizingConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableCellImageSizingConfiguration{}

	if v, ok := tfMap["table_cell_image_scaling_configuration"].(string); ok && v != "" {
		options.TableCellImageScalingConfiguration = aws.String(v)
	}

	return options
}

func expandTableFieldLinkConfiguration(tfList []interface{}) *quicksight.TableFieldLinkConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableFieldLinkConfiguration{}

	if v, ok := tfMap["target"].(string); ok && v != "" {
		options.Target = aws.String(v)
	}
	if v, ok := tfMap["content"].([]interface{}); ok && len(v) > 0 {
		options.Content = expandTableFieldLinkContentConfiguration(v)
	}

	return options
}

func expandTableFieldLinkContentConfiguration(tfList []interface{}) *quicksight.TableFieldLinkContentConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableFieldLinkContentConfiguration{}

	if v, ok := tfMap["custom_icon_content"].([]interface{}); ok && len(v) > 0 {
		options.CustomIconContent = expandTableFieldCustomIconContent(v)
	}
	if v, ok := tfMap["custom_text_content"].([]interface{}); ok && len(v) > 0 {
		options.CustomTextContent = expandTableFieldCustomTextContent(v)
	}

	return options
}

func expandTableFieldCustomIconContent(tfList []interface{}) *quicksight.TableFieldCustomIconContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableFieldCustomIconContent{}

	if v, ok := tfMap["icon"].(string); ok && v != "" {
		options.Icon = aws.String(v)
	}

	return options
}

func expandTableFieldCustomTextContent(tfList []interface{}) *quicksight.TableFieldCustomTextContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableFieldCustomTextContent{}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		options.Value = aws.String(v)
	}
	if v, ok := tfMap["custom_text_content"].([]interface{}); ok && len(v) > 0 {
		options.FontConfiguration = expandFontConfiguration(v)
	}

	return options
}

func expandTablePaginatedReportOptions(tfList []interface{}) *quicksight.TablePaginatedReportOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TablePaginatedReportOptions{}

	if v, ok := tfMap["overflow_column_header_visibility"].(string); ok && v != "" {
		options.OverflowColumnHeaderVisibility = aws.String(v)
	}
	if v, ok := tfMap["vertical_overflow_visibility"].(string); ok && v != "" {
		options.VerticalOverflowVisibility = aws.String(v)
	}

	return options
}

func expandTableOptions(tfList []interface{}) *quicksight.TableOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableOptions{}

	if v, ok := tfMap["orientation"].(string); ok && v != "" {
		options.Orientation = aws.String(v)
	}
	if v, ok := tfMap["cell_style"].([]interface{}); ok && len(v) > 0 {
		options.CellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["header_style"].([]interface{}); ok && len(v) > 0 {
		options.HeaderStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["row_alternate_color_options"].([]interface{}); ok && len(v) > 0 {
		options.RowAlternateColorOptions = expandRowAlternateColorOptions(v)
	}

	return options
}

func expandTableTotalOptions(tfList []interface{}) *quicksight.TotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TotalOptions{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		options.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["placement"].(string); ok && v != "" {
		options.Placement = aws.String(v)
	}
	if v, ok := tfMap["scroll_status"].(string); ok && v != "" {
		options.ScrollStatus = aws.String(v)
	}
	if v, ok := tfMap["totals_visibility"].(string); ok && v != "" {
		options.TotalsVisibility = aws.String(v)
	}
	if v, ok := tfMap["total_cell_style"].([]interface{}); ok && len(v) > 0 {
		options.TotalCellStyle = expandTableCellStyle(v)
	}

	return options
}

func expandTableConditionalFormatting(tfList []interface{}) *quicksight.TableConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]interface{}); ok && len(v) > 0 {
		options.ConditionalFormattingOptions = expandTableConditionalFormattingOptions(v)
	}

	return options
}

func expandTableConditionalFormattingOptions(tfList []interface{}) []*quicksight.TableConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.TableConditionalFormattingOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandTableConditionalFormattingOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandTableConditionalFormattingOption(tfMap map[string]interface{}) *quicksight.TableConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.TableConditionalFormattingOption{}

	if v, ok := tfMap["cell"].([]interface{}); ok && len(v) > 0 {
		options.Cell = expandTableCellConditionalFormatting(v)
	}
	if v, ok := tfMap["row"].([]interface{}); ok && len(v) > 0 {
		options.Row = expandTableRowConditionalFormatting(v)
	}

	return options
}

func expandTableCellConditionalFormatting(tfList []interface{}) *quicksight.TableCellConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableCellConditionalFormatting{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["text_format"].([]interface{}); ok && len(v) > 0 {
		options.TextFormat = expandTextConditionalFormat(v)
	}

	return options
}

func expandTableRowConditionalFormatting(tfList []interface{}) *quicksight.TableRowConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableRowConditionalFormatting{}

	if v, ok := tfMap["background_color"].([]interface{}); ok && len(v) > 0 {
		options.BackgroundColor = expandConditionalFormattingColor(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		options.TextColor = expandConditionalFormattingColor(v)
	}

	return options
}

func expandTableInlineVisualizations(tfList []interface{}) []*quicksight.TableInlineVisualization {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.TableInlineVisualization
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandTableInlineVisualization(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandTableInlineVisualization(tfMap map[string]interface{}) *quicksight.TableInlineVisualization {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.TableInlineVisualization{}

	if v, ok := tfMap["data_bars"].([]interface{}); ok && len(v) > 0 {
		options.DataBars = expandDataBarsOptions(v)
	}

	return options
}

func expandDataBarsOptions(tfList []interface{}) *quicksight.DataBarsOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.DataBarsOptions{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["negative_color"].(string); ok && v != "" {
		options.NegativeColor = aws.String(v)
	}
	if v, ok := tfMap["positive_color"].(string); ok && v != "" {
		options.PositiveColor = aws.String(v)
	}
	return options
}

func flattenTableVisual(apiObject *quicksight.TableVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id": aws.StringValue(apiObject.VisualId),
	}
	if apiObject.Actions != nil {
		tfMap["actions"] = flattenVisualCustomAction(apiObject.Actions)
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

	return []interface{}{tfMap}
}

func flattenTableConfiguration(apiObject *quicksight.TableConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenTableFieldOptions(apiObject *quicksight.TableFieldOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Order != nil {
		tfMap["order"] = flex.FlattenStringList(apiObject.Order)
	}
	if apiObject.SelectedFieldOptions != nil {
		tfMap["selected_field_options"] = flattenTableFieldOption(apiObject.SelectedFieldOptions)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldOption(apiObject []*quicksight.TableFieldOption) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.FieldId != nil {
			tfMap["field_id"] = aws.StringValue(config.FieldId)
		}
		if config.CustomLabel != nil {
			tfMap["custom_label"] = aws.StringValue(config.CustomLabel)
		}
		if config.URLStyling != nil {
			tfMap["url_styling"] = flattenTableFieldURLConfiguration(config.URLStyling)
		}
		if config.Visibility != nil {
			tfMap["visibility"] = aws.StringValue(config.Visibility)
		}
		if config.Width != nil {
			tfMap["width"] = aws.StringValue(config.Width)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTableFieldURLConfiguration(apiObject *quicksight.TableFieldURLConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ImageConfiguration != nil {
		tfMap["image_configuration"] = flattenTableFieldImageConfiguration(apiObject.ImageConfiguration)
	}
	if apiObject.LinkConfiguration != nil {
		tfMap["link_configuration"] = flattenTableFieldLinkConfiguration(apiObject.LinkConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldImageConfiguration(apiObject *quicksight.TableFieldImageConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SizingOptions != nil {
		tfMap["sizing_options"] = flattenTableCellImageSizingConfiguration(apiObject.SizingOptions)
	}

	return []interface{}{tfMap}
}

func flattenTableCellImageSizingConfiguration(apiObject *quicksight.TableCellImageSizingConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TableCellImageScalingConfiguration != nil {
		tfMap["table_cell_image_scaling_configuration"] = aws.StringValue(apiObject.TableCellImageScalingConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldLinkConfiguration(apiObject *quicksight.TableFieldLinkConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Content != nil {
		tfMap["content"] = flattenTableFieldLinkContentConfiguration(apiObject.Content)
	}
	if apiObject.Target != nil {
		tfMap["target"] = aws.StringValue(apiObject.Target)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldLinkContentConfiguration(apiObject *quicksight.TableFieldLinkContentConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomIconContent != nil {
		tfMap["custom_icon_content"] = flattenTableFieldCustomIconContent(apiObject.CustomIconContent)
	}
	if apiObject.CustomTextContent != nil {
		tfMap["custom_text_content"] = flattenTableFieldCustomTextContent(apiObject.CustomTextContent)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldCustomIconContent(apiObject *quicksight.TableFieldCustomIconContent) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Icon != nil {
		tfMap["icon"] = aws.StringValue(apiObject.Icon)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldCustomTextContent(apiObject *quicksight.TableFieldCustomTextContent) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	if apiObject.Value != nil {
		tfMap["value"] = aws.StringValue(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldWells(apiObject *quicksight.TableFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TableAggregatedFieldWells != nil {
		tfMap["table_aggregated_field_wells"] = flattenTableAggregatedFieldWells(apiObject.TableAggregatedFieldWells)
	}
	if apiObject.TableUnaggregatedFieldWells != nil {
		tfMap["table_unaggregated_field_wells"] = flattenTableUnaggregatedFieldWells(apiObject.TableUnaggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenTableAggregatedFieldWells(apiObject *quicksight.TableAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.GroupBy != nil {
		tfMap["group_by"] = flattenDimensionFields(apiObject.GroupBy)
	}
	if apiObject.Values != nil {
		tfMap["values"] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenTableUnaggregatedFieldWells(apiObject *quicksight.TableUnaggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Values != nil {
		tfMap["values"] = flattenUnaggregatedField(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenUnaggregatedField(apiObject []*quicksight.UnaggregatedField) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.Column != nil {
			tfMap["column"] = flattenColumnIdentifier(config.Column)
		}
		if config.FieldId != nil {
			tfMap["field_id"] = aws.StringValue(config.FieldId)
		}
		if config.FormatConfiguration != nil {
			tfMap["format_configuration"] = flattenFormatConfiguration(config.FormatConfiguration)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTablePaginatedReportOptions(apiObject *quicksight.TablePaginatedReportOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.OverflowColumnHeaderVisibility != nil {
		tfMap["overflow_column_header_visibility"] = aws.StringValue(apiObject.OverflowColumnHeaderVisibility)
	}
	if apiObject.VerticalOverflowVisibility != nil {
		tfMap["vertical_overflow_visibility"] = aws.StringValue(apiObject.VerticalOverflowVisibility)
	}

	return []interface{}{tfMap}
}

func flattenTableSortConfiguration(apiObject *quicksight.TableSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PaginationConfiguration != nil {
		tfMap["pagination_configuration"] = flattenPaginationConfiguration(apiObject.PaginationConfiguration)
	}
	if apiObject.RowSort != nil {
		tfMap["row_sort"] = flattenFieldSortOptions(apiObject.RowSort)
	}

	return []interface{}{tfMap}
}

func flattenTableInlineVisualization(apiObject []*quicksight.TableInlineVisualization) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.DataBars != nil {
			tfMap["data_bars"] = flattenDataBarsOptions(config.DataBars)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataBarsOptions(apiObject *quicksight.DataBarsOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.NegativeColor != nil {
		tfMap["negative_color"] = aws.StringValue(apiObject.NegativeColor)
	}
	if apiObject.PositiveColor != nil {
		tfMap["positive_color"] = aws.StringValue(apiObject.PositiveColor)
	}

	return []interface{}{tfMap}
}

func flattenTableOptions(apiObject *quicksight.TableOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CellStyle != nil {
		tfMap["cell_style"] = flattenTableCellStyle(apiObject.CellStyle)
	}
	if apiObject.HeaderStyle != nil {
		tfMap["header_style"] = flattenTableCellStyle(apiObject.HeaderStyle)
	}
	if apiObject.Orientation != nil {
		tfMap["orientation"] = aws.StringValue(apiObject.Orientation)
	}
	if apiObject.RowAlternateColorOptions != nil {
		tfMap["row_alternate_color_options"] = flattenRowAlternateColorOptions(apiObject.RowAlternateColorOptions)
	}

	return []interface{}{tfMap}
}

func flattenTotalOptions(apiObject *quicksight.TotalOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.StringValue(apiObject.CustomLabel)
	}
	if apiObject.Placement != nil {
		tfMap["placement"] = aws.StringValue(apiObject.Placement)
	}
	if apiObject.ScrollStatus != nil {
		tfMap["scroll_status"] = aws.StringValue(apiObject.ScrollStatus)
	}
	if apiObject.TotalCellStyle != nil {
		tfMap["total_cell_style"] = flattenTableCellStyle(apiObject.TotalCellStyle)
	}
	if apiObject.TotalsVisibility != nil {
		tfMap["totals_visibility"] = aws.StringValue(apiObject.TotalsVisibility)
	}

	return []interface{}{tfMap}
}

func flattenTableConditionalFormatting(apiObject *quicksight.TableConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenTableConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []interface{}{tfMap}
}

func flattenTableConditionalFormattingOption(apiObject []*quicksight.TableConditionalFormattingOption) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.Cell != nil {
			tfMap["cell"] = flattenTableCellConditionalFormatting(config.Cell)
		}
		if config.Row != nil {
			tfMap["row"] = flattenTableRowConditionalFormatting(config.Row)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTableCellConditionalFormatting(apiObject *quicksight.TableCellConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.TextFormat != nil {
		tfMap["text_format"] = flattenTextConditionalFormat(apiObject.TextFormat)
	}

	return []interface{}{tfMap}
}

func flattenTableRowConditionalFormatting(apiObject *quicksight.TableRowConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = flattenConditionalFormattingColor(apiObject.BackgroundColor)
	}
	if apiObject.TextColor != nil {
		tfMap["text_color"] = flattenConditionalFormattingColor(apiObject.TextColor)
	}

	return []interface{}{tfMap}
}
