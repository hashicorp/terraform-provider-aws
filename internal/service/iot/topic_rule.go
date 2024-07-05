// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_topic_rule", name="Topic Rule")
// @Tags(identifierAttribute="arn")
func resourceTopicRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicRuleCreate,
		ReadWithoutTimeout:   resourceTopicRuleRead,
		UpdateWithoutTimeout: resourceTopicRuleUpdate,
		DeleteWithoutTimeout: resourceTopicRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			topicRuleErrorActionExactlyOneOf := []string{
				"error_action.0.cloudwatch_alarm",
				"error_action.0.cloudwatch_logs",
				"error_action.0.cloudwatch_metric",
				"error_action.0.dynamodb",
				"error_action.0.dynamodbv2",
				"error_action.0.elasticsearch",
				"error_action.0.firehose",
				"error_action.0.http",
				"error_action.0.iot_analytics",
				"error_action.0.iot_events",
				"error_action.0.kafka",
				"error_action.0.kinesis",
				"error_action.0.lambda",
				"error_action.0.republish",
				"error_action.0.s3",
				"error_action.0.sns",
				"error_action.0.sqs",
				"error_action.0.step_functions",
				"error_action.0.timestream",
			}

			timestreamDimensionResource := func() *schema.Resource {
				return &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				}
			}

			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"cloudwatch_alarm": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"alarm_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"state_reason": {
								Type:     schema.TypeString,
								Required: true,
							},
							"state_value": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validTopicRuleCloudWatchAlarmStateValue,
							},
						},
					},
				},
				names.AttrCloudWatchLogs: {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_mode": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							names.AttrLogGroupName: {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"cloudwatch_metric": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrMetricName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"metric_namespace": {
								Type:     schema.TypeString,
								Required: true,
							},
							"metric_timestamp": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidUTCTimestamp,
							},
							"metric_unit": {
								Type:     schema.TypeString,
								Required: true,
							},
							"metric_value": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				"dynamodb": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"hash_key_field": {
								Type:     schema.TypeString,
								Required: true,
							},
							"hash_key_value": {
								Type:     schema.TypeString,
								Required: true,
							},
							"hash_key_type": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"operation": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									"DELETE",
									"INSERT",
									"UPDATE",
								}, false),
							},
							"payload_field": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"range_key_field": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"range_key_value": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"range_key_type": {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							names.AttrTableName: {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"dynamodbv2": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"put_item": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrTableName: {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"elasticsearch": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrEndpoint: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validTopicRuleElasticsearchEndpoint,
							},
							names.AttrID: {
								Type:     schema.TypeString,
								Required: true,
							},
							"index": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							names.AttrType: {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				names.AttrEnabled: {
					Type:     schema.TypeBool,
					Required: true,
				},
				"error_action": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cloudwatch_alarm": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"alarm_name": {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										"state_reason": {
											Type:     schema.TypeString,
											Required: true,
										},
										"state_value": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validTopicRuleCloudWatchAlarmStateValue,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							names.AttrCloudWatchLogs: {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"batch_mode": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  false,
										},
										names.AttrLogGroupName: {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"cloudwatch_metric": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrMetricName: {
											Type:     schema.TypeString,
											Required: true,
										},
										"metric_namespace": {
											Type:     schema.TypeString,
											Required: true,
										},
										"metric_timestamp": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verify.ValidUTCTimestamp,
										},
										"metric_unit": {
											Type:     schema.TypeString,
											Required: true,
										},
										"metric_value": {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"dynamodb": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"hash_key_field": {
											Type:     schema.TypeString,
											Required: true,
										},
										"hash_key_value": {
											Type:     schema.TypeString,
											Required: true,
										},
										"hash_key_type": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"operation": {
											Type:     schema.TypeString,
											Optional: true,
											ValidateFunc: validation.StringInSlice([]string{
												"DELETE",
												"INSERT",
												"UPDATE",
											}, false),
										},
										"payload_field": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"range_key_field": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"range_key_value": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"range_key_type": {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										names.AttrTableName: {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"dynamodbv2": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"put_item": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrTableName: {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"elasticsearch": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrEndpoint: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validTopicRuleElasticsearchEndpoint,
										},
										names.AttrID: {
											Type:     schema.TypeString,
											Required: true,
										},
										"index": {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										names.AttrType: {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"firehose": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"batch_mode": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  false,
										},
										"delivery_stream_name": {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										"separator": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validTopicRuleFirehoseSeparator,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"http": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"confirmation_url": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.IsURLWithHTTPS,
										},
										"http_header": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrKey: {
														Type:     schema.TypeString,
														Required: true,
													},
													names.AttrValue: {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
										names.AttrURL: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.IsURLWithHTTPS,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"iot_analytics": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"batch_mode": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  false,
										},
										"channel_name": {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"iot_events": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"batch_mode": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  false,
										},
										"input_name": {
											Type:     schema.TypeString,
											Required: true,
										},
										"message_id": {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"kafka": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"client_properties": {
											Type:     schema.TypeMap,
											Required: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrDestinationARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										names.AttrHeader: {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrKey: {
														Type:     schema.TypeString,
														Required: true,
													},
													names.AttrValue: {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
										names.AttrKey: {
											Type:     schema.TypeString,
											Optional: true,
										},
										"partition": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"topic": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"kinesis": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"partition_key": {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										"stream_name": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"lambda": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrFunctionARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"republish": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"qos": {
											Type:         schema.TypeInt,
											Optional:     true,
											Default:      0,
											ValidateFunc: validation.IntBetween(0, 1),
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										"topic": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"s3": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrBucketName: {
											Type:     schema.TypeString,
											Required: true,
										},
										"canned_acl": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[awstypes.CannedAccessControlList](),
										},
										names.AttrKey: {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"sns": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"message_format": {
											Type:     schema.TypeString,
											Default:  awstypes.MessageFormatRaw,
											Optional: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										names.AttrTargetARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"sqs": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"queue_url": {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										"use_base64": {
											Type:     schema.TypeBool,
											Required: true,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"step_functions": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"execution_name_prefix": {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										"state_machine_name": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
							"timestream": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabaseName: {
											Type:     schema.TypeString,
											Required: true,
										},
										"dimension": {
											Type:     schema.TypeSet,
											Required: true,
											Elem:     timestreamDimensionResource(),
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
										names.AttrTableName: {
											Type:     schema.TypeString,
											Required: true,
										},
										"timestamp": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrUnit: {
														Type:     schema.TypeString,
														Required: true,
														ValidateFunc: validation.StringInSlice([]string{
															"SECONDS",
															"MILLISECONDS",
															"MICROSECONDS",
															"NANOSECONDS",
														}, false),
													},
													names.AttrValue: {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
									},
								},
								ExactlyOneOf: topicRuleErrorActionExactlyOneOf,
							},
						},
					},
				},
				"firehose": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_mode": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"delivery_stream_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"separator": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validTopicRuleFirehoseSeparator,
							},
						},
					},
				},
				"http": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"confirmation_url": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.IsURLWithHTTPS,
							},
							"http_header": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrKey: {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrValue: {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							names.AttrURL: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.IsURLWithHTTPS,
							},
						},
					},
				},
				"iot_analytics": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_mode": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"channel_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"iot_events": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_mode": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"input_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"message_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"kafka": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_properties": {
								Type:     schema.TypeMap,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							names.AttrDestinationARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							names.AttrHeader: {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrKey: {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrValue: {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							names.AttrKey: {
								Type:     schema.TypeString,
								Optional: true,
							},
							"partition": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"topic": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"kinesis": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"partition_key": {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"stream_name": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"lambda": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrFunctionARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validTopicRuleName,
				},
				"republish": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"qos": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      0,
								ValidateFunc: validation.IntBetween(0, 1),
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"topic": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"s3": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrBucketName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"canned_acl": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.CannedAccessControlList](),
							},
							names.AttrKey: {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"sns": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"message_format": {
								Type:     schema.TypeString,
								Optional: true,
								Default:  awstypes.MessageFormatRaw,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							names.AttrTargetARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"sql": {
					Type:     schema.TypeString,
					Required: true,
				},
				"sql_version": {
					Type:     schema.TypeString,
					Required: true,
				},
				"sqs": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"queue_url": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"use_base64": {
								Type:     schema.TypeBool,
								Required: true,
							},
						},
					},
				},
				"step_functions": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"execution_name_prefix": {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"state_machine_name": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"timestream": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabaseName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"dimension": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     timestreamDimensionResource(),
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							names.AttrTableName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"timestamp": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrUnit: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.StringInSlice([]string{
												"SECONDS",
												"MILLISECONDS",
												"MICROSECONDS",
												"NANOSECONDS",
											}, false),
										},
										names.AttrValue: {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
						},
					},
				},
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTopicRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	ruleName := d.Get(names.AttrName).(string)
	input := &iot.CreateTopicRuleInput{
		RuleName:         aws.String(ruleName),
		Tags:             aws.String(KeyValueTags(ctx, getTagsIn(ctx)).URLQueryString()),
		TopicRulePayload: expandTopicRulePayload(d),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateTopicRule(ctx, input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Topic Rule (%s): %s", ruleName, err)
	}

	d.SetId(ruleName)

	return append(diags, resourceTopicRuleRead(ctx, d, meta)...)
}

func resourceTopicRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findTopicRuleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Topic Rule %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Topic Rule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.RuleArn)
	d.Set(names.AttrName, output.Rule.RuleName)
	d.Set(names.AttrDescription, output.Rule.Description)
	d.Set(names.AttrEnabled, !aws.ToBool(output.Rule.RuleDisabled))
	d.Set("sql", output.Rule.Sql)
	d.Set("sql_version", output.Rule.AwsIotSqlVersion)

	if err := d.Set("cloudwatch_alarm", flattenCloudWatchAlarmActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cloudwatch_alarm: %s", err)
	}

	if err := d.Set(names.AttrCloudWatchLogs, flattenCloudWatchLogsActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cloudwatch_logs: %s", err)
	}

	if err := d.Set("cloudwatch_metric", flattenCloudWatchMetricActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cloudwatch_metric: %s", err)
	}

	if err := d.Set("dynamodb", flattenDynamoDBActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dynamodb: %s", err)
	}

	if err := d.Set("dynamodbv2", flattenDynamoDBv2Actions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dynamodbv2: %s", err)
	}

	if err := d.Set("elasticsearch", flattenElasticsearchActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting elasticsearch: %s", err)
	}

	if err := d.Set("firehose", flattenFirehoseActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting firehose: %s", err)
	}

	if err := d.Set("http", flattenHTTPActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting http: %s", err)
	}

	if err := d.Set("iot_analytics", flattenAnalyticsActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting iot_analytics: %s", err)
	}

	if err := d.Set("iot_events", flattenEventsActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting iot_events: %s", err)
	}

	if err := d.Set("kafka", flattenKafkaActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kafka: %s", err)
	}

	if err := d.Set("kinesis", flattenKinesisActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kinesis: %s", err)
	}

	if err := d.Set("lambda", flattenLambdaActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda: %s", err)
	}

	if err := d.Set("republish", flattenRepublishActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting republish: %s", err)
	}

	if err := d.Set("s3", flattenS3Actions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3: %s", err)
	}

	if err := d.Set("sns", flattenSNSActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sns: %s", err)
	}

	if err := d.Set("sqs", flattenSQSActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sqs: %s", err)
	}

	if err := d.Set("step_functions", flattenStepFunctionsActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting step_functions: %s", err)
	}

	if err := d.Set("timestream", flattenTimestreamActions(output.Rule.Actions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting timestream: %s", err)
	}

	if err := d.Set("error_action", flattenErrorAction(output.Rule.ErrorAction)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting error_action: %s", err)
	}

	return diags
}

func resourceTopicRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &iot.ReplaceTopicRuleInput{
			RuleName:         aws.String(d.Id()),
			TopicRulePayload: expandTopicRulePayload(d),
		}

		_, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.ReplaceTopicRule(ctx, input)
			})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "replacing IoT Topic Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTopicRuleRead(ctx, d, meta)...)
}

func resourceTopicRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	log.Printf("[INFO] Deleting IoT Topic Rule: %s", d.Id())
	_, err := conn.DeleteTopicRule(ctx, &iot.DeleteTopicRuleInput{
		RuleName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Topic Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func findTopicRuleByName(ctx context.Context, conn *iot.Client, name string) (*iot.GetTopicRuleOutput, error) {
	// GetTopicRule returns unhelpful errors such as
	//	"An error occurred (UnauthorizedException) when calling the GetTopicRule operation: Access to topic rule 'xxxxxxxx' was denied"
	// when querying for a rule that doesn't exist.
	inputL := &iot.ListTopicRulesInput{}
	var rule *awstypes.TopicRuleListItem

	pages := iot.NewListTopicRulesPaginator(conn, inputL)
pageLoop:
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Rules {
			v := v
			if aws.ToString(v.RuleName) == name {
				rule = &v
				break pageLoop
			}
		}
	}

	if rule == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	inputG := &iot.GetTopicRuleInput{
		RuleName: aws.String(name),
	}

	output, err := conn.GetTopicRule(ctx, inputG)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: inputG,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(inputG)
	}

	return output, nil
}

func expandPutItemInput(tfList []interface{}) *awstypes.PutItemInput {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.PutItemInput{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrTableName].(string); ok && v != "" {
		apiObject.TableName = aws.String(v)
	}

	return apiObject
}

func expandCloudWatchAlarmAction(tfList []interface{}) *awstypes.CloudwatchAlarmAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.CloudwatchAlarmAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["alarm_name"].(string); ok && v != "" {
		apiObject.AlarmName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["state_reason"].(string); ok && v != "" {
		apiObject.StateReason = aws.String(v)
	}

	if v, ok := tfMap["state_value"].(string); ok && v != "" {
		apiObject.StateValue = aws.String(v)
	}

	return apiObject
}

func expandCloudWatchLogsAction(tfList []interface{}) *awstypes.CloudwatchLogsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.CloudwatchLogsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["batch_mode"].(bool); ok {
		apiObject.BatchMode = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrLogGroupName].(string); ok && v != "" {
		apiObject.LogGroupName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandCloudWatchMetricAction(tfList []interface{}) *awstypes.CloudwatchMetricAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.CloudwatchMetricAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrMetricName].(string); ok && v != "" {
		apiObject.MetricName = aws.String(v)
	}

	if v, ok := tfMap["metric_namespace"].(string); ok && v != "" {
		apiObject.MetricNamespace = aws.String(v)
	}

	if v, ok := tfMap["metric_timestamp"].(string); ok && v != "" {
		apiObject.MetricTimestamp = aws.String(v)
	}

	if v, ok := tfMap["metric_unit"].(string); ok && v != "" {
		apiObject.MetricUnit = aws.String(v)
	}

	if v, ok := tfMap["metric_value"].(string); ok && v != "" {
		apiObject.MetricValue = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandDynamoDBAction(tfList []interface{}) *awstypes.DynamoDBAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.DynamoDBAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["hash_key_field"].(string); ok && v != "" {
		apiObject.HashKeyField = aws.String(v)
	}

	if v, ok := tfMap["hash_key_type"].(string); ok && v != "" {
		apiObject.HashKeyType = awstypes.DynamoKeyType(v)
	}

	if v, ok := tfMap["hash_key_value"].(string); ok && v != "" {
		apiObject.HashKeyValue = aws.String(v)
	}

	if v, ok := tfMap["operation"].(string); ok && v != "" {
		apiObject.Operation = aws.String(v)
	}

	if v, ok := tfMap["payload_field"].(string); ok && v != "" {
		apiObject.PayloadField = aws.String(v)
	}

	if v, ok := tfMap["range_key_field"].(string); ok && v != "" {
		apiObject.RangeKeyField = aws.String(v)
	}

	if v, ok := tfMap["range_key_type"].(string); ok && v != "" {
		apiObject.RangeKeyType = awstypes.DynamoKeyType(v)
	}

	if v, ok := tfMap["range_key_value"].(string); ok && v != "" {
		apiObject.RangeKeyValue = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTableName].(string); ok && v != "" {
		apiObject.TableName = aws.String(v)
	}

	return apiObject
}

func expandDynamoDBv2Action(tfList []interface{}) *awstypes.DynamoDBv2Action {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.DynamoDBv2Action{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["put_item"].([]interface{}); ok {
		apiObject.PutItem = expandPutItemInput(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandElasticsearchAction(tfList []interface{}) *awstypes.ElasticsearchAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.ElasticsearchAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrEndpoint].(string); ok && v != "" {
		apiObject.Endpoint = aws.String(v)
	}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap["index"].(string); ok && v != "" {
		apiObject.Index = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandFirehoseAction(tfList []interface{}) *awstypes.FirehoseAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.FirehoseAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["batch_mode"].(bool); ok {
		apiObject.BatchMode = aws.Bool(v)
	}

	if v, ok := tfMap["delivery_stream_name"].(string); ok && v != "" {
		apiObject.DeliveryStreamName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["separator"].(string); ok && v != "" {
		apiObject.Separator = aws.String(v)
	}

	return apiObject
}

func expandHTTPAction(tfList []interface{}) *awstypes.HttpAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.HttpAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrURL].(string); ok && v != "" {
		apiObject.Url = aws.String(v)
	}

	if v, ok := tfMap["confirmation_url"].(string); ok && v != "" {
		apiObject.ConfirmationUrl = aws.String(v)
	}

	if v, ok := tfMap["http_header"].([]interface{}); ok {
		headerObjs := []awstypes.HttpActionHeader{}
		for _, val := range v {
			if m, ok := val.(map[string]interface{}); ok {
				headerObj := awstypes.HttpActionHeader{}
				if v, ok := m[names.AttrKey].(string); ok && v != "" {
					headerObj.Key = aws.String(v)
				}
				if v, ok := m[names.AttrValue].(string); ok && v != "" {
					headerObj.Value = aws.String(v)
				}
				headerObjs = append(headerObjs, headerObj)
			}
		}
		apiObject.Headers = headerObjs
	}

	return apiObject
}

func expandAnalyticsAction(tfList []interface{}) *awstypes.IotAnalyticsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.IotAnalyticsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["batch_mode"].(bool); ok {
		apiObject.BatchMode = aws.Bool(v)
	}

	if v, ok := tfMap["channel_name"].(string); ok && v != "" {
		apiObject.ChannelName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandEventsAction(tfList []interface{}) *awstypes.IotEventsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.IotEventsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["batch_mode"].(bool); ok {
		apiObject.BatchMode = aws.Bool(v)
	}

	if v, ok := tfMap["input_name"].(string); ok && v != "" {
		apiObject.InputName = aws.String(v)
	}

	if v, ok := tfMap["message_id"].(string); ok && v != "" {
		apiObject.MessageId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandKafkaAction(tfList []interface{}) *awstypes.KafkaAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.KafkaAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["client_properties"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.ClientProperties = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap[names.AttrDestinationARN].(string); ok && v != "" {
		apiObject.DestinationArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrHeader].([]interface{}); ok && len(v) > 0 {
		apiObject.Headers = expandKafkaHeader(v)
	}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["partition"].(string); ok && v != "" {
		apiObject.Partition = aws.String(v)
	}

	if v, ok := tfMap["topic"].(string); ok && v != "" {
		apiObject.Topic = aws.String(v)
	}

	if reflect.DeepEqual(&awstypes.KafkaAction{}, apiObject) {
		return nil
	}

	return apiObject
}

func expandKafkaHeader(tfList []interface{}) []awstypes.KafkaActionHeader {
	var apiObjects []awstypes.KafkaActionHeader
	for _, elem := range tfList {
		tfMap := elem.(map[string]interface{})

		apiObject := awstypes.KafkaActionHeader{}
		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.Key = aws.String(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandKinesisAction(tfList []interface{}) *awstypes.KinesisAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.KinesisAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["partition_key"].(string); ok && v != "" {
		apiObject.PartitionKey = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["stream_name"].(string); ok && v != "" {
		apiObject.StreamName = aws.String(v)
	}

	return apiObject
}

func expandLambdaAction(tfList []interface{}) *awstypes.LambdaAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.LambdaAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrFunctionARN].(string); ok && v != "" {
		apiObject.FunctionArn = aws.String(v)
	}

	return apiObject
}

func expandRepublishAction(tfList []interface{}) *awstypes.RepublishAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.RepublishAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["qos"].(int); ok {
		apiObject.Qos = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["topic"].(string); ok && v != "" {
		apiObject.Topic = aws.String(v)
	}

	return apiObject
}

func expandS3Action(tfList []interface{}) *awstypes.S3Action {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.S3Action{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		apiObject.BucketName = aws.String(v)
	}

	if v, ok := tfMap["canned_acl"].(string); ok && v != "" {
		apiObject.CannedAcl = awstypes.CannedAccessControlList(v)
	}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandSNSAction(tfList []interface{}) *awstypes.SnsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.SnsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["message_format"].(string); ok && v != "" {
		apiObject.MessageFormat = awstypes.MessageFormat(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTargetARN].(string); ok && v != "" {
		apiObject.TargetArn = aws.String(v)
	}

	return apiObject
}

func expandSQSAction(tfList []interface{}) *awstypes.SqsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.SqsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["queue_url"].(string); ok && v != "" {
		apiObject.QueueUrl = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["use_base64"].(bool); ok {
		apiObject.UseBase64 = aws.Bool(v)
	}

	return apiObject
}

func expandStepFunctionsAction(tfList []interface{}) *awstypes.StepFunctionsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.StepFunctionsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["execution_name_prefix"].(string); ok && v != "" {
		apiObject.ExecutionNamePrefix = aws.String(v)
	}

	if v, ok := tfMap["state_machine_name"].(string); ok && v != "" {
		apiObject.StateMachineName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandTimestreamAction(tfList []interface{}) *awstypes.TimestreamAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.TimestreamAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrDatabaseName].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["dimension"].(*schema.Set); ok {
		apiObject.Dimensions = expandTimestreamDimensions(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTableName].(string); ok && v != "" {
		apiObject.TableName = aws.String(v)
	}

	if v, ok := tfMap["timestamp"].([]interface{}); ok {
		apiObject.Timestamp = expandTimestreamTimestamp(v)
	}

	return apiObject
}

func expandTimestreamDimensions(tfSet *schema.Set) []awstypes.TimestreamDimension {
	if tfSet == nil || tfSet.Len() == 0 {
		return nil
	}

	apiObjects := make([]awstypes.TimestreamDimension, tfSet.Len())
	for i, elem := range tfSet.List() {
		if tfMap, ok := elem.(map[string]interface{}); ok {
			apiObject := awstypes.TimestreamDimension{}

			if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
				apiObject.Name = aws.String(v)
			}

			if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
				apiObject.Value = aws.String(v)
			}

			apiObjects[i] = apiObject
		}
	}

	return apiObjects
}

func expandTimestreamTimestamp(tfList []interface{}) *awstypes.TimestreamTimestamp {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.TimestreamTimestamp{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrUnit].(string); ok && v != "" {
		apiObject.Unit = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandTopicRulePayload(d *schema.ResourceData) *awstypes.TopicRulePayload {
	var actions []awstypes.Action

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("cloudwatch_alarm").(*schema.Set).List() {
		action := expandCloudWatchAlarmAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{CloudwatchAlarm: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get(names.AttrCloudWatchLogs).(*schema.Set).List() {
		action := expandCloudWatchLogsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{CloudwatchLogs: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("cloudwatch_metric").(*schema.Set).List() {
		action := expandCloudWatchMetricAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{CloudwatchMetric: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("dynamodb").(*schema.Set).List() {
		action := expandDynamoDBAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{DynamoDB: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("dynamodbv2").(*schema.Set).List() {
		action := expandDynamoDBv2Action([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{DynamoDBv2: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("elasticsearch").(*schema.Set).List() {
		action := expandElasticsearchAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Elasticsearch: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("firehose").(*schema.Set).List() {
		action := expandFirehoseAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Firehose: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("http").(*schema.Set).List() {
		action := expandHTTPAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Http: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("iot_analytics").(*schema.Set).List() {
		action := expandAnalyticsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{IotAnalytics: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("iot_events").(*schema.Set).List() {
		action := expandEventsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{IotEvents: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("kafka").(*schema.Set).List() {
		action := expandKafkaAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Kafka: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("kinesis").(*schema.Set).List() {
		action := expandKinesisAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Kinesis: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("lambda").(*schema.Set).List() {
		action := expandLambdaAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Lambda: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("republish").(*schema.Set).List() {
		action := expandRepublishAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Republish: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("s3").(*schema.Set).List() {
		action := expandS3Action([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{S3: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("sns").(*schema.Set).List() {
		action := expandSNSAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Sns: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("sqs").(*schema.Set).List() {
		action := expandSQSAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Sqs: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("step_functions").(*schema.Set).List() {
		action := expandStepFunctionsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{StepFunctions: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("timestream").(*schema.Set).List() {
		action := expandTimestreamAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, awstypes.Action{Timestream: action})
	}

	// Prevent sending empty Actions:
	// - missing required field, CreateTopicRuleInput.TopicRulePayload.Actions
	if len(actions) == 0 {
		actions = []awstypes.Action{}
	}

	var iotErrorAction *awstypes.Action
	if errorAction := d.Get("error_action").([]interface{}); len(errorAction) > 0 {
		for k, v := range errorAction[0].(map[string]interface{}) {
			switch k {
			case "cloudwatch_alarm":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandCloudWatchAlarmAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{CloudwatchAlarm: action}
				}
			case "cloudwatch_logs":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandCloudWatchLogsAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{CloudwatchLogs: action}
				}
			case "cloudwatch_metric":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandCloudWatchMetricAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{CloudwatchMetric: action}
				}
			case "dynamodb":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandDynamoDBAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{DynamoDB: action}
				}
			case "dynamodbv2":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandDynamoDBv2Action([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{DynamoDBv2: action}
				}
			case "elasticsearch":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandElasticsearchAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Elasticsearch: action}
				}
			case "firehose":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandFirehoseAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Firehose: action}
				}
			case "http":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandHTTPAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Http: action}
				}
			case "iot_analytics":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandAnalyticsAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{IotAnalytics: action}
				}
			case "iot_events":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandEventsAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{IotEvents: action}
				}
			case "kafka":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandKafkaAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Kafka: action}
				}
			case "kinesis":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandKinesisAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Kinesis: action}
				}
			case "lambda":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandLambdaAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Lambda: action}
				}
			case "republish":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandRepublishAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Republish: action}
				}
			case "s3":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandS3Action([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{S3: action}
				}
			case "sns":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandSNSAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Sns: action}
				}
			case "sqs":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandSQSAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Sqs: action}
				}
			case "step_functions":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandStepFunctionsAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{StepFunctions: action}
				}
			case "timestream":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandTimestreamAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &awstypes.Action{Timestream: action}
				}
			}
		}
	}

	return &awstypes.TopicRulePayload{
		Actions:          actions,
		AwsIotSqlVersion: aws.String(d.Get("sql_version").(string)),
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		ErrorAction:      iotErrorAction,
		RuleDisabled:     aws.Bool(!d.Get(names.AttrEnabled).(bool)),
		Sql:              aws.String(d.Get("sql").(string)),
	}
}

func flattenCloudWatchAlarmAction(apiObject *awstypes.CloudwatchAlarmAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.AlarmName; v != nil {
		tfMap["alarm_name"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.StateReason; v != nil {
		tfMap["state_reason"] = aws.ToString(v)
	}

	if v := apiObject.StateValue; v != nil {
		tfMap["state_value"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenCloudWatchAlarmActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.CloudwatchAlarm; v != nil {
			tfList = append(tfList, flattenCloudWatchAlarmAction(v)...)
		}
	}

	return tfList
}

func flattenCloudWatchLogsAction(apiObject *awstypes.CloudwatchLogsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.BatchMode; v != nil {
		tfMap["batch_mode"] = aws.ToBool(v)
	}

	if v := apiObject.LogGroupName; v != nil {
		tfMap[names.AttrLogGroupName] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenCloudWatchLogsActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.CloudwatchLogs; v != nil {
			tfList = append(tfList, flattenCloudWatchLogsAction(v)...)
		}
	}

	return tfList
}

// Legacy root attribute handling
func flattenCloudWatchMetricActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.CloudwatchMetric; v != nil {
			tfList = append(tfList, flattenCloudWatchMetricAction(v)...)
		}
	}

	return tfList
}

func flattenCloudWatchMetricAction(apiObject *awstypes.CloudwatchMetricAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.MetricName; v != nil {
		tfMap[names.AttrMetricName] = aws.ToString(v)
	}

	if v := apiObject.MetricNamespace; v != nil {
		tfMap["metric_namespace"] = aws.ToString(v)
	}

	if v := apiObject.MetricTimestamp; v != nil {
		tfMap["metric_timestamp"] = aws.ToString(v)
	}

	if v := apiObject.MetricUnit; v != nil {
		tfMap["metric_unit"] = aws.ToString(v)
	}

	if v := apiObject.MetricValue; v != nil {
		tfMap["metric_value"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenDynamoDBActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.DynamoDB; v != nil {
			tfList = append(tfList, flattenDynamoDBAction(v)...)
		}
	}

	return tfList
}

func flattenDynamoDBAction(apiObject *awstypes.DynamoDBAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.HashKeyField; v != nil {
		tfMap["hash_key_field"] = aws.ToString(v)
	}

	tfMap["hash_key_type"] = apiObject.HashKeyType

	if v := apiObject.HashKeyValue; v != nil {
		tfMap["hash_key_value"] = aws.ToString(v)
	}

	if v := apiObject.PayloadField; v != nil {
		tfMap["payload_field"] = aws.ToString(v)
	}

	if v := apiObject.Operation; v != nil {
		tfMap["operation"] = aws.ToString(v)
	}

	if v := apiObject.RangeKeyField; v != nil {
		tfMap["range_key_field"] = aws.ToString(v)
	}

	tfMap["range_key_type"] = apiObject.RangeKeyType

	if v := apiObject.RangeKeyValue; v != nil {
		tfMap["range_key_value"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.TableName; v != nil {
		tfMap[names.AttrTableName] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenDynamoDBv2Actions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.DynamoDBv2; v != nil {
			tfList = append(tfList, flattenDynamoDBv2Action(v)...)
		}
	}

	return tfList
}

func flattenDynamoDBv2Action(apiObject *awstypes.DynamoDBv2Action) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.PutItem; v != nil {
		tfMap["put_item"] = flattenPutItemInput(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenElasticsearchActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Elasticsearch; v != nil {
			tfList = append(tfList, flattenElasticsearchAction(v)...)
		}
	}

	return tfList
}

func flattenElasticsearchAction(apiObject *awstypes.ElasticsearchAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Endpoint; v != nil {
		tfMap[names.AttrEndpoint] = aws.ToString(v)
	}

	if v := apiObject.Id; v != nil {
		tfMap[names.AttrID] = aws.ToString(v)
	}

	if v := apiObject.Index; v != nil {
		tfMap["index"] = aws.ToString(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenFirehoseActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Firehose; v != nil {
			tfList = append(tfList, flattenFirehoseAction(v)...)
		}
	}

	return tfList
}

func flattenFirehoseAction(apiObject *awstypes.FirehoseAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.BatchMode; v != nil {
		tfMap["batch_mode"] = aws.ToBool(v)
	}

	if v := apiObject.DeliveryStreamName; v != nil {
		tfMap["delivery_stream_name"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.Separator; v != nil {
		tfMap["separator"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenHTTPActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Http; v != nil {
			tfList = append(tfList, flattenHTTPAction(v)...)
		}
	}

	return tfList
}

func flattenHTTPAction(apiObject *awstypes.HttpAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Url; v != nil {
		tfMap[names.AttrURL] = aws.ToString(v)
	}

	if v := apiObject.ConfirmationUrl; v != nil {
		tfMap["confirmation_url"] = aws.ToString(v)
	}

	if v := apiObject.Headers; v != nil {
		headers := []map[string]string{}

		for _, h := range v {
			m := map[string]string{
				names.AttrKey:   aws.ToString(h.Key),
				names.AttrValue: aws.ToString(h.Value),
			}
			headers = append(headers, m)
		}
		tfMap["http_header"] = headers
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenAnalyticsActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.IotAnalytics; v != nil {
			tfList = append(tfList, flattenAnalyticsAction(v)...)
		}
	}

	return tfList
}

func flattenAnalyticsAction(apiObject *awstypes.IotAnalyticsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.BatchMode; v != nil {
		tfMap["batch_mode"] = aws.ToBool(v)
	}

	if v := apiObject.ChannelName; v != nil {
		tfMap["channel_name"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenEventsActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.IotEvents; v != nil {
			tfList = append(tfList, flattenEventsAction(v)...)
		}
	}

	return tfList
}

func flattenEventsAction(apiObject *awstypes.IotEventsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.BatchMode; v != nil {
		tfMap["batch_mode"] = aws.ToBool(v)
	}

	if v := apiObject.InputName; v != nil {
		tfMap["input_name"] = aws.ToString(v)
	}

	if v := apiObject.MessageId; v != nil {
		tfMap["message_id"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenKafkaActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Kafka; v != nil {
			tfList = append(tfList, flattenKafkaAction(v)...)
		}
	}

	return tfList
}

func flattenKafkaAction(apiObject *awstypes.KafkaAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.ClientProperties; v != nil {
		tfMap["client_properties"] = aws.StringMap(v)
	}

	if v := apiObject.DestinationArn; v != nil {
		tfMap[names.AttrDestinationARN] = aws.ToString(v)
	}

	if v := apiObject.Headers; v != nil {
		tfMap[names.AttrHeader] = flattenKafkaHeaders(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	if v := apiObject.Partition; v != nil {
		tfMap["partition"] = aws.ToString(v)
	}

	if v := apiObject.Topic; v != nil {
		tfMap["topic"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenKafkaHeaders(apiObjects []awstypes.KafkaActionHeader) []interface{} {
	results := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{})

		if v := apiObject.Key; v != nil {
			tfMap[names.AttrKey] = aws.ToString(v)
		}

		if v := apiObject.Value; v != nil {
			tfMap[names.AttrValue] = aws.ToString(v)
		}
		results = append(results, tfMap)
	}

	return results
}

// Legacy root attribute handling
func flattenKinesisActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Kinesis; v != nil {
			tfList = append(tfList, flattenKinesisAction(v)...)
		}
	}

	return tfList
}

func flattenKinesisAction(apiObject *awstypes.KinesisAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.PartitionKey; v != nil {
		tfMap["partition_key"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.StreamName; v != nil {
		tfMap["stream_name"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenLambdaActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Lambda; v != nil {
			tfList = append(tfList, flattenLambdaAction(v)...)
		}
	}

	return tfList
}

func flattenLambdaAction(apiObject *awstypes.LambdaAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.FunctionArn; v != nil {
		tfMap[names.AttrFunctionARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenPutItemInput(apiObject *awstypes.PutItemInput) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.TableName; v != nil {
		tfMap[names.AttrTableName] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenRepublishActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Republish; v != nil {
			tfList = append(tfList, flattenRepublishAction(v)...)
		}
	}

	return tfList
}

func flattenRepublishAction(apiObject *awstypes.RepublishAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Qos; v != nil {
		tfMap["qos"] = aws.ToInt32(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.Topic; v != nil {
		tfMap["topic"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenS3Actions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.S3; v != nil {
			tfList = append(tfList, flattenS3Action(v)...)
		}
	}

	return tfList
}

func flattenS3Action(apiObject *awstypes.S3Action) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucketName] = aws.ToString(v)
	}

	tfMap["canned_acl"] = apiObject.CannedAcl

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenSNSActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Sns; v != nil {
			tfList = append(tfList, flattenSNSAction(v)...)
		}
	}

	return tfList
}

func flattenSNSAction(apiObject *awstypes.SnsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	tfMap["message_format"] = apiObject.MessageFormat

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.TargetArn; v != nil {
		tfMap[names.AttrTargetARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenSQSActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Sqs; v != nil {
			tfList = append(tfList, flattenSQSAction(v)...)
		}
	}

	return tfList
}

func flattenSQSAction(apiObject *awstypes.SqsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.QueueUrl; v != nil {
		tfMap["queue_url"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.UseBase64; v != nil {
		tfMap["use_base64"] = aws.ToBool(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenStepFunctionsActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.StepFunctions; v != nil {
			tfList = append(tfList, flattenStepFunctionsAction(v)...)
		}
	}

	return tfList
}

func flattenStepFunctionsAction(apiObject *awstypes.StepFunctionsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.ExecutionNamePrefix; v != nil {
		tfMap["execution_name_prefix"] = aws.ToString(v)
	}

	if v := apiObject.StateMachineName; v != nil {
		tfMap["state_machine_name"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenTimestreamActions(apiObjects []awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		if v := apiObject.Timestream; v != nil {
			tfList = append(tfList, flattenTimestreamAction(v)...)
		}
	}

	return tfList
}

func flattenTimestreamAction(apiObject *awstypes.TimestreamAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.DatabaseName; v != nil {
		tfMap[names.AttrDatabaseName] = aws.ToString(v)
	}

	if v := apiObject.Dimensions; v != nil {
		tfMap["dimension"] = flattenTimestreamDimensions(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.TableName; v != nil {
		tfMap[names.AttrTableName] = aws.ToString(v)
	}

	if v := apiObject.Timestamp; v != nil {
		tfMap["timestamp"] = flattenTimestreamTimestamp(v)
	}

	return []interface{}{tfMap}
}

func flattenTimestreamDimensions(apiObjects []awstypes.TimestreamDimension) []interface{} {
	if apiObjects == nil {
		return nil
	}

	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{})

		if v := apiObject.Name; v != nil {
			tfMap[names.AttrName] = aws.ToString(v)
		}

		if v := apiObject.Value; v != nil {
			tfMap[names.AttrValue] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTimestreamTimestamp(apiObject *awstypes.TimestreamTimestamp) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Unit; v != nil {
		tfMap[names.AttrUnit] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenErrorAction(apiObject *awstypes.Action) []interface{} {
	tfList := make([]interface{}, 0)

	if apiObject == nil {
		return nil
	}

	input := []awstypes.Action{*apiObject}
	if apiObject.CloudwatchAlarm != nil {
		tfList = append(tfList, map[string]interface{}{"cloudwatch_alarm": flattenCloudWatchAlarmActions(input)})
		return tfList
	}
	if apiObject.CloudwatchLogs != nil {
		tfList = append(tfList, map[string]interface{}{names.AttrCloudWatchLogs: flattenCloudWatchLogsActions(input)})
		return tfList
	}
	if apiObject.CloudwatchMetric != nil {
		tfList = append(tfList, map[string]interface{}{"cloudwatch_metric": flattenCloudWatchMetricActions(input)})
		return tfList
	}
	if apiObject.DynamoDB != nil {
		tfList = append(tfList, map[string]interface{}{"dynamodb": flattenDynamoDBActions(input)})
		return tfList
	}
	if apiObject.DynamoDBv2 != nil {
		tfList = append(tfList, map[string]interface{}{"dynamodbv2": flattenDynamoDBv2Actions(input)})
		return tfList
	}
	if apiObject.Elasticsearch != nil {
		tfList = append(tfList, map[string]interface{}{"elasticsearch": flattenElasticsearchActions(input)})
		return tfList
	}
	if apiObject.Firehose != nil {
		tfList = append(tfList, map[string]interface{}{"firehose": flattenFirehoseActions(input)})
		return tfList
	}
	if apiObject.Http != nil {
		tfList = append(tfList, map[string]interface{}{"http": flattenHTTPActions(input)})
		return tfList
	}
	if apiObject.IotAnalytics != nil {
		tfList = append(tfList, map[string]interface{}{"iot_analytics": flattenAnalyticsActions(input)})
		return tfList
	}
	if apiObject.IotEvents != nil {
		tfList = append(tfList, map[string]interface{}{"iot_events": flattenEventsActions(input)})
		return tfList
	}
	if apiObject.Kafka != nil {
		tfList = append(tfList, map[string]interface{}{"kafka": flattenKafkaActions(input)})
		return tfList
	}
	if apiObject.Kinesis != nil {
		tfList = append(tfList, map[string]interface{}{"kinesis": flattenKinesisActions(input)})
		return tfList
	}
	if apiObject.Lambda != nil {
		tfList = append(tfList, map[string]interface{}{"lambda": flattenLambdaActions(input)})
		return tfList
	}
	if apiObject.Republish != nil {
		tfList = append(tfList, map[string]interface{}{"republish": flattenRepublishActions(input)})
		return tfList
	}
	if apiObject.S3 != nil {
		tfList = append(tfList, map[string]interface{}{"s3": flattenS3Actions(input)})
		return tfList
	}
	if apiObject.Sns != nil {
		tfList = append(tfList, map[string]interface{}{"sns": flattenSNSActions(input)})
		return tfList
	}
	if apiObject.Sqs != nil {
		tfList = append(tfList, map[string]interface{}{"sqs": flattenSQSActions(input)})
		return tfList
	}
	if apiObject.StepFunctions != nil {
		tfList = append(tfList, map[string]interface{}{"step_functions": flattenStepFunctionsActions(input)})
		return tfList
	}
	if apiObject.Timestream != nil {
		tfList = append(tfList, map[string]interface{}{"timestream": flattenTimestreamActions(input)})
		return tfList
	}

	return tfList
}
