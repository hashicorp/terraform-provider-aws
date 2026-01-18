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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_notifications_organizations_access", name="Organizations Access")
func newOrganizationsAccessResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &organizationsAccessResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultReadTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type organizationsAccessResource struct {
	framework.ResourceWithModel[organizationsAccessResourceModel]
	framework.WithTimeouts
}

func (r *organizationsAccessResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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

func (r *organizationsAccessResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data organizationsAccessResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	// Set ID to account ID
	data.ID = types.StringValue(r.Meta().AccountID(ctx))

	enabled := fwflex.BoolValueFromFramework(ctx, data.Enabled)
	if enabled {
		input := notifications.EnableNotificationsAccessForOrganizationInput{}
		_, err := conn.EnableNotificationsAccessForOrganization(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError("enabling User Notifications Organizations Access", err.Error())
			return
		}

		// Wait for enabled state
		_, err = waitOrganizationsAccessEnabled(ctx, conn, r.CreateTimeout(ctx, data.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Organizations Access (%s) to be enabled", data.ID.ValueString()), err.Error())
			return
		}
	} else {
		input := notifications.DisableNotificationsAccessForOrganizationInput{}
		_, err := conn.DisableNotificationsAccessForOrganization(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError("disabling User Notifications Organizations Access", err.Error())
			return
		}

		// Wait for disabled state
		_, err = waitOrganizationsAccessDisabled(ctx, conn, r.CreateTimeout(ctx, data.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Organizations Access (%s) to be disabled", data.ID.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *organizationsAccessResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data organizationsAccessResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	status, err := waitOrganizationsAccessStable(ctx, conn, r.ReadTimeout(ctx, data.Timeouts))

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Organizations Access (%s)", data.ID.ValueString()), err.Error())
		return
	}

	if status == "" {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Organizations Access (%s)", data.ID.ValueString()), "empty response")
		return
	}

	data.Enabled = types.BoolValue(status == string(awstypes.AccessStatusEnabled))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *organizationsAccessResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new organizationsAccessResourceModel
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
			input := notifications.EnableNotificationsAccessForOrganizationInput{}
			_, err := conn.EnableNotificationsAccessForOrganization(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError("enabling User Notifications Organizations Access", err.Error())
				return
			}

			// Wait for enabled state
			_, err = waitOrganizationsAccessEnabled(ctx, conn, r.UpdateTimeout(ctx, new.Timeouts))
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Organizations Access (%s) to be enabled", new.ID.ValueString()), err.Error())
				return
			}
		} else {
			input := notifications.DisableNotificationsAccessForOrganizationInput{}
			_, err := conn.DisableNotificationsAccessForOrganization(ctx, &input)

			if err != nil {
				response.Diagnostics.AddError("disabling User Notifications Organizations Access", err.Error())
				return
			}

			// Wait for disabled state
			_, err = waitOrganizationsAccessDisabled(ctx, conn, r.UpdateTimeout(ctx, new.Timeouts))
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Organizations Access (%s) to be disabled", new.ID.ValueString()), err.Error())
				return
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *organizationsAccessResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data organizationsAccessResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	// Always disable on delete
	input := notifications.DisableNotificationsAccessForOrganizationInput{}
	_, err := conn.DisableNotificationsAccessForOrganization(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("disabling User Notifications Organizations Access", err.Error())
		return
	}

	// Wait for disabled state
	_, err = waitOrganizationsAccessDisabled(ctx, conn, r.DeleteTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for User Notifications Organizations Access (%s) to be disabled", data.ID.ValueString()), err.Error())
		return
	}
}

const (
	organizationsAccessStableTimeout = 10 * time.Minute
)

func waitOrganizationsAccessEnabled(ctx context.Context, conn *notifications.Client, timeout time.Duration) (*notifications.GetNotificationsAccessForOrganizationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusDisabled, awstypes.AccessStatusPending),
		Target:  enum.Slice(awstypes.AccessStatusEnabled),
		Refresh: statusOrganizationsAccess(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetNotificationsAccessForOrganizationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitOrganizationsAccessDisabled(ctx context.Context, conn *notifications.Client, timeout time.Duration) (*notifications.GetNotificationsAccessForOrganizationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusEnabled, awstypes.AccessStatusPending),
		Target:  enum.Slice(awstypes.AccessStatusDisabled),
		Refresh: statusOrganizationsAccess(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*notifications.GetNotificationsAccessForOrganizationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitOrganizationsAccessStable(ctx context.Context, conn *notifications.Client, timeout time.Duration) (string, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusPending),
		Target:  enum.Slice(awstypes.AccessStatusEnabled, awstypes.AccessStatusDisabled),
		Refresh: statusOrganizationsAccess(ctx, conn),
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

func statusOrganizationsAccess(ctx context.Context, conn *notifications.Client) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := getOrganizationsAccess(ctx, conn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.NotificationsAccessForOrganization.AccessStatus), nil
	}
}

func getOrganizationsAccess(ctx context.Context, conn *notifications.Client) (*notifications.GetNotificationsAccessForOrganizationOutput, error) {
	input := notifications.GetNotificationsAccessForOrganizationInput{}

	output, err := conn.GetNotificationsAccessForOrganization(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.NotificationsAccessForOrganization == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type organizationsAccessResourceModel struct {
	Enabled  types.Bool     `tfsdk:"enabled"`
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
