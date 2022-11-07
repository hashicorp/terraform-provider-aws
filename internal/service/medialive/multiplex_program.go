package medialive

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	mltypes "github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resourceHelper "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/intf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	registerFrameworkResourceFactory(newResourceMultiplexProgram)
}

func newResourceMultiplexProgram(_ context.Context) (intf.ResourceWithConfigureAndImportState, error) {
	return &multiplexProgram{}, nil
}

func NewResourceMultiplexProgram(_ context.Context) resource.Resource {
	return &multiplexProgram{}
}

const (
	ResNameMultiplexProgram = "Multiplex Program"
)

type multiplexProgram struct {
	meta *conns.AWSClient
}

func (m *multiplexProgram) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_medialive_multiplex_program"
}

func (m *multiplexProgram) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"multiplex_id": {
				Type:     types.StringType,
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"program_name": {
				Type:     types.StringType,
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]tfsdk.Block{
			"multiplex_program_settings": {
				NestingMode: tfsdk.BlockNestingModeList,
				MinItems:    1,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"program_number": {
						Type:     types.Int64Type,
						Required: true,
					},
					"preferred_channel_pipeline": {
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf(preferredChannelPipelineToSlice(mltypes.PreferredChannelPipeline("").Values())...),
						},
					},
				},
				Blocks: map[string]tfsdk.Block{
					"service_descriptor": {
						NestingMode: tfsdk.BlockNestingModeList,
						MaxItems:    1,
						Attributes: map[string]tfsdk.Attribute{
							"provider_name": {
								Type:     types.StringType,
								Required: true,
							},
							"service_name": {
								Type:     types.StringType,
								Required: true,
							},
						},
					},
					"video_settings": {
						NestingMode: tfsdk.BlockNestingModeList,
						MaxItems:    1,
						Attributes: map[string]tfsdk.Attribute{
							"constant_bitrate": {
								Type:     types.Int64Type,
								Optional: true,
								Computed: true,
								PlanModifiers: []tfsdk.AttributePlanModifier{
									resource.UseStateForUnknown(),
								},
							},
						},
						Blocks: map[string]tfsdk.Block{
							"statemux_settings": {
								DeprecationMessage: "Configure statmux_settings instead of statemux_settings. This block will be removed in the next major version of the provider.",
								NestingMode:        tfsdk.BlockNestingModeList,
								MaxItems:           1,
								Attributes: map[string]tfsdk.Attribute{
									"minimum_bitrate": {
										Type:     types.Int64Type,
										Optional: true,
										Computed: true,
										PlanModifiers: []tfsdk.AttributePlanModifier{
											resource.UseStateForUnknown(),
										},
									},
									"maximum_bitrate": {
										Type:     types.Int64Type,
										Optional: true,
										Computed: true,
										PlanModifiers: []tfsdk.AttributePlanModifier{
											resource.UseStateForUnknown(),
										},
									},
									"priority": {
										Type:     types.Int64Type,
										Optional: true,
										Computed: true,
										PlanModifiers: []tfsdk.AttributePlanModifier{
											resource.UseStateForUnknown(),
										},
									},
								},
							},
							"statmux_settings": {
								NestingMode: tfsdk.BlockNestingModeList,
								MaxItems:    1,
								Attributes: map[string]tfsdk.Attribute{
									"minimum_bitrate": {
										Type:     types.Int64Type,
										Optional: true,
										Computed: true,
										PlanModifiers: []tfsdk.AttributePlanModifier{
											resource.UseStateForUnknown(),
										},
									},
									"maximum_bitrate": {
										Type:     types.Int64Type,
										Optional: true,
										Computed: true,
										PlanModifiers: []tfsdk.AttributePlanModifier{
											resource.UseStateForUnknown(),
										},
									},
									"priority": {
										Type:     types.Int64Type,
										Optional: true,
										Computed: true,
										PlanModifiers: []tfsdk.AttributePlanModifier{
											resource.UseStateForUnknown(),
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

	return schema, nil
}

func (m *multiplexProgram) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		m.meta = v
	}
}

func (m *multiplexProgram) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := m.meta.MediaLiveClient

	var plan resourceMultiplexProgramData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	multiplexId := plan.MultiplexID.Value
	programName := plan.ProgramName.Value

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

	result.ID = types.String{Value: fmt.Sprintf("%s/%s", programName, multiplexId)}
	result.ProgramName = types.String{Value: aws.ToString(out.MultiplexProgram.ProgramName)}
	result.MultiplexID = types.String{Value: plan.MultiplexID.Value}
	result.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgram.MultiplexProgramSettings, isStateMuxSet)

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := m.meta.MediaLiveClient

	var state resourceMultiplexProgramData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseMultiplexProgramID(state.ID.Value)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionReading, ResNameMultiplexProgram, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	out, err := FindMultiplexProgramByID(ctx, conn, multiplexId, programName)

	if tfresource.NotFound(err) {
		diag.NewWarningDiagnostic(
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
		if len(sm[0].StatemuxSettings.Elems) == 0 {
			stateMuxIsNull = true
		}
	}
	state.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgramSettings, stateMuxIsNull)
	state.ProgramName = types.String{Value: aws.ToString(out.ProgramName)}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := m.meta.MediaLiveClient

	var plan resourceMultiplexProgramData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseMultiplexProgramID(plan.ID.Value)

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
	conn := m.meta.MediaLiveClient

	var state resourceMultiplexProgramData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseMultiplexProgramID(state.ID.Value)

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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionDeleting, ResNameMultiplexProgram, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
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

	if len(mps[0].VideoSettings.Elems) > 0 || !mps[0].VideoSettings.IsNull() {
		vs := make([]videoSettings, 1)
		resp.Diagnostics.Append(mps[0].VideoSettings.ElementsAs(ctx, &vs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		statMuxSet := len(vs[0].StatmuxSettings.Elems) > 0
		stateMuxSet := len(vs[0].StatemuxSettings.Elems) > 0

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
		ProgramNumber:            int32(data.ProgramNumber.Value),
		PreferredChannelPipeline: mltypes.PreferredChannelPipeline(data.PreferredChannelPipeline.Value),
	}

	if len(data.ServiceDescriptor.Elems) > 0 && !data.ServiceDescriptor.IsNull() {
		sd := make([]serviceDescriptor, 1)
		err := data.ServiceDescriptor.ElementsAs(ctx, &sd, false)
		if err.HasError() {
			return nil, false, err
		}

		l.ServiceDescriptor = &mltypes.MultiplexProgramServiceDescriptor{
			ProviderName: aws.String(sd[0].ProviderName.Value),
			ServiceName:  aws.String(sd[0].ServiceName.Value),
		}
	}

	if len(data.VideoSettings.Elems) > 0 && !data.VideoSettings.IsNull() {
		vs := make([]videoSettings, 1)
		err := data.VideoSettings.ElementsAs(ctx, &vs, false)
		if err.HasError() {
			return nil, false, err
		}

		l.VideoSettings = &mltypes.MultiplexVideoSettings{
			ConstantBitrate: int32(vs[0].ConstantBitrate.Value),
		}

		// Deprecated: will be removed in the next major version
		if len(vs[0].StatemuxSettings.Elems) > 0 && !vs[0].StatemuxSettings.IsNull() {
			sms := make([]statmuxSettings, 1)
			err := vs[0].StatemuxSettings.ElementsAs(ctx, &sms, false)
			if err.HasError() {
				return nil, false, err
			}

			l.VideoSettings.StatmuxSettings = &mltypes.MultiplexStatmuxVideoSettings{
				MinimumBitrate: int32(sms[0].MinimumBitrate.Value),
				MaximumBitrate: int32(sms[0].MaximumBitrate.Value),
				Priority:       int32(sms[0].Priority.Value),
			}
		}

		if len(vs[0].StatmuxSettings.Elems) > 0 && !vs[0].StatmuxSettings.IsNull() {
			stateMuxIsNull = true
			sms := make([]statmuxSettings, 1)
			err := vs[0].StatmuxSettings.ElementsAs(ctx, &sms, false)
			if err.HasError() {
				return nil, false, err
			}

			l.VideoSettings.StatmuxSettings = &mltypes.MultiplexStatmuxVideoSettings{
				MinimumBitrate: int32(sms[0].MinimumBitrate.Value),
				MaximumBitrate: int32(sms[0].MaximumBitrate.Value),
				Priority:       int32(sms[0].Priority.Value),
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

	vals := types.Object{AttrTypes: multiplexProgramSettingsAttrs}
	attrs := map[string]attr.Value{}

	if mps == nil {
		return types.List{ElemType: elemType, Elems: []attr.Value{}}
	}

	attrs["program_number"] = types.Int64{Value: int64(mps.ProgramNumber)}
	attrs["preferred_channel_pipeline"] = types.String{Value: string(mps.PreferredChannelPipeline)}
	attrs["service_descriptor"] = flattenServiceDescriptor(mps.ServiceDescriptor)
	attrs["video_settings"] = flattenVideoSettings(mps.VideoSettings, stateMuxIsNull)

	vals.Attrs = attrs

	return types.List{
		Elems:    []attr.Value{vals},
		ElemType: elemType,
	}
}

func flattenServiceDescriptor(sd *mltypes.MultiplexProgramServiceDescriptor) types.List {
	elemType := types.ObjectType{AttrTypes: serviceDescriptorAttrs}

	vals := types.Object{AttrTypes: serviceDescriptorAttrs}
	attrs := map[string]attr.Value{}

	if sd == nil {
		return types.List{ElemType: elemType, Elems: []attr.Value{}}
	}

	attrs["provider_name"] = types.String{Value: aws.ToString(sd.ProviderName)}
	attrs["service_name"] = types.String{Value: aws.ToString(sd.ServiceName)}

	vals.Attrs = attrs

	return types.List{
		Elems:    []attr.Value{vals},
		ElemType: elemType,
	}
}

func flattenStatMuxSettings(mps *mltypes.MultiplexStatmuxVideoSettings) types.List {
	elemType := types.ObjectType{AttrTypes: statmuxAttrs}

	vals := types.Object{AttrTypes: statmuxAttrs}

	if mps == nil {
		return types.List{ElemType: elemType, Elems: []attr.Value{}}
	}

	attrs := map[string]attr.Value{}
	attrs["minimum_bitrate"] = types.Int64{Value: int64(mps.MinimumBitrate)}
	attrs["maximum_bitrate"] = types.Int64{Value: int64(mps.MaximumBitrate)}
	attrs["priority"] = types.Int64{Value: int64(mps.Priority)}

	vals.Attrs = attrs

	return types.List{
		Elems:    []attr.Value{vals},
		ElemType: elemType,
	}
}

func flattenVideoSettings(mps *mltypes.MultiplexVideoSettings, stateMuxIsNull bool) types.List {
	elemType := types.ObjectType{AttrTypes: videoSettingsAttrs}

	vals := types.Object{AttrTypes: videoSettingsAttrs}
	attrs := map[string]attr.Value{}

	if mps == nil {
		return types.List{ElemType: elemType, Elems: []attr.Value{}}
	}

	attrs["constant_bitrate"] = types.Int64{Value: int64(mps.ConstantBitrate)}

	if stateMuxIsNull {
		attrs["statmux_settings"] = flattenStatMuxSettings(mps.StatmuxSettings)
		attrs["statemux_settings"] = types.List{
			Elems:    []attr.Value{},
			ElemType: types.ObjectType{AttrTypes: statmuxAttrs},
		}
	} else {
		attrs["statmux_settings"] = types.List{
			Elems:    []attr.Value{},
			ElemType: types.ObjectType{AttrTypes: statmuxAttrs},
		}
		attrs["statemux_settings"] = flattenStatMuxSettings(mps.StatmuxSettings)
	}

	vals.Attrs = attrs

	return types.List{
		Elems:    []attr.Value{vals},
		ElemType: elemType,
	}
}

func preferredChannelPipelineToSlice(p []mltypes.PreferredChannelPipeline) []string {
	s := make([]string, 0)

	for _, v := range p {
		s = append(s, string(v))
	}
	return s
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
