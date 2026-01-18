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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_notifications_organizations_access", name="Organizations Access")
func newOrganizationsAccessResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &organizationsAccessResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
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
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
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

	if fwflex.BoolValueFromFramework(ctx, data.Enabled) {
		if err := enableOrganizationsAccess(ctx, conn, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError("creating Notifications Organizations Access", err.Error())
			return
		}
	} else {
		if err := disableOrganizationsAccess(ctx, conn, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError("creating Notifications Organizations Access", err.Error())
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

	output, err := findAccessForOrganization(ctx, conn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading User Notifications Organizations Access", err.Error())
		return
	}

	data.Enabled = types.BoolValue(output.AccessStatus == awstypes.AccessStatusEnabled)

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

	if fwflex.BoolValueFromFramework(ctx, new.Enabled) {
		if err := enableOrganizationsAccess(ctx, conn, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError("updating Notifications Organizations Access", err.Error())
			return
		}
	} else {
		if err := disableOrganizationsAccess(ctx, conn, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError("updating Notifications Organizations Access", err.Error())
			return
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

	if err := disableOrganizationsAccess(ctx, conn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError("deleting Notifications Organizations Access", err.Error())
		return
	}
}

func enableOrganizationsAccess(ctx context.Context, conn *notifications.Client, timeout time.Duration) error {
	var input notifications.EnableNotificationsAccessForOrganizationInput
	_, err := conn.EnableNotificationsAccessForOrganization(ctx, &input)
	if err != nil {
		return fmt.Errorf("enabling User Notifications Organizations Access: %w", err)
	}

	if _, err := waitOrganizationsAccessEnabled(ctx, conn, timeout); err != nil {
		return fmt.Errorf("waiting for User Notifications Organizations Access enable: %w", err)
	}

	return nil
}

func disableOrganizationsAccess(ctx context.Context, conn *notifications.Client, timeout time.Duration) error {
	var input notifications.DisableNotificationsAccessForOrganizationInput
	_, err := conn.DisableNotificationsAccessForOrganization(ctx, &input)
	if err != nil {
		return fmt.Errorf("disabling User Notifications Organizations Access: %w", err)
	}

	if _, err := waitOrganizationsAccessDisabled(ctx, conn, timeout); err != nil {
		return fmt.Errorf("waiting for User Notifications Organizations Access disable: %w", err)
	}

	return nil
}

func waitOrganizationsAccessEnabled(ctx context.Context, conn *notifications.Client, timeout time.Duration) (*awstypes.NotificationsAccessForOrganization, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusDisabled, awstypes.AccessStatusPending),
		Target:  enum.Slice(awstypes.AccessStatusEnabled),
		Refresh: statusOrganizationsAccess(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NotificationsAccessForOrganization); ok {
		return output, err
	}

	return nil, err
}

func waitOrganizationsAccessDisabled(ctx context.Context, conn *notifications.Client, timeout time.Duration) (*awstypes.NotificationsAccessForOrganization, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccessStatusEnabled, awstypes.AccessStatusPending),
		Target:  enum.Slice(awstypes.AccessStatusDisabled),
		Refresh: statusOrganizationsAccess(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NotificationsAccessForOrganization); ok {
		return output, err
	}

	return nil, err
}

func statusOrganizationsAccess(conn *notifications.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAccessForOrganization(ctx, conn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AccessStatus), nil
	}
}

func findAccessForOrganization(ctx context.Context, conn *notifications.Client) (*awstypes.NotificationsAccessForOrganization, error) {
	var input notifications.GetNotificationsAccessForOrganizationInput
	output, err := conn.GetNotificationsAccessForOrganization(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.NotificationsAccessForOrganization == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.NotificationsAccessForOrganization, nil
}

type organizationsAccessResourceModel struct {
	Enabled  types.Bool     `tfsdk:"enabled"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
