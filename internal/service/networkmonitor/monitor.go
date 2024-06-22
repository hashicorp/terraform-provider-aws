// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmonitor

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmonitor/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	MonitorTimeout = time.Minute * 10
	ResNameMonitor = "CloudWatch Network Monitor Monitor"
)

// @FrameworkResource(name="CloudWatch Network Monitor Monitor")
// @Tags(identifierAttribute="arn")
func newResourceMonitor(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceNetworkMonitorMonitor{}, nil
}

type resourceNetworkMonitorMonitor struct {
	framework.ResourceWithConfigure
}

func (r *resourceNetworkMonitorMonitor) Metadata(_ context.Context, request resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_networkmonitor_monitor"
}

func (r *resourceNetworkMonitorMonitor) Schema(ctx context.Context, request resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"aggregation_period": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.OneOf(30, 60),
				},
			},
			names.AttrCreatedAt: schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"monitor_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("[a-zA-Z0-9_-]+"), "Must match [a-zA-Z0-9_-]+"),
					stringvalidator.LengthBetween(1, 255),
				},
			},
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrState: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceNetworkMonitorMonitor) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkMonitorClient(ctx)

	var plan resourceNetworkMonitorMonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := networkmonitor.CreateMonitorInput{
		MonitorName:       plan.MonitorName.ValueStringPointer(),
		AggregationPeriod: plan.AggregationPeriod.ValueInt64Pointer(),
		Tags:              getTagsIn(ctx),
	}

	_, err := conn.CreateMonitor(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionCreating, ResNameMonitor, plan.MonitorName.String(), nil),
			err.Error(),
		)
		return
	}

	out, err := waitMonitorReady(ctx, conn, plan.MonitorName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionWaitingForCreation, ResNameMonitor, plan.MonitorName.ValueString(), nil),
			err.Error(),
		)
	}

	state := plan

	state.ID = flex.StringToFramework(ctx, out.MonitorName)
	state.Arn = flex.StringToFramework(ctx, out.MonitorArn)
	state.State = flex.StringToFramework(ctx, (*string)(&out.State))
	state.CreatedAt = flex.Int64ToFramework(ctx, (aws.Int64(out.CreatedAt.Unix())))
	state.ModifiedAt = flex.Int64ToFramework(ctx, (aws.Int64(out.ModifiedAt.Unix())))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNetworkMonitorMonitor) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkMonitorClient(ctx)

	var state resourceNetworkMonitorMonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := FindMonitorByName(ctx, conn, state.ID.ValueString())
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			resp.State.RemoveResource(ctx)
			return
		}

		if tfresource.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionReading, ResNameMonitor, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)

	state.AggregationPeriod = flex.Int64ToFramework(ctx, output.AggregationPeriod)
	state.MonitorName = flex.StringToFramework(ctx, output.MonitorName)
	state.Arn = flex.StringToFramework(ctx, output.MonitorArn)
	state.State = flex.StringToFramework(ctx, (*string)(&output.State))
	state.CreatedAt = flex.Int64ToFramework(ctx, (aws.Int64(output.CreatedAt.Unix())))
	state.ModifiedAt = flex.Int64ToFramework(ctx, (aws.Int64(output.ModifiedAt.Unix())))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNetworkMonitorMonitor) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkMonitorClient(ctx)

	var plan, state resourceNetworkMonitorMonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AggregationPeriod.Equal(state.AggregationPeriod) {
		input := networkmonitor.UpdateMonitorInput{
			MonitorName:       plan.MonitorName.ValueStringPointer(),
			AggregationPeriod: plan.AggregationPeriod.ValueInt64Pointer(),
		}

		_, err := conn.UpdateMonitor(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionUpdating, ResNameMonitor, state.ID.String(), nil),
				err.Error(),
			)
			return
		}

		_, err = waitMonitorReady(ctx, conn, plan.MonitorName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionCreating, ResNameMonitor, plan.MonitorName.ValueString(), nil),
				err.Error(),
			)
		}
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionWaitingForUpdate, ResNameMonitor, state.ID.String(), nil),
				err.Error(),
			)
			return
		}
	}

	state.AggregationPeriod = flex.Int64ToFramework(ctx, plan.AggregationPeriod.ValueInt64Pointer())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceNetworkMonitorMonitor) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkMonitorClient(ctx)

	var state resourceNetworkMonitorMonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := networkmonitor.DeleteMonitorInput{
		MonitorName: flex.StringFromFramework(ctx, state.MonitorName),
	}

	_, err := conn.DeleteMonitor(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionDeleting, ResNameMonitor, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	_, err = waitMonitorDeleted(ctx, conn, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionWaitingForDeletion, ResNameMonitor, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
}

func statusMonitor(ctx context.Context, conn *networkmonitor.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindMonitorByName(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitMonitorReady(ctx context.Context, conn *networkmonitor.Client, id string) (*networkmonitor.GetMonitorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.MonitorStatePending),
		Target:     enum.Slice(awstypes.MonitorStateActive, awstypes.MonitorStateInactive),
		Refresh:    statusMonitor(ctx, conn, id),
		Timeout:    MonitorTimeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*networkmonitor.GetMonitorOutput); ok {
		return output, err
	}

	return nil, err
}

func waitMonitorDeleted(ctx context.Context, conn *networkmonitor.Client, id string) (*networkmonitor.GetMonitorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.MonitorStateDeleting, awstypes.MonitorStateActive, awstypes.MonitorStateInactive),
		Target:     []string{},
		Refresh:    statusMonitor(ctx, conn, id),
		Timeout:    MonitorTimeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*networkmonitor.GetMonitorOutput); ok {
		return output, err
	}

	return nil, err
}

func FindMonitorByName(ctx context.Context, conn *networkmonitor.Client, name string) (*networkmonitor.GetMonitorOutput, error) {
	input := &networkmonitor.GetMonitorInput{
		MonitorName: &name,
	}

	output, err := conn.GetMonitor(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func (r *resourceNetworkMonitorMonitor) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func (r *resourceNetworkMonitorMonitor) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

type resourceNetworkMonitorMonitorModel struct {
	ID                types.String `tfsdk:"id"`
	Arn               types.String `tfsdk:"arn"`
	AggregationPeriod types.Int64  `tfsdk:"aggregation_period"`
	CreatedAt         types.Int64  `tfsdk:"created_at"`
	ModifiedAt        types.Int64  `tfsdk:"modified_at"`
	MonitorName       types.String `tfsdk:"monitor_name"`
	State             types.String `tfsdk:"state"`
	Tags              types.Map    `tfsdk:"tags"`
	TagsAll           types.Map    `tfsdk:"tags_all"`
}
