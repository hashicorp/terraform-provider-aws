// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pipes

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func enrichmentParametersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"http_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"header_parameters": {
								Type:     schema.TypeMap,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"path_parameter_values": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"query_string_parameters": {
								Type:     schema.TypeMap,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"input_template": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 8192),
				},
			},
		},
	}
}

func expandPipeEnrichmentParameters(tfMap map[string]interface{}) *types.PipeEnrichmentParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeEnrichmentParameters{}

	if v, ok := tfMap["http_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.HttpParameters = expandPipeEnrichmentHTTPParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["input_template"].(string); ok && v != "" {
		apiObject.InputTemplate = aws.String(v)
	}

	return apiObject
}

func expandPipeEnrichmentHTTPParameters(tfMap map[string]interface{}) *types.PipeEnrichmentHttpParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeEnrichmentHttpParameters{}

	if v, ok := tfMap["header_parameters"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.HeaderParameters = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["path_parameter_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.PathParameterValues = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["query_string_parameters"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.QueryStringParameters = flex.ExpandStringValueMap(v)
	}

	return apiObject
}

func flattenPipeEnrichmentParameters(apiObject *types.PipeEnrichmentParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HttpParameters; v != nil {
		tfMap["http_parameters"] = []interface{}{flattenPipeEnrichmentHTTPParameters(v)}
	}

	if v := apiObject.InputTemplate; v != nil {
		tfMap["input_template"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPipeEnrichmentHTTPParameters(apiObject *types.PipeEnrichmentHttpParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HeaderParameters; v != nil {
		tfMap["header_parameters"] = v
	}

	if v := apiObject.PathParameterValues; v != nil {
		tfMap["path_parameter_values"] = v
	}

	if v := apiObject.QueryStringParameters; v != nil {
		tfMap["query_string_parameters"] = v
	}

	return tfMap
}
