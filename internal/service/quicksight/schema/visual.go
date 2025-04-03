// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const customActionsMaxItems = 10
const referenceLinesMaxItems = 20
const dataPathValueMaxItems = 20

var visualsSchema = sync.OnceValue(func() *schema.Schema {
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
})

var legendOptionsSchema = sync.OnceValue(func() *schema.Schema {
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
				"position":   stringEnumSchema[awstypes.LegendPosition](attrOptional),
				"title":      labelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
				"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
				"width": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
})

var tooltipOptionsSchema = sync.OnceValue(func() *schema.Schema {
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
							"aggregation_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
													"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
													"field_id": stringLenBetweenSchema(attrRequired, 1, 512),
													"label": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
												},
											},
										},
									},
								},
							},
							"tooltip_title_type": stringEnumSchema[awstypes.TooltipTitleType](attrOptional),
						},
					},
				},
				"selected_tooltip_type": stringEnumSchema[awstypes.SelectedTooltipType](attrOptional),
				"tooltip_visibility":    stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

var visualPaletteSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"chart_color": hexColorSchema(attrOptional),
				"color_map": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathColor.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 5000,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":            hexColorSchema(attrRequired),
							"element":          dataPathValueSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathValue.html
							"time_granularity": stringEnumSchema[awstypes.TimeGranularity](attrOptional),
						},
					},
				},
			},
		},
	}
})

func dataPathValueSchema(maxItems int) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPathValue.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_id":    stringLenBetweenSchema(attrRequired, 1, 512),
				"field_value": stringLenBetweenSchema(attrRequired, 1, 2048),
			},
		},
	}
}

var columnHierarchiesSchema = sync.OnceValue(func() *schema.Schema {
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
							"hierarchy_id":       stringLenBetweenSchema(attrRequired, 1, 512),
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
										"column_name":         stringLenBetweenSchema(attrRequired, 1, 128),
										"data_set_identifier": stringLenBetweenSchema(attrRequired, 1, 2048),
									},
								},
							},
							"hierarchy_id":       stringLenBetweenSchema(attrRequired, 1, 512),
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
										"column_name":         stringLenBetweenSchema(attrRequired, 1, 128),
										"data_set_identifier": stringLenBetweenSchema(attrRequired, 1, 2048),
									},
								},
							},
							"hierarchy_id":       stringLenBetweenSchema(attrRequired, 1, 512),
							"drill_down_filters": drillDownFilterSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DrillDownFilter.html
						},
					},
				},
			},
		},
	}
})

var visualSubtitleLabelOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"format_text": longFormatTextSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LongFormatText.html
				"visibility":  stringEnumSchema[awstypes.Visibility](attrOptionalComputed),
			},
		},
	}
})

func longFormatTextSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LongFormatText.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"plain_text": stringLenBetweenSchema(attrOptional, 1, 1024),
				"rich_text":  stringLenBetweenSchema(attrOptional, 1, 2048),
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
				"plain_text": stringLenBetweenSchema(attrOptional, 1, 512),
				"rich_text":  stringLenBetweenSchema(attrOptional, 1, 1024),
			},
		},
	}
}

var visualTitleLabelOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"format_text": shortFormatTextSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ShortFormatText.html
				"visibility":  stringEnumSchema[awstypes.Visibility](attrOptionalComputed),
			},
		},
	}
})

var comparisonConfigurationSchema = sync.OnceValue(func() *schema.Schema {
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
				"comparison_method": stringEnumSchema[awstypes.ComparisonMethod](attrOptional),
			},
		},
	}
})

var colorScaleSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColorScale.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"color_fill_type": stringEnumSchema[awstypes.ColorFillType](attrRequired),
				"colors": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataColor.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 2,
					MaxItems: 3,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color": hexColorSchema(attrOptional),
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
							"color": hexColorSchema(attrOptional),
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
})

var dataLabelOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"category_label_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
										"field_id":    stringLenBetweenSchema(attrOptional, 1, 512),
										"field_value": stringLenBetweenSchema(attrOptional, 1, 2048),
										"visibility":  stringEnumSchema[awstypes.Visibility](attrOptional),
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
										"field_id":   stringLenBetweenSchema(attrOptional, 1, 512),
										"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
										"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
										"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
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
										"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
									},
								},
							},
						},
					},
				},
				"label_color":              hexColorSchema(attrOptional),
				"label_content":            stringEnumSchema[awstypes.DataLabelContent](attrOptional),
				"label_font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
				"measure_label_visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
				"overlap":                  stringEnumSchema[awstypes.DataLabelOverlap](attrOptional),
				"position":                 stringEnumSchema[awstypes.DataLabelPosition](attrOptional),
				"visibility":               stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

func hexColorSchema(handling attrHandling) *schema.Schema {
	return stringMatchSchema(handling, `^#[0-9A-F]{6}$`, "")
}

func expandVisual(tfMap map[string]any) *awstypes.Visual {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Visual{}

	if v, ok := tfMap["bar_chart_visual"].([]any); ok && len(v) > 0 {
		apiObject.BarChartVisual = expandBarChartVisual(v)
	}
	if v, ok := tfMap["box_plot_visual"].([]any); ok && len(v) > 0 {
		apiObject.BoxPlotVisual = expandBoxPlotVisual(v)
	}
	if v, ok := tfMap["combo_chart_visual"].([]any); ok && len(v) > 0 {
		apiObject.ComboChartVisual = expandComboChartVisual(v)
	}
	if v, ok := tfMap["custom_content_visual"].([]any); ok && len(v) > 0 {
		apiObject.CustomContentVisual = expandCustomContentVisual(v)
	}
	if v, ok := tfMap["empty_visual"].([]any); ok && len(v) > 0 {
		apiObject.EmptyVisual = expandEmptyVisual(v)
	}
	if v, ok := tfMap["filled_map_visual"].([]any); ok && len(v) > 0 {
		apiObject.FilledMapVisual = expandFilledMapVisual(v)
	}
	if v, ok := tfMap["funnel_chart_visual"].([]any); ok && len(v) > 0 {
		apiObject.FunnelChartVisual = expandFunnelChartVisual(v)
	}
	if v, ok := tfMap["gauge_chart_visual"].([]any); ok && len(v) > 0 {
		apiObject.GaugeChartVisual = expandGaugeChartVisual(v)
	}
	if v, ok := tfMap["geospatial_map_visual"].([]any); ok && len(v) > 0 {
		apiObject.GeospatialMapVisual = expandGeospatialMapVisual(v)
	}
	if v, ok := tfMap["heat_map_visual"].([]any); ok && len(v) > 0 {
		apiObject.HeatMapVisual = expandHeatMapVisual(v)
	}
	if v, ok := tfMap["histogram_visual"].([]any); ok && len(v) > 0 {
		apiObject.HistogramVisual = expandHistogramVisual(v)
	}
	if v, ok := tfMap["insight_visual"].([]any); ok && len(v) > 0 {
		apiObject.InsightVisual = expandInsightVisual(v)
	}
	if v, ok := tfMap["kpi_visual"].([]any); ok && len(v) > 0 {
		apiObject.KPIVisual = expandKPIVisual(v)
	}
	if v, ok := tfMap["line_chart_visual"].([]any); ok && len(v) > 0 {
		apiObject.LineChartVisual = expandLineChartVisual(v)
	}
	if v, ok := tfMap["pie_chart_visual"].([]any); ok && len(v) > 0 {
		apiObject.PieChartVisual = expandPieChartVisual(v)
	}
	if v, ok := tfMap["pivot_table_visual"].([]any); ok && len(v) > 0 {
		apiObject.PivotTableVisual = expandPivotTableVisual(v)
	}
	if v, ok := tfMap["radar_chart_visual"].([]any); ok && len(v) > 0 {
		apiObject.RadarChartVisual = expandRadarChartVisual(v)
	}
	if v, ok := tfMap["sankey_diagram_visual"].([]any); ok && len(v) > 0 {
		apiObject.SankeyDiagramVisual = expandSankeyDiagramVisual(v)
	}
	if v, ok := tfMap["scatter_plot_visual"].([]any); ok && len(v) > 0 {
		apiObject.ScatterPlotVisual = expandScatterPlotVisual(v)
	}
	if v, ok := tfMap["table_visual"].([]any); ok && len(v) > 0 {
		apiObject.TableVisual = expandTableVisual(v)
	}
	if v, ok := tfMap["tree_map_visual"].([]any); ok && len(v) > 0 {
		apiObject.TreeMapVisual = expandTreeMapVisual(v)
	}
	if v, ok := tfMap["waterfall_visual"].([]any); ok && len(v) > 0 {
		apiObject.WaterfallVisual = expandWaterfallVisual(v)
	}
	if v, ok := tfMap["word_cloud_visual"].([]any); ok && len(v) > 0 {
		apiObject.WordCloudVisual = expandWordCloudVisual(v)
	}

	return apiObject
}

func expandDataLabelOptions(tfList []any) *awstypes.DataLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataLabelOptions{}

	if v, ok := tfMap["category_label_visibility"].(string); ok && v != "" {
		apiObject.CategoryLabelVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["label_color"].(string); ok && v != "" {
		apiObject.LabelColor = aws.String(v)
	}
	if v, ok := tfMap["label_content"].(string); ok && v != "" {
		apiObject.LabelContent = awstypes.DataLabelContent(v)
	}
	if v, ok := tfMap["measure_label_visibility"].(string); ok && v != "" {
		apiObject.MeasureLabelVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["overlap"].(string); ok && v != "" {
		apiObject.Overlap = awstypes.DataLabelOverlap(v)
	}
	if v, ok := tfMap["position"].(string); ok && v != "" {
		apiObject.Position = awstypes.DataLabelPosition(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["data_label_types"].([]any); ok && len(v) > 0 {
		apiObject.DataLabelTypes = expandDataLabelTypes(v)
	}
	if v, ok := tfMap["label_font_configuration"].([]any); ok && len(v) > 0 {
		apiObject.LabelFontConfiguration = expandFontConfiguration(v)
	}

	return apiObject
}

func expandDataLabelTypes(tfList []any) []awstypes.DataLabelType {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DataLabelType

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandDataLabelType(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDataLabelType(tfMap map[string]any) *awstypes.DataLabelType {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataLabelType{}

	if v, ok := tfMap["data_path_label_type"].([]any); ok && len(v) > 0 {
		apiObject.DataPathLabelType = expandDataPathLabelType(v)
	}
	if v, ok := tfMap["field_label_type"].([]any); ok && len(v) > 0 {
		apiObject.FieldLabelType = expandFieldLabelType(v)
	}
	if v, ok := tfMap["maximum_label_type"].([]any); ok && len(v) > 0 {
		apiObject.MaximumLabelType = expandMaximumLabelType(v)
	}
	if v, ok := tfMap["minimum_label_type"].([]any); ok && len(v) > 0 {
		apiObject.MinimumLabelType = expandMinimumLabelType(v)
	}
	if v, ok := tfMap["range_ends_label_type"].([]any); ok && len(v) > 0 {
		apiObject.RangeEndsLabelType = expandRangeEndsLabelType(v)
	}

	return apiObject
}

func expandDataPathLabelType(tfList []any) *awstypes.DataPathLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataPathLabelType{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["field_value"].(string); ok && v != "" {
		apiObject.FieldValue = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandFieldLabelType(tfList []any) *awstypes.FieldLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FieldLabelType{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandMaximumLabelType(tfList []any) *awstypes.MaximumLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.MaximumLabelType{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandMinimumLabelType(tfList []any) *awstypes.MinimumLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.MinimumLabelType{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandRangeEndsLabelType(tfList []any) *awstypes.RangeEndsLabelType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RangeEndsLabelType{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandLegendOptions(tfList []any) *awstypes.LegendOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LegendOptions{}

	if v, ok := tfMap["height"].(string); ok && v != "" {
		apiObject.Height = aws.String(v)
	}
	if v, ok := tfMap["position"].(string); ok && v != "" {
		apiObject.Position = awstypes.LegendPosition(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["width"].(string); ok && v != "" {
		apiObject.Width = aws.String(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandLabelOptions(v)
	}

	return apiObject
}

func expandTooltipOptions(tfList []any) *awstypes.TooltipOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TooltipOptions{}

	if v, ok := tfMap["selected_tooltip_type"].(string); ok && v != "" {
		apiObject.SelectedTooltipType = awstypes.SelectedTooltipType(v)
	}
	if v, ok := tfMap["tooltip_visibility"].(string); ok && v != "" {
		apiObject.TooltipVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["field_base_tooltip"].([]any); ok && len(v) > 0 {
		apiObject.FieldBasedTooltip = expandFieldBasedTooltip(v)
	}

	return apiObject
}

func expandFieldBasedTooltip(tfList []any) *awstypes.FieldBasedTooltip {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FieldBasedTooltip{}

	if v, ok := tfMap["aggregation_visibility"].(string); ok && v != "" {
		apiObject.AggregationVisibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["tooltip_title_type"].(string); ok && v != "" {
		apiObject.TooltipTitleType = awstypes.TooltipTitleType(v)
	}
	if v, ok := tfMap["tooltip_fields"].([]any); ok && len(v) > 0 {
		apiObject.TooltipFields = expandTooltipItems(v)
	}

	return apiObject
}

func expandTooltipItems(tfList []any) []awstypes.TooltipItem {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.TooltipItem

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandTooltipItem(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandTooltipItem(tfMap map[string]any) *awstypes.TooltipItem {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TooltipItem{}

	if v, ok := tfMap["column_tooltip_item"].([]any); ok && len(v) > 0 {
		apiObject.ColumnTooltipItem = expandColumnTooltipItem(v)
	}
	if v, ok := tfMap["field_tooltip_item"].([]any); ok && len(v) > 0 {
		apiObject.FieldTooltipItem = expandFieldTooltipItem(v)
	}

	return apiObject
}

func expandColumnTooltipItem(tfList []any) *awstypes.ColumnTooltipItem {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ColumnTooltipItem{}

	if v, ok := tfMap["label"].(string); ok && v != "" {
		apiObject.Label = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["aggregation"].([]any); ok && len(v) > 0 {
		apiObject.Aggregation = expandAggregationFunction(v)
	}

	return apiObject
}

func expandFieldTooltipItem(tfList []any) *awstypes.FieldTooltipItem {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FieldTooltipItem{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["label"].(string); ok && v != "" {
		apiObject.Label = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandVisualPalette(tfList []any) *awstypes.VisualPalette {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.VisualPalette{}

	if v, ok := tfMap["chart_color"].(string); ok && v != "" {
		apiObject.ChartColor = aws.String(v)
	}
	if v, ok := tfMap["color_map"].([]any); ok && len(v) > 0 {
		apiObject.ColorMap = expandDataPathColors(v)
	}

	return apiObject
}

func expandDataPathColors(tfList []any) []awstypes.DataPathColor {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DataPathColor

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandDataPathColor(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDataPathColor(tfMap map[string]any) *awstypes.DataPathColor {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataPathColor{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		apiObject.TimeGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["element"].([]any); ok && len(v) > 0 {
		apiObject.Element = expandDataPathValue(v)
	}

	return apiObject
}

func expandDataPathValues(tfList []any) []awstypes.DataPathValue {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DataPathValue

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandDataPathValueInternal(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDataPathValueInternal(tfMap map[string]any) *awstypes.DataPathValue {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataPathValue{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["field_value"].(string); ok && v != "" {
		apiObject.FieldValue = aws.String(v)
	}

	return apiObject
}

func expandDataPathValue(tfList []any) *awstypes.DataPathValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	return expandDataPathValueInternal(tfMap)
}

func expandColumnHierarchies(tfList []any) []awstypes.ColumnHierarchy {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnHierarchy

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandColumnHierarchy(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandColumnHierarchy(tfMap map[string]any) *awstypes.ColumnHierarchy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColumnHierarchy{}

	if v, ok := tfMap["date_time_hierarchy"].([]any); ok && len(v) > 0 {
		apiObject.DateTimeHierarchy = expandDateTimeHierarchy(v)
	}
	if v, ok := tfMap["explicit_hierarchy"].([]any); ok && len(v) > 0 {
		apiObject.ExplicitHierarchy = expandExplicitHierarchy(v)
	}
	if v, ok := tfMap["predefined_hierarchy"].([]any); ok && len(v) > 0 {
		apiObject.PredefinedHierarchy = expandPredefinedHierarchy(v)
	}

	return apiObject
}

func expandDateTimeHierarchy(tfList []any) *awstypes.DateTimeHierarchy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateTimeHierarchy{}

	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		apiObject.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["drill_down_filters"].([]any); ok && len(v) > 0 {
		apiObject.DrillDownFilters = expandDrillDownFilters(v)
	}

	return apiObject
}

func expandExplicitHierarchy(tfList []any) *awstypes.ExplicitHierarchy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ExplicitHierarchy{}

	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		apiObject.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["columns"].([]any); ok && len(v) > 0 {
		apiObject.Columns = expandColumnIdentifiers(v)
	}
	if v, ok := tfMap["drill_down_filters"].([]any); ok && len(v) > 0 {
		apiObject.DrillDownFilters = expandDrillDownFilters(v)
	}

	return apiObject
}
func expandPredefinedHierarchy(tfList []any) *awstypes.PredefinedHierarchy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PredefinedHierarchy{}

	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		apiObject.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["columns"].([]any); ok && len(v) > 0 {
		apiObject.Columns = expandColumnIdentifiers(v)
	}
	if v, ok := tfMap["drill_down_filters"].([]any); ok && len(v) > 0 {
		apiObject.DrillDownFilters = expandDrillDownFilters(v)
	}

	return apiObject
}

func expandVisualSubtitleLabelOptions(tfList []any) *awstypes.VisualSubtitleLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.VisualSubtitleLabelOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["format_text"].([]any); ok && len(v) > 0 {
		apiObject.FormatText = expandLongFormatText(v)
	}

	return apiObject
}

func expandLongFormatText(tfList []any) *awstypes.LongFormatText {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LongFormatText{}

	if v, ok := tfMap["plain_text"].(string); ok && v != "" {
		apiObject.PlainText = aws.String(v)
	}
	if v, ok := tfMap["rich_text"].(string); ok && v != "" {
		apiObject.RichText = aws.String(v)
	}

	return apiObject
}

func expandVisualTitleLabelOptions(tfList []any) *awstypes.VisualTitleLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.VisualTitleLabelOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["format_text"].([]any); ok && len(v) > 0 {
		apiObject.FormatText = expandShortFormatText(v)
	}

	return apiObject
}

func expandShortFormatText(tfList []any) *awstypes.ShortFormatText {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ShortFormatText{}

	if v, ok := tfMap["plain_text"].(string); ok && v != "" {
		apiObject.PlainText = aws.String(v)
	}
	if v, ok := tfMap["rich_text"].(string); ok && v != "" {
		apiObject.RichText = aws.String(v)
	}

	return apiObject
}

func expandComparisonConfiguration(tfList []any) *awstypes.ComparisonConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ComparisonConfiguration{}

	if v, ok := tfMap["comparison_method"].(string); ok && v != "" {
		apiObject.ComparisonMethod = awstypes.ComparisonMethod(v)
	}
	if v, ok := tfMap["comparison_format"].([]any); ok && len(v) > 0 {
		apiObject.ComparisonFormat = expandComparisonFormatConfiguration(v)
	}

	return apiObject
}

func expandColorScale(tfList []any) *awstypes.ColorScale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ColorScale{}

	if v, ok := tfMap["color_fill_type"].(string); ok && v != "" {
		apiObject.ColorFillType = awstypes.ColorFillType(v)
	}
	if v, ok := tfMap["colors"].([]any); ok && len(v) > 0 {
		apiObject.Colors = expandDataColors(v)
	}
	if v, ok := tfMap["null_value_color"].([]any); ok && len(v) > 0 {
		apiObject.NullValueColor = expandDataColor(v)
	}

	return apiObject
}

func expandDataColor(tfList []any) *awstypes.DataColor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	return expandDataColorInternal(tfMap)
}

func expandDataColorInternal(tfMap map[string]any) *awstypes.DataColor {
	apiObject := &awstypes.DataColor{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap["data_value"].(float64); ok {
		apiObject.DataValue = aws.Float64(v)
	}

	return apiObject
}

func expandDataColors(tfList []any) []awstypes.DataColor {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DataColor

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandDataColorInternal(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenVisuals(apiObjects []awstypes.Visual) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.BarChartVisual != nil {
			tfMap["bar_chart_visual"] = flattenBarChartVisual(apiObject.BarChartVisual)
		}
		if apiObject.BoxPlotVisual != nil {
			tfMap["box_plot_visual"] = flattenBoxPlotVisual(apiObject.BoxPlotVisual)
		}
		if apiObject.ComboChartVisual != nil {
			tfMap["combo_chart_visual"] = flattenComboChartVisual(apiObject.ComboChartVisual)
		}
		if apiObject.CustomContentVisual != nil {
			tfMap["custom_content_visual"] = flattenCustomContentVisual(apiObject.CustomContentVisual)
		}
		if apiObject.EmptyVisual != nil {
			tfMap["empty_visual"] = flattenEmptyVisual(apiObject.EmptyVisual)
		}
		if apiObject.FilledMapVisual != nil {
			tfMap["filled_map_visual"] = flattenFilledMapVisual(apiObject.FilledMapVisual)
		}
		if apiObject.FunnelChartVisual != nil {
			tfMap["funnel_chart_visual"] = flattenFunnelChartVisual(apiObject.FunnelChartVisual)
		}
		if apiObject.GaugeChartVisual != nil {
			tfMap["gauge_chart_visual"] = flattenGaugeChartVisual(apiObject.GaugeChartVisual)
		}
		if apiObject.GeospatialMapVisual != nil {
			tfMap["geospatial_map_visual"] = flattenGeospatialMapVisual(apiObject.GeospatialMapVisual)
		}
		if apiObject.HeatMapVisual != nil {
			tfMap["heat_map_visual"] = flattenHeatMapVisual(apiObject.HeatMapVisual)
		}
		if apiObject.HistogramVisual != nil {
			tfMap["histogram_visual"] = flattenHistogramVisual(apiObject.HistogramVisual)
		}
		if apiObject.InsightVisual != nil {
			tfMap["insight_visual"] = flattenInsightVisual(apiObject.InsightVisual)
		}
		if apiObject.KPIVisual != nil {
			tfMap["kpi_visual"] = flattenKPIVisual(apiObject.KPIVisual)
		}
		if apiObject.LineChartVisual != nil {
			tfMap["line_chart_visual"] = flattenLineChartVisual(apiObject.LineChartVisual)
		}
		if apiObject.PieChartVisual != nil {
			tfMap["pie_chart_visual"] = flattenPieChartVisual(apiObject.PieChartVisual)
		}
		if apiObject.PivotTableVisual != nil {
			tfMap["pivot_table_visual"] = flattenPivotTableVisual(apiObject.PivotTableVisual)
		}
		if apiObject.RadarChartVisual != nil {
			tfMap["radar_chart_visual"] = flattenRadarChartVisual(apiObject.RadarChartVisual)
		}
		if apiObject.SankeyDiagramVisual != nil {
			tfMap["sankey_diagram_visual"] = flattenSankeyDiagramVisual(apiObject.SankeyDiagramVisual)
		}
		if apiObject.ScatterPlotVisual != nil {
			tfMap["scatter_plot_visual"] = flattenScatterPlotVisual(apiObject.ScatterPlotVisual)
		}
		if apiObject.TableVisual != nil {
			tfMap["table_visual"] = flattenTableVisual(apiObject.TableVisual)
		}
		if apiObject.TreeMapVisual != nil {
			tfMap["tree_map_visual"] = flattenTreeMapVisual(apiObject.TreeMapVisual)
		}
		if apiObject.WaterfallVisual != nil {
			tfMap["waterfall_visual"] = flattenWaterfallVisual(apiObject.WaterfallVisual)
		}
		if apiObject.WordCloudVisual != nil {
			tfMap["word_cloud_visual"] = flattenWordCloudVisual(apiObject.WordCloudVisual)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataLabelOptions(apiObject *awstypes.DataLabelOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["category_label_visibility"] = apiObject.CategoryLabelVisibility
	if apiObject.DataLabelTypes != nil {
		tfMap["data_label_types"] = flattenDataLabelType(apiObject.DataLabelTypes)
	}
	if apiObject.LabelColor != nil {
		tfMap["label_color"] = aws.ToString(apiObject.LabelColor)
	}
	tfMap["label_content"] = apiObject.LabelContent
	if apiObject.LabelFontConfiguration != nil {
		tfMap["label_font_configuration"] = flattenFontConfiguration(apiObject.LabelFontConfiguration)
	}
	tfMap["measure_label_visibility"] = apiObject.MeasureLabelVisibility
	tfMap["overlap"] = apiObject.Overlap
	tfMap["position"] = apiObject.Position
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenDataLabelType(apiObjects []awstypes.DataLabelType) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.DataPathLabelType != nil {
			tfMap["data_path_label_type"] = flattenDataPathLabelType(apiObject.DataPathLabelType)
		}
		if apiObject.FieldLabelType != nil {
			tfMap["field_label_type"] = flattenFieldLabelType(apiObject.FieldLabelType)
		}
		if apiObject.MaximumLabelType != nil {
			tfMap["maximum_label_type"] = flattenMaximumLabelType(apiObject.MaximumLabelType)
		}
		if apiObject.MinimumLabelType != nil {
			tfMap["minimum_label_type"] = flattenMinimumLabelType(apiObject.MinimumLabelType)
		}
		if apiObject.RangeEndsLabelType != nil {
			tfMap["range_ends_label_type"] = flattenRangeEndsLabelType(apiObject.RangeEndsLabelType)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataPathLabelType(apiObject *awstypes.DataPathLabelType) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FieldValue != nil {
		tfMap["field_value"] = aws.ToString(apiObject.FieldValue)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenFieldLabelType(apiObject *awstypes.FieldLabelType) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenMaximumLabelType(apiObject *awstypes.MaximumLabelType) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenMinimumLabelType(apiObject *awstypes.MinimumLabelType) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenRangeEndsLabelType(apiObject *awstypes.RangeEndsLabelType) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenLegendOptions(apiObject *awstypes.LegendOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Height != nil {
		tfMap["height"] = aws.ToString(apiObject.Height)
	}
	tfMap["position"] = apiObject.Position
	if apiObject.Title != nil {
		tfMap["title"] = flattenLabelOptions(apiObject.Title)
	}
	tfMap["visibility"] = apiObject.Visibility
	if apiObject.Width != nil {
		tfMap["width"] = aws.ToString(apiObject.Width)
	}

	return []any{tfMap}
}

func flattenTooltipOptions(apiObject *awstypes.TooltipOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldBasedTooltip != nil {
		tfMap["field_base_tooltip"] = flattenFieldBasedTooltip(apiObject.FieldBasedTooltip)
	}
	tfMap["selected_tooltip_type"] = apiObject.SelectedTooltipType
	tfMap["tooltip_visibility"] = apiObject.TooltipVisibility

	return []any{tfMap}
}

func flattenFieldBasedTooltip(apiObject *awstypes.FieldBasedTooltip) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["aggregation_visibility"] = apiObject.AggregationVisibility
	if apiObject.TooltipFields != nil {
		tfMap["tooltip_fields"] = flattenTooltipItem(apiObject.TooltipFields)
	}
	tfMap["tooltip_title_type"] = apiObject.TooltipTitleType

	return []any{tfMap}
}

func flattenTooltipItem(apiObjects []awstypes.TooltipItem) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.ColumnTooltipItem != nil {
			tfMap["column_tooltip_item"] = flattenColumnTooltipItem(apiObject.ColumnTooltipItem)
		}
		if apiObject.FieldTooltipItem != nil {
			tfMap["field_tooltip_item"] = flattenFieldTooltipItem(apiObject.FieldTooltipItem)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnTooltipItem(apiObject *awstypes.ColumnTooltipItem) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.Aggregation != nil {
		tfMap["aggregation"] = flattenAggregationFunction(apiObject.Aggregation)
	}
	if apiObject.Label != nil {
		tfMap["label"] = aws.ToString(apiObject.Label)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenFieldTooltipItem(apiObject *awstypes.FieldTooltipItem) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.Label != nil {
		tfMap["label"] = aws.ToString(apiObject.Label)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenVisualPalette(apiObject *awstypes.VisualPalette) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}
	if apiObject.ChartColor != nil {
		tfMap["chart_color"] = aws.ToString(apiObject.ChartColor)
	}
	if apiObject.ColorMap != nil {
		tfMap["color_map"] = flattenDataPathColor(apiObject.ColorMap)
	}

	return []any{tfMap}
}

func flattenDataPathColor(apiObjects []awstypes.DataPathColor) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Color != nil {
			tfMap["color"] = aws.ToString(apiObject.Color)
		}
		if apiObject.Element != nil {
			tfMap["element"] = flattenDataPathValue(apiObject.Element)
		}
		tfMap["time_granularity"] = apiObject.TimeGranularity

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataPathValue(apiObject *awstypes.DataPathValue) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FieldValue != nil {
		tfMap["field_value"] = aws.ToString(apiObject.FieldValue)
	}

	return []any{tfMap}
}

func flattenDataPathValues(apiObjects []awstypes.DataPathValue) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.FieldId != nil {
			tfMap["field_id"] = aws.ToString(apiObject.FieldId)
		}
		if apiObject.FieldValue != nil {
			tfMap["field_value"] = aws.ToString(apiObject.FieldValue)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnHierarchy(apiObjects []awstypes.ColumnHierarchy) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.DateTimeHierarchy != nil {
			tfMap["date_time_hierarchy"] = flattenDateTimeHierarchy(apiObject.DateTimeHierarchy)
		}
		if apiObject.ExplicitHierarchy != nil {
			tfMap["explicit_hierarchy"] = flattenExplicitHierarchy(apiObject.ExplicitHierarchy)
		}
		if apiObject.PredefinedHierarchy != nil {
			tfMap["predefined_hierarchy"] = flattenPredefinedHierarchy(apiObject.PredefinedHierarchy)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDateTimeHierarchy(apiObject *awstypes.DateTimeHierarchy) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}
	if apiObject.DrillDownFilters != nil {
		tfMap["drill_down_filters"] = flattenDrillDownFilter(apiObject.DrillDownFilters)
	}

	return []any{tfMap}
}

func flattenDrillDownFilter(apiObjects []awstypes.DrillDownFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.CategoryFilter != nil {
			tfMap["category_filter"] = flattenCategoryDrillDownFilter(apiObject.CategoryFilter)
		}
		if apiObject.NumericEqualityFilter != nil {
			tfMap["numeric_equality_filter"] = flattenNumericEqualityDrillDownFilter(apiObject.NumericEqualityFilter)
		}
		if apiObject.TimeRangeFilter != nil {
			tfMap["time_range_filter"] = flattenTimeRangeDrillDownFilter(apiObject.TimeRangeFilter)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCategoryDrillDownFilter(apiObject *awstypes.CategoryDrillDownFilter) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CategoryValues != nil {
		tfMap["category_values"] = apiObject.CategoryValues
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}

	return []any{tfMap}
}

func flattenNumericEqualityDrillDownFilter(apiObject *awstypes.NumericEqualityDrillDownFilter) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	tfMap[names.AttrValue] = apiObject.Value

	return []any{tfMap}
}

func flattenTimeRangeDrillDownFilter(apiObject *awstypes.TimeRangeDrillDownFilter) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.RangeMaximum != nil {
		tfMap["range_maximum"] = aws.ToTime(apiObject.RangeMaximum).Format(time.RFC3339)
	}
	if apiObject.RangeMinimum != nil {
		tfMap["range_minimum"] = aws.ToTime(apiObject.RangeMinimum).Format(time.RFC3339)
	}
	tfMap["time_granularity"] = apiObject.TimeGranularity

	return []any{tfMap}
}

func flattenExplicitHierarchy(apiObject *awstypes.ExplicitHierarchy) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Columns != nil {
		tfMap["columns"] = flattenColumnIdentifiers(apiObject.Columns)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}
	if apiObject.DrillDownFilters != nil {
		tfMap["drill_down_filters"] = flattenDrillDownFilter(apiObject.DrillDownFilters)
	}

	return []any{tfMap}
}

func flattenPredefinedHierarchy(apiObject *awstypes.PredefinedHierarchy) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Columns != nil {
		tfMap["columns"] = flattenColumnIdentifiers(apiObject.Columns)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}
	if apiObject.DrillDownFilters != nil {
		tfMap["drill_down_filters"] = flattenDrillDownFilter(apiObject.DrillDownFilters)
	}

	return []any{tfMap}
}

func flattenVisualSubtitleLabelOptions(apiObject *awstypes.VisualSubtitleLabelOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FormatText != nil {
		tfMap["format_text"] = flattenLongFormatText(apiObject.FormatText)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenLongFormatText(apiObject *awstypes.LongFormatText) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.PlainText != nil {
		tfMap["plain_text"] = aws.ToString(apiObject.PlainText)
	}
	if apiObject.RichText != nil {
		tfMap["rich_text"] = aws.ToString(apiObject.RichText)
	}

	return []any{tfMap}
}

func flattenVisualTitleLabelOptions(apiObject *awstypes.VisualTitleLabelOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FormatText != nil {
		tfMap["format_text"] = flattenShortFormatText(apiObject.FormatText)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenShortFormatText(apiObject *awstypes.ShortFormatText) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.PlainText != nil {
		tfMap["plain_text"] = aws.ToString(apiObject.PlainText)
	}
	if apiObject.RichText != nil {
		tfMap["rich_text"] = aws.ToString(apiObject.RichText)
	}

	return []any{tfMap}
}

func flattenColorScale(apiObject *awstypes.ColorScale) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["color_fill_type"] = apiObject.ColorFillType
	if apiObject.Colors != nil {
		tfMap["colors"] = flattenDataColors(apiObject.Colors)
	}
	if apiObject.NullValueColor != nil {
		tfMap["null_value_color"] = flattenDataColor(apiObject.NullValueColor)
	}

	return []any{tfMap}
}

func flattenDataColor(apiObject *awstypes.DataColor) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	if apiObject.DataValue != nil {
		tfMap["data_value"] = aws.ToFloat64(apiObject.DataValue)
	}

	return []any{tfMap}
}

func flattenDataColors(apiObject []awstypes.DataColor) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObject {
		tfMap := map[string]any{}

		if apiObject.Color != nil {
			tfMap["color"] = aws.ToString(apiObject.Color)
		}
		if apiObject.DataValue != nil {
			tfMap["data_value"] = aws.ToFloat64(apiObject.DataValue)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
