// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_notifications_notification_hub", name="Notification Hub")
func newNotificationHubResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &notificationHubResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type notificationHubResource struct {
	framework.ResourceWithModel[notificationHubResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *notificationHubResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"notification_hub_region": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *notificationHubResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data notificationHubResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	region := fwflex.StringValueFromFramework(ctx, data.NotificationHubRegion)
	input := notifications.RegisterNotificationHubInput{
		NotificationHubRegion: aws.String(region),
	}
	_, err := conn.RegisterNotificationHub(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("registering User Notifications Notification Hub (%s)", region), err.Error())

		return
	}

	if _, err := waitNotificationHubCreated(ctx, conn, region, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Notification Hub (%s) create", region), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *notificationHubResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data notificationHubResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	_, err := findNotificationHubByRegion(ctx, conn, data.NotificationHubRegion.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Notification Hub (%s)", data.NotificationHubRegion.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *notificationHubResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data notificationHubResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	region := fwflex.StringValueFromFramework(ctx, data.NotificationHubRegion)
	input := notifications.DeregisterNotificationHubInput{
		NotificationHubRegion: aws.String(region),
	}
	_, err := conn.DeregisterNotificationHub(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "Cannot deregister last ACTIVE notification hub") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deregistering User Notifications Notification Hub (%s)", region), err.Error())

		return
	}

	if _, err := waitNotificationHubDeleted(ctx, conn, region, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Notification Hub (%s) delete", region), err.Error())

		return
	}
}

func (r *notificationHubResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("notification_hub_region"), request, response)
}

func findNotificationHubByRegion(ctx context.Context, conn *notifications.Client, region string) (*awstypes.NotificationHubOverview, error) {
	var input notifications.ListNotificationHubsInput
	output, err := findNotificationHub(ctx, conn, &input, func(v *awstypes.NotificationHubOverview) bool {
		return aws.ToString(v.NotificationHubRegion) == region
	})

	if err != nil {
		return nil, err
	}

	if output.StatusSummary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findNotificationHub(ctx context.Context, conn *notifications.Client, input *notifications.ListNotificationHubsInput, filter tfslices.Predicate[*awstypes.NotificationHubOverview]) (*awstypes.NotificationHubOverview, error) {
	output, err := findNotificationHubs(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNotificationHubs(ctx context.Context, conn *notifications.Client, input *notifications.ListNotificationHubsInput, filter tfslices.Predicate[*awstypes.NotificationHubOverview]) ([]awstypes.NotificationHubOverview, error) {
	var output []awstypes.NotificationHubOverview

	pages := notifications.NewListNotificationHubsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.NotificationHubs {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusNotificationHub(ctx context.Context, conn *notifications.Client, region string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findNotificationHubByRegion(ctx, conn, region)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.StatusSummary.Status), nil
	}
}

func waitNotificationHubCreated(ctx context.Context, conn *notifications.Client, region string, timeout time.Duration) (*awstypes.NotificationHubOverview, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.NotificationHubStatusRegistering),
		Target:                    enum.Slice(awstypes.NotificationHubStatusActive),
		Refresh:                   statusNotificationHub(ctx, conn, region),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NotificationHubOverview); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusSummary.Reason)))

		return output, err
	}

	return nil, err
}

func waitNotificationHubDeleted(ctx context.Context, conn *notifications.Client, region string, timeout time.Duration) (*awstypes.NotificationHubOverview, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NotificationHubStatusDeregistering),
		Target:  []string{},
		Refresh: statusNotificationHub(ctx, conn, region),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NotificationHubOverview); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusSummary.Reason)))

		return output, err
	}

	return nil, err
}

type notificationHubResourceModel struct {
	NotificationHubRegion types.String   `tfsdk:"notification_hub_region"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}
