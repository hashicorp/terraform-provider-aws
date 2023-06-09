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
							"header": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"key": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(0, 512),
										},
										"value": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(0, 512),
										},
									},
								},
							},
							"path_parameters": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"query_string": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"key": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(0, 512),
										},
										"value": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(0, 512),
										},
									},
								},
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

func expandEnrichmentParameters(config []interface{}) *types.PipeEnrichmentParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeEnrichmentParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["input_template"].(string); ok && val != "" {
			parameters.InputTemplate = aws.String(val)
		}
		if val, ok := param["http_parameters"]; ok {
			parameters.HttpParameters = expandEnrichmentHTTPParameters(val.([]interface{}))
		}
	}
	return &parameters
}

func expandEnrichmentHTTPParameters(config []interface{}) *types.PipeEnrichmentHttpParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeEnrichmentHttpParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["path_parameters"]; ok {
			parameters.PathParameterValues = flex.ExpandStringValueList(val.([]interface{}))
		}

		if val, ok := param["header"]; ok {
			headers := map[string]string{}
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					if key, ok := valueParam["key"].(string); ok && key != "" {
						if value, ok := valueParam["value"].(string); ok && value != "" {
							headers[key] = value
						}
					}
				}
			}
			if len(headers) > 0 {
				parameters.HeaderParameters = headers
			}
		}

		if val, ok := param["query_string"]; ok {
			queryStrings := map[string]string{}
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					if key, ok := valueParam["key"].(string); ok && key != "" {
						if value, ok := valueParam["value"].(string); ok && value != "" {
							queryStrings[key] = value
						}
					}
				}
			}
			if len(queryStrings) > 0 {
				parameters.QueryStringParameters = queryStrings
			}
		}
	}
	return &parameters
}

func flattenEnrichmentParameters(enrichmentParameters *types.PipeEnrichmentParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if enrichmentParameters.InputTemplate != nil {
		config["input_template"] = *enrichmentParameters.InputTemplate
	}

	if enrichmentParameters.HttpParameters != nil {
		httpParameters := make(map[string]interface{})

		var headerParameters []map[string]interface{}
		for key, value := range enrichmentParameters.HttpParameters.HeaderParameters {
			header := make(map[string]interface{})
			header["key"] = key
			header["value"] = value
			headerParameters = append(headerParameters, header)
		}
		httpParameters["header"] = headerParameters

		var queryStringParameters []map[string]interface{}
		for key, value := range enrichmentParameters.HttpParameters.QueryStringParameters {
			queryString := make(map[string]interface{})
			queryString["key"] = key
			queryString["value"] = value
			queryStringParameters = append(queryStringParameters, queryString)
		}
		httpParameters["query_string"] = queryStringParameters
		httpParameters["path_parameters"] = flex.FlattenStringValueList(enrichmentParameters.HttpParameters.PathParameterValues)

		config["http_parameters"] = []map[string]interface{}{httpParameters}
	}

	if len(config) == 0 {
		return nil
	}

	result := []map[string]interface{}{config}
	return result
}
