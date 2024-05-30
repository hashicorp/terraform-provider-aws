// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceFrameworkShare(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceFrameworkShare{}, nil
}

const (
	ResNameFrameworkShare = "FrameworkShare"
)

type resourceFrameworkShare struct {
	framework.ResourceWithConfigure
}

func (r *resourceFrameworkShare) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_auditmanager_framework_share"
}

func (r *resourceFrameworkShare) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrComment: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_account": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_region": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"framework_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceFrameworkShare) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan resourceFrameworkShareData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := auditmanager.StartAssessmentFrameworkShareInput{
		DestinationAccount: aws.String(plan.DestinationAccount.ValueString()),
		DestinationRegion:  aws.String(plan.DestinationRegion.ValueString()),
		FrameworkId:        aws.String(plan.FrameworkID.ValueString()),
	}
	if !plan.Comment.IsNull() {
		in.Comment = aws.String(plan.Comment.ValueString())
	}
	out, err := conn.StartAssessmentFrameworkShare(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameFrameworkShare, plan.FrameworkID.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil || out.AssessmentFrameworkShareRequest == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameFrameworkShare, plan.FrameworkID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.refreshFromOutput(ctx, out.AssessmentFrameworkShareRequest)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceFrameworkShare) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceFrameworkShareData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindFrameworkShareByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionReading, ResNameFrameworkShare, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.refreshFromOutput(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op. Changing any of account, region, or framework_id will result
// a destroy and replace.
func (r *resourceFrameworkShare) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceFrameworkShare) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceFrameworkShareData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Framework share requests in certain statuses must be revoked before deletion
	if CanBeRevoked(state.Status.ValueString()) {
		in := auditmanager.UpdateAssessmentFrameworkShareInput{
			RequestId:   aws.String(state.ID.ValueString()),
			RequestType: awstypes.ShareRequestTypeSent,
			Action:      awstypes.ShareRequestActionRevoke,
		}
		_, err := conn.UpdateAssessmentFrameworkShare(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameFrameworkShare, state.ID.String(), nil),
				err.Error(),
			)
		}
	}

	in := auditmanager.DeleteAssessmentFrameworkShareInput{
		RequestId:   aws.String(state.ID.ValueString()),
		RequestType: awstypes.ShareRequestTypeSent,
	}
	_, err := conn.DeleteAssessmentFrameworkShare(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameFrameworkShare, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceFrameworkShare) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindFrameworkShareByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.AssessmentFrameworkShareRequest, error) {
	in := &auditmanager.ListAssessmentFrameworkShareRequestsInput{
		RequestType: awstypes.ShareRequestTypeSent,
	}
	pages := auditmanager.NewListAssessmentFrameworkShareRequestsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, share := range page.AssessmentFrameworkShareRequests {
			if id == aws.ToString(share.Id) {
				return &share, nil
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastRequest: in,
	}
}

// CanBeRevoked verifies a framework share is in a status which can be revoked
func CanBeRevoked(status string) bool {
	nonRevokable := enum.Slice(
		awstypes.ShareRequestStatusDeclined,
		awstypes.ShareRequestStatusExpired,
		awstypes.ShareRequestStatusFailed,
		awstypes.ShareRequestStatusRevoked,
	)
	for _, s := range nonRevokable {
		if s == status {
			return false
		}
	}
	return true
}

type resourceFrameworkShareData struct {
	Comment            types.String `tfsdk:"comment"`
	DestinationAccount types.String `tfsdk:"destination_account"`
	DestinationRegion  types.String `tfsdk:"destination_region"`
	FrameworkID        types.String `tfsdk:"framework_id"`
	ID                 types.String `tfsdk:"id"`
	Status             types.String `tfsdk:"status"`
}

// refreshFromOutput writes state data from an AWS response object
func (rd *resourceFrameworkShareData) refreshFromOutput(ctx context.Context, out *awstypes.AssessmentFrameworkShareRequest) {
	if out == nil {
		return
	}

	rd.Comment = flex.StringToFramework(ctx, out.Comment)
	rd.DestinationAccount = flex.StringToFramework(ctx, out.DestinationAccount)
	rd.DestinationRegion = flex.StringToFramework(ctx, out.DestinationRegion)
	rd.FrameworkID = flex.StringToFramework(ctx, out.FrameworkId)
	rd.ID = flex.StringToFramework(ctx, out.Id)
	rd.Status = flex.StringValueToFramework(ctx, out.Status)
}
