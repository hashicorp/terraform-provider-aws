// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func wordCloudVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_WordCloudVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
													"group_by":     dimensionFieldSchema(dimensionsFieldMaxItems10), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrSize: measureFieldSchema(1),                           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
										"category_items_limit": itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"category_sort":        fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
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
										"cloud_layout":          stringEnumSchema[awstypes.WordCloudCloudLayout](attrOptional),
										"maximum_string_length": intBetweenSchema(attrOptional, 1, 100),
										"word_casing":           stringEnumSchema[awstypes.WordCloudWordCasing](attrOptional),
										"word_orientation":      stringEnumSchema[awstypes.WordCloudWordOrientation](attrOptional),
										"word_padding":          stringEnumSchema[awstypes.WordCloudWordPadding](attrOptional),
										"word_scaling":          stringEnumSchema[awstypes.WordCloudWordScaling](attrOptional),
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

func expandWordCloudVisual(tfList []any) *awstypes.WordCloudVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.WordCloudVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]any); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandWordCloudChartConfiguration(v)
	}
	if v, ok := tfMap["column_hierarchies"].([]any); ok && len(v) > 0 {
		apiObject.ColumnHierarchies = expandColumnHierarchies(v)
	}
	if v, ok := tfMap["subtitle"].([]any); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandWordCloudChartConfiguration(tfList []any) *awstypes.WordCloudChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.WordCloudChartConfiguration{}

	if v, ok := tfMap["category_label_options"].([]any); ok && len(v) > 0 {
		apiObject.CategoryLabelOptions = expandChartAxisLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]any); ok && len(v) > 0 {
		apiObject.FieldWells = expandWordCloudFieldWells(v)
	}
	if v, ok := tfMap["sort_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandWordCloudSortConfiguration(v)
	}
	if v, ok := tfMap["word_cloud_options"].([]any); ok && len(v) > 0 {
		apiObject.WordCloudOptions = expandWordCloudOptions(v)
	}

	return apiObject
}

func expandWordCloudFieldWells(tfList []any) *awstypes.WordCloudFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.WordCloudFieldWells{}

	if v, ok := tfMap["word_cloud_aggregated_field_wells"].([]any); ok && len(v) > 0 {
		apiObject.WordCloudAggregatedFieldWells = expandWordCloudAggregatedFieldWells(v)
	}

	return apiObject
}

func expandWordCloudAggregatedFieldWells(tfList []any) *awstypes.WordCloudAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.WordCloudAggregatedFieldWells{}

	if v, ok := tfMap["group_by"].([]any); ok && len(v) > 0 {
		apiObject.GroupBy = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrSize].([]any); ok && len(v) > 0 {
		apiObject.Size = expandMeasureFields(v)
	}

	return apiObject
}

func expandWordCloudSortConfiguration(tfList []any) *awstypes.WordCloudSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.WordCloudSortConfiguration{}

	if v, ok := tfMap["category_items_limit"].([]any); ok && len(v) > 0 {
		apiObject.CategoryItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["category_sort"].([]any); ok && len(v) > 0 {
		apiObject.CategorySort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func expandWordCloudOptions(tfList []any) *awstypes.WordCloudOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.WordCloudOptions{}

	if v, ok := tfMap["cloud_layout"].(string); ok && v != "" {
		apiObject.CloudLayout = awstypes.WordCloudCloudLayout(v)
	}
	if v, ok := tfMap["maximum_string_length"].(int); ok {
		apiObject.MaximumStringLength = aws.Int32(int32(v))
	}
	if v, ok := tfMap["word_casing"].(string); ok && v != "" {
		apiObject.WordCasing = awstypes.WordCloudWordCasing(v)
	}
	if v, ok := tfMap["word_orientation"].(string); ok && v != "" {
		apiObject.WordOrientation = awstypes.WordCloudWordOrientation(v)
	}
	if v, ok := tfMap["word_padding"].(string); ok && v != "" {
		apiObject.WordPadding = awstypes.WordCloudWordPadding(v)
	}
	if v, ok := tfMap["word_padding"].(string); ok && v != "" {
		apiObject.WordScaling = awstypes.WordCloudWordScaling(v)
	}

	return apiObject
}

func flattenWordCloudVisual(apiObject *awstypes.WordCloudVisual) []any {
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

	return []any{tfMap}
}

func flattenWordCloudChartConfiguration(apiObject *awstypes.WordCloudChartConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}

func flattenWordCloudFieldWells(apiObject *awstypes.WordCloudFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}
	if apiObject.WordCloudAggregatedFieldWells != nil {
		tfMap["word_cloud_aggregated_field_wells"] = flattenWordCloudAggregatedFieldWells(apiObject.WordCloudAggregatedFieldWells)
	}

	return []any{tfMap}
}

func flattenWordCloudAggregatedFieldWells(apiObject *awstypes.WordCloudAggregatedFieldWells) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.GroupBy != nil {
		tfMap["group_by"] = flattenDimensionFields(apiObject.GroupBy)
	}
	if apiObject.Size != nil {
		tfMap[names.AttrSize] = flattenMeasureFields(apiObject.Size)
	}

	return []any{tfMap}
}

func flattenWordCloudSortConfiguration(apiObject *awstypes.WordCloudSortConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CategoryItemsLimit != nil {
		tfMap["category_items_limit"] = flattenItemsLimitConfiguration(apiObject.CategoryItemsLimit)
	}
	if apiObject.CategorySort != nil {
		tfMap["category_sort"] = flattenFieldSortOptions(apiObject.CategorySort)
	}

	return []any{tfMap}
}

func flattenWordCloudOptions(apiObject *awstypes.WordCloudOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["cloud_layout"] = apiObject.CloudLayout
	if apiObject.MaximumStringLength != nil {
		tfMap["maximum_string_length"] = aws.ToInt32(apiObject.MaximumStringLength)
	}
	tfMap["word_casing"] = apiObject.WordCasing
	tfMap["word_orientation"] = apiObject.WordOrientation
	tfMap["word_padding"] = apiObject.WordPadding
	tfMap["word_scaling"] = apiObject.WordScaling

	return []any{tfMap}
}
