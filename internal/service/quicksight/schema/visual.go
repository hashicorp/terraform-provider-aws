// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const customActionsMaxItems = 10
const referenceLinesMaxItems = 20
const dataPathValueMaxItems = 20

func visualsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetControlLayout.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 50,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"bar_chart_visual":      barCharVisualSchema(),
				"box_plot_visual":       boxPlotVisualSchema(),
				"combo_chart_visual":    comboChartVisualSchema(),
				"custom_content_visual": customContentVisualSchema(),
				"empty_visual":          emptyVisualSchema(),
				"filled_map_visual":     filledMapVisualSchema(),
				"funnel_chart_visual":   funnelChartVisualSchema(),
				"gauge_chart_visual":    gaugeChartVisualSchema(),
				"geospatial_map_visual": geospatialMapVisualSchema(),
				"heat_map_visual":       heatMapVisualSchema(),
				"histogram_visual":      histogramVisualSchema(),
				"insight_visual":        insightVisualSchema(),
				"kpi_visual":            kpiVisualSchema(),
				"line_chart_visual":     lineChartVisualSchema(),
				"pie_chart_visual":      pieChartVisualSchema(),
				"pivot_table_visual":    pivotTableVisualSchema(),
				"radar_chart_visual":    radarChartVisualSchema(),
				"sankey_diagram_visual": sankeyDiagramVisualSchema(),
				"scatter_plot_visual":   scatterPlotVisualSchema(),
				"table_visual":          tableVisualSchema(),
				"tree_map_visual":       treeMapVisualSchema(),
				"waterfall_visual":      waterfallVisualSchema(),
				"word_cloud_visual":     wordCloudVisualSchema(),
			},
		},
	}
}

func legendOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"height": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"position":   stringSchema(false, enum.Validate[types.LegendPosition]()),
				"title":      labelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
				"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
				"width": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func tooltipOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_base_tooltip": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldBasedTooltip.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"aggregation_visibility": stringSchema(false, enum.Validate[types.Visibility]()),
							"tooltip_fields": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipItem.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 100,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column_tooltip_item": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnTooltipItem.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"column":      columnSchema(true),               // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
													"aggregation": aggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
													"label": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
												},
											},
										},
										"field_tooltip_item": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldTooltipItem.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"field_id": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
													"label": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
												},
											},
										},
									},
								},
							},
							"tooltip_title_type": stringSchema(false, enum.Validate[types.TooltipTitleType]()),
						},
					},
				},
				"selected_tooltip_type": stringSchema(false, enum.Validate[types.SelectedTooltipType]()),
				"tooltip_visibility":    stringSchema(false, enum.Validate[types.Visibility]()),
			},
		},
	}
}

func visualPaletteSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"chart_color": stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
				"color_map": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathColor.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 5000,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":            stringSchema(true, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
							"element":          dataPathValueSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathValue.html
							"time_granularity": stringSchema(false, enum.Validate[types.TimeGranularity]()),
						},
					},
				},
			},
		},
	}
}

func dataPathValueSchema(maxItems int) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathValue.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_id":    stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
				"field_value": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
			},
		},
	}
}

func columnHierarchiesSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnHierarchy.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 2,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_hierarchy": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeHierarchy.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"hierarchy_id":       stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
							"drill_down_filters": drillDownFilterSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DrillDownFilter.html
						},
					},
				},
				"explicit_hierarchy": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExplicitHierarchy.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"columns": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 2,
								MaxItems: 10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
										"column_name":         stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 128))),
										"data_set_identifier": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
									},
								},
							},
							"hierarchy_id":       stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
							"drill_down_filters": drillDownFilterSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DrillDownFilter.html
						},
					},
				},
				"predefined_hierarchy": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PredefinedHierarchy.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"columns": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
										"column_name":         stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 128))),
										"data_set_identifier": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
									},
								},
							},
							"hierarchy_id":       stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
							"drill_down_filters": drillDownFilterSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DrillDownFilter.html
						},
					},
				},
			},
		},
	}
}

func visualSubtitleLabelOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"format_text": longFormatTextSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LongFormatText.html
				"visibility":  stringOptionalComputedSchema(enum.Validate[types.Visibility]()),
			},
		},
	}
}

func longFormatTextSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LongFormatText.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"plain_text": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
				},
				"rich_text": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
				},
			},
		},
	}
}

func shortFormatTextSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ShortFormatText.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"plain_text": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 512)),
				},
				"rich_text": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
				},
			},
		},
	}
}

func visualTitleLabelOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"format_text": shortFormatTextSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ShortFormatText.html
				"visibility":  stringOptionalComputedSchema(enum.Validate[types.Visibility]()),
			},
		},
	}
}

func comparisonConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComparisonConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison_format": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComparisonFormatConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"number_display_format_configuration":     numberDisplayFormatConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberDisplayFormatConfiguration.html
							"percentage_display_format_configuration": percentageDisplayFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PercentageDisplayFormatConfiguration.html
						},
					},
				},
				"comparison_method": stringSchema(false, enum.Validate[types.ComparisonMethod]()),
			},
		},
	}
}

func colorScaleSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColorScale.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"color_fill_type": stringSchema(false, enum.Validate[types.ColorFillType]()),
				"colors": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataColor.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 2,
					MaxItems: 3,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color": stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
							"data_value": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
						},
					},
				},
				"null_value_color": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataColor.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color": stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
							"data_value": {
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

func dataLabelOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"category_label_visibility": stringSchema(false, enum.Validate[types.Visibility]()),
				"data_label_types": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelType.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 100,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_path_label_type": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathLabelType.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"field_id":    stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
										"field_value": stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
										"visibility":  stringSchema(false, enum.Validate[types.Visibility]()),
									},
								},
							},
							"field_label_type": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldLabelType.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"field_id":   stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 512))),
										"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
									},
								},
							},
							"maximum_label_type": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MaximumLabelType.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
									},
								},
							},
							"minimum_label_type": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MinimumLabelType.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
									},
								},
							},
							"range_ends_label_type": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RangeEndsLabelType.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
									},
								},
							},
						},
					},
				},
				"label_color":              stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}$`), ""))),
				"label_content":            stringSchema(false, enum.Validate[types.DataLabelContent]()),
				"label_font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
				"measure_label_visibility": stringSchema(false, enum.Validate[types.Visibility]()),
				"overlap":                  stringSchema(false, enum.Validate[types.DataLabelOverlap]()),
				"position":                 stringSchema(false, enum.Validate[types.DataLabelPosition]()),
				"visibility":               stringSchema(false, enum.Validate[types.Visibility]()),
			},
		},
	}
}

func expandVisual(tfMap map[string]interface{}) *types.Visual {
	if tfMap == nil {
		return nil
	}

	visual := &types.Visual{}

	if v, ok := tfMap["bar_chart_visual"].([]interface{}); ok && len(v) > 0 {
		visual.BarChartVisual = expandBarChartVisual(v)
	}
	if v, ok := tfMap["box_plot_visual"].([]interface{}); ok && len(v) > 0 {
		visual.BoxPlotVisual = expandBoxPlotVisual(v)
	}
	if v, ok := tfMap["combo_chart_visual"].([]interface{}); ok && len(v) > 0 {
		visual.ComboChartVisual = expandComboChartVisual(v)
	}
	if v, ok := tfMap["custom_content_visual"].([]interface{}); ok && len(v) > 0 {
		visual.CustomContentVisual = expandCustomContentVisual(v)
	}
	if v, ok := tfMap["empty_visual"].([]interface{}); ok && len(v) > 0 {
		visual.EmptyVisual = expandEmptyVisual(v)
	}
	if v, ok := tfMap["filled_map_visual"].([]interface{}); ok && len(v) > 0 {
		visual.FilledMapVisual = expandFilledMapVisual(v)
	}
	if v, ok := tfMap["funnel_chart_visual"].([]interface{}); ok && len(v) > 0 {
		visual.FunnelChartVisual = expandFunnelChartVisual(v)
	}
	if v, ok := tfMap["gauge_chart_visual"].([]interface{}); ok && len(v) > 0 {
		visual.GaugeChartVisual = expandGaugeChartVisual(v)
	}
	if v, ok := tfMap["geospatial_map_visual"].([]interface{}); ok && len(v) > 0 {
		visual.GeospatialMapVisual = expandGeospatialMapVisual(v)
	}
	if v, ok := tfMap["heat_map_visual"].([]interface{}); ok && len(v) > 0 {
		visual.HeatMapVisual = expandHeatMapVisual(v)
	}
	if v, ok := tfMap["histogram_visual"].([]interface{}); ok && len(v) > 0 {
		visual.HistogramVisual = expandHistogramVisual(v)
	}
	if v, ok := tfMap["insight_visual"].([]interface{}); ok && len(v) > 0 {
		visual.InsightVisual = expandInsightVisual(v)
	}
	if v, ok := tfMap["kpi_visual"].([]interface{}); ok && len(v) > 0 {
		visual.KPIVisual = expandKPIVisual(v)
	}
	if v, ok := tfMap["line_chart_visual"].([]interface{}); ok && len(v) > 0 {
		visual.LineChartVisual = expandLineChartVisual(v)
	}
	if v, ok := tfMap["pie_chart_visual"].([]interface{}); ok && len(v) > 0 {
		visual.PieChartVisual = expandPieChartVisual(v)
	}
	if v, ok := tfMap["pivot_table_visual"].([]interface{}); ok && len(v) > 0 {
		visual.PivotTableVisual = expandPivotTableVisual(v)
	}
	if v, ok := tfMap["radar_chart_visual"].([]interface{}); ok && len(v) > 0 {
		visual.RadarChartVisual = expandRadarChartVisual(v)
	}
	if v, ok := tfMap["sankey_diagram_visual"].([]interface{}); ok && len(v) > 0 {
		visual.SankeyDiagramVisual = expandSankeyDiagramVisual(v)
	}
	if v, ok := tfMap["scatter_plot_visual"].([]interface{}); ok && len(v) > 0 {
		visual.ScatterPlotVisual = expandScatterPlotVisual(v)
	}
	if v, ok := tfMap["table_visual"].([]interface{}); ok && len(v) > 0 {
		visual.TableVisual = expandTableVisual(v)
	}
	if v, ok := tfMap["tree_map_visual"].([]interface{}); ok && len(v) > 0 {
		visual.TreeMapVisual = expandTreeMapVisual(v)
	}
	if v, ok := tfMap["waterfall_visual"].([]interface{}); ok && len(v) > 0 {
		visual.WaterfallVisual = expandWaterfallVisual(v)
	}
	if v, ok := tfMap["word_cloud_visual"].([]interface{}); ok && len(v) > 0 {
		visual.WordCloudVisual = expandWordCloudVisual(v)
	}

	return visual
}

func expandDataLabelOptions(tfList []interface{}) *types.DataLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.DataLabelOptions{}

	if v, ok := tfMap["category_label_visibility"].(string); ok && v != "" {
		options.CategoryLabelVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["label_color"].(string); ok && v != "" {
		options.LabelColor = aws.String(v)
	}
	if v, ok := tfMap["label_content"].(string); ok && v != "" {
		options.LabelContent = types.DataLabelContent(v)
	}
	if v, ok := tfMap["measure_label_visibility"].(string); ok && v != "" {
		options.MeasureLabelVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["overlap"].(string); ok && v != "" {
		options.Overlap = types.DataLabelOverlap(v)
	}
	if v, ok := tfMap["position"].(string); ok && v != "" {
		options.Position = types.DataLabelPosition(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}

	if v, ok := tfMap["data_label_types"].([]interface{}); ok && len(v) > 0 {
		options.DataLabelTypes = expandDataLabelTypes(v)
	}
	if v, ok := tfMap["label_font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.LabelFontConfiguration = expandFontConfiguration(v)
	}

	return options
}

func expandDataLabelTypes(tfList []interface{}) []types.DataLabelType {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.DataLabelType
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandDataLabelType(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandDataLabelType(tfMap map[string]interface{}) *types.DataLabelType {
	if tfMap == nil {
		return nil
	}

	options := &types.DataLabelType{}

	if v, ok := tfMap["data_path_label_type"].([]interface{}); ok && len(v) > 0 {
		options.DataPathLabelType = expandDataPathLabelType(v)
	}
	if v, ok := tfMap["field_label_type"].([]interface{}); ok && len(v) > 0 {
		options.FieldLabelType = expandFieldLabelType(v)
	}
	if v, ok := tfMap["maximum_label_type"].([]interface{}); ok && len(v) > 0 {
		options.MaximumLabelType = expandMaximumLabelType(v)
	}
	if v, ok := tfMap["minimum_label_type"].([]interface{}); ok && len(v) > 0 {
		options.MinimumLabelType = expandMinimumLabelType(v)
	}
	if v, ok := tfMap["range_ends_label_type"].([]interface{}); ok && len(v) > 0 {
		options.RangeEndsLabelType = expandRangeEndsLabelType(v)
	}

	return options
}

func expandDataPathLabelType(tfList []interface{}) *types.DataPathLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.DataPathLabelType{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["field_value"].(string); ok && v != "" {
		options.FieldValue = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}

	return options
}

func expandFieldLabelType(tfList []interface{}) *types.FieldLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.FieldLabelType{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		options.FieldId = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}

	return options
}

func expandMaximumLabelType(tfList []interface{}) *types.MaximumLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.MaximumLabelType{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}

	return options
}

func expandMinimumLabelType(tfList []interface{}) *types.MinimumLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.MinimumLabelType{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}

	return options
}

func expandRangeEndsLabelType(tfList []interface{}) *types.RangeEndsLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.RangeEndsLabelType{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}

	return options
}

func expandLegendOptions(tfList []interface{}) *types.LegendOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.LegendOptions{}

	if v, ok := tfMap["height"].(string); ok && v != "" {
		options.Height = aws.String(v)
	}
	if v, ok := tfMap["position"].(string); ok && v != "" {
		options.Position = types.LegendPosition(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}
	if v, ok := tfMap["width"].(string); ok && v != "" {
		options.Width = aws.String(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		options.Title = expandLabelOptions(v)
	}

	return options
}

func expandTooltipOptions(tfList []interface{}) *types.TooltipOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.TooltipOptions{}

	if v, ok := tfMap["selected_tooltip_type"].(string); ok && v != "" {
		options.SelectedTooltipType = types.SelectedTooltipType(v)
	}
	if v, ok := tfMap["tooltip_visibility"].(string); ok && v != "" {
		options.TooltipVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["field_base_tooltip"].([]interface{}); ok && len(v) > 0 {
		options.FieldBasedTooltip = expandFieldBasedTooltip(v)
	}

	return options
}

func expandFieldBasedTooltip(tfList []interface{}) *types.FieldBasedTooltip {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.FieldBasedTooltip{}

	if v, ok := tfMap["aggregation_visibility"].(string); ok && v != "" {
		options.AggregationVisibility = types.Visibility(v)
	}
	if v, ok := tfMap["tooltip_title_type"].(string); ok && v != "" {
		options.TooltipTitleType = types.TooltipTitleType(v)
	}
	if v, ok := tfMap["tooltip_fields"].([]interface{}); ok && len(v) > 0 {
		options.TooltipFields = expandTooltipItems(v)
	}

	return options
}

func expandTooltipItems(tfList []interface{}) []types.TooltipItem {
	if len(tfList) == 0 {
		return nil
	}

	var items []types.TooltipItem
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandTooltipItem(tfMap)
		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return items
}

func expandTooltipItem(tfMap map[string]interface{}) *types.TooltipItem {
	if tfMap == nil {
		return nil
	}

	item := &types.TooltipItem{}

	if v, ok := tfMap["column_tooltip_item"].([]interface{}); ok && len(v) > 0 {
		item.ColumnTooltipItem = expandColumnTooltipItem(v)
	}
	if v, ok := tfMap["field_tooltip_item"].([]interface{}); ok && len(v) > 0 {
		item.FieldTooltipItem = expandFieldTooltipItem(v)
	}

	return item
}

func expandColumnTooltipItem(tfList []interface{}) *types.ColumnTooltipItem {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	item := &types.ColumnTooltipItem{}

	if v, ok := tfMap["label"].(string); ok && v != "" {
		item.Label = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		item.Visibility = types.Visibility(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		item.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["aggregation"].([]interface{}); ok && len(v) > 0 {
		item.Aggregation = expandAggregationFunction(v)
	}

	return item
}

func expandFieldTooltipItem(tfList []interface{}) *types.FieldTooltipItem {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	item := &types.FieldTooltipItem{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		item.FieldId = aws.String(v)
	}
	if v, ok := tfMap["label"].(string); ok && v != "" {
		item.Label = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		item.Visibility = types.Visibility(v)
	}

	return item
}

func expandVisualPalette(tfList []interface{}) *types.VisualPalette {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.VisualPalette{}

	if v, ok := tfMap["chart_color"].(string); ok && v != "" {
		config.ChartColor = aws.String(v)
	}
	if v, ok := tfMap["color_map"].([]interface{}); ok && len(v) > 0 {
		config.ColorMap = expandDataPathColors(v)
	}

	return config
}

func expandDataPathColors(tfList []interface{}) []types.DataPathColor {
	if len(tfList) == 0 {
		return nil
	}

	var colors []types.DataPathColor
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		color := expandDataPathColor(tfMap)
		if color == nil {
			continue
		}

		colors = append(colors, *color)
	}

	return colors
}

func expandDataPathColor(tfMap map[string]interface{}) *types.DataPathColor {
	if tfMap == nil {
		return nil
	}

	color := &types.DataPathColor{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		color.Color = aws.String(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		color.TimeGranularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["element"].([]interface{}); ok && len(v) > 0 {
		color.Element = expandDataPathValue(v)
	}

	return color
}

func expandDataPathValues(tfList []interface{}) []types.DataPathValue {
	if len(tfList) == 0 {
		return nil
	}

	var values []types.DataPathValue
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		value := expandDataPathValueInternal(tfMap)
		if value == nil {
			continue
		}

		values = append(values, *value)
	}

	return values
}

func expandDataPathValueInternal(tfMap map[string]interface{}) *types.DataPathValue {
	if tfMap == nil {
		return nil
	}

	value := &types.DataPathValue{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		value.FieldId = aws.String(v)
	}
	if v, ok := tfMap["field_value"].(string); ok && v != "" {
		value.FieldValue = aws.String(v)
	}

	return value
}

func expandDataPathValue(tfList []interface{}) *types.DataPathValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return expandDataPathValueInternal(tfMap)
}

func expandColumnHierarchies(tfList []interface{}) []types.ColumnHierarchy {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.ColumnHierarchy
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandColumnHierarchy(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandColumnHierarchy(tfMap map[string]interface{}) *types.ColumnHierarchy {
	if tfMap == nil {
		return nil
	}

	options := &types.ColumnHierarchy{}

	if v, ok := tfMap["date_time_hierarchy"].([]interface{}); ok && len(v) > 0 {
		options.DateTimeHierarchy = expandDateTimeHierarchy(v)
	}
	if v, ok := tfMap["explicit_hierarchy"].([]interface{}); ok && len(v) > 0 {
		options.ExplicitHierarchy = expandExplicitHierarchy(v)
	}
	if v, ok := tfMap["predefined_hierarchy"].([]interface{}); ok && len(v) > 0 {
		options.PredefinedHierarchy = expandPredefinedHierarchy(v)
	}

	return options
}

func expandDateTimeHierarchy(tfList []interface{}) *types.DateTimeHierarchy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DateTimeHierarchy{}

	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		config.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["drill_down_filters"].([]interface{}); ok && len(v) > 0 {
		config.DrillDownFilters = expandDrillDownFilters(v)
	}

	return config
}

func expandExplicitHierarchy(tfList []interface{}) *types.ExplicitHierarchy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.ExplicitHierarchy{}

	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		config.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["columns"].([]interface{}); ok && len(v) > 0 {
		config.Columns = expandColumnIdentifiers(v)
	}
	if v, ok := tfMap["drill_down_filters"].([]interface{}); ok && len(v) > 0 {
		config.DrillDownFilters = expandDrillDownFilters(v)
	}

	return config
}
func expandPredefinedHierarchy(tfList []interface{}) *types.PredefinedHierarchy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.PredefinedHierarchy{}

	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		config.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["columns"].([]interface{}); ok && len(v) > 0 {
		config.Columns = expandColumnIdentifiers(v)
	}
	if v, ok := tfMap["drill_down_filters"].([]interface{}); ok && len(v) > 0 {
		config.DrillDownFilters = expandDrillDownFilters(v)
	}

	return config
}

func expandVisualSubtitleLabelOptions(tfList []interface{}) *types.VisualSubtitleLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.VisualSubtitleLabelOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}
	if v, ok := tfMap["format_text"].([]interface{}); ok && len(v) > 0 {
		options.FormatText = expandLongFormatText(v)
	}

	return options
}

func expandLongFormatText(tfList []interface{}) *types.LongFormatText {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	format := &types.LongFormatText{}

	if v, ok := tfMap["plain_text"].(string); ok && v != "" {
		format.PlainText = aws.String(v)
	}
	if v, ok := tfMap["rich_text"].(string); ok && v != "" {
		format.RichText = aws.String(v)
	}

	return format
}

func expandVisualTitleLabelOptions(tfList []interface{}) *types.VisualTitleLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.VisualTitleLabelOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = types.Visibility(v)
	}
	if v, ok := tfMap["format_text"].([]interface{}); ok && len(v) > 0 {
		options.FormatText = expandShortFormatText(v)
	}

	return options
}

func expandShortFormatText(tfList []interface{}) *types.ShortFormatText {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	format := &types.ShortFormatText{}

	if v, ok := tfMap["plain_text"].(string); ok && v != "" {
		format.PlainText = aws.String(v)
	}
	if v, ok := tfMap["rich_text"].(string); ok && v != "" {
		format.RichText = aws.String(v)
	}

	return format
}

func expandComparisonConfiguration(tfList []interface{}) *types.ComparisonConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.ComparisonConfiguration{}

	if v, ok := tfMap["comparison_method"].(string); ok && v != "" {
		config.ComparisonMethod = types.ComparisonMethod(v)
	}
	if v, ok := tfMap["comparison_format"].([]interface{}); ok && len(v) > 0 {
		config.ComparisonFormat = expandComparisonFormatConfiguration(v)
	}

	return config
}

func expandColorScale(tfList []interface{}) *types.ColorScale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.ColorScale{}

	if v, ok := tfMap["color_fill_type"].(string); ok && v != "" {
		config.ColorFillType = types.ColorFillType(v)
	}
	if v, ok := tfMap["colors"].([]interface{}); ok && len(v) > 0 {
		config.Colors = expandDataColors(v)
	}
	if v, ok := tfMap["null_value_color"].([]interface{}); ok && len(v) > 0 {
		config.NullValueColor = expandDataColor(v)
	}

	return config
}

func expandDataColor(tfList []interface{}) *types.DataColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return expandDataColorInternal(tfMap)
}

func expandDataColorInternal(tfMap map[string]interface{}) *types.DataColor {
	color := &types.DataColor{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		color.Color = aws.String(v)
	}
	if v, ok := tfMap["data_value"].(float64); ok {
		color.DataValue = aws.Float64(v)
	}

	return color
}

func expandDataColors(tfList []interface{}) []types.DataColor {
	if len(tfList) == 0 {
		return nil
	}

	var colors []types.DataColor
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		color := expandDataColorInternal(tfMap)
		if color == nil {
			continue
		}

		colors = append(colors, *color)
	}

	return colors
}

func flattenVisuals(apiObject []types.Visual) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.BarChartVisual != nil {
			tfMap["bar_chart_visual"] = flattenBarChartVisual(config.BarChartVisual)
		}
		if config.BoxPlotVisual != nil {
			tfMap["box_plot_visual"] = flattenBoxPlotVisual(config.BoxPlotVisual)
		}
		if config.ComboChartVisual != nil {
			tfMap["combo_chart_visual"] = flattenComboChartVisual(config.ComboChartVisual)
		}
		if config.CustomContentVisual != nil {
			tfMap["custom_content_visual"] = flattenCustomContentVisual(config.CustomContentVisual)
		}
		if config.EmptyVisual != nil {
			tfMap["empty_visual"] = flattenEmptyVisual(config.EmptyVisual)
		}
		if config.FilledMapVisual != nil {
			tfMap["filled_map_visual"] = flattenFilledMapVisual(config.FilledMapVisual)
		}
		if config.FunnelChartVisual != nil {
			tfMap["funnel_chart_visual"] = flattenFunnelChartVisual(config.FunnelChartVisual)
		}
		if config.GaugeChartVisual != nil {
			tfMap["gauge_chart_visual"] = flattenGaugeChartVisual(config.GaugeChartVisual)
		}
		if config.GeospatialMapVisual != nil {
			tfMap["geospatial_map_visual"] = flattenGeospatialMapVisual(config.GeospatialMapVisual)
		}
		if config.HeatMapVisual != nil {
			tfMap["heat_map_visual"] = flattenHeatMapVisual(config.HeatMapVisual)
		}
		if config.HistogramVisual != nil {
			tfMap["histogram_visual"] = flattenHistogramVisual(config.HistogramVisual)
		}
		if config.InsightVisual != nil {
			tfMap["insight_visual"] = flattenInsightVisual(config.InsightVisual)
		}
		if config.KPIVisual != nil {
			tfMap["kpi_visual"] = flattenKPIVisual(config.KPIVisual)
		}
		if config.LineChartVisual != nil {
			tfMap["line_chart_visual"] = flattenLineChartVisual(config.LineChartVisual)
		}
		if config.PieChartVisual != nil {
			tfMap["pie_chart_visual"] = flattenPieChartVisual(config.PieChartVisual)
		}
		if config.PivotTableVisual != nil {
			tfMap["pivot_table_visual"] = flattenPivotTableVisual(config.PivotTableVisual)
		}
		if config.RadarChartVisual != nil {
			tfMap["radar_chart_visual"] = flattenRadarChartVisual(config.RadarChartVisual)
		}
		if config.SankeyDiagramVisual != nil {
			tfMap["sankey_diagram_visual"] = flattenSankeyDiagramVisual(config.SankeyDiagramVisual)
		}
		if config.ScatterPlotVisual != nil {
			tfMap["scatter_plot_visual"] = flattenScatterPlotVisual(config.ScatterPlotVisual)
		}
		if config.TableVisual != nil {
			tfMap["table_visual"] = flattenTableVisual(config.TableVisual)
		}
		if config.TreeMapVisual != nil {
			tfMap["tree_map_visual"] = flattenTreeMapVisual(config.TreeMapVisual)
		}
		if config.WaterfallVisual != nil {
			tfMap["waterfall_visual"] = flattenWaterfallVisual(config.WaterfallVisual)
		}
		if config.WordCloudVisual != nil {
			tfMap["word_cloud_visual"] = flattenWordCloudVisual(config.WordCloudVisual)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataLabelOptions(apiObject *types.DataLabelOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["category_label_visibility"] = apiObject.CategoryLabelVisibility
	if apiObject.DataLabelTypes != nil {
		tfMap["data_label_types"] = flattenDataLabelType(apiObject.DataLabelTypes)
	}
	if apiObject.LabelColor != nil {
		tfMap["label_color"] = aws.ToString(apiObject.LabelColor)
	}
	tfMap["label_content"] = types.DataLabelContent(apiObject.LabelContent)
	if apiObject.LabelFontConfiguration != nil {
		tfMap["label_font_configuration"] = flattenFontConfiguration(apiObject.LabelFontConfiguration)
	}
	tfMap["measure_label_visibility"] = types.Visibility(apiObject.MeasureLabelVisibility)
	tfMap["overlap"] = types.DataLabelOverlap(apiObject.Overlap)
	tfMap["position"] = types.DataLabelPosition(apiObject.Position)
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenDataLabelType(apiObject []types.DataLabelType) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.DataPathLabelType != nil {
			tfMap["data_path_label_type"] = flattenDataPathLabelType(config.DataPathLabelType)
		}
		if config.FieldLabelType != nil {
			tfMap["field_label_type"] = flattenFieldLabelType(config.FieldLabelType)
		}
		if config.MaximumLabelType != nil {
			tfMap["maximum_label_type"] = flattenMaximumLabelType(config.MaximumLabelType)
		}
		if config.MinimumLabelType != nil {
			tfMap["minimum_label_type"] = flattenMinimumLabelType(config.MinimumLabelType)
		}
		if config.RangeEndsLabelType != nil {
			tfMap["range_ends_label_type"] = flattenRangeEndsLabelType(config.RangeEndsLabelType)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataPathLabelType(apiObject *types.DataPathLabelType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FieldValue != nil {
		tfMap["field_value"] = aws.ToString(apiObject.FieldValue)
	}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenFieldLabelType(apiObject *types.FieldLabelType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenMaximumLabelType(apiObject *types.MaximumLabelType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenMinimumLabelType(apiObject *types.MinimumLabelType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenRangeEndsLabelType(apiObject *types.RangeEndsLabelType) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenLegendOptions(apiObject *types.LegendOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Height != nil {
		tfMap["height"] = aws.ToString(apiObject.Height)
	}
	tfMap["position"] = types.LegendPosition(apiObject.Position)
	if apiObject.Title != nil {
		tfMap["title"] = flattenLabelOptions(apiObject.Title)
	}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	if apiObject.Width != nil {
		tfMap["width"] = aws.ToString(apiObject.Width)
	}

	return []interface{}{tfMap}
}

func flattenTooltipOptions(apiObject *types.TooltipOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldBasedTooltip != nil {
		tfMap["field_base_tooltip"] = flattenFieldBasedTooltip(apiObject.FieldBasedTooltip)
	}
	tfMap["selected_tooltip_type"] = types.SelectedTooltipType(apiObject.SelectedTooltipType)
	tfMap["tooltip_visibility"] = types.Visibility(apiObject.TooltipVisibility)

	return []interface{}{tfMap}
}

func flattenFieldBasedTooltip(apiObject *types.FieldBasedTooltip) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["aggregation_visibility"] = types.Visibility(apiObject.AggregationVisibility)
	if apiObject.TooltipFields != nil {
		tfMap["tooltip_fields"] = flattenTooltipItem(apiObject.TooltipFields)
	}
	tfMap["tooltip_title_type"] = types.TooltipTitleType(apiObject.TooltipTitleType)

	return []interface{}{tfMap}
}

func flattenTooltipItem(apiObject []types.TooltipItem) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.ColumnTooltipItem != nil {
			tfMap["column_tooltip_item"] = flattenColumnTooltipItem(config.ColumnTooltipItem)
		}
		if config.FieldTooltipItem != nil {
			tfMap["field_tooltip_item"] = flattenFieldTooltipItem(config.FieldTooltipItem)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnTooltipItem(apiObject *types.ColumnTooltipItem) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.Aggregation != nil {
		tfMap["aggregation"] = flattenAggregationFunction(apiObject.Aggregation)
	}
	if apiObject.Label != nil {
		tfMap["label"] = aws.ToString(apiObject.Label)
	}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenFieldTooltipItem(apiObject *types.FieldTooltipItem) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.Label != nil {
		tfMap["label"] = aws.ToString(apiObject.Label)
	}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenVisualPalette(apiObject *types.VisualPalette) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ChartColor != nil {
		tfMap["chart_color"] = aws.ToString(apiObject.ChartColor)
	}
	if apiObject.ColorMap != nil {
		tfMap["color_map"] = flattenDataPathColor(apiObject.ColorMap)
	}

	return []interface{}{tfMap}
}

func flattenDataPathColor(apiObject []types.DataPathColor) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.Color != nil {
			tfMap["color"] = aws.ToString(config.Color)
		}
		if config.Element != nil {
			tfMap["element"] = flattenDataPathValue(config.Element)
		}
		tfMap["time_granularity"] = types.TimeGranularity(config.TimeGranularity)

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataPathValue(apiObject *types.DataPathValue) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FieldValue != nil {
		tfMap["field_value"] = aws.ToString(apiObject.FieldValue)
	}

	return []interface{}{tfMap}
}

func flattenDataPathValues(apiObject []types.DataPathValue) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.FieldId != nil {
			tfMap["field_id"] = aws.ToString(config.FieldId)
		}
		if config.FieldValue != nil {
			tfMap["field_value"] = aws.ToString(config.FieldValue)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnHierarchy(apiObject []types.ColumnHierarchy) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.DateTimeHierarchy != nil {
			tfMap["date_time_hierarchy"] = flattenDateTimeHierarchy(config.DateTimeHierarchy)
		}
		if config.ExplicitHierarchy != nil {
			tfMap["explicit_hierarchy"] = flattenExplicitHierarchy(config.ExplicitHierarchy)
		}
		if config.PredefinedHierarchy != nil {
			tfMap["predefined_hierarchy"] = flattenPredefinedHierarchy(config.PredefinedHierarchy)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDateTimeHierarchy(apiObject *types.DateTimeHierarchy) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}
	if apiObject.DrillDownFilters != nil {
		tfMap["drill_down_filters"] = flattenDrillDownFilter(apiObject.DrillDownFilters)
	}

	return []interface{}{tfMap}
}

func flattenDrillDownFilter(apiObject []types.DrillDownFilter) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.CategoryFilter != nil {
			tfMap["category_filter"] = flattenCategoryDrillDownFilter(config.CategoryFilter)
		}
		if config.NumericEqualityFilter != nil {
			tfMap["numeric_equality_filter"] = flattenNumericEqualityDrillDownFilter(config.NumericEqualityFilter)
		}
		if config.TimeRangeFilter != nil {
			tfMap["time_range_filter"] = flattenTimeRangeDrillDownFilter(config.TimeRangeFilter)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCategoryDrillDownFilter(apiObject *types.CategoryDrillDownFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryValues != nil {
		tfMap["category_values"] = apiObject.CategoryValues
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}

	return []interface{}{tfMap}
}

func flattenNumericEqualityDrillDownFilter(apiObject *types.NumericEqualityDrillDownFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	tfMap["value"] = float64(apiObject.Value)

	return []interface{}{tfMap}
}

func flattenTimeRangeDrillDownFilter(apiObject *types.TimeRangeDrillDownFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.RangeMaximum != nil {
		tfMap["range_maximum"] = apiObject.RangeMaximum.Format(time.RFC3339)
	}
	if apiObject.RangeMinimum != nil {
		tfMap["range_minimum"] = apiObject.RangeMinimum.Format(time.RFC3339)
	}
	tfMap["time_granularity"] = types.TimeGranularity(apiObject.TimeGranularity)

	return []interface{}{tfMap}
}

func flattenExplicitHierarchy(apiObject *types.ExplicitHierarchy) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Columns != nil {
		tfMap["columns"] = flattenColumnIdentifiers(apiObject.Columns)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}
	if apiObject.DrillDownFilters != nil {
		tfMap["drill_down_filters"] = flattenDrillDownFilter(apiObject.DrillDownFilters)
	}

	return []interface{}{tfMap}
}

func flattenPredefinedHierarchy(apiObject *types.PredefinedHierarchy) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Columns != nil {
		tfMap["columns"] = flattenColumnIdentifiers(apiObject.Columns)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}
	if apiObject.DrillDownFilters != nil {
		tfMap["drill_down_filters"] = flattenDrillDownFilter(apiObject.DrillDownFilters)
	}

	return []interface{}{tfMap}
}

func flattenVisualSubtitleLabelOptions(apiObject *types.VisualSubtitleLabelOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FormatText != nil {
		tfMap["format_text"] = flattenLongFormatText(apiObject.FormatText)
	}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenLongFormatText(apiObject *types.LongFormatText) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PlainText != nil {
		tfMap["plain_text"] = aws.ToString(apiObject.PlainText)
	}
	if apiObject.RichText != nil {
		tfMap["rich_text"] = aws.ToString(apiObject.RichText)
	}

	return []interface{}{tfMap}
}

func flattenVisualTitleLabelOptions(apiObject *types.VisualTitleLabelOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FormatText != nil {
		tfMap["format_text"] = flattenShortFormatText(apiObject.FormatText)
	}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenShortFormatText(apiObject *types.ShortFormatText) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PlainText != nil {
		tfMap["plain_text"] = aws.ToString(apiObject.PlainText)
	}
	if apiObject.RichText != nil {
		tfMap["rich_text"] = aws.ToString(apiObject.RichText)
	}

	return []interface{}{tfMap}
}

func flattenColorScale(apiObject *types.ColorScale) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["color_fill_type"] = types.ColorFillType(apiObject.ColorFillType)
	if apiObject.Colors != nil {
		tfMap["colors"] = flattenDataColors(apiObject.Colors)
	}
	if apiObject.NullValueColor != nil {
		tfMap["null_value_color"] = flattenDataColor(apiObject.NullValueColor)
	}

	return []interface{}{tfMap}
}

func flattenDataColor(apiObject *types.DataColor) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	if apiObject.DataValue != nil {
		tfMap["data_value"] = aws.ToFloat64(apiObject.DataValue)
	}

	return []interface{}{tfMap}
}

func flattenDataColors(apiObject []types.DataColor) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.Color != nil {
			tfMap["color"] = aws.ToString(config.Color)
		}
		if config.DataValue != nil {
			tfMap["data_value"] = aws.ToFloat64(config.DataValue)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
