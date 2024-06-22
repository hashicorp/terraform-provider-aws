// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmonitor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmonitor/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	ProbeTimeout = time.Minute * 15
	ResNameProbe = "CloudWatch Network Monitor Probe"
)

// @FrameworkResource(name="CloudWatch Network Monitor Probe")
// @Tags(identifierAttribute="arn")
func newResourceProbe(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceNetworkMonitorProbe{}, nil
}

type resourceNetworkMonitorProbe struct {
	framework.ResourceWithConfigure
}

func (r *resourceNetworkMonitorProbe) Metadata(_ context.Context, request resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_networkmonitor_probe"
}

func (r *resourceNetworkMonitorProbe) Schema(ctx context.Context, request resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrARN:        framework.ARNAttributeComputedOnly(),
			"monitor_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("[a-zA-Z0-9_-]+"), "Must match [a-zA-Z0-9_-]+"),
					stringvalidator.LengthBetween(1, 255),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"probe": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"address_family": schema.StringAttribute{
						Computed: true,
					},
					names.AttrCreatedAt: schema.Int64Attribute{
						Computed: true,
					},
					names.AttrDestination: schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 255),
						},
					},
					"destination_port": schema.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							int64validator.Between(0, 65536),
						},
					},
					"modified_at": schema.Int64Attribute{
						Computed: true,
					},
					"packet_size": schema.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							int64validator.Between(56, 8500),
						},
					},
					"probe_arn": schema.StringAttribute{
						Computed: true,
					},
					"probe_id": schema.StringAttribute{
						Computed: true,
					},
					names.AttrProtocol: schema.StringAttribute{
						Required: true,
					},
					"source_arn": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(20, 2048),
							stringvalidator.RegexMatches(regexache.MustCompile("arn:.*"), "Must match pattern arn:*"),
						},
					},
					names.AttrTags: schema.MapAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
					names.AttrState: schema.StringAttribute{
						Computed: true,
					},
					names.AttrVPCID: schema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func (r *resourceNetworkMonitorProbe) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkMonitorClient(ctx)

	var state resourceNetworkMonitorProbeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var probe probeConfigModel
	resp.Diagnostics.Append(
		state.Probe.As(
			ctx,
			&probe,
			basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})...)

	probeConfig := expandProbeConfig(ctx, probe)
	if resp.Diagnostics.HasError() {
		return
	}

	input := networkmonitor.CreateProbeInput{
		MonitorName: state.MonitorName.ValueStringPointer(),
		Probe:       &probeConfig,
		Tags:        flex.ExpandFrameworkStringValueMap(ctx, state.Tags),
	}

	out, err := conn.CreateProbe(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionCreating, ResNameProbe, state.MonitorName.String(), nil),
			err.Error(),
		)
		return
	}

	probeID := fmt.Sprintf("%s:%s", *out.ProbeId, *state.MonitorName.ValueStringPointer())

	outWait, err := waitProbeReady(ctx, conn, probeID)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionWaitingForCreation, ResNameProbe, probeID, nil),
			err.Error(),
		)
	}

	state.ID = flex.StringToFramework(ctx, &probeID)
	state.MonitorName = flex.StringToFramework(ctx, state.MonitorName.ValueStringPointer())
	p, d := flattenProbeConfig(ctx, *outWait)
	resp.Diagnostics.Append(d...)
	state.Probe = p
	state.Arn = flex.StringToFramework(ctx, out.ProbeArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNetworkMonitorProbe) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkMonitorClient(ctx)

	var state resourceNetworkMonitorProbeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindProbeByID(ctx, conn, state.ID.ValueString())
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
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionReading, ResNameProbe, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, state.ID.ValueStringPointer())
	_, monitorName, err := probeParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionReading, ResNameProbe, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.MonitorName = flex.StringToFramework(ctx, &monitorName)
	state.Arn = flex.StringToFramework(ctx, out.ProbeArn)
	p, d := flattenProbeConfig(ctx, *out)
	resp.Diagnostics.Append(d...)
	state.Probe = p

	setTagsOut(ctx, out.Tags)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNetworkMonitorProbe) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkMonitorClient(ctx)

	var plan, state resourceNetworkMonitorProbeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var probePlan, probeState probeConfigModel
	resp.Diagnostics.Append(
		plan.Probe.As(
			ctx,
			&probePlan,
			basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})...)

	resp.Diagnostics.Append(
		state.Probe.As(
			ctx,
			&probeState,
			basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})...)

	probeID, monitorName, err := probeParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionUpdating, ResNameProbe, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	in := networkmonitor.UpdateProbeInput{
		MonitorName: &monitorName,
		ProbeId:     &probeID,
	}

	if !probePlan.Destination.Equal(probeState.Destination) {
		in.Destination = probePlan.Destination.ValueStringPointer()
	}
	if !probePlan.DestinationPort.Equal(probeState.DestinationPort) {
		in.DestinationPort = aws.Int32(int32(probePlan.DestinationPort.ValueInt64()))
	}
	if !probePlan.PacketSize.Equal(probeState.PacketSize) {
		in.PacketSize = aws.Int32(int32(probePlan.PacketSize.ValueInt64()))
	}
	if !probePlan.Protocol.Equal(probeState.Protocol) {
		in.Protocol = awstypes.Protocol(*probePlan.Protocol.ValueStringPointer())
	}

	_, err = conn.UpdateProbe(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionUpdating, ResNameProbe, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	out, err := waitProbeReady(ctx, conn, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionWaitingForUpdate, ResNameProbe, state.ID.String(), nil),
			err.Error(),
		)
	}
	state = plan
	p, d := flattenProbeConfig(ctx, *out)
	resp.Diagnostics.Append(d...)
	state.Probe = p

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNetworkMonitorProbe) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkMonitorClient(ctx)

	var state resourceNetworkMonitorProbeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	probeID, monitorName, err := probeParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionDeleting, ResNameProbe, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	input := networkmonitor.DeleteProbeInput{
		MonitorName: &monitorName,
		ProbeId:     &probeID,
	}
	_, err = conn.DeleteProbe(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionDeleting, ResNameProbe, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	_, err = waitProbeDeleted(ctx, conn, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkMonitor, create.ErrActionWaitingForDeletion, ResNameProbe, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
}

func (r *resourceNetworkMonitorProbe) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func (r *resourceNetworkMonitorProbe) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func statusProbe(ctx context.Context, conn *networkmonitor.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProbeByID(ctx, conn, id)

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

func waitProbeReady(ctx context.Context, conn *networkmonitor.Client, id string) (*networkmonitor.GetProbeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ProbeStatePending),
		Target:     enum.Slice(awstypes.ProbeStateActive, awstypes.ProbeStateInactive),
		Refresh:    statusProbe(ctx, conn, id),
		Timeout:    ProbeTimeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*networkmonitor.GetProbeOutput); ok {
		return output, err
	}

	return nil, err
}

func waitProbeDeleted(ctx context.Context, conn *networkmonitor.Client, id string) (*networkmonitor.GetProbeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ProbeStateActive, awstypes.ProbeStateInactive, awstypes.ProbeStateDeleting),
		Target:     []string{},
		Refresh:    statusProbe(ctx, conn, id),
		Timeout:    ProbeTimeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*networkmonitor.GetProbeOutput); ok {
		return output, err
	}

	return nil, err
}

func FindProbeByID(ctx context.Context, conn *networkmonitor.Client, id string) (*networkmonitor.GetProbeOutput, error) {
	probeID, monitorName, err := probeParseID(id)
	if err != nil {
		return nil, err
	}

	input := &networkmonitor.GetProbeInput{
		ProbeId:     &probeID,
		MonitorName: &monitorName,
	}

	output, err := conn.GetProbe(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func probeParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected probeID:monitorName", id)
	}

	return parts[0], parts[1], nil
}

var probeConfigTypes = map[string]attr.Type{
	"address_family":   types.StringType,
	names.AttrCreatedAt:       types.Int64Type,
	names.AttrDestination:      types.StringType,
	"destination_port": types.Int64Type,
	"modified_at":      types.Int64Type,
	"packet_size":      types.Int64Type,
	"probe_arn":        types.StringType,
	"probe_id":         types.StringType,
	names.AttrProtocol:         types.StringType,
	"source_arn":       types.StringType,
	names.AttrState:            types.StringType,
	names.AttrTags:             types.MapType{ElemType: types.StringType},
	names.AttrVPCID:           types.StringType,
}

type resourceNetworkMonitorProbeModel struct {
	ID          types.String `tfsdk:"id"`
	Arn         types.String `tfsdk:"arn"`
	MonitorName types.String `tfsdk:"monitor_name"`
	Probe       types.Object `tfsdk:"probe"`
	Tags        types.Map    `tfsdk:"tags"`
	TagsAll     types.Map    `tfsdk:"tags_all"`
}

type probeConfigModel struct {
	AddressFamily   types.String `tfsdk:"address_family"`
	CreatedAt       types.Int64  `tfsdk:"created_at"`
	Destination     types.String `tfsdk:"destination"`
	DestinationPort types.Int64  `tfsdk:"destination_port"`
	ModifiedAt      types.Int64  `tfsdk:"modified_at"`
	PacketSize      types.Int64  `tfsdk:"packet_size"`
	ProbeArn        types.String `tfsdk:"probe_arn"`
	ProbeId         types.String `tfsdk:"probe_id"`
	Tags            types.Map    `tfsdk:"tags"`
	Protocol        types.String `tfsdk:"protocol"`
	SourceArn       types.String `tfsdk:"source_arn"`
	State           types.String `tfsdk:"state"`
	VpcId           types.String `tfsdk:"vpc_id"`
}

func flattenProbeConfig(ctx context.Context, object networkmonitor.GetProbeOutput) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	t := map[string]attr.Value{
		"address_family":   flex.StringToFramework(ctx, (*string)(&object.AddressFamily)),
		names.AttrCreatedAt:       flex.Int64ToFramework(ctx, aws.Int64(object.CreatedAt.Unix())),
		names.AttrDestination:      flex.StringToFramework(ctx, object.Destination),
		"destination_port": flex.Int64ToFramework(ctx, aws.Int64(int64(*object.DestinationPort))),
		"modified_at":      flex.Int64ToFramework(ctx, aws.Int64(object.ModifiedAt.Unix())),
		"packet_size":      flex.Int64ToFramework(ctx, aws.Int64(int64(*object.PacketSize))),
		"probe_arn":        flex.StringToFramework(ctx, object.ProbeArn),
		"probe_id":         flex.StringToFramework(ctx, object.ProbeId),
		names.AttrProtocol:         flex.StringToFramework(ctx, (*string)(&object.Protocol)),
		"source_arn":       flex.StringToFramework(ctx, object.SourceArn),
		names.AttrState:            flex.StringToFramework(ctx, (*string)(&object.State)),
		names.AttrTags:             flex.FlattenFrameworkStringValueMap(ctx, object.Tags),
		names.AttrVPCID:           flex.StringToFramework(ctx, object.VpcId),
	}
	objVal, d := types.ObjectValue(probeConfigTypes, t)
	diags.Append(d...)

	return objVal, diags
}

func expandProbeConfig(ctx context.Context, object probeConfigModel) awstypes.ProbeInput {
	return awstypes.ProbeInput{
		Destination:     object.Destination.ValueStringPointer(),
		DestinationPort: aws.Int32(int32(object.DestinationPort.ValueInt64())),
		PacketSize:      aws.Int32(int32(object.PacketSize.ValueInt64())),
		Protocol:        awstypes.Protocol(object.Protocol.ValueString()),
		SourceArn:       object.SourceArn.ValueStringPointer(),
		Tags:            flex.ExpandFrameworkStringValueMap(ctx, object.Tags),
	}
}
