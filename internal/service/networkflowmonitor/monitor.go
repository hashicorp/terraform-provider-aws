// Copyright IBM Corp. 2014, 2026
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
	set "github.com/hashicorp/go-set/v3"
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
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

	if retry.NotFound(err) {
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
		var oldLocalResources, newLocalResources []awstypes.MonitorLocalResource
		response.Diagnostics.Append(fwflex.Expand(ctx, old.LocalResources, &oldLocalResources)...)
		if response.Diagnostics.HasError() {
			return
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new.LocalResources, &newLocalResources)...)
		if response.Diagnostics.HasError() {
			return
		}

		hashLocalResource := func(v awstypes.MonitorLocalResource) string {
			return string(v.Type) + ":" + aws.ToString(v.Identifier)
		}
		osLocalResource, nsLocalResource := set.HashSetFromFunc(oldLocalResources, hashLocalResource), set.HashSetFromFunc(newLocalResources, hashLocalResource)

		var oldRemoteResources, newRemoteResources []awstypes.MonitorRemoteResource
		response.Diagnostics.Append(fwflex.Expand(ctx, old.RemoteResources, &oldRemoteResources)...)
		if response.Diagnostics.HasError() {
			return
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new.RemoteResources, &newRemoteResources)...)
		if response.Diagnostics.HasError() {
			return
		}

		hashRemoteResource := func(v awstypes.MonitorRemoteResource) string {
			return string(v.Type) + ":" + aws.ToString(v.Identifier)
		}
		osRemoteResource, nsRemoteResource := set.HashSetFromFunc(oldRemoteResources, hashRemoteResource), set.HashSetFromFunc(newRemoteResources, hashRemoteResource)

		monitorName := fwflex.StringValueFromFramework(ctx, new.MonitorName)
		input := networkflowmonitor.UpdateMonitorInput{
			LocalResourcesToAdd:     nsLocalResource.Difference(osLocalResource).Slice(),
			LocalResourcesToRemove:  osLocalResource.Difference(nsLocalResource).Slice(),
			MonitorName:             aws.String(monitorName),
			RemoteResourcesToAdd:    nsRemoteResource.Difference(osRemoteResource).Slice(),
			RemoteResourcesToRemove: osRemoteResource.Difference(nsRemoteResource).Slice(),
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
	resource.ImportStatePassthroughID(ctx, path.Root("monitor_name"), request, response)
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
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusMonitor(ctx context.Context, conn *networkflowmonitor.Client, name string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findMonitorByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.MonitorStatus), nil
	}
}

func waitMonitorCreated(ctx context.Context, conn *networkflowmonitor.Client, name string, timeout time.Duration) (*networkflowmonitor.GetMonitorOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
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
	stateConf := &sdkretry.StateChangeConf{
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
	stateConf := &sdkretry.StateChangeConf{
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
