// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func upgradeJobQueueResourceStateV0toV1(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	type resourceJobQueueDataV0 struct {
		ComputeEnvironments types.List     `tfsdk:"compute_environments"`
		ID                  types.String   `tfsdk:"id"`
		JobQueueARN         types.String   `tfsdk:"arn"`
		JobQueueName        types.String   `tfsdk:"name"`
		Priority            types.Int64    `tfsdk:"priority"`
		SchedulingPolicyARN types.String   `tfsdk:"scheduling_policy_arn"`
		State               types.String   `tfsdk:"state"`
		Tags                tftags.Map     `tfsdk:"tags"`
		TagsAll             tftags.Map     `tfsdk:"tags_all"`
		Timeouts            timeouts.Value `tfsdk:"timeouts"`
	}

	var jobQueueDataV0 resourceJobQueueDataV0
	response.Diagnostics.Append(request.State.Get(ctx, &jobQueueDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	jobQueueDataV1 := jobQueueResourceModel{
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
