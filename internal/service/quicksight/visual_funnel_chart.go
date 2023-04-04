package quicksight

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func funnelChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"category_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"data_label_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartDataLabelOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_label_visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
										"label_color":               stringSchema(false, validation.StringMatch(regexp.MustCompile(`^#[A-F0-9]{6}$`), "")),
										"label_font_configuration":  fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
										"measure_data_label_style":  stringSchema(false, validation.StringInSlice(quicksight.FunnelChartMeasureDataLabelStyle_Values(), false)),
										"measure_label_visibility":  stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
										"position":                  stringSchema(false, validation.StringInSlice(quicksight.DataLabelPosition_Values(), false)),
										"visibility":                stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
									},
								},
							},
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"funnel_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"category": dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"values":   measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FunnelChartSortConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":        fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
									},
								},
							},
							"tooltip":             tooltipOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"value_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"visual_palette":      visualPaletteSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
						},
					},
				},
				"column_hierarchies": columnHierarchiesSchema(),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnHierarchy.html
				"subtitle":           visualSubtitleLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
				"title":              visualTitleLabelOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
			},
		},
	}
}

func expandFunnelChartVisual(tfList []interface{}) *quicksight.FunnelChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.FunnelChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandFunnelChartConfiguration(v)
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

func expandFunnelChartConfiguration(tfList []interface{}) *quicksight.FunnelChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FunnelChartConfiguration{}

	if v, ok := tfMap["category_label_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["data_label_options"].([]interface{}); ok && len(v) > 0 {
		config.DataLabelOptions = expandFunnelChartDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandFunnelChartFieldWells(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandFunnelChartSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["value_label_options"].([]interface{}); ok && len(v) > 0 {
		config.ValueLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}

	return config
}

func expandFunnelChartFieldWells(tfList []interface{}) *quicksight.FunnelChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FunnelChartFieldWells{}

	if v, ok := tfMap["funnel_chart_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FunnelChartAggregatedFieldWells = expandFunnelChartAggregatedFieldWells(v)
	}

	return config
}

func expandFunnelChartAggregatedFieldWells(tfList []interface{}) *quicksight.FunnelChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FunnelChartAggregatedFieldWells{}

	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		config.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		config.Values = expandMeasureFields(v)
	}

	return config
}

func expandFunnelChartSortConfiguration(tfList []interface{}) *quicksight.FunnelChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FunnelChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandFunnelChartDataLabelOptions(tfList []interface{}) *quicksight.FunnelChartDataLabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.FunnelChartDataLabelOptions{}

	if v, ok := tfMap["category_label_visibility"].(string); ok && v != "" {
		options.CategoryLabelVisibility = aws.String(v)
	}
	if v, ok := tfMap["label_color"].(string); ok && v != "" {
		options.LabelColor = aws.String(v)
	}
	if v, ok := tfMap["measure_data_label_style"].(string); ok && v != "" {
		options.MeasureDataLabelStyle = aws.String(v)
	}
	if v, ok := tfMap["measure_label_visibility"].(string); ok && v != "" {
		options.MeasureLabelVisibility = aws.String(v)
	}
	if v, ok := tfMap["position"].(string); ok && v != "" {
		options.Position = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}
	if v, ok := tfMap["label_font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.LabelFontConfiguration = expandFontConfiguration(v)
	}

	return options
}
