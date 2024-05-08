// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func expandParameters(params map[string]interface{}) []awstypes.Parameter {
	var cfParams []awstypes.Parameter
	for k, v := range params {
		cfParams = append(cfParams, awstypes.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v.(string)),
		})
	}

	return cfParams
}

func flattenAllParameters(cfParams []awstypes.Parameter) map[string]interface{} {
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		params[aws.ToString(p.ParameterKey)] = aws.ToString(p.ParameterValue)
	}
	return params
}

func flattenOutputs(cfOutputs []awstypes.Output) map[string]string {
	outputs := make(map[string]string, len(cfOutputs))
	for _, o := range cfOutputs {
		outputs[aws.ToString(o.OutputKey)] = aws.ToString(o.OutputValue)
	}
	return outputs
}

// flattenParameters is flattening list of
// *cloudformation.Parameters and only returning existing
// parameters to avoid clash with default values
func flattenParameters(cfParams []awstypes.Parameter,
	originalParams map[string]interface{}) map[string]interface{} {
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		_, isConfigured := originalParams[aws.ToString(p.ParameterKey)]
		if isConfigured {
			params[aws.ToString(p.ParameterKey)] = aws.ToString(p.ParameterValue)
		}
	}
	return params
}
