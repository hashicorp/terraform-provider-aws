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
	awstypes "github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_medialive_multiplex_program", name="Multiplex Program")
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiplexProgramSettings](ctx),
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
							CustomType: fwtypes.StringEnumType[awstypes.PreferredChannelPipeline](),
							Required:   true,
							//Validators: []validator.String{
							//	enum.FrameworkValidate[awstypes.PreferredChannelPipeline](),
							//},
						},
					},
					Blocks: map[string]schema.Block{
						"service_descriptor": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[serviceDescriptor](ctx),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[videoSettings](ctx),
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
										CustomType: fwtypes.NewListNestedObjectTypeOf[statmuxSettings](ctx),
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

	input := medialive.CreateMultiplexProgramInput{
		RequestId: aws.String(id.UniqueId()),
	}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	multiplexId := plan.MultiplexID.ValueString()
	programName := plan.ProgramName.ValueString()
	//
	//in := &medialive.CreateMultiplexProgramInput{
	//	MultiplexId: aws.String(multiplexId),
	//	ProgramName: aws.String(programName),
	//	RequestId:   aws.String(id.UniqueId()),
	//}
	//
	//mps := make(multiplexProgramSettingsObject, 1)
	//resp.Diagnostics.Append(plan.MultiplexProgramSettings.ElementsAs(ctx, &mps, false)...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//mpSettings, err := mps.expand(ctx)
	//
	//resp.Diagnostics.Append(err...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//in.MultiplexProgramSettings = mpSettings

	out, err := conn.CreateMultiplexProgram(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionCreating, ResNameMultiplexProgram, plan.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = fwflex.StringValueToFramework(ctx, fmt.Sprintf("%s/%s", programName, multiplexId))
	//result.ProgramName = flex.StringToFrameworkLegacy(ctx, out.MultiplexProgram.ProgramName)
	//result.MultiplexID = plan.MultiplexID
	//result.MultiplexProgramSettings = flattenMultiplexProgramSettings(ctx, out.MultiplexProgram.MultiplexProgramSettings)

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, state.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	out, err := FindMultiplexProgramByID(ctx, conn, multiplexId, programName)

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, state.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	//state.MultiplexProgramSettings = flattenMultiplexProgramSettings(ctx, out.MultiplexProgramSettings)
	//state.ProgramName = types.StringValue(aws.ToString(out.ProgramName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (m *multiplexProgram) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := m.Meta().MediaLiveClient(ctx)

	var plan, state resourceMultiplexProgramData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseMultiplexProgramID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, plan.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	diff, d := fwflex.Calculate(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := medialive.UpdateMultiplexProgramInput{}
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.ProgramName = aws.String(programName)
		input.MultiplexId = aws.String(multiplexId)

		//mps := make(multiplexProgramSettingsObject, 1)
		//resp.Diagnostics.Append(plan.MultiplexProgramSettings.ElementsAs(ctx, &mps, false)...)
		//if resp.Diagnostics.HasError() {
		//	return
		//}
		//
		//mpSettings, errExpand := mps.expand(ctx)
		//
		//resp.Diagnostics.Append(errExpand...)
		//if resp.Diagnostics.HasError() {
		//	return
		//}
		//
		//in := &medialive.UpdateMultiplexProgramInput{
		//	MultiplexId:              aws.String(multiplexId),
		//	ProgramName:              aws.String(programName),
		//	MultiplexProgramSettings: mpSettings,
		//}

		_, err = conn.UpdateMultiplexProgram(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaLive, create.ErrActionUpdating, ResNameMultiplexProgram, plan.ProgramName.String(), err),
				err.Error(),
			)
			return
		}

		//Need to find multiplex program because output from update does not provide state data
		output, err := FindMultiplexProgramByID(ctx, conn, multiplexId, programName)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaLive, create.ErrActionUpdating, ResNameMultiplexProgram, plan.ProgramName.String(), err),
				err.Error(),
			)
			return
		}

		// plan.MultiplexProgramSettings = flattenMultiplexProgramSettings(ctx, out.MultiplexProgramSettings)

		resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

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
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionDeleting, ResNameMultiplexProgram, state.ProgramName.String(), err),
			err.Error(),
		)
		return
	}

	input := medialive.DeleteMultiplexProgramInput{
		MultiplexId: aws.String(multiplexId),
		ProgramName: aws.String(programName),
	}
	_, err = conn.DeleteMultiplexProgram(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionDeleting, ResNameMultiplexProgram, state.ProgramName.String(), err),
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
		var nfe *awstypes.NotFoundException
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

//type multiplexProgramSettingsObject []multiplexProgramSettings
//
//func (mps multiplexProgramSettingsObject) expand(ctx context.Context) (*awstypes.MultiplexProgramSettings, diag.Diagnostics) {
//	if len(mps) == 0 {
//		return nil, nil
//	}
//
//	data := mps[0]
//
//	l := &awstypes.MultiplexProgramSettings{
//		ProgramNumber:            flex.Int32FromFramework(ctx, data.ProgramNumber),
//		PreferredChannelPipeline: awstypes.PreferredChannelPipeline(data.PreferredChannelPipeline.ValueString()),
//	}
//
//	if len(data.ServiceDescriptor.Elements()) > 0 && !data.ServiceDescriptor.IsNull() {
//		sd := make(serviceDescriptorObject, 1)
//		err := data.ServiceDescriptor.ElementsAs(ctx, &sd, false)
//		if err.HasError() {
//			return nil, err
//		}
//
//		l.ServiceDescriptor = sd.expand(ctx)
//	}
//
//	if len(data.VideoSettings.Elements()) > 0 && !data.VideoSettings.IsNull() {
//		vs := make(videoSettingsObject, 1)
//		err := data.VideoSettings.ElementsAs(ctx, &vs, false)
//		if err.HasError() {
//			return nil, err
//		}
//
//		l.VideoSettings = vs.expand(ctx)
//
//		if len(vs[0].StatmuxSettings.Elements()) > 0 && !vs[0].StatmuxSettings.IsNull() {
//			sms := make(statmuxSettingsObject, 1)
//			err := vs[0].StatmuxSettings.ElementsAs(ctx, &sms, false)
//			if err.HasError() {
//				return nil, err
//			}
//
//			l.VideoSettings.StatmuxSettings = sms.expand(ctx)
//		}
//	}
//
//	return l, nil
//}
//
//type serviceDescriptorObject []serviceDescriptor
//
//func (sd serviceDescriptorObject) expand(ctx context.Context) *awstypes.MultiplexProgramServiceDescriptor {
//	if len(sd) == 0 {
//		return nil
//	}
//
//	return &awstypes.MultiplexProgramServiceDescriptor{
//		ProviderName: flex.StringFromFramework(ctx, sd[0].ProviderName),
//		ServiceName:  flex.StringFromFramework(ctx, sd[0].ServiceName),
//	}
//}
//
//type videoSettingsObject []videoSettings
//
//func (vs videoSettingsObject) expand(ctx context.Context) *awstypes.MultiplexVideoSettings {
//	if len(vs) == 0 {
//		return nil
//	}
//
//	return &awstypes.MultiplexVideoSettings{
//		ConstantBitrate: flex.Int32FromFramework(ctx, vs[0].ConstantBitrate),
//	}
//}
//
//type statmuxSettingsObject []statmuxSettings
//
//func (sms statmuxSettingsObject) expand(ctx context.Context) *awstypes.MultiplexStatmuxVideoSettings {
//	if len(sms) == 0 {
//		return nil
//	}
//
//	return &awstypes.MultiplexStatmuxVideoSettings{
//		MaximumBitrate: flex.Int32FromFramework(ctx, sms[0].MaximumBitrate),
//		MinimumBitrate: flex.Int32FromFramework(ctx, sms[0].MinimumBitrate),
//		Priority:       flex.Int32FromFramework(ctx, sms[0].Priority),
//	}
//}
//
//var (
//	statmuxAttrs = map[string]attr.Type{
//		"minimum_bitrate":  types.Int64Type,
//		"maximum_bitrate":  types.Int64Type,
//		names.AttrPriority: types.Int64Type,
//	}
//
//	videoSettingsAttrs = map[string]attr.Type{
//		"constant_bitrate": types.Int64Type,
//		"statmux_settings": types.ListType{ElemType: types.ObjectType{AttrTypes: statmuxAttrs}},
//	}
//
//	serviceDescriptorAttrs = map[string]attr.Type{
//		names.AttrProviderName: types.StringType,
//		names.AttrServiceName:  types.StringType,
//	}
//
//	multiplexProgramSettingsAttrs = map[string]attr.Type{
//		"program_number":             types.Int64Type,
//		"preferred_channel_pipeline": types.StringType,
//		"service_descriptor":         types.ListType{ElemType: types.ObjectType{AttrTypes: serviceDescriptorAttrs}},
//		"video_settings":             types.ListType{ElemType: types.ObjectType{AttrTypes: videoSettingsAttrs}},
//	}
//)
//
//func flattenMultiplexProgramSettings(ctx context.Context, mps *awstypes.MultiplexProgramSettings) types.List {
//	elemType := types.ObjectType{AttrTypes: multiplexProgramSettingsAttrs}
//
//	if mps == nil {
//		return types.ListValueMust(elemType, []attr.Value{})
//	}
//
//	attrs := map[string]attr.Value{}
//	attrs["program_number"] = flex.Int32ToFramework(ctx, mps.ProgramNumber)
//	attrs["preferred_channel_pipeline"] = flex.StringValueToFrameworkLegacy(ctx, mps.PreferredChannelPipeline)
//	attrs["service_descriptor"] = flattenServiceDescriptor(ctx, mps.ServiceDescriptor)
//	attrs["video_settings"] = flattenVideoSettings(ctx, mps.VideoSettings)
//
//	vals := types.ObjectValueMust(multiplexProgramSettingsAttrs, attrs)
//
//	return types.ListValueMust(elemType, []attr.Value{vals})
//}
//
//func flattenServiceDescriptor(ctx context.Context, sd *awstypes.MultiplexProgramServiceDescriptor) types.List {
//	elemType := types.ObjectType{AttrTypes: serviceDescriptorAttrs}
//
//	if sd == nil {
//		return types.ListValueMust(elemType, []attr.Value{})
//	}
//
//	attrs := map[string]attr.Value{}
//	attrs[names.AttrProviderName] = flex.StringToFrameworkLegacy(ctx, sd.ProviderName)
//	attrs[names.AttrServiceName] = flex.StringToFrameworkLegacy(ctx, sd.ServiceName)
//
//	vals := types.ObjectValueMust(serviceDescriptorAttrs, attrs)
//
//	return types.ListValueMust(elemType, []attr.Value{vals})
//}
//
//func flattenStatMuxSettings(ctx context.Context, mps *awstypes.MultiplexStatmuxVideoSettings) types.List {
//	elemType := types.ObjectType{AttrTypes: statmuxAttrs}
//
//	if mps == nil {
//		return types.ListValueMust(elemType, []attr.Value{})
//	}
//
//	attrs := map[string]attr.Value{}
//	attrs["minimum_bitrate"] = flex.Int32ToFramework(ctx, mps.MinimumBitrate)
//	attrs["maximum_bitrate"] = flex.Int32ToFramework(ctx, mps.MaximumBitrate)
//	attrs[names.AttrPriority] = flex.Int32ToFramework(ctx, mps.Priority)
//
//	vals := types.ObjectValueMust(statmuxAttrs, attrs)
//
//	return types.ListValueMust(elemType, []attr.Value{vals})
//}
//
//func flattenVideoSettings(ctx context.Context, mps *awstypes.MultiplexVideoSettings) types.List {
//	elemType := types.ObjectType{AttrTypes: videoSettingsAttrs}
//
//	if mps == nil {
//		return types.ListValueMust(elemType, []attr.Value{})
//	}
//
//	attrs := map[string]attr.Value{}
//	attrs["constant_bitrate"] = flex.Int32ToFramework(ctx, mps.ConstantBitrate)
//	attrs["statmux_settings"] = flattenStatMuxSettings(ctx, mps.StatmuxSettings)
//
//	vals := types.ObjectValueMust(videoSettingsAttrs, attrs)
//
//	return types.ListValueMust(elemType, []attr.Value{vals})
//}

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
	ID                       types.String                                              `tfsdk:"id"`
	MultiplexID              types.String                                              `tfsdk:"multiplex_id"`
	MultiplexProgramSettings fwtypes.ListNestedObjectValueOf[multiplexProgramSettings] `tfsdk:"multiplex_program_settings"`
	ProgramName              types.String                                              `tfsdk:"program_name"`
}

type multiplexProgramSettings struct {
	ProgramNumber            types.Int64                                           `tfsdk:"program_number"`
	PreferredChannelPipeline fwtypes.StringEnum[awstypes.PreferredChannelPipeline] `tfsdk:"preferred_channel_pipeline"`
	ServiceDescriptor        fwtypes.ListNestedObjectValueOf[serviceDescriptor]    `tfsdk:"service_descriptor"`
	VideoSettings            fwtypes.ListNestedObjectValueOf[videoSettings]        `tfsdk:"video_settings"`
}

type serviceDescriptor struct {
	ProviderName types.String `tfsdk:"provider_name"`
	ServiceName  types.String `tfsdk:"service_name"`
}

type videoSettings struct {
	ConstantBitrate types.Int64                                      `tfsdk:"constant_bitrate"`
	StatmuxSettings fwtypes.ListNestedObjectValueOf[statmuxSettings] `tfsdk:"statmux_settings"`
}

type statmuxSettings struct {
	MaximumBitrate types.Int64 `tfsdk:"maximum_bitrate"`
	MinimumBitrate types.Int64 `tfsdk:"minimum_bitrate"`
	Priority       types.Int64 `tfsdk:"priority"`
}
