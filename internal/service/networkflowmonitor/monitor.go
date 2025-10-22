// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkflowmonitor/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkflowmonitor_monitor", name="Monitor")
// @Tags(identifierAttribute="arn")
func newMonitorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &monitorResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type monitorResource struct {
	framework.ResourceWithConfigure
	framework.ResourceWithModel[monitorResourceModel]
	framework.WithTimeouts
}

func (r *monitorResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_networkflowmonitor_monitor"
}

func (r *monitorResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"monitor_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"monitor_status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"local_resources": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[monitorResourceConfigModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
						"identifier": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"remote_resources": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[monitorResourceConfigModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
						"identifier": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

type monitorResourceModel struct {
	ARN             types.String                                               `tfsdk:"arn"`
	ID              types.String                                               `tfsdk:"id"`
	MonitorName     types.String                                               `tfsdk:"monitor_name"`
	ScopeArn        types.String                                               `tfsdk:"scope_arn"`
	MonitorStatus   types.String                                               `tfsdk:"monitor_status"`
	LocalResources  fwtypes.SetNestedObjectValueOf[monitorResourceConfigModel] `tfsdk:"local_resources"`
	RemoteResources fwtypes.SetNestedObjectValueOf[monitorResourceConfigModel] `tfsdk:"remote_resources"`
	Tags            tftags.Map                                                 `tfsdk:"tags"`
	TagsAll         tftags.Map                                                 `tfsdk:"tags_all"`
	Timeouts        timeouts.Value                                             `tfsdk:"timeouts"`
}

type monitorResourceConfigModel struct {
	Type       types.String `tfsdk:"type"`
	Identifier types.String `tfsdk:"identifier"`
}

func (r *monitorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data monitorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	input := &networkflowmonitor.CreateMonitorInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set additional fields that need special handling
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateMonitor(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Network Flow Monitor Monitor", err.Error())
		return
	}

	data.ID = types.StringValue(aws.ToString(output.MonitorArn))
	data.ARN = types.StringValue(aws.ToString(output.MonitorArn))

	monitor, err := waitMonitorCreated(ctx, conn, data.MonitorName.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Monitor (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	// Use fwflex.Flatten to set all attributes including tags_all
	response.Diagnostics.Append(fwflex.Flatten(ctx, monitor, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *monitorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	monitor, err := findMonitorByName(ctx, conn, data.MonitorName.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Network Flow Monitor Monitor (%s)", data.ID.ValueString()), err.Error())
		return
	}

	// Use fwflex.Flatten to automatically map API response to model
	response.Diagnostics.Append(fwflex.Flatten(ctx, monitor, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Ensure ID and ARN are set properly (especially important for import)
	if data.ID.IsNull() || data.ID.IsUnknown() {
		data.ID = types.StringValue(aws.ToString(monitor.MonitorArn))
	}
	if data.ARN.IsNull() || data.ARN.IsUnknown() {
		data.ARN = types.StringValue(aws.ToString(monitor.MonitorArn))
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *monitorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	// Check if local_resources or remote_resources have changed
	if !new.LocalResources.Equal(old.LocalResources) || !new.RemoteResources.Equal(old.RemoteResources) {
		input := &networkflowmonitor.UpdateMonitorInput{
			MonitorName: aws.String(new.MonitorName.ValueString()),
		}

		// Calculate local resources diff
		oldLocalResources := make(map[string]awstypes.MonitorLocalResource)
		newLocalResources := make(map[string]awstypes.MonitorLocalResource)

		// Build map of old local resources
		if !old.LocalResources.IsNull() && !old.LocalResources.IsUnknown() {
			oldLocalSlice, diags := old.LocalResources.ToSlice(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			for _, resource := range oldLocalSlice {
				key := resource.Type.ValueString() + "|" + resource.Identifier.ValueString()
				oldLocalResources[key] = awstypes.MonitorLocalResource{
					Type:       awstypes.MonitorLocalResourceType(resource.Type.ValueString()),
					Identifier: aws.String(resource.Identifier.ValueString()),
				}
			}
		}

		// Build map of new local resources
		if !new.LocalResources.IsNull() && !new.LocalResources.IsUnknown() {
			newLocalSlice, diags := new.LocalResources.ToSlice(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			for _, resource := range newLocalSlice {
				key := resource.Type.ValueString() + "|" + resource.Identifier.ValueString()
				newLocalResources[key] = awstypes.MonitorLocalResource{
					Type:       awstypes.MonitorLocalResourceType(resource.Type.ValueString()),
					Identifier: aws.String(resource.Identifier.ValueString()),
				}
			}
		}

		// Find resources to add (in new but not in old)
		var localResourcesToAdd []awstypes.MonitorLocalResource
		for key, resource := range newLocalResources {
			if _, exists := oldLocalResources[key]; !exists {
				localResourcesToAdd = append(localResourcesToAdd, resource)
			}
		}

		// Find resources to remove (in old but not in new)
		var localResourcesToRemove []awstypes.MonitorLocalResource
		for key, resource := range oldLocalResources {
			if _, exists := newLocalResources[key]; !exists {
				localResourcesToRemove = append(localResourcesToRemove, resource)
			}
		}

		// Calculate remote resources diff
		oldRemoteResources := make(map[string]awstypes.MonitorRemoteResource)
		newRemoteResources := make(map[string]awstypes.MonitorRemoteResource)

		// Build map of old remote resources
		if !old.RemoteResources.IsNull() && !old.RemoteResources.IsUnknown() {
			oldRemoteSlice, diags := old.RemoteResources.ToSlice(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			for _, resource := range oldRemoteSlice {
				key := resource.Type.ValueString() + "|" + resource.Identifier.ValueString()
				oldRemoteResources[key] = awstypes.MonitorRemoteResource{
					Type:       awstypes.MonitorRemoteResourceType(resource.Type.ValueString()),
					Identifier: aws.String(resource.Identifier.ValueString()),
				}
			}
		}

		// Build map of new remote resources
		if !new.RemoteResources.IsNull() && !new.RemoteResources.IsUnknown() {
			newRemoteSlice, diags := new.RemoteResources.ToSlice(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			for _, resource := range newRemoteSlice {
				key := resource.Type.ValueString() + "|" + resource.Identifier.ValueString()
				newRemoteResources[key] = awstypes.MonitorRemoteResource{
					Type:       awstypes.MonitorRemoteResourceType(resource.Type.ValueString()),
					Identifier: aws.String(resource.Identifier.ValueString()),
				}
			}
		}

		// Find resources to add (in new but not in old)
		var remoteResourcesToAdd []awstypes.MonitorRemoteResource
		for key, resource := range newRemoteResources {
			if _, exists := oldRemoteResources[key]; !exists {
				remoteResourcesToAdd = append(remoteResourcesToAdd, resource)
			}
		}

		// Find resources to remove (in old but not in new)
		var remoteResourcesToRemove []awstypes.MonitorRemoteResource
		for key, resource := range oldRemoteResources {
			if _, exists := newRemoteResources[key]; !exists {
				remoteResourcesToRemove = append(remoteResourcesToRemove, resource)
			}
		}

		// Set the calculated diffs in the input
		if len(localResourcesToAdd) > 0 {
			input.LocalResourcesToAdd = localResourcesToAdd
		}
		if len(localResourcesToRemove) > 0 {
			input.LocalResourcesToRemove = localResourcesToRemove
		}
		if len(remoteResourcesToAdd) > 0 {
			input.RemoteResourcesToAdd = remoteResourcesToAdd
		}
		if len(remoteResourcesToRemove) > 0 {
			input.RemoteResourcesToRemove = remoteResourcesToRemove
		}

		_, err := conn.UpdateMonitor(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Network Flow Monitor Monitor (%s)", new.ID.ValueString()), err.Error())
			return
		}

		// Wait for the update to complete
		monitor, err := waitMonitorUpdated(ctx, conn, new.MonitorName.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Monitor (%s) update", new.ID.ValueString()), err.Error())
			return
		}

		// Update only the computed attributes, preserve the order of resources from configuration
		new.MonitorStatus = types.StringValue(string(monitor.MonitorStatus))

		// Ensure ID and ARN are set properly
		if new.ID.IsNull() || new.ID.IsUnknown() {
			new.ID = types.StringValue(aws.ToString(monitor.MonitorArn))
		}
		if new.ARN.IsNull() || new.ARN.IsUnknown() {
			new.ARN = types.StringValue(aws.ToString(monitor.MonitorArn))
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *monitorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	_, err := conn.DeleteMonitor(ctx, &networkflowmonitor.DeleteMonitorInput{
		MonitorName: aws.String(data.MonitorName.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Network Flow Monitor Monitor (%s)", data.ID.ValueString()), err.Error())
		return
	}

	if _, err := waitMonitorDeleted(ctx, conn, data.MonitorName.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Monitor (%s) delete", data.ID.ValueString()), err.Error())
		return
	}
}

func (r *monitorResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// The import ID can be either an ARN or a monitor name
	id := request.ID

	// If it's an ARN, extract the monitor name
	if strings.HasPrefix(id, "arn:aws:networkflowmonitor:") {
		// ARN format: arn:aws:networkflowmonitor:region:account:monitor/monitor-name
		parts := strings.Split(id, "/")
		if len(parts) != 2 {
			response.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected ARN format 'arn:aws:networkflowmonitor:region:account:monitor/monitor-name', got: %s", id))
			return
		}
		monitorName := parts[1]

		// Set both ID (ARN) and monitor_name
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), id)...)
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("monitor_name"), monitorName)...)
	} else {
		// Assume it's a monitor name, we'll need to construct the ARN during read
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("monitor_name"), id)...)
		// ID will be set during the subsequent Read operation
	}
}

func findMonitorByName(ctx context.Context, conn *networkflowmonitor.Client, name string) (*networkflowmonitor.GetMonitorOutput, error) {
	input := &networkflowmonitor.GetMonitorInput{
		MonitorName: aws.String(name),
	}

	output, err := conn.GetMonitor(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusMonitor(ctx context.Context, conn *networkflowmonitor.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findMonitorByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.MonitorStatus), nil
	}
}

func waitMonitorCreated(ctx context.Context, conn *networkflowmonitor.Client, name string, timeout time.Duration) (*networkflowmonitor.GetMonitorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MonitorStatusPending),
		Target:  enum.Slice(awstypes.MonitorStatusActive),
		Refresh: statusMonitor(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkflowmonitor.GetMonitorOutput); ok {
		return output, err
	}

	return nil, err
}

func waitMonitorUpdated(ctx context.Context, conn *networkflowmonitor.Client, name string, timeout time.Duration) (*networkflowmonitor.GetMonitorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MonitorStatusPending),
		Target:  enum.Slice(awstypes.MonitorStatusActive),
		Refresh: statusMonitor(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkflowmonitor.GetMonitorOutput); ok {
		return output, err
	}

	return nil, err
}

func waitMonitorDeleted(ctx context.Context, conn *networkflowmonitor.Client, name string, timeout time.Duration) (*networkflowmonitor.GetMonitorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MonitorStatusDeleting),
		Target:  []string{},
		Refresh: statusMonitor(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkflowmonitor.GetMonitorOutput); ok {
		return output, err
	}

	return nil, err
}
