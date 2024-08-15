// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	mltypes "github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceMultiplexProgram(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &multiplexProgram{}, nil
}

const (
	ResNameMultiplexProgram = "Multiplex Program"
)

type multiplexProgram struct {
	framework.ResourceWithConfigure
}

func (m *multiplexProgram) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_medialive_multiplex_program"
}

func (m *multiplexProgram) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"multiplex_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"program_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"multiplex_program_settings": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"program_number": schema.Int64Attribute{
							Required: true,
						},
						"preferred_channel_pipeline": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[mltypes.PreferredChannelPipeline](),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"service_descriptor": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrProviderName: schema.StringAttribute{
										Required: true,
									},
									names.AttrServiceName: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"video_settings": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"constant_bitrate": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.UseStateForUnknown(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"statmux_settings": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"minimum_bitrate": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
												"maximum_bitrate": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
												names.AttrPriority: schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (m *multiplexProgram) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var plan resourceMultiplexProgramData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	multiplexId := plan.MultiplexID.ValueString()
	programName := plan.ProgramName.ValueString()

	in := &medialive.CreateMultiplexProgramInput{
		MultiplexId: aws.String(multiplexId),
		ProgramName: aws.String(programName),
		RequestId:   aws.String(id.UniqueId()),
	}

	mps := make(multiplexProgramSettingsObject, 1)
	resp.Diagnostics.Append(plan.MultiplexProgramSettings.ElementsAs(ctx, &mps, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpSettings, err := mps.expand(ctx)

	resp.Diagnostics.Append(err...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.MultiplexProgramSettings = mpSettings

	out, errCreate := conn.CreateMultiplexProgram(ctx, in)

	if errCreate != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionCreating, ResNameMultiplexProgram, plan.ProgramName.String(), nil),
			errCreate.Error(),
		)
		return
	}

	var result resourceMultiplexProgramData

	result.ID = flex.StringValueToFramework(ctx, fmt.Sprintf("%s/%s", programName, multiplexId))
	result.ProgramName = flex.StringToFrameworkLegacy(ctx, out.MultiplexProgram.ProgramName)
	result.MultiplexID = plan.MultiplexID
	result.MultiplexProgramSettings = flattenMultiplexProgramSettings(ctx, out.MultiplexProgram.MultiplexProgramSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var state resourceMultiplexProgramData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseMultiplexProgramID(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	out, err := FindMultiplexProgramByID(ctx, conn, multiplexId, programName)

	if tfresource.NotFound(err) {
		resp.Diagnostics.AddWarning(
			"AWS Resource Not Found During Refresh",
			fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original Error: %s", err.Error()),
		)
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	state.MultiplexProgramSettings = flattenMultiplexProgramSettings(ctx, out.MultiplexProgramSettings)
	state.ProgramName = types.StringValue(aws.ToString(out.ProgramName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var plan resourceMultiplexProgramData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseMultiplexProgramID(plan.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, plan.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	mps := make(multiplexProgramSettingsObject, 1)
	resp.Diagnostics.Append(plan.MultiplexProgramSettings.ElementsAs(ctx, &mps, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpSettings, errExpand := mps.expand(ctx)

	resp.Diagnostics.Append(errExpand...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &medialive.UpdateMultiplexProgramInput{
		MultiplexId:              aws.String(multiplexId),
		ProgramName:              aws.String(programName),
		MultiplexProgramSettings: mpSettings,
	}

	_, errUpdate := conn.UpdateMultiplexProgram(ctx, in)

	if errUpdate != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionUpdating, ResNameMultiplexProgram, plan.ProgramName.String(), nil),
			errUpdate.Error(),
		)
		return
	}

	//Need to find multiplex program because output from update does not provide state data
	out, errUpdate := FindMultiplexProgramByID(ctx, conn, multiplexId, programName)

	if errUpdate != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionUpdating, ResNameMultiplexProgram, plan.ProgramName.String(), nil),
			errUpdate.Error(),
		)
		return
	}

	plan.MultiplexProgramSettings = flattenMultiplexProgramSettings(ctx, out.MultiplexProgramSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (m *multiplexProgram) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var state resourceMultiplexProgramData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseMultiplexProgramID(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionDeleting, ResNameMultiplexProgram, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteMultiplexProgram(ctx, &medialive.DeleteMultiplexProgramInput{
		MultiplexId: aws.String(multiplexId),
		ProgramName: aws.String(programName),
	})

	if err != nil {
		var nfe *mltypes.NotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionDeleting, ResNameMultiplexProgram, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}
}

func (m *multiplexProgram) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindMultiplexProgramByID(ctx context.Context, conn *medialive.Client, multiplexId, programName string) (*medialive.DescribeMultiplexProgramOutput, error) {
	in := &medialive.DescribeMultiplexProgramInput{
		MultiplexId: aws.String(multiplexId),
		ProgramName: aws.String(programName),
	}
	out, err := conn.DescribeMultiplexProgram(ctx, in)
	if err != nil {
		var nfe *mltypes.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type multiplexProgramSettingsObject []multiplexProgramSettings

func (mps multiplexProgramSettingsObject) expand(ctx context.Context) (*mltypes.MultiplexProgramSettings, diag.Diagnostics) {
	if len(mps) == 0 {
		return nil, nil
	}

	data := mps[0]

	l := &mltypes.MultiplexProgramSettings{
		ProgramNumber:            flex.Int32FromFramework(ctx, data.ProgramNumber),
		PreferredChannelPipeline: mltypes.PreferredChannelPipeline(data.PreferredChannelPipeline.ValueString()),
	}

	if len(data.ServiceDescriptor.Elements()) > 0 && !data.ServiceDescriptor.IsNull() {
		sd := make(serviceDescriptorObject, 1)
		err := data.ServiceDescriptor.ElementsAs(ctx, &sd, false)
		if err.HasError() {
			return nil, err
		}

		l.ServiceDescriptor = sd.expand(ctx)
	}

	if len(data.VideoSettings.Elements()) > 0 && !data.VideoSettings.IsNull() {
		vs := make(videoSettingsObject, 1)
		err := data.VideoSettings.ElementsAs(ctx, &vs, false)
		if err.HasError() {
			return nil, err
		}

		l.VideoSettings = vs.expand(ctx)

		if len(vs[0].StatmuxSettings.Elements()) > 0 && !vs[0].StatmuxSettings.IsNull() {
			sms := make(statmuxSettingsObject, 1)
			err := vs[0].StatmuxSettings.ElementsAs(ctx, &sms, false)
			if err.HasError() {
				return nil, err
			}

			l.VideoSettings.StatmuxSettings = sms.expand(ctx)
		}
	}

	return l, nil
}

type serviceDescriptorObject []serviceDescriptor

func (sd serviceDescriptorObject) expand(ctx context.Context) *mltypes.MultiplexProgramServiceDescriptor {
	if len(sd) == 0 {
		return nil
	}

	return &mltypes.MultiplexProgramServiceDescriptor{
		ProviderName: flex.StringFromFramework(ctx, sd[0].ProviderName),
		ServiceName:  flex.StringFromFramework(ctx, sd[0].ServiceName),
	}
}

type videoSettingsObject []videoSettings

func (vs videoSettingsObject) expand(ctx context.Context) *mltypes.MultiplexVideoSettings {
	if len(vs) == 0 {
		return nil
	}

	return &mltypes.MultiplexVideoSettings{
		ConstantBitrate: flex.Int32FromFramework(ctx, vs[0].ConstantBitrate),
	}
}

type statmuxSettingsObject []statmuxSettings

func (sms statmuxSettingsObject) expand(ctx context.Context) *mltypes.MultiplexStatmuxVideoSettings {
	if len(sms) == 0 {
		return nil
	}

	return &mltypes.MultiplexStatmuxVideoSettings{
		MaximumBitrate: flex.Int32FromFramework(ctx, sms[0].MaximumBitrate),
		MinimumBitrate: flex.Int32FromFramework(ctx, sms[0].MinimumBitrate),
		Priority:       flex.Int32FromFramework(ctx, sms[0].Priority),
	}
}

var (
	statmuxAttrs = map[string]attr.Type{
		"minimum_bitrate":  types.Int64Type,
		"maximum_bitrate":  types.Int64Type,
		names.AttrPriority: types.Int64Type,
	}

	videoSettingsAttrs = map[string]attr.Type{
		"constant_bitrate": types.Int64Type,
		"statmux_settings": types.ListType{ElemType: types.ObjectType{AttrTypes: statmuxAttrs}},
	}

	serviceDescriptorAttrs = map[string]attr.Type{
		names.AttrProviderName: types.StringType,
		names.AttrServiceName:  types.StringType,
	}

	multiplexProgramSettingsAttrs = map[string]attr.Type{
		"program_number":             types.Int64Type,
		"preferred_channel_pipeline": types.StringType,
		"service_descriptor":         types.ListType{ElemType: types.ObjectType{AttrTypes: serviceDescriptorAttrs}},
		"video_settings":             types.ListType{ElemType: types.ObjectType{AttrTypes: videoSettingsAttrs}},
	}
)

func flattenMultiplexProgramSettings(ctx context.Context, mps *mltypes.MultiplexProgramSettings) types.List {
	elemType := types.ObjectType{AttrTypes: multiplexProgramSettingsAttrs}

	if mps == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{}
	attrs["program_number"] = flex.Int32ToFramework(ctx, mps.ProgramNumber)
	attrs["preferred_channel_pipeline"] = flex.StringValueToFrameworkLegacy(ctx, mps.PreferredChannelPipeline)
	attrs["service_descriptor"] = flattenServiceDescriptor(ctx, mps.ServiceDescriptor)
	attrs["video_settings"] = flattenVideoSettings(ctx, mps.VideoSettings)

	vals := types.ObjectValueMust(multiplexProgramSettingsAttrs, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}

func flattenServiceDescriptor(ctx context.Context, sd *mltypes.MultiplexProgramServiceDescriptor) types.List {
	elemType := types.ObjectType{AttrTypes: serviceDescriptorAttrs}

	if sd == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{}
	attrs[names.AttrProviderName] = flex.StringToFrameworkLegacy(ctx, sd.ProviderName)
	attrs[names.AttrServiceName] = flex.StringToFrameworkLegacy(ctx, sd.ServiceName)

	vals := types.ObjectValueMust(serviceDescriptorAttrs, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}

func flattenStatMuxSettings(ctx context.Context, mps *mltypes.MultiplexStatmuxVideoSettings) types.List {
	elemType := types.ObjectType{AttrTypes: statmuxAttrs}

	if mps == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{}
	attrs["minimum_bitrate"] = flex.Int32ToFramework(ctx, mps.MinimumBitrate)
	attrs["maximum_bitrate"] = flex.Int32ToFramework(ctx, mps.MaximumBitrate)
	attrs[names.AttrPriority] = flex.Int32ToFramework(ctx, mps.Priority)

	vals := types.ObjectValueMust(statmuxAttrs, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}

func flattenVideoSettings(ctx context.Context, mps *mltypes.MultiplexVideoSettings) types.List {
	elemType := types.ObjectType{AttrTypes: videoSettingsAttrs}

	if mps == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{}
	attrs["constant_bitrate"] = flex.Int32ToFramework(ctx, mps.ConstantBitrate)
	attrs["statmux_settings"] = flattenStatMuxSettings(ctx, mps.StatmuxSettings)

	vals := types.ObjectValueMust(videoSettingsAttrs, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}

func ParseMultiplexProgramID(id string) (programName string, multiplexId string, err error) {
	idParts := strings.Split(id, "/")

	if len(idParts) < 2 || (idParts[0] == "" || idParts[1] == "") {
		err = errors.New("invalid id")
		return
	}

	programName = idParts[0]
	multiplexId = idParts[1]

	return
}

type resourceMultiplexProgramData struct {
	ID                       types.String `tfsdk:"id"`
	MultiplexID              types.String `tfsdk:"multiplex_id"`
	MultiplexProgramSettings types.List   `tfsdk:"multiplex_program_settings"`
	ProgramName              types.String `tfsdk:"program_name"`
}

type multiplexProgramSettings struct {
	ProgramNumber            types.Int64  `tfsdk:"program_number"`
	PreferredChannelPipeline types.String `tfsdk:"preferred_channel_pipeline"`
	ServiceDescriptor        types.List   `tfsdk:"service_descriptor"`
	VideoSettings            types.List   `tfsdk:"video_settings"`
}

type serviceDescriptor struct {
	ProviderName types.String `tfsdk:"provider_name"`
	ServiceName  types.String `tfsdk:"service_name"`
}

type videoSettings struct {
	ConstantBitrate types.Int64 `tfsdk:"constant_bitrate"`
	StatmuxSettings types.List  `tfsdk:"statmux_settings"`
}

type statmuxSettings struct {
	MaximumBitrate types.Int64 `tfsdk:"maximum_bitrate"`
	MinimumBitrate types.Int64 `tfsdk:"minimum_bitrate"`
	Priority       types.Int64 `tfsdk:"priority"`
}
