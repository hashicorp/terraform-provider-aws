// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ParametersSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Parameters.html
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_parameters": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeParameter.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 100,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": stringSchema(true, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`.*\S.*`), ""))),
							"values": {
								Type:     schema.TypeList,
								MinItems: 1,
								Required: true,
								Elem: &schema.Schema{
									Type:             schema.TypeString,
									ValidateDiagFunc: validation.ToDiagFunc(verify.ValidUTCTimestamp),
								},
							},
						},
					},
				},
				"decimal_parameters": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalParameter.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 100,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": stringSchema(true, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`.*\S.*`), ""))),
							"values": {
								Type:     schema.TypeList,
								MinItems: 1,
								Required: true,
								Elem: &schema.Schema{
									Type: schema.TypeFloat,
								},
							},
						},
					},
				},
				"integer_parameters": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_IntegerParameter.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 100,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": stringSchema(true, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`.*\S.*`), ""))),
							"values": {
								Type:     schema.TypeList,
								MinItems: 1,
								Required: true,
								Elem: &schema.Schema{
									Type: schema.TypeInt,
								},
							},
						},
					},
				},
				"string_parameters": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringParameter.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 100,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": stringSchema(true, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`.*\S.*`), ""))),
							"values": {
								Type:     schema.TypeList,
								MinItems: 1,
								Required: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
		},
	}
}

func ExpandParameters(tfList []interface{}) *types.Parameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	parameters := &types.Parameters{}

	if v, ok := tfMap["date_time_parameters"].([]interface{}); ok && len(v) > 0 {
		parameters.DateTimeParameters = expandDateTimeParameters(v)
	}
	if v, ok := tfMap["decimal_parameters"].([]interface{}); ok && len(v) > 0 {
		parameters.DecimalParameters = expandDecimalParameters(v)
	}
	if v, ok := tfMap["integer_parameters"].([]interface{}); ok && len(v) > 0 {
		parameters.IntegerParameters = expandIntegerParameters(v)
	}
	if v, ok := tfMap["string_parameters"].([]interface{}); ok && len(v) > 0 {
		parameters.StringParameters = expandStringParameters(v)
	}

	return parameters
}

func expandDateTimeParameters(tfList []interface{}) []types.DateTimeParameter {
	if len(tfList) == 0 {
		return nil
	}

	var parameters []types.DateTimeParameter
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		parameter := expandDateTimeParameter(tfMap)
		if parameter == nil {
			continue
		}

		parameters = append(parameters, *parameter)
	}

	return parameters
}

func expandDateTimeParameter(tfMap map[string]interface{}) *types.DateTimeParameter {
	if tfMap == nil {
		return nil
	}

	parameter := &types.DateTimeParameter{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		parameter.Name = aws.String(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		parameter.Values = flex.ExpandStringTimeValueList(v, time.RFC3339)
	}

	return parameter
}

func expandDecimalParameters(tfList []interface{}) []types.DecimalParameter {
	if len(tfList) == 0 {
		return nil
	}

	var parameters []types.DecimalParameter
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		parameter := expandDecimalParameter(tfMap)
		if parameter == nil {
			continue
		}

		parameters = append(parameters, *parameter)
	}

	return parameters
}

func expandDecimalParameter(tfMap map[string]interface{}) *types.DecimalParameter {
	if tfMap == nil {
		return nil
	}

	parameter := &types.DecimalParameter{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		parameter.Name = aws.String(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		parameter.Values = flex.ExpandFloat64ValueList(v)
	}

	return parameter
}

func expandIntegerParameters(tfList []interface{}) []types.IntegerParameter {
	if len(tfList) == 0 {
		return nil
	}

	var parameters []types.IntegerParameter
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		parameter := expandIntegerParameter(tfMap)
		if parameter == nil {
			continue
		}

		parameters = append(parameters, *parameter)
	}

	return parameters
}

func expandIntegerParameter(tfMap map[string]interface{}) *types.IntegerParameter {
	if tfMap == nil {
		return nil
	}

	parameter := &types.IntegerParameter{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		parameter.Name = aws.String(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		parameter.Values = flex.ExpandInt64ValueList(v)
	}

	return parameter
}

func expandStringParameters(tfList []interface{}) []types.StringParameter {
	if len(tfList) == 0 {
		return nil
	}

	var parameters []types.StringParameter
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		parameter := expandStringParameter(tfMap)
		if parameter == nil {
			continue
		}

		parameters = append(parameters, *parameter)
	}

	return parameters
}

func expandStringParameter(tfMap map[string]interface{}) *types.StringParameter {
	if tfMap == nil {
		return nil
	}

	parameter := &types.StringParameter{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		parameter.Name = aws.String(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		parameter.Values = flex.ExpandStringValueList(v)
	}

	return parameter
}
