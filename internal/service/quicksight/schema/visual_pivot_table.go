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
																"field_id":    stringSchema(true, validation.StringLenBetween(1, 512)),
																"field_value": stringSchema(true, validation.StringLenBetween(1, 2048)),
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
													"field_id":     stringSchema(true, validation.StringLenBetween(1, 512)),
													"custom_label": stringSchema(false, validation.StringLenBetween(1, 2048)),
													"visibility":   stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
										"overflow_column_header_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
										"vertical_overflow_visibility":      stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
													"field_id": stringSchema(true, validation.StringLenBetween(1, 512)),
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
																			"direction":  stringSchema(true, validation.StringInSlice(quicksight.SortDirection_Values(), false)),
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
										"collapsed_row_dimensions_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
										"column_header_style":                 tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"column_names_visibility":             stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
										"metric_placement":                    stringSchema(false, validation.StringInSlice(quicksight.PivotTableMetricPlacement_Values(), false)),
										"row_alternate_color_options":         rowAlternateColorOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RowAlternateColorOptions.html
										"row_field_names_style":               tableCellStyleSchema(),           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"row_header_style":                    tableCellStyleSchema(),           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
										"single_metric_visibility":            stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
										"toggle_buttons_visibility":           stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
													"field_id": stringSchema(true, validation.StringLenBetween(1, 512)),
													names.AttrScope: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableConditionalFormattingScope.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrRole: stringSchema(false, validation.StringInSlice(quicksight.PivotTableConditionalFormattingScopeRole_Values(), false)),
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

func tableBorderOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableBorderOptions.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"color":     stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
				"style":     stringSchema(false, validation.StringInSlice(quicksight.TableBorderStyle_Values(), false)),
				"thickness": intSchema(false, validation.IntBetween(1, 4)),
			},
		},
	}
}

func tableCellStyleSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"background_color": stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), "")),
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
				"height":                    intSchema(false, validation.IntBetween(8, 500)),
				"horizontal_text_alignment": stringSchema(false, validation.StringInSlice(quicksight.HorizontalTextAlignment_Values(), false)),
				"text_wrap":                 stringSchema(false, validation.StringInSlice(quicksight.TextWrap_Values(), false)),
				"vertical_text_alignment":   stringSchema(false, validation.StringInSlice(quicksight.VerticalTextAlignment_Values(), false)),
				"visibility":                stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
			},
		},
	}
}

func subtotalOptionsSchema() *schema.Schema {
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
				"field_level": stringSchema(false, validation.StringInSlice(quicksight.PivotTableSubtotalLevel_Values(), false)),
				"field_level_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PivotTableFieldSubtotalOptions.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 100,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"field_id": stringSchema(false, validation.StringLenBetween(1, 512)),
						},
					},
				},
				"metric_header_cell_style": tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
				"total_cell_style":         tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
				"totals_visibility":        stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
				"value_cell_style":         tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
			},
		},
	}
}

func pivotTotalOptionsSchema() *schema.Schema {
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
				"placement":                stringSchema(false, validation.StringInSlice(quicksight.TableTotalsPlacement_Values(), false)),
				"scroll_status":            stringSchema(false, validation.StringInSlice(quicksight.TableTotalsScrollStatus_Values(), false)),
				"total_cell_style":         tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
				"totals_visibility":        stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
				"value_cell_style":         tableCellStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TableCellStyle.html
			},
		},
	}
}

func rowAlternateColorOptionsSchema() *schema.Schema {
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
				names.AttrStatus: stringSchema(false, validation.StringInSlice(quicksight.Status_Values(), false)),
			},
		},
	}
}

func textConditionalFormatSchema() *schema.Schema {
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
}

func expandPivotTableVisual(tfList []interface{}) *quicksight.PivotTableVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.PivotTableVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandPivotTableConfiguration(v)
	}
	if v, ok := tfMap["conditional_formatting"].([]interface{}); ok && len(v) > 0 {
		visual.ConditionalFormatting = expandPivotTableConditionalFormatting(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		visual.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		visual.Title = expandVisualTitleLabelOptions(v)
	}

	return visual
}

func expandPivotTableConfiguration(tfList []interface{}) *quicksight.PivotTableConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PivotTableConfiguration{}

	if v, ok := tfMap["field_options"].([]interface{}); ok && len(v) > 0 {
		config.FieldOptions = expandPivotTableFieldOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandPivotTableFieldWells(v)
	}
	if v, ok := tfMap["paginated_report_options"].([]interface{}); ok && len(v) > 0 {
		config.PaginatedReportOptions = expandPivotTablePaginatedReportOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandPivotTableSortConfiguration(v)
	}
	if v, ok := tfMap["table_options"].([]interface{}); ok && len(v) > 0 {
		config.TableOptions = expandPivotTableOptions(v)
	}
	if v, ok := tfMap["total_options"].([]interface{}); ok && len(v) > 0 {
		config.TotalOptions = expandPivotTableTotalOptions(v)
	}

	return config
}

func expandPivotTableFieldWells(tfList []interface{}) *quicksight.PivotTableFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PivotTableFieldWells{}

	if v, ok := tfMap["pivot_table_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.PivotTableAggregatedFieldWells = expandPivotTableAggregatedFieldWells(v)
	}

	return config
}

func expandPivotTableAggregatedFieldWells(tfList []interface{}) *quicksight.PivotTableAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PivotTableAggregatedFieldWells{}

	if v, ok := tfMap["columns"].([]interface{}); ok && len(v) > 0 {
		config.Columns = expandDimensionFields(v)
	}
	if v, ok := tfMap["rows"].([]interface{}); ok && len(v) > 0 {
		config.Rows = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandPivotTableSortConfiguration(tfList []interface{}) *quicksight.PivotTableSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PivotTableSortConfiguration{}

	if v, ok := tfMap["field_sort_options"].([]interface{}); ok && len(v) > 0 {
		config.FieldSortOptions = expandPivotFieldSortOptionsList(v)
	}

	return config
}

func expandPivotFieldSortOptionsList(tfList []interface{}) []*quicksight.PivotFieldSortOptions {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.PivotFieldSortOptions
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandPivotFieldSortOptions(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandPivotFieldSortOptions(tfMap map[string]interface{}) *quicksight.PivotFieldSortOptions {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.PivotFieldSortOptions{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["sort_by"].([]interface{}); ok && len(v) > 0 {
		options.SortBy = expandPivotTableSortBy(v)
	}

	return options
}

func expandPivotTableSortBy(tfList []interface{}) *quicksight.PivotTableSortBy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PivotTableSortBy{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		config.Column = expandColumnSort(v)
	}
	if v, ok := tfMap["data_path"].([]interface{}); ok && len(v) > 0 {
		config.DataPath = expandDataPathSort(v)
	}
	if v, ok := tfMap[names.AttrField].([]interface{}); ok && len(v) > 0 {
		config.Field = expandFieldSort(v)
	}

	return config
}

func expandDataPathSort(tfList []interface{}) *quicksight.DataPathSort {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DataPathSort{}

	if v, ok := tfMap["direction"].(string); ok && v != "" {
		config.Direction = aws.String(v)
	}
	if v, ok := tfMap["sort_paths"].([]interface{}); ok && len(v) > 0 {
		config.SortPaths = expandDataPathValues(v)
	}

	return config
}

func expandPivotTableFieldOptions(tfList []interface{}) *quicksight.PivotTableFieldOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PivotTableFieldOptions{}

	if v, ok := tfMap["data_path_options"].([]interface{}); ok && len(v) > 0 {
		options.DataPathOptions = expandPivotTableDataPathOptions(v)
	}
	if v, ok := tfMap["selected_field_options"].([]interface{}); ok && len(v) > 0 {
		options.SelectedFieldOptions = expandPivotTableFieldOptionsList(v)
	}

	return options
}

func expandPivotTableDataPathOptions(tfList []interface{}) []*quicksight.PivotTableDataPathOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.PivotTableDataPathOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandPivotTableDataPathOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandPivotTableDataPathOption(tfMap map[string]interface{}) *quicksight.PivotTableDataPathOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.PivotTableDataPathOption{}

	if v, ok := tfMap["width"].(string); ok && v != "" {
		options.Width = aws.String(v)
	}
	if v, ok := tfMap["data_path_list"].([]interface{}); ok && len(v) > 0 {
		options.DataPathList = expandDataPathValues(v)
	}

	return options
}

func expandPivotTableFieldOptionsList(tfList []interface{}) []*quicksight.PivotTableFieldOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.PivotTableFieldOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandPivotTableFieldOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandPivotTableFieldOption(tfMap map[string]interface{}) *quicksight.PivotTableFieldOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.PivotTableFieldOption{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		options.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}

	return options
}

func expandPivotTablePaginatedReportOptions(tfList []interface{}) *quicksight.PivotTablePaginatedReportOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PivotTablePaginatedReportOptions{}

	if v, ok := tfMap["overflow_column_header_visibility"].(string); ok && v != "" {
		options.OverflowColumnHeaderVisibility = aws.String(v)
	}
	if v, ok := tfMap["vertical_overflow_visibility"].(string); ok && v != "" {
		options.VerticalOverflowVisibility = aws.String(v)
	}

	return options
}

func expandPivotTableOptions(tfList []interface{}) *quicksight.PivotTableOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PivotTableOptions{}

	if v, ok := tfMap["collapsed_row_dimensions_visibility"].(string); ok && v != "" {
		options.CollapsedRowDimensionsVisibility = aws.String(v)
	}
	if v, ok := tfMap["column_names_visibility"].(string); ok && v != "" {
		options.ColumnNamesVisibility = aws.String(v)
	}
	if v, ok := tfMap["metric_placement"].(string); ok && v != "" {
		options.MetricPlacement = aws.String(v)
	}
	if v, ok := tfMap["single_metric_visibility"].(string); ok && v != "" {
		options.SingleMetricVisibility = aws.String(v)
	}
	if v, ok := tfMap["toggle_buttons_visibility"].(string); ok && v != "" {
		options.ToggleButtonsVisibility = aws.String(v)
	}
	if v, ok := tfMap["cell_style"].([]interface{}); ok && len(v) > 0 {
		options.CellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["column_header_style"].([]interface{}); ok && len(v) > 0 {
		options.ColumnHeaderStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["row_alternate_color_options"].([]interface{}); ok && len(v) > 0 {
		options.RowAlternateColorOptions = expandRowAlternateColorOptions(v)
	}
	if v, ok := tfMap["row_field_names_style"].([]interface{}); ok && len(v) > 0 {
		options.RowFieldNamesStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["row_header_style"].([]interface{}); ok && len(v) > 0 {
		options.RowHeaderStyle = expandTableCellStyle(v)
	}

	return options
}

func expandTableCellStyle(tfList []interface{}) *quicksight.TableCellStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	style := &quicksight.TableCellStyle{}

	if v, ok := tfMap["background_color"].(string); ok && v != "" {
		style.BackgroundColor = aws.String(v)
	}
	if v, ok := tfMap["height"].(int); ok {
		style.Height = aws.Int64(int64(v))
	}
	if v, ok := tfMap["horizontal_text_alignment"].(string); ok && v != "" {
		style.HorizontalTextAlignment = aws.String(v)
	}
	if v, ok := tfMap["text_wrap"].(string); ok && v != "" {
		style.TextWrap = aws.String(v)
	}
	if v, ok := tfMap["vertical_text_alignment"].(string); ok && v != "" {
		style.VerticalTextAlignment = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		style.Visibility = aws.String(v)
	}
	if v, ok := tfMap["border"].([]interface{}); ok && len(v) > 0 {
		style.Border = expandGlobalTableBorderOptions(v)
	}
	if v, ok := tfMap["font_configuration"].([]interface{}); ok && len(v) > 0 {
		style.FontConfiguration = expandFontConfiguration(v)
	}

	return style
}

func expandGlobalTableBorderOptions(tfList []interface{}) *quicksight.GlobalTableBorderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.GlobalTableBorderOptions{}

	if v, ok := tfMap["side_specific_border"].([]interface{}); ok && len(v) > 0 {
		options.SideSpecificBorder = expandTableSideBorderOptions(v)
	}
	if v, ok := tfMap["uniform_border"].([]interface{}); ok && len(v) > 0 {
		options.UniformBorder = expandTableBorderOptions(v)
	}

	return options
}

func expandTableSideBorderOptions(tfList []interface{}) *quicksight.TableSideBorderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableSideBorderOptions{}

	if v, ok := tfMap["bottom"].([]interface{}); ok && len(v) > 0 {
		options.Bottom = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["inner_horizontal"].([]interface{}); ok && len(v) > 0 {
		options.InnerHorizontal = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["inner_vertical"].([]interface{}); ok && len(v) > 0 {
		options.InnerVertical = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["left"].([]interface{}); ok && len(v) > 0 {
		options.Left = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["right"].([]interface{}); ok && len(v) > 0 {
		options.Right = expandTableBorderOptions(v)
	}
	if v, ok := tfMap["top"].([]interface{}); ok && len(v) > 0 {
		options.Top = expandTableBorderOptions(v)
	}

	return options
}

func expandTableBorderOptions(tfList []interface{}) *quicksight.TableBorderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TableBorderOptions{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		options.Color = aws.String(v)
	}
	if v, ok := tfMap["style"].(string); ok && v != "" {
		options.Style = aws.String(v)
	}
	if v, ok := tfMap["thickness"].(int); ok {
		options.Thickness = aws.Int64(int64(v))
	}

	return options
}

func expandPivotTableTotalOptions(tfList []interface{}) *quicksight.PivotTableTotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PivotTableTotalOptions{}

	if v, ok := tfMap["column_subtotal_options"].([]interface{}); ok && len(v) > 0 {
		options.ColumnSubtotalOptions = expandSubtotalOptions(v)
	}
	if v, ok := tfMap["column_total_options"].([]interface{}); ok && len(v) > 0 {
		options.ColumnTotalOptions = expandPivotTotalOptions(v)
	}
	if v, ok := tfMap["row_subtotal_options"].([]interface{}); ok && len(v) > 0 {
		options.RowSubtotalOptions = expandSubtotalOptions(v)
	}
	if v, ok := tfMap["row_total_options"].([]interface{}); ok && len(v) > 0 {
		options.RowTotalOptions = expandPivotTotalOptions(v)
	}

	return options
}

func expandSubtotalOptions(tfList []interface{}) *quicksight.SubtotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.SubtotalOptions{}

	if v, ok := tfMap["custom_label"].(string); ok && v != "" {
		options.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["field_level"].(string); ok && v != "" {
		options.FieldLevel = aws.String(v)
	}
	if v, ok := tfMap["totals_visibility"].(string); ok && v != "" {
		options.TotalsVisibility = aws.String(v)
	}
	if v, ok := tfMap["field_level_options"].([]interface{}); ok && len(v) > 0 {
		options.FieldLevelOptions = expandPivotTableFieldSubtotalOptionsList(v)
	}
	if v, ok := tfMap["metric_header_cell_style"].([]interface{}); ok && len(v) > 0 {
		options.MetricHeaderCellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["total_cell_style"].([]interface{}); ok && len(v) > 0 {
		options.TotalCellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["value_cell_style"].([]interface{}); ok && len(v) > 0 {
		options.ValueCellStyle = expandTableCellStyle(v)
	}

	return options
}

func expandPivotTableFieldSubtotalOptionsList(tfList []interface{}) []*quicksight.PivotTableFieldSubtotalOptions {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.PivotTableFieldSubtotalOptions
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandPivotTableFieldSubtotalOptions(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandPivotTableFieldSubtotalOptions(tfMap map[string]interface{}) *quicksight.PivotTableFieldSubtotalOptions {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.PivotTableFieldSubtotalOptions{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	return options
}

func expandPivotTotalOptions(tfList []interface{}) *quicksight.PivotTotalOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PivotTotalOptions{}

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
	if v, ok := tfMap["metric_header_cell_style"].([]interface{}); ok && len(v) > 0 {
		options.MetricHeaderCellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["total_cell_style"].([]interface{}); ok && len(v) > 0 {
		options.TotalCellStyle = expandTableCellStyle(v)
	}
	if v, ok := tfMap["value_cell_style"].([]interface{}); ok && len(v) > 0 {
		options.ValueCellStyle = expandTableCellStyle(v)
	}

	return options
}

func expandRowAlternateColorOptions(tfList []interface{}) *quicksight.RowAlternateColorOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.RowAlternateColorOptions{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		options.Status = aws.String(v)
	}
	if v, ok := tfMap["row_alternate_colors"].([]interface{}); ok && len(v) > 0 {
		options.RowAlternateColors = flex.ExpandStringList(v)
	}

	return options
}

func expandPivotTableConditionalFormatting(tfList []interface{}) *quicksight.PivotTableConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PivotTableConditionalFormatting{}

	if v, ok := tfMap["conditional_formatting_options"].([]interface{}); ok && len(v) > 0 {
		options.ConditionalFormattingOptions = expandPivotTableConditionalFormattingOptions(v)
	}

	return options
}

func expandPivotTableConditionalFormattingOptions(tfList []interface{}) []*quicksight.PivotTableConditionalFormattingOption {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.PivotTableConditionalFormattingOption
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandPivotTableConditionalFormattingOption(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandPivotTableConditionalFormattingOption(tfMap map[string]interface{}) *quicksight.PivotTableConditionalFormattingOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.PivotTableConditionalFormattingOption{}

	if v, ok := tfMap["cell"].([]interface{}); ok && len(v) > 0 {
		options.Cell = expandPivotTableCellConditionalFormatting(v)
	}

	return options
}

func expandPivotTableCellConditionalFormatting(tfList []interface{}) *quicksight.PivotTableCellConditionalFormatting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PivotTableCellConditionalFormatting{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrScope].([]interface{}); ok && len(v) > 0 {
		options.Scope = expandPivotTableConditionalFormattingScope(v)
	}
	if v, ok := tfMap["text_format"].([]interface{}); ok && len(v) > 0 {
		options.TextFormat = expandTextConditionalFormat(v)
	}

	return options
}

func expandPivotTableConditionalFormattingScope(tfList []interface{}) *quicksight.PivotTableConditionalFormattingScope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.PivotTableConditionalFormattingScope{}

	if v, ok := tfMap[names.AttrRole].(string); ok && v != "" {
		options.Role = aws.String(v)
	}

	return options
}

func flattenPivotTableVisual(apiObject *quicksight.PivotTableVisual) []interface{} {
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

	return []interface{}{tfMap}
}

func flattenPivotTableConfiguration(apiObject *quicksight.PivotTableConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenPivotTableFieldOptions(apiObject *quicksight.PivotTableFieldOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DataPathOptions != nil {
		tfMap["data_path_options"] = flattenPivotTableDataPathOption(apiObject.DataPathOptions)
	}
	if apiObject.SelectedFieldOptions != nil {
		tfMap["selected_field_options"] = flattenPivotTableFieldOption(apiObject.SelectedFieldOptions)
	}

	return []interface{}{tfMap}
}

func flattenPivotTableDataPathOption(apiObject []*quicksight.PivotTableDataPathOption) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.DataPathList != nil {
			tfMap["data_path_list"] = flattenDataPathValues(config.DataPathList)
		}
		if config.Width != nil {
			tfMap["width"] = aws.StringValue(config.Width)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTableFieldWells(apiObject *quicksight.PivotTableFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PivotTableAggregatedFieldWells != nil {
		tfMap["pivot_table_aggregated_field_wells"] = flattenPivotTableAggregatedFieldWells(apiObject.PivotTableAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenPivotTableAggregatedFieldWells(apiObject *quicksight.PivotTableAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Columns != nil {
		tfMap["columns"] = flattenDimensionFields(apiObject.Columns)
	}
	if apiObject.Rows != nil {
		tfMap["rows"] = flattenDimensionFields(apiObject.Rows)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenPivotTablePaginatedReportOptions(apiObject *quicksight.PivotTablePaginatedReportOptions) []interface{} {
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

func flattenPivotTableSortConfiguration(apiObject *quicksight.PivotTableSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldSortOptions != nil {
		tfMap["field_sort_options"] = flattenPivotFieldSortOptions(apiObject.FieldSortOptions)
	}

	return []interface{}{tfMap}
}

func flattenPivotFieldSortOptions(apiObject []*quicksight.PivotFieldSortOptions) []interface{} {
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
		if config.SortBy != nil {
			tfMap["sort_by"] = flattenPivotTableSortBy(config.SortBy)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTableSortBy(apiObject *quicksight.PivotTableSortBy) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnSort(apiObject.Column)
	}
	if apiObject.DataPath != nil {
		tfMap["data_path"] = flattenDataPathSort(apiObject.DataPath)
	}
	if apiObject.Field != nil {
		tfMap[names.AttrField] = flattenFieldSort(apiObject.Field)
	}

	return []interface{}{tfMap}
}

func flattenDataPathSort(apiObject *quicksight.DataPathSort) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Direction != nil {
		tfMap["direction"] = aws.StringValue(apiObject.Direction)
	}
	if apiObject.SortPaths != nil {
		tfMap["sort_paths"] = flattenDataPathValues(apiObject.SortPaths)
	}

	return []interface{}{tfMap}
}

func flattenPivotTableOptions(apiObject *quicksight.PivotTableOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CellStyle != nil {
		tfMap["cell_style"] = flattenTableCellStyle(apiObject.CellStyle)
	}
	if apiObject.CollapsedRowDimensionsVisibility != nil {
		tfMap["collapsed_row_dimensions_visibility"] = aws.StringValue(apiObject.CollapsedRowDimensionsVisibility)
	}
	if apiObject.ColumnHeaderStyle != nil {
		tfMap["column_header_style"] = flattenTableCellStyle(apiObject.ColumnHeaderStyle)
	}
	if apiObject.ColumnNamesVisibility != nil {
		tfMap["column_names_visibility"] = aws.StringValue(apiObject.ColumnNamesVisibility)
	}
	if apiObject.MetricPlacement != nil {
		tfMap["metric_placement"] = aws.StringValue(apiObject.MetricPlacement)
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
	if apiObject.SingleMetricVisibility != nil {
		tfMap["single_metric_visibility"] = aws.StringValue(apiObject.SingleMetricVisibility)
	}
	if apiObject.ToggleButtonsVisibility != nil {
		tfMap["toggle_buttons_visibility"] = aws.StringValue(apiObject.ToggleButtonsVisibility)
	}

	return []interface{}{tfMap}
}

func flattenTableCellStyle(apiObject *quicksight.TableCellStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = aws.StringValue(apiObject.BackgroundColor)
	}
	if apiObject.Border != nil {
		tfMap["border"] = flattenGlobalTableBorderOptions(apiObject.Border)
	}
	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	if apiObject.Height != nil {
		tfMap["height"] = aws.Int64Value(apiObject.Height)
	}
	if apiObject.HorizontalTextAlignment != nil {
		tfMap["horizontal_text_alignment"] = aws.StringValue(apiObject.HorizontalTextAlignment)
	}
	if apiObject.TextWrap != nil {
		tfMap["text_wrap"] = aws.StringValue(apiObject.TextWrap)
	}
	if apiObject.VerticalTextAlignment != nil {
		tfMap["vertical_text_alignment"] = aws.StringValue(apiObject.VerticalTextAlignment)
	}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenGlobalTableBorderOptions(apiObject *quicksight.GlobalTableBorderOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SideSpecificBorder != nil {
		tfMap["side_specific_border"] = flattenTableSideBorderOptions(apiObject.SideSpecificBorder)
	}
	if apiObject.UniformBorder != nil {
		tfMap["uniform_border"] = flattenTableBorderOptions(apiObject.UniformBorder)
	}

	return []interface{}{tfMap}
}

func flattenTableSideBorderOptions(apiObject *quicksight.TableSideBorderOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenTableBorderOptions(apiObject *quicksight.TableBorderOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.StringValue(apiObject.Color)
	}
	if apiObject.Style != nil {
		tfMap["style"] = aws.StringValue(apiObject.Style)
	}
	if apiObject.Thickness != nil {
		tfMap["thickness"] = aws.Int64Value(apiObject.Thickness)
	}

	return []interface{}{tfMap}
}

func flattenRowAlternateColorOptions(apiObject *quicksight.RowAlternateColorOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.RowAlternateColors != nil {
		tfMap["row_alternate_colors"] = flex.FlattenStringList(apiObject.RowAlternateColors)
	}
	if apiObject.Status != nil {
		tfMap[names.AttrStatus] = aws.StringValue(apiObject.Status)
	}

	return []interface{}{tfMap}
}

func flattenPivotTableTotalOptions(apiObject *quicksight.PivotTableTotalOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenSubtotalOptions(apiObject *quicksight.SubtotalOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.StringValue(apiObject.CustomLabel)
	}
	if apiObject.FieldLevel != nil {
		tfMap["field_level"] = aws.StringValue(apiObject.FieldLevel)
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
	if apiObject.TotalsVisibility != nil {
		tfMap["totals_visibility"] = aws.StringValue(apiObject.TotalsVisibility)
	}
	if apiObject.ValueCellStyle != nil {
		tfMap["value_cell_style"] = flattenTableCellStyle(apiObject.ValueCellStyle)
	}

	return []interface{}{tfMap}
}

func flattenPivotTableFieldSubtotalOptions(apiObject []*quicksight.PivotTableFieldSubtotalOptions) []interface{} {
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

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTotalOptions(apiObject *quicksight.PivotTotalOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.StringValue(apiObject.CustomLabel)
	}
	if apiObject.MetricHeaderCellStyle != nil {
		tfMap["metric_header_cell_style"] = flattenTableCellStyle(apiObject.MetricHeaderCellStyle)
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
	if apiObject.ValueCellStyle != nil {
		tfMap["value_cell_style"] = flattenTableCellStyle(apiObject.ValueCellStyle)
	}

	return []interface{}{tfMap}
}

func flattenPivotTableFieldOption(apiObject []*quicksight.PivotTableFieldOption) []interface{} {
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
		if config.Visibility != nil {
			tfMap["visibility"] = aws.StringValue(config.Visibility)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTableConditionalFormatting(apiObject *quicksight.PivotTableConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ConditionalFormattingOptions != nil {
		tfMap["conditional_formatting_options"] = flattenPivotTableConditionalFormattingOption(apiObject.ConditionalFormattingOptions)
	}

	return []interface{}{tfMap}
}

func flattenPivotTableConditionalFormattingOption(apiObject []*quicksight.PivotTableConditionalFormattingOption) []interface{} {
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
			tfMap["cell"] = flattenPivotTableCellConditionalFormatting(config.Cell)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPivotTableCellConditionalFormatting(apiObject *quicksight.PivotTableCellConditionalFormatting) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"field_id": aws.StringValue(apiObject.FieldId),
	}
	if apiObject.Scope != nil {
		tfMap[names.AttrScope] = flattenPivotTableConditionalFormattingScope(apiObject.Scope)
	}
	if apiObject.TextFormat != nil {
		tfMap["text_format"] = flattenTextConditionalFormat(apiObject.TextFormat)
	}

	return []interface{}{tfMap}
}

func flattenPivotTableConditionalFormattingScope(apiObject *quicksight.PivotTableConditionalFormattingScope) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Role != nil {
		tfMap[names.AttrRole] = aws.StringValue(apiObject.Role)
	}

	return []interface{}{tfMap}
}

func flattenTextConditionalFormat(apiObject *quicksight.TextConditionalFormat) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BackgroundColor != nil {
		tfMap["background_color"] = flattenConditionalFormattingColor(apiObject.BackgroundColor)
	}
	if apiObject.Icon != nil {
		tfMap["icon"] = flattenConditionalFormattingIcon(apiObject.Icon)
	}
	if apiObject.TextColor != nil {
		tfMap["text_color"] = flattenConditionalFormattingColor(apiObject.TextColor)
	}

	return []interface{}{tfMap}
}
