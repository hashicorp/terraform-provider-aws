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

// @FrameworkResource("aws_workspacesweb_user_settings_association", name="User Settings Association")
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.UserSettings")
// @Testing(importStateIdAttribute="user_settings_arn,portal_arn")
func newUserSettingsAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &userSettingsAssociationResource{}, nil
}

const (
	ResNameUserSettingsAssociation = "User Settings Association"
)

type userSettingsAssociationResource struct {
	framework.ResourceWithModel[userSettingsAssociationResourceModel]
}

func (r *userSettingsAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user_settings_arn": schema.StringAttribute{
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

func (r *userSettingsAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data userSettingsAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.AssociateUserSettingsInput{
		UserSettingsArn: data.UserSettingsARN.ValueStringPointer(),
		PortalArn:       data.PortalARN.ValueStringPointer(),
	}

	_, err := conn.AssociateUserSettings(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating WorkSpacesWeb %s", ResNameUserSettingsAssociation), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *userSettingsAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data userSettingsAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	// Check if the association exists by getting the user settings and checking associated portals
	output, err := findUserSettingsByARN(ctx, conn, data.UserSettingsARN.ValueString())
	if tfretry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb %s (%s)", ResNameUserSettingsAssociation, data.UserSettingsARN.ValueString()), err.Error())
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

func (r *userSettingsAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// This resource requires replacement on update since there's no true update operation
	response.Diagnostics.AddError("Update not supported", "This resource must be replaced to update")
}

func (r *userSettingsAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data userSettingsAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DisassociateUserSettingsInput{
		PortalArn: data.PortalARN.ValueStringPointer(),
	}

	_, err := conn.DisassociateUserSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb %s (%s)", ResNameUserSettingsAssociation, data.UserSettingsARN.ValueString()), err.Error())
		return
	}
}

func (r *userSettingsAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		userSettingsAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, userSettingsAssociationIDParts, true)
	if err != nil {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: user_settings_arn,portal_arn. Got: %q", request.ID),
		)
		return
	}
	userSettingsARN := parts[0]
	portalARN := parts[1]

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_settings_arn"), userSettingsARN)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("portal_arn"), portalARN)...)
}

type userSettingsAssociationResourceModel struct {
	framework.WithRegionModel
	UserSettingsARN types.String `tfsdk:"user_settings_arn"`
	PortalARN       types.String `tfsdk:"portal_arn"`
}
