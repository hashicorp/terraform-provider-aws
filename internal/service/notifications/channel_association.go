// Copyright (c) HashiCorp, Inc.
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

// @FrameworkResource("aws_notifications_channel_association", name="Channel Association")
func newChannelAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &channelAssociationResource{}

	return r, nil
}

type channelAssociationResource struct {
	framework.ResourceWithModel[channelAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *channelAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"notification_configuration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *channelAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data channelAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	var input notifications.AssociateChannelInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.AssociateChannel(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating User Notifications Channel Association (%s,%s)", data.NotificationConfigurationARN.ValueString(), data.ARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *channelAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data channelAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	notificationConfigurationARN, arn := fwflex.StringValueFromFramework(ctx, data.NotificationConfigurationARN), fwflex.StringValueFromFramework(ctx, data.ARN)
	err := findChannelAssociationByTwoPartKey(ctx, conn, notificationConfigurationARN, arn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Channel Association (%s,%s)", notificationConfigurationARN, arn), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *channelAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data channelAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	var input notifications.DisassociateChannelInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DisassociateChannel(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting User Notifications Channel Association (%s,%s)", data.NotificationConfigurationARN.ValueString(), data.ARN.ValueString()), err.Error())

		return
	}
}

func (r *channelAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		channelAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, channelAssociationIDParts, false)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrARN), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("notification_configuration_arn"), parts[0])...)
}

func findChannelAssociationByTwoPartKey(ctx context.Context, conn *notifications.Client, notificationConfigurationArn, arn string) error {
	input := notifications.ListChannelsInput{
		NotificationConfigurationArn: aws.String(notificationConfigurationArn),
	}

	pages := notifications.NewListChannelsPaginator(conn, &input)
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

		if slices.Contains(page.Channels, arn) {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastRequest: &input,
	}
}

type channelAssociationResourceModel struct {
	ARN                          fwtypes.ARN `tfsdk:"arn"`
	NotificationConfigurationARN fwtypes.ARN `tfsdk:"notification_configuration_arn"`
}
