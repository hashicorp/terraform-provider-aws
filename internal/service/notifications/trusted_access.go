// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_notifications_trusted_access", name="Trusted Access")
func newTrustedAccessResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &trustedAccessResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultReadTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type trustedAccessResource struct {
	framework.ResourceWithModel[trustedAccessResourceModel]
	framework.WithTimeouts
}

func (r *trustedAccessResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrEnabled: schema.BoolAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *trustedAccessResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data trustedAccessResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	// Set ID to account ID
	data.ID = types.StringValue(r.Meta().AccountID(ctx))

	enabled := fwflex.BoolValueFromFramework(ctx, data.Enabled)
	if enabled {
		_, err := conn.EnableNotificationsAccessForOrganization(ctx, &notifications.EnableNotificationsAccessForOrganizationInput{})

		if err != nil {
			response.Diagnostics.AddError("enabling User Notifications Trusted Access", err.Error())
			return
		}

		// Wait for enabled state
		_, err = waitTrustedAccessEnabled(ctx, conn, r.CreateTimeout(ctx, data.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Trusted Access (%s) to be enabled", data.ID.ValueString()), err.Error())
			return
		}
	} else {
		_, err := conn.DisableNotificationsAccessForOrganization(ctx, &notifications.DisableNotificationsAccessForOrganizationInput{})

		if err != nil {
			response.Diagnostics.AddError("disabling User Notifications Trusted Access", err.Error())
			return
		}

		// Wait for disabled state
		_, err = waitTrustedAccessDisabled(ctx, conn, r.CreateTimeout(ctx, data.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Trusted Access (%s) to be disabled", data.ID.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *trustedAccessResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data trustedAccessResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	status, err := waitTrustedAccessStable(ctx, conn, r.ReadTimeout(ctx, data.Timeouts))

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Trusted Access (%s)", data.ID.ValueString()), err.Error())
		return
	}

	if status == "" {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Trusted Access (%s)", data.ID.ValueString()), "empty response")
		return
	}

	data.Enabled = types.BoolValue(status == string(awstypes.AccessStatusEnabled))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *trustedAccessResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new trustedAccessResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	if !new.Enabled.Equal(old.Enabled) {
		enabled := fwflex.BoolValueFromFramework(ctx, new.Enabled)
		if enabled {
			_, err := conn.EnableNotificationsAccessForOrganization(ctx, &notifications.EnableNotificationsAccessForOrganizationInput{})

			if err != nil {
				response.Diagnostics.AddError("enabling User Notifications Trusted Access", err.Error())
				return
			}

			// Wait for enabled state
			_, err = waitTrustedAccessEnabled(ctx, conn, r.UpdateTimeout(ctx, new.Timeouts))
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Trusted Access (%s) to be enabled", new.ID.ValueString()), err.Error())
				return
			}
		} else {
			_, err := conn.DisableNotificationsAccessForOrganization(ctx, &notifications.DisableNotificationsAccessForOrganizationInput{})

			if err != nil {
				response.Diagnostics.AddError("disabling User Notifications Trusted Access", err.Error())
				return
			}

			// Wait for disabled state
			_, err = waitTrustedAccessDisabled(ctx, conn, r.UpdateTimeout(ctx, new.Timeouts))
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Trusted Access (%s) to be disabled", new.ID.ValueString()), err.Error())
				return
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *trustedAccessResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data trustedAccessResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	// Always disable on delete
	_, err := conn.DisableNotificationsAccessForOrganization(ctx, &notifications.DisableNotificationsAccessForOrganizationInput{})

	if err != nil {
		response.Diagnostics.AddError("disabling User Notifications Trusted Access", err.Error())
		return
	}

	// Wait for disabled state
	_, err = waitTrustedAccessDisabled(ctx, conn, r.DeleteTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Trusted Access (%s) to be disabled", data.ID.ValueString()), err.Error())
		return
	}
}

const (
	trustedAccessStableTimeout = 10 * time.Minute
	statusNotFound             = "NotFound"
	statusUnavailable          = "Unavailable"
	trustedAccessStatusError   = "Error"
)

func waitTrustedAccessEnabled(ctx context.Context, conn *notifications.Client, timeout time.Duration) (*notifications.GetNotificationsAccessForOrganizationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusDisabled, awstypes.AccessStatusPending, statusNotFound, statusUnavailable),
		Target:  enum.Slice(awstypes.AccessStatusEnabled),
		Refresh: statusTrustedAccess(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetNotificationsAccessForOrganizationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTrustedAccessDisabled(ctx context.Context, conn *notifications.Client, timeout time.Duration) (*notifications.GetNotificationsAccessForOrganizationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusEnabled, awstypes.AccessStatusPending, statusNotFound, statusUnavailable),
		Target:  enum.Slice(awstypes.AccessStatusDisabled),
		Refresh: statusTrustedAccess(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetNotificationsAccessForOrganizationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTrustedAccessStable(ctx context.Context, conn *notifications.Client, timeout time.Duration) (string, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusPending, statusNotFound, statusUnavailable),
		Target:  enum.Slice(awstypes.AccessStatusEnabled, awstypes.AccessStatusDisabled),
		Refresh: statusTrustedAccess(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetNotificationsAccessForOrganizationOutput); ok {
		if output.NotificationsAccessForOrganization != nil {
			return string(output.NotificationsAccessForOrganization.AccessStatus), err
		}
	}

	return "", err
}

func statusTrustedAccess(ctx context.Context, conn *notifications.Client) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &notifications.GetNotificationsAccessForOrganizationInput{}

		output, err := conn.GetNotificationsAccessForOrganization(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, statusNotFound, err
		}

		if err != nil {
			return nil, trustedAccessStatusError, fmt.Errorf("getting User Notifications Trusted Access: %w", err)
		}

		if output == nil || output.NotificationsAccessForOrganization == nil {
			return nil, statusUnavailable, fmt.Errorf("getting User Notifications Trusted Access: empty response")
		}

		return output, string(output.NotificationsAccessForOrganization.AccessStatus), err
	}
}

type trustedAccessResourceModel struct {
	Enabled  types.Bool     `tfsdk:"enabled"`
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
