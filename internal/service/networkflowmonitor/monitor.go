// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkflowmonitor/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	framework.WithImportByID
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
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"local_resources": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[monitorResourceConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
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
			"remote_resources": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[monitorResourceConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

type monitorResourceModel struct {
	ARN             types.String                                                `tfsdk:"arn"`
	ID              types.String                                                `tfsdk:"id"`
	MonitorName     types.String                                                `tfsdk:"monitor_name"`
	ScopeArn        types.String                                                `tfsdk:"scope_arn"`
	MonitorStatus   types.String                                                `tfsdk:"monitor_status"`
	LocalResources  fwtypes.ListNestedObjectValueOf[monitorResourceConfigModel] `tfsdk:"local_resources"`
	RemoteResources fwtypes.ListNestedObjectValueOf[monitorResourceConfigModel] `tfsdk:"remote_resources"`
	Tags            tftags.Map                                                  `tfsdk:"tags"`
	TagsAll         tftags.Map                                                  `tfsdk:"tags_all"`
	Timeouts        timeouts.Value                                              `tfsdk:"timeouts"`
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

	// Set computed attributes
	data.MonitorStatus = types.StringValue(string(monitor.MonitorStatus))

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

	// Use flex.Flatten to automatically map API response to model
	response.Diagnostics.Append(fwflex.Flatten(ctx, monitor, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set tags from API response
	setTagsOut(ctx, monitor.Tags)

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

	if !new.Tags.Equal(old.Tags) {
		if err := updateTags(ctx, conn, new.ID.ValueString(), old.Tags, new.Tags); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Network Flow Monitor Monitor (%s) tags", new.ID.ValueString()), err.Error())
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
