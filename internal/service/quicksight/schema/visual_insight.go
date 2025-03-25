// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func insightVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_InsightVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_identifier": stringLenBetweenSchema(attrRequired, 1, 2048),
				"visual_id":           idSchema(),
				names.AttrActions:     visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"insight_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_InsightConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"computation": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Computation.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 100,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"forecast": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ForecastComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id":           idSchema(),
													"time":                     dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"custom_seasonality_value": intBetweenSchema(attrOptional, 1, 180),
													"lower_boundary": {
														Type:     schema.TypeFloat,
														Optional: true,
													},
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													"periods_backward":    intBetweenSchema(attrOptional, 0, 1000),
													"periods_forward":     intBetweenSchema(attrOptional, 1, 1000),
													"prediction_interval": intBetweenSchema(attrOptional, 50, 95),
													"seasonality":         stringEnumSchema[awstypes.ForecastComputationSeasonality](attrRequired),
													"upper_boundary": {
														Type:     schema.TypeFloat,
														Optional: true,
													},
													names.AttrValue: measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"growth_rate": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GrowthRateComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id": idSchema(),
													"time":           dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													"period_size":   intBetweenSchema(attrOptional, 2, 52),
													names.AttrValue: measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"maximum_minimum": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MaximumMinimumComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id": idSchema(),
													"time":           dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrType:   stringEnumSchema[awstypes.MaximumMinimumComputationType](attrRequired),
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													names.AttrValue: measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"metric_comparison": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MetricComparisonComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id": idSchema(),
													"time":           dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"from_value":     measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
													"target_value":   measureFieldSchema(1),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										"period_over_period": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PeriodOverPeriodComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id": idSchema(),
													"time":           dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													names.AttrValue: measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"period_to_date": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PeriodToDateComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id": idSchema(),
													"time":           dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													"period_time_granularity": stringEnumSchema[awstypes.TimeGranularity](attrRequired),
													names.AttrValue:           measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"top_bottom_movers": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TopBottomMoversComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id": idSchema(),
													"category":       dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"time":           dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrType:   stringEnumSchema[awstypes.TopBottomComputationType](attrRequired),
													"mover_size":     intBetweenSchema(attrOptional, 1, 20),
													"sort_order":     stringEnumSchema[awstypes.TopBottomSortOrder](attrRequired),
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													names.AttrValue: measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"top_bottom_ranked": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TopBottomRankedComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id": idSchema(),
													"category":       dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													"result_size":   intBetweenSchema(attrOptional, 1, 20),
													names.AttrType:  stringEnumSchema[awstypes.TopBottomComputationType](attrRequired),
													names.AttrValue: measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"total_aggregation": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TotalAggregationComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"computation_id": idSchema(),
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													names.AttrValue: measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
										"unique_values": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_UniqueValuesComputation.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"category":       dimensionFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"computation_id": idSchema(),
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"custom_narrative": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomNarrativeOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"narrative": stringLenBetweenSchema(attrRequired, 1, 150000),
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

func expandInsightVisual(tfList []any) *awstypes.InsightVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.InsightVisual{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		apiObject.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["insight_configuration"].([]any); ok && len(v) > 0 {
		apiObject.InsightConfiguration = expandInsightConfiguration(v)
	}
	if v, ok := tfMap["subtitle"].([]any); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]any); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandInsightConfiguration(tfList []any) *awstypes.InsightConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.InsightConfiguration{}

	if v, ok := tfMap["computation"].([]any); ok && len(v) > 0 {
		apiObject.Computations = expandComputations(v)
	}
	if v, ok := tfMap["custom_narrative"].([]any); ok && len(v) > 0 {
		apiObject.CustomNarrative = expandCustomNarrativeOptions(v)
	}

	return apiObject
}

func expandComputations(tfList []any) []awstypes.Computation {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Computation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandComputation(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandComputation(tfMap map[string]any) *awstypes.Computation {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Computation{}

	if v, ok := tfMap["forecast"].([]any); ok && len(v) > 0 {
		apiObject.Forecast = expandForecastComputation(v)
	}
	if v, ok := tfMap["growth_rate"].([]any); ok && len(v) > 0 {
		apiObject.GrowthRate = expandGrowthRateComputation(v)
	}
	if v, ok := tfMap["maximum_minimum"].([]any); ok && len(v) > 0 {
		apiObject.MaximumMinimum = expandMaximumMinimumComputation(v)
	}
	if v, ok := tfMap["metric_comparison"].([]any); ok && len(v) > 0 {
		apiObject.MetricComparison = expandMetricComparisonComputation(v)
	}
	if v, ok := tfMap["period_over_period"].([]any); ok && len(v) > 0 {
		apiObject.PeriodOverPeriod = expandPeriodOverPeriodComputation(v)
	}
	if v, ok := tfMap["period_to_date"].([]any); ok && len(v) > 0 {
		apiObject.PeriodToDate = expandPeriodToDateComputation(v)
	}
	if v, ok := tfMap["top_bottom_movers"].([]any); ok && len(v) > 0 {
		apiObject.TopBottomMovers = expandTopBottomMoversComputation(v)
	}
	if v, ok := tfMap["top_bottom_ranked"].([]any); ok && len(v) > 0 {
		apiObject.TopBottomRanked = expandTopBottomRankedComputation(v)
	}
	if v, ok := tfMap["total_aggregation"].([]any); ok && len(v) > 0 {
		apiObject.TotalAggregation = expandTotalAggregationComputation(v)
	}
	if v, ok := tfMap["unique_values"].([]any); ok && len(v) > 0 {
		apiObject.UniqueValues = expandUniqueValuesComputation(v)
	}

	return apiObject
}

func expandForecastComputation(tfList []any) *awstypes.ForecastComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ForecastComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap["custom_seasonality_value"].(int); ok {
		apiObject.CustomSeasonalityValue = aws.Int32(int32(v))
	}
	if v, ok := tfMap["lower_boundary"].(float64); ok {
		apiObject.LowerBoundary = aws.Float64(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["periods_backward"].(int); ok {
		apiObject.PeriodsBackward = aws.Int32(int32(v))
	}
	if v, ok := tfMap["periods_forward"].(int); ok {
		apiObject.PeriodsForward = aws.Int32(int32(v))
	}
	if v, ok := tfMap["prediction_interval"].(int); ok {
		apiObject.PredictionInterval = aws.Int32(int32(v))
	}
	if v, ok := tfMap["seasonality"].(string); ok && v != "" {
		apiObject.Seasonality = awstypes.ForecastComputationSeasonality(v)
	}
	if v, ok := tfMap["upper_boundary"].(float64); ok {
		apiObject.UpperBoundary = aws.Float64(v)
	}
	if v, ok := tfMap["time"].([]any); ok && len(v) > 0 {
		apiObject.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 {
		apiObject.Value = expandMeasureField(v)
	}

	return apiObject
}

func expandGrowthRateComputation(tfList []any) *awstypes.GrowthRateComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GrowthRateComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["period_size"].(int); ok {
		apiObject.PeriodSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["time"].([]any); ok && len(v) > 0 {
		apiObject.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 {
		apiObject.Value = expandMeasureField(v)
	}

	return apiObject
}

func expandMaximumMinimumComputation(tfList []any) *awstypes.MaximumMinimumComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.MaximumMinimumComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.MaximumMinimumComputationType(v)
	}
	if v, ok := tfMap["time"].([]any); ok && len(v) > 0 {
		apiObject.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 {
		apiObject.Value = expandMeasureField(v)
	}

	return apiObject
}

func expandMetricComparisonComputation(tfList []any) *awstypes.MetricComparisonComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.MetricComparisonComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["time"].([]any); ok && len(v) > 0 {
		apiObject.Time = expandDimensionField(v)
	}
	if v, ok := tfMap["from_value"].([]any); ok && len(v) > 0 {
		apiObject.FromValue = expandMeasureField(v)
	}
	if v, ok := tfMap["target_value"].([]any); ok && len(v) > 0 {
		apiObject.TargetValue = expandMeasureField(v)
	}

	return apiObject
}

func expandPeriodOverPeriodComputation(tfList []any) *awstypes.PeriodOverPeriodComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PeriodOverPeriodComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["time"].([]any); ok && len(v) > 0 {
		apiObject.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 {
		apiObject.Value = expandMeasureField(v)
	}

	return apiObject
}

func expandPeriodToDateComputation(tfList []any) *awstypes.PeriodToDateComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PeriodToDateComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["period_time_granularity"].(string); ok && v != "" {
		apiObject.PeriodTimeGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["time"].([]any); ok && len(v) > 0 {
		apiObject.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 {
		apiObject.Value = expandMeasureField(v)
	}

	return apiObject
}

func expandTopBottomMoversComputation(tfList []any) *awstypes.TopBottomMoversComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TopBottomMoversComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["sort_order"].(string); ok && v != "" {
		apiObject.SortOrder = awstypes.TopBottomSortOrder(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.TopBottomComputationType(v)
	}
	if v, ok := tfMap["mover_size"].(int); ok {
		apiObject.MoverSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionField(v)
	}
	if v, ok := tfMap["time"].([]any); ok && len(v) > 0 {
		apiObject.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 {
		apiObject.Value = expandMeasureField(v)
	}

	return apiObject
}

func expandTopBottomRankedComputation(tfList []any) *awstypes.TopBottomRankedComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TopBottomRankedComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.TopBottomComputationType(v)
	}
	if v, ok := tfMap["result_size"].(int); ok {
		apiObject.ResultSize = aws.Int32(int32(v))
	}
	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 {
		apiObject.Value = expandMeasureField(v)
	}

	return apiObject
}

func expandTotalAggregationComputation(tfList []any) *awstypes.TotalAggregationComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TotalAggregationComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 {
		apiObject.Value = expandMeasureField(v)
	}

	return apiObject
}

func expandUniqueValuesComputation(tfList []any) *awstypes.UniqueValuesComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.UniqueValuesComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		apiObject.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["category"].([]any); ok && len(v) > 0 {
		apiObject.Category = expandDimensionField(v)
	}

	return apiObject
}

func expandCustomNarrativeOptions(tfList []any) *awstypes.CustomNarrativeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomNarrativeOptions{}

	if v, ok := tfMap["narrative"].(string); ok && v != "" {
		apiObject.Narrative = aws.String(v)
	}

	return apiObject
}

func flattenInsightVisual(apiObject *awstypes.InsightVisual) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"visual_id":           aws.ToString(apiObject.VisualId),
		"data_set_identifier": aws.ToString(apiObject.DataSetIdentifier),
	}

	if apiObject.Actions != nil {
		tfMap[names.AttrActions] = flattenVisualCustomAction(apiObject.Actions)
	}
	if apiObject.InsightConfiguration != nil {
		tfMap["insight_configuration"] = flattenInsightConfiguration(apiObject.InsightConfiguration)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []any{tfMap}
}

func flattenInsightConfiguration(apiObject *awstypes.InsightConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Computations != nil {
		tfMap["computation"] = flattenComputation(apiObject.Computations)
	}
	if apiObject.CustomNarrative != nil {
		tfMap["custom_narrative"] = flattenCustomNarrativeOptions(apiObject.CustomNarrative)
	}

	return []any{tfMap}
}

func flattenComputation(apiObjects []awstypes.Computation) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Forecast != nil {
			tfMap["forecast"] = flattenForecastComputation(apiObject.Forecast)
		}
		if apiObject.GrowthRate != nil {
			tfMap["growth_rate"] = flattenGrowthRateComputation(apiObject.GrowthRate)
		}
		if apiObject.MaximumMinimum != nil {
			tfMap["maximum_minimum"] = flattenMaximumMinimumComputation(apiObject.MaximumMinimum)
		}
		if apiObject.MetricComparison != nil {
			tfMap["metric_comparison"] = flattenMetricComparisonComputation(apiObject.MetricComparison)
		}
		if apiObject.PeriodOverPeriod != nil {
			tfMap["period_over_period"] = flattenPeriodOverPeriodComputation(apiObject.PeriodOverPeriod)
		}
		if apiObject.PeriodToDate != nil {
			tfMap["period_to_date"] = flattenPeriodToDateComputation(apiObject.PeriodToDate)
		}
		if apiObject.TopBottomMovers != nil {
			tfMap["top_bottom_movers"] = flattenTopBottomMoversComputation(apiObject.TopBottomMovers)
		}
		if apiObject.TopBottomRanked != nil {
			tfMap["top_bottom_ranked"] = flattenTopBottomRankedComputation(apiObject.TopBottomRanked)
		}
		if apiObject.TotalAggregation != nil {
			tfMap["total_aggregation"] = flattenTotalAggregationComputation(apiObject.TotalAggregation)
		}
		if apiObject.UniqueValues != nil {
			tfMap["unique_values"] = flattenUniqueValuesComputation(apiObject.UniqueValues)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenForecastComputation(apiObject *awstypes.ForecastComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.CustomSeasonalityValue != nil {
		tfMap["custom_seasonality_value"] = aws.ToInt32(apiObject.CustomSeasonalityValue)
	}
	if apiObject.LowerBoundary != nil {
		tfMap["lower_boundary"] = aws.ToFloat64(apiObject.LowerBoundary)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	if apiObject.PeriodsBackward != nil {
		tfMap["periods_backward"] = aws.ToInt32(apiObject.PeriodsBackward)
	}
	if apiObject.PeriodsForward != nil {
		tfMap["periods_forward"] = aws.ToInt32(apiObject.PeriodsForward)
	}
	if apiObject.PredictionInterval != nil {
		tfMap["prediction_interval"] = aws.ToInt32(apiObject.PredictionInterval)
	}
	tfMap["seasonality"] = apiObject.Seasonality
	if apiObject.UpperBoundary != nil {
		tfMap["upper_boundary"] = aws.ToFloat64(apiObject.UpperBoundary)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenGrowthRateComputation(apiObject *awstypes.GrowthRateComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	if apiObject.PeriodSize != nil {
		tfMap["period_size"] = aws.ToInt32(apiObject.PeriodSize)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenMaximumMinimumComputation(apiObject *awstypes.MaximumMinimumComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	tfMap[names.AttrType] = apiObject.Type
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenMetricComparisonComputation(apiObject *awstypes.MetricComparisonComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.FromValue != nil {
		tfMap["from_value"] = flattenMeasureField(apiObject.FromValue)
	}
	if apiObject.TargetValue != nil {
		tfMap["target_value"] = flattenMeasureField(apiObject.TargetValue)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	return []any{tfMap}
}

func flattenPeriodOverPeriodComputation(apiObject *awstypes.PeriodOverPeriodComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenPeriodToDateComputation(apiObject *awstypes.PeriodToDateComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	tfMap["period_time_granularity"] = apiObject.PeriodTimeGranularity
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenTopBottomMoversComputation(apiObject *awstypes.TopBottomMoversComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionField(apiObject.Category)
	}
	if apiObject.MoverSize != nil {
		tfMap["mover_size"] = aws.ToInt32(apiObject.MoverSize)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	tfMap["sort_order"] = apiObject.SortOrder
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	tfMap[names.AttrType] = apiObject.Type
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenTopBottomRankedComputation(apiObject *awstypes.TopBottomRankedComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionField(apiObject.Category)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	if apiObject.ResultSize != nil {
		tfMap["result_size"] = aws.ToInt32(apiObject.ResultSize)
	}
	tfMap[names.AttrType] = apiObject.Type
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenTotalAggregationComputation(apiObject *awstypes.TotalAggregationComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []any{tfMap}
}

func flattenUniqueValuesComputation(apiObject *awstypes.UniqueValuesComputation) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.ToString(apiObject.ComputationId)
	}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionField(apiObject.Category)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	return []any{tfMap}
}

func flattenCustomNarrativeOptions(apiObject *awstypes.CustomNarrativeOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Narrative != nil {
		tfMap["narrative"] = aws.ToString(apiObject.Narrative)
	}

	return []any{tfMap}
}
