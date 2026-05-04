// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecs_daemon_task_definition", name="Daemon Task Definition")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
func newDaemonTaskDefinitionResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &daemonTaskDefinitionResource{}, nil
}

type daemonTaskDefinitionResource struct {
	framework.ResourceWithModel[daemonTaskDefinitionResourceModel]
	framework.WithImportByIdentity
}

func (r *daemonTaskDefinitionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delete_requested_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrFamily: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(regexache.MustCompile("^[0-9A-Za-z_-]+$"), "must contain only alphanumeric characters, hyphens, and underscores"),
				},
			},
			"memory": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registered_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registered_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"revision": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"task_role_arn": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"container_definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[containerDefinitionModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"command": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Optional:   true,
						},
						"cpu": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"entry_point": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Optional:   true,
						},
						"essential": schema.BoolAttribute{
							Required: true,
						},
						"image": schema.StringAttribute{
							Required: true,
						},
						"interactive": schema.BoolAttribute{
							Optional: true,
						},
						"memory": schema.Int64Attribute{
							Optional: true,
						},
						"memory_reservation": schema.Int64Attribute{
							Optional: true,
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						"privileged": schema.BoolAttribute{
							Optional: true,
						},
						"pseudo_terminal": schema.BoolAttribute{
							Optional: true,
						},
						"readonly_root_filesystem": schema.BoolAttribute{
							Optional: true,
						},
						"start_timeout": schema.Int64Attribute{
							Optional: true,
						},
						"stop_timeout": schema.Int64Attribute{
							Optional: true,
						},
						"user": schema.StringAttribute{
							Optional: true,
						},
						"working_directory": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"depends_on": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerDependencyModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrCondition: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ContainerCondition](),
										Required:   true,
									},
									"container_name": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						names.AttrEnvironment: schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[keyValuePairModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
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
										Required:   true,
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
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
										Required:   true,
									},
									"options": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						names.AttrHealthCheck: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[healthCheckModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"command": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Required:   true,
									},
									names.AttrInterval: schema.Int64Attribute{
										Optional: true,
									},
									"retries": schema.Int64Attribute{
										Optional: true,
									},
									"start_period": schema.Int64Attribute{
										Optional: true,
									},
									names.AttrTimeout: schema.Int64Attribute{
										Optional: true,
									},
								},
							},
						},
						"linux_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[daemonLinuxParametersModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"init_process_enabled": schema.BoolAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"capabilities": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[kernelCapabilitiesModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"add": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Optional:   true,
												},
												"drop": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Optional:   true,
												},
											},
										},
									},
									"device": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[deviceModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"container_path": schema.StringAttribute{
													Optional: true,
												},
												"host_path": schema.StringAttribute{
													Required: true,
												},
												names.AttrPermissions: schema.ListAttribute{
													Optional:    true,
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
													Required: true,
												},
												"mount_options": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Optional:   true,
												},
												names.AttrSize: schema.Int64Attribute{
													Required: true,
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
										Required:   true,
									},
									"options": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
								},
								Blocks: map[string]schema.Block{
									"secret_option": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[secretModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrName: schema.StringAttribute{
													Required: true,
												},
												"value_from": schema.StringAttribute{
													Required: true,
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
										Optional: true,
									},
									"read_only": schema.BoolAttribute{
										Optional: true,
									},
									"source_volume": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"repository_credentials": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[repositoryCredentialsModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_parameter": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"restart_policy": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerRestartPolicyModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrEnabled: schema.BoolAttribute{
										Optional: true,
									},
									"ignored_exit_codes": schema.ListAttribute{
										Optional:    true,
										ElementType: types.Int64Type,
									},
									"restart_attempt_period": schema.Int64Attribute{
										Optional: true,
									},
								},
							},
						},
						"secret": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[secretModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									"value_from": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"system_control": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[systemControlModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrNamespace: schema.StringAttribute{
										Optional: true,
									},
									names.AttrValue: schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"ulimit": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ulimitModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"hard_limit": schema.Int64Attribute{
										Required: true,
									},
									names.AttrName: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.UlimitName](),
										Required:   true,
									},
									"soft_limit": schema.Int64Attribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"volume": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[volumeModel](ctx),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"host_path": schema.StringAttribute{
							Optional: true,
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *daemonTaskDefinitionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan daemonTaskDefinitionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	var input ecs.RegisterDaemonTaskDefinitionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Fields AutoFlex can't handle
	input.Tags = getTagsIn(ctx)
	if !plan.Volumes.IsNull() && !plan.Volumes.IsUnknown() {
		volumeSlice, diags := plan.Volumes.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		input.Volumes = expandDaemonVolumesFromModel(volumeSlice)
	}

	output, err := conn.RegisterDaemonTaskDefinition(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating ECS Daemon Task Definition (%s)", plan.Family.ValueString()), err.Error())
		return
	}

	plan.DaemonTaskDefinitionArn = types.StringPointerValue(output.DaemonTaskDefinitionArn)

	// Read back to populate all computed attributes
	dtd, err := findDaemonTaskDefinitionByARN(ctx, conn, plan.DaemonTaskDefinitionArn.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Task Definition (%s)", plan.DaemonTaskDefinitionArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(flattenDaemonTaskDefinition(ctx, dtd, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *daemonTaskDefinitionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state daemonTaskDefinitionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	dtd, err := findDaemonTaskDefinitionByARN(ctx, conn, state.DaemonTaskDefinitionArn.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Task Definition (%s)", state.DaemonTaskDefinitionArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(flattenDaemonTaskDefinition(ctx, dtd, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *daemonTaskDefinitionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state daemonTaskDefinitionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	log.Printf("[DEBUG] Deleting ECS Daemon Task Definition: %s", state.DaemonTaskDefinitionArn.ValueString())

	_, err := conn.DeleteDaemonTaskDefinition(ctx, &ecs.DeleteDaemonTaskDefinitionInput{
		DaemonTaskDefinition: state.DaemonTaskDefinitionArn.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ClientException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting ECS Daemon Task Definition (%s)", state.DaemonTaskDefinitionArn.ValueString()), err.Error())
	}
}

// flattenDaemonTaskDefinition populates the model from a DaemonTaskDefinition using AutoFlex
// for matching fields and manual handling for fields that require transformation.
func flattenDaemonTaskDefinition(ctx context.Context, dtd *awstypes.DaemonTaskDefinition, model *daemonTaskDefinitionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// AutoFlex handles DaemonTaskDefinitionArn, Family, Cpu, Memory, ExecutionRoleArn,
	// TaskRoleArn, Revision, Status, RegisteredBy, ContainerDefinitions
	diags.Append(fwflex.Flatten(ctx, dtd, model)...)
	if diags.HasError() {
		return diags
	}

	// Manual: volumes (structural mismatch — Host.SourcePath → HostPath)
	model.Volumes, diags = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, flattenDaemonVolumes(dtd.Volumes))

	return diags
}

// flattenDaemonVolumes converts SDK Volume types to the Terraform model.
// AutoFlex cannot handle this because the SDK nests SourcePath under Host
// (Volume.Host.SourcePath) while the Terraform model flattens it (volumeModel.HostPath).
func flattenDaemonVolumes(volumes []awstypes.DaemonVolume) []volumeModel {
	models := make([]volumeModel, len(volumes))
	for i, v := range volumes {
		models[i] = volumeModel{
			Name: types.StringPointerValue(v.Name),
		}
		if v.Host != nil && v.Host.SourcePath != nil {
			models[i].HostPath = types.StringPointerValue(v.Host.SourcePath)
		}
	}
	return models
}

func expandDaemonVolumesFromModel(volumes []*volumeModel) []awstypes.DaemonVolume {
	if len(volumes) == 0 {
		return nil
	}

	var apiObjects []awstypes.DaemonVolume
	for _, v := range volumes {
		apiObject := awstypes.DaemonVolume{
			Name: v.Name.ValueStringPointer(),
		}
		if !v.HostPath.IsNull() && v.HostPath.ValueString() != "" {
			apiObject.Host = &awstypes.HostVolumeProperties{
				SourcePath: v.HostPath.ValueStringPointer(),
			}
		}
		apiObjects = append(apiObjects, apiObject)
	}
	return apiObjects
}

// Helper functions used by both resource and data sources.

type daemonTaskDefinitionResourceModel struct {
	framework.WithRegionModel
	DaemonTaskDefinitionArn types.String                                              `tfsdk:"arn"`
	ContainerDefinitions    fwtypes.ListNestedObjectValueOf[containerDefinitionModel] `tfsdk:"container_definition"`
	Cpu                     types.String                                              `tfsdk:"cpu"`
	DeleteRequestedAt       timetypes.RFC3339                                         `tfsdk:"delete_requested_at"`
	ExecutionRoleArn        types.String                                              `tfsdk:"execution_role_arn"`
	Family                  types.String                                              `tfsdk:"family"`
	Memory                  types.String                                              `tfsdk:"memory"`
	RegisteredAt            timetypes.RFC3339                                         `tfsdk:"registered_at"`
	RegisteredBy            types.String                                              `tfsdk:"registered_by"`
	Revision                types.Int64                                               `tfsdk:"revision"`
	Status                  types.String                                              `tfsdk:"status"`
	Tags                    tftags.Map                                                `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                `tfsdk:"tags_all"`
	TaskRoleArn             types.String                                              `tfsdk:"task_role_arn"`
	Volumes                 fwtypes.SetNestedObjectValueOf[volumeModel]               `tfsdk:"volume"`
}

type containerDefinitionModel struct {
	Command                fwtypes.ListOfString                                         `tfsdk:"command"`
	Cpu                    types.Int64                                                  `tfsdk:"cpu"`
	DependsOn              fwtypes.ListNestedObjectValueOf[containerDependencyModel]    `tfsdk:"depends_on"`
	EntryPoint             fwtypes.ListOfString                                         `tfsdk:"entry_point"`
	Environment            fwtypes.SetNestedObjectValueOf[keyValuePairModel]            `tfsdk:"environment"`
	EnvironmentFiles       fwtypes.ListNestedObjectValueOf[environmentFileModel]        `tfsdk:"environment_file"`
	Essential              types.Bool                                                   `tfsdk:"essential"`
	FirelensConfiguration  fwtypes.ListNestedObjectValueOf[firelensConfigurationModel]  `tfsdk:"firelens_configuration"`
	HealthCheck            fwtypes.ListNestedObjectValueOf[healthCheckModel]            `tfsdk:"health_check"`
	Image                  types.String                                                 `tfsdk:"image"`
	Interactive            types.Bool                                                   `tfsdk:"interactive"`
	LinuxParameters        fwtypes.ListNestedObjectValueOf[daemonLinuxParametersModel]  `tfsdk:"linux_parameters"`
	LogConfiguration       fwtypes.ListNestedObjectValueOf[logConfigurationModel]       `tfsdk:"log_configuration"`
	Memory                 types.Int64                                                  `tfsdk:"memory"`
	MemoryReservation      types.Int64                                                  `tfsdk:"memory_reservation"`
	MountPoints            fwtypes.ListNestedObjectValueOf[mountPointModel]             `tfsdk:"mount_point"`
	Name                   types.String                                                 `tfsdk:"name"`
	Privileged             types.Bool                                                   `tfsdk:"privileged"`
	PseudoTerminal         types.Bool                                                   `tfsdk:"pseudo_terminal"`
	ReadonlyRootFilesystem types.Bool                                                   `tfsdk:"readonly_root_filesystem"`
	RepositoryCredentials  fwtypes.ListNestedObjectValueOf[repositoryCredentialsModel]  `tfsdk:"repository_credentials"`
	RestartPolicy          fwtypes.ListNestedObjectValueOf[containerRestartPolicyModel] `tfsdk:"restart_policy"`
	Secrets                fwtypes.ListNestedObjectValueOf[secretModel]                 `tfsdk:"secret"`
	StartTimeout           types.Int64                                                  `tfsdk:"start_timeout"`
	StopTimeout            types.Int64                                                  `tfsdk:"stop_timeout"`
	SystemControls         fwtypes.ListNestedObjectValueOf[systemControlModel]          `tfsdk:"system_control"`
	Ulimits                fwtypes.ListNestedObjectValueOf[ulimitModel]                 `tfsdk:"ulimit"`
	User                   types.String                                                 `tfsdk:"user"`
	WorkingDirectory       types.String                                                 `tfsdk:"working_directory"`
}

type containerDependencyModel struct {
	Condition     fwtypes.StringEnum[awstypes.ContainerCondition] `tfsdk:"condition"`
	ContainerName types.String                                    `tfsdk:"container_name"`
}

type environmentFileModel struct {
	Type  fwtypes.StringEnum[awstypes.EnvironmentFileType] `tfsdk:"type"`
	Value types.String                                     `tfsdk:"value"`
}

type firelensConfigurationModel struct {
	Type    fwtypes.StringEnum[awstypes.FirelensConfigurationType] `tfsdk:"type"`
	Options fwtypes.MapOfString                                    `tfsdk:"options"`
}

type healthCheckModel struct {
	Command     fwtypes.ListOfString `tfsdk:"command"`
	Interval    types.Int64          `tfsdk:"interval"`
	Retries     types.Int64          `tfsdk:"retries"`
	StartPeriod types.Int64          `tfsdk:"start_period"`
	Timeout     types.Int64          `tfsdk:"timeout"`
}

type daemonLinuxParametersModel struct {
	Capabilities       fwtypes.ListNestedObjectValueOf[kernelCapabilitiesModel] `tfsdk:"capabilities"`
	Devices            fwtypes.ListNestedObjectValueOf[deviceModel]             `tfsdk:"device"`
	InitProcessEnabled types.Bool                                               `tfsdk:"init_process_enabled"`
	Tmpfs              fwtypes.ListNestedObjectValueOf[tmpfsModel]              `tfsdk:"tmpfs"`
}

type kernelCapabilitiesModel struct {
	Add  fwtypes.ListOfString `tfsdk:"add"`
	Drop fwtypes.ListOfString `tfsdk:"drop"`
}

type deviceModel struct {
	ContainerPath types.String                                                             `tfsdk:"container_path"`
	HostPath      types.String                                                             `tfsdk:"host_path"`
	Permissions   fwtypes.ListValueOf[fwtypes.StringEnum[awstypes.DeviceCgroupPermission]] `tfsdk:"permissions"`
}

type tmpfsModel struct {
	ContainerPath types.String         `tfsdk:"container_path"`
	MountOptions  fwtypes.ListOfString `tfsdk:"mount_options"`
	Size          types.Int64          `tfsdk:"size"`
}

type logConfigurationModel struct {
	LogDriver     fwtypes.StringEnum[awstypes.LogDriver]       `tfsdk:"log_driver"`
	Options       fwtypes.MapOfString                          `tfsdk:"options"`
	SecretOptions fwtypes.ListNestedObjectValueOf[secretModel] `tfsdk:"secret_option"`
}

type mountPointModel struct {
	ContainerPath types.String `tfsdk:"container_path"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	SourceVolume  types.String `tfsdk:"source_volume"`
}

type repositoryCredentialsModel struct {
	CredentialsParameter types.String `tfsdk:"credentials_parameter"`
}

type containerRestartPolicyModel struct {
	Enabled              types.Bool          `tfsdk:"enabled"`
	IgnoredExitCodes     fwtypes.ListOfInt64 `tfsdk:"ignored_exit_codes"`
	RestartAttemptPeriod types.Int64         `tfsdk:"restart_attempt_period"`
}

type systemControlModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Value     types.String `tfsdk:"value"`
}

type ulimitModel struct {
	HardLimit types.Int64                             `tfsdk:"hard_limit"`
	Name      fwtypes.StringEnum[awstypes.UlimitName] `tfsdk:"name"`
	SoftLimit types.Int64                             `tfsdk:"soft_limit"`
}

func findDaemonTaskDefinitionByARN(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.DaemonTaskDefinition, error) {
	input := &ecs.DescribeDaemonTaskDefinitionInput{
		DaemonTaskDefinition: aws.String(arn),
	}

	output, err := conn.DescribeDaemonTaskDefinition(ctx, input)

	if errs.Contains(err, "DaemonTaskDefinitionNotFoundException") || errs.Contains(err, "not found") {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DaemonTaskDefinition == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if output.DaemonTaskDefinition.Status == awstypes.DaemonTaskDefinitionStatusDeleteInProgress || output.DaemonTaskDefinition.Status == awstypes.DaemonTaskDefinitionStatusDeleted {
		return nil, &sdkretry.NotFoundError{
			Message:     string(output.DaemonTaskDefinition.Status),
			LastRequest: input,
		}
	}

	return output.DaemonTaskDefinition, nil
}

func findDaemonTaskDefinitions(ctx context.Context, conn *ecs.Client, input *ecs.ListDaemonTaskDefinitionsInput) ([]awstypes.DaemonTaskDefinitionSummary, error) {
	var result []awstypes.DaemonTaskDefinitionSummary

	err := listDaemonTaskDefinitionsPages(ctx, conn, input, func(page *ecs.ListDaemonTaskDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		result = append(result, page.DaemonTaskDefinitions...)
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
