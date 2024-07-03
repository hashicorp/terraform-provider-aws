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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_route53profiles_resource_association", name="ResourceAssociation")
func newResourceResourceAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceResourceAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameResourceAssociation = "ResourceAssociation"
)

type resourceResourceAssociation struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceResourceAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_route53profiles_resource_association"
}

func (r *resourceResourceAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_properties": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.JSON(),
				},
			},
			names.AttrResourceType: schema.StringAttribute{
				Computed: true,
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

func (r *resourceResourceAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var plan resourceAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &route53profiles.AssociateResourceToProfileInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

	out, err := conn.AssociateResourceToProfile(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionCreating, ResNameResourceAssociation, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ProfileResourceAssociation == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionCreating, ResNameResourceAssociation, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.ProfileResourceAssociation.Id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	resourceAssociation, err := waitResourceAssociationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionWaitingForCreation, ResNameResourceAssociation, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if resourceAssociation != nil {
		plan.OwnerId = flex.StringToFramework(ctx, resourceAssociation.OwnerId)
		plan.ResourceProperties = flex.StringToFramework(ctx, resourceAssociation.ResourceProperties)
		plan.ResourceType = flex.StringToFramework(ctx, resourceAssociation.ResourceType)
		plan.Status = fwtypes.StringEnumValue[awstypes.ProfileStatus](resourceAssociation.Status)
		plan.StatusMessage = flex.StringToFramework(ctx, resourceAssociation.StatusMessage)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResourceAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var state resourceAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourceAssociationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionSetting, ResNameResourceAssociation, state.ID.String(), err),
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

func (r *resourceResourceAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var state resourceAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &route53profiles.DisassociateResourceFromProfileInput{
		ProfileId:   aws.String(state.ProfileID.ValueString()),
		ResourceArn: aws.String(state.ResourceArn.ValueString()),
	}

	_, err := conn.DisassociateResourceFromProfile(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionDeleting, ResNameResourceAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitResourceAssociationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionWaitingForDeletion, ResNameResourceAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResourceAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitResourceAssociationCreated(ctx context.Context, conn *route53profiles.Client, id string, timeout time.Duration) (*awstypes.ProfileResourceAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProfileStatusCreating, awstypes.ProfileStatusUpdating),
		Target:                    enum.Slice(awstypes.ProfileStatusComplete),
		Refresh:                   statusResourceAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ProfileResourceAssociation); ok {
		return out, err
	}

	return nil, err
}

func waitResourceAssociationDeleted(ctx context.Context, conn *route53profiles.Client, id string, timeout time.Duration) (*awstypes.ProfileResourceAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProfileStatusDeleting),
		Target:  []string{},
		Refresh: statusResourceAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ProfileResourceAssociation); ok {
		return out, err
	}

	return nil, err
}

func statusResourceAssociation(ctx context.Context, conn *route53profiles.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findResourceAssociationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findResourceAssociationByID(ctx context.Context, conn *route53profiles.Client, id string) (*awstypes.ProfileResourceAssociation, error) {
	in := &route53profiles.GetProfileResourceAssociationInput{
		ProfileResourceAssociationId: aws.String(id),
	}

	out, err := conn.GetProfileResourceAssociation(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ProfileResourceAssociation == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ProfileResourceAssociation, nil
}

type resourceAssociationResourceModel struct {
	ID                 types.String                               `tfsdk:"id"`
	Name               types.String                               `tfsdk:"name"`
	OwnerId            types.String                               `tfsdk:"owner_id"`
	ProfileID          types.String                               `tfsdk:"profile_id"`
	ResourceArn        types.String                               `tfsdk:"resource_arn"`
	ResourceProperties types.String                               `tfsdk:"resource_properties"`
	ResourceType       types.String                               `tfsdk:"resource_type"`
	Status             fwtypes.StringEnum[awstypes.ProfileStatus] `tfsdk:"status"`
	StatusMessage      types.String                               `tfsdk:"status_message"`
	Timeouts           timeouts.Value                             `tfsdk:"timeouts"`
}
