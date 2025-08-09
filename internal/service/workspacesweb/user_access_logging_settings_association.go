// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfretry "github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @FrameworkResource("aws_workspacesweb_user_access_logging_settings_association", name="User Access Logging Settings Association")
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.UserAccessLoggingSettings")
// @Testing(importStateIdAttribute="user_access_logging_settings_arn,portal_arn")
func newUserAccessLoggingSettingsAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &userAccessLoggingSettingsAssociationResource{}, nil
}

const (
	ResNameUserAccessLoggingSettingsAssociation = "User Access Logging Settings Association"
)

type userAccessLoggingSettingsAssociationResource struct {
	framework.ResourceWithModel[userAccessLoggingSettingsAssociationResourceModel]
}

func (r *userAccessLoggingSettingsAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user_access_logging_settings_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"portal_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *userAccessLoggingSettingsAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data userAccessLoggingSettingsAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.AssociateUserAccessLoggingSettingsInput{
		UserAccessLoggingSettingsArn: data.UserAccessLoggingSettingsARN.ValueStringPointer(),
		PortalArn:                    data.PortalARN.ValueStringPointer(),
	}

	_, err := conn.AssociateUserAccessLoggingSettings(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating WorkSpacesWeb %s", ResNameUserAccessLoggingSettingsAssociation), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *userAccessLoggingSettingsAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data userAccessLoggingSettingsAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	// Check if the association exists by getting the user access logging settings and checking associated portals
	output, err := findUserAccessLoggingSettingsByARN(ctx, conn, data.UserAccessLoggingSettingsARN.ValueString())
	if tfretry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb %s (%s)", ResNameUserAccessLoggingSettingsAssociation, data.UserAccessLoggingSettingsARN.ValueString()), err.Error())
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

func (r *userAccessLoggingSettingsAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// This resource requires replacement on update since there's no true update operation
	response.Diagnostics.AddError("Update not supported", "This resource must be replaced to update")
}

func (r *userAccessLoggingSettingsAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data userAccessLoggingSettingsAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DisassociateUserAccessLoggingSettingsInput{
		PortalArn: data.PortalARN.ValueStringPointer(),
	}

	_, err := conn.DisassociateUserAccessLoggingSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb %s (%s)", ResNameUserAccessLoggingSettingsAssociation, data.UserAccessLoggingSettingsARN.ValueString()), err.Error())
		return
	}
}

func (r *userAccessLoggingSettingsAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		userAccessLoggingSettingsAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, userAccessLoggingSettingsAssociationIDParts, true)
	if err != nil {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: user_access_logging_settings_arn,portal_arn. Got: %q", request.ID),
		)
		return
	}
	userAccessLoggingSettingsARN := parts[0]
	portalARN := parts[1]

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_access_logging_settings_arn"), userAccessLoggingSettingsARN)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("portal_arn"), portalARN)...)
}

type userAccessLoggingSettingsAssociationResourceModel struct {
	framework.WithRegionModel
	UserAccessLoggingSettingsARN types.String `tfsdk:"user_access_logging_settings_arn"`
	PortalARN                    types.String `tfsdk:"portal_arn"`
}
