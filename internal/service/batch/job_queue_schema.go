// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(batch.JQState_Values()...),
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
		},
	}
}

func upgradeJobQueueResourceStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	type resourceJobQueueDataV0 struct {
		ARN                 types.String   `tfsdk:"arn"`
		ComputeEnvironments types.List     `tfsdk:"compute_environments"`
		ID                  types.String   `tfsdk:"id"`
		Name                types.String   `tfsdk:"name"`
		Priority            types.Int64    `tfsdk:"priority"`
		SchedulingPolicyARN types.String   `tfsdk:"scheduling_policy_arn"`
		State               types.String   `tfsdk:"state"`
		Tags                types.Map      `tfsdk:"tags"`
		TagsAll             types.Map      `tfsdk:"tags_all"`
		Timeouts            timeouts.Value `tfsdk:"timeouts"`
	}

	var jobQueueDataV0 resourceJobQueueDataV0

	resp.Diagnostics.Append(req.State.Get(ctx, &jobQueueDataV0)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ceo := fwtypes.NewListNestedObjectValueOfNull[computeEnvironmentOrder](ctx)

	jobQueueDataV2 := resourceJobQueueData{
		ComputeEnvironments:     jobQueueDataV0.ComputeEnvironments,
		ComputeEnvironmentOrder: ceo,
		ID:                      jobQueueDataV0.ID,
		JobQueueName:            jobQueueDataV0.Name,
		Priority:                jobQueueDataV0.Priority,
		State:                   jobQueueDataV0.State,
		Tags:                    jobQueueDataV0.Tags,
		TagsAll:                 jobQueueDataV0.TagsAll,
		Timeouts:                jobQueueDataV0.Timeouts,
	}

	if jobQueueDataV0.SchedulingPolicyARN.ValueString() == "" {
		jobQueueDataV2.SchedulingPolicyARN = fwtypes.ARNNull()
	}

	diags := resp.State.Set(ctx, jobQueueDataV2)
	resp.Diagnostics.Append(diags...)
}
