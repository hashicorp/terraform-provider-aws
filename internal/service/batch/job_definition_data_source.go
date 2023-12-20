// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/batch"
	batchtypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Job Definition")
func newDataSourceJobDefinition(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceJobDefinition{}, nil
}

const (
	DSNameJobDefinition = "Job Definition Data Source"
)

func (r *resourceJobQueue) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("arn"),
			path.MatchRoot("name"),
		),
	}
}

type dataSourceJobDefinition struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceJobDefinition) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_batch_job_definition"
}

func (d *dataSourceJobDefinition) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
			},
			"container_orchestration_type": schema.StringAttribute{
				Computed: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Optional: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"revision": schema.Int64Attribute{
				Optional: true,
			},
			"status": schema.StringAttribute{
				Optional: true,
				// Default: JobDefinitionStatusActive,
				// https://github.com/hashicorp/terraform-plugin-framework/issues/751#issuecomment-1799757575
				Validators: []validator.String{
					stringvalidator.OneOf(JobDefinitionStatus_Values()...),
				},
			},
			"scheduling_priority": schema.Int64Attribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"eks_properties": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{
					"pod_properties": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"dns_policy": schema.StringAttribute{
								Computed: true,
							},
							"host_network": schema.BoolAttribute{
								Computed: true,
							},
							"service_account_name": schema.StringAttribute{
								Computed: true,
							},
						},
						Blocks: map[string]schema.Block{
							"containers": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"args": schema.ListAttribute{
											Computed:    true,
											ElementType: types.StringType,
										},
										"commands": schema.ListAttribute{
											Computed:    true,
											ElementType: types.StringType,
										},
										"image": schema.StringAttribute{
											Computed: true,
										},
										"image_pull_policy": schema.StringAttribute{
											Computed: true,
										},
										"name": schema.StringAttribute{
											Computed: true,
										},
									},
									Blocks: map[string]schema.Block{
										"env": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
										"resources": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"limits": schema.MapAttribute{
													Computed:    true,
													ElementType: types.StringType,
												},
												"requests": schema.MapAttribute{
													Computed:    true,
													ElementType: types.StringType,
												},
											},
										},
										"security_context": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"privileged": schema.BoolAttribute{
													Computed: true,
												},
												"run_as_user": schema.Int64Attribute{
													Computed: true,
												},
												"run_as_group": schema.Int64Attribute{
													Computed: true,
												},
												"run_as_non_root": schema.BoolAttribute{
													Computed: true,
												},
												"read_only_root_filesystem": schema.BoolAttribute{
													Computed: true,
												},
											},
										},
										"volume_mounts": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"mount_path": schema.StringAttribute{
														Computed: true,
													},
													"name": schema.StringAttribute{
														Computed: true,
													},
													"read_only": schema.BoolAttribute{
														Computed: true,
													},
												},
											},
										},
									},
								},
							},
							"metadata": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"labels": schema.MapAttribute{
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
							"volumes": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Computed: true,
										},
									},
									Blocks: map[string]schema.Block{
										"empty_dir": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"medium": schema.StringAttribute{
													Computed: true,
												},
												"size_limit": schema.Int64Attribute{
													Computed: true,
												},
											},
										},
										"host_path": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"path": schema.StringAttribute{
													Computed: true,
												},
											},
										},
										"secret": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"secret_name": schema.StringAttribute{
													Computed: true,
												},
												"optional": schema.BoolAttribute{
													Computed: true,
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
			"node_properties": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"main_node": schema.Int64Attribute{
						Computed: true,
					},
					"num_nodes": schema.Int64Attribute{
						Computed: true,
					},
				},
				Blocks: map[string]schema.Block{
					"node_range_properties": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"target_nodes": schema.StringAttribute{
									Computed: true,
								},
							},
							Blocks: map[string]schema.Block{
								"container": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"command": schema.ListAttribute{
											Computed:    true,
											ElementType: types.StringType,
										},
										"execution_role_arn": schema.StringAttribute{
											Computed: true,
										},
										"image": schema.StringAttribute{
											Computed: true,
										},
										"instance_type": schema.StringAttribute{
											Computed: true,
										},
										"job_role_arn": schema.StringAttribute{
											Computed: true,
										},
										"privileged": schema.BoolAttribute{
											Computed: true,
										},
										"readonly_root_filesystem": schema.BoolAttribute{
											Computed: true,
										},
										"user": schema.StringAttribute{
											Computed: true,
										},
									},
									Blocks: map[string]schema.Block{
										"environment": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
										"ephemeral_storage": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"size_in_gib": schema.Int64Attribute{
													Computed: true,
												},
											},
										},
										"fargate_platform_configuration": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"platform_version": schema.StringAttribute{
													Computed: true,
												},
											},
										},
										"linux_parameters": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"init_process_enabled": schema.BoolAttribute{
													Computed: true,
												},
												"max_swap": schema.Int64Attribute{
													Computed: true,
												},
												"shared_memory_size": schema.Int64Attribute{
													Computed: true,
												},
												"swappiness": schema.Int64Attribute{
													Computed: true,
												},
											},
											Blocks: map[string]schema.Block{
												"devices": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"host_path": schema.StringAttribute{
																Computed: true,
															},
															"container_path": schema.StringAttribute{
																Computed: true,
															},
															"permissions": schema.ListAttribute{
																Computed:    true,
																ElementType: types.StringType,
															},
														},
													},
												},
												"tmpfs": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"container_path": schema.StringAttribute{
																Computed: true,
															},
															"size": schema.Int64Attribute{
																Computed: true,
															},
															"mount_options": schema.ListAttribute{
																Computed:    true,
																ElementType: types.StringType,
															},
														},
													},
												},
											},
										},
										"log_configuration": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"options": schema.MapAttribute{
													Computed:    true,
													ElementType: types.StringType,
												},
												"log_driver": schema.StringAttribute{
													Computed: true,
												},
											},
											Blocks: map[string]schema.Block{
												"secret_options": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"name": schema.StringAttribute{
																Computed: true,
															},
															"value_from": schema.StringAttribute{
																Computed: true,
															},
														},
													},
												},
											},
										},
										"mount_points": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"container_path": schema.StringAttribute{
														Computed: true,
													},
													"read_only": schema.BoolAttribute{
														Computed: true,
													},
													"source_volume": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
										"network_configuration": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"assign_public_ip": schema.StringAttribute{
													Computed: true,
												},
											},
										},
										"resource_requirements": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"type": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
										"runtime_platform": schema.SingleNestedBlock{
											Attributes: map[string]schema.Attribute{
												"cpu_architecture": schema.StringAttribute{
													Computed: true,
												},
												"operating_system_family": schema.StringAttribute{
													Computed: true,
												},
											},
										},
										"secrets": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Computed: true,
													},
													"value_from": schema.StringAttribute{
														Computed: true,
													},
												},
											},
										},
										"ulimits": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"hard_limit": schema.Int64Attribute{
														Computed: true,
													},
													"name": schema.StringAttribute{
														Computed: true,
													},
													"soft_limit": schema.Int64Attribute{
														Computed: true,
													},
												},
											},
										},
										"volumes": schema.ListNestedBlock{
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Computed: true,
													},
												},
												Blocks: map[string]schema.Block{
													"host": schema.SingleNestedBlock{
														Attributes: map[string]schema.Attribute{
															"source_path": schema.StringAttribute{
																Computed: true,
															},
														},
													},
													"efs_volume_configuration": schema.SingleNestedBlock{
														Attributes: map[string]schema.Attribute{
															"file_system_id": schema.StringAttribute{
																Computed: true,
															},
															"root_directory": schema.StringAttribute{
																Computed: true,
															},
															"transit_encryption": schema.StringAttribute{
																Computed: true,
															},
															"transit_encryption_port": schema.Int64Attribute{
																Computed: true,
															},
														},
														Blocks: map[string]schema.Block{
															"authorization_config": schema.SingleNestedBlock{
																Attributes: map[string]schema.Attribute{
																	"access_point_id": schema.StringAttribute{
																		Computed: true,
																	},
																	"iam": schema.StringAttribute{
																		Computed: true,
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
							},
						},
					},
				},
			},
			"retry_strategy": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"attempts": schema.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							int64validator.Between(1, 10),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"evaluate_on_exit": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"action": schema.StringAttribute{
									Computed: true,
								},
								"on_exit_code": schema.StringAttribute{
									Computed: true,
								},
								"on_reason": schema.StringAttribute{
									Computed: true,
								},
								"on_status_reason": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
			"timeout": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"attempt_duration_seconds": schema.Int64Attribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func (d *dataSourceJobDefinition) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().BatchClient(ctx)

	var data dataSourceJobDefinitionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jd := batchtypes.JobDefinition{}

	if !data.ARN.IsNull() {
		out, err := FindJobDefinitionV2ByARN(ctx, conn, aws.StringValue(flex.StringFromFramework(ctx, data.ARN)))

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionReading, DSNameJobDefinition, data.Name.String(), err),
				err.Error(),
			)
			return
		}
		jd = *out
	}

	if !data.Name.IsNull() {
		input := &batch.DescribeJobDefinitionsInput{
			JobDefinitionName: flex.StringFromFramework(ctx, data.Name),
		}

		if data.Status.IsNull() {
			active := JobDefinitionStatusActive
			input.Status = &active
		} else {
			input.Status = flex.StringFromFramework(ctx, data.Status)
		}

		jds, err := ListJobDefinitionsV2ByNameWithStatus(ctx, conn, input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionReading, DSNameJobDefinition, data.Name.String(), err),
				err.Error(),
			)
		}

		if !data.Revision.IsNull() {
			for _, _jd := range jds {
				if aws.Int32Value(_jd.Revision) == int32(data.Revision.ValueInt64()) {
					jd = _jd
				}
			}

			if jd.JobDefinitionArn == nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.Batch, create.ErrActionReading, DSNameJobDefinition, data.Name.String(), fmt.Errorf("job definition revision %d not found", data.Revision.ValueInt64())),
					fmt.Sprintf("job definition revision %d not found with name %s", data.Revision.ValueInt64(), data.Name.String()),
				)
			}
		}

		if data.Revision.IsNull() {
			var latestRevision int32 = 0
			for _, _jd := range jds {
				if aws.Int32Value(_jd.Revision) > latestRevision {
					latestRevision = aws.Int32Value(_jd.Revision)
					jd = _jd
				}
			}
		}
	}

	// These fields don't have the same name as their api
	data.ARN = flex.StringToFrameworkARN(ctx, jd.JobDefinitionArn)
	data.ID = flex.StringToFramework(ctx, jd.JobDefinitionArn)
	data.Name = flex.StringToFramework(ctx, jd.JobDefinitionName)
	data.Revision = flex.Int32ToFramework(ctx, jd.Revision)
	data.Status = flex.StringToFramework(ctx, jd.Status)
	data.Type = flex.StringToFramework(ctx, jd.Type)
	data.ContainerOrchestrationType = types.StringValue(string(jd.ContainerOrchestrationType))
	data.SchedulingPriority = flex.Int32ToFramework(ctx, jd.SchedulingPriority)
	if jd.Timeout != nil {
		data.Timeout = types.ObjectValueMust(timeoutAttr, map[string]attr.Value{
			"attempt_duration_seconds": flex.Int32ToFramework(ctx, jd.Timeout.AttemptDurationSeconds),
		})
	} else {
		data.Timeout = types.ObjectNull(timeoutAttr)
	}

	resp.Diagnostics.Append(frameworkFlattenNodeProperties(ctx, jd.NodeProperties, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(frameworkFlattenEKSproperties(ctx, jd.EksProperties, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(frameworkFlattenRetryStrategy(ctx, jd.RetryStrategy, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func frameworkFlattenEKSproperties(ctx context.Context, apiObject *batchtypes.EksProperties, data *dataSourceJobDefinitionData) (diags diag.Diagnostics) {
	if apiObject == nil {
		data.EksProperties = types.ObjectNull(eksPropertiesAttr)
		return
	}
	props := map[string]attr.Value{
		"dns_policy":           flex.StringToFramework(ctx, apiObject.PodProperties.DnsPolicy),
		"host_network":         flex.BoolToFramework(ctx, apiObject.PodProperties.HostNetwork),
		"service_account_name": flex.StringToFramework(ctx, apiObject.PodProperties.ServiceAccountName),
	}

	if apiObject.PodProperties.Metadata != nil {
		props["metadata"] = types.ObjectValueMust(eksMetadataAttr, map[string]attr.Value{
			"labels": flex.FlattenFrameworkStringMap(ctx, aws.StringMap(apiObject.PodProperties.Metadata.Labels)),
		})
	} else {
		props["metadata"] = types.ObjectNull(eksMetadataAttr)
	}

	if len(apiObject.PodProperties.Containers) > 0 {
		container, d := types.ListValue(types.ObjectType{AttrTypes: eksContainerAttr}, frameworkFlattenEKSContainer(ctx, apiObject.PodProperties.Containers))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		props["containers"] = container
	} else {
		props["containers"] = types.ListNull(types.ObjectType{AttrTypes: eksContainerAttr})
	}
	if len(apiObject.PodProperties.Volumes) > 0 {
		volume, d := types.ListValue(types.ObjectType{AttrTypes: eksVolumeAttr}, frameworkFlattenEKSVolume(ctx, apiObject.PodProperties.Volumes))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		props["volumes"] = volume
	} else {
		props["volumes"] = types.ListNull(types.ObjectType{AttrTypes: eksVolumeAttr})
	}

	podProps, d := types.ObjectValue(eksPodPropertiesAttr, props)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	data.EksProperties = types.ObjectValueMust(eksPropertiesAttr, map[string]attr.Value{
		"pod_properties": podProps,
	})
	return diags
}

func frameworkFlattenEKSContainer(ctx context.Context, apiObject []batchtypes.EksContainer) []attr.Value {
	var containers []attr.Value
	for _, c := range apiObject {
		props := map[string]attr.Value{
			"image":             flex.StringToFramework(ctx, c.Image),
			"image_pull_policy": flex.StringToFramework(ctx, c.ImagePullPolicy),
			"name":              flex.StringToFramework(ctx, c.Name),
			"args":              flex.FlattenFrameworkStringList(ctx, aws.StringSlice(c.Args)),
			"commands":          flex.FlattenFrameworkStringList(ctx, aws.StringSlice(c.Command)),
		}
		if c.SecurityContext != nil {
			props["security_context"] = types.ObjectValueMust(eksContainerSecurityContextAttr, map[string]attr.Value{
				"privileged":                flex.BoolToFramework(ctx, c.SecurityContext.Privileged),
				"run_as_user":               flex.Int64ToFramework(ctx, c.SecurityContext.RunAsUser),
				"run_as_group":              flex.Int64ToFramework(ctx, c.SecurityContext.RunAsGroup),
				"run_as_non_root":           flex.BoolToFramework(ctx, c.SecurityContext.RunAsNonRoot),
				"read_only_root_filesystem": flex.BoolToFramework(ctx, c.SecurityContext.ReadOnlyRootFilesystem),
			})
		} else {
			props["security_context"] = types.ObjectNull(eksContainerSecurityContextAttr)
		}
		if len(c.VolumeMounts) > 0 {
			props["volume_mounts"] = types.ListValueMust(types.ObjectType{AttrTypes: eksContainerVolumeMountAttr}, frameworkFlattenEKSContainerVolumeMount(ctx, c.VolumeMounts))
		} else {
			props["volume_mounts"] = types.ListNull(types.ObjectType{AttrTypes: eksContainerVolumeMountAttr})
		}

		if len(c.Env) > 0 {
			props["env"] = types.ListValueMust(types.ObjectType{AttrTypes: eksContainerEnvironmentVariableAttr}, frameworkFlattenEKSContainerEnv(ctx, c.Env))
		} else {
			props["env"] = types.ListNull(types.ObjectType{AttrTypes: eksContainerEnvironmentVariableAttr})
		}

		if c.Resources != nil {
			props["resources"] = types.ObjectValueMust(eksContainerResourceRequirementsAttr, map[string]attr.Value{
				"limits":   flex.FlattenFrameworkStringMap(ctx, aws.StringMap(c.Resources.Limits)),
				"requests": flex.FlattenFrameworkStringMap(ctx, aws.StringMap(c.Resources.Requests)),
			})
		} else {
			props["resources"] = types.ObjectNull(eksContainerResourceRequirementsAttr)
		}

		containers = append(containers, types.ObjectValueMust(eksContainerAttr, props))
	}
	return containers
}

func frameworkFlattenNodeProperties(ctx context.Context, props *batchtypes.NodeProperties, data *dataSourceJobDefinitionData) (diags diag.Diagnostics) {
	att := fwtypes.AttributeTypesMust[frameworkNodeProperties](ctx)
	if props == nil {
		data.EksProperties = types.ObjectNull(att)
		return
	}
	att["node_range_properties"] = types.ListType{ElemType: types.ObjectType{AttrTypes: nodeRangePropertiesAttr}}
	if props == nil {
		data.NodeProperties = types.ObjectNull(att)
		return
	}

	var properties []attr.Value
	for _, v := range props.NodeRangeProperties {
		container, d := types.ObjectValue(containerPropertiesAttr, frameworkFlattenContainerProperties(ctx, v.Container))
		diags = append(diags, d...)
		if diags.HasError() {
			return
		}
		properties = append(properties, types.ObjectValueMust(nodeRangePropertiesAttr, map[string]attr.Value{
			"container":    container,
			"target_nodes": flex.StringToFramework(ctx, v.TargetNodes),
		}))
	}
	data.NodeProperties = types.ObjectValueMust(att, map[string]attr.Value{
		"main_node":             flex.Int32ToFramework(ctx, props.MainNode),
		"num_nodes":             flex.Int32ToFramework(ctx, props.NumNodes),
		"node_range_properties": types.ListValueMust(types.ObjectType{AttrTypes: nodeRangePropertiesAttr}, properties),
	})
	return
}

func frameworkFlattenEKSVolume(ctx context.Context, apiObject []batchtypes.EksVolume) (volumes []attr.Value) {
	for _, v := range apiObject {
		volume := map[string]attr.Value{
			"name": flex.StringToFramework(ctx, v.Name),
		}
		if v.EmptyDir != nil {
			volume["empty_dir"] = types.ObjectValueMust(eksVolumeEmptyDirAttr, map[string]attr.Value{
				"medium":     flex.StringToFramework(ctx, v.EmptyDir.Medium),
				"size_limit": flex.StringToFramework(ctx, v.EmptyDir.SizeLimit),
			})
		} else {
			volume["empty_dir"] = types.ObjectNull(eksVolumeEmptyDirAttr)
		}
		if v.HostPath != nil {
			volume["host"] = types.ObjectValueMust(eksVolumeHostPathAttr, map[string]attr.Value{
				"path": flex.StringToFramework(ctx, v.HostPath.Path),
			})
		} else {
			volume["host"] = types.ObjectNull(eksVolumeHostPathAttr)
		}
		if v.Secret != nil {
			volume["secret"] = types.ObjectValueMust(eksVolumeSecretAttr, map[string]attr.Value{
				"secret_name": flex.StringToFramework(ctx, v.Secret.SecretName),
				"optional":    flex.BoolToFramework(ctx, v.Secret.Optional),
			})
		} else {
			volume["secret"] = types.ObjectNull(eksVolumeSecretAttr)
		}
		volumes = append(volumes, types.ObjectValueMust(eksVolumeAttr, volume))
	}
	return volumes
}

func frameworkFlattenEKSContainerVolumeMount(ctx context.Context, apiObject []batchtypes.EksContainerVolumeMount) (volumeMounts []attr.Value) {
	for _, v := range apiObject {
		volumeMounts = append(volumeMounts, types.ObjectValueMust(eksContainerVolumeMountAttr, map[string]attr.Value{
			"mount_path": flex.StringToFramework(ctx, v.MountPath),
			"name":       flex.StringToFramework(ctx, v.Name),
			"read_only":  flex.BoolToFramework(ctx, v.ReadOnly),
		}))
	}
	return
}

func frameworkFlattenEKSContainerEnv(ctx context.Context, apiObject []batchtypes.EksContainerEnvironmentVariable) (env []attr.Value) {
	for _, v := range apiObject {
		env = append(env, types.ObjectValueMust(eksContainerEnvironmentVariableAttr, map[string]attr.Value{
			"name":  flex.StringToFramework(ctx, v.Name),
			"value": flex.StringToFramework(ctx, v.Value),
		}))
	}
	return
}

func frameworkFlattenContainerProperties(ctx context.Context, c *batchtypes.ContainerProperties) map[string]attr.Value {
	containerProps := map[string]attr.Value{
		"command":                  flex.FlattenFrameworkStringList(ctx, aws.StringSlice(c.Command)),
		"execution_role_arn":       flex.StringToFramework(ctx, c.ExecutionRoleArn),
		"image":                    flex.StringToFramework(ctx, c.Image),
		"instance_type":            flex.StringToFramework(ctx, c.InstanceType),
		"job_role_arn":             flex.StringToFramework(ctx, c.JobRoleArn),
		"privileged":               flex.BoolToFramework(ctx, c.Privileged),
		"readonly_root_filesystem": flex.BoolToFramework(ctx, c.ReadonlyRootFilesystem),
		"user":                     flex.StringToFramework(ctx, c.User),
	}

	if (c.EphemeralStorage != nil) && (c.EphemeralStorage.SizeInGiB != nil) {
		containerProps["ephemeral_storage"] = types.ObjectValueMust(ephemeralStorageAttr, map[string]attr.Value{
			"size_in_gib": flex.Int32ToFramework(ctx, c.EphemeralStorage.SizeInGiB),
		})
	} else {
		containerProps["ephemeral_storage"] = types.ObjectNull(ephemeralStorageAttr)
	}

	if c.LinuxParameters != nil {
		containerProps["linux_parameters"] = types.ObjectValueMust(
			linuxParametersAttr,
			frameworkFlattenContainerLinuxParameters(ctx, c.LinuxParameters),
		)
	} else {
		containerProps["linux_parameters"] = types.ObjectNull(linuxParametersAttr)
	}

	if c.FargatePlatformConfiguration != nil {
		containerProps["fargate_platform_configuration"] = types.ObjectValueMust(fargatePlatformConfigurationAttr, map[string]attr.Value{
			"platform_version": flex.StringToFramework(ctx, c.FargatePlatformConfiguration.PlatformVersion),
		})
	} else {
		containerProps["fargate_platform_configuration"] = types.ObjectNull(fargatePlatformConfigurationAttr)
	}

	if c.NetworkConfiguration != nil {
		containerProps["network_configuration"] = types.ObjectValueMust(networkConfigurationAttr, map[string]attr.Value{
			"assign_public_ip": flex.StringToFramework(ctx, aws.String(string(c.NetworkConfiguration.AssignPublicIp))),
		})
	} else {
		containerProps["network_configuration"] = types.ObjectNull(networkConfigurationAttr)
	}

	if c.RuntimePlatform != nil {
		containerProps["runtime_platform"] = types.ObjectValueMust(runtimePlatformAttr, map[string]attr.Value{
			"cpu_architecture":        flex.StringToFramework(ctx, c.RuntimePlatform.CpuArchitecture),
			"operating_system_family": flex.StringToFramework(ctx, c.RuntimePlatform.OperatingSystemFamily),
		})
	} else {
		containerProps["runtime_platform"] = types.ObjectNull(runtimePlatformAttr)
	}

	var environment []attr.Value
	if len(c.Environment) > 0 {
		for _, env := range c.Environment {
			environment = append(environment, types.ObjectValueMust(keyValuePairAttr, map[string]attr.Value{
				"name":  flex.StringToFramework(ctx, env.Name),
				"value": flex.StringToFramework(ctx, env.Value),
			}))
		}
		containerProps["environment"] = types.ListValueMust(types.ObjectType{AttrTypes: keyValuePairAttr}, environment)
	} else {
		containerProps["environment"] = types.ListNull(types.ObjectType{AttrTypes: keyValuePairAttr})
	}
	if len(c.Environment) > 0 {
		var mountPoints []attr.Value
		for _, m := range c.MountPoints {
			mountPoints = append(environment, types.ObjectValueMust(mountPointAttr, map[string]attr.Value{
				"container_path": flex.StringToFramework(ctx, m.ContainerPath),
				"read_only":      flex.BoolToFramework(ctx, m.ReadOnly),
				"source_volume":  flex.StringToFramework(ctx, m.SourceVolume),
			}))
		}
		containerProps["mount_points"] = types.ListValueMust(types.ObjectType{AttrTypes: mountPointAttr}, mountPoints)
	} else {
		containerProps["mount_points"] = types.ListNull(types.ObjectType{AttrTypes: mountPointAttr})
	}

	var logConfigurationSecrets []attr.Value
	if c.LogConfiguration != nil {
		if len(c.LogConfiguration.SecretOptions) > 0 {
			for _, sec := range c.LogConfiguration.SecretOptions {
				logConfigurationSecrets = append(logConfigurationSecrets, types.ObjectValueMust(secretAttr, map[string]attr.Value{
					"name":       flex.StringToFramework(ctx, sec.Name),
					"value_from": flex.StringToFramework(ctx, sec.ValueFrom),
				}))
			}
			containerProps["log_configuration"] = types.ObjectValueMust(logConfigurationAttr, map[string]attr.Value{
				"options":     flex.FlattenFrameworkStringMap(ctx, aws.StringMap(c.LogConfiguration.Options)),
				"log_driver":  flex.StringToFramework(ctx, aws.String(string(c.LogConfiguration.LogDriver))),
				"secret_opts": types.ListValueMust(types.ObjectType{AttrTypes: secretAttr}, logConfigurationSecrets),
			})
		} else {
			containerProps["log_configuration"] = types.ObjectValueMust(logConfigurationAttr, map[string]attr.Value{
				"options":     flex.FlattenFrameworkStringMap(ctx, aws.StringMap(c.LogConfiguration.Options)),
				"log_driver":  flex.StringToFramework(ctx, aws.String(string(c.LogConfiguration.LogDriver))),
				"secret_opts": types.ListNull(types.ObjectType{AttrTypes: secretAttr}),
			})
		}
	} else {
		containerProps["log_configuration"] = types.ObjectNull(logConfigurationAttr)
	}

	var resourceRequirements []attr.Value
	if len(c.ResourceRequirements) > 0 {
		for _, res := range c.ResourceRequirements {
			resourceRequirements = append(resourceRequirements, types.ObjectValueMust(resourceRequirementsAttr, map[string]attr.Value{
				"type":  flex.StringToFramework(ctx, aws.String(string(res.Type))),
				"value": flex.StringToFramework(ctx, res.Value),
			}))
		}
		containerProps["resource_requirements"] = types.ListValueMust(types.ObjectType{AttrTypes: resourceRequirementsAttr}, resourceRequirements)
	} else {
		containerProps["resource_requirements"] = types.ListNull(types.ObjectType{AttrTypes: resourceRequirementsAttr})
	}

	var secrets []attr.Value
	if len(c.Secrets) > 0 {
		for _, sec := range c.Secrets {
			secrets = append(secrets, types.ObjectValueMust(secretAttr, map[string]attr.Value{
				"name":       flex.StringToFramework(ctx, sec.Name),
				"value_from": flex.StringToFramework(ctx, sec.ValueFrom),
			}))
		}
		containerProps["secrets"] = types.ListValueMust(types.ObjectType{AttrTypes: secretAttr}, secrets)
	} else {
		containerProps["secrets"] = types.ListNull(types.ObjectType{AttrTypes: secretAttr})
	}

	if len(c.Ulimits) > 0 {
		var ulimits []attr.Value
		for _, ul := range c.Ulimits {
			ulimits = append(ulimits, types.ObjectValueMust(ulimitsAttr, map[string]attr.Value{
				"hard_limit": flex.Int32ToFramework(ctx, ul.HardLimit),
				"name":       flex.StringToFramework(ctx, ul.Name),
				"soft_limit": flex.Int32ToFramework(ctx, ul.SoftLimit),
			}))
		}
		containerProps["ulimits"] = types.ListValueMust(types.ObjectType{AttrTypes: ulimitsAttr}, ulimits)
	} else {
		containerProps["ulimits"] = types.ListNull(types.ObjectType{AttrTypes: ulimitsAttr})
	}

	if len(c.Volumes) > 0 {
		var volumes []attr.Value
		for _, vol := range c.Volumes {
			volume := map[string]attr.Value{
				"name": flex.StringToFramework(ctx, vol.Name),
			}
			if vol.Host != nil {
				volume["host"] = types.ObjectValueMust(hostAttr, map[string]attr.Value{
					"source_path": flex.StringToFramework(ctx, vol.Host.SourcePath),
				})
			}
			if vol.EfsVolumeConfiguration != nil {
				volume["efs_volume_configuration"] = types.ObjectValueMust(efsVolumeConfigurationAttr, map[string]attr.Value{
					"file_system_id":          flex.StringToFramework(ctx, vol.EfsVolumeConfiguration.FileSystemId),
					"root_directory":          flex.StringToFramework(ctx, vol.EfsVolumeConfiguration.RootDirectory),
					"transit_encryption":      flex.StringToFramework(ctx, aws.String(string(vol.EfsVolumeConfiguration.TransitEncryption))),
					"transit_encryption_port": flex.Int32ToFramework(ctx, vol.EfsVolumeConfiguration.TransitEncryptionPort),
					"authorization_config": types.ObjectValueMust(authorizationConfigAttr, map[string]attr.Value{
						"access_point_id": flex.StringToFramework(ctx, vol.EfsVolumeConfiguration.AuthorizationConfig.AccessPointId),
						"iam":             flex.StringToFramework(ctx, aws.String(string(vol.EfsVolumeConfiguration.AuthorizationConfig.Iam))),
					}),
				})

			}
			volumes = append(volumes, types.ObjectValueMust(volumeAttr, volume))
		}
	} else {
		containerProps["volumes"] = types.ListNull(types.ObjectType{AttrTypes: volumeAttr})
	}
	return containerProps
}

func frameworkFlattenContainerLinuxParameters(ctx context.Context, lp *batchtypes.LinuxParameters) map[string]attr.Value {
	linuxProps := map[string]attr.Value{
		"init_process_enabled": flex.BoolToFramework(ctx, lp.InitProcessEnabled),
		"max_swap":             flex.Int32ToFramework(ctx, lp.MaxSwap),
		"shared_memory_size":   flex.Int32ToFramework(ctx, lp.SharedMemorySize),
		"swappiness":           flex.Int32ToFramework(ctx, lp.Swappiness),
	}
	if len(lp.Devices) > 0 {
		linuxProps["devices"] = types.ListValueMust(types.ObjectType{AttrTypes: deviceAttr}, frameworkFlattenContainerDevices(ctx, lp.Devices))
	} else {
		linuxProps["devices"] = types.ListNull(types.ObjectType{AttrTypes: deviceAttr})
	}
	if len(lp.Tmpfs) > 0 {
		linuxProps["tmpfs"] = types.ListValueMust(types.ObjectType{AttrTypes: tmpfsAttr}, flattenContainerTmpfs(ctx, lp.Tmpfs))
	} else {
		linuxProps["tmpfs"] = types.ListNull(types.ObjectType{AttrTypes: tmpfsAttr})
	}
	linuxProps["linux_parameters"] = types.ObjectValueMust(linuxParametersAttr, linuxProps)
	return linuxProps
}

func frameworkFlattenContainerDevices(ctx context.Context, devices []batchtypes.Device) (data []attr.Value) {
	for _, dev := range devices {
		var perms []string
		for _, perm := range dev.Permissions {
			perms = append(perms, string(perm))
		}
		data = append(data, types.ObjectValueMust(deviceAttr, map[string]attr.Value{
			"host_path":      flex.StringToFramework(ctx, dev.HostPath),
			"container_path": flex.StringToFramework(ctx, dev.ContainerPath),
			"permissions":    flex.FlattenFrameworkStringList(ctx, aws.StringSlice(perms)),
		}))
	}
	return
}

func flattenContainerTmpfs(ctx context.Context, tmpfs []batchtypes.Tmpfs) (data []attr.Value) {
	for _, tmp := range tmpfs {
		data = append(data, types.ObjectValueMust(tmpfsAttr, map[string]attr.Value{
			"container_path": flex.StringToFramework(ctx, tmp.ContainerPath),
			"size":           flex.Int32ToFramework(ctx, tmp.Size),
			"mount_options":  flex.FlattenFrameworkStringList(ctx, aws.StringSlice(tmp.MountOptions)),
		}))
	}
	return
}

func frameworkFlattenRetryStrategy(ctx context.Context, jd *batchtypes.RetryStrategy, data *dataSourceJobDefinitionData) (diags diag.Diagnostics) {
	att := fwtypes.AttributeTypesMust[retryStrategy](ctx)
	att["evaluate_on_exit"] = types.ListType{ElemType: types.ObjectType{AttrTypes: evaluateOnExitAttr}}
	if jd == nil {
		data.RetryStrategy = types.ObjectNull(att)
		return
	}

	var elems []attr.Value
	for _, apiObject := range jd.EvaluateOnExit {
		obj := map[string]attr.Value{
			"action":           flex.StringToFramework(ctx, aws.String(string(apiObject.Action))),
			"on_exit_code":     flex.StringToFramework(ctx, apiObject.OnExitCode),
			"on_reason":        flex.StringToFramework(ctx, apiObject.OnReason),
			"on_status_reason": flex.StringToFramework(ctx, apiObject.OnStatusReason),
		}
		elems = append(elems, types.ObjectValueMust(evaluateOnExitAttr, obj))
	}

	if elems == nil {
		data.RetryStrategy = types.ObjectValueMust(att, map[string]attr.Value{
			"attempts":         flex.Int32ToFramework(ctx, jd.Attempts),
			"evaluate_on_exit": types.ListNull(types.ObjectType{AttrTypes: evaluateOnExitAttr}),
		})
	} else {
		data.RetryStrategy = types.ObjectValueMust(att, map[string]attr.Value{
			"attempts":         flex.Int32ToFramework(ctx, jd.Attempts),
			"evaluate_on_exit": types.ListValueMust(types.ObjectType{AttrTypes: evaluateOnExitAttr}, elems),
		})
	}
	return diags
}

type dataSourceJobDefinitionData struct {
	ARN                        fwtypes.ARN  `tfsdk:"arn"`
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Revision                   types.Int64  `tfsdk:"revision"`
	Status                     types.String `tfsdk:"status"`
	Tags                       types.Map    `tfsdk:"tags"`
	Type                       types.String `tfsdk:"type"`
	ContainerOrchestrationType types.String `tfsdk:"container_orchestration_type"`
	SchedulingPriority         types.Int64  `tfsdk:"scheduling_priority"`
	RetryStrategy              types.Object `tfsdk:"retry_strategy"`
	Timeout                    types.Object `tfsdk:"timeout"`
	NodeProperties             types.Object `tfsdk:"node_properties"`
	EksProperties              types.Object `tfsdk:"eks_properties"`
}

type retryStrategy struct {
	Attempts       types.Int64  `tfsdk:"attempts"`
	EvaluateOnExit types.Object `tfsdk:"evaluate_on_exit"`
}

var timeoutAttr = map[string]attr.Type{
	"attempt_duration_seconds": types.Int64Type,
}

var eksPropertiesAttr = map[string]attr.Type{
	"pod_properties": types.ObjectType{AttrTypes: eksPodPropertiesAttr},
}

var eksPodPropertiesAttr = map[string]attr.Type{
	"containers":           types.ListType{ElemType: types.ObjectType{AttrTypes: eksContainerAttr}},
	"dns_policy":           types.StringType,
	"host_network":         types.BoolType,
	"metadata":             types.ObjectType{AttrTypes: eksMetadataAttr},
	"service_account_name": types.StringType,
	"volumes":              types.ListType{ElemType: types.ObjectType{AttrTypes: eksVolumeAttr}},
}

var eksContainerAttr = map[string]attr.Type{
	"args":              types.ListType{ElemType: types.StringType},
	"commands":          types.ListType{ElemType: types.StringType},
	"env":               types.ListType{ElemType: types.ObjectType{AttrTypes: eksContainerEnvironmentVariableAttr}},
	"image":             types.StringType,
	"image_pull_policy": types.StringType,
	"name":              types.StringType,
	"resources":         types.ObjectType{AttrTypes: eksContainerResourceRequirementsAttr},
	"security_context":  types.ObjectType{AttrTypes: eksContainerSecurityContextAttr},
	"volume_mounts":     types.ListType{ElemType: types.ObjectType{AttrTypes: eksContainerVolumeMountAttr}},
}

var eksContainerEnvironmentVariableAttr = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}

var eksContainerResourceRequirementsAttr = map[string]attr.Type{
	"limits":   types.MapType{ElemType: types.StringType},
	"requests": types.MapType{ElemType: types.StringType},
}

var eksContainerSecurityContextAttr = map[string]attr.Type{
	"privileged":                types.BoolType,
	"run_as_user":               types.Int64Type,
	"run_as_group":              types.Int64Type,
	"run_as_non_root":           types.BoolType,
	"read_only_root_filesystem": types.BoolType,
}

var eksContainerVolumeMountAttr = map[string]attr.Type{
	"mount_path": types.StringType,
	"name":       types.StringType,
	"read_only":  types.BoolType,
}

var eksMetadataAttr = map[string]attr.Type{
	"labels": types.MapType{ElemType: types.StringType},
}

var eksVolumeAttr = map[string]attr.Type{
	"name":      types.StringType,
	"empty_dir": types.ObjectType{AttrTypes: eksVolumeEmptyDirAttr},
	"host_path": types.ObjectType{AttrTypes: eksVolumeHostPathAttr},
	"secret":    types.ObjectType{AttrTypes: eksVolumeSecretAttr},
}

var eksVolumeEmptyDirAttr = map[string]attr.Type{
	"medium":     types.StringType,
	"size_limit": types.Int64Type,
}

var eksVolumeHostPathAttr = map[string]attr.Type{
	"path": types.StringType,
}

var eksVolumeSecretAttr = map[string]attr.Type{
	"secret_name": types.StringType,
	"optional":    types.BoolType,
}

type frameworkNodeProperties struct {
	MainNode            types.Int64 `tfsdk:"main_node"`
	NodeRangeProperties types.List  `tfsdk:"node_range_properties"`
	NumNodes            types.Int64 `tfsdk:"num_nodes"`
}

var evaluateOnExitAttr = map[string]attr.Type{
	"action":           types.StringType,
	"on_exit_code":     types.StringType,
	"on_reason":        types.StringType,
	"on_status_reason": types.StringType,
}

var nodeRangePropertiesAttr = map[string]attr.Type{
	"container":    types.ObjectType{AttrTypes: containerPropertiesAttr},
	"target_nodes": types.StringType,
}

var containerPropertiesAttr = map[string]attr.Type{
	"command":                        types.ListType{ElemType: types.StringType},
	"environment":                    types.ListType{ElemType: types.ObjectType{AttrTypes: keyValuePairAttr}},
	"ephemeral_storage":              types.ObjectType{AttrTypes: ephemeralStorageAttr},
	"execution_role_arn":             types.StringType,
	"fargate_platform_configuration": types.ObjectType{AttrTypes: fargatePlatformConfigurationAttr},
	"image":                          types.StringType,
	"instance_type":                  types.StringType,
	"job_role_arn":                   types.StringType,
	"linux_parameters":               types.ObjectType{AttrTypes: linuxParametersAttr},
	"log_configuration":              types.ObjectType{AttrTypes: logConfigurationAttr},
	"mount_points":                   types.ListType{ElemType: types.ObjectType{AttrTypes: mountPointAttr}},
	"network_configuration":          types.ObjectType{AttrTypes: networkConfigurationAttr},
	"privileged":                     types.BoolType,
	"readonly_root_filesystem":       types.BoolType,
	"resource_requirements":          types.ListType{ElemType: types.ObjectType{AttrTypes: resourceRequirementsAttr}},
	"runtime_platform":               types.ObjectType{AttrTypes: runtimePlatformAttr},
	"secrets":                        types.ListType{ElemType: types.ObjectType{AttrTypes: secretAttr}},
	"ulimits":                        types.ListType{ElemType: types.ObjectType{AttrTypes: ulimitsAttr}},
	"user":                           types.StringType,
	"volumes":                        types.ListType{ElemType: types.ObjectType{AttrTypes: volumeAttr}},
}

var keyValuePairAttr = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}

var ephemeralStorageAttr = map[string]attr.Type{
	"size_in_gib": types.Int64Type,
}

var fargatePlatformConfigurationAttr = map[string]attr.Type{
	"platform_version": types.StringType,
}

var linuxParametersAttr = map[string]attr.Type{
	"devices":              types.ListType{ElemType: types.ObjectType{AttrTypes: deviceAttr}},
	"init_process_enabled": types.BoolType,
	"max_swap":             types.Int64Type,
	"shared_memory_size":   types.Int64Type,
	"swappiness":           types.Int64Type,
	"tmpfs":                types.ListType{ElemType: types.ObjectType{AttrTypes: tmpfsAttr}},
}

var logConfigurationAttr = map[string]attr.Type{
	"options":        types.MapType{ElemType: types.StringType},
	"secret_options": types.ListType{ElemType: types.ObjectType{AttrTypes: secretAttr}},
	"log_driver":     types.StringType,
}
var tmpfsAttr = map[string]attr.Type{
	"container_path": types.StringType,
	"mount_options":  types.ListType{ElemType: types.StringType},
	"size":           types.Int64Type,
}

var deviceAttr = map[string]attr.Type{
	"container_path": types.StringType,
	"host_path":      types.StringType,
	"permissions":    types.ListType{ElemType: types.StringType},
}

var mountPointAttr = map[string]attr.Type{
	"container_path": types.StringType,
	"read_only":      types.BoolType,
	"source_volume":  types.StringType,
}

var networkConfigurationAttr = map[string]attr.Type{
	"assign_public_ip": types.StringType,
}

var resourceRequirementsAttr = map[string]attr.Type{
	"type":  types.StringType,
	"value": types.StringType,
}

var runtimePlatformAttr = map[string]attr.Type{
	"cpu_architecture":        types.StringType,
	"operating_system_family": types.StringType,
}

var secretAttr = map[string]attr.Type{
	"name":       types.StringType,
	"value_from": types.StringType,
}

var ulimitsAttr = map[string]attr.Type{
	"hard_limit": types.Int64Type,
	"name":       types.StringType,
	"soft_limit": types.Int64Type,
}

var volumeAttr = map[string]attr.Type{
	"efs_volume_configuration": types.ObjectType{AttrTypes: efsVolumeConfigurationAttr},
	"host":                     types.ObjectType{AttrTypes: hostAttr},
	"name":                     types.StringType,
}

var efsVolumeConfigurationAttr = map[string]attr.Type{
	"authorization_config":    types.ObjectType{AttrTypes: authorizationConfigAttr},
	"file_system_id":          types.StringType,
	"root_directory":          types.StringType,
	"transit_encryption":      types.StringType,
	"transit_encryption_port": types.Int64Type,
}

var authorizationConfigAttr = map[string]attr.Type{
	"access_point_id": types.StringType,
	"iam":             types.StringType,
}

var hostAttr = map[string]attr.Type{
	"source_path": types.StringType,
}
