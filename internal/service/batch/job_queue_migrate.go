// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func jobQueueSchema0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compute_environments": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrPriority: schema.Int64Attribute{
				Required: true,
			},
			"scheduling_policy_arn": schema.StringAttribute{
				Optional: true,
			},
			names.AttrState: schema.StringAttribute{
				Required: true,
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
		},
	}
}

func jobQueueSchema1(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compute_environments": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z]{1}[0-9A-Za-z_-]{0,127}$`),
						"must be up to 128 letters (uppercase and lowercase), numbers, underscores and dashes, and must start with an alphanumeric"),
				},
			},
			names.AttrPriority: schema.Int64Attribute{
				Required: true,
			},
			"scheduling_policy_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrState: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidateIgnoreCase[awstypes.JQState](),
				},
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
			"compute_environment_order": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[computeEnvironmentOrderModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"compute_environment": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						"order": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
			"job_state_time_limit_action": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[jobStateTimeLimitActionModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrAction: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.JobStateTimeLimitActionsAction](),
							Required:   true,
						},
						"max_time_seconds": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(600, 86400),
							},
						},
						"reason": schema.StringAttribute{
							Required: true,
						},
						names.AttrState: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.JobStateTimeLimitActionsState](),
							Required:   true,
						},
					},
				},
			},
		},
	}
}

type resourceJobQueueDataV1 struct {
	ComputeEnvironments      fwtypes.ListOfString                                          `tfsdk:"compute_environments"`
	ComputeEnvironmentOrder  fwtypes.ListNestedObjectValueOf[computeEnvironmentOrderModel] `tfsdk:"compute_environment_order"`
	ID                       types.String                                                  `tfsdk:"id"`
	JobQueueARN              types.String                                                  `tfsdk:"arn"`
	JobQueueName             types.String                                                  `tfsdk:"name"`
	JobStateTimeLimitActions fwtypes.ListNestedObjectValueOf[jobStateTimeLimitActionModel] `tfsdk:"job_state_time_limit_action"`
	Priority                 types.Int64                                                   `tfsdk:"priority"`
	SchedulingPolicyARN      fwtypes.ARN                                                   `tfsdk:"scheduling_policy_arn"`
	State                    types.String                                                  `tfsdk:"state"`
	Tags                     tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                                `tfsdk:"timeouts"`
}

func upgradeJobQueueResourceStateV0toV1(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	type resourceJobQueueDataV0 struct {
		ComputeEnvironments fwtypes.ListOfString `tfsdk:"compute_environments"`
		ID                  types.String         `tfsdk:"id"`
		JobQueueARN         types.String         `tfsdk:"arn"`
		JobQueueName        types.String         `tfsdk:"name"`
		Priority            types.Int64          `tfsdk:"priority"`
		SchedulingPolicyARN types.String         `tfsdk:"scheduling_policy_arn"`
		State               types.String         `tfsdk:"state"`
		Tags                tftags.Map           `tfsdk:"tags"`
		TagsAll             tftags.Map           `tfsdk:"tags_all"`
		Timeouts            timeouts.Value       `tfsdk:"timeouts"`
	}

	var jobQueueDataV0 resourceJobQueueDataV0
	response.Diagnostics.Append(request.State.Get(ctx, &jobQueueDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	jobQueueDataV1 := resourceJobQueueDataV1{
		ComputeEnvironments:      jobQueueDataV0.ComputeEnvironments,
		ComputeEnvironmentOrder:  fwtypes.NewListNestedObjectValueOfNull[computeEnvironmentOrderModel](ctx),
		ID:                       jobQueueDataV0.ID,
		JobQueueARN:              jobQueueDataV0.JobQueueARN,
		JobQueueName:             jobQueueDataV0.JobQueueName,
		JobStateTimeLimitActions: fwtypes.NewListNestedObjectValueOfNull[jobStateTimeLimitActionModel](ctx),
		Priority:                 jobQueueDataV0.Priority,
		State:                    jobQueueDataV0.State,
		Tags:                     jobQueueDataV0.Tags,
		TagsAll:                  jobQueueDataV0.TagsAll,
		Timeouts:                 jobQueueDataV0.Timeouts,
	}

	if jobQueueDataV0.SchedulingPolicyARN.ValueString() == "" {
		jobQueueDataV1.SchedulingPolicyARN = fwtypes.ARNNull()
	}

	response.Diagnostics.Append(response.State.Set(ctx, jobQueueDataV1)...)
}

func upgradeJobQueueResourceStateV1toV2(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var jobQueueDataV1 resourceJobQueueDataV1
	response.Diagnostics.Append(request.State.Get(ctx, &jobQueueDataV1)...)
	if response.Diagnostics.HasError() {
		return
	}

	jobQueueDataV2 := jobQueueResourceModel{
		ComputeEnvironmentOrder:  jobQueueDataV1.ComputeEnvironmentOrder,
		ID:                       jobQueueDataV1.ID,
		JobQueueARN:              jobQueueDataV1.JobQueueARN,
		JobQueueName:             jobQueueDataV1.JobQueueName,
		JobStateTimeLimitActions: jobQueueDataV1.JobStateTimeLimitActions,
		Priority:                 jobQueueDataV1.Priority,
		State:                    jobQueueDataV1.State,
		Tags:                     jobQueueDataV1.Tags,
		TagsAll:                  jobQueueDataV1.TagsAll,
		Timeouts:                 jobQueueDataV1.Timeouts,
		SchedulingPolicyARN:      jobQueueDataV1.SchedulingPolicyARN,
	}

	if !jobQueueDataV1.ComputeEnvironments.IsNull() {
		var computeEnvironments []string
		diags := jobQueueDataV1.ComputeEnvironments.ElementsAs(ctx, &computeEnvironments, false)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		computeEnvironmentOrder := make([]*computeEnvironmentOrderModel, len(computeEnvironments))
		for i, env := range computeEnvironments {
			computeEnvironmentOrder[i] = &computeEnvironmentOrderModel{
				ComputeEnvironment: fwtypes.ARNValue(env),
				Order:              types.Int64Value(int64(i)),
			}
		}
		jobQueueDataV2.ComputeEnvironmentOrder = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, computeEnvironmentOrder)
	}

	response.Diagnostics.Append(response.State.Set(ctx, jobQueueDataV2)...)
}
