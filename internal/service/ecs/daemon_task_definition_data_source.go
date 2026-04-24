// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ecs_daemon_task_definition", name="Daemon Task Definition")
func newDaemonTaskDefinitionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &daemonTaskDefinitionDataSource{}, nil
}

type daemonTaskDefinitionDataSource struct {
	framework.DataSourceWithModel[daemonTaskDefinitionDataSourceModel]
}

func (d *daemonTaskDefinitionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"cpu": schema.StringAttribute{
				Computed: true,
			},
			"delete_requested_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrFamily: schema.StringAttribute{
				Computed: true,
			},
			"memory": schema.StringAttribute{
				Computed: true,
			},
			"registered_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"registered_by": schema.StringAttribute{
				Computed: true,
			},
			"revision": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			"task_definition": schema.StringAttribute{
				Required: true,
			},
			"task_role_arn": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"container_definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[containerDefinitionModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"command": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Computed:   true,
						},
						"cpu": schema.Int64Attribute{
							Computed: true,
						},
						"entry_point": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Computed:   true,
						},
						"essential": schema.BoolAttribute{
							Computed: true,
						},
						"image": schema.StringAttribute{
							Computed: true,
						},
						"interactive": schema.BoolAttribute{
							Computed: true,
						},
						"memory": schema.Int64Attribute{
							Computed: true,
						},
						"memory_reservation": schema.Int64Attribute{
							Computed: true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						"privileged": schema.BoolAttribute{
							Computed: true,
						},
						"pseudo_terminal": schema.BoolAttribute{
							Computed: true,
						},
						"readonly_root_filesystem": schema.BoolAttribute{
							Computed: true,
						},
						"start_timeout": schema.Int64Attribute{
							Computed: true,
						},
						"stop_timeout": schema.Int64Attribute{
							Computed: true,
						},
						"user": schema.StringAttribute{
							Computed: true,
						},
						"working_directory": schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"depends_on": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerDependencyModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"condition": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ContainerCondition](),
										Computed:   true,
									},
									"container_name": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"environment": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[keyValuePairModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Computed: true,
									},
									names.AttrValue: schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"environment_file": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[environmentFileModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.EnvironmentFileType](),
										Computed:   true,
									},
									names.AttrValue: schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"firelens_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[firelensConfigurationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FirelensConfigurationType](),
										Computed:   true,
									},
									"options": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"health_check": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[healthCheckModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"command": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Computed:   true,
									},
									"interval": schema.Int64Attribute{
										Computed: true,
									},
									"retries": schema.Int64Attribute{
										Computed: true,
									},
									"start_period": schema.Int64Attribute{
										Computed: true,
									},
									"timeout": schema.Int64Attribute{
										Computed: true,
									},
								},
							},
						},
						"linux_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[daemonLinuxParametersModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"init_process_enabled": schema.BoolAttribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"capabilities": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[kernelCapabilitiesModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"add": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Computed:   true,
												},
												"drop": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Computed:   true,
												},
											},
										},
									},
									"device": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[deviceModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"container_path": schema.StringAttribute{
													Computed: true,
												},
												"host_path": schema.StringAttribute{
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
										CustomType: fwtypes.NewListNestedObjectTypeOf[tmpfsModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"container_path": schema.StringAttribute{
													Computed: true,
												},
												"mount_options": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Computed:   true,
												},
												"size": schema.Int64Attribute{
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"log_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[logConfigurationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_driver": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.LogDriver](),
										Computed:   true,
									},
									"options": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Computed:    true,
										ElementType: types.StringType,
									},
								},
								Blocks: map[string]schema.Block{
									"secret_option": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[secretModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrName: schema.StringAttribute{
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
						},
						"mount_point": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mountPointModel](ctx),
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
						"repository_credentials": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[repositoryCredentialsModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_parameter": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"restart_policy": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerRestartPolicyModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrEnabled: schema.BoolAttribute{
										Computed: true,
									},
									"ignored_exit_codes": schema.ListAttribute{
										Computed:    true,
										ElementType: types.Int64Type,
									},
									"restart_attempt_period": schema.Int64Attribute{
										Computed: true,
									},
								},
							},
						},
						"secret": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[secretModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Computed: true,
									},
									"value_from": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"system_control": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[systemControlModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"namespace": schema.StringAttribute{
										Computed: true,
									},
									names.AttrValue: schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"ulimit": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ulimitModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"hard_limit": schema.Int64Attribute{
										Computed: true,
									},
									names.AttrName: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.UlimitName](),
										Computed:   true,
									},
									"soft_limit": schema.Int64Attribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"volume": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[volumeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"host_path": schema.StringAttribute{
							Computed: true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *daemonTaskDefinitionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data daemonTaskDefinitionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECSClient(ctx)

	input := &ecs.DescribeDaemonTaskDefinitionInput{
		DaemonTaskDefinition: aws.String(data.TaskDefinition.ValueString()),
	}

	output, err := conn.DescribeDaemonTaskDefinition(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Task Definition (%s)", data.TaskDefinition.ValueString()), err.Error())
		return
	}

	dtd := output.DaemonTaskDefinition

	// AutoFlex handles DaemonTaskDefinitionArn, Family, Revision, Cpu, Memory,
	// ExecutionRoleArn, TaskRoleArn, Status, RegisteredBy
	response.Diagnostics.Append(fwflex.Flatten(ctx, dtd, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Manual: volumes (structural mismatch — Host.SourcePath → HostPath)
	volumes := make([]volumeModel, len(dtd.Volumes))
	for i, v := range dtd.Volumes {
		volumes[i] = volumeModel{
			HostPath: types.StringPointerValue(v.Host.SourcePath),
			Name:     types.StringPointerValue(v.Name),
		}
	}
	var volumeDiags diag.Diagnostics
	data.Volumes, volumeDiags = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, volumes)
	response.Diagnostics.Append(volumeDiags...)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type daemonTaskDefinitionDataSourceModel struct {
	DaemonTaskDefinitionArn types.String                                              `tfsdk:"arn"`
	ContainerDefinitions    fwtypes.ListNestedObjectValueOf[containerDefinitionModel] `tfsdk:"container_definition"`
	Cpu                     types.String                                              `tfsdk:"cpu"`
	DeleteRequestedAt       timetypes.RFC3339                                         `tfsdk:"delete_requested_at"`
	ExecutionRoleArn        types.String                                              `tfsdk:"execution_role_arn"`
	Family                  types.String                                              `tfsdk:"family"`
	Memory                  types.String                                              `tfsdk:"memory"`
	Region                  types.String                                              `tfsdk:"region"`
	RegisteredAt            timetypes.RFC3339                                         `tfsdk:"registered_at"`
	RegisteredBy            types.String                                              `tfsdk:"registered_by"`
	Revision                types.Int64                                               `tfsdk:"revision"`
	Status                  types.String                                              `tfsdk:"status"`
	TaskDefinition          types.String                                              `tfsdk:"task_definition"`
	TaskRoleArn             types.String                                              `tfsdk:"task_role_arn"`
	Volumes                 fwtypes.SetNestedObjectValueOf[volumeModel]               `tfsdk:"volume"`
}

type volumeModel struct {
	HostPath types.String `tfsdk:"host_path"`
	Name     types.String `tfsdk:"name"`
}
