package medialive

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	mltypes "github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resourceHelper "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func NewResourceMultiplexProgramType(_ context.Context, meta interface{ Meta() interface{} }) provider.ResourceType {
	return &resourceMultiplexProgramType{
		meta: meta,
	}
}

type resourceMultiplexProgramType struct {
	meta interface{ Meta() interface{} }
}

func (t *resourceMultiplexProgramType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"program_number": {
						Type:     types.NumberType,
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
								Type:     types.NumberType,
								Optional: true,
							},
						},
						Blocks: map[string]tfsdk.Block{
							"statemux_settings": {
								NestingMode: tfsdk.BlockNestingModeList,
								MaxItems:    1,
								Attributes: map[string]tfsdk.Attribute{
									"maximum_bitrate": {
										Type:     types.NumberType,
										Optional: true,
									},
									"minimum_bitrate": {
										Type:     types.NumberType,
										Optional: true,
									},
									"priority": {
										Type:     types.NumberType,
										Optional: true,
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

const (
	ResNameMultiplexProgram = "Multiplex Program"
)

func (t *resourceMultiplexProgramType) NewResource(ctx context.Context, provider provider.Provider) (resource.Resource, diag.Diagnostics) {
	return &multiplexProgram{
		meta: t.meta,
	}, nil
}

type multiplexProgram struct {
	meta interface{ Meta() interface{} }
}

func (m *multiplexProgram) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := m.meta.Meta().(*conns.AWSClient).MediaLiveConn

	var plan resourceMultiplexProgramData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &medialive.CreateMultiplexProgramInput{
		MultiplexId: aws.String(plan.MultiplexID.Value),
		ProgramName: aws.String(plan.ProgramName.Value),
		RequestId:   aws.String(resourceHelper.UniqueId()),
	}

	in.MultiplexProgramSettings = expandMultiplexProgramSettings(plan.MultiplexProgramSettings)

	out, err := conn.CreateMultiplexProgram(ctx, in)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionCreating, ResNameMultiplexProgram, plan.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	var result resourceMultiplexProgramData

	result.ID.Value = fmt.Sprintf("%s/%s", plan.ProgramName.Value, plan.MultiplexID.Value)
	result.ProgramName.Value = aws.ToString(out.MultiplexProgram.ProgramName)
	result.MultiplexID.Value = plan.MultiplexID.Value
	result.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgram.MultiplexProgramSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := m.meta.Meta().(*conns.AWSClient).MediaLiveConn

	var state resourceMultiplexProgramData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	multiplexId := state.MultiplexID.Value
	programName := state.ProgramName.Value

	out, err := FindMultipleProgramByID(ctx, conn, multiplexId, programName)

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

	state.ID.Value = fmt.Sprintf("%s/%s", state.ProgramName.Value, state.MultiplexID.Value)
	state.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgramSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := m.meta.Meta().(*conns.AWSClient).MediaLiveConn

	var plan resourceMultiplexProgramData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state resourceMultiplexProgramData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &medialive.UpdateMultiplexProgramInput{
		MultiplexId:              aws.String(state.MultiplexID.Value),
		ProgramName:              aws.String(state.ProgramName.Value),
		MultiplexProgramSettings: expandMultiplexProgramSettings(plan.MultiplexProgramSettings),
	}

	out, err := conn.UpdateMultiplexProgram(ctx, in)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaLive, create.ErrActionUpdating, ResNameMultiplexProgram, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	//var result resourceMultiplexProgramData

	state.ProgramName.Value = aws.ToString(out.MultiplexProgram.ProgramName)
	//state.MultiplexID.Value = state.MultiplexID.Value
	state.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgram.MultiplexProgramSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := m.meta.Meta().(*conns.AWSClient).MediaLiveConn

	var state resourceMultiplexProgramData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	multiplexId := state.MultiplexID.Value
	programName := state.ProgramName.Value

	_, err := conn.DeleteMultiplexProgram(ctx, &medialive.DeleteMultiplexProgramInput{
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

func FindMultipleProgramByID(ctx context.Context, conn *medialive.Client, multiplexId, programName string) (*medialive.DescribeMultiplexProgramOutput, error) {
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

func expandMultiplexProgramSettings(mps []multiplexProgramSettings) *mltypes.MultiplexProgramSettings {
	if len(mps) == 0 {
		return nil
	}

	data := mps[0]

	l := &mltypes.MultiplexProgramSettings{
		ProgramNumber:            data.ProgramNumber,
		PreferredChannelPipeline: mltypes.PreferredChannelPipeline(data.PreferredChannelPipeline.Value),
	}

	if len(data.ServiceDescriptor) > 0 {
		l.ServiceDescriptor = &mltypes.MultiplexProgramServiceDescriptor{
			ProviderName: aws.String(data.ServiceDescriptor[0].ProviderName.Value),
			ServiceName:  aws.String(data.ServiceDescriptor[0].ServiceName.Value),
		}
	}

	if len(data.VideoSettings) > 0 {
		l.VideoSettings = &mltypes.MultiplexVideoSettings{
			ConstantBitrate: data.VideoSettings[0].ConstantBitrate,
		}

		if len(data.VideoSettings[0].StatemuxSettings) > 0 {
			l.VideoSettings.StatmuxSettings = &mltypes.MultiplexStatmuxVideoSettings{
				MinimumBitrate: data.VideoSettings[0].StatemuxSettings[0].MinimumBitrate,
				MaximumBitrate: data.VideoSettings[0].StatemuxSettings[0].MaximumBitrate,
				Priority:       data.VideoSettings[0].StatemuxSettings[0].Priority,
			}
		}
	}

	return l
}

func flattenMultiplexProgramSettings(mps *mltypes.MultiplexProgramSettings) []multiplexProgramSettings {
	if mps == nil {
		return nil
	}

	m := multiplexProgramSettings{
		ProgramNumber:            mps.ProgramNumber,
		PreferredChannelPipeline: types.String{Value: string(mps.PreferredChannelPipeline)},
	}

	if mps.ServiceDescriptor != nil {
		m.ServiceDescriptor = []serviceDescriptor{
			{
				ProviderName: types.String{Value: aws.ToString(mps.ServiceDescriptor.ProviderName)},
				ServiceName:  types.String{Value: aws.ToString(mps.ServiceDescriptor.ServiceName)},
			},
		}
	}

	if mps.VideoSettings != nil {
		var vs videoSettings
		vs.ConstantBitrate = mps.VideoSettings.ConstantBitrate
		if mps.VideoSettings.StatmuxSettings != nil {
			vs.StatemuxSettings = []statemuxSettings{
				{
					MinimumBitrate: mps.VideoSettings.StatmuxSettings.MinimumBitrate,
					MaximumBitrate: mps.VideoSettings.StatmuxSettings.MaximumBitrate,
					Priority:       mps.VideoSettings.StatmuxSettings.Priority,
				},
			}
		}

		m.VideoSettings = []videoSettings{vs}
	}

	return []multiplexProgramSettings{m}
}

func preferredChannelPipelineToSlice(p []mltypes.PreferredChannelPipeline) []string {
	s := make([]string, 0)

	for _, v := range p {
		s = append(s, string(v))
	}
	return s
}

type resourceMultiplexProgramData struct {
	ID                       types.String               `tfsdk:"id"`
	MultiplexID              types.String               `tfsdk:"multiplex_id"`
	MultiplexProgramSettings []multiplexProgramSettings `tfsdk:"multiplex_program_settings"`
	ProgramName              types.String               `tfsdk:"program_name"`
}

type multiplexProgramSettings struct {
	ProgramNumber            int32               `tfsdk:"program_number"`
	PreferredChannelPipeline types.String        `tfsdk:"preferred_channel_pipeline"`
	ServiceDescriptor        []serviceDescriptor `tfsdk:"service_descriptor"`
	VideoSettings            []videoSettings     `tfsdk:"video_settings"`
}

type serviceDescriptor struct {
	ProviderName types.String `tfsdk:"provider_name"`
	ServiceName  types.String `tfsdk:"service_name"`
}

type videoSettings struct {
	ConstantBitrate  int32              `tfsdk:"constant_bitrate"`
	StatemuxSettings []statemuxSettings `tfsdk:"statemux_settings"`
}

type statemuxSettings struct {
	MaximumBitrate int32 `tfsdk:"maximum_bitrate"`
	MinimumBitrate int32 `tfsdk:"minimum_bitrate"`
	Priority       int32 `tfsdk:"priority"`
}
