// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package workspacesweb

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @FrameworkResource("aws_workspacesweb_session_logger_association", name="Session Logger Association")
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.SessionLogger")
// @Testing(importStateIdAttribute="session_logger_arn,portal_arn")
func newSessionLoggerAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &sessionLoggerAssociationResource{}, nil
}

type sessionLoggerAssociationResource struct {
	framework.ResourceWithModel[sessionLoggerAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *sessionLoggerAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"portal_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"session_logger_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *sessionLoggerAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data sessionLoggerAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.AssociateSessionLoggerInput{
		PortalArn:        data.PortalARN.ValueStringPointer(),
		SessionLoggerArn: data.SessionLoggerARN.ValueStringPointer(),
	}

	_, err := conn.AssociateSessionLogger(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating WorkSpacesWeb Session Logger Association", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *sessionLoggerAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data sessionLoggerAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	// Check if the association exists by getting the session logger and checking associated portals
	output, err := findSessionLoggerByARN(ctx, conn, data.SessionLoggerARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Session Logger Association (%s)", data.SessionLoggerARN.ValueString()), err.Error())
		return
	}

	// Check if the portal is in the associated portals list
	if !slices.Contains(output.AssociatedPortalArns, data.PortalARN.ValueString()) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(fmt.Errorf("association not found")))
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *sessionLoggerAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data sessionLoggerAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DisassociateSessionLoggerInput{
		PortalArn: data.PortalARN.ValueStringPointer(),
	}

	_, err := conn.DisassociateSessionLogger(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb Session Logger Association (%s)", data.SessionLoggerARN.ValueString()), err.Error())
		return
	}
}

func (r *sessionLoggerAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		sessionLoggerAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, sessionLoggerAssociationIDParts, true)
	if err != nil {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: session_logger_arn,portal_arn. Got: %q", request.ID),
		)
		return
	}
	sessionLoggerARN := parts[0]
	portalARN := parts[1]

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("session_logger_arn"), sessionLoggerARN)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("portal_arn"), portalARN)...)
}

type sessionLoggerAssociationResourceModel struct {
	framework.WithRegionModel
	PortalARN        fwtypes.ARN `tfsdk:"portal_arn"`
	SessionLoggerARN fwtypes.ARN `tfsdk:"session_logger_arn"`
}
