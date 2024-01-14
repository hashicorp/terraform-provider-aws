// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qbusiness"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func keySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 200),
			validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`),
				"must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters")),
	}
}

func valueSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_value": {
					Type:          schema.TypeString,
					Optional:      true,
					ValidateFunc:  validation.IsRFC3339Time,
					ConflictsWith: []string{"long_value", "string_list_value", "strings_value"},
					AtLeastOneOf:  []string{"date_value", "long_value", "string_list_value", "strings_value"},
				},
				"long_value": {
					Type:          schema.TypeInt,
					Optional:      true,
					ConflictsWith: []string{"date_value", "string_list_value", "strings_value"},
					AtLeastOneOf:  []string{"date_value", "long_value", "string_list_value", "strings_value"},
				},
				"string_list_value": {
					Type:          schema.TypeList,
					Optional:      true,
					MinItems:      1,
					MaxItems:      2048,
					Elem:          &schema.Schema{Type: schema.TypeString},
					ConflictsWith: []string{"date_value", "long_value", "strings_value"},
					AtLeastOneOf:  []string{"date_value", "long_value", "string_list_value", "strings_value"},
				},
				"strings_value": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"date_value", "long_value", "string_list_value"},
					AtLeastOneOf:  []string{"date_value", "long_value", "string_list_value", "strings_value"},
				},
			},
		},
	}
}

func documentAttributeConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": keySchema(),
				"operator": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(qbusiness.DocumentEnrichmentConditionOperator_Values(), false),
				},
				"value": valueSchema(),
			},
		},
	}
}

func hookConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"invocation_condition": documentAttributeConditionSchema(),
			"lambda_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"s3_bucket_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9]$`), "must be a valid bucket name")),
			},
		}},
	}
}

// @SDKResource("aws_qbusiness_datasource", name="Datasource")
// @Tags(identifierAttribute="arn")
func ResourceDatasource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the Amazon Q application the data source will be attached to.",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				),
			},
			"configration": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Configuration information (JSON) to connect to your data source repository.",
				ValidateFunc: validation.StringIsJSON,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description for the data source connector.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1000),
					validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				),
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A name for the data source connector.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
				),
			},
			"document_enrichment_configuration": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration information for altering document metadata and content during the document ingestion process.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inline_configurations": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Information to alter document attributes or metadata fields and content when ingesting documents into Amazon Q",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"configuration": {
										Type:     schema.TypeList,
										MinItems: 1,
										MaxItems: 100,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"condition": documentAttributeConditionSchema(),
												"document_content_operator": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(qbusiness.DocumentContentOperator_Values(), false),
												},
												"target": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"key": keySchema(),
															"attribute_value_operator": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(qbusiness.AttributeValueOperator_Values(), false),
															},
															"value": valueSchema(),
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"post_extraction_hook_configuration": hookConfigurationSchema(),
						"pre_extraction_hook_configuration":  hookConfigurationSchema(),
					},
				},
			},
			"iam_service_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The Amazon Resource Name (ARN) of an IAM role with permission to access the data source and required resources.",
				ValidateFunc: verify.ValidARN,
			},
			"index_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the index that you want to use with the data source connector.",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				),
			},
			"sync_schedule": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Frequency for Amazon Q to check the documents in your data source repository and update your index.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 998),
					validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				),
			},
			"vpc_configuration": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Information for an VPC to connect to your data source.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 200,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 200,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func expandHookConfiguration(v []interface{}) *qbusiness.HookConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &qbusiness.HookConfiguration{
		InvocationCondition: expandDocumentAttributeCondition(m["invocation_condition"].([]interface{})),
		RoleArn:             aws.String(m["role_arn"].(string)),
		LambdaArn:           aws.String(m["lambda_arn"].(string)),
		S3BucketName:        aws.String(m["s3_bucket_name"].(string)),
	}
}

func expandDocumentAttributeCondition(v []interface{}) *qbusiness.DocumentAttributeCondition {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &qbusiness.DocumentAttributeCondition{
		Key:      aws.String(m["key"].(string)),
		Operator: aws.String(m["operator"].(string)),
		Value:    expandValueSchema(m["value"].([]interface{})),
	}
}

func expandValueSchema(v []interface{}) *qbusiness.DocumentAttributeValue {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	t, _ := time.Parse(time.RFC3339, m["date_value"].(string))

	return &qbusiness.DocumentAttributeValue{
		DateValue:       aws.Time(t),
		LongValue:       aws.Int64(int64(m["long_value"].(int))),
		StringListValue: flex.ExpandStringList(m["string_list_value"].([]interface{})),
		StringValue:     aws.String(m["strings_value"].(string)),
	}
}

func expandVPCConfiguration(cfg []interface{}) *qbusiness.DataSourceVpcConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := qbusiness.DataSourceVpcConfiguration{}

	if v, ok := conf["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		out.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := conf["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		out.SubnetIds = flex.ExpandStringSet(v)
	}

	return &out
}

func flattenVPCConfiguration(rs *qbusiness.DataSourceVpcConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.SecurityGroupIds != nil {
		m["security_group_ids"] = flex.FlattenStringSet(rs.SecurityGroupIds)
	}
	if rs.SubnetIds != nil {
		m["subnet_ids"] = flex.FlattenStringSet(rs.SubnetIds)
	}

	return []interface{}{m}
}
