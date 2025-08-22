// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func pivotTableVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"field_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableFieldOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"data_path_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableDataPathOption.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"data_path_list": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathValue.html
														Type:     schema.TypeList,
														Required: true,
														MinItems: 1,
														MaxItems: 20,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"field_id":    stringLenBetweenSchema(attrRequired, 1, 512),
																"field_value": stringLenBetweenSchema(attrRequired, 1, 2048),
															},
														},
													},
													"width": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										"selected_field_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableFieldOption.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id":     stringLenBetweenSchema(attrRequired, 1, 512),
													"custom_label": stringLenBetweenSchema(attrOptional, 1, 2048),
													"visibility":   stringEnumSchema[awstypes.Visibility](attrOptional),
												},
											},
										},
									},
								},
							},
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"pivot_table_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"columns":        dimensionFieldSchema(dimensionsFieldMaxItems40), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"rows":           dimensionFieldSchema(dimensionsFieldMaxItems40), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(measureFieldsMaxItems40),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"paginated_report_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTablePaginatedReportOptions.html
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
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"field_sort_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotFieldSortOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 200,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id": stringLenBetweenSchema(attrRequired, 1, 512),
													"sort_by": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableSortBy.html
														Type:     schema.TypeList,
														Required: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"column": columnSortSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnSort.html
																"data_path": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathSort.html
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"direction":  stringEnumSchema[awstypes.SortDirection](attrRequired),
																			"sort_paths": dataPathValueSchema(dataPathValueMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathValue.html
																		},
																	},
																},
																names.AttrField: fieldSortSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSort.html
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"table_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"cell_style":                          tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"collapsed_row_dimensions_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
										"column_header_style":                 tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"column_names_visibility":             stringEnumSchema[awstypes.Visibility](attrOptional),
										"metric_placement":                    stringEnumSchema[awstypes.PivotTableMetricPlacement](attrOptional),
										"row_alternate_color_options":         rowAlternateColorOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RowAlternateColorOptions.html
										"row_field_names_style":               tableCellStyleSchema(),           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"row_header_style":                    tableCellStyleSchema(),           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"single_metric_visibility":            stringEnumSchema[awstypes.Visibility](attrOptional),
										"toggle_buttons_visibility":           stringEnumSchema[awstypes.Visibility](attrOptional),
									},
								},
							},
							"total_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableTotalOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column_subtotal_options": subtotalOptionsSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SubtotalOptions.html
										"column_total_options":    pivotTotalOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTotalOptions.html
										"row_subtotal_options":    subtotalOptionsSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SubtotalOptions.html
										"row_total_options":       pivotTotalOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTotalOptions.html
									},
								},
							},
						},
					},
				},
				"conditional_formatting": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableConditionalFormatting.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"conditional_formatting_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableConditionalFormattingOption.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 100,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"cell": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableCellConditionalFormatting.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id": stringLenBetweenSchema(attrRequired, 1, 512),
													names.AttrScope: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableConditionalFormattingScope.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrRole: stringEnumSchema[awstypes.PivotTableConditionalFormattingScopeRole](attrOptional),
															},
														},
													},
													"text_format": textConditionalFormatSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextConditionalFormat.html
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

var tableBorderOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableBorderOptions.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"color":     hexColorSchema(attrOptional),
				"style":     stringEnumSchema[awstypes.TableBorderStyle](attrOptional),
				"thickness": intBetweenSchema(attrOptional, 1, 4),
			},
		},
	}
})

var tableCellStyleSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"background_color": hexColorSchema(attrOptional),
				"border": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GlobalTableBorderOptions.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"side_specific_border": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableSideBorderOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"bottom":           tableBorderOptionsSchema(),
										"inner_horizontal": tableBorderOptionsSchema(),
										"inner_vertical":   tableBorderOptionsSchema(),
										"left":             tableBorderOptionsSchema(),
										"right":            tableBorderOptionsSchema(),
										"top":              tableBorderOptionsSchema(),
									},
								},
							},
							"uniform_border": tableBorderOptionsSchema(),
						},
					},
				},
				"font_configuration":        fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
				"height":                    intBetweenSchema(attrOptional, 8, 500),
				"horizontal_text_alignment": stringEnumSchema[awstypes.HorizontalTextAlignment](attrOptional),
				"text_wrap":                 stringEnumSchema[awstypes.TextWrap](attrOptional),
				"vertical_text_alignment":   stringEnumSchema[awstypes.VerticalTextAlignment](attrOptional),
				"visibility":                stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

var subtotalOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SubtotalOptions.html
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
				"field_level": stringEnumSchema[awstypes.PivotTableSubtotalLevel](attrOptional),
				"field_level_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableFieldSubtotalOptions.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 100,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"field_id": stringLenBetweenSchema(attrOptional, 1, 512),
						},
					},
				},
				"metric_header_cell_style": tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
				"total_cell_style":         tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
				"totals_visibility":        stringEnumSchema[awstypes.Visibility](attrOptional),
				"value_cell_style":         tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
			},
		},
	}
})

var pivotTotalOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTotalOptions.html
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
				"metric_header_cell_style": tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
				"placement":                stringEnumSchema[awstypes.TableTotalsPlacement](attrOptional),
				"scroll_status":            stringEnumSchema[awstypes.TableTotalsScrollStatus](attrOptional),
				"total_cell_style":         tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
				"totals_visibility":        stringEnumSchema[awstypes.Visibility](attrOptional),
				"value_cell_style":         tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
			},
		},
	}
})

var rowAlternateColorOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RowAlternateColorOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"row_alternate_colors": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")},
				},
				names.AttrStatus: stringEnumSchema[awstypes.Status](attrOptional),
			},
		},
	}
})

var textConditionalFormatSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextConditionalFormat.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"background_color": conditionalFormattingColorSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingColor.html
				"icon":             conditionalFormattingIconSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingIcon.html
				"text_color":       conditionalFormattingColorSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ConditionalFormattingColor.html
			},
		},
	}
})

func expandPivotTableVisual(tfList []any) *awstypes.PivotTableVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandPivotTableConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]any); ok && len(v) > 0 {
		apiObject.ConditionalFormatting = expandPivotTableConditionalFormatting(v)
	}
	if v, ok := tfMap["subtitle"].([]any); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandPivotTableConfiguration(tfList []any) *awstypes.PivotTableConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableConfiguration{}

	if v, ok := tfMap["field_options"].([]any); ok && len(v) > 0 {
		apiObject.FieldOptions = expandPivotTableFieldOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandPivotTableFieldWells(v)
	}
	if v, ok := tfMap["paginated_report_options"].([]any); ok && len(v) > 0 {
		apiObject.PaginatedReportOptions = expandPivotTablePaginatedReportOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandPivotTableSortConfiguration(v)
	}
	if v, ok := tfMap["table_options"].([]any); ok && len(v) > 0 {
		apiObject.TableOptions = expandPivotTableOptions(v)
	}
	if v, ok := tfMap["total_options"].([]any); ok && len(v) > 0 {
		apiObject.TotalOptions = expandPivotTableTotalOptions(v)
	}

	return apiObject
}

func expandPivotTableFieldWells(tfList []any) *awstypes.PivotTableFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableFieldWells{}

	if v, ok := tfMap["pivot_table_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.PivotTableAggregatedFieldWells = expandPivotTableAggregatedFieldWells(v)
	}

	return apiObject
}

func expandPivotTableAggregatedFieldWells(tfList []any) *awstypes.PivotTableAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableAggregatedFieldWells{}

	if v, ok := tfMap["columns"].([]any); ok && len(v) > 0 {
		apiObject.Columns = expandDimensionFields(v)
	}
	if v, ok := tfMap["rows"].([]any); ok && len(v) > 0 {
		apiObject.Rows = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandPivotTableSortConfiguration(tfList []any) *awstypes.PivotTableSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableSortConfiguration{}

	if v, ok := tfMap["field_sort_options"].([]any); ok && len(v) > 0 {
		apiObject.FieldSortOptions = expandPivotFieldSortOptionsList(v)
	}

	return apiObject
}

func expandPivotFieldSortOptionsList(tfList []any) []awstypes.PivotFieldSortOptions {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PivotFieldSortOptions

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandPivotFieldSortOptions(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPivotFieldSortOptions(tfMap map[string]any) *awstypes.PivotFieldSortOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PivotFieldSortOptions{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["sort_by"].([]any); ok && len(v) > 0 {
		apiObject.SortBy = expandPivotTableSortBy(v)
	}

	return apiObject
}

func expandPivotTableSortBy(tfList []any) *awstypes.PivotTableSortBy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableSortBy{}

	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnSort(v)
	}
	if v, ok := tfMap["data_path"].([]any); ok && len(v) > 0 {
		apiObject.DataPath = expandDataPathSort(v)
	}
	if v, ok := tfMap[names.AttrField].([]any); ok && len(v) > 0 {
		apiObject.Field = expandFieldSort(v)
	}

	return apiObject
}

func expandDataPathSort(tfList []any) *awstypes.DataPathSort {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataPathSort{}

	if v, ok := tfMap["direction"].(string); ok && v != "" {
		apiObject.Direction = awstypes.SortDirection(v)
	}
	if v, ok := tfMap["sort_paths"].([]any); ok && len(v) > 0 {
		apiObject.SortPaths = expandDataPathValues(v)
	}

	return apiObject
}

func expandPivotTableFieldOptions(tfList []any) *awstypes.PivotTableFieldOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableFieldOptions{}

	if v, ok := tfMap["data_path_options"].([]any); ok && len(v) > 0 {
		apiObject.DataPathOptions = expandPivotTableDataPathOptions(v)
	}
	if v, ok := tfMap["selected_field_options"].([]any); ok && len(v) > 0 {
		apiObject.SelectedFieldOptions = expandPivotTableFieldOptionsList(v)
	}

	return apiObject
}

func expandPivotTableDataPathOptions(tfList []any) []awstypes.PivotTableDataPathOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PivotTableDataPathOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandPivotTableDataPathOption(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPivotTableDataPathOption(tfMap map[string]any) *awstypes.PivotTableDataPathOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PivotTableDataPathOption{}

	if v, ok := tfMap["width"].(string); ok && v != "" {
		apiObject.Width = aws.String(v)
	}
	if v, ok := tfMap["data_path_list"].([]any); ok && len(v) > 0 {
		apiObject.DataPathList = expandDataPathValues(v)
	}

	return apiObject
}

func expandPivotTableFieldOptionsList(tfList []any) []awstypes.PivotTableFieldOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PivotTableFieldOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandPivotTableFieldOption(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPivotTableFieldOption(tfMap map[string]any) *awstypes.PivotTableFieldOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PivotTableFieldOption{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		apiObject.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandPivotTablePaginatedReportOptions(tfList []any) *awstypes.PivotTablePaginatedReportOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTablePaginatedReportOptions{}

	if v, ok := tfMap["overflow_column_header_visibility"].(string); ok && v != "" {
		apiObject.OverflowColumnHeaderVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["vertical_overflow_visibility"].(string); ok && v != "" {
		apiObject.VerticalOverflowVisibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandPivotTableOptions(tfList []any) *awstypes.PivotTableOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableOptions{}

	if v, ok := tfMap["collapsed_row_dimensions_visibility"].(string); ok && v != "" {
		apiObject.CollapsedRowDimensionsVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["column_names_visibility"].(string); ok && v != "" {
		apiObject.ColumnNamesVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["metric_placement"].(string); ok && v != "" {
		apiObject.MetricPlacement = awstypes.PivotTableMetricPlacement(v)
	}
	if v, ok := tfMap["single_metric_visibility"].(string); ok && v != "" {
		apiObject.SingleMetricVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["toggle_buttons_visibility"].(string); ok && v != "" {
		apiObject.ToggleButtonsVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["cell_style"].([]any); ok && len(v) > 0 {
		apiObject.CellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["column_header_style"].([]any); ok && len(v) > 0 {
		apiObject.ColumnHeaderStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["row_alternate_color_options"].([]any); ok && len(v) > 0 {
		apiObject.RowAlternateColorOptions = expandRowAlternateColorOptions(v)
	}
	if v, ok := tfMap["row_field_names_style"].([]any); ok && len(v) > 0 {
		apiObject.RowFieldNamesStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["row_header_style"].([]any); ok && len(v) > 0 {
		apiObject.RowHeaderStyle = expandTableCellStyle(v)
	}

	return apiObject
}

func expandTableCellStyle(tfList []any) *awstypes.TableCellStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableCellStyle{}

	if v, ok := tfMap["background_color"].(string); ok && v != "" {
		apiObject.BackgroundColor = aws.String(v)
	}
	if v, ok := tfMap["height"].(int); ok {
		apiObject.Height = aws.Int32(int32(v))
	}
	if v, ok := tfMap["horizontal_text_alignment"].(string); ok && v != "" {
		apiObject.HorizontalTextAlignment = awstypes.HorizontalTextAlignment(v)
	}
	if v, ok := tfMap["text_wrap"].(string); ok && v != "" {
		apiObject.TextWrap = awstypes.TextWrap(v)
	}
	if v, ok := tfMap["vertical_text_alignment"].(string); ok && v != "" {
		apiObject.VerticalTextAlignment = awstypes.VerticalTextAlignment(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["border"].([]any); ok && len(v) > 0 {
		apiObject.Border = expandGlobalTableBorderOptions(v)
	}
	if v, ok := tfMap["font_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FontConfiguration = expandFontConfiguration(v)
	}

	return apiObject
}

func expandGlobalTableBorderOptions(tfList []any) *awstypes.GlobalTableBorderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GlobalTableBorderOptions{}

	if v, ok := tfMap["side_specific_border"].([]any); ok && len(v) > 0 {
		apiObject.SideSpecificBorder = expandTableSideBorderOptions(v)
	}
	if v, ok := tfMap["uniform_border"].([]any); ok && len(v) > 0 {
		apiObject.UniformBorder = expandTableBorderOptions(v)
	}

	return apiObject
}

func expandTableSideBorderOptions(tfList []any) *awstypes.TableSideBorderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableSideBorderOptions{}

	if v, ok := tfMap["bottom"].([]any); ok && len(v) > 0 {
		apiObject.Bottom = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["inner_horizontal"].([]any); ok && len(v) > 0 {
		apiObject.InnerHorizontal = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["inner_vertical"].([]any); ok && len(v) > 0 {
		apiObject.InnerVertical = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["left"].([]any); ok && len(v) > 0 {
		apiObject.Left = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["right"].([]any); ok && len(v) > 0 {
		apiObject.Right = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["top"].([]any); ok && len(v) > 0 {
		apiObject.Top = expandTableBorderOptions(v)
	}

	return apiObject
}

func expandTableBorderOptions(tfList []any) *awstypes.TableBorderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TableBorderOptions{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap["style"].(string); ok && v != "" {
		apiObject.Style = awstypes.TableBorderStyle(v)
	}
	if v, ok := tfMap["thickness"].(int); ok {
		apiObject.Thickness = aws.Int32(int32(v))
	}

	return apiObject
}

func expandPivotTableTotalOptions(tfList []any) *awstypes.PivotTableTotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableTotalOptions{}

	if v, ok := tfMap["column_subtotal_options"].([]any); ok && len(v) > 0 {
		apiObject.ColumnSubtotalOptions = expandSubtotalOptions(v)
	}
	if v, ok := tfMap["column_total_options"].([]any); ok && len(v) > 0 {
		apiObject.ColumnTotalOptions = expandPivotTotalOptions(v)
	}
	if v, ok := tfMap["row_subtotal_options"].([]any); ok && len(v) > 0 {
		apiObject.RowSubtotalOptions = expandSubtotalOptions(v)
	}
	if v, ok := tfMap["row_total_options"].([]any); ok && len(v) > 0 {
		apiObject.RowTotalOptions = expandPivotTotalOptions(v)
	}

	return apiObject
}

func expandSubtotalOptions(tfList []any) *awstypes.SubtotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.SubtotalOptions{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		apiObject.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["field_level"].(string); ok && v != "" {
		apiObject.FieldLevel = awstypes.PivotTableSubtotalLevel(v)
	}
	if v, ok := tfMap["totals_visibility"].(string); ok && v != "" {
		apiObject.TotalsVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["field_level_options"].([]any); ok && len(v) > 0 {
		apiObject.FieldLevelOptions = expandPivotTableFieldSubtotalOptionsList(v)
	}
	if v, ok := tfMap["metric_header_cell_style"].([]any); ok && len(v) > 0 {
		apiObject.MetricHeaderCellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["total_cell_style"].([]any); ok && len(v) > 0 {
		apiObject.TotalCellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["value_cell_style"].([]any); ok && len(v) > 0 {
		apiObject.ValueCellStyle = expandTableCellStyle(v)
	}

	return apiObject
}

func expandPivotTableFieldSubtotalOptionsList(tfList []any) []awstypes.PivotTableFieldSubtotalOptions {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PivotTableFieldSubtotalOptions

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandPivotTableFieldSubtotalOptions(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPivotTableFieldSubtotalOptions(tfMap map[string]any) *awstypes.PivotTableFieldSubtotalOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PivotTableFieldSubtotalOptions{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}

	return apiObject
}

func expandPivotTotalOptions(tfList []any) *awstypes.PivotTotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTotalOptions{}

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
	if v, ok := tfMap["metric_header_cell_style"].([]any); ok && len(v) > 0 {
		apiObject.MetricHeaderCellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["total_cell_style"].([]any); ok && len(v) > 0 {
		apiObject.TotalCellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["value_cell_style"].([]any); ok && len(v) > 0 {
		apiObject.ValueCellStyle = expandTableCellStyle(v)
	}

	return apiObject
}

func expandRowAlternateColorOptions(tfList []any) *awstypes.RowAlternateColorOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RowAlternateColorOptions{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.WidgetStatus(v)
	}
	if v, ok := tfMap["row_alternate_colors"].([]any); ok && len(v) > 0 {
		apiObject.RowAlternateColors = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandPivotTableConditionalFormatting(tfList []any) *awstypes.PivotTableConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]any); ok && len(v) > 0 {
		apiObject.ConditionalFormattingOptions = expandPivotTableConditionalFormattingOptions(v)
	}

	return apiObject
}

func expandPivotTableConditionalFormattingOptions(tfList []any) []awstypes.PivotTableConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PivotTableConditionalFormattingOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandPivotTableConditionalFormattingOption(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPivotTableConditionalFormattingOption(tfMap map[string]any) *awstypes.PivotTableConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PivotTableConditionalFormattingOption{}

	if v, ok := tfMap["cell"].([]any); ok && len(v) > 0 {
		apiObject.Cell = expandPivotTableCellConditionalFormatting(v)
	}

	return apiObject
}

func expandPivotTableCellConditionalFormatting(tfList []any) *awstypes.PivotTableCellConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableCellConditionalFormatting{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrScope].([]any); ok && len(v) > 0 {
		apiObject.Scope = expandPivotTableConditionalFormattingScope(v)
	}
	if v, ok := tfMap["text_format"].([]any); ok && len(v) > 0 {
		apiObject.TextFormat = expandTextConditionalFormat(v)
	}

	return apiObject
}

func expandPivotTableConditionalFormattingScope(tfList []any) *awstypes.PivotTableConditionalFormattingScope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PivotTableConditionalFormattingScope{}

	if v, ok := tfMap[names.AttrRole].(string); ok && v != "" {
		apiObject.Role = awstypes.PivotTableConditionalFormattingScopeRole(v)
	}

	return apiObject
}

func flattenPivotTableVisual(apiObject *awstypes.PivotTableVisual) []any {
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
		tfMap["chart_configuration"] = flattenPivotTableConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.ConditionalFormatting != nil {
		tfMap["conditional_formatting"] = flattenPivotTableConditionalFormatting(apiObject.ConditionalFormatting)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []any{tfMap}
}

func flattenPivotTableConfiguration(apiObject *awstypes.PivotTableConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldOptions != nil {
		tfMap["field_options"] = flattenPivotTableFieldOptions(apiObject.FieldOptions)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenPivotTableFieldWells(apiObject.FieldWells)
	}
	if apiObject.PaginatedReportOptions != nil {
		tfMap["paginated_report_options"] = flattenPivotTablePaginatedReportOptions(apiObject.PaginatedReportOptions)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenPivotTableSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.TableOptions != nil {
		tfMap["table_options"] = flattenPivotTableOptions(apiObject.TableOptions)
	}
	if apiObject.TotalOptions != nil {
		tfMap["total_options"] = flattenPivotTableTotalOptions(apiObject.TotalOptions)
	}

	return []any{tfMap}
}

func flattenPivotTableFieldOptions(apiObject *awstypes.PivotTableFieldOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DataPathOptions != nil {
		tfMap["data_path_options"] = flattenPivotTableDataPathOption(apiObject.DataPathOptions)
	}
	if apiObject.SelectedFieldOptions != nil {
		tfMap["selected_field_options"] = flattenPivotTableFieldOption(apiObject.SelectedFieldOptions)
	}

	return []any{tfMap}
}

func flattenPivotTableDataPathOption(apiObjects []awstypes.PivotTableDataPathOption) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.DataPathList != nil {
			tfMap["data_path_list"] = flattenDataPathValues(apiObject.DataPathList)
		}
		if apiObject.Width != nil {
			tfMap["width"] = aws.ToString(apiObject.Width)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTableFieldWells(apiObject *awstypes.PivotTableFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.PivotTableAggregatedFieldWells != nil {
		tfMap["pivot_table_aggregated_field_wells"] = flattenPivotTableAggregatedFieldWells(apiObject.PivotTableAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenPivotTableAggregatedFieldWells(apiObject *awstypes.PivotTableAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Columns != nil {
		tfMap["columns"] = flattenDimensionFields(apiObject.Columns)
	}
	if apiObject.Rows != nil {
		tfMap["rows"] = flattenDimensionFields(apiObject.Rows)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []any{tfMap}
}

func flattenPivotTablePaginatedReportOptions(apiObject *awstypes.PivotTablePaginatedReportOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"overflow_column_header_visibility": apiObject.OverflowColumnHeaderVisibility,
		"vertical_overflow_visibility":      apiObject.VerticalOverflowVisibility,
	}

	return []any{tfMap}
}

func flattenPivotTableSortConfiguration(apiObject *awstypes.PivotTableSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldSortOptions != nil {
		tfMap["field_sort_options"] = flattenPivotFieldSortOptions(apiObject.FieldSortOptions)
	}

	return []any{tfMap}
}

func flattenPivotFieldSortOptions(apiObjects []awstypes.PivotFieldSortOptions) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.FieldId != nil {
			tfMap["field_id"] = aws.ToString(apiObject.FieldId)
		}
		if apiObject.SortBy != nil {
			tfMap["sort_by"] = flattenPivotTableSortBy(apiObject.SortBy)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTableSortBy(apiObject *awstypes.PivotTableSortBy) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnSort(apiObject.Column)
	}
	if apiObject.DataPath != nil {
		tfMap["data_path"] = flattenDataPathSort(apiObject.DataPath)
	}
	if apiObject.Field != nil {
		tfMap[names.AttrField] = flattenFieldSort(apiObject.Field)
	}

	return []any{tfMap}
}

func flattenDataPathSort(apiObject *awstypes.DataPathSort) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"direction": apiObject.Direction,
	}

	if apiObject.SortPaths != nil {
		tfMap["sort_paths"] = flattenDataPathValues(apiObject.SortPaths)
	}

	return []any{tfMap}
}

func flattenPivotTableOptions(apiObject *awstypes.PivotTableOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"collapsed_row_dimensions_visibility": apiObject.CollapsedRowDimensionsVisibility,
		"column_names_visibility":             apiObject.ColumnNamesVisibility,
		"metric_placement":                    apiObject.MetricPlacement,
		"single_metric_visibility":            apiObject.SingleMetricVisibility,
		"toggle_buttons_visibility":           apiObject.ToggleButtonsVisibility,
	}

	if apiObject.CellStyle != nil {
		tfMap["cell_style"] = flattenTableCellStyle(apiObject.CellStyle)
	}
	if apiObject.ColumnHeaderStyle != nil {
		tfMap["column_header_style"] = flattenTableCellStyle(apiObject.ColumnHeaderStyle)
	}
	if apiObject.RowAlternateColorOptions != nil {
		tfMap["row_alternate_color_options"] = flattenRowAlternateColorOptions(apiObject.RowAlternateColorOptions)
	}
	if apiObject.RowFieldNamesStyle != nil {
		tfMap["row_field_names_style"] = flattenTableCellStyle(apiObject.RowFieldNamesStyle)
	}
	if apiObject.RowHeaderStyle != nil {
		tfMap["row_header_style"] = flattenTableCellStyle(apiObject.RowHeaderStyle)
	}

	return []any{tfMap}
}

func flattenTableCellStyle(apiObject *awstypes.TableCellStyle) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"horizontal_text_alignment": apiObject.HorizontalTextAlignment,
		"text_wrap":                 apiObject.TextWrap,
		"vertical_text_alignment":   apiObject.VerticalTextAlignment,
		"visibility":                apiObject.Visibility,
	}

	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = aws.ToString(apiObject.BackgroundColor)
	}
	if apiObject.Border != nil {
		tfMap["border"] = flattenGlobalTableBorderOptions(apiObject.Border)
	}
	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	if apiObject.Height != nil {
		tfMap["height"] = aws.ToInt32(apiObject.Height)
	}

	return []any{tfMap}
}

func flattenGlobalTableBorderOptions(apiObject *awstypes.GlobalTableBorderOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.SideSpecificBorder != nil {
		tfMap["side_specific_border"] = flattenTableSideBorderOptions(apiObject.SideSpecificBorder)
	}
	if apiObject.UniformBorder != nil {
		tfMap["uniform_border"] = flattenTableBorderOptions(apiObject.UniformBorder)
	}

	return []any{tfMap}
}

func flattenTableSideBorderOptions(apiObject *awstypes.TableSideBorderOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Bottom != nil {
		tfMap["bottom"] = flattenTableBorderOptions(apiObject.Bottom)
	}
	if apiObject.InnerHorizontal != nil {
		tfMap["inner_horizontal"] = flattenTableBorderOptions(apiObject.InnerHorizontal)
	}
	if apiObject.InnerVertical != nil {
		tfMap["inner_vertical"] = flattenTableBorderOptions(apiObject.InnerVertical)
	}
	if apiObject.Left != nil {
		tfMap["left"] = flattenTableBorderOptions(apiObject.Left)
	}
	if apiObject.Right != nil {
		tfMap["right"] = flattenTableBorderOptions(apiObject.Right)
	}
	if apiObject.Top != nil {
		tfMap["top"] = flattenTableBorderOptions(apiObject.Top)
	}

	return []any{tfMap}
}

func flattenTableBorderOptions(apiObject *awstypes.TableBorderOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"style": apiObject.Style,
	}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	if apiObject.Thickness != nil {
		tfMap["thickness"] = aws.ToInt32(apiObject.Thickness)
	}

	return []any{tfMap}
}

func flattenRowAlternateColorOptions(apiObject *awstypes.RowAlternateColorOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrStatus: apiObject.Status,
	}

	if apiObject.RowAlternateColors != nil {
		tfMap["row_alternate_colors"] = apiObject.RowAlternateColors
	}

	return []any{tfMap}
}

func flattenPivotTableTotalOptions(apiObject *awstypes.PivotTableTotalOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ColumnSubtotalOptions != nil {
		tfMap["column_subtotal_options"] = flattenSubtotalOptions(apiObject.ColumnSubtotalOptions)
	}
	if apiObject.ColumnTotalOptions != nil {
		tfMap["column_total_options"] = flattenPivotTotalOptions(apiObject.ColumnTotalOptions)
	}
	if apiObject.RowSubtotalOptions != nil {
		tfMap["row_subtotal_options"] = flattenSubtotalOptions(apiObject.RowSubtotalOptions)
	}
	if apiObject.RowTotalOptions != nil {
		tfMap["row_total_options"] = flattenPivotTotalOptions(apiObject.RowTotalOptions)
	}

	return []any{tfMap}
}

func flattenSubtotalOptions(apiObject *awstypes.SubtotalOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"field_level":       apiObject.FieldLevel,
		"totals_visibility": apiObject.TotalsVisibility,
	}

	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
	}
	if apiObject.FieldLevelOptions != nil {
		tfMap["field_level_options"] = flattenPivotTableFieldSubtotalOptions(apiObject.FieldLevelOptions)
	}
	if apiObject.MetricHeaderCellStyle != nil {
		tfMap["metric_header_cell_style"] = flattenTableCellStyle(apiObject.MetricHeaderCellStyle)
	}
	if apiObject.TotalCellStyle != nil {
		tfMap["total_cell_style"] = flattenTableCellStyle(apiObject.TotalCellStyle)
	}
	if apiObject.ValueCellStyle != nil {
		tfMap["value_cell_style"] = flattenTableCellStyle(apiObject.ValueCellStyle)
	}

	return []any{tfMap}
}

func flattenPivotTableFieldSubtotalOptions(apiObjects []awstypes.PivotTableFieldSubtotalOptions) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.FieldId != nil {
			tfMap["field_id"] = aws.ToString(apiObject.FieldId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTotalOptions(apiObject *awstypes.PivotTotalOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"placement":         apiObject.Placement,
		"scroll_status":     apiObject.ScrollStatus,
		"totals_visibility": apiObject.TotalsVisibility,
	}

	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
	}
	if apiObject.MetricHeaderCellStyle != nil {
		tfMap["metric_header_cell_style"] = flattenTableCellStyle(apiObject.MetricHeaderCellStyle)
	}
	if apiObject.TotalCellStyle != nil {
		tfMap["total_cell_style"] = flattenTableCellStyle(apiObject.TotalCellStyle)
	}
	if apiObject.ValueCellStyle != nil {
		tfMap["value_cell_style"] = flattenTableCellStyle(apiObject.ValueCellStyle)
	}

	return []any{tfMap}
}

func flattenPivotTableFieldOption(apiObjects []awstypes.PivotTableFieldOption) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"visibility": apiObject.Visibility,
		}

		if apiObject.FieldId != nil {
			tfMap["field_id"] = aws.ToString(apiObject.FieldId)
		}
		if apiObject.CustomLabel != nil {
			tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTableConditionalFormatting(apiObject *awstypes.PivotTableConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenPivotTableConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []any{tfMap}
}

func flattenPivotTableConditionalFormattingOption(apiObjects []awstypes.PivotTableConditionalFormattingOption) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Cell != nil {
			tfMap["cell"] = flattenPivotTableCellConditionalFormatting(apiObject.Cell)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTableCellConditionalFormatting(apiObject *awstypes.PivotTableCellConditionalFormatting) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"field_id": aws.ToString(apiObject.FieldId),
	}

	if apiObject.Scope != nil {
		tfMap[names.AttrScope] = flattenPivotTableConditionalFormattingScope(apiObject.Scope)
	}
	if apiObject.TextFormat != nil {
		tfMap["text_format"] = flattenTextConditionalFormat(apiObject.TextFormat)
	}

	return []any{tfMap}
}

func flattenPivotTableConditionalFormattingScope(apiObject *awstypes.PivotTableConditionalFormattingScope) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrRole: apiObject.Role,
	}

	return []any{tfMap}
}

func flattenTextConditionalFormat(apiObject *awstypes.TextConditionalFormat) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = flattenConditionalFormattingColor(apiObject.BackgroundColor)
	}
	if apiObject.Icon != nil {
		tfMap["icon"] = flattenConditionalFormattingIcon(apiObject.Icon)
	}
	if apiObject.TextColor != nil {
		tfMap["text_color"] = flattenConditionalFormattingColor(apiObject.TextColor)
	}

	return []any{tfMap}
}
