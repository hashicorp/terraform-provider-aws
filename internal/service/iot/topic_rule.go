package iot

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTopicRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceTopicRuleCreate,
		Read:   resourceTopicRuleRead,
		Update: resourceTopicRuleUpdate,
		Delete: resourceTopicRuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
						"role_arn": {
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
			"cloudwatch_logs": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
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
						"metric_name": {
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
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"description": {
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
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"table_name": {
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
									"table_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"role_arn": {
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
						"endpoint": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validTopicRuleElasticsearchEndpoint,
						},
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"index": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"enabled": {
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
									"role_arn": {
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
						"cloudwatch_logs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_group_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
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
									"metric_name": {
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
									"role_arn": {
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
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"table_name": {
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
												"table_name": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"role_arn": {
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
									"endpoint": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validTopicRuleElasticsearchEndpoint,
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"index": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"type": {
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
									"delivery_stream_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
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
												"key": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"url": {
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
									"channel_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
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
									"input_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"message_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"role_arn": {
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
									"destination_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"key": {
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
									"role_arn": {
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
									"function_arn": {
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
									"role_arn": {
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
									"bucket_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"canned_acl": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(iot.CannedAccessControlList_Values(), false),
									},
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
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
										Default:  iot.MessageFormatRaw,
										Optional: true,
									},
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"target_arn": {
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
									"role_arn": {
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
									"role_arn": {
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
									"database_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"dimension": {
										Type:     schema.TypeSet,
										Required: true,
										Elem:     timestreamDimensionResource,
									},
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"table_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"timestamp": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"unit": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														"SECONDS",
														"MILLISECONDS",
														"MICROSECONDS",
														"NANOSECONDS",
													}, false),
												},
												"value": {
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
						"delivery_stream_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
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
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"url": {
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
						"channel_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
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
						"input_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"message_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
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
						"destination_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"key": {
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
						"role_arn": {
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
						"function_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"name": {
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
						"role_arn": {
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
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"canned_acl": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(iot.CannedAccessControlList_Values(), false),
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
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
							Default:  iot.MessageFormatRaw,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"target_arn": {
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
						"role_arn": {
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
						"role_arn": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"timestream": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"dimension": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     timestreamDimensionResource,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"table_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"timestamp": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"unit": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"SECONDS",
											"MILLISECONDS",
											"MICROSECONDS",
											"NANOSECONDS",
										}, false),
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

var topicRuleErrorActionExactlyOneOf = []string{
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

var timestreamDimensionResource *schema.Resource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"value": {
			Type:     schema.TypeString,
			Required: true,
		},
	},
}

func resourceTopicRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	ruleName := d.Get("name").(string)
	input := &iot.CreateTopicRuleInput{
		RuleName:         aws.String(ruleName),
		Tags:             aws.String(tags.IgnoreAWS().URLQueryString()),
		TopicRulePayload: expandTopicRulePayload(d),
	}

	log.Printf("[INFO] Creating IoT Topic Rule: %s", input)
	_, err := tfresource.RetryWhenAWSErrMessageContains(propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateTopicRule(input)
		},
		iot.ErrCodeInvalidRequestException, "sts:AssumeRole")

	if err != nil {
		return fmt.Errorf("creating IoT Topic Rule (%s): %w", ruleName, err)
	}

	d.SetId(ruleName)

	return resourceTopicRuleRead(d, meta)
}

func resourceTopicRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindTopicRuleByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Topic Rule %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading IoT Topic Rule (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.RuleArn)
	d.Set("name", output.Rule.RuleName)
	d.Set("description", output.Rule.Description)
	d.Set("enabled", !aws.BoolValue(output.Rule.RuleDisabled))
	d.Set("sql", output.Rule.Sql)
	d.Set("sql_version", output.Rule.AwsIotSqlVersion)

	if err := d.Set("cloudwatch_alarm", flattenCloudWatchAlarmActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting cloudwatch_alarm: %w", err)
	}

	if err := d.Set("cloudwatch_logs", flattenCloudWatchLogsActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting cloudwatch_logs: %w", err)
	}

	if err := d.Set("cloudwatch_metric", flattenCloudWatchMetricActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting cloudwatch_metric: %w", err)
	}

	if err := d.Set("dynamodb", flattenDynamoDBActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting dynamodb: %w", err)
	}

	if err := d.Set("dynamodbv2", flattenDynamoDBv2Actions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting dynamodbv2: %w", err)
	}

	if err := d.Set("elasticsearch", flattenElasticsearchActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting elasticsearch: %w", err)
	}

	if err := d.Set("firehose", flattenFirehoseActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting firehose: %w", err)
	}

	if err := d.Set("http", flattenHTTPActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting http: %w", err)
	}

	if err := d.Set("iot_analytics", flattenAnalyticsActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting iot_analytics: %w", err)
	}

	if err := d.Set("iot_events", flattenEventsActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting iot_events: %w", err)
	}

	if err := d.Set("kafka", flattenKafkaActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting kafka: %w", err)
	}

	if err := d.Set("kinesis", flattenKinesisActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting kinesis: %w", err)
	}

	if err := d.Set("lambda", flattenLambdaActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting lambda: %w", err)
	}

	if err := d.Set("republish", flattenRepublishActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting republish: %w", err)
	}

	if err := d.Set("s3", flattenS3Actions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting s3: %w", err)
	}

	if err := d.Set("sns", flattenSNSActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting sns: %w", err)
	}

	if err := d.Set("sqs", flattenSQSActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting sqs: %w", err)
	}

	if err := d.Set("step_functions", flattenStepFunctionsActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting step_functions: %w", err)
	}

	if err := d.Set("timestream", flattenTimestreamActions(output.Rule.Actions)); err != nil {
		return fmt.Errorf("setting timestream: %w", err)
	}

	if err := d.Set("error_action", flattenErrorAction(output.Rule.ErrorAction)); err != nil {
		return fmt.Errorf("setting error_action: %w", err)
	}

	tags, err := ListTags(conn, aws.StringValue(output.RuleArn))

	if err != nil {
		return fmt.Errorf("listing tags for IoT Topic Rule (%s): %w", aws.StringValue(output.RuleArn), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceTopicRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &iot.ReplaceTopicRuleInput{
			RuleName:         aws.String(d.Get("name").(string)),
			TopicRulePayload: expandTopicRulePayload(d),
		}

		log.Printf("[INFO] Replacing IoT Topic Rule: %s", input)
		_, err := conn.ReplaceTopicRule(input)

		if err != nil {
			return fmt.Errorf("replacing IoT Topic Rule (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating tags: %w", err)
		}
	}

	return resourceTopicRuleRead(d, meta)
}

func resourceTopicRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	log.Printf("[INFO] Deleting IoT Topic Rule: %s", d.Id())
	_, err := conn.DeleteTopicRule(&iot.DeleteTopicRuleInput{
		RuleName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("deleting IoT Topic Rule (%s): %w", d.Id(), err)
	}

	return nil
}

func expandPutItemInput(tfList []interface{}) *iot.PutItemInput {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.PutItemInput{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["table_name"].(string); ok && v != "" {
		apiObject.TableName = aws.String(v)
	}

	return apiObject
}

func expandCloudWatchAlarmAction(tfList []interface{}) *iot.CloudwatchAlarmAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.CloudwatchAlarmAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["alarm_name"].(string); ok && v != "" {
		apiObject.AlarmName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
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

func expandCloudWatchLogsAction(tfList []interface{}) *iot.CloudwatchLogsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.CloudwatchLogsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["log_group_name"].(string); ok && v != "" {
		apiObject.LogGroupName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandCloudWatchMetricAction(tfList []interface{}) *iot.CloudwatchMetricAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.CloudwatchMetricAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["metric_name"].(string); ok && v != "" {
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

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandDynamoDBAction(tfList []interface{}) *iot.DynamoDBAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.DynamoDBAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["hash_key_field"].(string); ok && v != "" {
		apiObject.HashKeyField = aws.String(v)
	}

	if v, ok := tfMap["hash_key_type"].(string); ok && v != "" {
		apiObject.HashKeyType = aws.String(v)
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
		apiObject.RangeKeyType = aws.String(v)
	}

	if v, ok := tfMap["range_key_value"].(string); ok && v != "" {
		apiObject.RangeKeyValue = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["table_name"].(string); ok && v != "" {
		apiObject.TableName = aws.String(v)
	}

	return apiObject
}

func expandDynamoDBv2Action(tfList []interface{}) *iot.DynamoDBv2Action {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.DynamoDBv2Action{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["put_item"].([]interface{}); ok {
		apiObject.PutItem = expandPutItemInput(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandElasticsearchAction(tfList []interface{}) *iot.ElasticsearchAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.ElasticsearchAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["endpoint"].(string); ok && v != "" {
		apiObject.Endpoint = aws.String(v)
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap["index"].(string); ok && v != "" {
		apiObject.Index = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandFirehoseAction(tfList []interface{}) *iot.FirehoseAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.FirehoseAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["delivery_stream_name"].(string); ok && v != "" {
		apiObject.DeliveryStreamName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["separator"].(string); ok && v != "" {
		apiObject.Separator = aws.String(v)
	}

	return apiObject
}

func expandHTTPAction(tfList []interface{}) *iot.HttpAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.HttpAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["url"].(string); ok && v != "" {
		apiObject.Url = aws.String(v)
	}

	if v, ok := tfMap["confirmation_url"].(string); ok && v != "" {
		apiObject.ConfirmationUrl = aws.String(v)
	}

	if v, ok := tfMap["http_header"].([]interface{}); ok {
		headerObjs := []*iot.HttpActionHeader{}
		for _, val := range v {
			if m, ok := val.(map[string]interface{}); ok {
				headerObj := &iot.HttpActionHeader{}
				if v, ok := m["key"].(string); ok && v != "" {
					headerObj.Key = aws.String(v)
				}
				if v, ok := m["value"].(string); ok && v != "" {
					headerObj.Value = aws.String(v)
				}
				headerObjs = append(headerObjs, headerObj)
			}
		}
		apiObject.Headers = headerObjs
	}

	return apiObject
}

func expandAnalyticsAction(tfList []interface{}) *iot.IotAnalyticsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.IotAnalyticsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["channel_name"].(string); ok && v != "" {
		apiObject.ChannelName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandEventsAction(tfList []interface{}) *iot.IotEventsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.IotEventsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["input_name"].(string); ok && v != "" {
		apiObject.InputName = aws.String(v)
	}

	if v, ok := tfMap["message_id"].(string); ok && v != "" {
		apiObject.MessageId = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandKafkaAction(tfList []interface{}) *iot.KafkaAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.KafkaAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["client_properties"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.ClientProperties = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["destination_arn"].(string); ok && v != "" {
		apiObject.DestinationArn = aws.String(v)
	}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["partition"].(string); ok && v != "" {
		apiObject.Partition = aws.String(v)
	}

	if v, ok := tfMap["topic"].(string); ok && v != "" {
		apiObject.Topic = aws.String(v)
	}

	return apiObject
}

func expandKinesisAction(tfList []interface{}) *iot.KinesisAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.KinesisAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["partition_key"].(string); ok && v != "" {
		apiObject.PartitionKey = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["stream_name"].(string); ok && v != "" {
		apiObject.StreamName = aws.String(v)
	}

	return apiObject
}

func expandLambdaAction(tfList []interface{}) *iot.LambdaAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.LambdaAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["function_arn"].(string); ok && v != "" {
		apiObject.FunctionArn = aws.String(v)
	}

	return apiObject
}

func expandRepublishAction(tfList []interface{}) *iot.RepublishAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.RepublishAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["qos"].(int); ok {
		apiObject.Qos = aws.Int64(int64(v))
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["topic"].(string); ok && v != "" {
		apiObject.Topic = aws.String(v)
	}

	return apiObject
}

func expandS3Action(tfList []interface{}) *iot.S3Action {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.S3Action{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		apiObject.BucketName = aws.String(v)
	}

	if v, ok := tfMap["canned_acl"].(string); ok && v != "" {
		apiObject.CannedAcl = aws.String(v)
	}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandSNSAction(tfList []interface{}) *iot.SnsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.SnsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["message_format"].(string); ok && v != "" {
		apiObject.MessageFormat = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["target_arn"].(string); ok && v != "" {
		apiObject.TargetArn = aws.String(v)
	}

	return apiObject
}

func expandSQSAction(tfList []interface{}) *iot.SqsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.SqsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["queue_url"].(string); ok && v != "" {
		apiObject.QueueUrl = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["use_base64"].(bool); ok {
		apiObject.UseBase64 = aws.Bool(v)
	}

	return apiObject
}

func expandStepFunctionsAction(tfList []interface{}) *iot.StepFunctionsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.StepFunctionsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["execution_name_prefix"].(string); ok && v != "" {
		apiObject.ExecutionNamePrefix = aws.String(v)
	}

	if v, ok := tfMap["state_machine_name"].(string); ok && v != "" {
		apiObject.StateMachineName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandTimestreamAction(tfList []interface{}) *iot.TimestreamAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.TimestreamAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["dimension"].(*schema.Set); ok {
		apiObject.Dimensions = expandTimestreamDimensions(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["table_name"].(string); ok && v != "" {
		apiObject.TableName = aws.String(v)
	}

	if v, ok := tfMap["timestamp"].([]interface{}); ok {
		apiObject.Timestamp = expandTimestreamTimestamp(v)
	}

	return apiObject
}

func expandTimestreamDimensions(tfSet *schema.Set) []*iot.TimestreamDimension {
	if tfSet == nil || tfSet.Len() == 0 {
		return nil
	}

	apiObjects := make([]*iot.TimestreamDimension, tfSet.Len())
	for i, elem := range tfSet.List() {
		if tfMap, ok := elem.(map[string]interface{}); ok {
			apiObject := &iot.TimestreamDimension{}

			if v, ok := tfMap["name"].(string); ok && v != "" {
				apiObject.Name = aws.String(v)
			}

			if v, ok := tfMap["value"].(string); ok && v != "" {
				apiObject.Value = aws.String(v)
			}

			apiObjects[i] = apiObject
		}
	}

	return apiObjects
}

func expandTimestreamTimestamp(tfList []interface{}) *iot.TimestreamTimestamp {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.TimestreamTimestamp{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["unit"].(string); ok && v != "" {
		apiObject.Unit = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandTopicRulePayload(d *schema.ResourceData) *iot.TopicRulePayload {
	var actions []*iot.Action

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("cloudwatch_alarm").(*schema.Set).List() {
		action := expandCloudWatchAlarmAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{CloudwatchAlarm: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("cloudwatch_logs").(*schema.Set).List() {
		action := expandCloudWatchLogsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{CloudwatchLogs: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("cloudwatch_metric").(*schema.Set).List() {
		action := expandCloudWatchMetricAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{CloudwatchMetric: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("dynamodb").(*schema.Set).List() {
		action := expandDynamoDBAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{DynamoDB: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("dynamodbv2").(*schema.Set).List() {
		action := expandDynamoDBv2Action([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{DynamoDBv2: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("elasticsearch").(*schema.Set).List() {
		action := expandElasticsearchAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Elasticsearch: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("firehose").(*schema.Set).List() {
		action := expandFirehoseAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Firehose: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("http").(*schema.Set).List() {
		action := expandHTTPAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Http: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("iot_analytics").(*schema.Set).List() {
		action := expandAnalyticsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{IotAnalytics: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("iot_events").(*schema.Set).List() {
		action := expandEventsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{IotEvents: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("kafka").(*schema.Set).List() {
		action := expandKafkaAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Kafka: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("kinesis").(*schema.Set).List() {
		action := expandKinesisAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Kinesis: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("lambda").(*schema.Set).List() {
		action := expandLambdaAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Lambda: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("republish").(*schema.Set).List() {
		action := expandRepublishAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Republish: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("s3").(*schema.Set).List() {
		action := expandS3Action([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{S3: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("sns").(*schema.Set).List() {
		action := expandSNSAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Sns: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("sqs").(*schema.Set).List() {
		action := expandSQSAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Sqs: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("step_functions").(*schema.Set).List() {
		action := expandStepFunctionsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{StepFunctions: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("timestream").(*schema.Set).List() {
		action := expandTimestreamAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Timestream: action})
	}

	// Prevent sending empty Actions:
	// - missing required field, CreateTopicRuleInput.TopicRulePayload.Actions
	if len(actions) == 0 {
		actions = []*iot.Action{}
	}

	var iotErrorAction *iot.Action
	errorAction := d.Get("error_action").([]interface{})
	if len(errorAction) > 0 {
		for k, v := range errorAction[0].(map[string]interface{}) {
			switch k {
			case "cloudwatch_alarm":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandCloudWatchAlarmAction([]interface{}{tfMapRaw})
					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{CloudwatchAlarm: action}

				}
			case "cloudwatch_logs":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandCloudWatchLogsAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{CloudwatchLogs: action}
				}
			case "cloudwatch_metric":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandCloudWatchMetricAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{CloudwatchMetric: action}
				}
			case "dynamodb":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandDynamoDBAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{DynamoDB: action}
				}
			case "dynamodbv2":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandDynamoDBv2Action([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{DynamoDBv2: action}
				}
			case "elasticsearch":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandElasticsearchAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Elasticsearch: action}
				}
			case "firehose":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandFirehoseAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Firehose: action}
				}
			case "http":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandHTTPAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Http: action}
				}
			case "iot_analytics":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandAnalyticsAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{IotAnalytics: action}
				}
			case "iot_events":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandEventsAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{IotEvents: action}
				}
			case "kafka":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandKafkaAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Kafka: action}
				}
			case "kinesis":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandKinesisAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Kinesis: action}
				}
			case "lambda":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandLambdaAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Lambda: action}
				}
			case "republish":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandRepublishAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Republish: action}
				}
			case "s3":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandS3Action([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{S3: action}
				}
			case "sns":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandSNSAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Sns: action}
				}
			case "sqs":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandSQSAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Sqs: action}
				}
			case "step_functions":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandStepFunctionsAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{StepFunctions: action}
				}
			case "timestream":
				for _, tfMapRaw := range v.([]interface{}) {
					action := expandTimestreamAction([]interface{}{tfMapRaw})

					if action == nil {
						continue
					}

					iotErrorAction = &iot.Action{Timestream: action}
				}
			}
		}
	}

	return &iot.TopicRulePayload{
		Actions:          actions,
		AwsIotSqlVersion: aws.String(d.Get("sql_version").(string)),
		Description:      aws.String(d.Get("description").(string)),
		ErrorAction:      iotErrorAction,
		RuleDisabled:     aws.Bool(!d.Get("enabled").(bool)),
		Sql:              aws.String(d.Get("sql").(string)),
	}
}

func flattenCloudWatchAlarmAction(apiObject *iot.CloudwatchAlarmAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.AlarmName; v != nil {
		tfMap["alarm_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.StateReason; v != nil {
		tfMap["state_reason"] = aws.StringValue(v)
	}

	if v := apiObject.StateValue; v != nil {
		tfMap["state_value"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenCloudWatchAlarmActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.CloudwatchAlarm; v != nil {
			results = append(results, flattenCloudWatchAlarmAction(v)...)
		}
	}

	return results
}

func flattenCloudWatchLogsAction(apiObject *iot.CloudwatchLogsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.LogGroupName; v != nil {
		tfMap["log_group_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenCloudWatchLogsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.CloudwatchLogs; v != nil {
			results = append(results, flattenCloudWatchLogsAction(v)...)
		}
	}

	return results
}

// Legacy root attribute handling
func flattenCloudWatchMetricActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.CloudwatchMetric; v != nil {
			results = append(results, flattenCloudWatchMetricAction(v)...)
		}
	}

	return results
}

func flattenCloudWatchMetricAction(apiObject *iot.CloudwatchMetricAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.MetricName; v != nil {
		tfMap["metric_name"] = aws.StringValue(v)
	}

	if v := apiObject.MetricNamespace; v != nil {
		tfMap["metric_namespace"] = aws.StringValue(v)
	}

	if v := apiObject.MetricTimestamp; v != nil {
		tfMap["metric_timestamp"] = aws.StringValue(v)
	}

	if v := apiObject.MetricUnit; v != nil {
		tfMap["metric_unit"] = aws.StringValue(v)
	}

	if v := apiObject.MetricValue; v != nil {
		tfMap["metric_value"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenDynamoDBActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.DynamoDB; v != nil {
			results = append(results, flattenDynamoDBAction(v)...)
		}
	}

	return results
}

func flattenDynamoDBAction(apiObject *iot.DynamoDBAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.HashKeyField; v != nil {
		tfMap["hash_key_field"] = aws.StringValue(v)
	}

	if v := apiObject.HashKeyType; v != nil {
		tfMap["hash_key_type"] = aws.StringValue(v)
	}

	if v := apiObject.HashKeyValue; v != nil {
		tfMap["hash_key_value"] = aws.StringValue(v)
	}

	if v := apiObject.PayloadField; v != nil {
		tfMap["payload_field"] = aws.StringValue(v)
	}

	if v := apiObject.Operation; v != nil {
		tfMap["operation"] = aws.StringValue(v)
	}

	if v := apiObject.RangeKeyField; v != nil {
		tfMap["range_key_field"] = aws.StringValue(v)
	}

	if v := apiObject.RangeKeyType; v != nil {
		tfMap["range_key_type"] = aws.StringValue(v)
	}

	if v := apiObject.RangeKeyValue; v != nil {
		tfMap["range_key_value"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.TableName; v != nil {
		tfMap["table_name"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenDynamoDBv2Actions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.DynamoDBv2; v != nil {
			results = append(results, flattenDynamoDBv2Action(v)...)
		}
	}

	return results
}

func flattenDynamoDBv2Action(apiObject *iot.DynamoDBv2Action) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.PutItem; v != nil {
		tfMap["put_item"] = flattenPutItemInput(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenElasticsearchActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Elasticsearch; v != nil {
			results = append(results, flattenElasticsearchAction(v)...)
		}
	}

	return results
}

func flattenElasticsearchAction(apiObject *iot.ElasticsearchAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Endpoint; v != nil {
		tfMap["endpoint"] = aws.StringValue(v)
	}

	if v := apiObject.Id; v != nil {
		tfMap["id"] = aws.StringValue(v)
	}

	if v := apiObject.Index; v != nil {
		tfMap["index"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenFirehoseActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Firehose; v != nil {
			results = append(results, flattenFirehoseAction(v)...)
		}
	}

	return results
}

func flattenFirehoseAction(apiObject *iot.FirehoseAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.DeliveryStreamName; v != nil {
		tfMap["delivery_stream_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.Separator; v != nil {
		tfMap["separator"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenHTTPActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Http; v != nil {
			results = append(results, flattenHTTPAction(v)...)
		}
	}

	return results
}

func flattenHTTPAction(apiObject *iot.HttpAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Url; v != nil {
		tfMap["url"] = aws.StringValue(v)
	}

	if v := apiObject.ConfirmationUrl; v != nil {
		tfMap["confirmation_url"] = aws.StringValue(v)
	}

	if v := apiObject.Headers; v != nil {
		headers := []map[string]string{}

		for _, h := range v {
			m := map[string]string{
				"key":   aws.StringValue(h.Key),
				"value": aws.StringValue(h.Value),
			}
			headers = append(headers, m)
		}
		tfMap["http_header"] = headers
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenAnalyticsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.IotAnalytics; v != nil {
			results = append(results, flattenAnalyticsAction(v)...)
		}
	}

	return results
}

func flattenAnalyticsAction(apiObject *iot.IotAnalyticsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.ChannelName; v != nil {
		tfMap["channel_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenEventsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.IotEvents; v != nil {
			results = append(results, flattenEventsAction(v)...)
		}
	}

	return results
}

func flattenEventsAction(apiObject *iot.IotEventsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.InputName; v != nil {
		tfMap["input_name"] = aws.StringValue(v)
	}

	if v := apiObject.MessageId; v != nil {
		tfMap["message_id"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenKafkaActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Kafka; v != nil {
			results = append(results, flattenKafkaAction(v)...)
		}
	}

	return results
}

func flattenKafkaAction(apiObject *iot.KafkaAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.ClientProperties; v != nil {
		tfMap["client_properties"] = aws.StringValueMap(v)
	}

	if v := apiObject.DestinationArn; v != nil {
		tfMap["destination_arn"] = aws.StringValue(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap["key"] = aws.StringValue(v)
	}

	if v := apiObject.Partition; v != nil {
		tfMap["partition"] = aws.StringValue(v)
	}

	if v := apiObject.Topic; v != nil {
		tfMap["topic"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenKinesisActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Kinesis; v != nil {
			results = append(results, flattenKinesisAction(v)...)
		}
	}

	return results
}

func flattenKinesisAction(apiObject *iot.KinesisAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.PartitionKey; v != nil {
		tfMap["partition_key"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.StreamName; v != nil {
		tfMap["stream_name"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenLambdaActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Lambda; v != nil {
			results = append(results, flattenLambdaAction(v)...)
		}
	}

	return results
}

func flattenLambdaAction(apiObject *iot.LambdaAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.FunctionArn; v != nil {
		tfMap["function_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func flattenPutItemInput(apiObject *iot.PutItemInput) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.TableName; v != nil {
		tfMap["table_name"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenRepublishActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Republish; v != nil {
			results = append(results, flattenRepublishAction(v)...)
		}
	}

	return results
}

func flattenRepublishAction(apiObject *iot.RepublishAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Qos; v != nil {
		tfMap["qos"] = aws.Int64Value(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.Topic; v != nil {
		tfMap["topic"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenS3Actions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.S3; v != nil {
			results = append(results, flattenS3Action(v)...)
		}
	}

	return results
}

func flattenS3Action(apiObject *iot.S3Action) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.BucketName; v != nil {
		tfMap["bucket_name"] = aws.StringValue(v)
	}

	if v := apiObject.CannedAcl; v != nil {
		tfMap["canned_acl"] = aws.StringValue(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap["key"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenSNSActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Sns; v != nil {
			results = append(results, flattenSNSAction(v)...)
		}
	}

	return results
}

func flattenSNSAction(apiObject *iot.SnsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.MessageFormat; v != nil {
		tfMap["message_format"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.TargetArn; v != nil {
		tfMap["target_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenSQSActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Sqs; v != nil {
			results = append(results, flattenSQSAction(v)...)
		}
	}

	return results
}

func flattenSQSAction(apiObject *iot.SqsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.QueueUrl; v != nil {
		tfMap["queue_url"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.UseBase64; v != nil {
		tfMap["use_base64"] = aws.BoolValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenStepFunctionsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.StepFunctions; v != nil {
			results = append(results, flattenStepFunctionsAction(v)...)
		}
	}

	return results
}

func flattenStepFunctionsAction(apiObject *iot.StepFunctionsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.ExecutionNamePrefix; v != nil {
		tfMap["execution_name_prefix"] = aws.StringValue(v)
	}

	if v := apiObject.StateMachineName; v != nil {
		tfMap["state_machine_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenTimestreamActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Timestream; v != nil {
			results = append(results, flattenTimestreamAction(v)...)
		}
	}

	return results
}

func flattenTimestreamAction(apiObject *iot.TimestreamAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	if v := apiObject.Dimensions; v != nil {
		tfMap["dimension"] = flattenTimestreamDimensions(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.TableName; v != nil {
		tfMap["table_name"] = aws.StringValue(v)
	}

	if v := apiObject.Timestamp; v != nil {
		tfMap["timestamp"] = flattenTimestreamTimestamp(v)
	}

	return []interface{}{tfMap}
}

func flattenTimestreamDimensions(apiObjects []*iot.TimestreamDimension) *schema.Set {
	if apiObjects == nil {
		return nil
	}

	tfSet := schema.NewSet(schema.HashResource(timestreamDimensionResource), []interface{}{})

	for _, apiObject := range apiObjects {
		if apiObject != nil {
			tfMap := make(map[string]interface{})

			if v := apiObject.Name; v != nil {
				tfMap["name"] = aws.StringValue(v)
			}

			if v := apiObject.Value; v != nil {
				tfMap["value"] = aws.StringValue(v)
			}

			tfSet.Add(tfMap)
		}
	}

	return tfSet
}

func flattenTimestreamTimestamp(apiObject *iot.TimestreamTimestamp) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Unit; v != nil {
		tfMap["unit"] = aws.StringValue(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap["value"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func flattenErrorAction(errorAction *iot.Action) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)

	if errorAction == nil {
		return results
	}
	input := []*iot.Action{errorAction}
	if errorAction.CloudwatchAlarm != nil {
		results = append(results, map[string]interface{}{"cloudwatch_alarm": flattenCloudWatchAlarmActions(input)})
		return results
	}
	if errorAction.CloudwatchLogs != nil {
		results = append(results, map[string]interface{}{"cloudwatch_logs": flattenCloudWatchLogsActions(input)})
		return results
	}
	if errorAction.CloudwatchMetric != nil {
		results = append(results, map[string]interface{}{"cloudwatch_metric": flattenCloudWatchMetricActions(input)})
		return results
	}
	if errorAction.DynamoDB != nil {
		results = append(results, map[string]interface{}{"dynamodb": flattenDynamoDBActions(input)})
		return results
	}
	if errorAction.DynamoDBv2 != nil {
		results = append(results, map[string]interface{}{"dynamodbv2": flattenDynamoDBv2Actions(input)})
		return results
	}
	if errorAction.Elasticsearch != nil {
		results = append(results, map[string]interface{}{"elasticsearch": flattenElasticsearchActions(input)})
		return results
	}
	if errorAction.Firehose != nil {
		results = append(results, map[string]interface{}{"firehose": flattenFirehoseActions(input)})
		return results
	}
	if errorAction.Http != nil {
		results = append(results, map[string]interface{}{"http": flattenHTTPActions(input)})
		return results
	}
	if errorAction.IotAnalytics != nil {
		results = append(results, map[string]interface{}{"iot_analytics": flattenAnalyticsActions(input)})
		return results
	}
	if errorAction.IotEvents != nil {
		results = append(results, map[string]interface{}{"iot_events": flattenEventsActions(input)})
		return results
	}
	if errorAction.Kafka != nil {
		results = append(results, map[string]interface{}{"kafka": flattenKafkaActions(input)})
		return results
	}
	if errorAction.Kinesis != nil {
		results = append(results, map[string]interface{}{"kinesis": flattenKinesisActions(input)})
		return results
	}
	if errorAction.Lambda != nil {
		results = append(results, map[string]interface{}{"lambda": flattenLambdaActions(input)})
		return results
	}
	if errorAction.Republish != nil {
		results = append(results, map[string]interface{}{"republish": flattenRepublishActions(input)})
		return results
	}
	if errorAction.S3 != nil {
		results = append(results, map[string]interface{}{"s3": flattenS3Actions(input)})
		return results
	}
	if errorAction.Sns != nil {
		results = append(results, map[string]interface{}{"sns": flattenSNSActions(input)})
		return results
	}
	if errorAction.Sqs != nil {
		results = append(results, map[string]interface{}{"sqs": flattenSQSActions(input)})
		return results
	}
	if errorAction.StepFunctions != nil {
		results = append(results, map[string]interface{}{"step_functions": flattenStepFunctionsActions(input)})
		return results
	}
	if errorAction.Timestream != nil {
		results = append(results, map[string]interface{}{"timestream": flattenTimestreamActions(input)})
		return results
	}

	return results
}
