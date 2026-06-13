// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_compute_quota", name="Compute Quota")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sagemaker;sagemaker.DescribeComputeQuotaOutput")
func newComputeQuotaResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &computeQuotaResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameComputeQuota = "Compute Quota"
)

type computeQuotaResource struct {
	framework.ResourceWithModel[computeQuotaResourceModel]
	framework.WithTimeouts
}

func (r *computeQuotaResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"activation_state": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ActivationState](),
				Optional:   true,
				Computed:   true,
				Default:    fwtypes.StringEnumType[awstypes.ActivationState]().AttributeDefault(awstypes.ActivationStateEnabled),
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"cluster_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compute_quota_version": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`\S`), "must contain at least one non-whitespace character"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SchedulerResourceStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"compute_quota_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[computeQuotaConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"preempt_team_tasks": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PreemptTeamTasks](),
							Optional:   true,
							Computed:   true,
							Default:    fwtypes.StringEnumType[awstypes.PreemptTeamTasks]().AttributeDefault(awstypes.PreemptTeamTasksLowerpriority),
						},
					},
					Blocks: map[string]schema.Block{
						"compute_quota_resources": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[computeQuotaResourceConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: computeQuotaResourceConfigNestedObject(ctx),
						},
						"resource_sharing_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[resourceSharingConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"borrow_limit": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Default:  int32default.StaticInt32(50),
										Validators: []validator.Int32{
											int32validator.Between(1, 500),
										},
									},
									"strategy": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ResourceSharingStrategy](),
										Optional:   true,
										Computed:   true,
										Default:    fwtypes.StringEnumType[awstypes.ResourceSharingStrategy]().AttributeDefault(awstypes.ResourceSharingStrategyLendandborrow),
									},
								},
								Blocks: map[string]schema.Block{
									"absolute_borrow_limits": schema.ListNestedBlock{
										CustomType:   fwtypes.NewListNestedObjectTypeOf[computeQuotaResourceConfigModel](ctx),
										NestedObject: computeQuotaResourceConfigNestedObject(ctx),
									},
								},
							},
						},
					},
				},
			},
			"compute_quota_target": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[computeQuotaTargetModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"fair_share_weight": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							Default:  int32default.StaticInt32(0),
							Validators: []validator.Int32{
								int32validator.Between(0, 100),
							},
						},
						"team_name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(`\S`), "must contain at least one non-whitespace character"),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func computeQuotaResourceConfigNestedObject(ctx context.Context) schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: computeQuotaResourceConfigSchema(),
		Blocks: map[string]schema.Block{
			"accelerator_partition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[acceleratorPartitionConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"count": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(1),
							},
						},
						"type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MIGProfileType](),
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func computeQuotaResourceConfigSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"accelerators": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.AtLeast(1),
			},
		},
		"count": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.AtLeast(1),
			},
		},
		"instance_type": schema.StringAttribute{
			CustomType: fwtypes.StringEnumType[awstypes.ClusterInstanceType](),
			Required:   true,
		},
		"memory_in_gib": schema.Float32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Float32{
				float32validator.AtLeast(0),
				float32validator.NoneOf(0),
			},
		},
		"vcpu": schema.Float32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Float32{
				float32validator.AtLeast(0),
				float32validator.NoneOf(0),
			},
		},
	}
}

func (r *computeQuotaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan computeQuotaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)
	input, diags := expandCreateComputeQuotaInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateComputeQuota(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating SageMaker Compute Quota (%s)", plan.Name.ValueString()), err.Error())
		return
	}

	plan.ID = types.StringPointerValue(output.ComputeQuotaId)
	plan.ARN = types.StringPointerValue(output.ComputeQuotaArn)

	outputWait, err := waitComputeQuotaCreated(ctx, conn, plan.ID.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		if cleanupErr := deleteComputeQuotaAfterFailedCreate(ctx, conn, plan.ID.ValueString(), r.DeleteTimeout(ctx, plan.Timeouts)); cleanupErr != nil {
			resp.State.SetAttribute(ctx, path.Root(names.AttrID), plan.ID) // Set 'id' so as to taint the resource.
			err = errors.Join(err, fmt.Errorf("deleting failed SageMaker Compute Quota (%s): %w", plan.ID.ValueString(), cleanupErr))
		}

		resp.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker Compute Quota (%s) create", plan.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(flattenComputeQuota(ctx, outputWait, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *computeQuotaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state computeQuotaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)
	output, err := findComputeQuotaByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading SageMaker Compute Quota (%s)", state.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(flattenComputeQuota(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *computeQuotaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state computeQuotaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)
	current, err := findComputeQuotaByID(ctx, conn, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading SageMaker Compute Quota (%s)", state.ID.ValueString()), err.Error())
		return
	}

	input, diags := expandUpdateComputeQuotaInput(ctx, plan, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = conn.UpdateComputeQuota(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("updating SageMaker Compute Quota (%s)", state.ID.ValueString()), err.Error())
		return
	}

	outputWait, err := waitComputeQuotaUpdated(ctx, conn, state.ID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker Compute Quota (%s) update", state.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(flattenComputeQuota(ctx, outputWait, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *computeQuotaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state computeQuotaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)
	_, err := conn.DeleteComputeQuota(ctx, &sagemaker.DeleteComputeQuotaInput{
		ComputeQuotaId: state.ID.ValueStringPointer(),
	})
	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting SageMaker Compute Quota (%s)", state.ID.ValueString()), err.Error())
		return
	}

	if err := waitComputeQuotaDeleted(ctx, conn, state.ID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker Compute Quota (%s) delete", state.ID.ValueString()), err.Error())
	}
}

func (r *computeQuotaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func expandCreateComputeQuotaInput(ctx context.Context, data computeQuotaResourceModel) (*sagemaker.CreateComputeQuotaInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	config, d := expandComputeQuotaConfig(ctx, data.ComputeQuotaConfig)
	diags.Append(d...)
	target, d := expandComputeQuotaTarget(ctx, data.ComputeQuotaTarget)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &sagemaker.CreateComputeQuotaInput{
		ActivationState:    data.ActivationState.ValueEnum(),
		ClusterArn:         data.ClusterARN.ValueStringPointer(),
		ComputeQuotaConfig: config,
		ComputeQuotaTarget: target,
		Description:        data.Description.ValueStringPointer(),
		Name:               data.Name.ValueStringPointer(),
	}, diags
}

func expandUpdateComputeQuotaInput(ctx context.Context, data computeQuotaResourceModel, current *sagemaker.DescribeComputeQuotaOutput) (*sagemaker.UpdateComputeQuotaInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	config, d := expandComputeQuotaConfig(ctx, data.ComputeQuotaConfig)
	diags.Append(d...)
	target, d := expandComputeQuotaTarget(ctx, data.ComputeQuotaTarget)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &sagemaker.UpdateComputeQuotaInput{
		ActivationState:    data.ActivationState.ValueEnum(),
		ComputeQuotaConfig: config,
		ComputeQuotaId:     data.ID.ValueStringPointer(),
		ComputeQuotaTarget: target,
		Description:        data.Description.ValueStringPointer(),
		TargetVersion:      current.ComputeQuotaVersion,
	}, diags
}

func expandComputeQuotaConfig(ctx context.Context, value fwtypes.ListNestedObjectValueOf[computeQuotaConfigModel]) (*awstypes.ComputeQuotaConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	model, d := value.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || model == nil {
		return nil, diags
	}

	resources, d := expandComputeQuotaResourceConfigs(ctx, model.ComputeQuotaResources)
	diags.Append(d...)
	sharingConfig, d := expandResourceSharingConfig(ctx, model.ResourceSharingConfig)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &awstypes.ComputeQuotaConfig{
		ComputeQuotaResources: resources,
		PreemptTeamTasks:      model.PreemptTeamTasks.ValueEnum(),
		ResourceSharingConfig: sharingConfig,
	}, diags
}

func expandResourceSharingConfig(ctx context.Context, value fwtypes.ListNestedObjectValueOf[resourceSharingConfigModel]) (*awstypes.ResourceSharingConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	model, d := value.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || model == nil {
		return nil, diags
	}

	limits, d := expandComputeQuotaResourceConfigs(ctx, model.AbsoluteBorrowLimits)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	return &awstypes.ResourceSharingConfig{
		AbsoluteBorrowLimits: limits,
		BorrowLimit:          int32ValuePointer(model.BorrowLimit),
		Strategy:             model.Strategy.ValueEnum(),
	}, diags
}

func expandComputeQuotaTarget(ctx context.Context, value fwtypes.ListNestedObjectValueOf[computeQuotaTargetModel]) (*awstypes.ComputeQuotaTarget, diag.Diagnostics) {
	var diags diag.Diagnostics

	model, d := value.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || model == nil {
		return nil, diags
	}

	return &awstypes.ComputeQuotaTarget{
		FairShareWeight: int32ValuePointer(model.FairShareWeight),
		TeamName:        model.TeamName.ValueStringPointer(),
	}, diags
}

func expandComputeQuotaResourceConfigs(ctx context.Context, value fwtypes.ListNestedObjectValueOf[computeQuotaResourceConfigModel]) ([]awstypes.ComputeQuotaResourceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	models, d := value.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	output := make([]awstypes.ComputeQuotaResourceConfig, 0, len(models))
	for _, model := range models {
		if model == nil {
			continue
		}

		if !computeQuotaResourceConfigHasAllocation(model) {
			diags.AddError(
				"Missing SageMaker Compute Quota Resource Allocation",
				"Each compute_quota_resources and absolute_borrow_limits block must set at least one of count, accelerators, vcpu, memory_in_gib, or accelerator_partition.",
			)

			return nil, diags
		}

		partition, d := expandAcceleratorPartitionConfig(ctx, model.AcceleratorPartition)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		output = append(output, awstypes.ComputeQuotaResourceConfig{
			AcceleratorPartition: partition,
			Accelerators:         int32ValuePointer(model.Accelerators),
			Count:                int32ValuePointer(model.Count),
			InstanceType:         model.InstanceType.ValueEnum(),
			MemoryInGiB:          float32ValuePointer(model.MemoryInGiB),
			VCpu:                 float32ValuePointer(model.VCpu),
		})
	}

	return output, diags
}

func computeQuotaResourceConfigHasAllocation(model *computeQuotaResourceConfigModel) bool {
	return int32ValueIsConfigured(model.Accelerators) ||
		int32ValueIsConfigured(model.Count) ||
		float32ValueIsConfigured(model.MemoryInGiB) ||
		float32ValueIsConfigured(model.VCpu) ||
		!model.AcceleratorPartition.IsNull() && !model.AcceleratorPartition.IsUnknown()
}

func expandAcceleratorPartitionConfig(ctx context.Context, value fwtypes.ListNestedObjectValueOf[acceleratorPartitionConfigModel]) (*awstypes.AcceleratorPartitionConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	model, d := value.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || model == nil {
		return nil, diags
	}

	return &awstypes.AcceleratorPartitionConfig{
		Count: int32ValuePointer(model.Count),
		Type:  model.Type.ValueEnum(),
	}, diags
}

func int32ValuePointer(value types.Int32) *int32 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	return value.ValueInt32Pointer()
}

func int32ValueIsConfigured(value types.Int32) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func float32ValuePointer(value types.Float32) *float32 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	return value.ValueFloat32Pointer()
}

func float32ValueIsConfigured(value types.Float32) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func flattenComputeQuota(ctx context.Context, output *sagemaker.DescribeComputeQuotaOutput, data *computeQuotaResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ActivationState = fwtypes.StringEnumValue(output.ActivationState)
	data.ARN = fwflex.StringToFramework(ctx, output.ComputeQuotaArn)
	data.ClusterARN = fwflex.StringToFrameworkARN(ctx, output.ClusterArn)
	data.ComputeQuotaConfig, diags = flattenComputeQuotaConfig(ctx, output.ComputeQuotaConfig)
	if diags.HasError() {
		return diags
	}
	data.ComputeQuotaTarget, diags = flattenComputeQuotaTarget(ctx, output.ComputeQuotaTarget)
	if diags.HasError() {
		return diags
	}
	data.ComputeQuotaVersion = types.Int32PointerValue(output.ComputeQuotaVersion)
	data.CreationTime = fwflex.TimeToFramework(ctx, output.CreationTime)
	data.Description = fwflex.StringToFramework(ctx, output.Description)
	data.FailureReason = fwflex.StringToFramework(ctx, output.FailureReason)
	data.ID = fwflex.StringToFramework(ctx, output.ComputeQuotaId)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.LastModifiedTime)
	data.Name = fwflex.StringToFramework(ctx, output.Name)
	data.Status = fwtypes.StringEnumValue(output.Status)

	return diags
}

func flattenComputeQuotaConfig(ctx context.Context, apiObject *awstypes.ComputeQuotaConfig) (fwtypes.ListNestedObjectValueOf[computeQuotaConfigModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[computeQuotaConfigModel](ctx), diags
	}

	model := &computeQuotaConfigModel{
		PreemptTeamTasks: fwtypes.StringEnumValue(apiObject.PreemptTeamTasks),
	}
	model.ComputeQuotaResources, diags = flattenComputeQuotaResourceConfigs(ctx, apiObject.ComputeQuotaResources)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfUnknown[computeQuotaConfigModel](ctx), diags
	}
	model.ResourceSharingConfig, diags = flattenResourceSharingConfig(ctx, apiObject.ResourceSharingConfig)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfUnknown[computeQuotaConfigModel](ctx), diags
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, model), diags
}

func flattenResourceSharingConfig(ctx context.Context, apiObject *awstypes.ResourceSharingConfig) (fwtypes.ListNestedObjectValueOf[resourceSharingConfigModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[resourceSharingConfigModel](ctx), diags
	}

	model := &resourceSharingConfigModel{
		BorrowLimit: types.Int32PointerValue(apiObject.BorrowLimit),
		Strategy:    fwtypes.StringEnumValue(apiObject.Strategy),
	}
	model.AbsoluteBorrowLimits, diags = flattenComputeQuotaResourceConfigs(ctx, apiObject.AbsoluteBorrowLimits)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfUnknown[resourceSharingConfigModel](ctx), diags
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, model), diags
}

func flattenComputeQuotaTarget(ctx context.Context, apiObject *awstypes.ComputeQuotaTarget) (fwtypes.ListNestedObjectValueOf[computeQuotaTargetModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[computeQuotaTargetModel](ctx), diags
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &computeQuotaTargetModel{
		FairShareWeight: types.Int32PointerValue(apiObject.FairShareWeight),
		TeamName:        fwflex.StringToFramework(ctx, apiObject.TeamName),
	}), diags
}

func flattenComputeQuotaResourceConfigs(ctx context.Context, apiObjects []awstypes.ComputeQuotaResourceConfig) (fwtypes.ListNestedObjectValueOf[computeQuotaResourceConfigModel], diag.Diagnostics) {
	if len(apiObjects) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[computeQuotaResourceConfigModel](ctx), nil
	}

	models := make([]*computeQuotaResourceConfigModel, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		acceleratorPartition, diags := flattenAcceleratorPartitionConfig(ctx, apiObject.AcceleratorPartition)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfUnknown[computeQuotaResourceConfigModel](ctx), diags
		}

		models = append(models, &computeQuotaResourceConfigModel{
			AcceleratorPartition: acceleratorPartition,
			Accelerators:         types.Int32PointerValue(apiObject.Accelerators),
			Count:                types.Int32PointerValue(apiObject.Count),
			InstanceType:         fwtypes.StringEnumValue(apiObject.InstanceType),
			MemoryInGiB:          types.Float32PointerValue(apiObject.MemoryInGiB),
			VCpu:                 types.Float32PointerValue(apiObject.VCpu),
		})
	}

	return fwtypes.NewListNestedObjectValueOfSlice(ctx, models, nil)
}

func flattenAcceleratorPartitionConfig(ctx context.Context, apiObject *awstypes.AcceleratorPartitionConfig) (fwtypes.ListNestedObjectValueOf[acceleratorPartitionConfigModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[acceleratorPartitionConfigModel](ctx), diags
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &acceleratorPartitionConfigModel{
		Count: types.Int32PointerValue(apiObject.Count),
		Type:  fwtypes.StringEnumValue(apiObject.Type),
	}), diags
}

func waitComputeQuotaCreated(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*sagemaker.DescribeComputeQuotaOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.SchedulerResourceStatusCreating,
		),
		Target: enum.Slice(
			awstypes.SchedulerResourceStatusCreated,
			awstypes.SchedulerResourceStatusUpdated,
		),
		Refresh:        statusComputeQuota(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: 20,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*sagemaker.DescribeComputeQuotaOutput); ok {
		if isComputeQuotaFailureStatus(output.Status) && aws.ToString(output.FailureReason) != "" {
			retry.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeQuotaUpdated(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*sagemaker.DescribeComputeQuotaOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.SchedulerResourceStatusUpdating,
		),
		Target: enum.Slice(
			awstypes.SchedulerResourceStatusCreated,
			awstypes.SchedulerResourceStatusUpdated,
		),
		Refresh: statusComputeQuota(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*sagemaker.DescribeComputeQuotaOutput); ok {
		if isComputeQuotaFailureStatus(output.Status) && aws.ToString(output.FailureReason) != "" {
			retry.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeQuotaDeleted(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.SchedulerResourceStatusCreateFailed,
			awstypes.SchedulerResourceStatusCreateRollbackFailed,
			awstypes.SchedulerResourceStatusCreated,
			awstypes.SchedulerResourceStatusUpdated,
			awstypes.SchedulerResourceStatusUpdateFailed,
			awstypes.SchedulerResourceStatusUpdateRollbackFailed,
			awstypes.SchedulerResourceStatusDeleting,
		),
		Target:  []string{},
		Refresh: statusComputeQuota(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func deleteComputeQuotaAfterFailedCreate(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) error {
	output, err := findComputeQuotaByID(ctx, conn, id)
	if retry.NotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if !isComputeQuotaCreateFailureStatus(output.Status) {
		return nil
	}

	_, err = conn.DeleteComputeQuota(ctx, &sagemaker.DeleteComputeQuotaInput{
		ComputeQuotaId: aws.String(id),
	})
	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil
	}
	if err != nil {
		return err
	}

	return waitComputeQuotaDeleted(ctx, conn, id, timeout)
}

func statusComputeQuota(_ context.Context, conn *sagemaker.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findComputeQuotaByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findComputeQuotaByID(ctx context.Context, conn *sagemaker.Client, id string) (*sagemaker.DescribeComputeQuotaOutput, error) {
	output, err := conn.DescribeComputeQuota(ctx, &sagemaker.DescribeComputeQuotaInput{
		ComputeQuotaId: aws.String(id),
	})
	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}
	if output.Status == awstypes.SchedulerResourceStatusDeleted {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func isComputeQuotaFailureStatus(status awstypes.SchedulerResourceStatus) bool {
	return status == awstypes.SchedulerResourceStatusCreateFailed ||
		status == awstypes.SchedulerResourceStatusCreateRollbackFailed ||
		status == awstypes.SchedulerResourceStatusUpdateFailed ||
		status == awstypes.SchedulerResourceStatusUpdateRollbackFailed ||
		status == awstypes.SchedulerResourceStatusDeleteFailed ||
		status == awstypes.SchedulerResourceStatusDeleteRollbackFailed
}

func isComputeQuotaCreateFailureStatus(status awstypes.SchedulerResourceStatus) bool {
	return status == awstypes.SchedulerResourceStatusCreateFailed ||
		status == awstypes.SchedulerResourceStatusCreateRollbackFailed
}

type computeQuotaResourceModel struct {
	framework.WithRegionModel
	ActivationState     fwtypes.StringEnum[awstypes.ActivationState]             `tfsdk:"activation_state"`
	ARN                 types.String                                             `tfsdk:"arn"`
	ClusterARN          fwtypes.ARN                                              `tfsdk:"cluster_arn"`
	ComputeQuotaConfig  fwtypes.ListNestedObjectValueOf[computeQuotaConfigModel] `tfsdk:"compute_quota_config"`
	ComputeQuotaTarget  fwtypes.ListNestedObjectValueOf[computeQuotaTargetModel] `tfsdk:"compute_quota_target"`
	ComputeQuotaVersion types.Int32                                              `tfsdk:"compute_quota_version"`
	CreationTime        timetypes.RFC3339                                        `tfsdk:"creation_time"`
	Description         types.String                                             `tfsdk:"description"`
	FailureReason       types.String                                             `tfsdk:"failure_reason"`
	ID                  types.String                                             `tfsdk:"id"`
	LastModifiedTime    timetypes.RFC3339                                        `tfsdk:"last_modified_time"`
	Name                types.String                                             `tfsdk:"name"`
	Status              fwtypes.StringEnum[awstypes.SchedulerResourceStatus]     `tfsdk:"status"`
	Tags                tftags.Map                                               `tfsdk:"tags"`
	TagsAll             tftags.Map                                               `tfsdk:"tags_all"`
	Timeouts            timeouts.Value                                           `tfsdk:"timeouts"`
}

type computeQuotaConfigModel struct {
	ComputeQuotaResources fwtypes.ListNestedObjectValueOf[computeQuotaResourceConfigModel] `tfsdk:"compute_quota_resources"`
	PreemptTeamTasks      fwtypes.StringEnum[awstypes.PreemptTeamTasks]                    `tfsdk:"preempt_team_tasks"`
	ResourceSharingConfig fwtypes.ListNestedObjectValueOf[resourceSharingConfigModel]      `tfsdk:"resource_sharing_config"`
}

type computeQuotaResourceConfigModel struct {
	AcceleratorPartition fwtypes.ListNestedObjectValueOf[acceleratorPartitionConfigModel] `tfsdk:"accelerator_partition"`
	Accelerators         types.Int32                                                      `tfsdk:"accelerators"`
	Count                types.Int32                                                      `tfsdk:"count"`
	InstanceType         fwtypes.StringEnum[awstypes.ClusterInstanceType]                 `tfsdk:"instance_type"`
	MemoryInGiB          types.Float32                                                    `tfsdk:"memory_in_gib"`
	VCpu                 types.Float32                                                    `tfsdk:"vcpu"`
}

type acceleratorPartitionConfigModel struct {
	Count types.Int32                                 `tfsdk:"count"`
	Type  fwtypes.StringEnum[awstypes.MIGProfileType] `tfsdk:"type"`
}

type resourceSharingConfigModel struct {
	AbsoluteBorrowLimits fwtypes.ListNestedObjectValueOf[computeQuotaResourceConfigModel] `tfsdk:"absolute_borrow_limits"`
	BorrowLimit          types.Int32                                                      `tfsdk:"borrow_limit"`
	Strategy             fwtypes.StringEnum[awstypes.ResourceSharingStrategy]             `tfsdk:"strategy"`
}

type computeQuotaTargetModel struct {
	FairShareWeight types.Int32  `tfsdk:"fair_share_weight"`
	TeamName        types.String `tfsdk:"team_name"`
}

func sweepComputeQuotas(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	input := &sagemaker.ListComputeQuotasInput{
		NameContains: aws.String(sweep.ResourcePrefix),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListComputeQuotasPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ComputeQuotaSummaries {
			if v.Status == awstypes.SchedulerResourceStatusDeleted {
				continue
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newComputeQuotaResource, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.ComputeQuotaId))),
			)
		}
	}

	return sweepResources, nil
}
