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

func sankeyDiagramVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SankeyDiagramVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SankeyDiagramChartConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_labels": dataLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataLabelOptions.html
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SankeyDiagramFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"sankey_diagram_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SankeyDiagramAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrDestination: dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrSource:      dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrWeight:      measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SankeyDiagramSortConfiguration.html
								Type:             schema.TypeList,
								Optional:         true,
								MinItems:         1,
								MaxItems:         1,
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"destination_items_limit": itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"source_items_limit":      itemsLimitConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ItemsLimitConfiguration.html
										"weight_sort":             fieldSortOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
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

func expandSankeyDiagramVisual(tfList []interface{}) *awstypes.SankeyDiagramVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SankeyDiagramVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandSankeyDiagramConfiguration(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandSankeyDiagramConfiguration(tfList []interface{}) *awstypes.SankeyDiagramChartConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SankeyDiagramChartConfiguration{}

	if v, ok := tfMap["data_labels"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataLabels = expandDataLabelOptions(v)
	}
	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		apiObject.FieldWells = expandSankeyDiagramFieldWells(v)
	}
	if v, ok := tfMap["sort_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.SortConfiguration = expandSankeyDiagramSortConfiguration(v)
	}

	return apiObject
}

func expandSankeyDiagramFieldWells(tfList []interface{}) *awstypes.SankeyDiagramFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SankeyDiagramFieldWells{}

	if v, ok := tfMap["sankey_diagram_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		apiObject.SankeyDiagramAggregatedFieldWells = expandSankeyDiagramAggregatedFieldWells(v)
	}

	return apiObject
}

func expandSankeyDiagramAggregatedFieldWells(tfList []interface{}) *awstypes.SankeyDiagramAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SankeyDiagramAggregatedFieldWells{}

	if v, ok := tfMap[names.AttrDestination].([]interface{}); ok && len(v) > 0 {
		apiObject.Destination = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrSource].([]interface{}); ok && len(v) > 0 {
		apiObject.Source = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrWeight].([]interface{}); ok && len(v) > 0 {
		apiObject.Weight = expandMeasureFields(v)
	}

	return apiObject
}

func expandSankeyDiagramSortConfiguration(tfList []interface{}) *awstypes.SankeyDiagramSortConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SankeyDiagramSortConfiguration{}

	if v, ok := tfMap["destination_items_limit"].([]interface{}); ok && len(v) > 0 {
		apiObject.DestinationItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["source_items_limit"].([]interface{}); ok && len(v) > 0 {
		apiObject.SourceItemsLimit = expandItemsLimitConfiguration(v)
	}
	if v, ok := tfMap["weight_sort"].([]interface{}); ok && len(v) > 0 {
		apiObject.WeightSort = expandFieldSortOptionsList(v)
	}

	return apiObject
}

func flattenSankeyDiagramVisual(apiObject *awstypes.SankeyDiagramVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id": aws.ToString(apiObject.VisualId),
	}

	if apiObject.Actions != nil {
		tfMap[names.AttrActions] = flattenVisualCustomAction(apiObject.Actions)
	}
	if apiObject.ChartConfiguration != nil {
		tfMap["chart_configuration"] = flattenSankeyDiagramChartConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []interface{}{tfMap}
}

func flattenSankeyDiagramChartConfiguration(apiObject *awstypes.SankeyDiagramChartConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DataLabels != nil {
		tfMap["data_labels"] = flattenDataLabelOptions(apiObject.DataLabels)
	}
	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenSankeyDiagramFieldWells(apiObject.FieldWells)
	}
	if apiObject.SortConfiguration != nil {
		tfMap["sort_configuration"] = flattenSankeyDiagramSortConfiguration(apiObject.SortConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenSankeyDiagramFieldWells(apiObject *awstypes.SankeyDiagramFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SankeyDiagramAggregatedFieldWells != nil {
		tfMap["sankey_diagram_aggregated_field_wells"] = flattenSankeyDiagramAggregatedFieldWells(apiObject.SankeyDiagramAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenSankeyDiagramAggregatedFieldWells(apiObject *awstypes.SankeyDiagramAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Destination != nil {
		tfMap[names.AttrDestination] = flattenDimensionFields(apiObject.Destination)
	}
	if apiObject.Source != nil {
		tfMap[names.AttrSource] = flattenDimensionFields(apiObject.Source)
	}
	if apiObject.Weight != nil {
		tfMap[names.AttrWeight] = flattenMeasureFields(apiObject.Weight)
	}

	return []interface{}{tfMap}
}

func flattenSankeyDiagramSortConfiguration(apiObject *awstypes.SankeyDiagramSortConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DestinationItemsLimit != nil {
		tfMap["destination_items_limit"] = flattenItemsLimitConfiguration(apiObject.DestinationItemsLimit)
	}
	if apiObject.SourceItemsLimit != nil {
		tfMap["source_items_limit"] = flattenItemsLimitConfiguration(apiObject.SourceItemsLimit)
	}
	if apiObject.WeightSort != nil {
		tfMap["weight_sort"] = flattenFieldSortOptions(apiObject.WeightSort)
	}

	return []interface{}{tfMap}
}
