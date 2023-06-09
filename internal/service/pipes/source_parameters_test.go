package pipes

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_expandSourceParameters(t *testing.T) {
	tests := map[string]struct {
		config   map[string]interface{}
		expected *types.PipeSourceParameters
	}{
		"active_mq_broker config": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"queue":                              "test",
						"credentials": []interface{}{
							map[string]interface{}{
								"basic_auth": "arn:secrets",
							},
						},
					},
				},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue":               []interface{}{},
				"filter_criteria":         []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				ActiveMQBrokerParameters: &types.PipeSourceActiveMQBrokerParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					QueueName:                      aws.String("test"),
					Credentials: &types.MQBrokerAccessCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"dynamo_db_stream config": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{},
				"dynamo_db_stream": []interface{}{
					map[string]interface{}{
						"starting_position":                  "LATEST",
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"maximum_record_age_in_seconds":      int32(120),
						"maximum_retry_attempts":             int32(3),
						"parallelization_factor":             int32(1),
						"on_partial_batch_item_failure":      "AUTOMATIC_BISECT",
						"dead_letter_config": []interface{}{
							map[string]interface{}{
								"arn": "arn:queue",
							},
						},
					},
				},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue":               []interface{}{},
				"filter_criteria":         []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				DynamoDBStreamParameters: &types.PipeSourceDynamoDBStreamParameters{
					StartingPosition:               "LATEST",
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					MaximumRecordAgeInSeconds:      aws.Int32(120),
					MaximumRetryAttempts:           aws.Int32(3),
					ParallelizationFactor:          aws.Int32(1),
					OnPartialBatchItemFailure:      "AUTOMATIC_BISECT",
					DeadLetterConfig: &types.DeadLetterConfig{
						Arn: aws.String("arn:queue"),
					},
				},
			},
		},
		"kinesis_stream config": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{},
				"dynamo_db_stream": []interface{}{},
				"kinesis_stream": []interface{}{
					map[string]interface{}{
						"starting_position":                  "AT_TIMESTAMP",
						"starting_position_timestamp":        "2020-01-01T00:00:00Z",
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"maximum_record_age_in_seconds":      int32(120),
						"maximum_retry_attempts":             int32(3),
						"parallelization_factor":             int32(1),
						"on_partial_batch_item_failure":      "AUTOMATIC_BISECT",
						"dead_letter_config": []interface{}{
							map[string]interface{}{
								"arn": "arn:queue",
							},
						},
					},
				},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue":               []interface{}{},
				"filter_criteria":         []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				KinesisStreamParameters: &types.PipeSourceKinesisStreamParameters{
					StartingPosition:               "AT_TIMESTAMP",
					StartingPositionTimestamp:      aws.Time(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					MaximumRecordAgeInSeconds:      aws.Int32(120),
					MaximumRetryAttempts:           aws.Int32(3),
					ParallelizationFactor:          aws.Int32(1),
					OnPartialBatchItemFailure:      "AUTOMATIC_BISECT",
					DeadLetterConfig: &types.DeadLetterConfig{
						Arn: aws.String("arn:queue"),
					},
				},
			},
		},
		"managed_streaming_kafka config with client_certificate_tls_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{},
				"dynamo_db_stream": []interface{}{},
				"kinesis_stream":   []interface{}{},
				"managed_streaming_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"topic":                              "test",
						"consumer_group_id":                  "group",
						"starting_position":                  "LATEST",
						"credentials": []interface{}{
							map[string]interface{}{
								"client_certificate_tls_auth": "arn:secrets",
							},
						},
					},
				},
				"rabbit_mq_broker":   []interface{}{},
				"self_managed_kafka": []interface{}{},
				"sqs_queue":          []interface{}{},
				"filter_criteria":    []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				ManagedStreamingKafkaParameters: &types.PipeSourceManagedStreamingKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					Credentials: &types.MSKAccessCredentialsMemberClientCertificateTlsAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"managed_streaming_kafka config with sasl_scram_512_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{},
				"dynamo_db_stream": []interface{}{},
				"kinesis_stream":   []interface{}{},
				"managed_streaming_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"topic":                              "test",
						"consumer_group_id":                  "group",
						"starting_position":                  "LATEST",
						"credentials": []interface{}{
							map[string]interface{}{
								"sasl_scram_512_auth": "arn:secrets",
							},
						},
					},
				},
				"rabbit_mq_broker":   []interface{}{},
				"self_managed_kafka": []interface{}{},
				"sqs_queue":          []interface{}{},
				"filter_criteria":    []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				ManagedStreamingKafkaParameters: &types.PipeSourceManagedStreamingKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					Credentials: &types.MSKAccessCredentialsMemberSaslScram512Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"rabbit_mq_broker config": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"queue":                              "test",
						"virtual_host":                       "hosting",
						"credentials": []interface{}{
							map[string]interface{}{
								"basic_auth": "arn:secrets",
							},
						},
					},
				},
				"self_managed_kafka": []interface{}{},
				"sqs_queue":          []interface{}{},
				"filter_criteria":    []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				RabbitMQBrokerParameters: &types.PipeSourceRabbitMQBrokerParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					QueueName:                      aws.String("test"),
					VirtualHost:                    aws.String("hosting"),
					Credentials: &types.MQBrokerAccessCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with basic_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"topic":                              "test",
						"consumer_group_id":                  "group",
						"starting_position":                  "LATEST",
						"server_root_ca_certificate":         "arn:ca:cert",
						"servers": schema.NewSet(schema.HashString, []interface{}{
							"server1",
							"server2",
						}),
						"vpc": []interface{}{
							map[string]interface{}{
								"security_groups": schema.NewSet(schema.HashString, []interface{}{
									"sg1",
									"sg2",
								}),
								"subnets": schema.NewSet(schema.HashString, []interface{}{
									"subnet1",
									"subnet2",
								}),
							},
						},
						"credentials": []interface{}{
							map[string]interface{}{
								"basic_auth": "arn:secrets",
							},
						},
					},
				},
				"sqs_queue":       []interface{}{},
				"filter_criteria": []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				SelfManagedKafkaParameters: &types.PipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					AdditionalBootstrapServers:     []string{"server2", "server1"},
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with client_certificate_tls_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"topic":                              "test",
						"consumer_group_id":                  "group",
						"starting_position":                  "LATEST",
						"server_root_ca_certificate":         "arn:ca:cert",
						"servers": schema.NewSet(schema.HashString, []interface{}{
							"server1",
							"server2",
						}),
						"vpc": []interface{}{
							map[string]interface{}{
								"security_groups": schema.NewSet(schema.HashString, []interface{}{
									"sg1",
									"sg2",
								}),
								"subnets": schema.NewSet(schema.HashString, []interface{}{
									"subnet1",
									"subnet2",
								}),
							},
						},
						"credentials": []interface{}{
							map[string]interface{}{
								"client_certificate_tls_auth": "arn:secrets",
							},
						},
					},
				},
				"sqs_queue":       []interface{}{},
				"filter_criteria": []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				SelfManagedKafkaParameters: &types.PipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					AdditionalBootstrapServers:     []string{"server2", "server1"},
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberClientCertificateTlsAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with sasl_scram_512_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"topic":                              "test",
						"consumer_group_id":                  "group",
						"starting_position":                  "LATEST",
						"server_root_ca_certificate":         "arn:ca:cert",
						"servers": schema.NewSet(schema.HashString, []interface{}{
							"server1",
							"server2",
						}),
						"vpc": []interface{}{
							map[string]interface{}{
								"security_groups": schema.NewSet(schema.HashString, []interface{}{
									"sg1",
									"sg2",
								}),
								"subnets": schema.NewSet(schema.HashString, []interface{}{
									"subnet1",
									"subnet2",
								}),
							},
						},
						"credentials": []interface{}{
							map[string]interface{}{
								"sasl_scram_512_auth": "arn:secrets",
							},
						},
					},
				},
				"sqs_queue":       []interface{}{},
				"filter_criteria": []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				SelfManagedKafkaParameters: &types.PipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					AdditionalBootstrapServers:     []string{"server2", "server1"},
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram512Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with sasl_scram_256_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"topic":                              "test",
						"consumer_group_id":                  "group",
						"starting_position":                  "LATEST",
						"server_root_ca_certificate":         "arn:ca:cert",
						"servers": schema.NewSet(schema.HashString, []interface{}{
							"server1",
							"server2",
						}),
						"vpc": []interface{}{
							map[string]interface{}{
								"security_groups": schema.NewSet(schema.HashString, []interface{}{
									"sg1",
									"sg2",
								}),
								"subnets": schema.NewSet(schema.HashString, []interface{}{
									"subnet1",
									"subnet2",
								}),
							},
						},
						"credentials": []interface{}{
							map[string]interface{}{
								"sasl_scram_256_auth": "arn:secrets",
							},
						},
					},
				},
				"sqs_queue":       []interface{}{},
				"filter_criteria": []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				SelfManagedKafkaParameters: &types.PipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					AdditionalBootstrapServers:     []string{"server2", "server1"},
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram256Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"sqs_queue config": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
					},
				},
				"filter_criteria": []interface{}{},
			},
			expected: &types.PipeSourceParameters{
				SqsQueueParameters: &types.PipeSourceSqsQueueParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
				},
			},
		},
		"filter_criteria config": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue":               []interface{}{},
				"filter_criteria": []interface{}{
					map[string]interface{}{
						"filter": []interface{}{
							map[string]interface{}{
								"pattern": "1",
							},
							map[string]interface{}{
								"pattern": "2",
							},
						},
					},
				},
			},
			expected: &types.PipeSourceParameters{
				FilterCriteria: &types.FilterCriteria{
					Filters: []types.Filter{
						{
							Pattern: aws.String("1"),
						},
						{
							Pattern: aws.String("2"),
						},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := expandSourceParameters([]interface{}{tt.config})

			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_flattenSourceParameters(t *testing.T) {
	tests := map[string]struct {
		config   *types.PipeSourceParameters
		expected []map[string]interface{}
	}{
		"active_mq_broker config": {
			expected: []map[string]interface{}{
				{
					"active_mq_broker": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"queue":                              "test",
							"credentials": []map[string]interface{}{
								{
									"basic_auth": "arn:secrets",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				ActiveMQBrokerParameters: &types.PipeSourceActiveMQBrokerParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					QueueName:                      aws.String("test"),
					Credentials: &types.MQBrokerAccessCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"dynamo_db_stream config": {
			expected: []map[string]interface{}{
				{
					"dynamo_db_stream": []map[string]interface{}{
						{
							"starting_position":                  types.DynamoDBStreamStartPositionLatest,
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"maximum_record_age_in_seconds":      int32(120),
							"maximum_retry_attempts":             int32(3),
							"parallelization_factor":             int32(1),
							"on_partial_batch_item_failure":      types.OnPartialBatchItemFailureStreamsAutomaticBisect,
							"dead_letter_config": []map[string]interface{}{
								{
									"arn": "arn:queue",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				DynamoDBStreamParameters: &types.PipeSourceDynamoDBStreamParameters{
					StartingPosition:               "LATEST",
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					MaximumRecordAgeInSeconds:      aws.Int32(120),
					MaximumRetryAttempts:           aws.Int32(3),
					ParallelizationFactor:          aws.Int32(1),
					OnPartialBatchItemFailure:      "AUTOMATIC_BISECT",
					DeadLetterConfig: &types.DeadLetterConfig{
						Arn: aws.String("arn:queue"),
					},
				},
			},
		},
		"kinesis_stream config": {
			expected: []map[string]interface{}{
				{
					"kinesis_stream": []map[string]interface{}{
						{
							"starting_position":                  types.KinesisStreamStartPositionAtTimestamp,
							"starting_position_timestamp":        "2020-01-01T00:00:00Z",
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"maximum_record_age_in_seconds":      int32(120),
							"maximum_retry_attempts":             int32(3),
							"parallelization_factor":             int32(1),
							"on_partial_batch_item_failure":      types.OnPartialBatchItemFailureStreamsAutomaticBisect,
							"dead_letter_config": []map[string]interface{}{
								{
									"arn": "arn:queue",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				KinesisStreamParameters: &types.PipeSourceKinesisStreamParameters{
					StartingPosition:               "AT_TIMESTAMP",
					StartingPositionTimestamp:      aws.Time(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					MaximumRecordAgeInSeconds:      aws.Int32(120),
					MaximumRetryAttempts:           aws.Int32(3),
					ParallelizationFactor:          aws.Int32(1),
					OnPartialBatchItemFailure:      "AUTOMATIC_BISECT",
					DeadLetterConfig: &types.DeadLetterConfig{
						Arn: aws.String("arn:queue"),
					},
				},
			},
		},
		"managed_streaming_kafka config with client_certificate_tls_auth authentication": {
			expected: []map[string]interface{}{
				{
					"managed_streaming_kafka": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"topic":                              "test",
							"consumer_group_id":                  "group",
							"starting_position":                  types.MSKStartPositionLatest,
							"credentials": []map[string]interface{}{
								{
									"client_certificate_tls_auth": "arn:secrets",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				ManagedStreamingKafkaParameters: &types.PipeSourceManagedStreamingKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					Credentials: &types.MSKAccessCredentialsMemberClientCertificateTlsAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"managed_streaming_kafka config with sasl_scram_512_auth authentication": {
			expected: []map[string]interface{}{
				{
					"managed_streaming_kafka": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"topic":                              "test",
							"consumer_group_id":                  "group",
							"starting_position":                  types.MSKStartPositionLatest,
							"credentials": []map[string]interface{}{
								{
									"sasl_scram_512_auth": "arn:secrets",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				ManagedStreamingKafkaParameters: &types.PipeSourceManagedStreamingKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					Credentials: &types.MSKAccessCredentialsMemberSaslScram512Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"rabbit_mq_broker config": {
			expected: []map[string]interface{}{
				{
					"rabbit_mq_broker": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"queue":                              "test",
							"virtual_host":                       "hosting",
							"credentials": []map[string]interface{}{
								{
									"basic_auth": "arn:secrets",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				RabbitMQBrokerParameters: &types.PipeSourceRabbitMQBrokerParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					QueueName:                      aws.String("test"),
					VirtualHost:                    aws.String("hosting"),
					Credentials: &types.MQBrokerAccessCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with basic_auth authentication": {
			expected: []map[string]interface{}{
				{
					"self_managed_kafka": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"topic":                              "test",
							"consumer_group_id":                  "group",
							"starting_position":                  types.SelfManagedKafkaStartPositionLatest,
							"server_root_ca_certificate":         "arn:ca:cert",
							"servers": schema.NewSet(schema.HashString, []interface{}{
								"server1",
								"server2",
							}),
							"vpc": []map[string]interface{}{
								{
									"security_groups": schema.NewSet(schema.HashString, []interface{}{
										"sg1",
										"sg2",
									}),
									"subnets": schema.NewSet(schema.HashString, []interface{}{
										"subnet1",
										"subnet2",
									}),
								},
							},
							"credentials": []map[string]interface{}{
								{
									"basic_auth": "arn:secrets",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				SelfManagedKafkaParameters: &types.PipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					AdditionalBootstrapServers:     []string{"server2", "server1"},
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with client_certificate_tls_auth authentication": {
			expected: []map[string]interface{}{
				{
					"self_managed_kafka": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"topic":                              "test",
							"consumer_group_id":                  "group",
							"starting_position":                  types.SelfManagedKafkaStartPositionLatest,
							"server_root_ca_certificate":         "arn:ca:cert",
							"servers": schema.NewSet(schema.HashString, []interface{}{
								"server1",
								"server2",
							}),
							"vpc": []map[string]interface{}{
								{
									"security_groups": schema.NewSet(schema.HashString, []interface{}{
										"sg1",
										"sg2",
									}),
									"subnets": schema.NewSet(schema.HashString, []interface{}{
										"subnet1",
										"subnet2",
									}),
								},
							},
							"credentials": []map[string]interface{}{
								{
									"client_certificate_tls_auth": "arn:secrets",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				SelfManagedKafkaParameters: &types.PipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					AdditionalBootstrapServers:     []string{"server2", "server1"},
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberClientCertificateTlsAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with sasl_scram_512_auth authentication": {
			expected: []map[string]interface{}{
				{
					"self_managed_kafka": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"topic":                              "test",
							"consumer_group_id":                  "group",
							"starting_position":                  types.SelfManagedKafkaStartPositionLatest,
							"server_root_ca_certificate":         "arn:ca:cert",
							"servers": schema.NewSet(schema.HashString, []interface{}{
								"server1",
								"server2",
							}),
							"vpc": []map[string]interface{}{
								{
									"security_groups": schema.NewSet(schema.HashString, []interface{}{
										"sg1",
										"sg2",
									}),
									"subnets": schema.NewSet(schema.HashString, []interface{}{
										"subnet1",
										"subnet2",
									}),
								},
							},
							"credentials": []map[string]interface{}{
								{
									"sasl_scram_512_auth": "arn:secrets",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				SelfManagedKafkaParameters: &types.PipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					AdditionalBootstrapServers:     []string{"server2", "server1"},
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram512Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with sasl_scram_256_auth authentication": {
			expected: []map[string]interface{}{
				{
					"self_managed_kafka": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
							"topic":                              "test",
							"consumer_group_id":                  "group",
							"starting_position":                  types.SelfManagedKafkaStartPositionLatest,
							"server_root_ca_certificate":         "arn:ca:cert",
							"servers": schema.NewSet(schema.HashString, []interface{}{
								"server1",
								"server2",
							}),
							"vpc": []map[string]interface{}{
								{
									"security_groups": schema.NewSet(schema.HashString, []interface{}{
										"sg1",
										"sg2",
									}),
									"subnets": schema.NewSet(schema.HashString, []interface{}{
										"subnet1",
										"subnet2",
									}),
								},
							},
							"credentials": []map[string]interface{}{
								{
									"sasl_scram_256_auth": "arn:secrets",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				SelfManagedKafkaParameters: &types.PipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					TopicName:                      aws.String("test"),
					ConsumerGroupID:                aws.String("group"),
					StartingPosition:               "LATEST",
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					AdditionalBootstrapServers:     []string{"server2", "server1"},
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram256Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"sqs_queue config": {
			expected: []map[string]interface{}{
				{
					"sqs_queue": []map[string]interface{}{
						{
							"batch_size":                         int32(10),
							"maximum_batching_window_in_seconds": int32(60),
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				SqsQueueParameters: &types.PipeSourceSqsQueueParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
				},
			},
		},
		"filter_criteria config": {
			expected: []map[string]interface{}{
				{
					"filter_criteria": []map[string]interface{}{
						{
							"filter": []map[string]interface{}{
								{
									"pattern": "1",
								},
								{
									"pattern": "2",
								},
							},
						},
					},
				},
			},
			config: &types.PipeSourceParameters{
				FilterCriteria: &types.FilterCriteria{
					Filters: []types.Filter{
						{
							Pattern: aws.String("1"),
						},
						{
							Pattern: aws.String("2"),
						},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := flattenSourceParameters(tt.config)

			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_expandSourceUpdateParameters(t *testing.T) {
	tests := map[string]struct {
		config   map[string]interface{}
		expected *types.UpdatePipeSourceParameters
	}{
		"active_mq_broker config": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"credentials": []interface{}{
							map[string]interface{}{
								"basic_auth": "arn:secrets",
							},
						},
					},
				},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue":               []interface{}{},
				"filter_criteria":         []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				ActiveMQBrokerParameters: &types.UpdatePipeSourceActiveMQBrokerParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					Credentials: &types.MQBrokerAccessCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"dynamo_db_stream config": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{},
				"dynamo_db_stream": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"maximum_record_age_in_seconds":      int32(120),
						"maximum_retry_attempts":             int32(3),
						"parallelization_factor":             int32(1),
						"on_partial_batch_item_failure":      "AUTOMATIC_BISECT",
						"dead_letter_config": []interface{}{
							map[string]interface{}{
								"arn": "arn:queue",
							},
						},
					},
				},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue":               []interface{}{},
				"filter_criteria":         []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				DynamoDBStreamParameters: &types.UpdatePipeSourceDynamoDBStreamParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					MaximumRecordAgeInSeconds:      aws.Int32(120),
					MaximumRetryAttempts:           aws.Int32(3),
					ParallelizationFactor:          aws.Int32(1),
					OnPartialBatchItemFailure:      "AUTOMATIC_BISECT",
					DeadLetterConfig: &types.DeadLetterConfig{
						Arn: aws.String("arn:queue"),
					},
				},
			},
		},
		"kinesis_stream config": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{},
				"dynamo_db_stream": []interface{}{},
				"kinesis_stream": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"maximum_record_age_in_seconds":      int32(120),
						"maximum_retry_attempts":             int32(3),
						"parallelization_factor":             int32(1),
						"on_partial_batch_item_failure":      "AUTOMATIC_BISECT",
						"dead_letter_config": []interface{}{
							map[string]interface{}{
								"arn": "arn:queue",
							},
						},
					},
				},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue":               []interface{}{},
				"filter_criteria":         []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				KinesisStreamParameters: &types.UpdatePipeSourceKinesisStreamParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					MaximumRecordAgeInSeconds:      aws.Int32(120),
					MaximumRetryAttempts:           aws.Int32(3),
					ParallelizationFactor:          aws.Int32(1),
					OnPartialBatchItemFailure:      "AUTOMATIC_BISECT",
					DeadLetterConfig: &types.DeadLetterConfig{
						Arn: aws.String("arn:queue"),
					},
				},
			},
		},
		"managed_streaming_kafka config with client_certificate_tls_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{},
				"dynamo_db_stream": []interface{}{},
				"kinesis_stream":   []interface{}{},
				"managed_streaming_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"credentials": []interface{}{
							map[string]interface{}{
								"client_certificate_tls_auth": "arn:secrets",
							},
						},
					},
				},
				"rabbit_mq_broker":   []interface{}{},
				"self_managed_kafka": []interface{}{},
				"sqs_queue":          []interface{}{},
				"filter_criteria":    []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				ManagedStreamingKafkaParameters: &types.UpdatePipeSourceManagedStreamingKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					Credentials: &types.MSKAccessCredentialsMemberClientCertificateTlsAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"managed_streaming_kafka config with sasl_scram_512_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker": []interface{}{},
				"dynamo_db_stream": []interface{}{},
				"kinesis_stream":   []interface{}{},
				"managed_streaming_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"credentials": []interface{}{
							map[string]interface{}{
								"sasl_scram_512_auth": "arn:secrets",
							},
						},
					},
				},
				"rabbit_mq_broker":   []interface{}{},
				"self_managed_kafka": []interface{}{},
				"sqs_queue":          []interface{}{},
				"filter_criteria":    []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				ManagedStreamingKafkaParameters: &types.UpdatePipeSourceManagedStreamingKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					Credentials: &types.MSKAccessCredentialsMemberSaslScram512Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"rabbit_mq_broker config": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"credentials": []interface{}{
							map[string]interface{}{
								"basic_auth": "arn:secrets",
							},
						},
					},
				},
				"self_managed_kafka": []interface{}{},
				"sqs_queue":          []interface{}{},
				"filter_criteria":    []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				RabbitMQBrokerParameters: &types.UpdatePipeSourceRabbitMQBrokerParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					Credentials: &types.MQBrokerAccessCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with basic_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"server_root_ca_certificate":         "arn:ca:cert",
						"vpc": []interface{}{
							map[string]interface{}{
								"security_groups": schema.NewSet(schema.HashString, []interface{}{
									"sg1",
									"sg2",
								}),
								"subnets": schema.NewSet(schema.HashString, []interface{}{
									"subnet1",
									"subnet2",
								}),
							},
						},
						"credentials": []interface{}{
							map[string]interface{}{
								"basic_auth": "arn:secrets",
							},
						},
					},
				},
				"sqs_queue":       []interface{}{},
				"filter_criteria": []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				SelfManagedKafkaParameters: &types.UpdatePipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberBasicAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with client_certificate_tls_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"server_root_ca_certificate":         "arn:ca:cert",
						"vpc": []interface{}{
							map[string]interface{}{
								"security_groups": schema.NewSet(schema.HashString, []interface{}{
									"sg1",
									"sg2",
								}),
								"subnets": schema.NewSet(schema.HashString, []interface{}{
									"subnet1",
									"subnet2",
								}),
							},
						},
						"credentials": []interface{}{
							map[string]interface{}{
								"client_certificate_tls_auth": "arn:secrets",
							},
						},
					},
				},
				"sqs_queue":       []interface{}{},
				"filter_criteria": []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				SelfManagedKafkaParameters: &types.UpdatePipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberClientCertificateTlsAuth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with sasl_scram_512_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"server_root_ca_certificate":         "arn:ca:cert",
						"vpc": []interface{}{
							map[string]interface{}{
								"security_groups": schema.NewSet(schema.HashString, []interface{}{
									"sg1",
									"sg2",
								}),
								"subnets": schema.NewSet(schema.HashString, []interface{}{
									"subnet1",
									"subnet2",
								}),
							},
						},
						"credentials": []interface{}{
							map[string]interface{}{
								"sasl_scram_512_auth": "arn:secrets",
							},
						},
					},
				},
				"sqs_queue":       []interface{}{},
				"filter_criteria": []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				SelfManagedKafkaParameters: &types.UpdatePipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram512Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"self_managed_kafka config with sasl_scram_256_auth authentication": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
						"server_root_ca_certificate":         "arn:ca:cert",
						"vpc": []interface{}{
							map[string]interface{}{
								"security_groups": schema.NewSet(schema.HashString, []interface{}{
									"sg1",
									"sg2",
								}),
								"subnets": schema.NewSet(schema.HashString, []interface{}{
									"subnet1",
									"subnet2",
								}),
							},
						},
						"credentials": []interface{}{
							map[string]interface{}{
								"sasl_scram_256_auth": "arn:secrets",
							},
						},
					},
				},
				"sqs_queue":       []interface{}{},
				"filter_criteria": []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				SelfManagedKafkaParameters: &types.UpdatePipeSourceSelfManagedKafkaParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
					ServerRootCaCertificate:        aws.String("arn:ca:cert"),
					Vpc: &types.SelfManagedKafkaAccessConfigurationVpc{
						SecurityGroup: []string{"sg2", "sg1"},
						Subnets:       []string{"subnet1", "subnet2"},
					},
					Credentials: &types.SelfManagedKafkaAccessConfigurationCredentialsMemberSaslScram256Auth{
						Value: "arn:secrets",
					},
				},
			},
		},
		"sqs_queue config": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue": []interface{}{
					map[string]interface{}{
						"batch_size":                         int32(10),
						"maximum_batching_window_in_seconds": int32(60),
					},
				},
				"filter_criteria": []interface{}{},
			},
			expected: &types.UpdatePipeSourceParameters{
				SqsQueueParameters: &types.UpdatePipeSourceSqsQueueParameters{
					BatchSize:                      aws.Int32(10),
					MaximumBatchingWindowInSeconds: aws.Int32(60),
				},
			},
		},
		"filter_criteria config": {
			config: map[string]interface{}{
				"active_mq_broker":        []interface{}{},
				"dynamo_db_stream":        []interface{}{},
				"kinesis_stream":          []interface{}{},
				"managed_streaming_kafka": []interface{}{},
				"rabbit_mq_broker":        []interface{}{},
				"self_managed_kafka":      []interface{}{},
				"sqs_queue":               []interface{}{},
				"filter_criteria": []interface{}{
					map[string]interface{}{
						"filter": []interface{}{
							map[string]interface{}{
								"pattern": "1",
							},
							map[string]interface{}{
								"pattern": "2",
							},
						},
					},
				},
			},
			expected: &types.UpdatePipeSourceParameters{
				FilterCriteria: &types.FilterCriteria{
					Filters: []types.Filter{
						{
							Pattern: aws.String("1"),
						},
						{
							Pattern: aws.String("2"),
						},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := expandSourceUpdateParameters([]interface{}{tt.config})

			assert.Equal(t, tt.expected, got)
		})
	}
}
