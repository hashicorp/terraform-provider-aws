// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceCredentialsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"copy_source_arn": {
					Type:          schema.TypeString,
					Optional:      true,
					ValidateFunc:  verify.ValidARN,
					ConflictsWith: []string{"credentials.0.credential_pair", "credentials.0.secret_arn"},
				},
				"credential_pair": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrPassword: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.NoZeroValues,
									validation.StringLenBetween(1, 1024),
								),
								Sensitive: true,
							},
							names.AttrUsername: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.NoZeroValues,
									validation.StringLenBetween(1, 64),
								),
								Sensitive: true,
							},
						},
					},
					ConflictsWith: []string{"credentials.0.copy_source_arn", "credentials.0.secret_arn"},
				},
				"secret_arn": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"credentials.0.credential_pair", "credentials.0.copy_source_arn"},
				},
			},
		},
	}
}

func DataSourceParametersSchema() *schema.Schema {
	exactlyOneOf := []string{
		"parameters.0.amazon_elasticsearch",
		"parameters.0.athena",
		"parameters.0.aurora",
		"parameters.0.aurora_postgresql",
		"parameters.0.aws_iot_analytics",
		"parameters.0.databricks",
		"parameters.0.jira",
		"parameters.0.maria_db",
		"parameters.0.mysql",
		"parameters.0.oracle",
		"parameters.0.postgresql",
		"parameters.0.presto",
		"parameters.0.rds",
		"parameters.0.redshift",
		"parameters.0.s3",
		"parameters.0.service_now",
		"parameters.0.snowflake",
		"parameters.0.spark",
		"parameters.0.sql_server",
		"parameters.0.teradata",
		"parameters.0.twitter",
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"amazon_elasticsearch": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDomain: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"athena": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"work_group": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"aurora": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"aurora_postgresql": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"aws_iot_analytics": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_set_name": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"databricks": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
							"sql_endpoint_path": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"jira": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"site_base_url": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"maria_db": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"mysql": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"oracle": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"postgresql": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"presto": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"catalog": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"rds": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrInstanceID: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"redshift": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cluster_id": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrPort: {
								Type:     schema.TypeInt,
								Optional: true,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"s3": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"manifest_file_location": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrBucket: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrKey: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"service_now": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"site_base_url": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"snowflake": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"warehouse": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"spark": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"sql_server": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"teradata": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							"host": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
							names.AttrPort: {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
				"twitter": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_rows": {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntAtLeast(1),
							},
							"query": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
					ExactlyOneOf: exactlyOneOf,
				},
			},
		},
	}
}

func SSLPropertiesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disable_ssl": {
					Type:     schema.TypeBool,
					Required: true,
				},
			},
		},
	}
}

func VPCConnectionPropertiesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"vpc_connection_arn": arnStringSchema(attrRequired),
			},
		},
	}
}

func ExpandDataSourceCredentials(tfList []interface{}) *awstypes.DataSourceCredentials {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataSourceCredentials{}

	if v, ok := tfMap["copy_source_arn"].(string); ok && v != "" {
		apiObject.CopySourceArn = aws.String(v)
	}

	if v, ok := tfMap["credential_pair"].([]interface{}); ok && len(v) > 0 {
		apiObject.CredentialPair = expandCredentialPair(v)
	}

	if v, ok := tfMap["secret_arn"].(string); ok && v != "" {
		apiObject.SecretArn = aws.String(v)
	}

	return apiObject
}

func expandCredentialPair(tfList []interface{}) *awstypes.CredentialPair {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.CredentialPair{}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
		apiObject.Username = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPassword].(string); ok && v != "" {
		apiObject.Password = aws.String(v)
	}

	return apiObject
}

func ExpandDataSourceParameters(tfList []interface{}) awstypes.DataSourceParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var apiObject awstypes.DataSourceParameters

	if v, ok := tfMap["amazon_elasticsearch"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAmazonElasticsearchParameters{}

			if v, ok := tfMap[names.AttrDomain].(string); ok && v != "" {
				ps.Value.Domain = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["athena"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAthenaParameters{}

			if v, ok := tfMap["work_group"].(string); ok && v != "" {
				ps.Value.WorkGroup = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["aurora"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAuroraParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v, ok := tfMap["aurora_postgresql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAuroraPostgreSqlParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["aws_iot_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAwsIotAnalyticsParameters{}

			if v, ok := tfMap["data_set_name"].(string); ok && v != "" {
				ps.Value.DataSetName = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["databricks"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberDatabricksParameters{}

			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}
			if v, ok := tfMap["sql_endpoint_path"].(string); ok && v != "" {
				ps.Value.SqlEndpointPath = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["jira"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberJiraParameters{}

			if v, ok := tfMap["site_base_url"].(string); ok && v != "" {
				ps.Value.SiteBaseUrl = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["maria_db"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberMariaDbParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["mysql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberMySqlParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["oracle"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberOracleParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["postgresql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberPostgreSqlParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["presto"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberPrestoParameters{}

			if v, ok := tfMap["catalog"].(string); ok && v != "" {
				ps.Value.Catalog = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["rds"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberRdsParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap[names.AttrInstanceID].(string); ok && v != "" {
				ps.Value.InstanceId = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["redshift"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberRedshiftParameters{}

			if v, ok := tfMap["cluster_id"].(string); ok && v != "" {
				ps.Value.ClusterId = aws.String(v)
			}
			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = int32(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["s3"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberS3Parameters{}

			if v, ok := tfMap["manifest_file_location"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				if tfMap, ok := v[0].(map[string]interface{}); ok {
					apiObject := &awstypes.ManifestFileLocation{}

					if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
						apiObject.Bucket = aws.String(v)
					}
					if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
						apiObject.Key = aws.String(v)
					}

					ps.Value.ManifestFileLocation = apiObject
				}
			}

			apiObject = ps
		}
	}

	if v := tfMap["service_now"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberServiceNowParameters{}

			if v, ok := tfMap["site_base_url"].(string); ok && v != "" {
				ps.Value.SiteBaseUrl = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["snowflake"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberSnowflakeParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap["warehouse"].(string); ok && v != "" {
				ps.Value.Warehouse = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["spark"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberSparkParameters{}

			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["sql_server"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberSqlServerParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["teradata"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberTeradataParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["twitter"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberTwitterParameters{}

			if v, ok := tfMap["max_rows"].(int); ok {
				ps.Value.MaxRows = aws.Int32(int32(v))
			}
			if v, ok := tfMap["query"].(string); ok && v != "" {
				ps.Value.Query = aws.String(v)
			}

			apiObject = ps
		}
	}

	return apiObject
}

func FlattenDataSourceParameters(apiObject awstypes.DataSourceParameters) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	switch v := apiObject.(type) {
	case *awstypes.DataSourceParametersMemberAmazonElasticsearchParameters:
		tfMap["amazon_elasticsearch"] = []interface{}{
			map[string]interface{}{
				names.AttrDomain: aws.ToString(v.Value.Domain),
			},
		}
	case *awstypes.DataSourceParametersMemberAthenaParameters:
		tfMap["athena"] = []interface{}{
			map[string]interface{}{
				"work_group": aws.ToString(v.Value.WorkGroup),
			},
		}
	case *awstypes.DataSourceParametersMemberAuroraParameters:
		tfMap["aurora"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberAuroraPostgreSqlParameters:
		tfMap["aurora_postgresql"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberAwsIotAnalyticsParameters:
		tfMap["aws_iot_analytics"] = []interface{}{
			map[string]interface{}{
				"data_set_name": aws.ToString(v.Value.DataSetName),
			},
		}
	case *awstypes.DataSourceParametersMemberDatabricksParameters:
		tfMap["databricks"] = []interface{}{
			map[string]interface{}{
				"host":              aws.ToString(v.Value.Host),
				names.AttrPort:      aws.ToInt32(v.Value.Port),
				"sql_endpoint_path": aws.ToString(v.Value.SqlEndpointPath),
			},
		}
	case *awstypes.DataSourceParametersMemberJiraParameters:
		tfMap["jira"] = []interface{}{
			map[string]interface{}{
				"site_base_url": aws.ToString(v.Value.SiteBaseUrl),
			},
		}
	case *awstypes.DataSourceParametersMemberMariaDbParameters:
		tfMap["maria_db"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberMySqlParameters:
		tfMap["mysql"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberOracleParameters:
		tfMap["oracle"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberPostgreSqlParameters:
		tfMap["postgresql"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberPrestoParameters:
		tfMap["postgresql"] = []interface{}{
			map[string]interface{}{
				"catalog":      aws.ToString(v.Value.Catalog),
				"host":         aws.ToString(v.Value.Host),
				names.AttrPort: aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberRdsParameters:
		tfMap["rds"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase:   aws.ToString(v.Value.Database),
				names.AttrInstanceID: aws.ToString(v.Value.InstanceId),
			},
		}
	case *awstypes.DataSourceParametersMemberRedshiftParameters:
		tfMap["redshift"] = []interface{}{
			map[string]interface{}{
				"cluster_id":       aws.ToString(v.Value.ClusterId),
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     v.Value.Port,
			},
		}
	case *awstypes.DataSourceParametersMemberS3Parameters:
		tfMap["s3"] = []interface{}{
			map[string]interface{}{
				"manifest_file_location": []interface{}{
					map[string]interface{}{
						names.AttrBucket: aws.ToString(v.Value.ManifestFileLocation.Bucket),
						names.AttrKey:    aws.ToString(v.Value.ManifestFileLocation.Key),
					},
				},
			},
		}
	case *awstypes.DataSourceParametersMemberServiceNowParameters:
		tfMap["service_now"] = []interface{}{
			map[string]interface{}{
				"site_base_url": aws.ToString(v.Value.SiteBaseUrl),
			},
		}
	case *awstypes.DataSourceParametersMemberSnowflakeParameters:
		tfMap["snowflake"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				"warehouse":        aws.ToString(v.Value.Warehouse),
			},
		}
	case *awstypes.DataSourceParametersMemberSparkParameters:
		tfMap["snowflake"] = []interface{}{
			map[string]interface{}{
				"host":         aws.ToString(v.Value.Host),
				names.AttrPort: aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberSqlServerParameters:
		tfMap["sql_server"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     v.Value.Port,
			},
		}
	case *awstypes.DataSourceParametersMemberTeradataParameters:
		tfMap["teradata"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     v.Value.Port,
			},
		}
	case *awstypes.DataSourceParametersMemberTwitterParameters:
		tfMap["teradata"] = []interface{}{
			map[string]interface{}{
				"max_rows": aws.ToInt32(v.Value.MaxRows),
				"query":    aws.ToString(v.Value.Query),
			},
		}
	default:
		return nil
	}

	return []interface{}{tfMap}
}

func ExpandSSLProperties(tfList []interface{}) *awstypes.SslProperties {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SslProperties{}

	if v, ok := tfMap["disable_ssl"].(bool); ok {
		apiObject.DisableSsl = v
	}

	return apiObject
}

func FlattenSSLProperties(apiObject *awstypes.SslProperties) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"disable_ssl": apiObject.DisableSsl,
	}

	return []interface{}{tfMap}
}

func ExpandVPCConnectionProperties(tfList []interface{}) *awstypes.VpcConnectionProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.VpcConnectionProperties{}

	if v, ok := tfMap["vpc_connection_arn"].(string); ok && v != "" {
		apiObject.VpcConnectionArn = aws.String(v)
	}

	return apiObject
}

func FlattenVPCConnectionProperties(apiObject *awstypes.VpcConnectionProperties) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if apiObject.VpcConnectionArn != nil {
		tfMap["vpc_connection_arn"] = aws.ToString(apiObject.VpcConnectionArn)
	}

	return []interface{}{tfMap}
}
