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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const userJourneyImportIDPartCount = 2

// @FrameworkResource("aws_resiliencehubv2_user_journey", name="User Journey")
func newResourceUserJourney(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceUserJourney{}, nil
}

type resourceUserJourney struct {
	framework.ResourceWithModel[resourceUserJourneyModel]
}

func (r *resourceUserJourney) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrID: fwschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"system_arn": fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Required: true,
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
			},
			"policy_arn": fwschema.StringAttribute{
				Optional: true,
			},
			"user_journey_id": fwschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceUserJourney) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceUserJourneyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.CreateUserJourneyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateUserJourney(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	plan.UserJourneyId = types.StringPointerValue(output.UserJourney.UserJourneyId)
	plan.ID = types.StringValue(plan.SystemArn.ValueString() + "," + plan.UserJourneyId.ValueString())

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.UserJourney, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(plan.SystemArn.ValueString() + "," + plan.UserJourneyId.ValueString())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceUserJourney) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceUserJourneyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	uj, err := findUserJourneyByID(ctx, conn, state.SystemArn.ValueString(), state.UserJourneyId.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, uj, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(state.SystemArn.ValueString() + "," + state.UserJourneyId.ValueString())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *resourceUserJourney) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceUserJourneyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.UpdateUserJourneyInput{
		SystemArn:     state.SystemArn.ValueStringPointer(),
		UserJourneyId: state.UserJourneyId.ValueStringPointer(),
		Name:          plan.Name.ValueStringPointer(),
		Description:   plan.Description.ValueStringPointer(),
	}
	if !plan.PolicyArn.IsNull() {
		input.PolicyArn = plan.PolicyArn.ValueStringPointer()
	}

	output, err := conn.UpdateUserJourney(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.UserJourney, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SystemArn = state.SystemArn
	plan.ID = state.ID

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceUserJourney) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceUserJourneyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.DeleteUserJourneyInput{
		SystemArn:     state.SystemArn.ValueStringPointer(),
		UserJourneyId: state.UserJourneyId.ValueStringPointer(),
	}
	_, err := conn.DeleteUserJourney(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
	}
}

func (r *resourceUserJourney) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, userJourneyImportIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("system_arn"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_journey_id"), parts[1])...)
}

func findUserJourneyByID(ctx context.Context, conn *resiliencehubv2.Client, systemArn, userJourneyId string) (*awstypes.UserJourney, error) {
	input := resiliencehubv2.GetUserJourneyInput{
		SystemArn:     aws.String(systemArn),
		UserJourneyId: aws.String(userJourneyId),
	}
	output, err := conn.GetUserJourney(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		return nil, smarterr.NewError(err)
	}
	if output == nil || output.UserJourney == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return output.UserJourney, nil
}

type resourceUserJourneyModel struct {
	framework.WithRegionModel
	Description   types.String `tfsdk:"description"`
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	PolicyArn     types.String `tfsdk:"policy_arn"`
	SystemArn     types.String `tfsdk:"system_arn"`
	UserJourneyId types.String `tfsdk:"user_journey_id"`
}
