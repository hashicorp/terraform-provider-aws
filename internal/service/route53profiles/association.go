// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_route53profiles_association", name="Association")
// @Tags(identifierAttribute="arn")
func newResourceAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute) // tags only
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAssociation = "Association"
)

var (
	// Converted from (?!^[0-9]+$)([a-zA-Z0-9\-_' ']+) because of the negative lookahead.
	resourceAssociationNameRegex = regexache.MustCompile("(^[^0-9][a-zA-Z0-9\\-_' ']+$)")
)

type resourceAssociation struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[associationResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(resourceAssociationNameRegex, ""),
					stringvalidator.LengthBetween(0, 64),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"profile_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrResourceID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProfileStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *resourceAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var data associationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &route53profiles.AssociateProfileInput{
		Name: data.Name.ValueStringPointer(),
		Tags: getTagsInSlice(ctx),
	}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)

	out, err := conn.AssociateProfile(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionCreating, ResNameAssociation, data.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ProfileAssociation == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionCreating, ResNameAssociation, data.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	data.ID = flex.StringToFramework(ctx, out.ProfileAssociation.Id)
	associationARN := r.Meta().RegionalARN(ctx, "route53profiles", fmt.Sprintf("profile-association/%s", aws.ToString(out.ProfileAssociation.Id)))
	data.ARN = flex.StringValueToFramework(ctx, associationARN)

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	profileAssociation, err := waitAssociationCreated(ctx, conn, data.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionWaitingForCreation, ResNameAssociation, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, profileAssociation, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *resourceAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var state associationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAssociationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionSetting, ResNameAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	associationARN := r.Meta().RegionalARN(ctx, "route53profiles", fmt.Sprintf("profile-association/%s", aws.ToString(out.Id)))
	state.ARN = flex.StringValueToFramework(ctx, associationARN)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var state associationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &route53profiles.DisassociateProfileInput{
		ProfileId:  state.ProfileID.ValueStringPointer(),
		ResourceId: state.ResourceID.ValueStringPointer(),
	}

	_, err := conn.DisassociateProfile(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionDeleting, ResNameAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAssociationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionWaitingForDeletion, ResNameAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitAssociationCreated(ctx context.Context, conn *route53profiles.Client, id string, timeout time.Duration) (*awstypes.ProfileAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProfileStatusCreating),
		Target:                    enum.Slice(awstypes.ProfileStatusComplete),
		Refresh:                   statusAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ProfileAssociation); ok {
		return out, err
	}

	return nil, err
}

func waitAssociationDeleted(ctx context.Context, conn *route53profiles.Client, id string, timeout time.Duration) (*awstypes.ProfileAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProfileStatusDeleting),
		Target:  []string{},
		Refresh: statusAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ProfileAssociation); ok {
		return out, err
	}

	return nil, err
}

func statusAssociation(ctx context.Context, conn *route53profiles.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findAssociationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findAssociationByID(ctx context.Context, conn *route53profiles.Client, id string) (*awstypes.ProfileAssociation, error) {
	in := &route53profiles.GetProfileAssociationInput{
		ProfileAssociationId: aws.String(id),
	}

	out, err := conn.GetProfileAssociation(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ProfileAssociation == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ProfileAssociation, nil
}

type associationResourceModel struct {
	ARN           types.String                               `tfsdk:"arn"`
	ID            types.String                               `tfsdk:"id"`
	ResourceID    types.String                               `tfsdk:"resource_id"`
	ProfileID     types.String                               `tfsdk:"profile_id"`
	Name          types.String                               `tfsdk:"name"`
	OwnerId       types.String                               `tfsdk:"owner_id"`
	Status        fwtypes.StringEnum[awstypes.ProfileStatus] `tfsdk:"status"`
	StatusMessage types.String                               `tfsdk:"status_message"`
	Tags          tftags.Map                                 `tfsdk:"tags"`
	TagsAll       tftags.Map                                 `tfsdk:"tags_all"`
	Timeouts      timeouts.Value                             `tfsdk:"timeouts"`
}
