// Copyright (c) 2025 Altos Labs, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"errors"
	"maps"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func newResourceCluster(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCluster{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameCluster = "Cluster"
)

type resourceCluster struct {
	framework.ResourceWithModel[resourceClusterModel]
	framework.WithTimeouts
}

func (r *resourceCluster) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	var vpcConfigModelAttributeSchema = map[string]schema.Attribute{
		"security_group_ids": schema.SetAttribute{
			CustomType: fwtypes.SetOfStringType,
			Required:   true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(1, 5),
				setvalidator.ValueStringsAre(
					stringvalidator.RegexMatches(regexache.MustCompile(`[-0-9a-zA-Z]+`), `must match [-0-9a-zA-Z]+`),
					stringvalidator.LengthAtMost(32),
				),
			},
		},
		"subnets": schema.SetAttribute{
			CustomType: fwtypes.SetOfStringType,
			Required:   true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
			Validators: []validator.Set{
				setvalidator.SizeBetween(1, 16),
				setvalidator.ValueStringsAre(
					stringvalidator.RegexMatches(regexache.MustCompile(`[-0-9a-zA-Z]+`), `must match [-0-9a-zA-Z]+`),
					stringvalidator.LengthAtMost(32),
				),
			},
		},
	}
	var capacitySizeConfigModelAttributeSchema = map[string]schema.Attribute{
		"type": schema.StringAttribute{
			CustomType: fwtypes.StringEnumType[awstypes.NodeUnavailabilityType](),
			Required:   true,
		},
		"value": schema.Int32Attribute{
			Required: true,
			Validators: []validator.Int32{
				int32validator.AtLeast(1),
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"cluster_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "must match ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$"),
				},
			},
			"cluster_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ClusterStatus](),
				Computed:   true,
			},
			"node_recovery": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ClusterNodeRecovery](),
				Optional:   true,
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"instance_group": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[clusterInstanceGroupSpecificationModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeBetween(1, 100),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplaceIf(instanceGroupReplaceIf, "Replace instance group diff", "Replace instance group diff"),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"execution_role": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[a-z\-]*:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+$`), `must match ^arn:aws[a-z\-]*:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+$`),
							},
						},
						"instance_count": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(0, 6758),
							},
						},
						"instance_group_name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 63),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "must match ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$"),
							},
						},
						"instance_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ClusterInstanceType](),
							Required:   true,
						},
						"nodes": schema.ListAttribute{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[clusterNodeDetailsModel](ctx),
							Computed:    true,
							ElementType: fwtypes.NewObjectTypeOf[clusterNodeDetailsModel](ctx),
						},
						"on_start_deep_healthchecks": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringEnumType[awstypes.DeepHealthCheckType](),
							Optional:   true,
							Validators: []validator.List{
								listvalidator.AlsoRequires(
									path.MatchRoot("orchestrator").AtAnyListIndex().AtName("eks"),
								),
							},
						},
						"threads_per_core": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Int32{
								int32validator.OneOf(1, 2),
							},
						},
						"training_plan_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:training-plan/.*`), `must match arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:training-plan/.*`),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"instance_storage_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[clusterInstanceStorageConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"ebs_volume_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[clusterEbsVolumeConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"volume_size_in_gb": schema.Int32Attribute{
													Required: true,
													Validators: []validator.Int32{
														int32validator.Between(1, 16384),
													},
												},
											},
										},
									},
								},
							},
						},
						"lifecycle_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[clusterLifeCycleConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"on_create": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[\S\s]+$`), `must match ^[\S\s]+$`),
										},
									},
									"source_s3_uri": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(1024),
											stringvalidator.RegexMatches(regexache.MustCompile(`^(https|s3)://([^/]+)/?(.*)$`), `must match ^(https|s3)://([^/]+)/?(.*)$`),
										},
									},
								},
							},
						},
						"override_vpc_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: vpcConfigModelAttributeSchema,
							},
						},
						"scheduled_update_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[scheduledUpdateConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"schedule_expression": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"deployment_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[deploymentConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"wait_interval": schema.Int32Attribute{
													Optional: true,
													Validators: []validator.Int32{
														int32validator.Between(0, 3600),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"auto_rollback_configuration": schema.SetNestedBlock{
													CustomType: fwtypes.NewSetNestedObjectTypeOf[alarmDetailsModel](ctx),
													Validators: []validator.Set{
														setvalidator.SizeBetween(1, 10),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"alarm_name": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 255),
																	stringvalidator.RegexMatches(regexache.MustCompile(`\S`), `must not contain whitespace`),
																},
															},
														},
													},
												},
												"rolling_update_policy": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[rollingDeploymentPolicyModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"maximum_batch_size": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[capacitySizeConfigModel](ctx),
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: capacitySizeConfigModelAttributeSchema,
																},
															},
															"rollback_maximum_batch_size": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[capacitySizeConfigModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: capacitySizeConfigModelAttributeSchema,
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
			"orchestrator": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[clusterOrchestratorModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"eks": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[clusterOrchestratorEksConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"cluster_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(20, 2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[a-z\-]*:eks:[a-z0-9\-]*:[0-9]{12}:cluster\/[0-9A-Za-z][A-Za-z0-9\-_]{0,99}$`), `must match ^arn:aws[a-z\-]*:eks:[a-z0-9\-]*:[0-9]{12}:cluster\/[0-9A-Za-z][A-Za-z0-9\-_]{0,99}$`),
										},
									},
								},
							},
						},
					},
				},
			},
			"vpc_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: vpcConfigModelAttributeSchema,
				},
			},
		},
	}
}

func instanceGroupReplaceIf(ctx context.Context, req planmodifier.SetRequest, resp *setplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state resourceClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.InstanceGroups.Equal(state.InstanceGroups) {
		return
	}

	groupsPlanMap := createInstanceGroupsMap(ctx, resp.Diagnostics, plan.InstanceGroups)
	groupsStateMap := createInstanceGroupsMap(ctx, resp.Diagnostics, state.InstanceGroups)

	for name, planGroup := range groupsPlanMap {
		stateGroup, exists := groupsStateMap[name]
		if !exists {
			continue
		}

		if planGroup == stateGroup {
			continue
		}

		if !planGroup.ExecutionRole.Equal(stateGroup.ExecutionRole) ||
			!planGroup.InstanceStorageConfigs.Equal(stateGroup.InstanceStorageConfigs) ||
			!planGroup.InstanceType.Equal(stateGroup.InstanceType) ||
			!planGroup.OverrideVpcConfig.Equal(stateGroup.OverrideVpcConfig) ||
			!planGroup.ScheduledUpdateConfig.Equal(stateGroup.ScheduledUpdateConfig) ||
			(!planGroup.ThreadsPerCore.IsUnknown() && !planGroup.ThreadsPerCore.Equal(stateGroup.ThreadsPerCore)) ||
			!planGroup.TrainingPlanArn.Equal(stateGroup.TrainingPlanArn) {
			resp.RequiresReplace = true
			return
		}
	}
}

func (r *resourceCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var plan resourceClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input sagemaker.CreateClusterInput
	input.Tags = getTagsIn(ctx)

	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Cluster"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateCluster(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionCreating, ResNameCluster, plan.ClusterName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ClusterArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionCreating, ResNameCluster, plan.ClusterName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan, fwflex.WithFieldNamePrefix("Cluster"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	clusterCreateOutput, err := waitClusterCreated(ctx, conn, plan.ClusterName.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionWaitingForCreation, ResNameCluster, plan.ClusterName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, clusterCreateOutput, &plan, fwflex.WithFieldNamePrefix("Cluster"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	if clusterCreateOutput.ClusterStatus == awstypes.ClusterStatusFailed {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionConfiguring, ResNameCluster, plan.ClusterName.String(), err),
			*clusterCreateOutput.FailureMessage,
		)
	}

	clusterNodes, err := findClusterNodesByName(ctx, conn, plan.ClusterName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionReading, ResNameCluster, plan.ClusterName.String(), err),
			err.Error(),
		)
	}

	// Set values for unknowns.
	plan.InstanceGroups = mapNodesToInstanceGroup(ctx, resp.Diagnostics, plan.InstanceGroups, clusterNodes)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var state resourceClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterOutput, err := findClusterByName(ctx, conn, state.ClusterName.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionReading, ResNameCluster, state.ClusterName.String(), err),
			err.Error(),
		)
		return
	}

	// Set attributes for import.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, clusterOutput, &state, fwflex.WithFieldNamePrefix("Cluster"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterNodes, err := findClusterNodesByName(ctx, conn, state.ClusterName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionReading, ResNameCluster, state.ClusterName.String(), err),
			err.Error(),
		)
	}

	state.InstanceGroups = mapNodesToInstanceGroup(ctx, resp.Diagnostics, state.InstanceGroups, clusterNodes)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var plan, state resourceClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input sagemaker.UpdateClusterInput

		if !plan.InstanceGroups.Equal(state.InstanceGroups) {
			groupsPlanMap := createInstanceGroupsMap(ctx, resp.Diagnostics, plan.InstanceGroups)
			groupsStateMap := createInstanceGroupsMap(ctx, resp.Diagnostics, state.InstanceGroups)

			input.InstanceGroupsToDelete = difference(slices.Collect(maps.Keys(groupsStateMap)), slices.Collect(maps.Keys(groupsPlanMap)))

			if resp.Diagnostics.HasError() {
				return
			}
		}

		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Cluster"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateCluster(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SageMaker, create.ErrActionUpdating, ResNameCluster, plan.ClusterName.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.ClusterArn == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SageMaker, create.ErrActionUpdating, ResNameCluster, plan.ClusterName.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan, fwflex.WithFieldNamePrefix("Cluster"))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	clusterUpdateOutput, err := waitClusterUpdated(ctx, conn, plan.ClusterName.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionWaitingForUpdate, ResNameCluster, plan.ClusterName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, clusterUpdateOutput, &plan, fwflex.WithFieldNamePrefix("Cluster"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	if clusterUpdateOutput.ClusterStatus == awstypes.ClusterStatusFailed {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionConfiguring, ResNameCluster, plan.ClusterName.String(), err),
			*clusterUpdateOutput.FailureMessage,
		)
	}

	clusterNodes, err := findClusterNodesByName(ctx, conn, plan.ClusterName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionReading, ResNameCluster, state.ClusterName.String(), err),
			err.Error(),
		)
	}

	// Set values for unknowns.
	plan.InstanceGroups = mapNodesToInstanceGroup(ctx, resp.Diagnostics, plan.InstanceGroups, clusterNodes)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var state resourceClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := sagemaker.DeleteClusterInput{
		ClusterName: state.ClusterName.ValueStringPointer(),
	}

	_, err := conn.DeleteCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFound](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionDeleting, ResNameCluster, state.ClusterName.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitClusterDeleted(ctx, conn, state.ClusterName.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SageMaker, create.ErrActionWaitingForDeletion, ResNameCluster, state.ClusterName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCluster) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("cluster_name"), req, resp)
}

func waitClusterCreated(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*sagemaker.DescribeClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.ClusterStatusCreating), string(awstypes.ClusterStatusRollingback)},
		Target:                    []string{string(awstypes.ClusterStatusInservice), string(awstypes.ClusterStatusFailed)},
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*sagemaker.DescribeClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func waitClusterUpdated(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*sagemaker.DescribeClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.ClusterStatusRollingback), string(awstypes.ClusterStatusUpdating), string(awstypes.ClusterStatusSystemupdating)},
		Target:                    []string{string(awstypes.ClusterStatusInservice), string(awstypes.ClusterStatusFailed)},
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*sagemaker.DescribeClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*sagemaker.DescribeClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.ClusterStatusDeleting)},
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*sagemaker.DescribeClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func statusCluster(ctx context.Context, conn *sagemaker.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findClusterByName(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.ClusterStatus)), nil
	}
}

func findClusterByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeClusterOutput, error) {
	input := sagemaker.DescribeClusterInput{
		ClusterName: aws.String(name),
	}

	out, err := conn.DescribeCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.ClusterArn == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

func findClusterNodesByName(ctx context.Context, conn *sagemaker.Client, name string) ([]*awstypes.ClusterNodeDetails, error) {
	input := sagemaker.ListClusterNodesInput{
		ClusterName: aws.String(name),
	}
	pages := sagemaker.NewListClusterNodesPaginator(conn, &input)
	var clusterNodes []*awstypes.ClusterNodeDetails

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ClusterNodeSummaries {
			if v.InstanceId == nil {
				continue
			}

			clusterNode, err := conn.DescribeClusterNode(ctx, &sagemaker.DescribeClusterNodeInput{
				ClusterName: aws.String(name),
				NodeId:      v.InstanceId,
			})
			if err != nil {
				return nil, err
			}

			if clusterNode != nil && clusterNode.NodeDetails != nil {
				clusterNodes = append(clusterNodes, clusterNode.NodeDetails)
			}
		}
	}

	return clusterNodes, nil
}

func mapNodesToInstanceGroup(ctx context.Context, diagnostics diag.Diagnostics, groups fwtypes.SetNestedObjectValueOf[clusterInstanceGroupSpecificationModel], nodes []*awstypes.ClusterNodeDetails) fwtypes.SetNestedObjectValueOf[clusterInstanceGroupSpecificationModel] {
	groupsSlice, diags := groups.ToSlice(ctx)
	diagnostics.Append(diags...)

	nodesMap := createAggregatedClusterNodesMap(nodes)

	for i := range groupsSlice {

		group := groupsSlice[i]
		nodes := nodesMap[group.InstanceGroupName.ValueString()]
		var nodesFramework fwtypes.ListNestedObjectValueOf[clusterNodeDetailsModel]
		diagnostics.Append(fwflex.Flatten(ctx, nodes, &nodesFramework)...)

		group.Nodes = nodesFramework
	}

	return fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, groupsSlice)

}

func createAggregatedClusterNodesMap(nodes []*awstypes.ClusterNodeDetails) map[string][]*awstypes.ClusterNodeDetails {
	nodesMap := make(map[string][]*awstypes.ClusterNodeDetails)
	for i := range nodes {
		node := nodes[i]
		if node.InstanceGroupName == nil || *node.InstanceGroupName == "" {
			continue
		}
		nodesMap[*node.InstanceGroupName] = append(nodesMap[*node.InstanceGroupName], node)
	}

	return nodesMap
}

func createInstanceGroupsMap(ctx context.Context, diagnostics diag.Diagnostics, groups fwtypes.SetNestedObjectValueOf[clusterInstanceGroupSpecificationModel]) map[string]*clusterInstanceGroupSpecificationModel {
	groupsSlice, diags := groups.ToSlice(ctx)
	diagnostics.Append(diags...)

	groupsMap := make(map[string]*clusterInstanceGroupSpecificationModel)
	for i := range groupsSlice {
		ig := groupsSlice[i]
		groupsMap[ig.InstanceGroupName.ValueString()] = ig
	}

	return groupsMap
}

// Function to calculate set difference (a - b)
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

type resourceClusterModel struct {
	framework.WithRegionModel
	Arn            types.String                                                           `tfsdk:"arn"`
	ClusterName    types.String                                                           `tfsdk:"cluster_name"`
	ClusterStatus  fwtypes.StringEnum[awstypes.ClusterStatus]                             `tfsdk:"cluster_status"`
	InstanceGroups fwtypes.SetNestedObjectValueOf[clusterInstanceGroupSpecificationModel] `tfsdk:"instance_group"`
	NodeRecovery   fwtypes.StringEnum[awstypes.ClusterNodeRecovery]                       `tfsdk:"node_recovery"`
	Orchestrator   fwtypes.ListNestedObjectValueOf[clusterOrchestratorModel]              `tfsdk:"orchestrator"`
	Tags           tftags.Map                                                             `tfsdk:"tags"`
	TagsAll        tftags.Map                                                             `tfsdk:"tags_all"`
	Timeouts       timeouts.Value                                                         `tfsdk:"timeouts"`
	VpcConfig      fwtypes.ListNestedObjectValueOf[vpcConfigModel]                        `tfsdk:"vpc_config"`
}

type clusterInstanceGroupSpecificationModel struct {
	ExecutionRole           types.String                                                       `tfsdk:"execution_role"`
	InstanceCount           types.Int32                                                        `tfsdk:"instance_count"`
	InstanceGroupName       types.String                                                       `tfsdk:"instance_group_name"`
	InstanceType            fwtypes.StringEnum[awstypes.ClusterInstanceType]                   `tfsdk:"instance_type"`
	LifeCycleConfig         fwtypes.ListNestedObjectValueOf[clusterLifeCycleConfigModel]       `tfsdk:"lifecycle_config"`
	InstanceStorageConfigs  fwtypes.ListNestedObjectValueOf[clusterInstanceStorageConfigModel] `tfsdk:"instance_storage_config"`
	Nodes                   fwtypes.ListNestedObjectValueOf[clusterNodeDetailsModel]           `tfsdk:"nodes"`
	OnStartDeepHealthChecks fwtypes.ListValueOf[types.String]                                  `tfsdk:"on_start_deep_healthchecks"`
	OverrideVpcConfig       fwtypes.ListNestedObjectValueOf[vpcConfigModel]                    `tfsdk:"override_vpc_config"`
	ScheduledUpdateConfig   fwtypes.ListNestedObjectValueOf[scheduledUpdateConfigModel]        `tfsdk:"scheduled_update_config"`
	ThreadsPerCore          types.Int32                                                        `tfsdk:"threads_per_core"`
	TrainingPlanArn         fwtypes.ARN                                                        `tfsdk:"training_plan_arn"`
}

var (
	_ fwflex.Flattener = &clusterInstanceGroupSpecificationModel{}
)

func (m *clusterInstanceGroupSpecificationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	if v, ok := v.(awstypes.ClusterInstanceGroupDetails); ok {
		diags.Append(fwflex.Flatten(ctx, v.CurrentCount, &m.InstanceCount)...)
		diags.Append(fwflex.Flatten(ctx, v.ExecutionRole, &m.ExecutionRole)...)
		diags.Append(fwflex.Flatten(ctx, v.InstanceGroupName, &m.InstanceGroupName)...)
		diags.Append(fwflex.Flatten(ctx, v.InstanceStorageConfigs, &m.InstanceStorageConfigs)...)
		diags.Append(fwflex.Flatten(ctx, v.InstanceType, &m.InstanceType)...)
		diags.Append(fwflex.Flatten(ctx, v.LifeCycleConfig, &m.LifeCycleConfig)...)
		diags.Append(fwflex.Flatten(ctx, v.OnStartDeepHealthChecks, &m.OnStartDeepHealthChecks)...)
		diags.Append(fwflex.Flatten(ctx, v.OverrideVpcConfig, &m.OverrideVpcConfig)...)
		diags.Append(fwflex.Flatten(ctx, v.ScheduledUpdateConfig, &m.ScheduledUpdateConfig)...)
		diags.Append(fwflex.Flatten(ctx, v.ThreadsPerCore, &m.ThreadsPerCore)...)
		diags.Append(fwflex.Flatten(ctx, v.TrainingPlanArn, &m.TrainingPlanArn)...)
	}
	return diags
}

type clusterLifeCycleConfigModel struct {
	OnCreate    types.String `tfsdk:"on_create"`
	SourceS3Uri types.String `tfsdk:"source_s3_uri"`
}

type clusterInstanceStorageConfigModel struct {
	EbsVolumeConfig fwtypes.ListNestedObjectValueOf[clusterEbsVolumeConfigModel] `tfsdk:"ebs_volume_config"`
}

type clusterEbsVolumeConfigModel struct {
	VolumeSizeInGB types.Int32 `tfsdk:"volume_size_in_gb"`
}

var (
	_ fwflex.Expander  = clusterInstanceStorageConfigModel{}
	_ fwflex.Flattener = &clusterInstanceStorageConfigModel{}
)

func (m clusterInstanceStorageConfigModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.EbsVolumeConfig.IsNull():
		clusterEbsVolumeConfigModel := fwdiag.Must(m.EbsVolumeConfig.ToPtr(ctx))

		return &awstypes.ClusterInstanceStorageConfigMemberEbsVolumeConfig{
			Value: awstypes.ClusterEbsVolumeConfig{
				VolumeSizeInGB: fwflex.Int32FromFramework(ctx, clusterEbsVolumeConfigModel.VolumeSizeInGB),
			},
		}, diags
	}

	return nil, diags
}

func (m *clusterInstanceStorageConfigModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ClusterInstanceStorageConfigMemberEbsVolumeConfig:
		var model clusterEbsVolumeConfigModel
		d := fwflex.Flatten(ctx, t.Value, &model)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		m.EbsVolumeConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags
	}

	return diags
}

type clusterNodeDetailsModel struct {
	InstanceGroupName      types.String                                                       `tfsdk:"instance_group_name"`
	InstanceID             types.String                                                       `tfsdk:"instance_id"`
	InstanceStatus         fwtypes.ListNestedObjectValueOf[clusterInstanceStatusDetailsModel] `tfsdk:"instance_status"`
	InstanceStorageConfigs fwtypes.ListNestedObjectValueOf[clusterInstanceStorageConfigModel] `tfsdk:"instance_storage_configs"`
	InstanceType           fwtypes.StringEnum[awstypes.ClusterInstanceType]                   `tfsdk:"instance_type"`
	LastSoftwareUpdateTime timetypes.RFC3339                                                  `tfsdk:"last_software_update_time"`
	LaunchTime             timetypes.RFC3339                                                  `tfsdk:"launch_time"`
	LifeCycleConfig        fwtypes.ListNestedObjectValueOf[clusterLifeCycleConfigModel]       `tfsdk:"lifecycle_config"`
	OverrideVpcConfig      fwtypes.ListNestedObjectValueOf[vpcConfigModel]                    `tfsdk:"override_vpc_config"`
	Placement              fwtypes.ListNestedObjectValueOf[clusterInstancePlacementModel]     `tfsdk:"placement"`
	PrivateDnsHostname     types.String                                                       `tfsdk:"private_dns_hostname"`
	PrivatePrimaryIP       types.String                                                       `tfsdk:"private_primary_ip"`
	PrivatePrimaryIpv6     types.String                                                       `tfsdk:"private_primary_ipv6"`
	ThreadsPerCore         types.Int64                                                        `tfsdk:"threads_per_core"`
}

type clusterInstanceStatusDetailsModel struct {
	Message types.String                                       `tfsdk:"message"`
	Status  fwtypes.StringEnum[awstypes.ClusterInstanceStatus] `tfsdk:"status"`
}

type clusterInstancePlacementModel struct {
	AvailabilityZone   types.String `tfsdk:"availability_zone"`
	AvailabilityZoneID types.String `tfsdk:"availability_zone_id"`
}

type scheduledUpdateConfigModel struct {
	ScheduleExpression types.String                                                  `tfsdk:"schedule_expression"`
	DeploymentConfig   fwtypes.ListNestedObjectValueOf[deploymentConfigurationModel] `tfsdk:"deployment_config"`
}

type deploymentConfigurationModel struct {
	AutoRollbackConfiguration fwtypes.SetNestedObjectValueOf[alarmDetailsModel]             `tfsdk:"auto_rollback_configuration"`
	RollingUpdatePolicy       fwtypes.ListNestedObjectValueOf[rollingDeploymentPolicyModel] `tfsdk:"rolling_update_policy"`
	WaitInterval              types.Int32                                                   `tfsdk:"wait_interval"`
}

type alarmDetailsModel struct {
	AlarmName types.String `tfsdk:"alarm_name"`
}

type rollingDeploymentPolicyModel struct {
	MaximumBatchSize         fwtypes.ListNestedObjectValueOf[capacitySizeConfigModel] `tfsdk:"maximum_batch_size"`
	RollbackMaximumBatchSize fwtypes.ListNestedObjectValueOf[capacitySizeConfigModel] `tfsdk:"rollback_maximum_batch_size"`
}

type capacitySizeConfigModel struct {
	Type  fwtypes.StringEnum[awstypes.NodeUnavailabilityType] `tfsdk:"type"`
	Value types.Int32                                         `tfsdk:"value"`
}

type clusterOrchestratorModel struct {
	Eks fwtypes.ListNestedObjectValueOf[clusterOrchestratorEksConfigModel] `tfsdk:"eks"`
}

type clusterOrchestratorEksConfigModel struct {
	ClusterArn fwtypes.ARN `tfsdk:"cluster_arn"`
}

type vpcConfigModel struct {
	SecurityGroupIds fwtypes.SetOfString `tfsdk:"security_group_ids"`
	Subnets          fwtypes.SetOfString `tfsdk:"subnets"`
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := sagemaker.ListClustersInput{}
	conn := client.SageMakerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := sagemaker.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ClusterSummaries {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceCluster, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.ClusterArn))),
			)
		}
	}

	return sweepResources, nil
}
