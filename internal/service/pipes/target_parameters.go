// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pipes

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func targetParametersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"batch_job_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"array_properties": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrSize: {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(2, 10000),
										},
									},
								},
							},
							"container_overrides": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"command": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										names.AttrEnvironment: {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													names.AttrValue: {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										names.AttrInstanceType: {
											Type:     schema.TypeString,
											Optional: true,
										},
										"resource_requirement": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrType: {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[types.BatchResourceRequirementType](),
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
							"depends_on": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 20,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"job_id": {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrType: {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.BatchJobDependencyType](),
										},
									},
								},
							},
							"job_definition": {
								Type:     schema.TypeString,
								Required: true,
							},
							"job_name": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
							names.AttrParameters: {
								Type:     schema.TypeMap,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"retry_strategy": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"attempts": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(1, 10),
										},
									},
								},
							},
						},
					},
				},
				"cloudwatch_logs_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"log_stream_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 256),
							},
							"timestamp": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 256),
									validation.StringMatch(regexache.MustCompile(`^\$(\.[\w/_-]+(\[(\d+|\*)\])*)*$`), ""),
								),
							},
						},
					},
				},
				"ecs_task_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrCapacityProviderStrategy: {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 6,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"base": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(0, 100000),
										},
										"capacity_provider": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
										names.AttrWeight: {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(0, 1000),
										},
									},
								},
							},
							"enable_ecs_managed_tags": {
								Type:     schema.TypeBool,
								Optional: true,
							},
							"enable_execute_command": {
								Type:     schema.TypeBool,
								Optional: true,
							},
							"group": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							"launch_type": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.LaunchType](),
							},
							names.AttrNetworkConfiguration: {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"aws_vpc_configuration": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"assign_public_ip": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.AssignPublicIp](),
													},
													names.AttrSecurityGroups: {
														Type:     schema.TypeSet,
														Optional: true,
														MaxItems: 5,
														Elem: &schema.Schema{
															Type: schema.TypeString,
															ValidateFunc: validation.All(
																validation.StringLenBetween(1, 1024),
																validation.StringMatch(regexache.MustCompile(`^sg-[0-9A-Za-z]*|(\$(\.[\w/_-]+(\[(\d+|\*)\])*)*)$`), ""),
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
																validation.StringMatch(regexache.MustCompile(`^subnet-[0-9a-z]*|(\$(\.[\w/_-]+(\[(\d+|\*)\])*)*)$`), ""),
															),
														},
													},
												},
											},
										},
									},
								},
							},
							"overrides": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"container_override": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"command": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Schema{
															Type: schema.TypeString,
														},
													},
													"cpu": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													names.AttrEnvironment: {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrName: {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																names.AttrValue: {
																	Type:     schema.TypeString,
																	Optional: true,
																},
															},
														},
													},
													"environment_file": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrType: {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[types.EcsEnvironmentFileType](),
																},
																names.AttrValue: {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
															},
														},
													},
													"memory": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"memory_reservation": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													names.AttrName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													"resource_requirement": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrType: {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[types.EcsResourceRequirementType](),
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
										"cpu": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"ephemeral_storage": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"size_in_gib": {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntBetween(21, 200),
													},
												},
											},
										},
										names.AttrExecutionRoleARN: {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verify.ValidARN,
										},
										"inference_accelerator_override": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrDeviceName: {
														Type:     schema.TypeString,
														Optional: true,
													},
													"device_type": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										"memory": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"task_role_arn": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
							"placement_constraint": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrExpression: {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(1, 2000),
										},
										names.AttrType: {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.PlacementConstraintType](),
										},
									},
								},
							},
							"placement_strategy": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 5,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrField: {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
										names.AttrType: {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.PlacementStrategyType](),
										},
									},
								},
							},
							"platform_version": {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrPropagateTags: {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.PropagateTags](),
							},
							"reference_id": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 1024),
							},
							names.AttrTags: tftags.TagsSchema(),
							"task_count": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"task_definition_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"eventbridge_event_bus_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"detail_type": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 128),
							},
							"endpoint_id": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 50),
									validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+[\.][0-9A-Za-z-]+$`), ""),
								),
							},
							names.AttrResources: {
								Type:     schema.TypeSet,
								Optional: true,
								MaxItems: 10,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: verify.ValidARN,
								},
							},
							names.AttrSource: {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 256),
								),
							},
							"time": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 256),
									validation.StringMatch(regexache.MustCompile(`^\$(\.[\w/_-]+(\[(\d+|\*)\])*)*$`), ""),
								),
							},
						},
					},
				},
				"http_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
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
				"kinesis_stream_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"partition_key": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
						},
					},
				},
				"lambda_function_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"invocation_type": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.PipeTargetInvocationType](),
							},
						},
					},
				},
				"redshift_data_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 64),
							},
							"db_user": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
							"secret_manager_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							"sqls": {
								Type:     schema.TypeSet,
								Required: true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(1, 100000),
								},
							},
							"statement_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 500),
							},
							"with_event": {
								Type:     schema.TypeBool,
								Optional: true,
							},
						},
					},
				},
				"sagemaker_pipeline_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sqs_queue_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"pipeline_parameter": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 200,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 256),
												validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*|(\$(\.[\w/_-]+(\[(\d+|\*)\])*)*)$`), ""),
											),
										},
										names.AttrValue: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 1024),
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
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.step_function_state_machine_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"message_deduplication_id": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 100),
							},
							"message_group_id": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 100),
							},
						},
					},
				},
				"step_function_state_machine_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_job_parameters",
						"target_parameters.0.cloudwatch_logs_parameters",
						"target_parameters.0.ecs_task_parameters",
						"target_parameters.0.eventbridge_event_bus_parameters",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream_parameters",
						"target_parameters.0.lambda_function_parameters",
						"target_parameters.0.redshift_data_parameters",
						"target_parameters.0.sagemaker_pipeline_parameters",
						"target_parameters.0.sqs_queue_parameters",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"invocation_type": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.PipeTargetInvocationType](),
							},
						},
					},
				},
			},
		},
	}
}

func expandPipeTargetParameters(tfMap map[string]interface{}) *types.PipeTargetParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetParameters{}

	if v, ok := tfMap["batch_job_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.BatchJobParameters = expandPipeTargetBatchJobParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["cloudwatch_logs_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CloudWatchLogsParameters = expandPipeTargetCloudWatchLogsParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["ecs_task_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EcsTaskParameters = expandPipeTargetECSTaskParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["eventbridge_event_bus_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EventBridgeEventBusParameters = expandPipeTargetEventBridgeEventBusParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["http_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.HttpParameters = expandPipeTargetHTTPParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["input_template"].(string); ok && v != "" {
		apiObject.InputTemplate = aws.String(v)
	}

	if v, ok := tfMap["kinesis_stream_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.KinesisStreamParameters = expandPipeTargetKinesisStreamParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["lambda_function_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.LambdaFunctionParameters = expandPipeTargetLambdaFunctionParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["redshift_data_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.RedshiftDataParameters = expandPipeTargetRedshiftDataParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sagemaker_pipeline_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SageMakerPipelineParameters = expandPipeTargetSageMakerPipelineParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sqs_queue_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SqsQueueParameters = expandPipeTargetSQSQueueParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["step_function_state_machine_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.StepFunctionStateMachineParameters = expandPipeTargetStateMachineParameters(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPipeTargetBatchJobParameters(tfMap map[string]interface{}) *types.PipeTargetBatchJobParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetBatchJobParameters{}

	if v, ok := tfMap["array_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ArrayProperties = expandBatchArrayProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["container_overrides"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ContainerOverrides = expandBatchContainerOverrides(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["depends_on"].([]interface{}); ok && len(v) > 0 {
		apiObject.DependsOn = expandBatchJobDependencies(v)
	}

	if v, ok := tfMap["job_definition"].(string); ok && v != "" {
		apiObject.JobDefinition = aws.String(v)
	}

	if v, ok := tfMap["job_name"].(string); ok && v != "" {
		apiObject.JobName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrParameters].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Parameters = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["retry_strategy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.RetryStrategy = expandBatchRetryStrategy(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandBatchArrayProperties(tfMap map[string]interface{}) *types.BatchArrayProperties {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BatchArrayProperties{}

	if v, ok := tfMap[names.AttrSize].(int); ok {
		apiObject.Size = aws.Int32(int32(v))
	}

	return apiObject
}

func expandBatchContainerOverrides(tfMap map[string]interface{}) *types.BatchContainerOverrides {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BatchContainerOverrides{}

	if v, ok := tfMap["command"].([]interface{}); ok && len(v) > 0 {
		apiObject.Command = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap[names.AttrEnvironment].([]interface{}); ok && len(v) > 0 {
		apiObject.Environment = expandBatchEnvironmentVariables(v)
	}

	if v, ok := tfMap[names.AttrInstanceType].(string); ok && v != "" {
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := tfMap["resource_requirement"].([]interface{}); ok && len(v) > 0 {
		apiObject.ResourceRequirements = expandBatchResourceRequirements(v)
	}

	return apiObject
}

func expandBatchEnvironmentVariable(tfMap map[string]interface{}) *types.BatchEnvironmentVariable {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BatchEnvironmentVariable{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandBatchEnvironmentVariables(tfList []interface{}) []types.BatchEnvironmentVariable {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.BatchEnvironmentVariable

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandBatchEnvironmentVariable(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandBatchResourceRequirement(tfMap map[string]interface{}) *types.BatchResourceRequirement {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BatchResourceRequirement{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.BatchResourceRequirementType(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandBatchResourceRequirements(tfList []interface{}) []types.BatchResourceRequirement {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.BatchResourceRequirement

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandBatchResourceRequirement(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandBatchJobDependency(tfMap map[string]interface{}) *types.BatchJobDependency {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BatchJobDependency{}

	if v, ok := tfMap["job_id"].(string); ok && v != "" {
		apiObject.JobId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.BatchJobDependencyType(v)
	}

	return apiObject
}

func expandBatchJobDependencies(tfList []interface{}) []types.BatchJobDependency {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.BatchJobDependency

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandBatchJobDependency(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandBatchRetryStrategy(tfMap map[string]interface{}) *types.BatchRetryStrategy {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BatchRetryStrategy{}

	if v, ok := tfMap["attempts"].(int); ok {
		apiObject.Attempts = aws.Int32(int32(v))
	}

	return apiObject
}

func expandPipeTargetCloudWatchLogsParameters(tfMap map[string]interface{}) *types.PipeTargetCloudWatchLogsParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetCloudWatchLogsParameters{}

	if v, ok := tfMap["log_stream_name"].(string); ok && v != "" {
		apiObject.LogStreamName = aws.String(v)
	}

	if v, ok := tfMap["timestamp"].(string); ok && v != "" {
		apiObject.Timestamp = aws.String(v)
	}

	return apiObject
}

func expandPipeTargetECSTaskParameters(tfMap map[string]interface{}) *types.PipeTargetEcsTaskParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetEcsTaskParameters{}

	if v, ok := tfMap[names.AttrCapacityProviderStrategy].([]interface{}); ok && len(v) > 0 {
		apiObject.CapacityProviderStrategy = expandCapacityProviderStrategyItems(v)
	}

	if v, ok := tfMap["enable_ecs_managed_tags"].(bool); ok {
		apiObject.EnableECSManagedTags = v
	}

	if v, ok := tfMap["enable_execute_command"].(bool); ok {
		apiObject.EnableExecuteCommand = v
	}

	if v, ok := tfMap["group"].(string); ok && v != "" {
		apiObject.Group = aws.String(v)
	}

	if v, ok := tfMap["launch_type"].(string); ok && v != "" {
		apiObject.LaunchType = types.LaunchType(v)
	}

	if v, ok := tfMap[names.AttrNetworkConfiguration].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.NetworkConfiguration = expandNetworkConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["overrides"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Overrides = expandECSTaskOverride(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["placement_constraint"].([]interface{}); ok && len(v) > 0 {
		apiObject.PlacementConstraints = expandPlacementConstraints(v)
	}

	if v, ok := tfMap["placement_strategy"].([]interface{}); ok && len(v) > 0 {
		apiObject.PlacementStrategy = expandPlacementStrategies(v)
	}

	if v, ok := tfMap["platform_version"].(string); ok && v != "" {
		apiObject.PlatformVersion = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPropagateTags].(string); ok && v != "" {
		apiObject.PropagateTags = types.PropagateTags(v)
	}

	if v, ok := tfMap["reference_id"].(string); ok && v != "" {
		apiObject.ReferenceId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTags].(map[string]interface{}); ok && len(v) > 0 {
		for k, v := range flex.ExpandStringValueMap(v) {
			apiObject.Tags = append(apiObject.Tags, types.Tag{Key: aws.String(k), Value: aws.String(v)})
		}
	}

	if v, ok := tfMap["task_count"].(int); ok {
		apiObject.TaskCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["task_definition_arn"].(string); ok && v != "" {
		apiObject.TaskDefinitionArn = aws.String(v)
	}

	return apiObject
}

func expandCapacityProviderStrategyItem(tfMap map[string]interface{}) *types.CapacityProviderStrategyItem {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CapacityProviderStrategyItem{}

	if v, ok := tfMap["base"].(int); ok {
		apiObject.Base = int32(v)
	}

	if v, ok := tfMap["capacity_provider"].(string); ok && v != "" {
		apiObject.CapacityProvider = aws.String(v)
	}

	if v, ok := tfMap[names.AttrWeight].(int); ok {
		apiObject.Weight = int32(v)
	}

	return apiObject
}

func expandCapacityProviderStrategyItems(tfList []interface{}) []types.CapacityProviderStrategyItem {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.CapacityProviderStrategyItem

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCapacityProviderStrategyItem(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandNetworkConfiguration(tfMap map[string]interface{}) *types.NetworkConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.NetworkConfiguration{}

	if v, ok := tfMap["aws_vpc_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AwsvpcConfiguration = expandVPCConfiguration(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandVPCConfiguration(tfMap map[string]interface{}) *types.AwsVpcConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AwsVpcConfiguration{}

	if v, ok := tfMap["assign_public_ip"].(string); ok && v != "" {
		apiObject.AssignPublicIp = types.AssignPublicIp(v)
	}

	if v, ok := tfMap[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnets].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandECSTaskOverride(tfMap map[string]interface{}) *types.EcsTaskOverride {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EcsTaskOverride{}

	if v, ok := tfMap["container_override"].([]interface{}); ok && len(v) > 0 {
		apiObject.ContainerOverrides = expandECSContainerOverrides(v)
	}

	if v, ok := tfMap["cpu"].(string); ok && v != "" {
		apiObject.Cpu = aws.String(v)
	}

	if v, ok := tfMap["ephemeral_storage"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EphemeralStorage = expandECSEphemeralStorage(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrExecutionRoleARN].(string); ok && v != "" {
		apiObject.ExecutionRoleArn = aws.String(v)
	}

	if v, ok := tfMap["inference_accelerator_override"].([]interface{}); ok && len(v) > 0 {
		apiObject.InferenceAcceleratorOverrides = expandECSInferenceAcceleratorOverrides(v)
	}

	if v, ok := tfMap["memory"].(string); ok && v != "" {
		apiObject.Memory = aws.String(v)
	}

	if v, ok := tfMap["task_role_arn"].(string); ok && v != "" {
		apiObject.TaskRoleArn = aws.String(v)
	}

	return apiObject
}

func expandECSContainerOverride(tfMap map[string]interface{}) *types.EcsContainerOverride {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EcsContainerOverride{}

	if v, ok := tfMap["command"].([]interface{}); ok && len(v) > 0 {
		apiObject.Command = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["cpu"].(int); ok {
		apiObject.Cpu = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrEnvironment].([]interface{}); ok && len(v) > 0 {
		apiObject.Environment = expandECSEnvironmentVariables(v)
	}

	if v, ok := tfMap["environment_file"].([]interface{}); ok && len(v) > 0 {
		apiObject.EnvironmentFiles = expandECSEnvironmentFiles(v)
	}

	if v, ok := tfMap["memory"].(int); ok {
		apiObject.Memory = aws.Int32(int32(v))
	}

	if v, ok := tfMap["memory_reservation"].(int); ok {
		apiObject.MemoryReservation = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["resource_requirement"].([]interface{}); ok && len(v) > 0 {
		apiObject.ResourceRequirements = expandECSResourceRequirements(v)
	}

	return apiObject
}

func expandECSContainerOverrides(tfList []interface{}) []types.EcsContainerOverride {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.EcsContainerOverride

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandECSContainerOverride(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandECSEnvironmentVariable(tfMap map[string]interface{}) *types.EcsEnvironmentVariable {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EcsEnvironmentVariable{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandECSEnvironmentVariables(tfList []interface{}) []types.EcsEnvironmentVariable {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.EcsEnvironmentVariable

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandECSEnvironmentVariable(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandECSEnvironmentFile(tfMap map[string]interface{}) *types.EcsEnvironmentFile {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EcsEnvironmentFile{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.EcsEnvironmentFileType(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandECSEnvironmentFiles(tfList []interface{}) []types.EcsEnvironmentFile {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.EcsEnvironmentFile

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandECSEnvironmentFile(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandECSResourceRequirement(tfMap map[string]interface{}) *types.EcsResourceRequirement {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EcsResourceRequirement{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.EcsResourceRequirementType(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandECSResourceRequirements(tfList []interface{}) []types.EcsResourceRequirement {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.EcsResourceRequirement

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandECSResourceRequirement(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandECSEphemeralStorage(tfMap map[string]interface{}) *types.EcsEphemeralStorage {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EcsEphemeralStorage{}

	if v, ok := tfMap["size_in_gib"].(int); ok {
		apiObject.SizeInGiB = aws.Int32(int32(v))
	}

	return apiObject
}

func expandECSInferenceAcceleratorOverride(tfMap map[string]interface{}) *types.EcsInferenceAcceleratorOverride {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EcsInferenceAcceleratorOverride{}

	if v, ok := tfMap[names.AttrDeviceName].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["device_type"].(string); ok && v != "" {
		apiObject.DeviceType = aws.String(v)
	}

	return apiObject
}

func expandECSInferenceAcceleratorOverrides(tfList []interface{}) []types.EcsInferenceAcceleratorOverride {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.EcsInferenceAcceleratorOverride

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandECSInferenceAcceleratorOverride(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPlacementConstraint(tfMap map[string]interface{}) *types.PlacementConstraint {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PlacementConstraint{}

	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.PlacementConstraintType(v)
	}

	return apiObject
}

func expandPlacementConstraints(tfList []interface{}) []types.PlacementConstraint {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.PlacementConstraint

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPlacementConstraint(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPlacementStrategy(tfMap map[string]interface{}) *types.PlacementStrategy {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PlacementStrategy{}

	if v, ok := tfMap[names.AttrField].(string); ok && v != "" {
		apiObject.Field = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.PlacementStrategyType(v)
	}

	return apiObject
}

func expandPlacementStrategies(tfList []interface{}) []types.PlacementStrategy {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.PlacementStrategy

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPlacementStrategy(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPipeTargetEventBridgeEventBusParameters(tfMap map[string]interface{}) *types.PipeTargetEventBridgeEventBusParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetEventBridgeEventBusParameters{}

	if v, ok := tfMap["detail_type"].(string); ok && v != "" {
		apiObject.DetailType = aws.String(v)
	}

	if v, ok := tfMap["endpoint_id"].(string); ok && v != "" {
		apiObject.EndpointId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrResources].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Resources = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSource].(string); ok && v != "" {
		apiObject.Source = aws.String(v)
	}

	if v, ok := tfMap["time"].(string); ok && v != "" {
		apiObject.Time = aws.String(v)
	}

	return apiObject
}

func expandPipeTargetHTTPParameters(tfMap map[string]interface{}) *types.PipeTargetHttpParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetHttpParameters{}

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

func expandPipeTargetKinesisStreamParameters(tfMap map[string]interface{}) *types.PipeTargetKinesisStreamParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetKinesisStreamParameters{}

	if v, ok := tfMap["partition_key"].(string); ok && v != "" {
		apiObject.PartitionKey = aws.String(v)
	}

	return apiObject
}

func expandPipeTargetLambdaFunctionParameters(tfMap map[string]interface{}) *types.PipeTargetLambdaFunctionParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetLambdaFunctionParameters{}

	if v, ok := tfMap["invocation_type"].(string); ok && v != "" {
		apiObject.InvocationType = types.PipeTargetInvocationType(v)
	}

	return apiObject
}

func expandPipeTargetRedshiftDataParameters(tfMap map[string]interface{}) *types.PipeTargetRedshiftDataParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetRedshiftDataParameters{}

	if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
		apiObject.Database = aws.String(v)
	}

	if v, ok := tfMap["db_user"].(string); ok && v != "" {
		apiObject.DbUser = aws.String(v)
	}

	if v, ok := tfMap["secret_manager_arn"].(string); ok && v != "" {
		apiObject.SecretManagerArn = aws.String(v)
	}

	if v, ok := tfMap["sqls"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Sqls = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["statement_name"].(string); ok && v != "" {
		apiObject.StatementName = aws.String(v)
	}

	if v, ok := tfMap["with_event"].(bool); ok {
		apiObject.WithEvent = v
	}

	return apiObject
}

func expandPipeTargetSageMakerPipelineParameters(tfMap map[string]interface{}) *types.PipeTargetSageMakerPipelineParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetSageMakerPipelineParameters{}

	if v, ok := tfMap["pipeline_parameter"].([]interface{}); ok && len(v) > 0 {
		apiObject.PipelineParameterList = expandSageMakerPipelineParameters(v)
	}

	return apiObject
}

func expandSageMakerPipelineParameter(tfMap map[string]interface{}) *types.SageMakerPipelineParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SageMakerPipelineParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandSageMakerPipelineParameters(tfList []interface{}) []types.SageMakerPipelineParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.SageMakerPipelineParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandSageMakerPipelineParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPipeTargetSQSQueueParameters(tfMap map[string]interface{}) *types.PipeTargetSqsQueueParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetSqsQueueParameters{}

	if v, ok := tfMap["message_deduplication_id"].(string); ok && v != "" {
		apiObject.MessageDeduplicationId = aws.String(v)
	}

	if v, ok := tfMap["message_group_id"].(string); ok && v != "" {
		apiObject.MessageGroupId = aws.String(v)
	}

	return apiObject
}

func expandPipeTargetStateMachineParameters(tfMap map[string]interface{}) *types.PipeTargetStateMachineParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipeTargetStateMachineParameters{}

	if v, ok := tfMap["invocation_type"].(string); ok && v != "" {
		apiObject.InvocationType = types.PipeTargetInvocationType(v)
	}

	return apiObject
}

func flattenPipeTargetParameters(apiObject *types.PipeTargetParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BatchJobParameters; v != nil {
		tfMap["batch_job_parameters"] = []interface{}{flattenPipeTargetBatchJobParameters(v)}
	}

	if v := apiObject.CloudWatchLogsParameters; v != nil {
		tfMap["cloudwatch_logs_parameters"] = []interface{}{flattenPipeTargetCloudWatchLogsParameters(v)}
	}

	if v := apiObject.EcsTaskParameters; v != nil {
		tfMap["ecs_task_parameters"] = []interface{}{flattenPipeTargetECSTaskParameters(v)}
	}

	if v := apiObject.EventBridgeEventBusParameters; v != nil {
		tfMap["eventbridge_event_bus_parameters"] = []interface{}{flattenPipeTargetEventBridgeEventBusParameters(v)}
	}

	if v := apiObject.HttpParameters; v != nil {
		tfMap["http_parameters"] = []interface{}{flattenPipeTargetHTTPParameters(v)}
	}

	if v := apiObject.InputTemplate; v != nil {
		tfMap["input_template"] = aws.ToString(v)
	}

	if v := apiObject.KinesisStreamParameters; v != nil {
		tfMap["kinesis_stream_parameters"] = []interface{}{flattenPipeTargetKinesisStreamParameters(v)}
	}

	if v := apiObject.LambdaFunctionParameters; v != nil {
		tfMap["lambda_function_parameters"] = []interface{}{flattenPipeTargetLambdaFunctionParameters(v)}
	}

	if v := apiObject.RedshiftDataParameters; v != nil {
		tfMap["redshift_data_parameters"] = []interface{}{flattenPipeTargetRedshiftDataParameters(v)}
	}

	if v := apiObject.SageMakerPipelineParameters; v != nil {
		tfMap["sagemaker_pipeline_parameters"] = []interface{}{flattenPipeTargetSageMakerPipelineParameters(v)}
	}

	if v := apiObject.SqsQueueParameters; v != nil {
		tfMap["sqs_queue_parameters"] = []interface{}{flattenPipeTargetSQSQueueParameters(v)}
	}

	if v := apiObject.StepFunctionStateMachineParameters; v != nil {
		tfMap["step_function_state_machine_parameters"] = []interface{}{flattenPipeTargetStateMachineParameters(v)}
	}

	return tfMap
}

func flattenPipeTargetBatchJobParameters(apiObject *types.PipeTargetBatchJobParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ArrayProperties; v != nil {
		tfMap["array_properties"] = []interface{}{flattenBatchArrayProperties(v)}
	}

	if v := apiObject.ContainerOverrides; v != nil {
		tfMap["container_overrides"] = []interface{}{flattenBatchContainerOverrides(v)}
	}

	if v := apiObject.DependsOn; v != nil {
		tfMap["depends_on"] = flattenBatchJobDependencies(v)
	}

	if v := apiObject.JobDefinition; v != nil {
		tfMap["job_definition"] = aws.ToString(v)
	}

	if v := apiObject.JobName; v != nil {
		tfMap["job_name"] = aws.ToString(v)
	}

	if v := apiObject.Parameters; v != nil {
		tfMap[names.AttrParameters] = v
	}

	if v := apiObject.RetryStrategy; v != nil {
		tfMap["retry_strategy"] = []interface{}{flattenBatchRetryStrategy(v)}
	}

	return tfMap
}

func flattenBatchArrayProperties(apiObject *types.BatchArrayProperties) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Size; v != nil {
		tfMap[names.AttrSize] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenBatchContainerOverrides(apiObject *types.BatchContainerOverrides) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Command; v != nil {
		tfMap["command"] = v
	}

	if v := apiObject.Environment; v != nil {
		tfMap[names.AttrEnvironment] = flattenBatchEnvironmentVariables(v)
	}

	if v := apiObject.InstanceType; v != nil {
		tfMap[names.AttrInstanceType] = aws.ToString(v)
	}

	if v := apiObject.ResourceRequirements; v != nil {
		tfMap["resource_requirement"] = flattenBatchResourceRequirements(v)
	}

	return tfMap
}

func flattenBatchEnvironmentVariable(apiObject types.BatchEnvironmentVariable) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenBatchEnvironmentVariables(apiObjects []types.BatchEnvironmentVariable) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenBatchEnvironmentVariable(apiObject))
	}

	return tfList
}

func flattenBatchResourceRequirement(apiObject types.BatchResourceRequirement) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenBatchResourceRequirements(apiObjects []types.BatchResourceRequirement) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenBatchResourceRequirement(apiObject))
	}

	return tfList
}

func flattenBatchJobDependency(apiObject types.BatchJobDependency) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.JobId; v != nil {
		tfMap["job_id"] = aws.ToString(v)
	}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

	return tfMap
}

func flattenBatchJobDependencies(apiObjects []types.BatchJobDependency) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenBatchJobDependency(apiObject))
	}

	return tfList
}

func flattenBatchRetryStrategy(apiObject *types.BatchRetryStrategy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Attempts; v != nil {
		tfMap["attempts"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenPipeTargetCloudWatchLogsParameters(apiObject *types.PipeTargetCloudWatchLogsParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogStreamName; v != nil {
		tfMap["log_stream_name"] = aws.ToString(v)
	}

	if v := apiObject.Timestamp; v != nil {
		tfMap["timestamp"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPipeTargetECSTaskParameters(apiObject *types.PipeTargetEcsTaskParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"enable_ecs_managed_tags": apiObject.EnableECSManagedTags,
		"enable_execute_command":  apiObject.EnableExecuteCommand,
	}

	if v := apiObject.CapacityProviderStrategy; v != nil {
		tfMap[names.AttrCapacityProviderStrategy] = flattenCapacityProviderStrategyItems(v)
	}

	if v := apiObject.Group; v != nil {
		tfMap["group"] = aws.ToString(v)
	}

	if v := apiObject.LaunchType; v != "" {
		tfMap["launch_type"] = v
	}

	if v := apiObject.NetworkConfiguration; v != nil {
		tfMap[names.AttrNetworkConfiguration] = []interface{}{flattenNetworkConfiguration(v)}
	}

	if v := apiObject.Overrides; v != nil {
		tfMap["overrides"] = []interface{}{flattenECSTaskOverride(v)}
	}

	if v := apiObject.PlacementConstraints; v != nil {
		tfMap["placement_constraint"] = flattenPlacementConstraints(v)
	}

	if v := apiObject.PlacementStrategy; v != nil {
		tfMap["placement_strategy"] = flattenPlacementStrategies(v)
	}

	if v := apiObject.PlatformVersion; v != nil {
		tfMap["platform_version"] = aws.ToString(v)
	}

	if v := apiObject.PropagateTags; v != "" {
		tfMap[names.AttrPropagateTags] = v
	}

	if v := apiObject.ReferenceId; v != nil {
		tfMap["reference_id"] = aws.ToString(v)
	}

	if v := apiObject.Tags; v != nil {
		tags := map[string]interface{}{}

		for _, apiObject := range v {
			tags[aws.ToString(apiObject.Key)] = aws.ToString(apiObject.Value)
		}

		tfMap[names.AttrTags] = tags
	}

	if v := apiObject.TaskCount; v != nil {
		tfMap["task_count"] = aws.ToInt32(v)
	}

	if v := apiObject.TaskDefinitionArn; v != nil {
		tfMap["task_definition_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenCapacityProviderStrategyItem(apiObject types.CapacityProviderStrategyItem) map[string]interface{} {
	tfMap := map[string]interface{}{
		"base":           apiObject.Base,
		names.AttrWeight: apiObject.Weight,
	}

	if v := apiObject.CapacityProvider; v != nil {
		tfMap["capacity_provider"] = aws.ToString(v)
	}

	return tfMap
}

func flattenCapacityProviderStrategyItems(apiObjects []types.CapacityProviderStrategyItem) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCapacityProviderStrategyItem(apiObject))
	}

	return tfList
}

func flattenECSTaskOverride(apiObject *types.EcsTaskOverride) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ContainerOverrides; v != nil {
		tfMap["container_override"] = flattenECSContainerOverrides(v)
	}

	if v := apiObject.Cpu; v != nil {
		tfMap["cpu"] = aws.ToString(v)
	}

	if v := apiObject.EphemeralStorage; v != nil {
		tfMap["ephemeral_storage"] = []interface{}{flattenECSEphemeralStorage(v)}
	}

	if v := apiObject.ExecutionRoleArn; v != nil {
		tfMap[names.AttrExecutionRoleARN] = aws.ToString(v)
	}

	if v := apiObject.InferenceAcceleratorOverrides; v != nil {
		tfMap["inference_accelerator_override"] = flattenECSInferenceAcceleratorOverrides(v)
	}

	if v := apiObject.Memory; v != nil {
		tfMap["memory"] = aws.ToString(v)
	}

	if v := apiObject.TaskRoleArn; v != nil {
		tfMap["task_role_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenECSContainerOverride(apiObject types.EcsContainerOverride) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Command; v != nil {
		tfMap["command"] = v
	}

	if v := apiObject.Cpu; v != nil {
		tfMap["cpu"] = aws.ToInt32(v)
	}

	if v := apiObject.Environment; v != nil {
		tfMap[names.AttrEnvironment] = flattenECSEnvironmentVariables(v)
	}

	if v := apiObject.EnvironmentFiles; v != nil {
		tfMap["environment_file"] = flattenECSEnvironmentFiles(v)
	}

	if v := apiObject.Memory; v != nil {
		tfMap["memory"] = aws.ToInt32(v)
	}

	if v := apiObject.MemoryReservation; v != nil {
		tfMap["memory_reservation"] = aws.ToInt32(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.ResourceRequirements; v != nil {
		tfMap["resource_requirement"] = flattenECSResourceRequirements(v)
	}

	return tfMap
}

func flattenECSContainerOverrides(apiObjects []types.EcsContainerOverride) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenECSContainerOverride(apiObject))
	}

	return tfList
}

func flattenECSResourceRequirement(apiObject types.EcsResourceRequirement) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrName] = v
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenECSResourceRequirements(apiObjects []types.EcsResourceRequirement) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenECSResourceRequirement(apiObject))
	}

	return tfList
}

func flattenECSEnvironmentFile(apiObject types.EcsEnvironmentFile) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrName] = v
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenECSEnvironmentVariable(apiObject types.EcsEnvironmentVariable) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenECSEnvironmentVariables(apiObjects []types.EcsEnvironmentVariable) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenECSEnvironmentVariable(apiObject))
	}

	return tfList
}

func flattenECSEnvironmentFiles(apiObjects []types.EcsEnvironmentFile) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenECSEnvironmentFile(apiObject))
	}

	return tfList
}

func flattenECSEphemeralStorage(apiObject *types.EcsEphemeralStorage) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"size_in_gib": apiObject.SizeInGiB,
	}

	return tfMap
}

func flattenECSInferenceAcceleratorOverride(apiObject types.EcsInferenceAcceleratorOverride) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.DeviceName; v != nil {
		tfMap[names.AttrDeviceName] = aws.ToString(v)
	}

	if v := apiObject.DeviceType; v != nil {
		tfMap["device_type"] = aws.ToString(v)
	}

	return tfMap
}

func flattenECSInferenceAcceleratorOverrides(apiObjects []types.EcsInferenceAcceleratorOverride) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenECSInferenceAcceleratorOverride(apiObject))
	}

	return tfList
}

func flattenNetworkConfiguration(apiObject *types.NetworkConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AwsvpcConfiguration; v != nil {
		tfMap["aws_vpc_configuration"] = []interface{}{flattenVPCConfiguration(v)}
	}

	return tfMap
}

func flattenVPCConfiguration(apiObject *types.AwsVpcConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AssignPublicIp; v != "" {
		tfMap["assign_public_ip"] = v
	}

	if v := apiObject.SecurityGroups; v != nil {
		tfMap[names.AttrSecurityGroups] = v
	}

	if v := apiObject.Subnets; v != nil {
		tfMap[names.AttrSubnets] = v
	}

	return tfMap
}

func flattenPlacementConstraint(apiObject types.PlacementConstraint) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Expression; v != nil {
		tfMap[names.AttrExpression] = aws.ToString(v)
	}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

	return tfMap
}

func flattenPlacementConstraints(apiObjects []types.PlacementConstraint) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPlacementConstraint(apiObject))
	}

	return tfList
}

func flattenPlacementStrategy(apiObject types.PlacementStrategy) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Field; v != nil {
		tfMap[names.AttrField] = aws.ToString(v)
	}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

	return tfMap
}

func flattenPlacementStrategies(apiObjects []types.PlacementStrategy) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPlacementStrategy(apiObject))
	}

	return tfList
}

func flattenPipeTargetEventBridgeEventBusParameters(apiObject *types.PipeTargetEventBridgeEventBusParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DetailType; v != nil {
		tfMap["detail_type"] = aws.ToString(v)
	}

	if v := apiObject.EndpointId; v != nil {
		tfMap["endpoint_id"] = aws.ToString(v)
	}

	if v := apiObject.Resources; v != nil {
		tfMap[names.AttrResources] = v
	}

	if v := apiObject.Source; v != nil {
		tfMap[names.AttrSource] = aws.ToString(v)
	}

	if v := apiObject.Time; v != nil {
		tfMap["time"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPipeTargetHTTPParameters(apiObject *types.PipeTargetHttpParameters) map[string]interface{} {
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

func flattenPipeTargetKinesisStreamParameters(apiObject *types.PipeTargetKinesisStreamParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PartitionKey; v != nil {
		tfMap["partition_key"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPipeTargetLambdaFunctionParameters(apiObject *types.PipeTargetLambdaFunctionParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.InvocationType; v != "" {
		tfMap["invocation_type"] = v
	}

	return tfMap
}

func flattenPipeTargetRedshiftDataParameters(apiObject *types.PipeTargetRedshiftDataParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"with_event": apiObject.WithEvent,
	}

	if v := apiObject.Database; v != nil {
		tfMap[names.AttrDatabase] = aws.ToString(v)
	}

	if v := apiObject.DbUser; v != nil {
		tfMap["db_user"] = aws.ToString(v)
	}

	if v := apiObject.SecretManagerArn; v != nil {
		tfMap["secret_manager_arn"] = aws.ToString(v)
	}

	if v := apiObject.Sqls; v != nil {
		tfMap["sqls"] = v
	}

	if v := apiObject.StatementName; v != nil {
		tfMap["statement_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPipeTargetSageMakerPipelineParameters(apiObject *types.PipeTargetSageMakerPipelineParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PipelineParameterList; v != nil {
		tfMap["pipeline_parameter"] = flattenSageMakerPipelineParameters(v)
	}

	return tfMap
}

func flattenSageMakerPipelineParameter(apiObject types.SageMakerPipelineParameter) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenSageMakerPipelineParameters(apiObjects []types.SageMakerPipelineParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenSageMakerPipelineParameter(apiObject))
	}

	return tfList
}

func flattenPipeTargetSQSQueueParameters(apiObject *types.PipeTargetSqsQueueParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MessageDeduplicationId; v != nil {
		tfMap["message_deduplication_id"] = aws.ToString(v)
	}

	if v := apiObject.MessageGroupId; v != nil {
		tfMap["message_group_id"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPipeTargetStateMachineParameters(apiObject *types.PipeTargetStateMachineParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.InvocationType; v != "" {
		tfMap["invocation_type"] = v
	}

	return tfMap
}
