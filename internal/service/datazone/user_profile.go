// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

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

func (r *resourceUserProfile) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"details": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[detailsData](ctx),
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UserProfileStatus](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrType: schema.StringAttribute{
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
					stringplanmodifier.RequiresReplace(),
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
	state.ID = flex.StringToFramework(ctx, out.Id)
	resp.State.SetAttribute(ctx, path.Root(names.AttrID), out.Id) // set partial state to taint if wait fails

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := tfresource.RetryGWhenNotFound(ctx, createTimeout, func() (*datazone.GetUserProfileOutput, error) {
		return findUserProfileByID(ctx, conn, plan.DomainIdentifier.ValueString(), plan.UserIdentifier.ValueString(), out.Type)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameUserProfile, plan.UserIdentifier.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		in := datazone.UpdateUserProfileInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &in)...)

		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateUserProfile(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameUserProfile, plan.UserIdentifier.ValueString(), err),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		output, err := tfresource.RetryGWhenNotFound(ctx, updateTimeout, func() (*datazone.GetUserProfileOutput, error) {
			return findUserProfileByID(ctx, conn, plan.DomainIdentifier.ValueString(), plan.UserIdentifier.ValueString(), out.Type)
		})

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameUserProfile, plan.UserIdentifier.ValueString(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceUserProfile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) != 3 {
		resp.Diagnostics.AddError("resource import invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "user_identifier,domain_identifier,type"`, req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrType), parts[2])...)
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
	Type             fwtypes.StringEnum[awstypes.UserProfileType]   `tfsdk:"type"`
	UserType         fwtypes.StringEnum[awstypes.UserType]          `tfsdk:"user_type"`
	Timeouts         timeouts.Value                                 `tfsdk:"timeouts"`
}

type detailsData struct {
	IAM fwtypes.ListNestedObjectValueOf[iamUserProfileDetailsData] `tfsdk:"iam"`
	SSO fwtypes.ListNestedObjectValueOf[ssoUserProfileDetailsData] `tfsdk:"sso"`
}

type iamUserProfileDetailsData struct {
	ARN types.String `tfsdk:"arn"`
}

type ssoUserProfileDetailsData struct {
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	UserName  types.String `tfsdk:"user_name"`
}

var (
	_ flex.Flattener = &detailsData{}
)

func (d *detailsData) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.UserProfileDetailsMemberIam:
		var model iamUserProfileDetailsData
		di := flex.Flatten(ctx, t.Value, &model)
		diags.Append(di...)
		if diags.HasError() {
			return diags
		}

		d.IAM = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags

	case awstypes.UserProfileDetailsMemberSso:
		var model ssoUserProfileDetailsData
		di := flex.Flatten(ctx, t.Value, &model)
		diags.Append(di...)
		if diags.HasError() {
			return diags
		}

		d.SSO = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags

	default:
		return diags
	}
}
