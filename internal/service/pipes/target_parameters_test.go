package pipes

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Test_expandTargetParameters(t *testing.T) {
	tests := map[string]struct {
		config   map[string]interface{}
		expected *types.PipeTargetParameters
	}{
		"batch_target config": {
			config: map[string]interface{}{
				"batch_target": []interface{}{
					map[string]interface{}{
						"job_definition": "job:test",
						"job_name":       "test",
						"retry_strategy": []interface{}{
							map[string]interface{}{
								"attempts": int32(2),
							},
						},
						"array_properties": []interface{}{
							map[string]interface{}{
								"size": int32(50),
							},
						},
						"parameters": []interface{}{
							map[string]interface{}{
								"key":   "key1",
								"value": "value1",
							},
							map[string]interface{}{
								"key":   "key2",
								"value": "value2",
							},
						},
						"depends_on": []interface{}{
							map[string]interface{}{
								"job_id": "jobID1",
								"type":   "N_TO_N",
							},
							map[string]interface{}{
								"job_id": "jobID2",
								"type":   "SEQUENTIAL",
							},
						},
						"container_overrides": []interface{}{
							map[string]interface{}{
								"command": schema.NewSet(schema.HashString, []interface{}{
									"command1",
									"command2",
								}),
								"environment": []interface{}{
									map[string]interface{}{
										"name":  "env1",
										"value": "valueEnv1",
									},
									map[string]interface{}{
										"name":  "env2",
										"value": "valueEnv2",
									},
								},
								"instance_type": "instanceType",
								"resource_requirements": []interface{}{
									map[string]interface{}{
										"type":  "VCPU",
										"value": "4",
									},
								},
							},
						},
					},
				},
			},
			expected: &types.PipeTargetParameters{
				BatchJobParameters: &types.PipeTargetBatchJobParameters{
					JobDefinition: aws.String("job:test"),
					JobName:       aws.String("test"),
					RetryStrategy: &types.BatchRetryStrategy{
						Attempts: 2,
					},
					ArrayProperties: &types.BatchArrayProperties{
						Size: 50,
					},
					Parameters: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					DependsOn: []types.BatchJobDependency{
						{
							JobId: aws.String("jobID1"),
							Type:  types.BatchJobDependencyTypeNToN,
						},
						{
							JobId: aws.String("jobID2"),
							Type:  types.BatchJobDependencyTypeSequential,
						},
					},
					ContainerOverrides: &types.BatchContainerOverrides{
						Command: []string{"command2", "command1"},
						Environment: []types.BatchEnvironmentVariable{
							{
								Name:  aws.String("env1"),
								Value: aws.String("valueEnv1"),
							},
							{
								Name:  aws.String("env2"),
								Value: aws.String("valueEnv2"),
							},
						},
						InstanceType: aws.String("instanceType"),
						ResourceRequirements: []types.BatchResourceRequirement{
							{
								Type:  types.BatchResourceRequirementTypeVcpu,
								Value: aws.String("4"),
							},
						},
					},
				},
			},
		},
		"cloudwatch_logs config": {
			config: map[string]interface{}{
				"cloudwatch_logs": []interface{}{
					map[string]interface{}{
						"log_stream_name": "job:test",
						"timestamp":       "2020-01-01T00:00:00Z",
					},
				},
			},
			expected: &types.PipeTargetParameters{
				CloudWatchLogsParameters: &types.PipeTargetCloudWatchLogsParameters{
					LogStreamName: aws.String("job:test"),
					Timestamp:     aws.String("2020-01-01T00:00:00Z"),
				},
			},
		},
		"ecs_task config": {
			config: map[string]interface{}{
				"ecs_task": []interface{}{
					map[string]interface{}{
						"task_definition_arn": "arn:test",
						"capacity_provider_strategy": []interface{}{
							map[string]interface{}{
								"capacity_provider": "capacityProvider",
								"weight":            int32(1),
								"base":              int32(10),
							},
						},
						"enable_ecs_managed_tags": true,
						"enable_execute_command":  true,
						"group":                   "group",
						"launch_type":             "FARGATE",
						"network_configuration": []interface{}{
							map[string]interface{}{
								"aws_vpc_configuration": []interface{}{
									map[string]interface{}{
										"assign_public_ip": "ENABLED",
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
							},
						},
						"placement_constraints": []interface{}{
							map[string]interface{}{
								"type":       "memberOf",
								"expression": "expression",
							},
						},
						"placement_strategy": []interface{}{
							map[string]interface{}{
								"type":  "binpack",
								"field": "field",
							},
						},
						"platform_version": "platformVersion",
						"propagate_tags":   "TASK_DEFINITION",
						"reference_id":     "referenceID",
						"task_count":       int32(1),
						"tags": []interface{}{
							map[string]interface{}{
								"key":   "key1",
								"value": "value1",
							},
						},
						"overrides": []interface{}{
							map[string]interface{}{
								"cpu":                "cpu1",
								"memory":             "mem2",
								"execution_role_arn": "arn:role",
								"task_role_arn":      "arn:role2",
								"inference_accelerator_overrides": []interface{}{
									map[string]interface{}{
										"device_name": "deviceName",
										"device_type": "deviceType",
									},
								},
								"ecs_ephemeral_storage": []interface{}{
									map[string]interface{}{
										"size_in_gib": int32(30),
									},
								},
								"container_overrides": []interface{}{
									map[string]interface{}{
										"cpu":                int32(5),
										"memory":             int32(6),
										"memory_reservation": int32(7),
										"name":               "name",
										"command": schema.NewSet(schema.HashString, []interface{}{
											"command1",
											"command2",
										}),
										"environment": []interface{}{
											map[string]interface{}{
												"name":  "env1",
												"value": "valueEnv1",
											},
										},
										"environment_files": []interface{}{
											map[string]interface{}{
												"value": "some:arnvalue",
												"type":  "s3",
											},
										},
										"resource_requirements": []interface{}{
											map[string]interface{}{
												"type":  "GPU",
												"value": "4",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &types.PipeTargetParameters{
				EcsTaskParameters: &types.PipeTargetEcsTaskParameters{
					TaskDefinitionArn: aws.String("arn:test"),
					CapacityProviderStrategy: []types.CapacityProviderStrategyItem{
						{
							CapacityProvider: aws.String("capacityProvider"),
							Weight:           1,
							Base:             10,
						},
					},
					EnableECSManagedTags: true,
					EnableExecuteCommand: true,
					Group:                aws.String("group"),
					LaunchType:           types.LaunchTypeFargate,
					NetworkConfiguration: &types.NetworkConfiguration{
						AwsvpcConfiguration: &types.AwsVpcConfiguration{
							AssignPublicIp: types.AssignPublicIpEnabled,
							SecurityGroups: []string{
								"sg2",
								"sg1",
							},
							Subnets: []string{
								"subnet1",
								"subnet2",
							},
						},
					},
					PlacementConstraints: []types.PlacementConstraint{
						{
							Type:       types.PlacementConstraintTypeMemberOf,
							Expression: aws.String("expression"),
						},
					},
					PlacementStrategy: []types.PlacementStrategy{
						{
							Type:  types.PlacementStrategyTypeBinpack,
							Field: aws.String("field"),
						},
					},
					PlatformVersion: aws.String("platformVersion"),
					PropagateTags:   types.PropagateTagsTaskDefinition,
					ReferenceId:     aws.String("referenceID"),
					TaskCount:       aws.Int32(1),
					Tags: []types.Tag{
						{
							Key:   aws.String("key1"),
							Value: aws.String("value1"),
						},
					},
					Overrides: &types.EcsTaskOverride{
						Cpu:              aws.String("cpu1"),
						Memory:           aws.String("mem2"),
						ExecutionRoleArn: aws.String("arn:role"),
						TaskRoleArn:      aws.String("arn:role2"),
						InferenceAcceleratorOverrides: []types.EcsInferenceAcceleratorOverride{
							{
								DeviceName: aws.String("deviceName"),
								DeviceType: aws.String("deviceType"),
							},
						},
						EphemeralStorage: &types.EcsEphemeralStorage{
							SizeInGiB: 30,
						},
						ContainerOverrides: []types.EcsContainerOverride{
							{
								Cpu:               aws.Int32(5),
								Memory:            aws.Int32(6),
								MemoryReservation: aws.Int32(7),
								Name:              aws.String("name"),
								Command:           []string{"command2", "command1"},
								Environment: []types.EcsEnvironmentVariable{
									{
										Name:  aws.String("env1"),
										Value: aws.String("valueEnv1"),
									},
								},
								EnvironmentFiles: []types.EcsEnvironmentFile{
									{
										Value: aws.String("some:arnvalue"),
										Type:  types.EcsEnvironmentFileTypeS3,
									},
								},
								ResourceRequirements: []types.EcsResourceRequirement{
									{
										Type:  types.EcsResourceRequirementTypeGpu,
										Value: aws.String("4"),
									},
								},
							},
						},
					},
				},
			},
		},
		"event_bridge_event_bus config": {
			config: map[string]interface{}{
				"event_bridge_event_bus": []interface{}{
					map[string]interface{}{
						"detail_type": "some.event",
						"endpoint_id": "endpointID",
						"source":      "source",
						"time":        "2020-01-01T00:00:00Z",
						"resources": schema.NewSet(schema.HashString, []interface{}{
							"id1",
							"id2",
						}),
					},
				},
			},
			expected: &types.PipeTargetParameters{
				EventBridgeEventBusParameters: &types.PipeTargetEventBridgeEventBusParameters{
					DetailType: aws.String("some.event"),
					EndpointId: aws.String("endpointID"),
					Source:     aws.String("source"),
					Time:       aws.String("2020-01-01T00:00:00Z"),
					Resources: []string{
						"id2",
						"id1",
					},
				},
			},
		},
		"http_parameters config": {
			config: map[string]interface{}{
				"http_parameters": []interface{}{
					map[string]interface{}{
						"path_parameters": []interface{}{"a", "b"},
						"header": []interface{}{
							map[string]interface{}{
								"key":   "key1",
								"value": "value1",
							},
							map[string]interface{}{
								"key":   "key2",
								"value": "value2",
							},
						},
						"query_string": []interface{}{
							map[string]interface{}{
								"key":   "key3",
								"value": "value3",
							},
							map[string]interface{}{
								"key":   "key4",
								"value": "value4",
							},
						},
					},
				},
			},
			expected: &types.PipeTargetParameters{
				HttpParameters: &types.PipeTargetHttpParameters{
					PathParameterValues: []string{"a", "b"},
					HeaderParameters: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					QueryStringParameters: map[string]string{
						"key3": "value3",
						"key4": "value4",
					},
				},
			},
		},
		"kinesis_stream config": {
			config: map[string]interface{}{
				"kinesis_stream": []interface{}{
					map[string]interface{}{
						"partition_key": "partitionKey",
					},
				},
			},
			expected: &types.PipeTargetParameters{
				KinesisStreamParameters: &types.PipeTargetKinesisStreamParameters{
					PartitionKey: aws.String("partitionKey"),
				},
			},
		},
		"lambda_function config": {
			config: map[string]interface{}{
				"lambda_function": []interface{}{
					map[string]interface{}{
						"invocation_type": "FIRE_AND_FORGET",
					},
				},
			},
			expected: &types.PipeTargetParameters{
				LambdaFunctionParameters: &types.PipeTargetLambdaFunctionParameters{
					InvocationType: types.PipeTargetInvocationTypeFireAndForget,
				},
			},
		},
		"redshift_data config": {
			config: map[string]interface{}{
				"redshift_data": []interface{}{
					map[string]interface{}{
						"database":           "database",
						"database_user":      "database_user",
						"secret_manager_arn": "arn:secrets",
						"statement_name":     "statement_name",
						"with_event":         true,
						"sqls": schema.NewSet(schema.HashString, []interface{}{
							"sql2",
							"sql1",
						}),
					},
				},
			},
			expected: &types.PipeTargetParameters{
				RedshiftDataParameters: &types.PipeTargetRedshiftDataParameters{
					Database:         aws.String("database"),
					DbUser:           aws.String("database_user"),
					SecretManagerArn: aws.String("arn:secrets"),
					StatementName:    aws.String("statement_name"),
					WithEvent:        true,
					Sqls:             []string{"sql2", "sql1"},
				},
			},
		},
		"sagemaker_pipeline config": {
			config: map[string]interface{}{
				"sagemaker_pipeline": []interface{}{
					map[string]interface{}{
						"parameters": []interface{}{
							map[string]interface{}{
								"name":  "name1",
								"value": "value1",
							},
							map[string]interface{}{
								"name":  "name2",
								"value": "value2",
							},
						},
					},
				},
			},
			expected: &types.PipeTargetParameters{
				SageMakerPipelineParameters: &types.PipeTargetSageMakerPipelineParameters{
					PipelineParameterList: []types.SageMakerPipelineParameter{
						{
							Name:  aws.String("name1"),
							Value: aws.String("value1"),
						},
						{
							Name:  aws.String("name2"),
							Value: aws.String("value2"),
						},
					},
				},
			},
		},
		"sqs_queue config": {
			config: map[string]interface{}{
				"sqs_queue": []interface{}{
					map[string]interface{}{
						"message_deduplication_id": "deduplication-id",
						"message_group_id":         "group-id",
					},
				},
			},
			expected: &types.PipeTargetParameters{
				SqsQueueParameters: &types.PipeTargetSqsQueueParameters{
					MessageDeduplicationId: aws.String("deduplication-id"),
					MessageGroupId:         aws.String("group-id"),
				},
			},
		},
		"step_function config": {
			config: map[string]interface{}{
				"step_function": []interface{}{
					map[string]interface{}{
						"invocation_type": "FIRE_AND_FORGET",
					},
				},
			},
			expected: &types.PipeTargetParameters{
				StepFunctionStateMachineParameters: &types.PipeTargetStateMachineParameters{
					InvocationType: types.PipeTargetInvocationTypeFireAndForget,
				},
			},
		},
		"input_template config": {
			config: map[string]interface{}{
				"input_template": "some template",
			},
			expected: &types.PipeTargetParameters{
				InputTemplate: aws.String("some template"),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := expandTargetParameters([]interface{}{tt.config})

			if diff := cmp.Diff(got, tt.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func Test_flattenTargetParameters(t *testing.T) {
	tests := map[string]struct {
		expected []map[string]interface{}
		config   *types.PipeTargetParameters
	}{
		"batch_target config": {
			expected: []map[string]interface{}{
				{
					"batch_target": []map[string]interface{}{
						{
							"job_definition": "job:test",
							"job_name":       "test",
							"retry_strategy": []map[string]interface{}{
								{
									"attempts": int32(2),
								},
							},
							"array_properties": []map[string]interface{}{
								{
									"size": int32(50),
								},
							},
							"parameters": []map[string]interface{}{
								{
									"key":   "key1",
									"value": "value1",
								},
								{
									"key":   "key2",
									"value": "value2",
								},
							},
							"depends_on": []map[string]interface{}{
								{
									"job_id": "jobID1",
									"type":   types.BatchJobDependencyTypeNToN,
								},
								{
									"job_id": "jobID2",
									"type":   types.BatchJobDependencyTypeSequential,
								},
							},
							"container_overrides": []map[string]interface{}{
								{
									"command": schema.NewSet(schema.HashString, []interface{}{
										"command1",
										"command2",
									}),
									"environment": []map[string]interface{}{
										{
											"name":  "env1",
											"value": "valueEnv1",
										},
										{
											"name":  "env2",
											"value": "valueEnv2",
										},
									},
									"instance_type": "instanceType",
									"resource_requirements": []map[string]interface{}{
										{
											"type":  types.BatchResourceRequirementTypeVcpu,
											"value": "4",
										},
									},
								},
							},
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				BatchJobParameters: &types.PipeTargetBatchJobParameters{
					JobDefinition: aws.String("job:test"),
					JobName:       aws.String("test"),
					RetryStrategy: &types.BatchRetryStrategy{
						Attempts: 2,
					},
					ArrayProperties: &types.BatchArrayProperties{
						Size: 50,
					},
					Parameters: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					DependsOn: []types.BatchJobDependency{
						{
							JobId: aws.String("jobID1"),
							Type:  types.BatchJobDependencyTypeNToN,
						},
						{
							JobId: aws.String("jobID2"),
							Type:  types.BatchJobDependencyTypeSequential,
						},
					},
					ContainerOverrides: &types.BatchContainerOverrides{
						Command: []string{"command2", "command1"},
						Environment: []types.BatchEnvironmentVariable{
							{
								Name:  aws.String("env1"),
								Value: aws.String("valueEnv1"),
							},
							{
								Name:  aws.String("env2"),
								Value: aws.String("valueEnv2"),
							},
						},
						InstanceType: aws.String("instanceType"),
						ResourceRequirements: []types.BatchResourceRequirement{
							{
								Type:  types.BatchResourceRequirementTypeVcpu,
								Value: aws.String("4"),
							},
						},
					},
				},
			},
		},
		"cloudwatch_logs config": {
			expected: []map[string]interface{}{
				{
					"cloudwatch_logs": []map[string]interface{}{
						{
							"log_stream_name": "job:test",
							"timestamp":       "2020-01-01T00:00:00Z",
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				CloudWatchLogsParameters: &types.PipeTargetCloudWatchLogsParameters{
					LogStreamName: aws.String("job:test"),
					Timestamp:     aws.String("2020-01-01T00:00:00Z"),
				},
			},
		},
		"ecs_task config": {
			expected: []map[string]interface{}{
				{
					"ecs_task": []map[string]interface{}{
						{
							"task_definition_arn": "arn:test",
							"capacity_provider_strategy": []map[string]interface{}{
								{
									"capacity_provider": "capacityProvider",
									"weight":            int32(1),
									"base":              int32(10),
								},
							},
							"enable_ecs_managed_tags": true,
							"enable_execute_command":  true,
							"group":                   "group",
							"launch_type":             types.LaunchTypeFargate,
							"network_configuration": []map[string]interface{}{
								{
									"aws_vpc_configuration": []map[string]interface{}{
										{
											"assign_public_ip": types.AssignPublicIpEnabled,
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
								},
							},
							"placement_constraints": []map[string]interface{}{
								{
									"type":       types.PlacementConstraintTypeMemberOf,
									"expression": "expression",
								},
							},
							"placement_strategy": []map[string]interface{}{
								{
									"type":  types.PlacementStrategyTypeBinpack,
									"field": "field",
								},
							},
							"platform_version": "platformVersion",
							"propagate_tags":   types.PropagateTagsTaskDefinition,
							"reference_id":     "referenceID",
							"task_count":       int32(1),
							"tags": []map[string]interface{}{
								{
									"key":   "key1",
									"value": "value1",
								},
							},
							"overrides": []map[string]interface{}{
								{
									"cpu":                "cpu1",
									"memory":             "mem2",
									"execution_role_arn": "arn:role",
									"task_role_arn":      "arn:role2",
									"inference_accelerator_overrides": []map[string]interface{}{
										{
											"device_name": "deviceName",
											"device_type": "deviceType",
										},
									},
									"ecs_ephemeral_storage": []map[string]interface{}{
										{
											"size_in_gib": int32(30),
										},
									},
									"container_overrides": []map[string]interface{}{
										{
											"cpu":                int32(5),
											"memory":             int32(6),
											"memory_reservation": int32(7),
											"name":               "name",
											"command": schema.NewSet(schema.HashString, []interface{}{
												"command1",
												"command2",
											}),
											"environment": []map[string]interface{}{
												{
													"name":  "env1",
													"value": "valueEnv1",
												},
											},
											"environment_files": []map[string]interface{}{
												{
													"value": "some:arnvalue",
													"type":  types.EcsEnvironmentFileTypeS3,
												},
											},
											"resource_requirements": []map[string]interface{}{
												{
													"type":  types.EcsResourceRequirementTypeGpu,
													"value": "4",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				EcsTaskParameters: &types.PipeTargetEcsTaskParameters{
					TaskDefinitionArn: aws.String("arn:test"),
					CapacityProviderStrategy: []types.CapacityProviderStrategyItem{
						{
							CapacityProvider: aws.String("capacityProvider"),
							Weight:           1,
							Base:             10,
						},
					},
					EnableECSManagedTags: true,
					EnableExecuteCommand: true,
					Group:                aws.String("group"),
					LaunchType:           types.LaunchTypeFargate,
					NetworkConfiguration: &types.NetworkConfiguration{
						AwsvpcConfiguration: &types.AwsVpcConfiguration{
							AssignPublicIp: types.AssignPublicIpEnabled,
							SecurityGroups: []string{
								"sg2",
								"sg1",
							},
							Subnets: []string{
								"subnet1",
								"subnet2",
							},
						},
					},
					PlacementConstraints: []types.PlacementConstraint{
						{
							Type:       types.PlacementConstraintTypeMemberOf,
							Expression: aws.String("expression"),
						},
					},
					PlacementStrategy: []types.PlacementStrategy{
						{
							Type:  types.PlacementStrategyTypeBinpack,
							Field: aws.String("field"),
						},
					},
					PlatformVersion: aws.String("platformVersion"),
					PropagateTags:   types.PropagateTagsTaskDefinition,
					ReferenceId:     aws.String("referenceID"),
					TaskCount:       aws.Int32(1),
					Tags: []types.Tag{
						{
							Key:   aws.String("key1"),
							Value: aws.String("value1"),
						},
					},
					Overrides: &types.EcsTaskOverride{
						Cpu:              aws.String("cpu1"),
						Memory:           aws.String("mem2"),
						ExecutionRoleArn: aws.String("arn:role"),
						TaskRoleArn:      aws.String("arn:role2"),
						InferenceAcceleratorOverrides: []types.EcsInferenceAcceleratorOverride{
							{
								DeviceName: aws.String("deviceName"),
								DeviceType: aws.String("deviceType"),
							},
						},
						EphemeralStorage: &types.EcsEphemeralStorage{
							SizeInGiB: 30,
						},
						ContainerOverrides: []types.EcsContainerOverride{
							{
								Cpu:               aws.Int32(5),
								Memory:            aws.Int32(6),
								MemoryReservation: aws.Int32(7),
								Name:              aws.String("name"),
								Command:           []string{"command2", "command1"},
								Environment: []types.EcsEnvironmentVariable{
									{
										Name:  aws.String("env1"),
										Value: aws.String("valueEnv1"),
									},
								},
								EnvironmentFiles: []types.EcsEnvironmentFile{
									{
										Value: aws.String("some:arnvalue"),
										Type:  types.EcsEnvironmentFileTypeS3,
									},
								},
								ResourceRequirements: []types.EcsResourceRequirement{
									{
										Type:  types.EcsResourceRequirementTypeGpu,
										Value: aws.String("4"),
									},
								},
							},
						},
					},
				},
			},
		},
		"event_bridge_event_bus config": {
			expected: []map[string]interface{}{
				{
					"event_bridge_event_bus": []map[string]interface{}{
						{
							"detail_type": "some.event",
							"endpoint_id": "endpointID",
							"source":      "source",
							"time":        "2020-01-01T00:00:00Z",
							"resources": schema.NewSet(schema.HashString, []interface{}{
								"id1",
								"id2",
							}),
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				EventBridgeEventBusParameters: &types.PipeTargetEventBridgeEventBusParameters{
					DetailType: aws.String("some.event"),
					EndpointId: aws.String("endpointID"),
					Source:     aws.String("source"),
					Time:       aws.String("2020-01-01T00:00:00Z"),
					Resources: []string{
						"id2",
						"id1",
					},
				},
			},
		},
		"http_parameters config": {
			expected: []map[string]interface{}{
				{
					"http_parameters": []map[string]interface{}{
						{
							"path_parameters": []interface{}{"a", "b"},
							"header": []map[string]interface{}{
								{
									"key":   "key1",
									"value": "value1",
								},
								{
									"key":   "key2",
									"value": "value2",
								},
							},
							"query_string": []map[string]interface{}{
								{
									"key":   "key3",
									"value": "value3",
								},
								{
									"key":   "key4",
									"value": "value4",
								},
							},
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				HttpParameters: &types.PipeTargetHttpParameters{
					PathParameterValues: []string{"a", "b"},
					HeaderParameters: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					QueryStringParameters: map[string]string{
						"key3": "value3",
						"key4": "value4",
					},
				},
			},
		},
		"kinesis_stream config": {
			expected: []map[string]interface{}{
				{
					"kinesis_stream": []map[string]interface{}{
						{
							"partition_key": "partitionKey",
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				KinesisStreamParameters: &types.PipeTargetKinesisStreamParameters{
					PartitionKey: aws.String("partitionKey"),
				},
			},
		},
		"lambda_function config": {
			expected: []map[string]interface{}{
				{
					"lambda_function": []map[string]interface{}{
						{
							"invocation_type": types.PipeTargetInvocationTypeFireAndForget,
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				LambdaFunctionParameters: &types.PipeTargetLambdaFunctionParameters{
					InvocationType: types.PipeTargetInvocationTypeFireAndForget,
				},
			},
		},
		"redshift_data config": {
			expected: []map[string]interface{}{
				{
					"redshift_data": []map[string]interface{}{
						{
							"database":           "database",
							"database_user":      "database_user",
							"secret_manager_arn": "arn:secrets",
							"statement_name":     "statement_name",
							"with_event":         true,
							"sqls": schema.NewSet(schema.HashString, []interface{}{
								"sql2",
								"sql1",
							}),
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				RedshiftDataParameters: &types.PipeTargetRedshiftDataParameters{
					Database:         aws.String("database"),
					DbUser:           aws.String("database_user"),
					SecretManagerArn: aws.String("arn:secrets"),
					StatementName:    aws.String("statement_name"),
					WithEvent:        true,
					Sqls:             []string{"sql2", "sql1"},
				},
			},
		},
		"sagemaker_pipeline config": {
			expected: []map[string]interface{}{
				{
					"sagemaker_pipeline": []map[string]interface{}{
						{
							"parameters": []map[string]interface{}{
								{
									"name":  "name1",
									"value": "value1",
								},
								{
									"name":  "name2",
									"value": "value2",
								},
							},
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				SageMakerPipelineParameters: &types.PipeTargetSageMakerPipelineParameters{
					PipelineParameterList: []types.SageMakerPipelineParameter{
						{
							Name:  aws.String("name1"),
							Value: aws.String("value1"),
						},
						{
							Name:  aws.String("name2"),
							Value: aws.String("value2"),
						},
					},
				},
			},
		},
		"sqs_queue config": {
			expected: []map[string]interface{}{
				{
					"sqs_queue": []map[string]interface{}{
						{
							"message_deduplication_id": "deduplication-id",
							"message_group_id":         "group-id",
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				SqsQueueParameters: &types.PipeTargetSqsQueueParameters{
					MessageDeduplicationId: aws.String("deduplication-id"),
					MessageGroupId:         aws.String("group-id"),
				},
			},
		},
		"step_function config": {
			expected: []map[string]interface{}{
				{
					"step_function": []map[string]interface{}{
						{
							"invocation_type": types.PipeTargetInvocationTypeFireAndForget,
						},
					},
				},
			},
			config: &types.PipeTargetParameters{
				StepFunctionStateMachineParameters: &types.PipeTargetStateMachineParameters{
					InvocationType: types.PipeTargetInvocationTypeFireAndForget,
				},
			},
		},
		"input_template config": {
			expected: []map[string]interface{}{
				{
					"input_template": "some template",
				},
			},
			config: &types.PipeTargetParameters{
				InputTemplate: aws.String("some template"),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := flattenTargetParameters(tt.config)

			if diff := cmp.Diff(got, tt.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
