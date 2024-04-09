// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
)

func expandStepAdjustments(tfList []interface{}) []awstypes.StepAdjustment {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.StepAdjustment

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.StepAdjustment{
			ScalingAdjustment: aws.Int32(int32(tfMap["scaling_adjustment"].(int))),
		}

		if v, ok := tfMap["metric_interval_lower_bound"].(string); ok {
			if v, null, _ := nullable.Float(v).Value(); !null {
				apiObject.MetricIntervalLowerBound = aws.Float64(v)
			}
		}

		if v, ok := tfMap["metric_interval_upper_bound"].(string); ok {
			if v, null, _ := nullable.Float(v).Value(); !null {
				apiObject.MetricIntervalUpperBound = aws.Float64(v)
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenStepAdjustments(apiObjects []awstypes.StepAdjustment) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"scaling_adjustment": aws.ToInt32(apiObject.ScalingAdjustment),
		}

		if v := apiObject.MetricIntervalUpperBound; v != nil {
			tfMap["metric_interval_upper_bound"] = flex.Float64ToStringValue(v)
		}

		if v := apiObject.MetricIntervalLowerBound; v != nil {
			tfMap["metric_interval_lower_bound"] = flex.Float64ToStringValue(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
