package pipes

import (
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func targetParametersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"batch_target": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"array_properties": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"size": {
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
										"environment": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"value": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										"instance_type": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"resource_requirements": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"type": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[types.BatchResourceRequirementType](),
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
										"type": {
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
							"parameters": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"key": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"value": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
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
				"cloudwatch_logs": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
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
									validation.StringMatch(regexp.MustCompile(`^\$(\.[\w/_-]+(\[(\d+|\*)\])*)*$`), ""),
								),
							},
						},
					},
				},
				"ecs_task": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"capacity_provider_strategy": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 6,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"capacity_provider": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
										"base": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(0, 100000),
											Default:      0,
										},
										"weight": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(0, 1000),
											Default:      0,
										},
									},
								},
							},
							"enable_ecs_managed_tags": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"enable_execute_command": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
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
							"network_configuration": {
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
													"security_groups": {
														Type:     schema.TypeSet,
														Optional: true,
														MaxItems: 5,
														Elem: &schema.Schema{
															Type: schema.TypeString,
															ValidateFunc: validation.All(
																validation.StringLenBetween(1, 1024),
																validation.StringMatch(regexp.MustCompile(`^sg-[0-9a-zA-Z]*|(\$(\.[\w/_-]+(\[(\d+|\*)\])*)*)$`), ""),
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
																validation.StringMatch(regexp.MustCompile(`^subnet-[0-9a-z]*|(\$(\.[\w/_-]+(\[(\d+|\*)\])*)*)$`), ""),
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
										"container_overrides": {
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
													"environment": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"value": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
															},
														},
													},
													"environment_files": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"type": {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[types.EcsEnvironmentFileType](),
																},
																"value": {
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
													"name": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"resource_requirements": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"type": {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[types.EcsResourceRequirementType](),
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
										"cpu": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"ecs_ephemeral_storage": {
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
										"execution_role_arn": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verify.ValidARN,
										},
										"inference_accelerator_overrides": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"device_name": {
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
							"placement_constraints": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"expression": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(1, 2000),
										},
										"type": {
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
										"field": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
										"type": {
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
							"propagate_tags": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.PropagateTags](),
							},
							"reference_id": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 1024),
							},
							"tags": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"key": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										"value": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 256),
										},
									},
								},
							},
							"task_count": {
								Type:     schema.TypeInt,
								Optional: true,
								Default:  1,
							},
							"task_definition_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"event_bridge_event_bus": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
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
									validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9\-]+[\.][A-Za-z0-9\-]+$`), ""),
								),
							},
							"resources": {
								Type:     schema.TypeSet,
								Optional: true,
								MaxItems: 10,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: verify.ValidARN,
								},
							},
							"source": {
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
									validation.StringMatch(regexp.MustCompile(`^\$(\.[\w/_-]+(\[(\d+|\*)\])*)*$`), ""),
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
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
					},
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
				"kinesis_stream": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
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
				"lambda_function": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
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
				"redshift_data": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"database": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 64),
							},
							"database_user": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
							"secret_manager_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							"statement_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 500),
							},
							"sqls": {
								Type:     schema.TypeSet,
								Required: true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(1, 100000),
								},
							},
							"with_event": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
						},
					},
				},
				"sagemaker_pipeline": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sqs_queue",
						"target_parameters.0.step_function",
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"parameters": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 200,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 256),
												validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*|(\$(\.[\w/_-]+(\[(\d+|\*)\])*)*)$`), ""),
											),
										},
										"value": {
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
				"sqs_queue": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.step_function",
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
				"step_function": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ConflictsWith: []string{
						"target_parameters.0.batch_target",
						"target_parameters.0.cloudwatch_logs",
						"target_parameters.0.ecs_task",
						"target_parameters.0.event_bridge_event_bus",
						"target_parameters.0.http_parameters",
						"target_parameters.0.kinesis_stream",
						"target_parameters.0.lambda_function",
						"target_parameters.0.redshift_data",
						"target_parameters.0.sagemaker_pipeline",
						"target_parameters.0.sqs_queue",
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

	// ... nested attribute handling ...

	return apiObject
}

func flattenPipeTargetParameters(apiObject *types.PipeTargetParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	// ... nested attribute handling ...

	return tfMap
}

func expandTargetParameters(config []interface{}) *types.PipeTargetParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetParameters
	for _, c := range config {
		param, ok := c.(map[string]interface{})
		if !ok {
			return nil
		}

		if val, ok := param["batch_target"]; ok {
			parameters.BatchJobParameters = expandTargetBatchJobParameters(val.([]interface{}))
		}

		if val, ok := param["cloudwatch_logs"]; ok {
			parameters.CloudWatchLogsParameters = expandTargetCloudWatchLogsParameters(val.([]interface{}))
		}

		if val, ok := param["ecs_task"]; ok {
			parameters.EcsTaskParameters = expandTargetEcsTaskParameters(val.([]interface{}))
		}

		if val, ok := param["event_bridge_event_bus"]; ok {
			parameters.EventBridgeEventBusParameters = expandTargetEventBridgeEventBusParameters(val.([]interface{}))
		}

		if val, ok := param["http_parameters"]; ok {
			parameters.HttpParameters = expandTargetHTTPParameters(val.([]interface{}))
		}

		if val, ok := param["input_template"].(string); ok && val != "" {
			parameters.InputTemplate = aws.String(val)
		}

		if val, ok := param["kinesis_stream"]; ok {
			parameters.KinesisStreamParameters = expandTargetKinesisStreamParameters(val.([]interface{}))
		}

		if val, ok := param["lambda_function"]; ok {
			parameters.LambdaFunctionParameters = expandTargetLambdaFunctionParameters(val.([]interface{}))
		}

		if val, ok := param["redshift_data"]; ok {
			parameters.RedshiftDataParameters = expandTargetRedshiftDataParameters(val.([]interface{}))
		}

		if val, ok := param["sagemaker_pipeline"]; ok {
			parameters.SageMakerPipelineParameters = expandTargetSageMakerPipelineParameters(val.([]interface{}))
		}

		if val, ok := param["sqs_queue"]; ok {
			parameters.SqsQueueParameters = expandTargetSqsQueueParameters(val.([]interface{}))
		}

		if val, ok := param["step_function"]; ok {
			parameters.StepFunctionStateMachineParameters = expandTargetStepFunctionStateMachineParameters(val.([]interface{}))
		}
	}
	return &parameters
}

func expandTargetBatchJobParameters(config []interface{}) *types.PipeTargetBatchJobParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetBatchJobParameters
	for _, c := range config {
		param := c.(map[string]interface{})

		parameters.JobDefinition = expandString("job_definition", param)
		parameters.JobName = expandString("job_name", param)
		if val, ok := param["retry_strategy"]; ok {
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					if attempts, ok := valueParam["attempts"].(int32); ok {
						parameters.RetryStrategy = &types.BatchRetryStrategy{
							Attempts: attempts,
						}
					}
				}
			}
		}
		if val, ok := param["array_properties"]; ok {
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					if size, ok := valueParam["size"].(int32); ok {
						parameters.ArrayProperties = &types.BatchArrayProperties{
							Size: size,
						}
					}
				}
			}
		}

		if val, ok := param["parameters"]; ok {
			batchTargetParameters := map[string]string{}
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					if key, ok := valueParam["key"].(string); ok && key != "" {
						if value, ok := valueParam["value"].(string); ok && value != "" {
							batchTargetParameters[key] = value
						}
					}
				}
			}
			if len(batchTargetParameters) > 0 {
				parameters.Parameters = batchTargetParameters
			}
		}

		if val, ok := param["depends_on"]; ok {
			var dependsOn []types.BatchJobDependency
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					var dependency types.BatchJobDependency
					dependency.JobId = expandString("job_id", valueParam)
					dependancyType := expandStringValue("type", valueParam)
					if dependancyType != "" {
						dependency.Type = types.BatchJobDependencyType(dependancyType)
					}
					dependsOn = append(dependsOn, dependency)
				}
			}
			if len(dependsOn) > 0 {
				parameters.DependsOn = dependsOn
			}
		}

		if val, ok := param["container_overrides"]; ok {
			parameters.ContainerOverrides = expandTargetBatchContainerOverrides(val.([]interface{}))
		}
	}

	return &parameters
}

func expandTargetBatchContainerOverrides(config []interface{}) *types.BatchContainerOverrides {
	if len(config) == 0 {
		return nil
	}

	var parameters types.BatchContainerOverrides
	for _, c := range config {
		param := c.(map[string]interface{})
		if value, ok := param["command"]; ok {
			parameters.Command = flex.ExpandStringValueList(value.([]interface{}))
		}
		parameters.InstanceType = expandString("instance_type", param)

		if val, ok := param["environment"]; ok {
			var environment []types.BatchEnvironmentVariable
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					var env types.BatchEnvironmentVariable
					env.Name = expandString("name", valueParam)
					env.Value = expandString("value", valueParam)
					environment = append(environment, env)
				}
			}
			if len(environment) > 0 {
				parameters.Environment = environment
			}
		}

		if val, ok := param["resource_requirements"]; ok {
			var resourceRequirements []types.BatchResourceRequirement
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					var resourceRequirement types.BatchResourceRequirement
					resourceRequirementType := expandStringValue("type", valueParam)
					if resourceRequirementType != "" {
						resourceRequirement.Type = types.BatchResourceRequirementType(resourceRequirementType)
					}
					resourceRequirement.Value = expandString("value", valueParam)
					resourceRequirements = append(resourceRequirements, resourceRequirement)
				}
			}
			if len(resourceRequirements) > 0 {
				parameters.ResourceRequirements = resourceRequirements
			}
		}
	}

	return &parameters
}

func expandTargetCloudWatchLogsParameters(config []interface{}) *types.PipeTargetCloudWatchLogsParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetCloudWatchLogsParameters
	for _, c := range config {
		param := c.(map[string]interface{})

		parameters.LogStreamName = expandString("log_stream_name", param)
		parameters.Timestamp = expandString("timestamp", param)
	}

	return &parameters
}

func expandTargetEcsTaskParameters(config []interface{}) *types.PipeTargetEcsTaskParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetEcsTaskParameters
	for _, c := range config {
		param := c.(map[string]interface{})

		parameters.TaskDefinitionArn = expandString("task_definition_arn", param)
		parameters.EnableECSManagedTags = expandBool("enable_ecs_managed_tags", param)
		parameters.EnableExecuteCommand = expandBool("enable_execute_command", param)
		parameters.Group = expandString("group", param)
		launchType := expandStringValue("launch_type", param)
		if launchType != "" {
			parameters.LaunchType = types.LaunchType(launchType)
		}
		parameters.PlatformVersion = expandString("platform_version", param)
		propagateTags := expandStringValue("propagate_tags", param)
		if propagateTags != "" {
			parameters.PropagateTags = types.PropagateTags(propagateTags)
		}
		parameters.ReferenceId = expandString("reference_id", param)
		parameters.TaskCount = expandInt32("task_count", param)

		if val, ok := param["capacity_provider_strategy"]; ok {
			parameters.CapacityProviderStrategy = expandTargetCapacityProviderStrategy(val.([]interface{}))
		}
		if val, ok := param["network_configuration"]; ok {
			parameters.NetworkConfiguration = expandTargetNetworkConfiguration(val.([]interface{}))
		}
		if val, ok := param["placement_constraints"]; ok {
			parameters.PlacementConstraints = expandTargetPlacementConstraints(val.([]interface{}))
		}
		if val, ok := param["placement_strategy"]; ok {
			parameters.PlacementStrategy = expandTargetPlacementStrategies(val.([]interface{}))
		}
		if val, ok := param["tags"]; ok {
			parameters.Tags = expandTargetECSTaskTags(val.([]interface{}))
		}
		if val, ok := param["overrides"]; ok {
			parameters.Overrides = expandTargetECSTaskOverrides(val.([]interface{}))
		}
	}

	return &parameters
}

func expandTargetCapacityProviderStrategy(config []interface{}) []types.CapacityProviderStrategyItem {
	if len(config) == 0 {
		return nil
	}

	var parameters []types.CapacityProviderStrategyItem
	for _, c := range config {
		param := c.(map[string]interface{})

		var provider types.CapacityProviderStrategyItem
		provider.CapacityProvider = expandString("capacity_provider", param)
		base := expandInt32("base", param)
		if base != nil {
			provider.Base = aws.ToInt32(base)
		}
		weight := expandInt32("weight", param)
		if weight != nil {
			provider.Weight = aws.ToInt32(weight)
		}

		parameters = append(parameters, provider)
	}

	return parameters
}

func expandTargetNetworkConfiguration(config []interface{}) *types.NetworkConfiguration {
	if len(config) == 0 {
		return nil
	}

	var parameters types.NetworkConfiguration
	for _, c := range config {
		param := c.(map[string]interface{})

		if val, ok := param["aws_vpc_configuration"]; ok {
			parameters.AwsvpcConfiguration = expandTargetAWSVPCConfiguration(val.([]interface{}))
		}
	}

	return &parameters
}

func expandTargetAWSVPCConfiguration(config []interface{}) *types.AwsVpcConfiguration {
	if len(config) == 0 {
		return nil
	}

	var parameters types.AwsVpcConfiguration
	for _, c := range config {
		param := c.(map[string]interface{})
		assignPublicIp := expandStringValue("assign_public_ip", param)
		if assignPublicIp != "" {
			parameters.AssignPublicIp = types.AssignPublicIp(assignPublicIp)
		}

		if value, ok := param["security_groups"]; ok && value.(*schema.Set).Len() > 0 {
			parameters.SecurityGroups = flex.ExpandStringValueSet(value.(*schema.Set))
		}

		if value, ok := param["subnets"]; ok && value.(*schema.Set).Len() > 0 {
			parameters.Subnets = flex.ExpandStringValueSet(value.(*schema.Set))
		}
	}

	return &parameters
}

func expandTargetPlacementConstraints(config []interface{}) []types.PlacementConstraint {
	if len(config) == 0 {
		return nil
	}

	var parameters []types.PlacementConstraint
	for _, c := range config {
		param := c.(map[string]interface{})

		var constraint types.PlacementConstraint
		constraint.Expression = expandString("expression", param)
		constraintType := expandStringValue("type", param)
		if constraintType != "" {
			constraint.Type = types.PlacementConstraintType(constraintType)
		}

		parameters = append(parameters, constraint)
	}

	return parameters
}

func expandTargetPlacementStrategies(config []interface{}) []types.PlacementStrategy {
	if len(config) == 0 {
		return nil
	}

	var parameters []types.PlacementStrategy
	for _, c := range config {
		param := c.(map[string]interface{})

		var strategy types.PlacementStrategy
		strategy.Field = expandString("field", param)
		strategyType := expandStringValue("type", param)
		if strategyType != "" {
			strategy.Type = types.PlacementStrategyType(strategyType)
		}

		parameters = append(parameters, strategy)
	}

	return parameters
}

func expandTargetECSTaskTags(config []interface{}) []types.Tag {
	if len(config) == 0 {
		return nil
	}

	var parameters []types.Tag
	for _, c := range config {
		param := c.(map[string]interface{})

		var tag types.Tag
		tag.Key = expandString("key", param)
		tag.Value = expandString("value", param)

		parameters = append(parameters, tag)
	}

	return parameters
}

func expandTargetECSTaskOverrides(config []interface{}) *types.EcsTaskOverride {
	if len(config) == 0 {
		return nil
	}

	var parameters types.EcsTaskOverride
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.Cpu = expandString("cpu", param)
		parameters.Memory = expandString("memory", param)
		parameters.ExecutionRoleArn = expandString("execution_role_arn", param)
		parameters.TaskRoleArn = expandString("task_role_arn", param)

		if val, ok := param["inference_accelerator_overrides"]; ok {
			var inferenceAcceleratorOverrides []types.EcsInferenceAcceleratorOverride
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					var override types.EcsInferenceAcceleratorOverride
					override.DeviceName = expandString("device_name", valueParam)
					override.DeviceType = expandString("device_type", valueParam)
					inferenceAcceleratorOverrides = append(inferenceAcceleratorOverrides, override)
				}
			}
			if len(inferenceAcceleratorOverrides) > 0 {
				parameters.InferenceAcceleratorOverrides = inferenceAcceleratorOverrides
			}
		}

		if val, ok := param["ecs_ephemeral_storage"]; ok {
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					if size, ok := valueParam["size_in_gib"].(int32); ok {
						parameters.EphemeralStorage = &types.EcsEphemeralStorage{
							SizeInGiB: size,
						}
					}
				}
			}
		}

		if val, ok := param["container_overrides"]; ok {
			parameters.ContainerOverrides = expandTargetECSTaskOverrideContainerOverrides(val.([]interface{}))
		}
	}

	return &parameters
}

func expandTargetECSTaskOverrideContainerOverrides(config []interface{}) []types.EcsContainerOverride {
	if len(config) == 0 {
		return nil
	}

	var parameters []types.EcsContainerOverride
	for _, c := range config {
		param := c.(map[string]interface{})

		var override types.EcsContainerOverride
		override.Cpu = expandInt32("cpu", param)
		override.Memory = expandInt32("memory", param)
		override.MemoryReservation = expandInt32("memory_reservation", param)
		override.Name = expandString("name", param)
		if value, ok := param["command"]; ok {
			override.Command = flex.ExpandStringValueList(value.([]interface{}))
		}

		if val, ok := param["environment"]; ok {
			var environment []types.EcsEnvironmentVariable
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					var env types.EcsEnvironmentVariable
					env.Name = expandString("name", valueParam)
					env.Value = expandString("value", valueParam)
					environment = append(environment, env)
				}
			}
			if len(environment) > 0 {
				override.Environment = environment
			}
		}

		if val, ok := param["environment_files"]; ok {
			var environment []types.EcsEnvironmentFile
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					var env types.EcsEnvironmentFile
					envType := expandStringValue("type", valueParam)
					if envType != "" {
						env.Type = types.EcsEnvironmentFileType(envType)
					}
					env.Value = expandString("value", valueParam)
					environment = append(environment, env)
				}
			}
			if len(environment) > 0 {
				override.EnvironmentFiles = environment
			}
		}

		if val, ok := param["resource_requirements"]; ok {
			var resourceRequirements []types.EcsResourceRequirement
			if values, ok := val.([]interface{}); ok {
				for _, v := range values {
					valueParam := v.(map[string]interface{})

					var resourceRequirement types.EcsResourceRequirement
					resourceRequirementType := expandStringValue("type", valueParam)
					if resourceRequirementType != "" {
						resourceRequirement.Type = types.EcsResourceRequirementType(resourceRequirementType)
					}
					resourceRequirement.Value = expandString("value", valueParam)
					resourceRequirements = append(resourceRequirements, resourceRequirement)
				}
			}
			if len(resourceRequirements) > 0 {
				override.ResourceRequirements = resourceRequirements
			}
		}

		parameters = append(parameters, override)
	}

	return parameters
}

func expandTargetEventBridgeEventBusParameters(config []interface{}) *types.PipeTargetEventBridgeEventBusParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetEventBridgeEventBusParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.DetailType = expandString("detail_type", param)
		parameters.EndpointId = expandString("endpoint_id", param)
		parameters.Source = expandString("source", param)
		parameters.Time = expandString("time", param)
		if value, ok := param["resources"]; ok && value.(*schema.Set).Len() > 0 {
			parameters.Resources = flex.ExpandStringValueSet(value.(*schema.Set))
		}
	}

	return &parameters
}

func expandTargetHTTPParameters(config []interface{}) *types.PipeTargetHttpParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetHttpParameters
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

func expandTargetKinesisStreamParameters(config []interface{}) *types.PipeTargetKinesisStreamParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetKinesisStreamParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.PartitionKey = expandString("partition_key", param)
	}

	return &parameters
}

func expandTargetLambdaFunctionParameters(config []interface{}) *types.PipeTargetLambdaFunctionParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetLambdaFunctionParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		invocationType := expandStringValue("invocation_type", param)
		if invocationType != "" {
			parameters.InvocationType = types.PipeTargetInvocationType(invocationType)
		}
	}

	return &parameters
}

func expandTargetRedshiftDataParameters(config []interface{}) *types.PipeTargetRedshiftDataParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetRedshiftDataParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.Database = expandString("database", param)
		parameters.DbUser = expandString("database_user", param)
		parameters.SecretManagerArn = expandString("secret_manager_arn", param)
		parameters.StatementName = expandString("statement_name", param)
		parameters.WithEvent = expandBool("with_event", param)
		if value, ok := param["sqls"]; ok && value.(*schema.Set).Len() > 0 {
			parameters.Sqls = flex.ExpandStringValueSet(value.(*schema.Set))
		}
	}

	return &parameters
}

func expandTargetSageMakerPipelineParameters(config []interface{}) *types.PipeTargetSageMakerPipelineParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetSageMakerPipelineParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["parameters"]; ok {
			parametersConfig := val.([]interface{})
			var params []types.SageMakerPipelineParameter
			for _, p := range parametersConfig {
				pp := p.(map[string]interface{})
				name := expandString("name", pp)
				value := expandString("value", pp)
				if name != nil {
					params = append(params, types.SageMakerPipelineParameter{
						Name:  name,
						Value: value,
					})
				}
			}
			parameters.PipelineParameterList = params
		}
	}

	return &parameters
}

func expandTargetSqsQueueParameters(config []interface{}) *types.PipeTargetSqsQueueParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetSqsQueueParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		parameters.MessageDeduplicationId = expandString("message_deduplication_id", param)
		parameters.MessageGroupId = expandString("message_group_id", param)
	}

	return &parameters
}

func expandTargetStepFunctionStateMachineParameters(config []interface{}) *types.PipeTargetStateMachineParameters {
	if len(config) == 0 {
		return nil
	}

	var parameters types.PipeTargetStateMachineParameters
	for _, c := range config {
		param := c.(map[string]interface{})
		invocationType := expandStringValue("invocation_type", param)
		if invocationType != "" {
			parameters.InvocationType = types.PipeTargetInvocationType(invocationType)
		}
	}

	return &parameters
}

func flattenTargetParameters(targetParameters *types.PipeTargetParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if targetParameters.BatchJobParameters != nil {
		config["batch_target"] = flattenTargetBatchJobParameters(targetParameters.BatchJobParameters)
	}

	if targetParameters.CloudWatchLogsParameters != nil {
		config["cloudwatch_logs"] = flattenTargetCloudWatchLogsParameters(targetParameters.CloudWatchLogsParameters)
	}

	if targetParameters.EcsTaskParameters != nil {
		config["ecs_task"] = flattenTargetEcsTaskParameters(targetParameters.EcsTaskParameters)
	}

	if targetParameters.EventBridgeEventBusParameters != nil {
		config["event_bridge_event_bus"] = flattenTargetEventBridgeEventBusParameters(targetParameters.EventBridgeEventBusParameters)
	}

	if targetParameters.HttpParameters != nil {
		config["http_parameters"] = flattenTargetHttpParameters(targetParameters.HttpParameters)
	}

	if targetParameters.InputTemplate != nil {
		config["input_template"] = aws.ToString(targetParameters.InputTemplate)
	}

	if targetParameters.KinesisStreamParameters != nil {
		config["kinesis_stream"] = flattenTargetKinesisStreamParameters(targetParameters.KinesisStreamParameters)
	}

	if targetParameters.LambdaFunctionParameters != nil {
		config["lambda_function"] = flattenTargetLambdaFunctionParameters(targetParameters.LambdaFunctionParameters)
	}

	if targetParameters.RedshiftDataParameters != nil {
		config["redshift_data"] = flattenTargetRedshiftDataParameters(targetParameters.RedshiftDataParameters)
	}

	if targetParameters.SageMakerPipelineParameters != nil {
		config["sagemaker_pipeline"] = flattenTargetSageMakerPipelineParameters(targetParameters.SageMakerPipelineParameters)
	}

	if targetParameters.SqsQueueParameters != nil {
		config["sqs_queue"] = flattenTargetSqsQueueParameters(targetParameters.SqsQueueParameters)
	}

	if targetParameters.StepFunctionStateMachineParameters != nil {
		config["step_function"] = flattenTargetStepFunctionStateMachineParameters(targetParameters.StepFunctionStateMachineParameters)
	}

	if len(config) == 0 {
		return nil
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetBatchJobParameters(parameters *types.PipeTargetBatchJobParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.JobDefinition != nil {
		config["job_definition"] = aws.ToString(parameters.JobDefinition)
	}
	if parameters.JobName != nil {
		config["job_name"] = aws.ToString(parameters.JobName)
	}

	var parameterValues []map[string]interface{}
	for key, value := range parameters.Parameters {
		p := make(map[string]interface{})
		p["key"] = key
		p["value"] = value
		parameterValues = append(parameterValues, p)
	}
	config["parameters"] = parameterValues

	if parameters.RetryStrategy != nil {
		retryStrategyConfig := make(map[string]interface{})
		retryStrategyConfig["attempts"] = parameters.RetryStrategy.Attempts
		config["retry_strategy"] = []map[string]interface{}{retryStrategyConfig}
	}

	if parameters.ArrayProperties != nil {
		arrayPropertiesConfig := make(map[string]interface{})
		arrayPropertiesConfig["size"] = parameters.ArrayProperties.Size
		config["array_properties"] = []map[string]interface{}{arrayPropertiesConfig}
	}

	var dependsOnValues []map[string]interface{}
	for _, value := range parameters.DependsOn {
		dependsOn := make(map[string]interface{})
		dependsOn["job_id"] = aws.ToString(value.JobId)
		dependsOn["type"] = value.Type
		dependsOnValues = append(dependsOnValues, dependsOn)
	}
	config["depends_on"] = dependsOnValues

	if parameters.ContainerOverrides != nil {
		config["container_overrides"] = flattenTargetBatchContainerOverrides(parameters.ContainerOverrides)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetBatchContainerOverrides(parameters *types.BatchContainerOverrides) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.Command != nil {
		config["command"] = flex.FlattenStringValueSet(parameters.Command)
	}
	if parameters.InstanceType != nil {
		config["instance_type"] = aws.ToString(parameters.InstanceType)
	}

	var environmentValues []map[string]interface{}
	for _, value := range parameters.Environment {
		env := make(map[string]interface{})
		env["name"] = aws.ToString(value.Name)
		env["value"] = aws.ToString(value.Value)
		environmentValues = append(environmentValues, env)
	}
	config["environment"] = environmentValues

	var resourceRequirementsValues []map[string]interface{}
	for _, value := range parameters.ResourceRequirements {
		rr := make(map[string]interface{})
		rr["type"] = value.Type
		rr["value"] = aws.ToString(value.Value)
		resourceRequirementsValues = append(resourceRequirementsValues, rr)
	}
	config["resource_requirements"] = resourceRequirementsValues

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetCloudWatchLogsParameters(parameters *types.PipeTargetCloudWatchLogsParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.LogStreamName != nil {
		config["log_stream_name"] = aws.ToString(parameters.LogStreamName)
	}
	if parameters.Timestamp != nil {
		config["timestamp"] = aws.ToString(parameters.Timestamp)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetEcsTaskParameters(parameters *types.PipeTargetEcsTaskParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.TaskDefinitionArn != nil {
		config["task_definition_arn"] = aws.ToString(parameters.TaskDefinitionArn)
	}
	config["enable_ecs_managed_tags"] = parameters.EnableECSManagedTags
	config["enable_execute_command"] = parameters.EnableExecuteCommand
	if parameters.Group != nil {
		config["group"] = aws.ToString(parameters.Group)
	}
	if parameters.LaunchType != "" {
		config["launch_type"] = parameters.LaunchType
	}
	if parameters.PlatformVersion != nil {
		config["platform_version"] = aws.ToString(parameters.PlatformVersion)
	}
	if parameters.PropagateTags != "" {
		config["propagate_tags"] = parameters.PropagateTags
	}
	if parameters.ReferenceId != nil {
		config["reference_id"] = aws.ToString(parameters.ReferenceId)
	}
	if parameters.TaskCount != nil {
		config["task_count"] = aws.ToInt32(parameters.TaskCount)
	}

	var capacityProviderStrategyValues []map[string]interface{}
	for _, value := range parameters.CapacityProviderStrategy {
		strategy := make(map[string]interface{})
		strategy["capacity_provider"] = aws.ToString(value.CapacityProvider)
		strategy["base"] = value.Base
		strategy["weight"] = value.Weight
		capacityProviderStrategyValues = append(capacityProviderStrategyValues, strategy)
	}
	config["capacity_provider_strategy"] = capacityProviderStrategyValues

	var placementConstraintsValues []map[string]interface{}
	for _, value := range parameters.PlacementConstraints {
		constraint := make(map[string]interface{})
		constraint["expression"] = aws.ToString(value.Expression)
		constraint["type"] = value.Type
		placementConstraintsValues = append(placementConstraintsValues, constraint)
	}
	config["placement_constraints"] = placementConstraintsValues

	var placementStrategyValues []map[string]interface{}
	for _, value := range parameters.PlacementStrategy {
		strategy := make(map[string]interface{})
		strategy["field"] = aws.ToString(value.Field)
		strategy["type"] = value.Type
		placementStrategyValues = append(placementStrategyValues, strategy)
	}
	config["placement_strategy"] = placementStrategyValues

	var tagValues []map[string]interface{}
	for _, tag := range parameters.Tags {
		t := make(map[string]interface{})
		t["key"] = aws.ToString(tag.Key)
		t["value"] = aws.ToString(tag.Value)
		tagValues = append(tagValues, t)
	}
	config["tags"] = tagValues

	if parameters.NetworkConfiguration != nil {
		config["network_configuration"] = flattenTargetNetworkConfiguration(parameters.NetworkConfiguration)
	}

	if parameters.Overrides != nil {
		config["overrides"] = flattenTargetECSTaskOverrides(parameters.Overrides)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetNetworkConfiguration(parameters *types.NetworkConfiguration) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.AwsvpcConfiguration != nil {
		awsVpcConfiguration := make(map[string]interface{})
		awsVpcConfiguration["assign_public_ip"] = parameters.AwsvpcConfiguration.AssignPublicIp

		if parameters.AwsvpcConfiguration.SecurityGroups != nil {
			awsVpcConfiguration["security_groups"] = flex.FlattenStringValueSet(parameters.AwsvpcConfiguration.SecurityGroups)
		}

		if parameters.AwsvpcConfiguration.Subnets != nil {
			awsVpcConfiguration["subnets"] = flex.FlattenStringValueSet(parameters.AwsvpcConfiguration.Subnets)
		}

		config["aws_vpc_configuration"] = []map[string]interface{}{awsVpcConfiguration}
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetECSTaskOverrides(parameters *types.EcsTaskOverride) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.Cpu != nil {
		config["cpu"] = aws.ToString(parameters.Cpu)
	}
	if parameters.Memory != nil {
		config["memory"] = aws.ToString(parameters.Memory)
	}
	if parameters.ExecutionRoleArn != nil {
		config["execution_role_arn"] = aws.ToString(parameters.ExecutionRoleArn)
	}
	if parameters.TaskRoleArn != nil {
		config["task_role_arn"] = aws.ToString(parameters.TaskRoleArn)
	}

	if parameters.EphemeralStorage != nil {
		ecsEphemeralStorageConfig := make(map[string]interface{})
		ecsEphemeralStorageConfig["size_in_gib"] = parameters.EphemeralStorage.SizeInGiB
		config["ecs_ephemeral_storage"] = []map[string]interface{}{ecsEphemeralStorageConfig}
	}

	var inferenceAcceleratorOverridesValues []map[string]interface{}
	for _, value := range parameters.InferenceAcceleratorOverrides {
		override := make(map[string]interface{})
		override["device_name"] = aws.ToString(value.DeviceName)
		override["device_type"] = aws.ToString(value.DeviceType)
		inferenceAcceleratorOverridesValues = append(inferenceAcceleratorOverridesValues, override)
	}
	config["inference_accelerator_overrides"] = inferenceAcceleratorOverridesValues

	var overridesValues []map[string]interface{}
	for _, value := range parameters.ContainerOverrides {
		override := flattenTargetECSTaskOverrideContainerOverride(value)
		overridesValues = append(overridesValues, override)
	}
	config["container_overrides"] = overridesValues

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetECSTaskOverrideContainerOverride(parameters types.EcsContainerOverride) map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.Cpu != nil {
		config["cpu"] = aws.ToInt32(parameters.Cpu)
	}
	if parameters.Memory != nil {
		config["memory"] = aws.ToInt32(parameters.Memory)
	}
	if parameters.MemoryReservation != nil {
		config["memory_reservation"] = aws.ToInt32(parameters.MemoryReservation)
	}
	if parameters.Name != nil {
		config["name"] = aws.ToString(parameters.Name)
	}
	if parameters.Command != nil {
		config["command"] = flex.FlattenStringValueSet(parameters.Command)
	}

	var environmentValues []map[string]interface{}
	for _, value := range parameters.Environment {
		env := make(map[string]interface{})
		env["name"] = aws.ToString(value.Name)
		env["value"] = aws.ToString(value.Value)
		environmentValues = append(environmentValues, env)
	}
	config["environment"] = environmentValues

	var environmentFileValues []map[string]interface{}
	for _, value := range parameters.EnvironmentFiles {
		env := make(map[string]interface{})
		env["type"] = value.Type
		env["value"] = aws.ToString(value.Value)
		environmentFileValues = append(environmentFileValues, env)
	}
	config["environment_files"] = environmentFileValues

	var resourceRequirementsValues []map[string]interface{}
	for _, value := range parameters.ResourceRequirements {
		rr := make(map[string]interface{})
		rr["type"] = value.Type
		rr["value"] = aws.ToString(value.Value)
		resourceRequirementsValues = append(resourceRequirementsValues, rr)
	}
	config["resource_requirements"] = resourceRequirementsValues

	return config
}

func flattenTargetEventBridgeEventBusParameters(parameters *types.PipeTargetEventBridgeEventBusParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.DetailType != nil {
		config["detail_type"] = aws.ToString(parameters.DetailType)
	}
	if parameters.EndpointId != nil {
		config["endpoint_id"] = aws.ToString(parameters.EndpointId)
	}
	if parameters.Source != nil {
		config["source"] = aws.ToString(parameters.Source)
	}
	if parameters.Resources != nil {
		config["resources"] = flex.FlattenStringValueSet(parameters.Resources)
	}
	if parameters.Time != nil {
		config["time"] = aws.ToString(parameters.Time)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetHttpParameters(parameters *types.PipeTargetHttpParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	var headerParameters []map[string]interface{}
	for key, value := range parameters.HeaderParameters {
		header := make(map[string]interface{})
		header["key"] = key
		header["value"] = value
		headerParameters = append(headerParameters, header)
	}
	config["header"] = headerParameters

	var queryStringParameters []map[string]interface{}
	for key, value := range parameters.QueryStringParameters {
		queryString := make(map[string]interface{})
		queryString["key"] = key
		queryString["value"] = value
		queryStringParameters = append(queryStringParameters, queryString)
	}
	config["query_string"] = queryStringParameters
	config["path_parameters"] = flex.FlattenStringValueList(parameters.PathParameterValues)

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetKinesisStreamParameters(parameters *types.PipeTargetKinesisStreamParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.PartitionKey != nil {
		config["partition_key"] = aws.ToString(parameters.PartitionKey)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetLambdaFunctionParameters(parameters *types.PipeTargetLambdaFunctionParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.InvocationType != "" {
		config["invocation_type"] = parameters.InvocationType
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetRedshiftDataParameters(parameters *types.PipeTargetRedshiftDataParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.Database != nil {
		config["database"] = aws.ToString(parameters.Database)
	}
	if parameters.DbUser != nil {
		config["database_user"] = aws.ToString(parameters.DbUser)
	}
	if parameters.SecretManagerArn != nil {
		config["secret_manager_arn"] = aws.ToString(parameters.SecretManagerArn)
	}
	if parameters.StatementName != nil {
		config["statement_name"] = aws.ToString(parameters.StatementName)
	}
	config["with_event"] = parameters.WithEvent
	if parameters.Sqls != nil {
		config["sqls"] = flex.FlattenStringValueSet(parameters.Sqls)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetSageMakerPipelineParameters(parameters *types.PipeTargetSageMakerPipelineParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if len(parameters.PipelineParameterList) != 0 {
		var params []map[string]interface{}
		for _, param := range parameters.PipelineParameterList {
			item := make(map[string]interface{})
			item["name"] = aws.ToString(param.Name)
			item["value"] = aws.ToString(param.Value)
			params = append(params, item)
		}
		config["parameters"] = params
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetSqsQueueParameters(parameters *types.PipeTargetSqsQueueParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.MessageDeduplicationId != nil {
		config["message_deduplication_id"] = aws.ToString(parameters.MessageDeduplicationId)
	}
	if parameters.MessageGroupId != nil {
		config["message_group_id"] = aws.ToString(parameters.MessageGroupId)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenTargetStepFunctionStateMachineParameters(parameters *types.PipeTargetStateMachineParameters) []map[string]interface{} {
	config := make(map[string]interface{})

	if parameters.InvocationType != "" {
		config["invocation_type"] = parameters.InvocationType
	}

	result := []map[string]interface{}{config}
	return result
}
