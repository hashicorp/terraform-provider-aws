// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				"data_set_identifier": stringSchema(true, validation.StringLenBetween(1, 2048)),
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
													"custom_seasonality_value": intSchema(false, validation.IntBetween(1, 180)),
													"lower_boundary": {
														Type:     schema.TypeFloat,
														Optional: true,
													},
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													"periods_backward":    intSchema(false, validation.IntBetween(0, 1000)),
													"periods_forward":     intSchema(false, validation.IntBetween(1, 1000)),
													"prediction_interval": intSchema(false, validation.IntBetween(50, 95)),
													"seasonality":         stringSchema(true, validation.StringInSlice(quicksight.ForecastComputationSeasonality_Values(), false)),
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
													"period_size":   intSchema(false, validation.IntBetween(2, 52)),
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
													names.AttrType:   stringSchema(true, validation.StringInSlice(quicksight.MaximumMinimumComputationType_Values(), false)),
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
													"period_time_granularity": stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
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
													names.AttrType:   stringSchema(true, validation.StringInSlice(quicksight.TopBottomComputationType_Values(), false)),
													"mover_size":     intSchema(false, validation.IntBetween(1, 20)),
													"sort_order":     stringSchema(true, validation.StringInSlice(quicksight.TopBottomSortOrder_Values(), false)),
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
													"result_size":   intSchema(false, validation.IntBetween(1, 20)),
													names.AttrType:  stringSchema(true, validation.StringInSlice(quicksight.TopBottomComputationType_Values(), false)),
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
										"narrative": stringSchema(true, validation.StringLenBetween(1, 150000)),
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

func expandInsightVisual(tfList []interface{}) *quicksight.InsightVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	visual := &quicksight.InsightVisual{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		visual.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		visual.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		visual.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["insight_configuration"].([]interface{}); ok && len(v) > 0 {
		visual.InsightConfiguration = expandInsightConfiguration(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		visual.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		visual.Title = expandVisualTitleLabelOptions(v)
	}

	return visual
}

func expandInsightConfiguration(tfList []interface{}) *quicksight.InsightConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.InsightConfiguration{}

	if v, ok := tfMap["computation"].([]interface{}); ok && len(v) > 0 {
		config.Computations = expandComputations(v)
	}
	if v, ok := tfMap["custom_narrative"].([]interface{}); ok && len(v) > 0 {
		config.CustomNarrative = expandCustomNarrativeOptions(v)
	}

	return config
}

func expandComputations(tfList []interface{}) []*quicksight.Computation {
	if len(tfList) == 0 {
		return nil
	}

	var computations []*quicksight.Computation
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		computation := expandComputation(tfMap)
		if computation == nil {
			continue
		}

		computations = append(computations, computation)
	}

	return computations
}

func expandComputation(tfMap map[string]interface{}) *quicksight.Computation {
	if tfMap == nil {
		return nil
	}

	computation := &quicksight.Computation{}

	if v, ok := tfMap["forecast"].([]interface{}); ok && len(v) > 0 {
		computation.Forecast = expandForecastComputation(v)
	}
	if v, ok := tfMap["growth_rate"].([]interface{}); ok && len(v) > 0 {
		computation.GrowthRate = expandGrowthRateComputation(v)
	}
	if v, ok := tfMap["maximum_minimum"].([]interface{}); ok && len(v) > 0 {
		computation.MaximumMinimum = expandMaximumMinimumComputation(v)
	}
	if v, ok := tfMap["metric_comparison"].([]interface{}); ok && len(v) > 0 {
		computation.MetricComparison = expandMetricComparisonComputation(v)
	}
	if v, ok := tfMap["period_over_period"].([]interface{}); ok && len(v) > 0 {
		computation.PeriodOverPeriod = expandPeriodOverPeriodComputation(v)
	}
	if v, ok := tfMap["period_to_date"].([]interface{}); ok && len(v) > 0 {
		computation.PeriodToDate = expandPeriodToDateComputation(v)
	}
	if v, ok := tfMap["top_bottom_movers"].([]interface{}); ok && len(v) > 0 {
		computation.TopBottomMovers = expandTopBottomMoversComputation(v)
	}
	if v, ok := tfMap["top_bottom_ranked"].([]interface{}); ok && len(v) > 0 {
		computation.TopBottomRanked = expandTopBottomRankedComputation(v)
	}
	if v, ok := tfMap["total_aggregation"].([]interface{}); ok && len(v) > 0 {
		computation.TotalAggregation = expandTotalAggregationComputation(v)
	}
	if v, ok := tfMap["unique_values"].([]interface{}); ok && len(v) > 0 {
		computation.UniqueValues = expandUniqueValuesComputation(v)
	}

	return computation
}

func expandForecastComputation(tfList []interface{}) *quicksight.ForecastComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.ForecastComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap["custom_seasonality_value"].(int); ok {
		computation.CustomSeasonalityValue = aws.Int64(int64(v))
	}
	if v, ok := tfMap["lower_boundary"].(float64); ok {
		computation.LowerBoundary = aws.Float64(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["periods_backward"].(int); ok {
		computation.PeriodsBackward = aws.Int64(int64(v))
	}
	if v, ok := tfMap["periods_forward"].(int); ok {
		computation.PeriodsForward = aws.Int64(int64(v))
	}
	if v, ok := tfMap["prediction_interval"].(int); ok {
		computation.PredictionInterval = aws.Int64(int64(v))
	}
	if v, ok := tfMap["seasonality"].(string); ok && v != "" {
		computation.Seasonality = aws.String(v)
	}
	if v, ok := tfMap["upper_boundary"].(float64); ok {
		computation.UpperBoundary = aws.Float64(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		computation.Value = expandMeasureField(v)
	}

	return computation
}

func expandGrowthRateComputation(tfList []interface{}) *quicksight.GrowthRateComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.GrowthRateComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["period_size"].(int); ok {
		computation.PeriodSize = aws.Int64(int64(v))
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		computation.Value = expandMeasureField(v)
	}

	return computation
}

func expandMaximumMinimumComputation(tfList []interface{}) *quicksight.MaximumMinimumComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.MaximumMinimumComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		computation.Type = aws.String(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		computation.Value = expandMeasureField(v)
	}

	return computation
}

func expandMetricComparisonComputation(tfList []interface{}) *quicksight.MetricComparisonComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.MetricComparisonComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap["from_value"].([]interface{}); ok && len(v) > 0 {
		computation.FromValue = expandMeasureField(v)
	}
	if v, ok := tfMap["target_value"].([]interface{}); ok && len(v) > 0 {
		computation.TargetValue = expandMeasureField(v)
	}

	return computation
}

func expandPeriodOverPeriodComputation(tfList []interface{}) *quicksight.PeriodOverPeriodComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.PeriodOverPeriodComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		computation.Value = expandMeasureField(v)
	}

	return computation
}

func expandPeriodToDateComputation(tfList []interface{}) *quicksight.PeriodToDateComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.PeriodToDateComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["period_time_granularity"].(string); ok && v != "" {
		computation.PeriodTimeGranularity = aws.String(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		computation.Value = expandMeasureField(v)
	}

	return computation
}

func expandTopBottomMoversComputation(tfList []interface{}) *quicksight.TopBottomMoversComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.TopBottomMoversComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["sort_order"].(string); ok && v != "" {
		computation.SortOrder = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		computation.Type = aws.String(v)
	}
	if v, ok := tfMap["mover_size"].(int); ok {
		computation.MoverSize = aws.Int64(int64(v))
	}
	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		computation.Category = expandDimensionField(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		computation.Value = expandMeasureField(v)
	}

	return computation
}

func expandTopBottomRankedComputation(tfList []interface{}) *quicksight.TopBottomRankedComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.TopBottomRankedComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		computation.Type = aws.String(v)
	}
	if v, ok := tfMap["result_size"].(int); ok {
		computation.ResultSize = aws.Int64(int64(v))
	}
	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		computation.Category = expandDimensionField(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		computation.Value = expandMeasureField(v)
	}

	return computation
}

func expandTotalAggregationComputation(tfList []interface{}) *quicksight.TotalAggregationComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.TotalAggregationComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		computation.Value = expandMeasureField(v)
	}

	return computation
}

func expandUniqueValuesComputation(tfList []interface{}) *quicksight.UniqueValuesComputation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	computation := &quicksight.UniqueValuesComputation{}

	if v, ok := tfMap["computation_id"].(string); ok && v != "" {
		computation.ComputationId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		computation.Category = expandDimensionField(v)
	}

	return computation
}

func expandCustomNarrativeOptions(tfList []interface{}) *quicksight.CustomNarrativeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.CustomNarrativeOptions{}

	if v, ok := tfMap["narrative"].(string); ok && v != "" {
		options.Narrative = aws.String(v)
	}

	return options
}

func flattenInsightVisual(apiObject *quicksight.InsightVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id":           aws.StringValue(apiObject.VisualId),
		"data_set_identifier": aws.StringValue(apiObject.DataSetIdentifier),
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

	return []interface{}{tfMap}
}

func flattenInsightConfiguration(apiObject *quicksight.InsightConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Computations != nil {
		tfMap["computation"] = flattenComputation(apiObject.Computations)
	}
	if apiObject.CustomNarrative != nil {
		tfMap["custom_narrative"] = flattenCustomNarrativeOptions(apiObject.CustomNarrative)
	}

	return []interface{}{tfMap}
}

func flattenComputation(apiObject []*quicksight.Computation) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.Forecast != nil {
			tfMap["forecast"] = flattenForecastComputation(config.Forecast)
		}
		if config.GrowthRate != nil {
			tfMap["growth_rate"] = flattenGrowthRateComputation(config.GrowthRate)
		}
		if config.MaximumMinimum != nil {
			tfMap["maximum_minimum"] = flattenMaximumMinimumComputation(config.MaximumMinimum)
		}
		if config.MetricComparison != nil {
			tfMap["metric_comparison"] = flattenMetricComparisonComputation(config.MetricComparison)
		}
		if config.PeriodOverPeriod != nil {
			tfMap["period_over_period"] = flattenPeriodOverPeriodComputation(config.PeriodOverPeriod)
		}
		if config.PeriodToDate != nil {
			tfMap["period_to_date"] = flattenPeriodToDateComputation(config.PeriodToDate)
		}
		if config.TopBottomMovers != nil {
			tfMap["top_bottom_movers"] = flattenTopBottomMoversComputation(config.TopBottomMovers)
		}
		if config.TopBottomRanked != nil {
			tfMap["top_bottom_ranked"] = flattenTopBottomRankedComputation(config.TopBottomRanked)
		}
		if config.TotalAggregation != nil {
			tfMap["total_aggregation"] = flattenTotalAggregationComputation(config.TotalAggregation)
		}
		if config.UniqueValues != nil {
			tfMap["unique_values"] = flattenUniqueValuesComputation(config.UniqueValues)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenForecastComputation(apiObject *quicksight.ForecastComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.CustomSeasonalityValue != nil {
		tfMap["custom_seasonality_value"] = aws.Int64Value(apiObject.CustomSeasonalityValue)
	}
	if apiObject.LowerBoundary != nil {
		tfMap["lower_boundary"] = aws.Float64Value(apiObject.LowerBoundary)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.PeriodsBackward != nil {
		tfMap["periods_backward"] = aws.Int64Value(apiObject.PeriodsBackward)
	}
	if apiObject.PeriodsForward != nil {
		tfMap["periods_forward"] = aws.Int64Value(apiObject.PeriodsForward)
	}
	if apiObject.PredictionInterval != nil {
		tfMap["prediction_interval"] = aws.Int64Value(apiObject.PredictionInterval)
	}
	if apiObject.Seasonality != nil {
		tfMap["seasonality"] = aws.StringValue(apiObject.Seasonality)
	}
	if apiObject.UpperBoundary != nil {
		tfMap["upper_boundary"] = aws.Float64Value(apiObject.UpperBoundary)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenGrowthRateComputation(apiObject *quicksight.GrowthRateComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.PeriodSize != nil {
		tfMap["period_size"] = aws.Int64Value(apiObject.PeriodSize)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenMaximumMinimumComputation(apiObject *quicksight.MaximumMinimumComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenMetricComparisonComputation(apiObject *quicksight.MetricComparisonComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
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
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}

	return []interface{}{tfMap}
}

func flattenPeriodOverPeriodComputation(apiObject *quicksight.PeriodOverPeriodComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenPeriodToDateComputation(apiObject *quicksight.PeriodToDateComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.PeriodTimeGranularity != nil {
		tfMap["period_time_granularity"] = aws.StringValue(apiObject.PeriodTimeGranularity)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenTopBottomMoversComputation(apiObject *quicksight.TopBottomMoversComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionField(apiObject.Category)
	}
	if apiObject.MoverSize != nil {
		tfMap["mover_size"] = aws.Int64Value(apiObject.MoverSize)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.SortOrder != nil {
		tfMap["sort_order"] = aws.StringValue(apiObject.SortOrder)
	}
	if apiObject.Time != nil {
		tfMap["time"] = flattenDimensionField(apiObject.Time)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenTopBottomRankedComputation(apiObject *quicksight.TopBottomRankedComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionField(apiObject.Category)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.ResultSize != nil {
		tfMap["result_size"] = aws.Int64Value(apiObject.ResultSize)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenTotalAggregationComputation(apiObject *quicksight.TotalAggregationComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = flattenMeasureField(apiObject.Value)
	}

	return []interface{}{tfMap}
}

func flattenUniqueValuesComputation(apiObject *quicksight.UniqueValuesComputation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ComputationId != nil {
		tfMap["computation_id"] = aws.StringValue(apiObject.ComputationId)
	}
	if apiObject.Category != nil {
		tfMap["category"] = flattenDimensionField(apiObject.Category)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}

	return []interface{}{tfMap}
}

func flattenCustomNarrativeOptions(apiObject *quicksight.CustomNarrativeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Narrative != nil {
		tfMap["narrative"] = aws.StringValue(apiObject.Narrative)
	}

	return []interface{}{tfMap}
}
