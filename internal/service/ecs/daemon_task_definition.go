// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

const (
	healthCheckDefaultRetries = 3
	healthCheckDefaultTimeout = 5
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
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"cpu": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
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
			"ipc_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DaemonIpcMode](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"memory": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pid_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DaemonPidMode](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"revision": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DaemonTaskDefinitionStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"task_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"container_definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[containerDefinitionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
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
							Optional: true,
							Computed: true,
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
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(255),
								stringvalidator.RegexMatches(regexache.MustCompile("^[0-9A-Za-z_-]+$"), "must contain only alphanumeric characters, hyphens, and underscores"),
							},
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
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
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
										Optional: true,
									},
									names.AttrValue: schema.StringAttribute{
										Optional: true,
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
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
						"firelens_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[firelensConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"options": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										ElementType: types.StringType,
									},
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FirelensConfigurationType](),
										Required:   true,
									},
								},
							},
						},
						names.AttrHealthCheck: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[healthCheckModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"command": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Required:   true,
									},
									names.AttrInterval: schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(5, 300),
										},
									},
									"retries": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(healthCheckDefaultRetries),
										Validators: []validator.Int64{
											int64validator.Between(1, 10),
										},
									},
									"start_period": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(0, 300),
										},
									},
									names.AttrTimeout: schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Default:  int64default.StaticInt64(healthCheckDefaultTimeout),
										Validators: []validator.Int64{
											int64validator.Between(2, 60),
										},
									},
								},
							},
						},
						"linux_parameters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[daemonLinuxParametersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"init_process_enabled": schema.BoolAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"capabilities": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[kernelCapabilitiesModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
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
													CustomType: fwtypes.ListOfStringEnumType[awstypes.DeviceCgroupPermission](),
													Optional:   true,
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
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
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
										CustomType: fwtypes.NewListNestedObjectTypeOf[daemonSecretModel](ctx),
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
										Optional: true,
									},
								},
							},
						},
						"repository_credentials": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[repositoryCredentialsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_parameter": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
						"restart_policy": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerRestartPolicyModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
									},
									"ignored_exit_codes": schema.ListAttribute{
										Optional:   true,
										CustomType: fwtypes.ListOfInt64Type,
										Validators: []validator.List{
											listvalidator.SizeAtMost(50),
										},
									},
									"restart_attempt_period": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(60, 1800),
										},
									},
								},
							},
						},
						"secret": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[daemonSecretModel](ctx),
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
						names.AttrName: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(255),
								stringvalidator.RegexMatches(regexache.MustCompile("^[0-9A-Za-z_-]+$"), "must contain only alphanumeric characters, hyphens, and underscores"),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"host": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[hostModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"source_path": schema.StringAttribute{
										Optional: true,
									},
								},
							},
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

	output, err := conn.RegisterDaemonTaskDefinition(ctx, &input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(r.Meta().Partition(ctx), err) {
		input.Tags = nil
		output, err = conn.RegisterDaemonTaskDefinition(ctx, &input)
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating ECS Daemon Task Definition (%s)", plan.Family.ValueString()), err.Error())
		return
	}

	if output == nil || output.DaemonTaskDefinitionArn == nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating ECS Daemon Task Definition (%s)", plan.Family.ValueString()), "empty output from API")
		return
	}

	plan.DaemonTaskDefinitionArn = types.StringValue(aws.ToString(output.DaemonTaskDefinitionArn))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		if err := createTags(ctx, conn, aws.ToString(output.DaemonTaskDefinitionArn), tags); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("setting ECS Daemon Task Definition (%s) tags", aws.ToString(output.DaemonTaskDefinitionArn)), err.Error())
			return
		}
	}

	// Read back to populate all computed attributes
	outputFind, err := findDaemonTaskDefinitionByARN(ctx, conn, plan.DaemonTaskDefinitionArn.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Task Definition (%s)", plan.DaemonTaskDefinitionArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputFind, &plan)...)
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

	output, err := findDaemonTaskDefinitionByARN(ctx, conn, state.DaemonTaskDefinitionArn.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Task Definition (%s)", state.DaemonTaskDefinitionArn.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
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

	input := ecs.DeleteDaemonTaskDefinitionInput{
		DaemonTaskDefinition: state.DaemonTaskDefinitionArn.ValueStringPointer(),
	}

	_, err := conn.DeleteDaemonTaskDefinition(ctx, &input)
	if errs.Contains(err, "DaemonTaskDefinitionNotFoundException") || errs.IsAErrorMessageContains[*awstypes.ClientException](err, "not found") || errs.IsAErrorMessageContains[*awstypes.ClientException](err, "being deleted") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting ECS Daemon Task Definition (%s)", state.DaemonTaskDefinitionArn.ValueString()), err.Error())
	}
}

func findDaemonTaskDefinitionByARN(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.DaemonTaskDefinition, error) {
	input := &ecs.DescribeDaemonTaskDefinitionInput{
		DaemonTaskDefinition: aws.String(arn),
	}

	output, err := conn.DescribeDaemonTaskDefinition(ctx, input)

	if errs.Contains(err, "DaemonTaskDefinitionNotFoundException") || errs.Contains(err, "not found") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DaemonTaskDefinition == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if output.DaemonTaskDefinition.Status == awstypes.DaemonTaskDefinitionStatusDeleteInProgress || output.DaemonTaskDefinition.Status == awstypes.DaemonTaskDefinitionStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: string(output.DaemonTaskDefinition.Status),
		}
	}

	return output.DaemonTaskDefinition, nil
}

type daemonTaskDefinitionResourceModel struct {
	framework.WithRegionModel
	DaemonTaskDefinitionArn types.String                                              `tfsdk:"arn"`
	ContainerDefinitions    fwtypes.ListNestedObjectValueOf[containerDefinitionModel] `tfsdk:"container_definition"`
	Cpu                     types.String                                              `tfsdk:"cpu"`
	ExecutionRoleArn        fwtypes.ARN                                               `tfsdk:"execution_role_arn"`
	Family                  types.String                                              `tfsdk:"family"`
	IpcMode                 fwtypes.StringEnum[awstypes.DaemonIpcMode]                `tfsdk:"ipc_mode"`
	Memory                  types.String                                              `tfsdk:"memory"`
	PidMode                 fwtypes.StringEnum[awstypes.DaemonPidMode]                `tfsdk:"pid_mode"`
	Revision                types.Int64                                               `tfsdk:"revision"`
	Status                  fwtypes.StringEnum[awstypes.DaemonTaskDefinitionStatus]   `tfsdk:"status"`
	Tags                    tftags.Map                                                `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                `tfsdk:"tags_all"`
	TaskRoleArn             fwtypes.ARN                                               `tfsdk:"task_role_arn"`
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
	Secrets                fwtypes.ListNestedObjectValueOf[daemonSecretModel]           `tfsdk:"secret"`
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
	Value fwtypes.ARN                                      `tfsdk:"value"`
}

type firelensConfigurationModel struct {
	Options fwtypes.MapOfString                                    `tfsdk:"options"`
	Type    fwtypes.StringEnum[awstypes.FirelensConfigurationType] `tfsdk:"type"`
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
	LogDriver     fwtypes.StringEnum[awstypes.LogDriver]             `tfsdk:"log_driver"`
	Options       fwtypes.MapOfString                                `tfsdk:"options"`
	SecretOptions fwtypes.ListNestedObjectValueOf[daemonSecretModel] `tfsdk:"secret_option"`
}

type mountPointModel struct {
	ContainerPath types.String `tfsdk:"container_path"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	SourceVolume  types.String `tfsdk:"source_volume"`
}

type repositoryCredentialsModel struct {
	CredentialsParameter fwtypes.ARN `tfsdk:"credentials_parameter"`
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

type volumeModel struct {
	Host fwtypes.ListNestedObjectValueOf[hostModel] `tfsdk:"host" autoflex:",omitempty"`
	Name types.String                               `tfsdk:"name"`
}

type hostModel struct {
	SourcePath types.String `tfsdk:"source_path"`
}

type daemonSecretModel struct {
	Name      types.String `tfsdk:"name"`
	ValueFrom types.String `tfsdk:"value_from"`
}
