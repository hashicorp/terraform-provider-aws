package pipes

import (
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func sourceParametersSchema() *schema.Schema {
	verifySecretsManagerARN := validation.StringMatch(regexp.MustCompile(`^(^arn:aws([a-z]|\-)*:secretsmanager:([a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\d{1}):(\d{12}):secret:.+)$`), "")

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"active_mq_broker": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.dynamo_db_stream",
						"source_parameters.0.kinesis_stream",
						"source_parameters.0.managed_streaming_kafka",
						"source_parameters.0.rabbit_mq_broker",
						"source_parameters.0.self_managed_kafka",
						"source_parameters.0.sqs_queue",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "100"
								},
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
								ValidateFunc: validation.IntBetween(0, 300),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "0"
								},
							},
							"queue": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 1000),
									validation.StringMatch(regexp.MustCompile(`^[\s\S]*$`), ""),
								),
							},
						},
					},
				},
				"dynamo_db_stream": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.active_mq_broker",
						"source_parameters.0.kinesis_stream",
						"source_parameters.0.managed_streaming_kafka",
						"source_parameters.0.rabbit_mq_broker",
						"source_parameters.0.self_managed_kafka",
						"source_parameters.0.sqs_queue",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "100"
								},
							},
							"dead_letter_config": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"arn": {
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
								ValidateFunc: validation.IntBetween(0, 300),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "0"
								},
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
								ValidateFunc: validation.IntBetween(1, 10),
								Default:      1,
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
					DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter": {
								Type:     schema.TypeList,
								Required: true,
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
				"kinesis_stream": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.active_mq_broker",
						"source_parameters.0.dynamo_db_stream",
						"source_parameters.0.managed_streaming_kafka",
						"source_parameters.0.rabbit_mq_broker",
						"source_parameters.0.self_managed_kafka",
						"source_parameters.0.sqs_queue",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "100"
								},
							},
							"dead_letter_config": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"arn": {
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
								ValidateFunc: validation.IntBetween(0, 300),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "0"
								},
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
								ValidateFunc: validation.IntBetween(1, 10),
								Default:      1,
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
				"managed_streaming_kafka": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.active_mq_broker",
						"source_parameters.0.dynamo_db_stream",
						"source_parameters.0.kinesis_stream",
						"source_parameters.0.rabbit_mq_broker",
						"source_parameters.0.self_managed_kafka",
						"source_parameters.0.sqs_queue",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "100"
								},
							},
							"consumer_group_id": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 200),
									validation.StringMatch(regexp.MustCompile(`^[^.]([a-zA-Z0-9\-_.]+)$`), ""),
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
								ValidateFunc: validation.IntBetween(0, 300),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "0"
								},
							},
							"starting_position": {
								Type:             schema.TypeString,
								Optional:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[types.MSKStartPosition](),
							},
							"topic": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 249),
									validation.StringMatch(regexp.MustCompile(`^[^.]([a-zA-Z0-9\-_.]+)$`), ""),
								),
							},
						},
					},
				},
				"rabbit_mq_broker": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.active_mq_broker",
						"source_parameters.0.dynamo_db_stream",
						"source_parameters.0.kinesis_stream",
						"source_parameters.0.managed_streaming_kafka",
						"source_parameters.0.self_managed_kafka",
						"source_parameters.0.sqs_queue",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "100"
								},
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
								ValidateFunc: validation.IntBetween(0, 300),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "0"
								},
							},
							"queue": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 1000),
									validation.StringMatch(regexp.MustCompile(`^[\s\S]*$`), ""),
								),
							},
							"virtual_host": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 200),
									validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-\/*:_+=.@-]*$`), ""),
								),
							},
						},
					},
				},
				"self_managed_kafka": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.active_mq_broker",
						"source_parameters.0.dynamo_db_stream",
						"source_parameters.0.kinesis_stream",
						"source_parameters.0.managed_streaming_kafka",
						"source_parameters.0.rabbit_mq_broker",
						"source_parameters.0.sqs_queue",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "100"
								},
							},
							"consumer_group_id": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 200),
									validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-\/*:_+=.@-]*$`), ""),
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
											Required:     true,
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
								ValidateFunc: validation.IntBetween(0, 300),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "0"
								},
							},
							"server_root_ca_certificate": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							"servers": {
								Type:     schema.TypeSet,
								Optional: true,
								ForceNew: true,
								MaxItems: 2,
								Elem: &schema.Schema{
									Type: schema.TypeString,
									ValidateFunc: validation.All(
										validation.StringLenBetween(1, 300),
										validation.StringMatch(regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9]):[0-9]{1,5}$`), ""),
									),
								},
							},
							"starting_position": {
								Type:             schema.TypeString,
								Optional:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[types.SelfManagedKafkaStartPosition](),
							},
							"topic": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 249),
									validation.StringMatch(regexp.MustCompile(`^[^.]([a-zA-Z0-9\-_.]+)$`), ""),
								),
							},
							"vpc": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"security_groups": {
											Type:     schema.TypeSet,
											Optional: true,
											MaxItems: 5,
											Elem: &schema.Schema{
												Type: schema.TypeString,
												ValidateFunc: validation.All(
													validation.StringLenBetween(1, 1024),
													validation.StringMatch(regexp.MustCompile(`^sg-[0-9a-zA-Z]*$`), ""),
												),
											},
										},
										"subnets": {
											Type:     schema.TypeSet,
											Optional: true,
											MaxItems: 16,
											Elem: &schema.Schema{
												Type: schema.TypeString,
												ValidateFunc: validation.All(
													validation.StringLenBetween(1, 1024),
													validation.StringMatch(regexp.MustCompile(`^subnet-[0-9a-z]*$`), ""),
												),
											},
										},
									},
								},
							},
						},
					},
				},
				"sqs_queue": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"source_parameters.0.active_mq_broker",
						"source_parameters.0.dynamo_db_stream",
						"source_parameters.0.kinesis_stream",
						"source_parameters.0.managed_streaming_kafka",
						"source_parameters.0.rabbit_mq_broker",
						"source_parameters.0.self_managed_kafka",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"batch_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(1, 10000),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "10"
								},
							},
							"maximum_batching_window_in_seconds": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(0, 300),
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									if new != "" && new != "0" {
										return false
									}
									return old == "0"
								},
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

	// ... nested attribute handling ...

	return apiObject
}

func expandUpdatePipeSourceParameters(tfMap map[string]interface{}) *types.UpdatePipeSourceParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.UpdatePipeSourceParameters{}

	// ... nested attribute handling ...

	return apiObject
}

func flattenPipeSourceParameters(apiObject *types.PipeSourceParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	// ... nested attribute handling ...

	return tfMap
}

func expandSourceParameters(config []interface{}) *types.PipeSourceParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeSourceParameters
	for _, c := range config {
		param, ok := c.(map[string]interface{})
		if !ok {
			return nil
		}

		if val, ok := param["active_mq_broker"]; ok {
			parameters.ActiveMQBrokerParameters = expandSourceActiveMQBrokerParameters(val.([]interface{}))
		}

		if val, ok := param["dynamo_db_stream"]; ok {
			parameters.DynamoDBStreamParameters = expandSourceDynamoDBStreamParameters(val.([]interface{}))
		}

		if val, ok := param["kinesis_stream"]; ok {
			parameters.KinesisStreamParameters = expandSourceKinesisStreamParameters(val.([]interface{}))
		}

		if val, ok := param["managed_streaming_kafka"]; ok {
			parameters.ManagedStreamingKafkaParameters = expandSourceManagedStreamingKafkaParameters(val.([]interface{}))
		}

		if val, ok := param["rabbit_mq_broker"]; ok {
			parameters.RabbitMQBrokerParameters = expandSourceRabbitMQBrokerParameters(val.([]interface{}))
		}

		if val, ok := param["self_managed_kafka"]; ok {
			parameters.SelfManagedKafkaParameters = expandSourceSelfManagedKafkaParameters(val.([]interface{}))
		}

		if val, ok := param["sqs_queue"]; ok {
			parameters.SqsQueueParameters = expandSourceSqsQueueParameters(val.([]interface{}))
		}

		if val, ok := param["filter_criteria"]; ok {
			parameters.FilterCriteria = expandSourceFilterCriteria(val.([]interface{}))
		}
	}
	return &parameters
}

func expandSourceActiveMQBrokerParameters(config []interface{}) *types.PipeSourceActiveMQBrokerParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeSourceActiveMQBrokerParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.QueueName = expandString("queue", param)
		if val, ok := param["credentials"]; ok {
			credentialsConfig := val.([]interface{})
			if len(credentialsConfig) != 0 {
				var credentialsParameters types.MQBrokerAccessCredentialsMemberBasicAuth
				for _, cc := range credentialsConfig {
					credentialsParam := cc.(map[string]interface{})
					credentialsParameters.Value = expandStringValue("basic_auth", credentialsParam)
				}
				parameters.Credentials = &credentialsParameters
			}
		}
	}
	return &parameters
}

func expandSourceDynamoDBStreamParameters(config []interface{}) *types.PipeSourceDynamoDBStreamParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeSourceDynamoDBStreamParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.MaximumRecordAgeInSeconds = expandInt32("maximum_record_age_in_seconds", param)
		parameters.ParallelizationFactor = expandInt32("parallelization_factor", param)
		parameters.MaximumRetryAttempts = expandInt32("maximum_retry_attempts", param)
		startingPosition := expandStringValue("starting_position", param)
		if startingPosition != "" {
			parameters.StartingPosition = types.DynamoDBStreamStartPosition(startingPosition)
		}
		onPartialBatchItemFailure := expandStringValue("on_partial_batch_item_failure", param)
		if onPartialBatchItemFailure != "" {
			parameters.OnPartialBatchItemFailure = types.OnPartialBatchItemFailureStreams(onPartialBatchItemFailure)
		}
		if val, ok := param["dead_letter_config"]; ok {
			parameters.DeadLetterConfig = expandSourceDeadLetterConfig(val.([]interface{}))
		}
	}
	return &parameters
}

func expandSourceKinesisStreamParameters(config []interface{}) *types.PipeSourceKinesisStreamParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeSourceKinesisStreamParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.MaximumRecordAgeInSeconds = expandInt32("maximum_record_age_in_seconds", param)
		parameters.ParallelizationFactor = expandInt32("parallelization_factor", param)
		parameters.MaximumRetryAttempts = expandInt32("maximum_retry_attempts", param)

		startingPosition := expandStringValue("starting_position", param)
		if startingPosition != "" {
			parameters.StartingPosition = types.KinesisStreamStartPosition(startingPosition)
		}
		onPartialBatchItemFailure := expandStringValue("on_partial_batch_item_failure", param)
		if onPartialBatchItemFailure != "" {
			parameters.OnPartialBatchItemFailure = types.OnPartialBatchItemFailureStreams(onPartialBatchItemFailure)
		}
		if val, ok := param["starting_position_timestamp"]; ok {
			t, _ := time.Parse(time.RFC3339, val.(string))

			parameters.StartingPositionTimestamp = aws.Time(t)
		}
		if val, ok := param["dead_letter_config"]; ok {
			parameters.DeadLetterConfig = expandSourceDeadLetterConfig(val.([]interface{}))
		}
	}
	return &parameters
}

func expandSourceManagedStreamingKafkaParameters(config []interface{}) *types.PipeSourceManagedStreamingKafkaParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeSourceManagedStreamingKafkaParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.TopicName = expandString("topic", param)
		parameters.ConsumerGroupID = expandString("consumer_group_id", param)

		startingPosition := expandStringValue("starting_position", param)
		if startingPosition != "" {
			parameters.StartingPosition = types.MSKStartPosition(startingPosition)
		}

		if val, ok := param["credentials"]; ok {
			credentialsConfig := val.([]interface{})
			if len(credentialsConfig) != 0 {
				for _, cc := range credentialsConfig {
					credentialsParam := cc.(map[string]interface{})
					if _, ok := credentialsParam["client_certificate_tls_auth"]; ok {
						var credentialsParameters types.MSKAccessCredentialsMemberClientCertificateTlsAuth
						credentialsParameters.Value = expandStringValue("client_certificate_tls_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
					if _, ok := credentialsParam["sasl_scram_512_auth"]; ok {
						var credentialsParameters types.MSKAccessCredentialsMemberSaslScram512Auth
						credentialsParameters.Value = expandStringValue("sasl_scram_512_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
				}
			}
		}
	}
	return &parameters
}

func expandSourceRabbitMQBrokerParameters(config []interface{}) *types.PipeSourceRabbitMQBrokerParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeSourceRabbitMQBrokerParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.QueueName = expandString("queue", param)
		parameters.VirtualHost = expandString("virtual_host", param)

		if val, ok := param["credentials"]; ok {
			credentialsConfig := val.([]interface{})
			if len(credentialsConfig) != 0 {
				var credentialsParameters types.MQBrokerAccessCredentialsMemberBasicAuth
				for _, cc := range credentialsConfig {
					credentialsParam := cc.(map[string]interface{})
					credentialsParameters.Value = expandStringValue("basic_auth", credentialsParam)
				}
				parameters.Credentials = &credentialsParameters
			}
		}
	}
	return &parameters
}

func expandSourceSelfManagedKafkaParameters(config []interface{}) *types.PipeSourceSelfManagedKafkaParameters {
	if len(config) == 0 {
		return nil
	}
	var parameters types.PipeSourceSelfManagedKafkaParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.TopicName = expandString("topic", param)
		parameters.ConsumerGroupID = expandString("consumer_group_id", param)
		parameters.ServerRootCaCertificate = expandString("server_root_ca_certificate", param)
		startingPosition := expandStringValue("starting_position", param)
		if startingPosition != "" {
			parameters.StartingPosition = types.SelfManagedKafkaStartPosition(startingPosition)
		}
		if value, ok := param["servers"]; ok && value.(*schema.Set).Len() > 0 {
			parameters.AdditionalBootstrapServers = flex.ExpandStringValueSet(value.(*schema.Set))
		}

		if val, ok := param["credentials"]; ok {
			credentialsConfig := val.([]interface{})
			if len(credentialsConfig) != 0 {
				for _, cc := range credentialsConfig {
					credentialsParam := cc.(map[string]interface{})
					if _, ok := credentialsParam["basic_auth"]; ok {
						var credentialsParameters types.SelfManagedKafkaAccessConfigurationCredentialsMemberBasicAuth
						credentialsParameters.Value = expandStringValue("basic_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
					if _, ok := credentialsParam["client_certificate_tls_auth"]; ok {
						var credentialsParameters types.SelfManagedKafkaAccessConfigurationCredentialsMemberClientCertificateTlsAuth
						credentialsParameters.Value = expandStringValue("client_certificate_tls_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
					if _, ok := credentialsParam["sasl_scram_512_auth"]; ok {
						var credentialsParameters types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram512Auth
						credentialsParameters.Value = expandStringValue("sasl_scram_512_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
					if _, ok := credentialsParam["sasl_scram_256_auth"]; ok {
						var credentialsParameters types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram256Auth
						credentialsParameters.Value = expandStringValue("sasl_scram_256_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
				}
			}
		}

		if val, ok := param["vpc"]; ok {
			vpcConfig := val.([]interface{})
			if len(vpcConfig) != 0 {
				var vpcParameters types.SelfManagedKafkaAccessConfigurationVpc
				for _, vc := range vpcConfig {
					vpcParam := vc.(map[string]interface{})
					if value, ok := vpcParam["security_groups"]; ok && value.(*schema.Set).Len() > 0 {
						vpcParameters.SecurityGroup = flex.ExpandStringValueSet(value.(*schema.Set))
					}
					if value, ok := vpcParam["subnets"]; ok && value.(*schema.Set).Len() > 0 {
						vpcParameters.Subnets = flex.ExpandStringValueSet(value.(*schema.Set))
					}
				}
				parameters.Vpc = &vpcParameters
			}
		}
	}

	return &parameters
}

func expandSourceSqsQueueParameters(config []interface{}) *types.PipeSourceSqsQueueParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeSourceSqsQueueParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
	}

	return &parameters
}

func expandSourceDeadLetterConfig(config []interface{}) *types.DeadLetterConfig {
	if len(config) == 0 {
		return nil
	}

	var parameters types.DeadLetterConfig
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.Arn = expandString("arn", param)
	}

	return &parameters
}

func expandSourceFilterCriteria(config []interface{}) *types.FilterCriteria {
	if len(config) == 0 {
		return nil
	}

	var parameters types.FilterCriteria
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["filter"]; ok {
			filtersConfig := val.([]interface{})
			var filters []types.Filter
			for _, f := range filtersConfig {
				filterParam := f.(map[string]interface{})
				pattern := expandString("pattern", filterParam)
				if pattern != nil {
					filters = append(filters, types.Filter{
						Pattern: pattern,
					})
				}
			}
			if len(filters) > 0 {
				parameters.Filters = filters
			}
		}
	}

	return &parameters
}

func flattenSourceParameters(sourceParameters *types.PipeSourceParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if sourceParameters.ActiveMQBrokerParameters != nil {
		config["active_mq_broker"] = flattenSourceActiveMQBrokerParameters(sourceParameters.ActiveMQBrokerParameters)
	}

	if sourceParameters.DynamoDBStreamParameters != nil {
		config["dynamo_db_stream"] = flattenSourceDynamoDBStreamParameters(sourceParameters.DynamoDBStreamParameters)
	}

	if sourceParameters.KinesisStreamParameters != nil {
		config["kinesis_stream"] = flattenSourceKinesisStreamParameters(sourceParameters.KinesisStreamParameters)
	}

	if sourceParameters.ManagedStreamingKafkaParameters != nil {
		config["managed_streaming_kafka"] = flattenSourceManagedStreamingKafkaParameters(sourceParameters.ManagedStreamingKafkaParameters)
	}

	if sourceParameters.RabbitMQBrokerParameters != nil {
		config["rabbit_mq_broker"] = flattenSourceRabbitMQBrokerParameters(sourceParameters.RabbitMQBrokerParameters)
	}

	if sourceParameters.SelfManagedKafkaParameters != nil {
		config["self_managed_kafka"] = flattenSourceSelfManagedKafkaParameters(sourceParameters.SelfManagedKafkaParameters)
	}

	if sourceParameters.SqsQueueParameters != nil {
		config["sqs_queue"] = flattenSourceSqsQueueParameters(sourceParameters.SqsQueueParameters)
	}

	if sourceParameters.FilterCriteria != nil {
		criteria := flattenSourceFilterCriteria(sourceParameters.FilterCriteria)
		if len(criteria) > 0 {
			config["filter_criteria"] = criteria
		}
	}

	if len(config) == 0 {
		return nil
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceActiveMQBrokerParameters(parameters *types.PipeSourceActiveMQBrokerParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.BatchSize != nil {
		config["batch_size"] = aws.ToInt32(parameters.BatchSize)
	}
	if parameters.MaximumBatchingWindowInSeconds != nil && aws.ToInt32(parameters.MaximumBatchingWindowInSeconds) != 0 {
		config["maximum_batching_window_in_seconds"] = aws.ToInt32(parameters.MaximumBatchingWindowInSeconds)
	}
	if parameters.QueueName != nil {
		config["queue"] = aws.ToString(parameters.QueueName)
	}
	if parameters.Credentials != nil {
		credentialsConfig := make(map[string]interface{})
		switch v := parameters.Credentials.(type) {
		case *types.MQBrokerAccessCredentialsMemberBasicAuth:
			credentialsConfig["basic_auth"] = v.Value
		}
		config["credentials"] = []map[string]interface{}{credentialsConfig}
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceDynamoDBStreamParameters(parameters *types.PipeSourceDynamoDBStreamParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.BatchSize != nil {
		config["batch_size"] = aws.ToInt32(parameters.BatchSize)
	}
	if parameters.MaximumBatchingWindowInSeconds != nil && aws.ToInt32(parameters.MaximumBatchingWindowInSeconds) != 0 {
		config["maximum_batching_window_in_seconds"] = aws.ToInt32(parameters.MaximumBatchingWindowInSeconds)
	}
	if parameters.MaximumRecordAgeInSeconds != nil {
		config["maximum_record_age_in_seconds"] = aws.ToInt32(parameters.MaximumRecordAgeInSeconds)
	}
	if parameters.ParallelizationFactor != nil {
		config["parallelization_factor"] = aws.ToInt32(parameters.ParallelizationFactor)
	}
	if parameters.MaximumRetryAttempts != nil {
		config["maximum_retry_attempts"] = aws.ToInt32(parameters.MaximumRetryAttempts)
	}
	if parameters.StartingPosition != "" {
		config["starting_position"] = parameters.StartingPosition
	}
	if parameters.OnPartialBatchItemFailure != "" {
		config["on_partial_batch_item_failure"] = parameters.OnPartialBatchItemFailure
	}
	if parameters.DeadLetterConfig != nil {
		config["dead_letter_config"] = flattenSourceDeadLetterConfig(parameters.DeadLetterConfig)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceKinesisStreamParameters(parameters *types.PipeSourceKinesisStreamParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.BatchSize != nil {
		config["batch_size"] = aws.ToInt32(parameters.BatchSize)
	}
	if parameters.MaximumBatchingWindowInSeconds != nil && aws.ToInt32(parameters.MaximumBatchingWindowInSeconds) != 0 {
		config["maximum_batching_window_in_seconds"] = aws.ToInt32(parameters.MaximumBatchingWindowInSeconds)
	}
	if parameters.MaximumRecordAgeInSeconds != nil {
		config["maximum_record_age_in_seconds"] = aws.ToInt32(parameters.MaximumRecordAgeInSeconds)
	}
	if parameters.ParallelizationFactor != nil {
		config["parallelization_factor"] = aws.ToInt32(parameters.ParallelizationFactor)
	}
	if parameters.MaximumRetryAttempts != nil {
		config["maximum_retry_attempts"] = aws.ToInt32(parameters.MaximumRetryAttempts)
	}
	if parameters.StartingPosition != "" {
		config["starting_position"] = parameters.StartingPosition
	}
	if parameters.OnPartialBatchItemFailure != "" {
		config["on_partial_batch_item_failure"] = parameters.OnPartialBatchItemFailure
	}
	if parameters.StartingPositionTimestamp != nil {
		config["starting_position_timestamp"] = aws.ToTime(parameters.StartingPositionTimestamp).Format(time.RFC3339)
	}
	if parameters.DeadLetterConfig != nil {
		config["dead_letter_config"] = flattenSourceDeadLetterConfig(parameters.DeadLetterConfig)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceManagedStreamingKafkaParameters(parameters *types.PipeSourceManagedStreamingKafkaParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.BatchSize != nil {
		config["batch_size"] = aws.ToInt32(parameters.BatchSize)
	}
	if parameters.MaximumBatchingWindowInSeconds != nil && aws.ToInt32(parameters.MaximumBatchingWindowInSeconds) != 0 {
		config["maximum_batching_window_in_seconds"] = aws.ToInt32(parameters.MaximumBatchingWindowInSeconds)
	}
	if parameters.ConsumerGroupID != nil {
		config["consumer_group_id"] = aws.ToString(parameters.ConsumerGroupID)
	}
	if parameters.StartingPosition != "" {
		config["starting_position"] = parameters.StartingPosition
	}
	if parameters.TopicName != nil {
		config["topic"] = aws.ToString(parameters.TopicName)
	}
	if parameters.Credentials != nil {
		credentialsConfig := make(map[string]interface{})
		switch v := parameters.Credentials.(type) {
		case *types.MSKAccessCredentialsMemberClientCertificateTlsAuth:
			credentialsConfig["client_certificate_tls_auth"] = v.Value
		case *types.MSKAccessCredentialsMemberSaslScram512Auth:
			credentialsConfig["sasl_scram_512_auth"] = v.Value
		}
		config["credentials"] = []map[string]interface{}{credentialsConfig}
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceRabbitMQBrokerParameters(parameters *types.PipeSourceRabbitMQBrokerParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.BatchSize != nil {
		config["batch_size"] = aws.ToInt32(parameters.BatchSize)
	}
	if parameters.MaximumBatchingWindowInSeconds != nil && aws.ToInt32(parameters.MaximumBatchingWindowInSeconds) != 0 {
		config["maximum_batching_window_in_seconds"] = aws.ToInt32(parameters.MaximumBatchingWindowInSeconds)
	}
	if parameters.QueueName != nil {
		config["queue"] = aws.ToString(parameters.QueueName)
	}
	if parameters.VirtualHost != nil {
		config["virtual_host"] = aws.ToString(parameters.VirtualHost)
	}
	if parameters.Credentials != nil {
		credentialsConfig := make(map[string]interface{})
		switch v := parameters.Credentials.(type) {
		case *types.MQBrokerAccessCredentialsMemberBasicAuth:
			credentialsConfig["basic_auth"] = v.Value
		}
		config["credentials"] = []map[string]interface{}{credentialsConfig}
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceSelfManagedKafkaParameters(parameters *types.PipeSourceSelfManagedKafkaParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.BatchSize != nil {
		config["batch_size"] = aws.ToInt32(parameters.BatchSize)
	}
	if parameters.MaximumBatchingWindowInSeconds != nil && aws.ToInt32(parameters.MaximumBatchingWindowInSeconds) != 0 {
		config["maximum_batching_window_in_seconds"] = aws.ToInt32(parameters.MaximumBatchingWindowInSeconds)
	}
	if parameters.ConsumerGroupID != nil {
		config["consumer_group_id"] = aws.ToString(parameters.ConsumerGroupID)
	}
	if parameters.StartingPosition != "" {
		config["starting_position"] = parameters.StartingPosition
	}
	if parameters.TopicName != nil {
		config["topic"] = aws.ToString(parameters.TopicName)
	}
	if parameters.AdditionalBootstrapServers != nil {
		config["servers"] = flex.FlattenStringValueSet(parameters.AdditionalBootstrapServers)
	}
	if parameters.ServerRootCaCertificate != nil {
		config["server_root_ca_certificate"] = aws.ToString(parameters.ServerRootCaCertificate)
	}

	if parameters.Credentials != nil {
		credentialsConfig := make(map[string]interface{})
		switch v := parameters.Credentials.(type) {
		case *types.SelfManagedKafkaAccessConfigurationCredentialsMemberBasicAuth:
			credentialsConfig["basic_auth"] = v.Value
		case *types.SelfManagedKafkaAccessConfigurationCredentialsMemberClientCertificateTlsAuth:
			credentialsConfig["client_certificate_tls_auth"] = v.Value
		case *types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram256Auth:
			credentialsConfig["sasl_scram_256_auth"] = v.Value
		case *types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram512Auth:
			credentialsConfig["sasl_scram_512_auth"] = v.Value
		}
		config["credentials"] = []map[string]interface{}{credentialsConfig}
	}
	if parameters.Vpc != nil {
		vpcConfig := make(map[string]interface{})
		vpcConfig["security_groups"] = flex.FlattenStringValueSet(parameters.Vpc.SecurityGroup)
		vpcConfig["subnets"] = flex.FlattenStringValueSet(parameters.Vpc.Subnets)
		config["vpc"] = []map[string]interface{}{vpcConfig}
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceSqsQueueParameters(parameters *types.PipeSourceSqsQueueParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.BatchSize != nil {
		config["batch_size"] = aws.ToInt32(parameters.BatchSize)
	}
	if parameters.MaximumBatchingWindowInSeconds != nil && aws.ToInt32(parameters.MaximumBatchingWindowInSeconds) != 0 {
		config["maximum_batching_window_in_seconds"] = aws.ToInt32(parameters.MaximumBatchingWindowInSeconds)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceFilterCriteria(parameters *types.FilterCriteria) []map[string]interface{} {
	config := make(map[string]interface{})

	if len(parameters.Filters) != 0 {
		var filters []map[string]interface{}
		for _, filter := range parameters.Filters {
			pattern := make(map[string]interface{})
			pattern["pattern"] = aws.ToString(filter.Pattern)
			filters = append(filters, pattern)
		}
		if len(filters) != 0 {
			config["filter"] = filters
		}
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenSourceDeadLetterConfig(parameters *types.DeadLetterConfig) []map[string]interface{} {
	if parameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if parameters.Arn != nil {
		config["arn"] = aws.ToString(parameters.Arn)
	}

	result := []map[string]interface{}{config}
	return result
}

func expandSourceUpdateParameters(config []interface{}) *types.UpdatePipeSourceParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.UpdatePipeSourceParameters
	for _, c := range config {
		param, ok := c.(map[string]interface{})
		if !ok {
			return nil
		}

		if val, ok := param["active_mq_broker"]; ok {
			parameters.ActiveMQBrokerParameters = expandSourceUpdateActiveMQBrokerParameters(val.([]interface{}))
		}

		if val, ok := param["dynamo_db_stream"]; ok {
			parameters.DynamoDBStreamParameters = expandSourceUpdateDynamoDBStreamParameters(val.([]interface{}))
		}

		if val, ok := param["kinesis_stream"]; ok {
			parameters.KinesisStreamParameters = expandSourceUpdateKinesisStreamParameters(val.([]interface{}))
		}

		if val, ok := param["managed_streaming_kafka"]; ok {
			parameters.ManagedStreamingKafkaParameters = expandSourceUpdateManagedStreamingKafkaParameters(val.([]interface{}))
		}

		if val, ok := param["rabbit_mq_broker"]; ok {
			parameters.RabbitMQBrokerParameters = expandSourceUpdateRabbitMQBrokerParameters(val.([]interface{}))
		}

		if val, ok := param["self_managed_kafka"]; ok {
			parameters.SelfManagedKafkaParameters = expandSourceUpdateSelfManagedKafkaParameters(val.([]interface{}))
		}

		if val, ok := param["sqs_queue"]; ok {
			parameters.SqsQueueParameters = expandSourceUpdateSqsQueueParameters(val.([]interface{}))
		}

		if val, ok := param["filter_criteria"]; ok {
			parameters.FilterCriteria = expandSourceFilterCriteria(val.([]interface{}))
		}
	}
	return &parameters
}

func expandSourceUpdateActiveMQBrokerParameters(config []interface{}) *types.UpdatePipeSourceActiveMQBrokerParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.UpdatePipeSourceActiveMQBrokerParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		if val, ok := param["credentials"]; ok {
			credentialsConfig := val.([]interface{})
			if len(credentialsConfig) != 0 {
				var credentialsParameters types.MQBrokerAccessCredentialsMemberBasicAuth
				for _, cc := range credentialsConfig {
					credentialsParam := cc.(map[string]interface{})
					credentialsParameters.Value = expandStringValue("basic_auth", credentialsParam)
				}
				parameters.Credentials = &credentialsParameters
			}
		}
	}
	return &parameters
}

func expandSourceUpdateDynamoDBStreamParameters(config []interface{}) *types.UpdatePipeSourceDynamoDBStreamParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.UpdatePipeSourceDynamoDBStreamParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.MaximumRecordAgeInSeconds = expandInt32("maximum_record_age_in_seconds", param)
		parameters.ParallelizationFactor = expandInt32("parallelization_factor", param)
		parameters.MaximumRetryAttempts = expandInt32("maximum_retry_attempts", param)
		onPartialBatchItemFailure := expandStringValue("on_partial_batch_item_failure", param)
		if onPartialBatchItemFailure != "" {
			parameters.OnPartialBatchItemFailure = types.OnPartialBatchItemFailureStreams(onPartialBatchItemFailure)
		}
		if val, ok := param["dead_letter_config"]; ok {
			parameters.DeadLetterConfig = expandSourceDeadLetterConfig(val.([]interface{}))
		}
	}
	return &parameters
}

func expandSourceUpdateKinesisStreamParameters(config []interface{}) *types.UpdatePipeSourceKinesisStreamParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.UpdatePipeSourceKinesisStreamParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.MaximumRecordAgeInSeconds = expandInt32("maximum_record_age_in_seconds", param)
		parameters.ParallelizationFactor = expandInt32("parallelization_factor", param)
		parameters.MaximumRetryAttempts = expandInt32("maximum_retry_attempts", param)

		onPartialBatchItemFailure := expandStringValue("on_partial_batch_item_failure", param)
		if onPartialBatchItemFailure != "" {
			parameters.OnPartialBatchItemFailure = types.OnPartialBatchItemFailureStreams(onPartialBatchItemFailure)
		}
		if val, ok := param["dead_letter_config"]; ok {
			parameters.DeadLetterConfig = expandSourceDeadLetterConfig(val.([]interface{}))
		}
	}
	return &parameters
}

func expandSourceUpdateManagedStreamingKafkaParameters(config []interface{}) *types.UpdatePipeSourceManagedStreamingKafkaParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.UpdatePipeSourceManagedStreamingKafkaParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)

		if val, ok := param["credentials"]; ok {
			credentialsConfig := val.([]interface{})
			if len(credentialsConfig) != 0 {
				for _, cc := range credentialsConfig {
					credentialsParam := cc.(map[string]interface{})
					if _, ok := credentialsParam["client_certificate_tls_auth"]; ok {
						var credentialsParameters types.MSKAccessCredentialsMemberClientCertificateTlsAuth
						credentialsParameters.Value = expandStringValue("client_certificate_tls_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
					if _, ok := credentialsParam["sasl_scram_512_auth"]; ok {
						var credentialsParameters types.MSKAccessCredentialsMemberSaslScram512Auth
						credentialsParameters.Value = expandStringValue("sasl_scram_512_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
				}
			}
		}
	}
	return &parameters
}

func expandSourceUpdateRabbitMQBrokerParameters(config []interface{}) *types.UpdatePipeSourceRabbitMQBrokerParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.UpdatePipeSourceRabbitMQBrokerParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)

		if val, ok := param["credentials"]; ok {
			credentialsConfig := val.([]interface{})
			if len(credentialsConfig) != 0 {
				var credentialsParameters types.MQBrokerAccessCredentialsMemberBasicAuth
				for _, cc := range credentialsConfig {
					credentialsParam := cc.(map[string]interface{})
					credentialsParameters.Value = expandStringValue("basic_auth", credentialsParam)
				}
				parameters.Credentials = &credentialsParameters
			}
		}
	}
	return &parameters
}

func expandSourceUpdateSelfManagedKafkaParameters(config []interface{}) *types.UpdatePipeSourceSelfManagedKafkaParameters {
	if len(config) == 0 {
		return nil
	}
	var parameters types.UpdatePipeSourceSelfManagedKafkaParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
		parameters.ServerRootCaCertificate = expandString("server_root_ca_certificate", param)

		if val, ok := param["credentials"]; ok {
			credentialsConfig := val.([]interface{})
			if len(credentialsConfig) != 0 {
				for _, cc := range credentialsConfig {
					credentialsParam := cc.(map[string]interface{})
					if _, ok := credentialsParam["basic_auth"]; ok {
						var credentialsParameters types.SelfManagedKafkaAccessConfigurationCredentialsMemberBasicAuth
						credentialsParameters.Value = expandStringValue("basic_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
					if _, ok := credentialsParam["client_certificate_tls_auth"]; ok {
						var credentialsParameters types.SelfManagedKafkaAccessConfigurationCredentialsMemberClientCertificateTlsAuth
						credentialsParameters.Value = expandStringValue("client_certificate_tls_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
					if _, ok := credentialsParam["sasl_scram_512_auth"]; ok {
						var credentialsParameters types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram512Auth
						credentialsParameters.Value = expandStringValue("sasl_scram_512_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
					if _, ok := credentialsParam["sasl_scram_256_auth"]; ok {
						var credentialsParameters types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram256Auth
						credentialsParameters.Value = expandStringValue("sasl_scram_256_auth", credentialsParam)
						parameters.Credentials = &credentialsParameters
					}
				}
			}
		}

		if val, ok := param["vpc"]; ok {
			vpcConfig := val.([]interface{})
			if len(vpcConfig) != 0 {
				var vpcParameters types.SelfManagedKafkaAccessConfigurationVpc
				for _, vc := range vpcConfig {
					vpcParam := vc.(map[string]interface{})
					if value, ok := vpcParam["security_groups"]; ok && value.(*schema.Set).Len() > 0 {
						vpcParameters.SecurityGroup = flex.ExpandStringValueSet(value.(*schema.Set))
					}
					if value, ok := vpcParam["subnets"]; ok && value.(*schema.Set).Len() > 0 {
						vpcParameters.Subnets = flex.ExpandStringValueSet(value.(*schema.Set))
					}
				}
				parameters.Vpc = &vpcParameters
			}
		}
	}

	return &parameters
}

func expandSourceUpdateSqsQueueParameters(config []interface{}) *types.UpdatePipeSourceSqsQueueParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.UpdatePipeSourceSqsQueueParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.BatchSize = expandInt32("batch_size", param)
		parameters.MaximumBatchingWindowInSeconds = expandInt32("maximum_batching_window_in_seconds", param)
	}

	return &parameters
}
