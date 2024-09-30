// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_user_profile", name="User Profile")
func newResourceUserProfile(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceUserProfile{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameUserProfile = "User Profile"
)

type resourceUserProfile struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpDelete
}

func (r *resourceUserProfile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_datazone_user_profile"
}

func (r *resourceUserProfile) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain_identifier": schema.StringAttribute{
				Required: true,
			},
			"details": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[detailsData](ctx),
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UserProfileStatus](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_identifier": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UserProfileType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UserType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *resourceUserProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var plan userProfileData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.CreateUserProfileInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.ClientToken = aws.String(id.UniqueId())
	out, err := conn.CreateUserProfile(ctx, in)
	if resp.Diagnostics.HasError() {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameUserProfile, plan.UserIdentifier.ValueString(), err),
			err.Error(),
		)
		return
	}

	state := plan
	resp.State.SetAttribute(ctx, path.Root(names.AttrID), state.ID) // set partial state to taint if wait fails

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := waitUserProfileCreated(ctx, conn, plan.DomainIdentifier.ValueString(), plan.ID.ValueString(), out.Type, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForCreation, ResNameUserProfile, plan.UserIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceUserProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var state userProfileData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserProfileByID(ctx, conn, state.DomainIdentifier.ValueString(), state.UserIdentifier.ValueString(), state.Type.ValueEnum())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameUserProfile, state.UserIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceUserProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state userProfileData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.UserType.Equal(state.UserType) {
		in := &datazone.UpdateUserProfileInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

		if resp.Diagnostics.HasError() {
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		out, err := waitUserProfileUpdated(ctx, conn, plan.DomainIdentifier.ValueString(), plan.ID.ValueString(), plan.Type.ValueEnum(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForCreation, ResNameUserProfile, plan.UserIdentifier.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceUserProfile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitUserProfileCreated(ctx context.Context, conn *datazone.Client, domainId string, userId string, userProfileType awstypes.UserProfileType, timeout time.Duration) (*datazone.GetUserProfileOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.UserProfileStatusNotAssigned, awstypes.UserProfileStatusDeactivated),
		Target:                    enum.Slice(awstypes.UserProfileStatusActivated, awstypes.UserProfileStatusActivated),
		Refresh:                   statusUserProfile(ctx, conn, domainId, userId, userProfileType),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetUserProfileOutput); ok {
		return out, err
	}

	return nil, err
}

func waitUserProfileUpdated(ctx context.Context, conn *datazone.Client, domainId string, userId string, userProfileType awstypes.UserProfileType, timeout time.Duration) (*datazone.GetUserProfileOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.UserProfileStatusNotAssigned, awstypes.UserProfileStatusDeactivated),
		Target:                    enum.Slice(awstypes.UserProfileStatusActivated, awstypes.UserProfileStatusActivated),
		Refresh:                   statusUserProfile(ctx, conn, domainId, userId, userProfileType),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetUserProfileOutput); ok {
		return out, err
	}

	return nil, err
}

func statusUserProfile(ctx context.Context, conn *datazone.Client, domainId string, userId string, userProfileType awstypes.UserProfileType) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findUserProfileByID(ctx, conn, domainId, userId, userProfileType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findUserProfileByID(ctx context.Context, conn *datazone.Client, domainId string, userId string, userProfileType awstypes.UserProfileType) (*datazone.GetUserProfileOutput, error) {
	in := &datazone.GetUserProfileInput{
		UserIdentifier:   aws.String(userId),
		DomainIdentifier: aws.String(domainId),
		Type:             userProfileType,
	}

	out, err := conn.GetUserProfile(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type userProfileData struct {
	DomainIdentifier types.String                                   `tfsdk:"domain_identifier"`
	Details          fwtypes.ListNestedObjectValueOf[detailsData]   `tfsdk:"details"`
	ID               types.String                                   `tfsdk:"id"`
	Status           fwtypes.StringEnum[awstypes.UserProfileStatus] `tfsdk:"status"`
	UserIdentifier   types.String                                   `tfsdk:"user_identifier"`
	Type             fwtypes.StringEnum[awstypes.UserProfileType]   `tfsdk:"user_profile_type"`
	UserType         fwtypes.StringEnum[awstypes.UserType]          `tfsdk:"user_type"`
	Timeouts         timeouts.Value                                 `tfsdk:"timeouts"`
}

type detailsData struct {
	IamUserProfileDetails fwtypes.ListNestedObjectValueOf[iamUserProfileDetailsData] `tfsdk:"iam_user_profile_details"`
	SsoUserProfileDetails fwtypes.ListNestedObjectValueOf[ssoUserProfileDetailsData] `tfsdk:"sso_user_profile_details"`
}

type iamUserProfileDetailsData struct {
	Arn types.String `tfsdk:"arn"`
}

type ssoUserProfileDetailsData struct {
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	UserName  types.String `tfsdk:"user_name"`
}
