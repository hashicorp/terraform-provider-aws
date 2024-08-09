// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkResource("aws_route53profiles_association", name="Association")
func newResourceAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAssociation = "Association"
)

type resourceAssociation struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_route53profiles_association"
}

func (r *resourceAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
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
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var state associationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &route53profiles.AssociateProfileInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, state, input)...)

	out, err := conn.AssociateProfile(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionCreating, ResNameAssociation, state.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ProfileAssociation == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionCreating, ResNameAssociation, state.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.ProfileAssociation.Id)

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	profileAssociation, err := waitAssociationCreated(ctx, conn, state.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionWaitingForCreation, ResNameAssociation, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, profileAssociation, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
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
		ProfileId:  aws.String(state.ProfileID.ValueString()),
		ResourceId: aws.String(state.ResourceID.ValueString()),
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
	return func() (interface{}, string, error) {
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
	ID            types.String                               `tfsdk:"id"`
	ResourceID    types.String                               `tfsdk:"resource_id"`
	ProfileID     types.String                               `tfsdk:"profile_id"`
	Name          types.String                               `tfsdk:"name"`
	OwnerId       types.String                               `tfsdk:"owner_id"`
	Status        fwtypes.StringEnum[awstypes.ProfileStatus] `tfsdk:"status"`
	StatusMessage types.String                               `tfsdk:"status_message"`
	Timeouts      timeouts.Value                             `tfsdk:"timeouts"`
}
