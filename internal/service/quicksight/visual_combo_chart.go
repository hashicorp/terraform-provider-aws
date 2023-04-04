package quicksight

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func comboChartVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"bar_data_labels":        dataLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"bars_arrangement":       stringSchema(false, validation.StringInSlice(quicksight.BarsArrangement_Values(), false)),
							"category_axis":          axisDisplayOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"category_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"color_label_options":    chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"combo_chart_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"bar_values":  measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
													"category":    dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"colors":      dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"line_values": measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":                           legendOptionsSchema(),                       // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"line_data_labels":                 dataLabelOptionsSchema(),                    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"primary_y_axis_display_options":   axisDisplayOptionsSchema(),                  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"primary_y_axis_label_options":     chartAxisLabelOptionsSchema(),               // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"reference_lines":                  referenceLineSchema(referenceLinesMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ReferenceLine.html
							"secondary_y_axis_display_options": axisDisplayOptionsSchema(),                  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AxisDisplayOptions.html
							"secondary_y_axis_label_options":   chartAxisLabelOptionsSchema(),               // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ComboChartSortConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":        fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html,
										"color_items_limit":    itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"color_sort":           fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"tooltip":        tooltipOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"visual_palette": visualPaletteSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
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

func expandComboChartVisual(tfList []interface{}) *quicksight.ComboChartVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.ComboChartVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandComboChartConfiguration(v)
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

func expandComboChartConfiguration(tfList []interface{}) *quicksight.ComboChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ComboChartConfiguration{}

	if v, ok := tfMap["bars_arrangement"].(string); ok && v != "" {
		config.BarsArrangement = aws.String(v)
	}
	if v, ok := tfMap["bar_data_labels"].([]interface{}); ok && len(v) > 0 {
		config.BarDataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["category_axis"].([]interface{}); ok && len(v) > 0 {
		config.CategoryAxis = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["category_label_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["color_label_options"].([]interface{}); ok && len(v) > 0 {
		config.ColorLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandComboChartFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		config.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["line_data_labels"].([]interface{}); ok && len(v) > 0 {
		config.LineDataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.PrimaryYAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["primary_y_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.PrimaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["reference_lines"].([]interface{}); ok && len(v) > 0 {
		config.ReferenceLines = expandReferenceLines(v)
	}
	if v, ok := tfMap["secondary_y_axis_display_options"].([]interface{}); ok && len(v) > 0 {
		config.SecondaryYAxisDisplayOptions = expandAxisDisplayOptions(v)
	}
	if v, ok := tfMap["secondary_y_axis_label_options"].([]interface{}); ok && len(v) > 0 {
		config.SecondaryYAxisLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandComboChartSortConfiguration(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		config.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palette"].([]interface{}); ok && len(v) > 0 {
		config.VisualPalette = expandVisualPalette(v)
	}

	return config
}

func expandComboChartFieldWells(tfList []interface{}) *quicksight.ComboChartFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ComboChartFieldWells{}

	if v, ok := tfMap["combo_chart_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.ComboChartAggregatedFieldWells = expandComboChartAggregatedFieldWells(v)
	}

	return config
}

func expandComboChartAggregatedFieldWells(tfList []interface{}) *quicksight.ComboChartAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ComboChartAggregatedFieldWells{}

	if v, ok := tfMap["bar_values"].([]interface{}); ok && len(v) > 0 {
		config.BarValues = expandMeasureFields(v)
	}
	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		config.Category = expandDimensionFields(v)
	}
	if v, ok := tfMap["colors"].([]interface{}); ok && len(v) > 0 {
		config.Colors = expandDimensionFields(v)
	}
	if v, ok := tfMap["line_values"].([]interface{}); ok && len(v) > 0 {
		config.LineValues = expandMeasureFields(v)
	}

	return config
}

func expandComboChartSortConfiguration(tfList []interface{}) *quicksight.ComboChartSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ComboChartSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}
	if v, ok := tfMap["color_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.ColorItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["color_sort"].([]interface{}); ok && len(v) > 0 {
		config.ColorSort = expandFieldSortOptionsList(v)
	}

	return config
}
