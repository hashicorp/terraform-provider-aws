// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_user_profile", name="User Profile")
// @IdentityAttribute("domain_identifier")
// @IdentityAttribute("user_identifier")
// @ImportIDHandler("userProfileImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/datazone;datazone.GetUserProfileOutput", checkDestroyNoop=true)
// @Testing(importStateIdFunc="testAccUserProfileImportStateFunc", importStateIdAttribute="domain_identifier")
// @Testing(importIgnore="user_type", plannableImportAction=Replace)
// @Testing(preIdentityVersion="v6.47.0")
func newUserProfileResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &userProfileResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameUserProfile = "User Profile"
)

type userProfileResource struct {
	framework.ResourceWithModel[userProfileResourceModel]
	framework.WithTimeouts
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *userProfileResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *userProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var plan userProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.CreateUserProfileInput{}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, in))
	if resp.Diagnostics.HasError() {
		return
	}

	in.ClientToken = aws.String(create.UniqueId(ctx))
	out, err := conn.CreateUserProfile(ctx, in)
	if resp.Diagnostics.HasError() {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserIdentifier.ValueString())
		return
	}

	state := plan
	state.ID = flex.StringToFramework(ctx, out.Id)
	resp.State.SetAttribute(ctx, path.Root(names.AttrID), out.Id) // set partial state to taint if wait fails

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := tfresource.RetryWhenNotFound(ctx, createTimeout, func(ctx context.Context) (*datazone.GetUserProfileOutput, error) {
		return findUserProfileByID(ctx, conn, plan.DomainIdentifier.ValueString(), plan.UserIdentifier.ValueString(), out.Type)
	})

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserIdentifier.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *userProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var state userProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserProfileByID(ctx, conn, state.DomainIdentifier.ValueString(), state.UserIdentifier.ValueString(), state.Type.ValueEnum())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserIdentifier.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *userProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state userProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		in := datazone.UpdateUserProfileInput{}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &in))

		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateUserProfile(ctx, &in)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserIdentifier.ValueString())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		output, err := tfresource.RetryWhenNotFound(ctx, updateTimeout, func(ctx context.Context) (*datazone.GetUserProfileOutput, error) {
			return findUserProfileByID(ctx, conn, plan.DomainIdentifier.ValueString(), plan.UserIdentifier.ValueString(), out.Type)
		})

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserIdentifier.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func findUserProfileByID(ctx context.Context, conn *datazone.Client, domainId string, userId string, userProfileType awstypes.UserProfileType) (*datazone.GetUserProfileOutput, error) {
	in := &datazone.GetUserProfileInput{
		UserIdentifier:   aws.String(userId),
		DomainIdentifier: aws.String(domainId),
		Type:             userProfileType,
	}

	out, err := conn.GetUserProfile(ctx, in)

	if isResourceMissing(err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type userProfileImportID struct{}

var _ inttypes.ImportIDParser = userProfileImportID{}

func (userProfileImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.SplitN(id, ",", 3)
	if len(parts) != 3 {
		return "", nil, fmt.Errorf("id %q should be in the format <user-identifier>,<domain-identifier>,<type>", id)
	}

	result := map[string]any{
		"user_identifier":   parts[0],
		"domain_identifier": parts[1],
	}

	return id, result, nil
}

type userProfileResourceModel struct {
	framework.WithRegionModel
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
