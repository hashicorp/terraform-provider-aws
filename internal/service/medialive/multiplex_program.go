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
	resourceHelper "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	_sp.registerFrameworkResourceFactory(newResourceMultiplexProgram)
}

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
			"id": framework.IDAttribute(),
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
									"provider_name": schema.StringAttribute{
										Required: true,
									},
									"service_name": schema.StringAttribute{
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
									"statemux_settings": schema.ListNestedBlock{
										DeprecationMessage: "Configure statmux_settings instead of statemux_settings. This block will be removed in the next major version of the provider.",
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
												"priority": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
											},
										},
									},
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
												"priority": schema.Int64Attribute{
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
	conn := m.Meta().MediaLiveClient()

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
		RequestId:   aws.String(resourceHelper.UniqueId()),
	}

	mps := make([]multiplexProgramSettings, 1)
	resp.Diagnostics.Append(plan.MultiplexProgramSettings.ElementsAs(ctx, &mps, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpSettings, isStateMuxSet, err := expandMultiplexProgramSettings(ctx, mps)

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

	result.ID = types.StringValue(fmt.Sprintf("%s/%s", programName, multiplexId))
	result.ProgramName = types.StringValue(aws.ToString(out.MultiplexProgram.ProgramName))
	result.MultiplexID = types.StringValue(plan.MultiplexID.ValueString())
	result.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgram.MultiplexProgramSettings, isStateMuxSet)

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := m.Meta().MediaLiveClient()

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

	sm := make([]videoSettings, 1)
	attErr := req.State.GetAttribute(ctx, path.Root("multiplex_program_settings").
		AtListIndex(0).AtName("video_settings"), &sm)

	resp.Diagnostics.Append(attErr...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateMuxIsNull bool
	if len(sm) > 0 {
		if len(sm[0].StatemuxSettings.Elements()) == 0 {
			stateMuxIsNull = true
		}
	}
	state.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgramSettings, stateMuxIsNull)
	state.ProgramName = types.StringValue(aws.ToString(out.ProgramName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := m.Meta().MediaLiveClient()

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

	mps := make([]multiplexProgramSettings, 1)
	resp.Diagnostics.Append(plan.MultiplexProgramSettings.ElementsAs(ctx, &mps, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpSettings, stateMuxIsNull, errExpand := expandMultiplexProgramSettings(ctx, mps)

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

	plan.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgramSettings, stateMuxIsNull)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (m *multiplexProgram) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := m.Meta().MediaLiveClient()

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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m *multiplexProgram) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resourceMultiplexProgramData

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mps := make([]multiplexProgramSettings, 1)
	resp.Diagnostics.Append(data.MultiplexProgramSettings.ElementsAs(ctx, &mps, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(mps[0].VideoSettings.Elements()) > 0 || !mps[0].VideoSettings.IsNull() {
		vs := make([]videoSettings, 1)
		resp.Diagnostics.Append(mps[0].VideoSettings.ElementsAs(ctx, &vs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		statMuxSet := len(vs[0].StatmuxSettings.Elements()) > 0
		stateMuxSet := len(vs[0].StatemuxSettings.Elements()) > 0

		if statMuxSet && stateMuxSet {
			resp.Diagnostics.AddAttributeError(
				path.Root("multiplex_program_settings").AtListIndex(0).AtName("video_settings").AtListIndex(0).AtName("statmux_settings"),
				"Conflicting Attribute Configuration",
				"Attribute statmux_settings cannot be configured with statemux_settings.",
			)
		}
	}
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
			return nil, &resourceHelper.NotFoundError{
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

func expandMultiplexProgramSettings(ctx context.Context, mps []multiplexProgramSettings) (*mltypes.MultiplexProgramSettings, bool, diag.Diagnostics) {
	if len(mps) == 0 {
		return nil, false, nil
	}

	var stateMuxIsNull bool
	data := mps[0]

	l := &mltypes.MultiplexProgramSettings{
		ProgramNumber:            int32(data.ProgramNumber.ValueInt64()),
		PreferredChannelPipeline: mltypes.PreferredChannelPipeline(data.PreferredChannelPipeline.ValueString()),
	}

	if len(data.ServiceDescriptor.Elements()) > 0 && !data.ServiceDescriptor.IsNull() {
		sd := make([]serviceDescriptor, 1)
		err := data.ServiceDescriptor.ElementsAs(ctx, &sd, false)
		if err.HasError() {
			return nil, false, err
		}

		l.ServiceDescriptor = &mltypes.MultiplexProgramServiceDescriptor{
			ProviderName: aws.String(sd[0].ProviderName.ValueString()),
			ServiceName:  aws.String(sd[0].ServiceName.ValueString()),
		}
	}

	if len(data.VideoSettings.Elements()) > 0 && !data.VideoSettings.IsNull() {
		vs := make([]videoSettings, 1)
		err := data.VideoSettings.ElementsAs(ctx, &vs, false)
		if err.HasError() {
			return nil, false, err
		}

		l.VideoSettings = &mltypes.MultiplexVideoSettings{
			ConstantBitrate: int32(vs[0].ConstantBitrate.ValueInt64()),
		}

		// Deprecated: will be removed in the next major version
		if len(vs[0].StatemuxSettings.Elements()) > 0 && !vs[0].StatemuxSettings.IsNull() {
			sms := make([]statmuxSettings, 1)
			err := vs[0].StatemuxSettings.ElementsAs(ctx, &sms, false)
			if err.HasError() {
				return nil, false, err
			}

			l.VideoSettings.StatmuxSettings = &mltypes.MultiplexStatmuxVideoSettings{
				MinimumBitrate: int32(sms[0].MinimumBitrate.ValueInt64()),
				MaximumBitrate: int32(sms[0].MaximumBitrate.ValueInt64()),
				Priority:       int32(sms[0].Priority.ValueInt64()),
			}
		}

		if len(vs[0].StatmuxSettings.Elements()) > 0 && !vs[0].StatmuxSettings.IsNull() {
			stateMuxIsNull = true
			sms := make([]statmuxSettings, 1)
			err := vs[0].StatmuxSettings.ElementsAs(ctx, &sms, false)
			if err.HasError() {
				return nil, false, err
			}

			l.VideoSettings.StatmuxSettings = &mltypes.MultiplexStatmuxVideoSettings{
				MinimumBitrate: int32(sms[0].MinimumBitrate.ValueInt64()),
				MaximumBitrate: int32(sms[0].MaximumBitrate.ValueInt64()),
				Priority:       int32(sms[0].Priority.ValueInt64()),
			}
		}
	}

	return l, stateMuxIsNull, nil
}

var (
	statmuxAttrs = map[string]attr.Type{
		"minimum_bitrate": types.Int64Type,
		"maximum_bitrate": types.Int64Type,
		"priority":        types.Int64Type,
	}

	videoSettingsAttrs = map[string]attr.Type{
		"constant_bitrate":  types.Int64Type,
		"statemux_settings": types.ListType{ElemType: types.ObjectType{AttrTypes: statmuxAttrs}},
		"statmux_settings":  types.ListType{ElemType: types.ObjectType{AttrTypes: statmuxAttrs}},
	}

	serviceDescriptorAttrs = map[string]attr.Type{
		"provider_name": types.StringType,
		"service_name":  types.StringType,
	}

	multiplexProgramSettingsAttrs = map[string]attr.Type{
		"program_number":             types.Int64Type,
		"preferred_channel_pipeline": types.StringType,
		"service_descriptor":         types.ListType{ElemType: types.ObjectType{AttrTypes: serviceDescriptorAttrs}},
		"video_settings":             types.ListType{ElemType: types.ObjectType{AttrTypes: videoSettingsAttrs}},
	}
)

func flattenMultiplexProgramSettings(mps *mltypes.MultiplexProgramSettings, stateMuxIsNull bool) types.List {
	elemType := types.ObjectType{AttrTypes: multiplexProgramSettingsAttrs}

	if mps == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{}
	attrs["program_number"] = types.Int64Value(int64(mps.ProgramNumber))
	attrs["preferred_channel_pipeline"] = types.StringValue(string(mps.PreferredChannelPipeline))
	attrs["service_descriptor"] = flattenServiceDescriptor(mps.ServiceDescriptor)
	attrs["video_settings"] = flattenVideoSettings(mps.VideoSettings, stateMuxIsNull)

	vals := types.ObjectValueMust(multiplexProgramSettingsAttrs, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}

func flattenServiceDescriptor(sd *mltypes.MultiplexProgramServiceDescriptor) types.List {
	elemType := types.ObjectType{AttrTypes: serviceDescriptorAttrs}

	if sd == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{}
	attrs["provider_name"] = types.StringValue(aws.ToString(sd.ProviderName))
	attrs["service_name"] = types.StringValue(aws.ToString(sd.ServiceName))

	vals := types.ObjectValueMust(serviceDescriptorAttrs, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}

func flattenStatMuxSettings(mps *mltypes.MultiplexStatmuxVideoSettings) types.List {
	elemType := types.ObjectType{AttrTypes: statmuxAttrs}

	if mps == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{}
	attrs["minimum_bitrate"] = types.Int64Value(int64(mps.MinimumBitrate))
	attrs["maximum_bitrate"] = types.Int64Value(int64(mps.MaximumBitrate))
	attrs["priority"] = types.Int64Value(int64(mps.Priority))

	vals := types.ObjectValueMust(statmuxAttrs, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}

func flattenVideoSettings(mps *mltypes.MultiplexVideoSettings, stateMuxIsNull bool) types.List {
	elemType := types.ObjectType{AttrTypes: videoSettingsAttrs}

	if mps == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{}
	attrs["constant_bitrate"] = types.Int64Value(int64(mps.ConstantBitrate))

	if stateMuxIsNull {
		attrs["statmux_settings"] = flattenStatMuxSettings(mps.StatmuxSettings)
		attrs["statemux_settings"] = types.ListValueMust(types.ObjectType{AttrTypes: statmuxAttrs}, []attr.Value{})
	} else {
		attrs["statmux_settings"] = types.ListValueMust(types.ObjectType{AttrTypes: statmuxAttrs}, []attr.Value{})
		attrs["statemux_settings"] = flattenStatMuxSettings(mps.StatmuxSettings)
	}

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
	ConstantBitrate  types.Int64 `tfsdk:"constant_bitrate"`
	StatemuxSettings types.List  `tfsdk:"statemux_settings"` // Deprecated: will be removed in the next major version
	StatmuxSettings  types.List  `tfsdk:"statmux_settings"`
}

type statmuxSettings struct {
	MaximumBitrate types.Int64 `tfsdk:"maximum_bitrate"`
	MinimumBitrate types.Int64 `tfsdk:"minimum_bitrate"`
	Priority       types.Int64 `tfsdk:"priority"`
}
