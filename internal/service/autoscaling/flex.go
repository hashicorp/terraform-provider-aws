// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
)

// Takes the result of flatmap.Expand for an array of step adjustments and
// returns a []*awstypes.StepAdjustment.
func ExpandStepAdjustments(configured []interface{}) ([]awstypes.StepAdjustment, error) {
	var adjustments []awstypes.StepAdjustment

	// Loop over our configured step adjustments and create an array
	// of aws-sdk-go compatible objects. We're forced to convert strings
	// to floats here because there's no way to detect whether or not
	// an uninitialized, optional schema element is "0.0" deliberately.
	// With strings, we can test for "", which is definitely an empty
	// struct value.
	for _, raw := range configured {
		data := raw.(map[string]interface{})
		a := awstypes.StepAdjustment{
			ScalingAdjustment: aws.Int32(int32(data["scaling_adjustment"].(int))),
		}
		if data["metric_interval_lower_bound"] != "" {
			bound := data["metric_interval_lower_bound"]
			switch bound := bound.(type) {
			case string:
				f, err := strconv.ParseFloat(bound, 64)
				if err != nil {
					return nil, fmt.Errorf(
						"metric_interval_lower_bound must be a float value represented as a string")
				}
				a.MetricIntervalLowerBound = aws.Float64(f)
			default:
				return nil, fmt.Errorf(
					"metric_interval_lower_bound isn't a string. This is a bug. Please file an issue.")
			}
		}
		if data["metric_interval_upper_bound"] != "" {
			bound := data["metric_interval_upper_bound"]
			switch bound := bound.(type) {
			case string:
				f, err := strconv.ParseFloat(bound, 64)
				if err != nil {
					return nil, fmt.Errorf(
						"metric_interval_upper_bound must be a float value represented as a string")
				}
				a.MetricIntervalUpperBound = aws.Float64(f)
			default:
				return nil, fmt.Errorf(
					"metric_interval_upper_bound isn't a string. This is a bug. Please file an issue.")
			}
		}
		adjustments = append(adjustments, a)
	}

	return adjustments, nil
}

// Flattens step adjustments into a list of map[string]interface.
func FlattenStepAdjustments(adjustments []awstypes.StepAdjustment) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(adjustments))
	for _, raw := range adjustments {
		a := map[string]interface{}{
			"scaling_adjustment": aws.ToInt32(raw.ScalingAdjustment),
		}
		if raw.MetricIntervalUpperBound != nil {
			a["metric_interval_upper_bound"] = fmt.Sprintf("%g", aws.ToFloat64(raw.MetricIntervalUpperBound))
		}
		if raw.MetricIntervalLowerBound != nil {
			a["metric_interval_lower_bound"] = fmt.Sprintf("%g", aws.ToFloat64(raw.MetricIntervalLowerBound))
		}
		result = append(result, a)
	}
	return result
}
