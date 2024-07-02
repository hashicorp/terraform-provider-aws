// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmonitor/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Probe")
// @Tags(identifierAttribute="arn")
func newProbeResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &probeResource{}, nil
}

type probeResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (*probeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_networkmonitor_probe"
}

func (r *probeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address_family": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AddressFamily](),
				Computed:   true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
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
			names.AttrID: framework.IDAttribute(),
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
			"packet_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(56, 8500),
				},
			},
			"probe_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrProtocol: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Protocol](),
				Required:   true,
			},
			"source_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *probeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data probeResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkMonitorClient(ctx)

	probeInput := &awstypes.ProbeInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, probeInput)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &networkmonitor.CreateProbeInput{
		ClientToken: aws.String(id.UniqueId()),
		MonitorName: fwflex.StringFromFramework(ctx, data.MonitorName),
		Probe:       probeInput,
		Tags:        getTagsIn(ctx),
	}

	outputCP, err := conn.CreateProbe(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating CloudWatch Network Monitor Probe (%s)", err.Error())

		return
	}

	// Set values for unknowns.
	data.ProbeARN = fwflex.StringToFramework(ctx, outputCP.ProbeArn)
	data.ProbeID = fwflex.StringToFramework(ctx, outputCP.ProbeId)
	data.setID()

	outputGP, err := waitProbeReady(ctx, conn, data.MonitorName.ValueString(), data.ProbeID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudWatch Network Monitor Probe (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.AddressFamily = fwtypes.StringEnumValue(outputGP.AddressFamily)
	if data.PacketSize.IsUnknown() {
		data.PacketSize = fwflex.Int32ToFramework(ctx, outputGP.PacketSize)
	}
	data.VpcID = fwflex.StringToFramework(ctx, outputGP.VpcId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *probeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data probeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().NetworkMonitorClient(ctx)

	output, err := findProbeByTwoPartKey(ctx, conn, data.MonitorName.ValueString(), data.ProbeID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Network Monitor Probe (%s)", data.ID.String()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *probeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new probeResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkMonitorClient(ctx)

	if !new.Destination.Equal(old.Destination) ||
		!new.DestinationPort.Equal(old.DestinationPort) ||
		!new.PacketSize.Equal(old.PacketSize) ||
		!new.Protocol.Equal(old.Protocol) {
		input := &networkmonitor.UpdateProbeInput{
			MonitorName: fwflex.StringFromFramework(ctx, new.MonitorName),
			ProbeId:     fwflex.StringFromFramework(ctx, new.ProbeID),
		}

		if !new.Destination.Equal(old.Destination) {
			input.Destination = fwflex.StringFromFramework(ctx, new.Destination)
		}
		if !new.DestinationPort.Equal(old.DestinationPort) {
			input.DestinationPort = fwflex.Int32FromFramework(ctx, new.DestinationPort)
		}
		if !new.PacketSize.Equal(old.PacketSize) {
			input.PacketSize = fwflex.Int32FromFramework(ctx, new.PacketSize)
		}
		if !new.Protocol.Equal(old.Protocol) {
			input.Protocol = new.Protocol.ValueEnum()
		}

		_, err := conn.UpdateProbe(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudWatch Network Monitor Probe (%s)", new.ID.String()), err.Error())

			return
		}

		outputGP, err := waitProbeReady(ctx, conn, new.MonitorName.ValueString(), new.ProbeID.ValueString())

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudWatch Network Monitor Probe (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		new.AddressFamily = fwtypes.StringEnumValue(outputGP.AddressFamily)
	} else {
		new.AddressFamily = old.AddressFamily
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *probeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data probeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkMonitorClient(ctx)

	_, err := conn.DeleteProbe(ctx, &networkmonitor.DeleteProbeInput{
		MonitorName: fwflex.StringFromFramework(ctx, data.MonitorName),
		ProbeId:     fwflex.StringFromFramework(ctx, data.ProbeID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Network Monitor Probe (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitProbeDeleted(ctx, conn, data.MonitorName.ValueString(), data.ProbeID.ValueString()); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudWatch Network Monitor Probe (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *probeResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findProbeByTwoPartKey(ctx context.Context, conn *networkmonitor.Client, monitorName, probeID string) (*networkmonitor.GetProbeOutput, error) {
	input := &networkmonitor.GetProbeInput{
		MonitorName: aws.String(monitorName),
		ProbeId:     aws.String(probeID),
	}

	output, err := conn.GetProbe(ctx, input)

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

func statusProbe(ctx context.Context, conn *networkmonitor.Client, monitorName, probeID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findProbeByTwoPartKey(ctx, conn, monitorName, probeID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitProbeReady(ctx context.Context, conn *networkmonitor.Client, monitorName, probeID string) (*networkmonitor.GetProbeOutput, error) {
	const (
		timeout = time.Minute * 15
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ProbeStatePending),
		Target:     enum.Slice(awstypes.ProbeStateActive, awstypes.ProbeStateInactive),
		Refresh:    statusProbe(ctx, conn, monitorName, probeID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmonitor.GetProbeOutput); ok {
		return output, err
	}

	return nil, err
}

func waitProbeDeleted(ctx context.Context, conn *networkmonitor.Client, monitorName, probeID string) (*networkmonitor.GetProbeOutput, error) {
	const (
		timeout = time.Minute * 15
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ProbeStateActive, awstypes.ProbeStateInactive, awstypes.ProbeStateDeleting),
		Target:     []string{},
		Refresh:    statusProbe(ctx, conn, monitorName, probeID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmonitor.GetProbeOutput); ok {
		return output, err
	}

	return nil, err
}

type probeResourceModel struct {
	AddressFamily   fwtypes.StringEnum[awstypes.AddressFamily] `tfsdk:"address_family"`
	Destination     types.String                               `tfsdk:"destination"`
	DestinationPort types.Int64                                `tfsdk:"destination_port"`
	ID              types.String                               `tfsdk:"id"`
	MonitorName     types.String                               `tfsdk:"monitor_name"`
	PacketSize      types.Int64                                `tfsdk:"packet_size"`
	ProbeARN        types.String                               `tfsdk:"arn"`
	ProbeID         types.String                               `tfsdk:"probe_id"`
	Protocol        fwtypes.StringEnum[awstypes.Protocol]      `tfsdk:"protocol"`
	SourceARN       fwtypes.ARN                                `tfsdk:"source_arn"`
	Tags            types.Map                                  `tfsdk:"tags"`
	TagsAll         types.Map                                  `tfsdk:"tags_all"`
	VpcID           types.String                               `tfsdk:"vpc_id"`
}

const (
	probeResourceIDPartCount = 2
)

func (m *probeResourceModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, probeResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.MonitorName = types.StringValue(parts[0])
	m.ProbeID = types.StringValue(parts[1])

	return nil
}

func (m *probeResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.MonitorName.ValueString(), m.ProbeID.ValueString()}, probeResourceIDPartCount, false)))
}
