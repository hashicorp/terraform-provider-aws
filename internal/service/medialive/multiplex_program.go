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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resourceHelper "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
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
									"minimum_bitrate": {
										Type:     types.NumberType,
										Optional: true,
									},
									"maximum_bitrate": {
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

func (m *multiplexProgram) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		m.meta = v
	}
}

func (m *multiplexProgram) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := m.meta.MediaLiveConn

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

	result.ID = types.String{Value: fmt.Sprintf("%s/%s", programName, multiplexId)}
	result.ProgramName = types.String{Value: aws.ToString(out.MultiplexProgram.ProgramName)}
	result.MultiplexID = types.String{Value: plan.MultiplexID.Value}
	result.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgram.MultiplexProgramSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := m.meta.MediaLiveConn

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

	state.MultiplexProgramSettings = flattenMultiplexProgramSettings(out.MultiplexProgramSettings)
	state.ProgramName = types.String{Value: aws.ToString(out.ProgramName)}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model resourceMultiplexProgramData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (m *multiplexProgram) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := m.meta.MediaLiveConn

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

func (m *multiplexProgram) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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

	sdList := make([]serviceDescriptor, 0)
	if mps.ServiceDescriptor != nil {
		var sd serviceDescriptor
		sd.ProviderName = types.String{Value: aws.ToString(mps.ServiceDescriptor.ProviderName)}
		sd.ServiceName = types.String{Value: aws.ToString(mps.ServiceDescriptor.ServiceName)}

		sdList = append(sdList, sd)
	}
	m.ServiceDescriptor = sdList

	vsList := make([]videoSettings, 0)
	if mps.VideoSettings != nil {
		var vs videoSettings
		vs.ConstantBitrate = mps.VideoSettings.ConstantBitrate

		ssList := make([]statemuxSettings, 0)
		if mps.VideoSettings.StatmuxSettings != nil {
			var s statemuxSettings
			s.MinimumBitrate = mps.VideoSettings.StatmuxSettings.MinimumBitrate
			s.MaximumBitrate = mps.VideoSettings.StatmuxSettings.MaximumBitrate
			s.Priority = mps.VideoSettings.StatmuxSettings.Priority

			ssList = append(ssList, s)
		}
		vs.StatemuxSettings = ssList
		vsList = append(vsList, vs)
	}
	m.VideoSettings = vsList

	return []multiplexProgramSettings{m}
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

	if idParts[0] == "" || idParts[1] == "" {
		err = errors.New("invalid id")
		return
	}

	programName = idParts[0]
	multiplexId = idParts[1]

	return
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
