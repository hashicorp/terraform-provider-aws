// Copyright IBM Corp. 2014, 2025
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @FrameworkResource("aws_workspacesweb_network_settings_association", name="Network Settings Association")
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.NetworkSettings")
// @Testing(importStateIdAttribute="network_settings_arn,portal_arn")
func newNetworkSettingsAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &networkSettingsAssociationResource{}, nil
}

type networkSettingsAssociationResource struct {
	framework.ResourceWithModel[networkSettingsAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *networkSettingsAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"network_settings_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"portal_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *networkSettingsAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data networkSettingsAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.AssociateNetworkSettingsInput{
		NetworkSettingsArn: data.NetworkSettingsARN.ValueStringPointer(),
		PortalArn:          data.PortalARN.ValueStringPointer(),
	}

	_, err := conn.AssociateNetworkSettings(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating WorkSpacesWeb Network Settings Association", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *networkSettingsAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data networkSettingsAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	// Check if the association exists by getting the network settings and checking associated portals
	output, err := findNetworkSettingsByARN(ctx, conn, data.NetworkSettingsARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Network Settings Association (%s)", data.NetworkSettingsARN.ValueString()), err.Error())
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

func (r *networkSettingsAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data networkSettingsAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DisassociateNetworkSettingsInput{
		PortalArn: data.PortalARN.ValueStringPointer(),
	}

	_, err := conn.DisassociateNetworkSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb Network Settings Association (%s)", data.NetworkSettingsARN.ValueString()), err.Error())
		return
	}
}

func (r *networkSettingsAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		networkSettingsAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, networkSettingsAssociationIDParts, true)
	if err != nil {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: network_settings_arn,portal_arn. Got: %q", request.ID),
		)
		return
	}
	networkSettingsARN := parts[0]
	portalARN := parts[1]

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("network_settings_arn"), networkSettingsARN)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("portal_arn"), portalARN)...)
}

type networkSettingsAssociationResourceModel struct {
	framework.WithRegionModel
	NetworkSettingsARN fwtypes.ARN `tfsdk:"network_settings_arn"`
	PortalARN          fwtypes.ARN `tfsdk:"portal_arn"`
}
