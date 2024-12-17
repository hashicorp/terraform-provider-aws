// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pipes

import (
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func sourceParametersSchema() *schema.Schema {
	verifySecretsManagerARN := validation.StringMatch(regexache.MustCompile(`^(^arn:aws([a-z]|\-)*:secretsmanager:([a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\d{1}):(\d{12}):secret:.+)$`), "")

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"activemq_broker_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.dynamodb_stream_parameters",
						"source_parameters.0.kinesis_stream_parameters",
						"source_parameters.0.managed_streaming_kafka_parameters",
						"source_parameters.0.rabbitmq_broker_parameters",
						"source_parameters.0.self_managed_kafka_parameters",
						"source_parameters.0.sqs_queue_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
							},
							"credentials": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"basic_auth": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verifySecretsManagerARN,
										},
									},
								},
							},
							"maximum_batching_window_in_seconds": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(0, 300),
							},
							"queue_name": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 1000),
									validation.StringMatch(regexache.MustCompile(`^[\s\S]*$`), ""),
								),
							},
						},
					},
				},
				"dynamodb_stream_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.activemq_broker_parameters",
						"source_parameters.0.kinesis_stream_parameters",
						"source_parameters.0.managed_streaming_kafka_parameters",
						"source_parameters.0.rabbitmq_broker_parameters",
						"source_parameters.0.self_managed_kafka_parameters",
						"source_parameters.0.sqs_queue_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
							},
							"dead_letter_config": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrARN: {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
							"maximum_batching_window_in_seconds": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(0, 300),
							},
							"maximum_record_age_in_seconds": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
								ValidateFunc: validation.Any(
									validation.IntInSlice([]int{-1}),
									validation.IntBetween(60, 604_800),
								),
							},
							"maximum_retry_attempts": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(-1, 10_000),
							},
							"on_partial_batch_item_failure": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.OnPartialBatchItemFailureStreams](),
							},
							"parallelization_factor": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10),
							},
							"starting_position": {
								Type:             schema.TypeString,
								Required:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[types.DynamoDBStreamStartPosition](),
							},
						},
					},
				},
				"filter_criteria": {
					Type:             schema.TypeList,
					Optional:         true,
					MaxItems:         1,
					DiffSuppressFunc: suppressEmptyConfigurationBlock("source_parameters.0.filter_criteria"),
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrFilter: {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 5,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"pattern": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 4096),
										},
									},
								},
							},
						},
					},
				},
				"kinesis_stream_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.activemq_broker_parameters",
						"source_parameters.0.dynamodb_stream_parameters",
						"source_parameters.0.managed_streaming_kafka_parameters",
						"source_parameters.0.rabbitmq_broker_parameters",
						"source_parameters.0.self_managed_kafka_parameters",
						"source_parameters.0.sqs_queue_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
							},
							"dead_letter_config": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrARN: {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
							"maximum_batching_window_in_seconds": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(0, 300),
							},
							"maximum_record_age_in_seconds": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
								ValidateFunc: validation.Any(
									validation.IntInSlice([]int{-1}),
									validation.IntBetween(60, 604_800),
								),
							},
							"maximum_retry_attempts": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(-1, 10_000),
							},
							"on_partial_batch_item_failure": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.OnPartialBatchItemFailureStreams](),
							},
							"parallelization_factor": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10),
							},
							"starting_position": {
								Type:             schema.TypeString,
								Required:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[types.KinesisStreamStartPosition](),
							},
							"starting_position_timestamp": {
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								ValidateFunc: validation.IsRFC3339Time,
							},
						},
					},
				},
				"managed_streaming_kafka_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.activemq_broker_parameters",
						"source_parameters.0.dynamodb_stream_parameters",
						"source_parameters.0.kinesis_stream_parameters",
						"source_parameters.0.rabbitmq_broker_parameters",
						"source_parameters.0.self_managed_kafka_parameters",
						"source_parameters.0.sqs_queue_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
							},
							"consumer_group_id": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 200),
									validation.StringMatch(regexache.MustCompile(`^[^.]([0-9A-Za-z_.-]+)$`), ""),
								),
							},
							"credentials": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"client_certificate_tls_auth": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verifySecretsManagerARN,
										},
										"sasl_scram_512_auth": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verifySecretsManagerARN,
										},
									},
								},
							},
							"maximum_batching_window_in_seconds": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(0, 300),
							},
							"starting_position": {
								Type:             schema.TypeString,
								Optional:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[types.MSKStartPosition](),
							},
							"topic_name": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 249),
									validation.StringMatch(regexache.MustCompile(`^[^.]([0-9A-Za-z_.-]+)$`), ""),
								),
							},
						},
					},
				},
				"rabbitmq_broker_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.activemq_broker_parameters",
						"source_parameters.0.dynamodb_stream_parameters",
						"source_parameters.0.kinesis_stream_parameters",
						"source_parameters.0.managed_streaming_kafka_parameters",
						"source_parameters.0.self_managed_kafka_parameters",
						"source_parameters.0.sqs_queue_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
							},
							"credentials": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"basic_auth": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verifySecretsManagerARN,
										},
									},
								},
							},
							"maximum_batching_window_in_seconds": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(0, 300),
							},
							"queue_name": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 1000),
									validation.StringMatch(regexache.MustCompile(`^[\s\S]*$`), ""),
								),
							},
							"virtual_host": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 200),
									validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_\/*:+=.@-]*$`), ""),
								),
							},
						},
					},
				},
				"self_managed_kafka_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.activemq_broker_parameters",
						"source_parameters.0.dynamodb_stream_parameters",
						"source_parameters.0.kinesis_stream_parameters",
						"source_parameters.0.managed_streaming_kafka_parameters",
						"source_parameters.0.rabbitmq_broker_parameters",
						"source_parameters.0.sqs_queue_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"additional_bootstrap_servers": {
								Type:     schema.TypeSet,
								Optional: true,
								ForceNew: true,
								MaxItems: 2,
								Elem: &schema.Schema{
									Type: schema.TypeString,
									ValidateFunc: validation.All(
										validation.StringLenBetween(1, 300),
										validation.StringMatch(regexache.MustCompile(`^(([0-9A-Za-z]|[0-9A-Za-z][0-9A-Za-z-]*[0-9A-Za-z])\.)*([0-9A-Za-z]|[0-9A-Za-z][0-9A-Za-z-]*[0-9A-Za-z]):[0-9]{1,5}$`), ""),
									),
								},
							},
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
							},
							"consumer_group_id": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 200),
									validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_\/*:+=.@-]*$`), ""),
								),
							},
							"credentials": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"basic_auth": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verifySecretsManagerARN,
										},
										"client_certificate_tls_auth": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verifySecretsManagerARN,
										},
										"sasl_scram_256_auth": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verifySecretsManagerARN,
										},
										"sasl_scram_512_auth": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verifySecretsManagerARN,
										},
									},
								},
							},
							"maximum_batching_window_in_seconds": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(0, 300),
							},
							"server_root_ca_certificate": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							"starting_position": {
								Type:             schema.TypeString,
								Optional:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[types.SelfManagedKafkaStartPosition](),
							},
							"topic_name": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 249),
									validation.StringMatch(regexache.MustCompile(`^[^.]([0-9A-Za-z_.-]+)$`), ""),
								),
							},
							"vpc": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrSecurityGroups: {
											Type:     schema.TypeSet,
											Optional: true,
											MaxItems: 5,
											Elem: &schema.Schema{
												Type: schema.TypeString,
												ValidateFunc: validation.All(
													validation.StringLenBetween(1, 1024),
													validation.StringMatch(regexache.MustCompile(`^sg-[0-9A-Za-z]*$`), ""),
												),
											},
										},
										names.AttrSubnets: {
											Type:     schema.TypeSet,
											Optional: true,
											MaxItems: 16,
											Elem: &schema.Schema{
												Type: schema.TypeString,
												ValidateFunc: validation.All(
													validation.StringLenBetween(1, 1024),
													validation.StringMatch(regexache.MustCompile(`^subnet-[0-9a-z]*$`), ""),
												),
											},
										},
									},
								},
							},
						},
					},
				},
				"sqs_queue_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.activemq_broker_parameters",
						"source_parameters.0.dynamodb_stream_parameters",
						"source_parameters.0.kinesis_stream_parameters",
						"source_parameters.0.managed_streaming_kafka_parameters",
						"source_parameters.0.rabbitmq_broker_parameters",
						"source_parameters.0.self_managed_kafka_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
							},
							"maximum_batching_window_in_seconds": {
								Type:         schema.TypeInt,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IntBetween(0, 300),
							},
						},
					},
				},
			},
		},
	}
}

func expandPipeSourceParameters(tfMap map[string]interface{}) *types.PipeSourceParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeSourceParameters{}

	if v, ok := tfMap["activemq_broker_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ActiveMQBrokerParameters = expandPipeSourceActiveMQBrokerParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["dynamodb_stream_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DynamoDBStreamParameters = expandPipeSourceDynamoDBStreamParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["filter_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.FilterCriteria = expandFilterCriteria(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["kinesis_stream_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.KinesisStreamParameters = expandPipeSourceKinesisStreamParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["managed_streaming_kafka_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ManagedStreamingKafkaParameters = expandPipeSourceManagedStreamingKafkaParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["rabbitmq_broker_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.RabbitMQBrokerParameters = expandPipeSourceRabbitMQBrokerParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["self_managed_kafka_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SelfManagedKafkaParameters = expandPipeSourceSelfManagedKafkaParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sqs_queue_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SqsQueueParameters = expandPipeSourceSQSQueueParameters(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandUpdatePipeSourceParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceParameters{}

	if v, ok := tfMap["activemq_broker_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ActiveMQBrokerParameters = expandUpdatePipeSourceActiveMQBrokerParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["dynamodb_stream_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DynamoDBStreamParameters = expandUpdatePipeSourceDynamoDBStreamParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["filter_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.FilterCriteria = expandFilterCriteria(v[0].(map[string]interface{}))
	} else {
		apiObject.FilterCriteria = &types.FilterCriteria{}
	}

	if v, ok := tfMap["kinesis_stream_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.KinesisStreamParameters = expandUpdatePipeSourceKinesisStreamParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["managed_streaming_kafka_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ManagedStreamingKafkaParameters = expandUpdatePipeSourceManagedStreamingKafkaParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["rabbitmq_broker_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.RabbitMQBrokerParameters = expandUpdatePipeSourceRabbitMQBrokerParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["self_managed_kafka_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SelfManagedKafkaParameters = expandUpdatePipeSourceSelfManagedKafkaParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sqs_queue_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SqsQueueParameters = expandUpdatePipeSourceSQSQueueParameters(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFilterCriteria(tfMap map[string]interface{}) *types.FilterCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.FilterCriteria{}

	if v, ok := tfMap[names.AttrFilter].([]interface{}); ok && len(v) > 0 {
		apiObject.Filters = expandFilters(v)
	}

	return apiObject
}

func expandFilter(tfMap map[string]interface{}) *types.Filter {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Filter{}

	if v, ok := tfMap["pattern"].(string); ok && v != "" {
		apiObject.Pattern = aws.String(v)
	}

	return apiObject
}

func expandFilters(tfList []interface{}) []types.Filter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Filter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandFilter(tfMap)

		if apiObject == nil || apiObject.Pattern == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPipeSourceActiveMQBrokerParameters(tfMap map[string]interface{}) *types.PipeSourceActiveMQBrokerParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeSourceActiveMQBrokerParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Credentials = expandMQBrokerAccessCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["queue_name"].(string); ok && v != "" {
		apiObject.QueueName = aws.String(v)
	}

	return apiObject
}

func expandUpdatePipeSourceActiveMQBrokerParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceActiveMQBrokerParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceActiveMQBrokerParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Credentials = expandMQBrokerAccessCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandMQBrokerAccessCredentials(tfMap map[string]interface{}) types.MQBrokerAccessCredentials {
	if tfMap == nil {
		return nil
	}

	if v, ok := tfMap["basic_auth"].(string); ok && v != "" {
		apiObject := &types.MQBrokerAccessCredentialsMemberBasicAuth{
			Value: v,
		}

		return apiObject
	}

	return nil
}

func expandPipeSourceDynamoDBStreamParameters(tfMap map[string]interface{}) *types.PipeSourceDynamoDBStreamParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeSourceDynamoDBStreamParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["dead_letter_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DeadLetterConfig = expandDeadLetterConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_record_age_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumRecordAgeInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_retry_attempts"].(int); ok && v != 0 {
		apiObject.MaximumRetryAttempts = aws.Int32(int32(v))
	}

	if v, ok := tfMap["on_partial_batch_item_failure"].(string); ok && v != "" {
		apiObject.OnPartialBatchItemFailure = types.OnPartialBatchItemFailureStreams(v)
	}

	if v, ok := tfMap["parallelization_factor"].(int); ok && v != 0 {
		apiObject.ParallelizationFactor = aws.Int32(int32(v))
	}

	if v, ok := tfMap["starting_position"].(string); ok && v != "" {
		apiObject.StartingPosition = types.DynamoDBStreamStartPosition(v)
	}

	return apiObject
}

func expandUpdatePipeSourceDynamoDBStreamParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceDynamoDBStreamParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceDynamoDBStreamParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["dead_letter_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DeadLetterConfig = expandDeadLetterConfig(v[0].(map[string]interface{}))
	} else {
		apiObject.DeadLetterConfig = &types.DeadLetterConfig{}
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_record_age_in_seconds"].(int); ok {
		apiObject.MaximumRecordAgeInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_retry_attempts"].(int); ok {
		apiObject.MaximumRetryAttempts = aws.Int32(int32(v))
	}

	if v, ok := tfMap["on_partial_batch_item_failure"].(string); ok {
		apiObject.OnPartialBatchItemFailure = types.OnPartialBatchItemFailureStreams(v)
	}

	if v, ok := tfMap["parallelization_factor"].(int); ok && v != 0 {
		apiObject.ParallelizationFactor = aws.Int32(int32(v))
	}

	return apiObject
}

func expandDeadLetterConfig(tfMap map[string]interface{}) *types.DeadLetterConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DeadLetterConfig{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	return apiObject
}

func expandPipeSourceKinesisStreamParameters(tfMap map[string]interface{}) *types.PipeSourceKinesisStreamParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeSourceKinesisStreamParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["dead_letter_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DeadLetterConfig = expandDeadLetterConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_record_age_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumRecordAgeInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_retry_attempts"].(int); ok && v != 0 {
		apiObject.MaximumRetryAttempts = aws.Int32(int32(v))
	}

	if v, ok := tfMap["on_partial_batch_item_failure"].(string); ok && v != "" {
		apiObject.OnPartialBatchItemFailure = types.OnPartialBatchItemFailureStreams(v)
	}

	if v, ok := tfMap["parallelization_factor"].(int); ok && v != 0 {
		apiObject.ParallelizationFactor = aws.Int32(int32(v))
	}

	if v, ok := tfMap["starting_position"].(string); ok && v != "" {
		apiObject.StartingPosition = types.KinesisStreamStartPosition(v)
	}

	if v, ok := tfMap["starting_position_timestamp"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.StartingPositionTimestamp = aws.Time(v)
	}

	return apiObject
}

func expandUpdatePipeSourceKinesisStreamParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceKinesisStreamParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceKinesisStreamParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["dead_letter_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DeadLetterConfig = expandDeadLetterConfig(v[0].(map[string]interface{}))
	} else {
		apiObject.DeadLetterConfig = &types.DeadLetterConfig{}
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_record_age_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumRecordAgeInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_retry_attempts"].(int); ok {
		apiObject.MaximumRetryAttempts = aws.Int32(int32(v))
	}

	if v, ok := tfMap["on_partial_batch_item_failure"].(string); ok {
		apiObject.OnPartialBatchItemFailure = types.OnPartialBatchItemFailureStreams(v)
	}

	if v, ok := tfMap["parallelization_factor"].(int); ok && v != 0 {
		apiObject.ParallelizationFactor = aws.Int32(int32(v))
	}

	return apiObject
}

func expandPipeSourceManagedStreamingKafkaParameters(tfMap map[string]interface{}) *types.PipeSourceManagedStreamingKafkaParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeSourceManagedStreamingKafkaParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["consumer_group_id"].(string); ok && v != "" {
		apiObject.ConsumerGroupID = aws.String(v)
	}

	if v, ok := tfMap["credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Credentials = expandMSKAccessCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["starting_position"].(string); ok && v != "" {
		apiObject.StartingPosition = types.MSKStartPosition(v)
	}

	if v, ok := tfMap["topic_name"].(string); ok && v != "" {
		apiObject.TopicName = aws.String(v)
	}

	return apiObject
}

func expandUpdatePipeSourceManagedStreamingKafkaParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceManagedStreamingKafkaParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceManagedStreamingKafkaParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Credentials = expandMSKAccessCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandMSKAccessCredentials(tfMap map[string]interface{}) types.MSKAccessCredentials {
	if tfMap == nil {
		return nil
	}

	if v, ok := tfMap["client_certificate_tls_auth"].(string); ok && v != "" {
		apiObject := &types.MSKAccessCredentialsMemberClientCertificateTlsAuth{
			Value: v,
		}

		return apiObject
	}

	if v, ok := tfMap["sasl_scram_512_auth"].(string); ok && v != "" {
		apiObject := &types.MSKAccessCredentialsMemberSaslScram512Auth{
			Value: v,
		}

		return apiObject
	}

	return nil
}

func expandPipeSourceRabbitMQBrokerParameters(tfMap map[string]interface{}) *types.PipeSourceRabbitMQBrokerParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeSourceRabbitMQBrokerParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Credentials = expandMQBrokerAccessCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["queue_name"].(string); ok && v != "" {
		apiObject.QueueName = aws.String(v)
	}

	if v, ok := tfMap["virtual_host"].(string); ok && v != "" {
		apiObject.VirtualHost = aws.String(v)
	}

	return apiObject
}

func expandUpdatePipeSourceRabbitMQBrokerParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceRabbitMQBrokerParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceRabbitMQBrokerParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Credentials = expandMQBrokerAccessCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandPipeSourceSelfManagedKafkaParameters(tfMap map[string]interface{}) *types.PipeSourceSelfManagedKafkaParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeSourceSelfManagedKafkaParameters{}

	if v, ok := tfMap["additional_bootstrap_servers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AdditionalBootstrapServers = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["consumer_group_id"].(string); ok && v != "" {
		apiObject.ConsumerGroupID = aws.String(v)
	}

	if v, ok := tfMap["credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Credentials = expandSelfManagedKafkaAccessConfigurationCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["server_root_ca_certificate"].(string); ok && v != "" {
		apiObject.ServerRootCaCertificate = aws.String(v)
	}

	if v, ok := tfMap["starting_position"].(string); ok && v != "" {
		apiObject.StartingPosition = types.SelfManagedKafkaStartPosition(v)
	}

	if v, ok := tfMap["topic_name"].(string); ok && v != "" {
		apiObject.TopicName = aws.String(v)
	}

	if v, ok := tfMap["vpc"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Vpc = expandSelfManagedKafkaAccessConfigurationVPC(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandUpdatePipeSourceSelfManagedKafkaParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceSelfManagedKafkaParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceSelfManagedKafkaParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["credentials"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Credentials = expandSelfManagedKafkaAccessConfigurationCredentials(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["server_root_ca_certificate"].(string); ok {
		apiObject.ServerRootCaCertificate = aws.String(v)
	}

	if v, ok := tfMap["vpc"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Vpc = expandSelfManagedKafkaAccessConfigurationVPC(v[0].(map[string]interface{}))
	} else {
		apiObject.Vpc = &types.SelfManagedKafkaAccessConfigurationVpc{}
	}

	return apiObject
}

func expandSelfManagedKafkaAccessConfigurationCredentials(tfMap map[string]interface{}) types.SelfManagedKafkaAccessConfigurationCredentials {
	if tfMap == nil {
		return nil
	}

	if v, ok := tfMap["basic_auth"].(string); ok && v != "" {
		apiObject := &types.SelfManagedKafkaAccessConfigurationCredentialsMemberBasicAuth{
			Value: v,
		}

		return apiObject
	}

	if v, ok := tfMap["client_certificate_tls_auth"].(string); ok && v != "" {
		apiObject := &types.SelfManagedKafkaAccessConfigurationCredentialsMemberClientCertificateTlsAuth{
			Value: v,
		}

		return apiObject
	}

	if v, ok := tfMap["sasl_scram_256_auth"].(string); ok && v != "" {
		apiObject := &types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram256Auth{
			Value: v,
		}

		return apiObject
	}

	if v, ok := tfMap["sasl_scram_512_auth"].(string); ok && v != "" {
		apiObject := &types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram512Auth{
			Value: v,
		}

		return apiObject
	}

	return nil
}

func expandSelfManagedKafkaAccessConfigurationVPC(tfMap map[string]interface{}) *types.SelfManagedKafkaAccessConfigurationVpc {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SelfManagedKafkaAccessConfigurationVpc{}

	if v, ok := tfMap[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroup = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnets].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandPipeSourceSQSQueueParameters(tfMap map[string]interface{}) *types.PipeSourceSqsQueueParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeSourceSqsQueueParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok && v != 0 {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandUpdatePipeSourceSQSQueueParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceSqsQueueParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceSqsQueueParameters{}

	if v, ok := tfMap["batch_size"].(int); ok && v != 0 {
		apiObject.BatchSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_batching_window_in_seconds"].(int); ok {
		apiObject.MaximumBatchingWindowInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenPipeSourceParameters(apiObject *types.PipeSourceParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ActiveMQBrokerParameters; v != nil {
		tfMap["activemq_broker_parameters"] = []interface{}{flattenPipeSourceActiveMQBrokerParameters(v)}
	}

	if v := apiObject.DynamoDBStreamParameters; v != nil {
		tfMap["dynamodb_stream_parameters"] = []interface{}{flattenPipeSourceDynamoDBStreamParameters(v)}
	}

	if v := apiObject.FilterCriteria; v != nil {
		tfMap["filter_criteria"] = []interface{}{flattenFilterCriteria(v)}
	}

	if v := apiObject.KinesisStreamParameters; v != nil {
		tfMap["kinesis_stream_parameters"] = []interface{}{flattenPipeSourceKinesisStreamParameters(v)}
	}

	if v := apiObject.ManagedStreamingKafkaParameters; v != nil {
		tfMap["managed_streaming_kafka_parameters"] = []interface{}{flattenPipeSourceManagedStreamingKafkaParameters(v)}
	}

	if v := apiObject.RabbitMQBrokerParameters; v != nil {
		tfMap["rabbitmq_broker_parameters"] = []interface{}{flattenPipeSourceRabbitMQBrokerParameters(v)}
	}

	if v := apiObject.SelfManagedKafkaParameters; v != nil {
		tfMap["self_managed_kafka_parameters"] = []interface{}{flattenPipeSourceSelfManagedKafkaParameters(v)}
	}

	if v := apiObject.SqsQueueParameters; v != nil {
		tfMap["sqs_queue_parameters"] = []interface{}{flattenPipeSourceSQSQueueParameters(v)}
	}

	return tfMap
}

func flattenFilterCriteria(apiObject *types.FilterCriteria) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Filters; v != nil {
		tfMap[names.AttrFilter] = flattenFilters(v)
	}

	return tfMap
}

func flattenFilter(apiObject types.Filter) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Pattern; v != nil {
		tfMap["pattern"] = aws.ToString(v)
	}

	return tfMap
}

func flattenFilters(apiObjects []types.Filter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenFilter(apiObject))
	}

	return tfList
}

func flattenPipeSourceActiveMQBrokerParameters(apiObject *types.PipeSourceActiveMQBrokerParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BatchSize; v != nil {
		tfMap["batch_size"] = aws.ToInt32(v)
	}

	if v := apiObject.Credentials; v != nil {
		tfMap["credentials"] = []interface{}{flattenMQBrokerAccessCredentials(v)}
	}

	if v := apiObject.MaximumBatchingWindowInSeconds; v != nil {
		tfMap["maximum_batching_window_in_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.QueueName; v != nil {
		tfMap["queue_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenMQBrokerAccessCredentials(apiObject types.MQBrokerAccessCredentials) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject, ok := apiObject.(*types.MQBrokerAccessCredentialsMemberBasicAuth); ok {
		if v := apiObject.Value; v != "" {
			tfMap["basic_auth"] = v
		}
	}

	return tfMap
}

func flattenPipeSourceDynamoDBStreamParameters(apiObject *types.PipeSourceDynamoDBStreamParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BatchSize; v != nil {
		tfMap["batch_size"] = aws.ToInt32(v)
	}

	if v := apiObject.DeadLetterConfig; v != nil {
		tfMap["dead_letter_config"] = []interface{}{flattenDeadLetterConfig(v)}
	}

	if v := apiObject.MaximumBatchingWindowInSeconds; v != nil {
		tfMap["maximum_batching_window_in_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.MaximumRecordAgeInSeconds; v != nil {
		tfMap["maximum_record_age_in_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.MaximumRetryAttempts; v != nil {
		tfMap["maximum_retry_attempts"] = aws.ToInt32(v)
	}

	if v := apiObject.OnPartialBatchItemFailure; v != "" {
		tfMap["on_partial_batch_item_failure"] = v
	}

	if v := apiObject.ParallelizationFactor; v != nil {
		tfMap["parallelization_factor"] = aws.ToInt32(v)
	}

	if v := apiObject.StartingPosition; v != "" {
		tfMap["starting_position"] = v
	}

	return tfMap
}

func flattenPipeSourceKinesisStreamParameters(apiObject *types.PipeSourceKinesisStreamParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BatchSize; v != nil {
		tfMap["batch_size"] = aws.ToInt32(v)
	}

	if v := apiObject.DeadLetterConfig; v != nil {
		tfMap["dead_letter_config"] = []interface{}{flattenDeadLetterConfig(v)}
	}

	if v := apiObject.MaximumBatchingWindowInSeconds; v != nil {
		tfMap["maximum_batching_window_in_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.MaximumRecordAgeInSeconds; v != nil {
		tfMap["maximum_record_age_in_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.MaximumRetryAttempts; v != nil {
		tfMap["maximum_retry_attempts"] = aws.ToInt32(v)
	}

	if v := apiObject.OnPartialBatchItemFailure; v != "" {
		tfMap["on_partial_batch_item_failure"] = v
	}

	if v := apiObject.ParallelizationFactor; v != nil {
		tfMap["parallelization_factor"] = aws.ToInt32(v)
	}

	if v := apiObject.StartingPosition; v != "" {
		tfMap["starting_position"] = v
	}

	if v := apiObject.StartingPositionTimestamp; v != nil {
		tfMap["starting_position_timestamp"] = aws.ToTime(v).Format(time.RFC3339)
	}

	return tfMap
}

func flattenPipeSourceManagedStreamingKafkaParameters(apiObject *types.PipeSourceManagedStreamingKafkaParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BatchSize; v != nil {
		tfMap["batch_size"] = aws.ToInt32(v)
	}

	if v := apiObject.ConsumerGroupID; v != nil {
		tfMap["consumer_group_id"] = aws.ToString(v)
	}

	if v := apiObject.Credentials; v != nil {
		tfMap["credentials"] = []interface{}{flattenMSKAccessCredentials(v)}
	}

	if v := apiObject.MaximumBatchingWindowInSeconds; v != nil {
		tfMap["maximum_batching_window_in_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.StartingPosition; v != "" {
		tfMap["starting_position"] = v
	}

	if v := apiObject.TopicName; v != nil {
		tfMap["topic_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenMSKAccessCredentials(apiObject types.MSKAccessCredentials) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject, ok := apiObject.(*types.MSKAccessCredentialsMemberClientCertificateTlsAuth); ok {
		if v := apiObject.Value; v != "" {
			tfMap["client_certificate_tls_auth"] = v
		}
	}

	if apiObject, ok := apiObject.(*types.MSKAccessCredentialsMemberSaslScram512Auth); ok {
		if v := apiObject.Value; v != "" {
			tfMap["sasl_scram_512_auth"] = v
		}
	}

	return tfMap
}

func flattenPipeSourceRabbitMQBrokerParameters(apiObject *types.PipeSourceRabbitMQBrokerParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BatchSize; v != nil {
		tfMap["batch_size"] = aws.ToInt32(v)
	}

	if v := apiObject.Credentials; v != nil {
		tfMap["credentials"] = []interface{}{flattenMQBrokerAccessCredentials(v)}
	}

	if v := apiObject.MaximumBatchingWindowInSeconds; v != nil {
		tfMap["maximum_batching_window_in_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.QueueName; v != nil {
		tfMap["queue_name"] = aws.ToString(v)
	}

	if v := apiObject.VirtualHost; v != nil {
		tfMap["virtual_host"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPipeSourceSelfManagedKafkaParameters(apiObject *types.PipeSourceSelfManagedKafkaParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AdditionalBootstrapServers; v != nil {
		tfMap["additional_bootstrap_servers"] = v
	}

	if v := apiObject.BatchSize; v != nil {
		tfMap["batch_size"] = aws.ToInt32(v)
	}

	if v := apiObject.ConsumerGroupID; v != nil {
		tfMap["consumer_group_id"] = aws.ToString(v)
	}

	if v := apiObject.Credentials; v != nil {
		tfMap["credentials"] = []interface{}{flattenSelfManagedKafkaAccessConfigurationCredentials(v)}
	}

	if v := apiObject.MaximumBatchingWindowInSeconds; v != nil {
		tfMap["maximum_batching_window_in_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.ServerRootCaCertificate; v != nil {
		tfMap["server_root_ca_certificate"] = aws.ToString(v)
	}

	if v := apiObject.StartingPosition; v != "" {
		tfMap["starting_position"] = v
	}

	if v := apiObject.TopicName; v != nil {
		tfMap["topic_name"] = aws.ToString(v)
	}

	if v := apiObject.Vpc; v != nil {
		tfMap["vpc"] = []interface{}{flattenSelfManagedKafkaAccessConfigurationVPC(v)}
	}

	return tfMap
}

func flattenSelfManagedKafkaAccessConfigurationCredentials(apiObject types.SelfManagedKafkaAccessConfigurationCredentials) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject, ok := apiObject.(*types.SelfManagedKafkaAccessConfigurationCredentialsMemberBasicAuth); ok {
		if v := apiObject.Value; v != "" {
			tfMap["basic_auth"] = v
		}
	}

	if apiObject, ok := apiObject.(*types.SelfManagedKafkaAccessConfigurationCredentialsMemberClientCertificateTlsAuth); ok {
		if v := apiObject.Value; v != "" {
			tfMap["client_certificate_tls_auth"] = v
		}
	}

	if apiObject, ok := apiObject.(*types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram256Auth); ok {
		if v := apiObject.Value; v != "" {
			tfMap["sasl_scram_256_auth"] = v
		}
	}

	if apiObject, ok := apiObject.(*types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram512Auth); ok {
		if v := apiObject.Value; v != "" {
			tfMap["sasl_scram_512_auth"] = v
		}
	}

	return tfMap
}

func flattenSelfManagedKafkaAccessConfigurationVPC(apiObject *types.SelfManagedKafkaAccessConfigurationVpc) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroup; v != nil {
		tfMap[names.AttrSecurityGroups] = v
	}

	if v := apiObject.Subnets; v != nil {
		tfMap[names.AttrSubnets] = v
	}

	return tfMap
}

func flattenPipeSourceSQSQueueParameters(apiObject *types.PipeSourceSqsQueueParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BatchSize; v != nil {
		tfMap["batch_size"] = aws.ToInt32(v)
	}

	if v := apiObject.MaximumBatchingWindowInSeconds; v != nil {
		tfMap["maximum_batching_window_in_seconds"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenDeadLetterConfig(apiObject *types.DeadLetterConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	return tfMap
}
