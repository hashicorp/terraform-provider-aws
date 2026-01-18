// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_notifications_managed_notification_additional_channel_association", name="Managed Notification Additional Channel Association")
func newManagedNotificationAdditionalChannelAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &managedNotificationAdditionalChannelAssociationResource{}

	return r, nil
}

type managedNotificationAdditionalChannelAssociationResource struct {
	framework.ResourceWithModel[managedNotificationAdditionalChannelAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *managedNotificationAdditionalChannelAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"managed_notification_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *managedNotificationAdditionalChannelAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data managedNotificationAdditionalChannelAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	input := notifications.AssociateManagedNotificationAdditionalChannelInput{
		ChannelArn:                          fwflex.StringFromFramework(ctx, data.ARN),
		ManagedNotificationConfigurationArn: fwflex.StringFromFramework(ctx, data.ManagedNotificationConfigurationARN),
	}

	_, err := conn.AssociateManagedNotificationAdditionalChannel(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating User Notifications Managed Notification Additional Channel Association (%s,%s)", data.ManagedNotificationConfigurationARN.ValueString(), data.ARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *managedNotificationAdditionalChannelAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data managedNotificationAdditionalChannelAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	managedNotificationConfigurationARN, channelArn := fwflex.StringValueFromFramework(ctx, data.ManagedNotificationConfigurationARN), fwflex.StringValueFromFramework(ctx, data.ARN)
	err := findManagedNotificationAdditionalChannelAssociationByTwoPartKey(ctx, conn, managedNotificationConfigurationARN, channelArn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Managed Notification Additional Channel Association (%s,%s)", managedNotificationConfigurationARN, channelArn), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *managedNotificationAdditionalChannelAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data managedNotificationAdditionalChannelAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	input := notifications.DisassociateManagedNotificationAdditionalChannelInput{
		ChannelArn:                          fwflex.StringFromFramework(ctx, data.ARN),
		ManagedNotificationConfigurationArn: fwflex.StringFromFramework(ctx, data.ManagedNotificationConfigurationARN),
	}

	_, err := conn.DisassociateManagedNotificationAdditionalChannel(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting User Notifications Managed Notification Additional Channel Association (%s,%s)", data.ManagedNotificationConfigurationARN.ValueString(), data.ARN.ValueString()), err.Error())

		return
	}
}

func (r *managedNotificationAdditionalChannelAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		managedNotificationAdditionalChannelAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, managedNotificationAdditionalChannelAssociationIDParts, false)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrARN), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("managed_notification_arn"), parts[0])...)
}

func findManagedNotificationAdditionalChannelAssociationByTwoPartKey(ctx context.Context, conn *notifications.Client, managedNotificationConfigurationArn, channelArn string) error {
	input := notifications.ListManagedNotificationChannelAssociationsInput{
		ManagedNotificationConfigurationArn: aws.String(managedNotificationConfigurationArn),
	}

	pages := notifications.NewListManagedNotificationChannelAssociationsPaginator(conn, &input)
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

		for _, association := range page.ChannelAssociations {
			if aws.ToString(association.ChannelIdentifier) == channelArn {
				return nil
			}
		}
	}

	return &retry.NotFoundError{
		LastRequest: &input,
	}
}

type managedNotificationAdditionalChannelAssociationResourceModel struct {
	ARN                                 fwtypes.ARN `tfsdk:"arn"`
	ManagedNotificationConfigurationARN fwtypes.ARN `tfsdk:"managed_notification_arn"`
}
