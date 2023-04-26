package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				"actions":             visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
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
													"name": {
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
													"value": measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"period_size": intSchema(false, validation.IntBetween(2, 52)),
													"value":       measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"type":           stringSchema(true, validation.StringInSlice(quicksight.MaximumMinimumComputationType_Values(), false)),
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"name": {
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
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"period_time_granularity": stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
													"value":                   measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"type":           stringSchema(true, validation.StringInSlice(quicksight.TopBottomComputationType_Values(), false)),
													"mover_size":     intSchema(false, validation.IntBetween(1, 20)),
													"sort_order":     stringSchema(true, validation.StringInSlice(quicksight.TopBottomSortOrder_Values(), false)),
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"result_size": intSchema(false, validation.IntBetween(1, 20)),
													"type":        stringSchema(true, validation.StringInSlice(quicksight.TopBottomComputationType_Values(), false)),
													"value":       measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": measureFieldSchema(1), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
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
													"name": {
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
	if v, ok := tfMap["actions"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["custom_seasonality_value"].(int64); ok {
		computation.CustomSeasonalityValue = aws.Int64(v)
	}
	if v, ok := tfMap["lower_boundary"].(float64); ok {
		computation.LowerBoundary = aws.Float64(v)
	}
	if v, ok := tfMap["name"].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["periods_backward"].(int64); ok {
		computation.PeriodsBackward = aws.Int64(v)
	}
	if v, ok := tfMap["periods_forward"].(int64); ok {
		computation.PeriodsForward = aws.Int64(v)
	}
	if v, ok := tfMap["prediction_interval"].(int64); ok {
		computation.PredictionInterval = aws.Int64(v)
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
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["period_size"].(int64); ok {
		computation.PeriodSize = aws.Int64(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["type"].(string); ok && v != "" {
		computation.Type = aws.String(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["period_time_granularity"].(string); ok && v != "" {
		computation.PeriodTimeGranularity = aws.String(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["sort_order"].(string); ok && v != "" {
		computation.SortOrder = aws.String(v)
	}
	if v, ok := tfMap["type"].(string); ok && v != "" {
		computation.Type = aws.String(v)
	}
	if v, ok := tfMap["mover_size"].(int64); ok {
		computation.MoverSize = aws.Int64(v)
	}
	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		computation.Category = expandDimensionField(v)
	}
	if v, ok := tfMap["time"].([]interface{}); ok && len(v) > 0 {
		computation.Time = expandDimensionField(v)
	}
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["type"].(string); ok && v != "" {
		computation.Type = aws.String(v)
	}
	if v, ok := tfMap["result_size"].(int64); ok {
		computation.ResultSize = aws.Int64(v)
	}
	if v, ok := tfMap["category"].([]interface{}); ok && len(v) > 0 {
		computation.Category = expandDimensionField(v)
	}
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
		computation.Name = aws.String(v)
	}
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := tfMap["name"].(string); ok && v != "" {
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
