// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_notifications_channel_association", name="Channel Association")
func newResourceChannelAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceChannelAssociation{}

	return r, nil
}

const (
	ResNameChannelAssociation    = "Channel Association"
	ChannelAssociationsARNsCount = 2
)

type resourceChannelAssociation struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceChannelAssociationModel]
}

func (r *resourceChannelAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"notification_configuration_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceChannelAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var plan resourceChannelAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input notifications.AssociateChannelInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.AssociateChannel(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionCreating, ResNameChannelAssociation, plan.NotificationConfigurationARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionCreating, ResNameChannelAssociation, plan.NotificationConfigurationARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceChannelAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var state resourceChannelAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	exists, err := findChannelAssociationByARNs(ctx, conn, state.ARN.ValueString(), state.NotificationConfigurationARN.ValueString())
	if !exists {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionReading, ResNameChannelAssociation, state.NotificationConfigurationARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceChannelAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var state resourceChannelAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := notifications.DisassociateChannelInput{
		Arn:                          state.ARN.ValueStringPointer(),
		NotificationConfigurationArn: state.NotificationConfigurationARN.ValueStringPointer(),
	}

	_, err := conn.DisassociateChannel(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionDeleting, ResNameChannelAssociation, state.NotificationConfigurationARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceChannelAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, ChannelAssociationsARNsCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: channel_arn,notification_configuration_arn. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrARN), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("notification_configuration_arn"), parts[1])...)
}

func findChannelAssociationByARNs(ctx context.Context, conn *notifications.Client, arn string, notificationConfigurationArn string) (bool, error) {
	input := notifications.ListChannelsInput{
		NotificationConfigurationArn: aws.String(notificationConfigurationArn),
	}

	out, err := conn.ListChannels(ctx, &input)
	if err != nil {
		return false, err
	}

	if out == nil || out.Channels == nil || len(out.Channels) == 0 {
		return false, tfresource.NewEmptyResultError(&input)
	}

	if slices.Contains(out.Channels, arn) {
		return true, nil
	}

	return false, &retry.NotFoundError{
		LastError:   fmt.Errorf("association of channel %q to notification configuration %q not found", arn, notificationConfigurationArn),
		LastRequest: &input,
	}
}

type resourceChannelAssociationModel struct {
	ARN                          types.String `tfsdk:"arn"`
	NotificationConfigurationARN types.String `tfsdk:"notification_configuration_arn"`
}

func sweepChannelAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := notifications.ListChannelsInput{}
	conn := client.NotificationsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := notifications.NewListChannelsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, arn := range page.Channels {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceChannelAssociation, client,
				sweepfw.NewAttribute(names.AttrARN, arn)),
			)
		}
	}

	return sweepResources, nil
}
