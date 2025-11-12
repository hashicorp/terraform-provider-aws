// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkflowmonitor/types"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
// @Tags(identifierAttribute="monitor_arn")
func newMonitorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &monitorResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type monitorResource struct {
	framework.ResourceWithModel[monitorResourceModel]
	framework.WithTimeouts
}

func (r *monitorResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"monitor_arn": framework.ARNAttributeComputedOnly(),
			"monitor_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9_.-]+`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"local_resource": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[monitorLocalResourceModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrIdentifier: schema.StringAttribute{
							Required: true,
						},
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MonitorLocalResourceType](),
							Required:   true,
						},
					},
				},
			},
			"remote_resource": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[monitorRemoteResourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrIdentifier: schema.StringAttribute{
							Required: true,
						},
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MonitorRemoteResourceType](),
							Required:   true,
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

func (r *monitorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data monitorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	var input networkflowmonitor.CreateMonitorInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	uuid, _ := uuid.GenerateUUID()
	input.ClientToken = aws.String(uuid)
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateMonitor(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Network Flow Monitor Monitor", err.Error())
		return
	}

	// Set values for unknowns.
	data.MonitorARN = fwflex.StringToFramework(ctx, output.MonitorArn)

	monitorName := fwflex.StringValueFromFramework(ctx, data.MonitorName)
	if _, err := waitMonitorCreated(ctx, conn, monitorName, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Monitor (%s) create", monitorName), err.Error())
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

	monitorName := fwflex.StringValueFromFramework(ctx, data.MonitorName)
	output, err := findMonitorByName(ctx, conn, monitorName)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Network Flow Monitor Monitor (%s)", monitorName), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
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

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		monitorName := fwflex.StringValueFromFramework(ctx, new.MonitorName)
		input := networkflowmonitor.UpdateMonitorInput{
			MonitorName: aws.String(monitorName),
		}

		_, err := conn.UpdateMonitor(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Network Flow Monitor Monitor (%s)", monitorName), err.Error())
			return
		}

		if _, err := waitMonitorUpdated(ctx, conn, monitorName, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Monitor (%s) update", monitorName), err.Error())
			return
		}
	}

	// Check if local_resources or remote_resources have changed
	// if !new.LocalResources.Equal(old.LocalResources) || !new.RemoteResources.Equal(old.RemoteResources) {
	// 	input := networkflowmonitor.UpdateMonitorInput{
	// 		MonitorName: new.MonitorName.ValueStringPointer(),
	// 	}

	// 	// Calculate local resources diff
	// 	oldLocalResources := make(map[string]awstypes.MonitorLocalResource)
	// 	newLocalResources := make(map[string]awstypes.MonitorLocalResource)

	// 	// Build map of old local resources
	// 	if !old.LocalResources.IsNull() && !old.LocalResources.IsUnknown() {
	// 		oldLocalSlice, diags := old.LocalResources.ToSlice(ctx)
	// 		response.Diagnostics.Append(diags...)
	// 		if response.Diagnostics.HasError() {
	// 			return
	// 		}
	// 		for _, resource := range oldLocalSlice {
	// 			key := resource.Type.ValueString() + "|" + resource.Identifier.ValueString()
	// 			oldLocalResources[key] = awstypes.MonitorLocalResource{
	// 				Type:       awstypes.MonitorLocalResourceType(resource.Type.ValueString()),
	// 				Identifier: resource.Identifier.ValueStringPointer(),
	// 			}
	// 		}
	// 	}

	// 	// Build map of new local resources
	// 	if !new.LocalResources.IsNull() && !new.LocalResources.IsUnknown() {
	// 		newLocalSlice, diags := new.LocalResources.ToSlice(ctx)
	// 		response.Diagnostics.Append(diags...)
	// 		if response.Diagnostics.HasError() {
	// 			return
	// 		}
	// 		for _, resource := range newLocalSlice {
	// 			key := resource.Type.ValueString() + "|" + resource.Identifier.ValueString()
	// 			newLocalResources[key] = awstypes.MonitorLocalResource{
	// 				Type:       awstypes.MonitorLocalResourceType(resource.Type.ValueString()),
	// 				Identifier: resource.Identifier.ValueStringPointer(),
	// 			}
	// 		}
	// 	}

	// 	// Find resources to add (in new but not in old)
	// 	var localResourcesToAdd []awstypes.MonitorLocalResource
	// 	for key, resource := range newLocalResources {
	// 		if _, exists := oldLocalResources[key]; !exists {
	// 			localResourcesToAdd = append(localResourcesToAdd, resource)
	// 		}
	// 	}

	// 	// Find resources to remove (in old but not in new)
	// 	var localResourcesToRemove []awstypes.MonitorLocalResource
	// 	for key, resource := range oldLocalResources {
	// 		if _, exists := newLocalResources[key]; !exists {
	// 			localResourcesToRemove = append(localResourcesToRemove, resource)
	// 		}
	// 	}

	// 	// Calculate remote resources diff
	// 	oldRemoteResources := make(map[string]awstypes.MonitorRemoteResource)
	// 	newRemoteResources := make(map[string]awstypes.MonitorRemoteResource)

	// 	// Build map of old remote resources
	// 	if !old.RemoteResources.IsNull() && !old.RemoteResources.IsUnknown() {
	// 		oldRemoteSlice, diags := old.RemoteResources.ToSlice(ctx)
	// 		response.Diagnostics.Append(diags...)
	// 		if response.Diagnostics.HasError() {
	// 			return
	// 		}
	// 		for _, resource := range oldRemoteSlice {
	// 			key := resource.Type.ValueString() + "|" + resource.Identifier.ValueString()
	// 			oldRemoteResources[key] = awstypes.MonitorRemoteResource{
	// 				Type:       awstypes.MonitorRemoteResourceType(resource.Type.ValueString()),
	// 				Identifier: resource.Identifier.ValueStringPointer(),
	// 			}
	// 		}
	// 	}

	// 	// Build map of new remote resources
	// 	if !new.RemoteResources.IsNull() && !new.RemoteResources.IsUnknown() {
	// 		newRemoteSlice, diags := new.RemoteResources.ToSlice(ctx)
	// 		response.Diagnostics.Append(diags...)
	// 		if response.Diagnostics.HasError() {
	// 			return
	// 		}
	// 		for _, resource := range newRemoteSlice {
	// 			key := resource.Type.ValueString() + "|" + resource.Identifier.ValueString()
	// 			newRemoteResources[key] = awstypes.MonitorRemoteResource{
	// 				Type:       awstypes.MonitorRemoteResourceType(resource.Type.ValueString()),
	// 				Identifier: resource.Identifier.ValueStringPointer(),
	// 			}
	// 		}
	// 	}

	// 	// Find resources to add (in new but not in old)
	// 	var remoteResourcesToAdd []awstypes.MonitorRemoteResource
	// 	for key, resource := range newRemoteResources {
	// 		if _, exists := oldRemoteResources[key]; !exists {
	// 			remoteResourcesToAdd = append(remoteResourcesToAdd, resource)
	// 		}
	// 	}

	// 	// Find resources to remove (in old but not in new)
	// 	var remoteResourcesToRemove []awstypes.MonitorRemoteResource
	// 	for key, resource := range oldRemoteResources {
	// 		if _, exists := newRemoteResources[key]; !exists {
	// 			remoteResourcesToRemove = append(remoteResourcesToRemove, resource)
	// 		}
	// 	}

	// 	// Set the calculated diffs in the input
	// 	if len(localResourcesToAdd) > 0 {
	// 		input.LocalResourcesToAdd = localResourcesToAdd
	// 	}
	// 	if len(localResourcesToRemove) > 0 {
	// 		input.LocalResourcesToRemove = localResourcesToRemove
	// 	}
	// 	if len(remoteResourcesToAdd) > 0 {
	// 		input.RemoteResourcesToAdd = remoteResourcesToAdd
	// 	}
	// 	if len(remoteResourcesToRemove) > 0 {
	// 		input.RemoteResourcesToRemove = remoteResourcesToRemove
	// 	}

	// 	_, err := conn.UpdateMonitor(ctx, &input)
	// 	if err != nil {
	// 		response.Diagnostics.AddError(fmt.Sprintf("updating Network Flow Monitor Monitor (%s)", new.ID.ValueString()), err.Error())
	// 		return
	// 	}

	// 	// Wait for the update to complete
	// 	monitor, err := waitMonitorUpdated(ctx, conn, new.MonitorName.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
	// 	if err != nil {
	// 		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Monitor (%s) update", new.ID.ValueString()), err.Error())
	// 		return
	// 	}

	// 	// Update only the computed attributes, preserve the order of resources from configuration
	// 	new.MonitorStatus = types.StringValue(string(monitor.MonitorStatus))

	// 	// Ensure ID and ARN are set properly
	// 	if new.ID.IsNull() || new.ID.IsUnknown() {
	// 		new.ID = types.StringValue(aws.ToString(monitor.MonitorArn))
	// 	}
	// 	if new.ARN.IsNull() || new.ARN.IsUnknown() {
	// 		new.ARN = types.StringValue(aws.ToString(monitor.MonitorArn))
	// 	}
	// }

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *monitorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	monitorName := fwflex.StringValueFromFramework(ctx, data.MonitorName)
	input := networkflowmonitor.DeleteMonitorInput{
		MonitorName: aws.String(monitorName),
	}
	_, err := conn.DeleteMonitor(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Network Flow Monitor Monitor (%s)", monitorName), err.Error())
		return
	}

	if _, err := waitMonitorDeleted(ctx, conn, monitorName, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Monitor (%s) delete", monitorName), err.Error())
		return
	}
}

func (r *monitorResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("monitor_name"), request.ID)...)
}

func findMonitorByName(ctx context.Context, conn *networkflowmonitor.Client, name string) (*networkflowmonitor.GetMonitorOutput, error) {
	input := networkflowmonitor.GetMonitorInput{
		MonitorName: aws.String(name),
	}

	return findMonitor(ctx, conn, &input)
}

func findMonitor(ctx context.Context, conn *networkflowmonitor.Client, input *networkflowmonitor.GetMonitorInput) (*networkflowmonitor.GetMonitorOutput, error) {
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
	return func() (any, string, error) {
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

type monitorResourceModel struct {
	framework.WithRegionModel
	LocalResources  fwtypes.SetNestedObjectValueOf[monitorLocalResourceModel]  `tfsdk:"local_resource"`
	MonitorARN      types.String                                               `tfsdk:"monitor_arn"`
	MonitorName     types.String                                               `tfsdk:"monitor_name"`
	RemoteResources fwtypes.SetNestedObjectValueOf[monitorRemoteResourceModel] `tfsdk:"remote_resource"`
	ScopeARN        fwtypes.ARN                                                `tfsdk:"scope_arn"`
	Tags            tftags.Map                                                 `tfsdk:"tags"`
	TagsAll         tftags.Map                                                 `tfsdk:"tags_all"`
	Timeouts        timeouts.Value                                             `tfsdk:"timeouts"`
}

type monitorLocalResourceModel struct {
	Identifier types.String                                          `tfsdk:"identifier"`
	Type       fwtypes.StringEnum[awstypes.MonitorLocalResourceType] `tfsdk:"type"`
}

type monitorRemoteResourceModel struct {
	Identifier types.String                                           `tfsdk:"identifier"`
	Type       fwtypes.StringEnum[awstypes.MonitorRemoteResourceType] `tfsdk:"type"`
}
