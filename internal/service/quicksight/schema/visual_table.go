// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
											Elem:     &schema.Schema{Type: schema.TypeString, ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 512))},
										},
										"selected_field_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableFieldOption.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id":     stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
													"custom_label": stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
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
																						"table_cell_image_scaling_configuration": stringSchema(false, enum.Validate[types.TableCellImageScalingConfiguration]()),
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
																									"icon": stringSchema(false, enum.Validate[types.TableFieldIconSetType]()),
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
																			"target": stringSchema(false, enum.Validate[types.URLTargetConfiguration]()),
																		},
																	},
																},
															},
														},
													},
													"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
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
																"field_id":             stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
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
										"overflow_column_header_visibility": stringSchema(false, enum.Validate[types.Visibility]()),
										"vertical_overflow_visibility":      stringSchema(false, enum.Validate[types.Visibility]()),
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
													"field_id":       stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
													"negative_color": stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
													"positive_color": stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
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
										"orientation":                 stringSchema(false, enum.Validate[types.TableOrientation]()),
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
										"placement":         stringSchema(false, enum.Validate[types.TableTotalsPlacement]()),
										"scroll_status":     stringSchema(false, enum.Validate[types.TableTotalsScrollStatus]()),
										"total_cell_style":  tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"totals_visibility": stringSchema(false, enum.Validate[types.Visibility]()),
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
													"field_id":    stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
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

func expandTableVisual(tfList []interface{}) *types.TableVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.TableVisual{}

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

func expandTableConfiguration(tfList []interface{}) *types.TableConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TableConfiguration{}

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

func expandTableFieldWells(tfList []interface{}) *types.TableFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TableFieldWells{}

	if v, ok := tfMap["table_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.TableAggregatedFieldWells = expandTableAggregatedFieldWells(v)
	}
	if v, ok := tfMap["table_unaggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.TableUnaggregatedFieldWells = expandTableUnaggregatedFieldWells(v)
	}

	return config
}

func expandTableAggregatedFieldWells(tfList []interface{}) *types.TableAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TableAggregatedFieldWells{}

	if v, ok := tfMap["group_by"].([]interface{}); ok && len(v) > 0 {
		config.GroupBy = expandDimensionFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandTableUnaggregatedFieldWells(tfList []interface{}) *types.TableUnaggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TableUnaggregatedFieldWells{}

	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandUnaggregatedFields(v)
	}

	return config
}

func expandUnaggregatedFields(tfList []interface{}) []types.UnaggregatedField {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.UnaggregatedField
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandUnaggregatedField(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandUnaggregatedField(tfMap map[string]interface{}) *types.UnaggregatedField {
	if tfMap == nil {
		return nil
	}

	options := &types.UnaggregatedField{}

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

func expandTableSortConfiguration(tfList []interface{}) *types.TableSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.TableSortConfiguration{}

	if v, ok := tfMap["pagination_configuration"].([]interface{}); ok && len(v) > 0 {
		config.PaginationConfiguration = expandPaginationConfiguration(v)
	}
	if v, ok := tfMap["row_sort"].([]interface{}); ok && len(v) > 0 {
		config.RowSort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandTableFieldOptions(tfList []interface{}) *types.TableFieldOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableFieldOptions{}

	if v, ok := tfMap["order"].([]interface{}); ok && len(v) > 0 {
		options.Order = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["selected_field_options"].([]interface{}); ok && len(v) > 0 {
		options.SelectedFieldOptions = expandTableFieldOptionsList(v)
	}

	return options
}

func expandTableFieldOptionsList(tfList []interface{}) []types.TableFieldOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.TableFieldOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandTableFieldOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandTableFieldOption(tfMap map[string]interface{}) *types.TableFieldOption {
	if tfMap == nil {
		return nil
	}

	options := &types.TableFieldOption{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		options.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}
	if v, ok := tfMap["width"].(string); ok && v != "" {
		options.Width = aws.String(v)
	}
	if v, ok := tfMap["url_styling"].([]interface{}); ok && len(v) > 0 {
		options.URLStyling = expandTableFieldURLConfiguration(v)
	}

	return options
}

func expandTableFieldURLConfiguration(tfList []interface{}) *types.TableFieldURLConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableFieldURLConfiguration{}

	if v, ok := tfMap["image_configuration"].([]interface{}); ok && len(v) > 0 {
		options.ImageConfiguration = expandTableFieldImageConfiguration(v)
	}
	if v, ok := tfMap["link_configuration"].([]interface{}); ok && len(v) > 0 {
		options.LinkConfiguration = expandTableFieldLinkConfiguration(v)
	}

	return options
}

func expandTableFieldImageConfiguration(tfList []interface{}) *types.TableFieldImageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableFieldImageConfiguration{}

	if v, ok := tfMap["sizing_options"].([]interface{}); ok && len(v) > 0 {
		options.SizingOptions = expandTableCellImageSizingConfiguration(v)
	}

	return options
}

func expandTableCellImageSizingConfiguration(tfList []interface{}) *types.TableCellImageSizingConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableCellImageSizingConfiguration{}

	if v, ok := tfMap["table_cell_image_scaling_configuration"].(string); ok && v != "" {
		options.TableCellImageScalingConfiguration = types.TableCellImageScalingConfiguration(v)
	}

	return options
}

func expandTableFieldLinkConfiguration(tfList []interface{}) *types.TableFieldLinkConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableFieldLinkConfiguration{}

	if v, ok := tfMap["target"].(string); ok && v != "" {
		options.Target = types.URLTargetConfiguration(v)
	}
	if v, ok := tfMap["content"].([]interface{}); ok && len(v) > 0 {
		options.Content = expandTableFieldLinkContentConfiguration(v)
	}

	return options
}

func expandTableFieldLinkContentConfiguration(tfList []interface{}) *types.TableFieldLinkContentConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableFieldLinkContentConfiguration{}

	if v, ok := tfMap["custom_icon_content"].([]interface{}); ok && len(v) > 0 {
		options.CustomIconContent = expandTableFieldCustomIconContent(v)
	}
	if v, ok := tfMap["custom_text_content"].([]interface{}); ok && len(v) > 0 {
		options.CustomTextContent = expandTableFieldCustomTextContent(v)
	}

	return options
}

func expandTableFieldCustomIconContent(tfList []interface{}) *types.TableFieldCustomIconContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableFieldCustomIconContent{}

	if v, ok := tfMap["icon"].(string); ok && v != "" {
		options.Icon = types.TableFieldIconSetType(v)
	}

	return options
}

func expandTableFieldCustomTextContent(tfList []interface{}) *types.TableFieldCustomTextContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableFieldCustomTextContent{}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		options.Value = aws.String(v)
	}
	if v, ok := tfMap["custom_text_content"].([]interface{}); ok && len(v) > 0 {
		options.FontConfiguration = expandFontConfiguration(v)
	}

	return options
}

func expandTablePaginatedReportOptions(tfList []interface{}) *types.TablePaginatedReportOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TablePaginatedReportOptions{}

	if v, ok := tfMap["overflow_column_header_visibility"].(string); ok && v != "" {
		options.OverflowColumnHeaderVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["vertical_overflow_visibility"].(string); ok && v != "" {
		options.VerticalOverflowVisibility = types.Visibility(v)
	}

	return options
}

func expandTableOptions(tfList []interface{}) *types.TableOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableOptions{}

	if v, ok := tfMap["orientation"].(string); ok && v != "" {
		options.Orientation = types.TableOrientation(v)
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

func expandTableTotalOptions(tfList []interface{}) *types.TotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TotalOptions{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		options.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["placement"].(string); ok && v != "" {
		options.Placement = types.TableTotalsPlacement(v)
	}
	if v, ok := tfMap["scroll_status"].(string); ok && v != "" {
		options.ScrollStatus = types.TableTotalsScrollStatus(v)
	}
	if v, ok := tfMap["totals_visibility"].(string); ok && v != "" {
		options.TotalsVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["total_cell_style"].([]interface{}); ok && len(v) > 0 {
		options.TotalCellStyle = expandTableCellStyle(v)
	}

	return options
}

func expandTableConditionalFormatting(tfList []interface{}) *types.TableConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]interface{}); ok && len(v) > 0 {
		options.ConditionalFormattingOptions = expandTableConditionalFormattingOptions(v)
	}

	return options
}

func expandTableConditionalFormattingOptions(tfList []interface{}) []types.TableConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.TableConditionalFormattingOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandTableConditionalFormattingOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandTableConditionalFormattingOption(tfMap map[string]interface{}) *types.TableConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	options := &types.TableConditionalFormattingOption{}

	if v, ok := tfMap["cell"].([]interface{}); ok && len(v) > 0 {
		options.Cell = expandTableCellConditionalFormatting(v)
	}
	if v, ok := tfMap["row"].([]interface{}); ok && len(v) > 0 {
		options.Row = expandTableRowConditionalFormatting(v)
	}

	return options
}

func expandTableCellConditionalFormatting(tfList []interface{}) *types.TableCellConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableCellConditionalFormatting{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["text_format"].([]interface{}); ok && len(v) > 0 {
		options.TextFormat = expandTextConditionalFormat(v)
	}

	return options
}

func expandTableRowConditionalFormatting(tfList []interface{}) *types.TableRowConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TableRowConditionalFormatting{}

	if v, ok := tfMap["background_color"].([]interface{}); ok && len(v) > 0 {
		options.BackgroundColor = expandConditionalFormattingColor(v)
	}
	if v, ok := tfMap["text_color"].([]interface{}); ok && len(v) > 0 {
		options.TextColor = expandConditionalFormattingColor(v)
	}

	return options
}

func expandTableInlineVisualizations(tfList []interface{}) []types.TableInlineVisualization {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.TableInlineVisualization
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandTableInlineVisualization(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandTableInlineVisualization(tfMap map[string]interface{}) *types.TableInlineVisualization {
	if tfMap == nil {
		return nil
	}

	options := &types.TableInlineVisualization{}

	if v, ok := tfMap["data_bars"].([]interface{}); ok && len(v) > 0 {
		options.DataBars = expandDataBarsOptions(v)
	}

	return options
}

func expandDataBarsOptions(tfList []interface{}) *types.DataBarsOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.DataBarsOptions{}

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

func flattenTableVisual(apiObject *types.TableVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id": aws.ToString(apiObject.VisualId),
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

func flattenTableConfiguration(apiObject *types.TableConfiguration) []interface{} {
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

func flattenTableFieldOptions(apiObject *types.TableFieldOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Order != nil {
		tfMap["order"] = flex.FlattenStringValueList(apiObject.Order)
	}
	if apiObject.SelectedFieldOptions != nil {
		tfMap["selected_field_options"] = flattenTableFieldOption(apiObject.SelectedFieldOptions)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldOption(apiObject []types.TableFieldOption) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.FieldId != nil {
			tfMap["field_id"] = aws.ToString(config.FieldId)
		}
		if config.CustomLabel != nil {
			tfMap["custom_label"] = aws.ToString(config.CustomLabel)
		}
		if config.URLStyling != nil {
			tfMap["url_styling"] = flattenTableFieldURLConfiguration(config.URLStyling)
		}
		tfMap["visbility"] = types.Visibility(config.Visibility)
		if config.Width != nil {
			tfMap["width"] = aws.ToString(config.Width)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTableFieldURLConfiguration(apiObject *types.TableFieldURLConfiguration) []interface{} {
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

func flattenTableFieldImageConfiguration(apiObject *types.TableFieldImageConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SizingOptions != nil {
		tfMap["sizing_options"] = flattenTableCellImageSizingConfiguration(apiObject.SizingOptions)
	}

	return []interface{}{tfMap}
}

func flattenTableCellImageSizingConfiguration(apiObject *types.TableCellImageSizingConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["table_cell_image_scaling_configuration"] = types.TableCellImageScalingConfiguration(apiObject.TableCellImageScalingConfiguration)

	return []interface{}{tfMap}
}

func flattenTableFieldLinkConfiguration(apiObject *types.TableFieldLinkConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Content != nil {
		tfMap["content"] = flattenTableFieldLinkContentConfiguration(apiObject.Content)
	}
	tfMap["target"] = types.URLTargetConfiguration(apiObject.Target)

	return []interface{}{tfMap}
}

func flattenTableFieldLinkContentConfiguration(apiObject *types.TableFieldLinkContentConfiguration) []interface{} {
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

func flattenTableFieldCustomIconContent(apiObject *types.TableFieldCustomIconContent) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["icon"] = types.TableFieldIconSetType(apiObject.Icon)

	return []interface{}{tfMap}
}

func flattenTableFieldCustomTextContent(apiObject *types.TableFieldCustomTextContent) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	if apiObject.Value != nil {
		tfMap["value"] = aws.ToString(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenTableFieldWells(apiObject *types.TableFieldWells) []interface{} {
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

func flattenTableAggregatedFieldWells(apiObject *types.TableAggregatedFieldWells) []interface{} {
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

func flattenTableUnaggregatedFieldWells(apiObject *types.TableUnaggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Values != nil {
		tfMap["values"] = flattenUnaggregatedField(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenUnaggregatedField(apiObject []types.UnaggregatedField) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.Column != nil {
			tfMap["column"] = flattenColumnIdentifier(config.Column)
		}
		if config.FieldId != nil {
			tfMap["field_id"] = aws.ToString(config.FieldId)
		}
		if config.FormatConfiguration != nil {
			tfMap["format_configuration"] = flattenFormatConfiguration(config.FormatConfiguration)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTablePaginatedReportOptions(apiObject *types.TablePaginatedReportOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["overflow_column_header_visibility"] = types.Visibility(apiObject.OverflowColumnHeaderVisibility)
	tfMap["vertical_overflow_visibility"] = types.Visibility(apiObject.VerticalOverflowVisibility)

	return []interface{}{tfMap}
}

func flattenTableSortConfiguration(apiObject *types.TableSortConfiguration) []interface{} {
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

func flattenTableInlineVisualization(apiObject []types.TableInlineVisualization) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.DataBars != nil {
			tfMap["data_bars"] = flattenDataBarsOptions(config.DataBars)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataBarsOptions(apiObject *types.DataBarsOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.NegativeColor != nil {
		tfMap["negative_color"] = aws.ToString(apiObject.NegativeColor)
	}
	if apiObject.PositiveColor != nil {
		tfMap["positive_color"] = aws.ToString(apiObject.PositiveColor)
	}

	return []interface{}{tfMap}
}

func flattenTableOptions(apiObject *types.TableOptions) []interface{} {
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
	tfMap["orientation"] = types.TableOrientation(apiObject.Orientation)
	if apiObject.RowAlternateColorOptions != nil {
		tfMap["row_alternate_color_options"] = flattenRowAlternateColorOptions(apiObject.RowAlternateColorOptions)
	}

	return []interface{}{tfMap}
}

func flattenTotalOptions(apiObject *types.TotalOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
	tfMap["placement"] = types.TableTotalsPlacement(apiObject.Placement)
	tfMap["scroll_status"] = types.TableTotalsScrollStatus(apiObject.ScrollStatus)
	if apiObject.TotalCellStyle != nil {
		tfMap["total_cell_style"] = flattenTableCellStyle(apiObject.TotalCellStyle)
	}
	tfMap["totals_visibility"] = types.Visibility(apiObject.TotalsVisibility)

	return []interface{}{tfMap}
}

func flattenTableConditionalFormatting(apiObject *types.TableConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenTableConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []interface{}{tfMap}
}

func flattenTableConditionalFormattingOption(apiObject []types.TableConditionalFormattingOption) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

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

func flattenTableCellConditionalFormatting(apiObject *types.TableCellConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.TextFormat != nil {
		tfMap["text_format"] = flattenTextConditionalFormat(apiObject.TextFormat)
	}

	return []interface{}{tfMap}
}

func flattenTableRowConditionalFormatting(apiObject *types.TableRowConditionalFormatting) []interface{} {
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
