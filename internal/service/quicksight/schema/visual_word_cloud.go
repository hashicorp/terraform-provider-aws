// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
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
										"cloud_layout":          stringSchema(false, enum.Validate[types.WordCloudCloudLayout]()),
										"maximum_string_length": intSchema(false, validation.IntBetween(1, 100)),
										"word_casing":           stringSchema(false, enum.Validate[types.WordCloudWordCasing]()),
										"word_orientation":      stringSchema(false, enum.Validate[types.WordCloudWordOrientation]()),
										"word_padding":          stringSchema(false, enum.Validate[types.WordCloudWordPadding]()),
										"word_scaling":          stringSchema(false, enum.Validate[types.WordCloudWordScaling]()),
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

func expandWordCloudVisual(tfList []interface{}) *types.WordCloudVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &types.WordCloudVisual{}

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

func expandWordCloudChartConfiguration(tfList []interface{}) *types.WordCloudChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.WordCloudChartConfiguration{}

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

func expandWordCloudFieldWells(tfList []interface{}) *types.WordCloudFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.WordCloudFieldWells{}

	if v, ok := tfMap["word_cloud_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		config.WordCloudAggregatedFieldWells = expandWordCloudAggregatedFieldWells(v)
	}

	return config
}

func expandWordCloudAggregatedFieldWells(tfList []interface{}) *types.WordCloudAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.WordCloudAggregatedFieldWells{}

	if v, ok := tfMap["group_by"].([]interface{}); ok && len(v) > 0 {
		config.GroupBy = expandDimensionFields(v)
	}
	if v, ok := tfMap["size"].([]interface{}); ok && len(v) > 0 {
		config.Size = expandMeasureFields(v)
	}

	return config
}

func expandWordCloudSortConfiguration(tfList []interface{}) *types.WordCloudSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.WordCloudSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]interface{}); ok && len(v) > 0 {
		config.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]interface{}); ok && len(v) > 0 {
		config.CategorySort = expandFieldSortOptionsList(v)
	}

	return config
}

func expandWordCloudOptions(tfList []interface{}) *types.WordCloudOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.WordCloudOptions{}

	if v, ok := tfMap["cloud_layout"].(string); ok && v != "" {
		options.CloudLayout = types.WordCloudCloudLayout(v)
	}
	if v, ok := tfMap["maximum_string_length"].(int); ok {
		options.MaximumStringLength = aws.Int32(int32(v))
	}
	if v, ok := tfMap["word_casing"].(string); ok && v != "" {
		options.WordCasing = types.WordCloudWordCasing(v)
	}
	if v, ok := tfMap["word_orientation"].(string); ok && v != "" {
		options.WordOrientation = types.WordCloudWordOrientation(v)
	}
	if v, ok := tfMap["word_padding"].(string); ok && v != "" {
		options.WordPadding = types.WordCloudWordPadding(v)
	}
	if v, ok := tfMap["word_padding"].(string); ok && v != "" {
		options.WordScaling = types.WordCloudWordScaling(v)
	}

	return options
}

func flattenWordCloudVisual(apiObject *types.WordCloudVisual) []interface{} {
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
		tfMap["chart_configuration"] = flattenWordCloudChartConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.ColumnHierarchies != nil {
		tfMap["column_hierarchies"] = flattenColumnHierarchy(apiObject.ColumnHierarchies)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []interface{}{tfMap}
}

func flattenWordCloudChartConfiguration(apiObject *types.WordCloudChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryLabelOptions != nil {
		tfMap["category_label_options"] = flattenChartAxisLabelOptions(apiObject.CategoryLabelOptions)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenWordCloudFieldWells(apiObject.FieldWells)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenWordCloudSortConfiguration(apiObject.SortConfiguration)
	}
	if apiObject.WordCloudOptions != nil {
		tfMap["word_cloud_options"] = flattenWordCloudOptions(apiObject.WordCloudOptions)
	}

	return []interface{}{tfMap}
}

func flattenWordCloudFieldWells(apiObject *types.WordCloudFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.WordCloudAggregatedFieldWells != nil {
		tfMap["word_cloud_aggregated_field_wells"] = flattenWordCloudAggregatedFieldWells(apiObject.WordCloudAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenWordCloudAggregatedFieldWells(apiObject *types.WordCloudAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.GroupBy != nil {
		tfMap["group_by"] = flattenDimensionFields(apiObject.GroupBy)
	}
	if apiObject.Size != nil {
		tfMap["size"] = flattenMeasureFields(apiObject.Size)
	}

	return []interface{}{tfMap}
}

func flattenWordCloudSortConfiguration(apiObject *types.WordCloudSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryItemsLimit != nil {
		tfMap["category_items_limit"] = flattenItemsLimitConfiguration(apiObject.CategoryItemsLimit)
	}
	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}

	return []interface{}{tfMap}
}

func flattenWordCloudOptions(apiObject *types.WordCloudOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["cloud_layout"] = types.WordCloudCloudLayout(apiObject.CloudLayout)
	if apiObject.MaximumStringLength != nil {
		tfMap["maximum_string_length"] = aws.ToInt32(apiObject.MaximumStringLength)
	}
	tfMap["word_casing"] = types.WordCloudWordCasing(apiObject.WordCasing)
	tfMap["word_orientation"] = types.WordCloudWordOrientation(apiObject.WordOrientation)
	tfMap["word_padding"] = types.WordCloudWordPadding(apiObject.WordPadding)
	tfMap["word_scaling"] = types.WordCloudWordScaling(apiObject.WordScaling)

	return []interface{}{tfMap}
}
