package quicksight

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func wordCloudVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WordCloudVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id": idSchema(),
				"actions":   visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WordCloudChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"category_label_options": chartAxisLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ChartAxisLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WordCloudFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"word_cloud_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WordCloudAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"group_by": dimensionFieldSchema(dimensionsFieldMaxItems10), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"size":     measureFieldSchema(1),                           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WordCloudSortConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category_items_limit": itemsLimitConfigurationSchema(),                     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":        fieldSortOptionsSchema(fieldSortOptionsMaxItems100), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
									},
								},
							},
							"word_cloud_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WordCloudOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"cloud_layout":          stringSchema(false, validation.StringInSlice(quicksight.WordCloudCloudLayout_Values(), false)),
										"maximum_string_length": intSchema(false, validation.IntBetween(1, 100)),
										"word_casing":           stringSchema(false, validation.StringInSlice(quicksight.WordCloudWordCasing_Values(), false)),
										"word_orientation":      stringSchema(false, validation.StringInSlice(quicksight.WordCloudWordOrientation_Values(), false)),
										"word_padding":          stringSchema(false, validation.StringInSlice(quicksight.WordCloudWordPadding_Values(), false)),
										"word_scaling":          stringSchema(false, validation.StringInSlice(quicksight.WordCloudWordScaling_Values(), false)),
									},
								},
							},
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

func expandWordCloudVisual(tfList []interface{}) *quicksight.WordCloudVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.WordCloudVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.ChartConfiguration = expandWordCloudChartConfiguration(v)
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

func expandWordCloudChartConfiguration(tfList []interface{}) *quicksight.WordCloudChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.WordCloudChartConfiguration{}

	if v, ok := tfMap["category_label_options"].([]interface{}); ok && len(v) > 0 {
		config.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		config.FieldWells = expandWordCloudFieldWells(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SortConfiguration = expandWordCloudSortConfiguration(v)
	}
	if v, ok := tfMap["word_cloud_options"].([]interface{}); ok && len(v) > 0 {
		config.WordCloudOptions = expandWordCloudOptions(v)
	}

	return config
}

func expandWordCloudFieldWells(tfList []interface{}) *quicksight.WordCloudFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.WordCloudFieldWells{}

	if v, ok := tfMap["word_cloud_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.WordCloudAggregatedFieldWells = expandWordCloudAggregatedFieldWells(v)
	}

	return config
}

func expandWordCloudAggregatedFieldWells(tfList []interface{}) *quicksight.WordCloudAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.WordCloudAggregatedFieldWells{}

	if v, ok := tfMap["group_by"].([]interface{}); ok && len(v) > 0 {
		config.GroupBy = expandDimensionFields(v)
	}
	if v, ok := tfMap["size"].([]interface{}); ok && len(v) > 0 {
		config.Size = expandMeasureFields(v)
	}

	return config
}

func expandWordCloudSortConfiguration(tfList []interface{}) *quicksight.WordCloudSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.WordCloudSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandWordCloudOptions(tfList []interface{}) *quicksight.WordCloudOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.WordCloudOptions{}

	if v, ok := tfMap["cloud_layout"].(string); ok && v != "" {
		options.CloudLayout = aws.String(v)
	}
	if v, ok := tfMap["maximum_string_length"].(int64); ok {
		options.MaximumStringLength = aws.Int64(v)
	}
	if v, ok := tfMap["word_casing"].(string); ok && v != "" {
		options.WordCasing = aws.String(v)
	}
	if v, ok := tfMap["word_orientation"].(string); ok && v != "" {
		options.WordOrientation = aws.String(v)
	}
	if v, ok := tfMap["word_padding"].(string); ok && v != "" {
		options.WordPadding = aws.String(v)
	}
	if v, ok := tfMap["word_padding"].(string); ok && v != "" {
		options.WordScaling = aws.String(v)
	}

	return options
}
