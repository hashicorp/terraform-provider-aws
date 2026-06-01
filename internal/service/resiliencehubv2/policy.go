// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehubv2_policy", name="Policy")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.Policy")
// @Testing(hasNoPreExistingResource=true)
func newResourcePolicy(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourcePolicy{}, nil
}

type resourcePolicy struct {
	framework.ResourceWithModel[resourcePolicyModel]
	framework.WithImportByIdentity
}

func (r *resourcePolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"availability_slo": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[availabilitySloModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						names.AttrTarget: fwschema.Float64Attribute{
							Required: true,
						},
					},
				},
			},
			"data_recovery": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataRecoveryModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"time_between_backups_in_minutes": fwschema.Int32Attribute{
							Required: true,
						},
					},
				},
			},
			"multi_az": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiAzModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"disaster_recovery_approach": fwschema.StringAttribute{
							Required: true,
						},
						"rpo_in_minutes": fwschema.Int32Attribute{
							Optional: true,
						},
						"rto_in_minutes": fwschema.Int32Attribute{
							Optional: true,
						},
					},
				},
			},
			"multi_region": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiRegionModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"disaster_recovery_approach": fwschema.StringAttribute{
							Required: true,
						},
						"rpo_in_minutes": fwschema.Int32Attribute{
							Optional: true,
						},
						"rto_in_minutes": fwschema.Int32Attribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourcePolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourcePolicyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.CreatePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Policy, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = types.StringValue(aws.ToString(output.Policy.PolicyArn))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourcePolicyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	policy, err := findPolicyByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, policy, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.ARN = types.StringValue(aws.ToString(policy.PolicyArn))

	tags, err := listTags(ctx, conn, state.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
	setTagsOut(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *resourcePolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourcePolicyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.UpdatePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.PolicyArn = state.ARN.ValueStringPointer()

	output, err := conn.UpdatePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Policy, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = types.StringValue(aws.ToString(output.Policy.PolicyArn))

	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := updateTags(ctx, conn, state.ARN.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourcePolicyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.DeletePolicyInput{
		PolicyArn: state.ARN.ValueStringPointer(),
	}
	_, err := conn.DeletePolicy(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
	}
}

func findPolicyByARN(ctx context.Context, conn *resiliencehubv2.Client, arn string) (*awstypes.Policy, error) {
	input := resiliencehubv2.GetPolicyInput{
		PolicyArn: aws.String(arn),
	}
	output, err := conn.GetPolicy(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		return nil, smarterr.NewError(err)
	}
	if output == nil || output.Policy == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return output.Policy, nil
}

type resourcePolicyModel struct {
	framework.WithRegionModel
	ARN             types.String                                          `tfsdk:"arn"`
	AvailabilitySlo fwtypes.ListNestedObjectValueOf[availabilitySloModel] `tfsdk:"availability_slo"`
	DataRecovery    fwtypes.ListNestedObjectValueOf[dataRecoveryModel]    `tfsdk:"data_recovery"`
	Description     types.String                                          `tfsdk:"description"`
	MultiAz         fwtypes.ListNestedObjectValueOf[multiAzModel]         `tfsdk:"multi_az"`
	MultiRegion     fwtypes.ListNestedObjectValueOf[multiRegionModel]     `tfsdk:"multi_region"`
	Name            types.String                                          `tfsdk:"name"`
	Tags            tftags.Map                                            `tfsdk:"tags"`
	TagsAll         tftags.Map                                            `tfsdk:"tags_all"`
}

type availabilitySloModel struct {
	Target types.Float64 `tfsdk:"target"`
}

type dataRecoveryModel struct {
	TimeBetweenBackupsInMinutes types.Int32 `tfsdk:"time_between_backups_in_minutes"`
}

type multiAzModel struct {
	DisasterRecoveryApproach types.String `tfsdk:"disaster_recovery_approach"`
	RpoInMinutes             types.Int32  `tfsdk:"rpo_in_minutes"`
	RtoInMinutes             types.Int32  `tfsdk:"rto_in_minutes"`
}

type multiRegionModel struct {
	DisasterRecoveryApproach types.String `tfsdk:"disaster_recovery_approach"`
	RpoInMinutes             types.Int32  `tfsdk:"rpo_in_minutes"`
	RtoInMinutes             types.Int32  `tfsdk:"rto_in_minutes"`
}
