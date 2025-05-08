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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_notifications_notification_hub", name="Notification Hub")
func newResourceNotificationHub(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceNotificationHub{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

const (
	ResNameNotificationHub = "Notification Hub"
)

type resourceNotificationHub struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *resourceNotificationHub) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrRegion: schema.StringAttribute{
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

func (r *resourceNotificationHub) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var plan resourceNotificationHubModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input notifications.RegisterNotificationHubInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("NotificationHub"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.RegisterNotificationHub(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionCreating, ResNameNotificationHub, plan.Region.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionCreating, ResNameNotificationHub, plan.Region.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitNotificationHubCreated(ctx, conn, plan.Region.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionWaitingForCreation, ResNameNotificationHub, plan.Region.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNotificationHub) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var state resourceNotificationHubModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNotificationHubByRegion(ctx, conn, state.Region.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionReading, ResNameNotificationHub, state.Region.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNotificationHub) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var state resourceNotificationHubModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := notifications.DeregisterNotificationHubInput{
		NotificationHubRegion: state.Region.ValueStringPointer(),
	}

	_, err := conn.DeregisterNotificationHub(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionDeleting, ResNameNotificationHub, state.Region.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitNotificationHubDeleted(ctx, conn, state.Region.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionWaitingForDeletion, ResNameNotificationHub, state.Region.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceNotificationHub) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrRegion), req, resp)
}

func waitNotificationHubCreated(ctx context.Context, conn *notifications.Client, id string, timeout time.Duration) (*awstypes.NotificationHubOverview, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.NotificationHubStatusRegistering),
		Target:                    enum.Slice(awstypes.NotificationHubStatusActive),
		Refresh:                   statusNotificationHub(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.NotificationHubOverview); ok {
		return out, err
	}

	return nil, err
}

func waitNotificationHubDeleted(ctx context.Context, conn *notifications.Client, id string, timeout time.Duration) (*awstypes.NotificationHubOverview, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NotificationHubStatusDeregistering),
		Target:  []string{},
		Refresh: statusNotificationHub(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.NotificationHubOverview); ok {
		return out, err
	}

	return nil, err
}

func statusNotificationHub(ctx context.Context, conn *notifications.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findNotificationHubByRegion(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.StatusSummary.Status), nil
	}
}

func findNotificationHubByRegion(ctx context.Context, conn *notifications.Client, region string) (*awstypes.NotificationHubOverview, error) {
	var input notifications.ListNotificationHubsInput

	out, err := conn.ListNotificationHubs(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}

	if out == nil || out.NotificationHubs == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	for _, hub := range out.NotificationHubs {
		if aws.ToString(hub.NotificationHubRegion) == region {
			return &hub, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError:   fmt.Errorf("notification Hub for region %q not found", region),
		LastRequest: &input,
	}
}

type resourceNotificationHubModel struct {
	Region   types.String   `tfsdk:"region"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
