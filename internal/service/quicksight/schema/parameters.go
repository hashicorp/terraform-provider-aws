// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
							names.AttrName: stringNonEmptyRequiredSchema(),
							names.AttrValues: {
								Type:     schema.TypeList,
								MinItems: 1,
								Required: true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: verify.ValidUTCTimestamp,
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
							names.AttrName: stringNonEmptyRequiredSchema(),
							names.AttrValues: {
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
							names.AttrName: stringNonEmptyRequiredSchema(),
							names.AttrValues: {
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
							names.AttrName: stringNonEmptyRequiredSchema(),
							names.AttrValues: {
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

var stringNonEmptyRequiredSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringIsNotWhiteSpace,
	}
})

func ExpandParameters(tfList []any) *awstypes.Parameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.Parameters{}

	if v, ok := tfMap["date_time_parameters"].([]any); ok && len(v) > 0 {
		apiObject.DateTimeParameters = expandDateTimeParameters(v)
	}
	if v, ok := tfMap["decimal_parameters"].([]any); ok && len(v) > 0 {
		apiObject.DecimalParameters = expandDecimalParameters(v)
	}
	if v, ok := tfMap["integer_parameters"].([]any); ok && len(v) > 0 {
		apiObject.IntegerParameters = expandIntegerParameters(v)
	}
	if v, ok := tfMap["string_parameters"].([]any); ok && len(v) > 0 {
		apiObject.StringParameters = expandStringParameters(v)
	}

	return apiObject
}

func expandDateTimeParameters(tfList []any) []awstypes.DateTimeParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DateTimeParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandDateTimeParameter(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDateTimeParameter(tfMap map[string]any) *awstypes.DateTimeParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DateTimeParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringTimeValueList(v, time.RFC3339)
	}

	return apiObject
}

func expandDecimalParameters(tfList []any) []awstypes.DecimalParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DecimalParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandDecimalParameter(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDecimalParameter(tfMap map[string]any) *awstypes.DecimalParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DecimalParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandFloat64ValueList(v)
	}

	return apiObject
}

func expandIntegerParameters(tfList []any) []awstypes.IntegerParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.IntegerParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandIntegerParameter(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandIntegerParameter(tfMap map[string]any) *awstypes.IntegerParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.IntegerParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandInt64ValueList(v)
	}

	return apiObject
}

func expandStringParameters(tfList []any) []awstypes.StringParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.StringParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandStringParameter(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandStringParameter(tfMap map[string]any) *awstypes.StringParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.StringParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringValueList(v)
	}

	return apiObject
}
