package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func expandParameters(params map[string]interface{}) []*cloudformation.Parameter {
	var cfParams []*cloudformation.Parameter
	for k, v := range params {
		cfParams = append(cfParams, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v.(string)),
		})
	}

	return cfParams
}

func flattenAllCloudFormationParameters(cfParams []*cloudformation.Parameter) map[string]interface{} {
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		params[*p.ParameterKey] = *p.ParameterValue
	}
	return params
}

func flattenOutputs(cfOutputs []*cloudformation.Output) map[string]string {
	outputs := make(map[string]string, len(cfOutputs))
	for _, o := range cfOutputs {
		outputs[*o.OutputKey] = *o.OutputValue
	}
	return outputs
}

// flattenParameters is flattening list of
// *cloudformation.Parameters and only returning existing
// parameters to avoid clash with default values
func flattenParameters(cfParams []*cloudformation.Parameter,
	originalParams map[string]interface{}) map[string]interface{} {
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		_, isConfigured := originalParams[*p.ParameterKey]
		if isConfigured {
			params[*p.ParameterKey] = *p.ParameterValue
		}
	}
	return params
}
