// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_notifications_organizational_unit_association", name="Organizational Unit Association")
func newOrganizationalUnitAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &organizationalUnitAssociationResource{}

	return r, nil
}

type organizationalUnitAssociationResource struct {
	framework.ResourceWithModel[organizationalUnitAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *organizationalUnitAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"notification_configuration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organizational_unit_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *organizationalUnitAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data organizationalUnitAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	var input notifications.AssociateOrganizationalUnitInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.AssociateOrganizationalUnit(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating User Notifications Organizational Unit Association (%s,%s)", data.NotificationConfigurationARN.ValueString(), data.OrganizationalUnitID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *organizationalUnitAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data organizationalUnitAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	notificationConfigurationARN, organizationalUnitID := fwflex.StringValueFromFramework(ctx, data.NotificationConfigurationARN), fwflex.StringValueFromFramework(ctx, data.OrganizationalUnitID)
	err := findOrganizationalUnitAssociationByTwoPartKey(ctx, conn, notificationConfigurationARN, organizationalUnitID)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Organizational Unit Association (%s,%s)", notificationConfigurationARN, organizationalUnitID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *organizationalUnitAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data organizationalUnitAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	var input notifications.DisassociateOrganizationalUnitInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DisassociateOrganizationalUnit(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting User Notifications Organizational Unit Association (%s,%s)", data.NotificationConfigurationARN.ValueString(), data.OrganizationalUnitID.ValueString()), err.Error())

		return
	}
}

func (r *organizationalUnitAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		organizationalUnitAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, organizationalUnitAssociationIDParts, false)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("organizational_unit_id"), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("notification_configuration_arn"), parts[0])...)
}

func findOrganizationalUnitAssociationByTwoPartKey(ctx context.Context, conn *notifications.Client, notificationConfigurationArn, organizationalUnitID string) error {
	input := notifications.ListOrganizationalUnitsInput{
		NotificationConfigurationArn: aws.String(notificationConfigurationArn),
	}

	pages := notifications.NewListOrganizationalUnitsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		if err != nil {
			return err
		}

		if slices.Contains(page.OrganizationalUnits, organizationalUnitID) {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastRequest: &input,
	}
}

type organizationalUnitAssociationResourceModel struct {
	NotificationConfigurationARN fwtypes.ARN  `tfsdk:"notification_configuration_arn"`
	OrganizationalUnitID         types.String `tfsdk:"organizational_unit_id"`
}
