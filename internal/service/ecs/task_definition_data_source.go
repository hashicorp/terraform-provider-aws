// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecs_task_definition", name="Task Definition")
func dataSourceTaskDefinition() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTaskDefinitionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_without_revision": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_definitions": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_fault_injection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ephemeral_storage": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_in_gib": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrExecutionRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFamily: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inference_accelerator": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ipc_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pid_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrExpression: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"proxy_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrProperties: {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"requires_compatibilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"runtime_platform": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operating_system_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cpu_architecture": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"task_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configure_at_launch": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"docker_volume_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"autoprovision": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"driver": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"driver_opts": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Computed: true,
									},
									"labels": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Computed: true,
									},
									names.AttrScope: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"efs_volume_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorization_config": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_point_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"iam": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									names.AttrFileSystemID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"root_directory": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"transit_encryption": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"transit_encryption_port": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"fsx_windows_file_server_volume_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorization_config": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"credentials_parameter": {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrDomain: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									names.AttrFileSystemID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"root_directory": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"host_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	taskDefinitionName := d.Get("task_definition").(string)
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}

	taskDefinition, _, err := findTaskDefinition(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition (%s): %s", taskDefinitionName, err)
	}

	d.SetId(aws.ToString(taskDefinition.TaskDefinitionArn))
	d.Set(names.AttrARN, taskDefinition.TaskDefinitionArn)
	d.Set("arn_without_revision", taskDefinitionARNStripRevision(aws.ToString(taskDefinition.TaskDefinitionArn)))

	orderedCDs := taskDefinition.ContainerDefinitions
	containerDefinitions(orderedCDs).orderContainers()
	containerDefinitions(orderedCDs).orderEnvironmentVariables()
	containerDefinitions(orderedCDs).orderSecrets()
	containerDefinitions(orderedCDs).compactArrays()
	containerDefinitions, err := flattenContainerDefinitions(orderedCDs)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("container_definitions", containerDefinitions); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting container_definitions: %s", err)
	}

	d.Set("cpu", taskDefinition.Cpu)
	d.Set("enable_fault_injection", taskDefinition.EnableFaultInjection)
	if err := d.Set("ephemeral_storage", flattenEphemeralStorage(taskDefinition.EphemeralStorage)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ephemeral_storage: %s", err)
	}
	d.Set(names.AttrExecutionRoleARN, taskDefinition.ExecutionRoleArn)
	d.Set(names.AttrFamily, taskDefinition.Family)
	if err := d.Set("inference_accelerator", flattenInferenceAccelerators(taskDefinition.InferenceAccelerators)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting inference_accelerator: %s", err)
	}
	d.Set("ipc_mode", taskDefinition.IpcMode)
	d.Set("memory", taskDefinition.Memory)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("pid_mode", taskDefinition.PidMode)
	if err := d.Set("placement_constraints", flattenTaskDefinitionPlacementConstraints(taskDefinition.PlacementConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting placement_constraints: %s", err)
	}
	if err := d.Set("proxy_configuration", flattenProxyConfiguration(taskDefinition.ProxyConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting proxy_configuration: %s", err)
	}
	d.Set("requires_compatibilities", taskDefinition.RequiresCompatibilities)
	d.Set("revision", taskDefinition.Revision)
	if err := d.Set("runtime_platform", flattenRuntimePlatform(taskDefinition.RuntimePlatform)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting runtime_platform: %s", err)
	}
	d.Set(names.AttrStatus, taskDefinition.Status)
	d.Set("task_role_arn", taskDefinition.TaskRoleArn)
	if err := d.Set("volume", flattenVolumes(taskDefinition.Volumes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting volume: %s", err)
	}

	return diags
}
