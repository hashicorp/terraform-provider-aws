package medialive

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	mltypes "github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resourceHelper "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
						Optional: true,
						//Validators: []tfsdk.AttributeValidator{
						//	stringvalidator.OneOf(preferredChannelPipelineToSlice(mltypes.PreferredChannelPipeline("").Values())...),
						//},
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
					"statmux_settings": {
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
	}

	return schema, nil
}

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

	in.MultiplexProgramSettings = &mltypes.MultiplexProgramSettings{
		ProgramNumber:            plan.MultiplexProgramSettings[0].ProgramNumber,
		PreferredChannelPipeline: mltypes.PreferredChannelPipeline(plan.MultiplexProgramSettings[0].PreferredChannelPipeline.Value),
	}

	out, err := conn.CreateMultiplexProgram(ctx, in)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating MediaLive Multiplex Program",
			err.Error(),
		)
		return
	}

	result := resourceMultiplexProgramData{
		ProgramName: types.String{Value: aws.ToString(out.MultiplexProgram.ProgramName)},
		MultiplexID: types.String{Value: plan.MultiplexID.Value},
		MultiplexProgramSettings: []multiplexProgramSettings{
			{
				ProgramNumber:            out.MultiplexProgram.MultiplexProgramSettings.ProgramNumber,
				PreferredChannelPipeline: types.String{Value: string(out.MultiplexProgram.MultiplexProgramSettings.PreferredChannelPipeline)},
			},
		},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)

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

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MediaLive Multiplex Program",
			err.Error(),
		)
		return
	}

	state.ProgramName = types.String{Value: aws.ToString(out.ProgramName)}
	state.MultiplexProgramSettings = []multiplexProgramSettings{
		{
			ProgramNumber:            out.MultiplexProgramSettings.ProgramNumber,
			PreferredChannelPipeline: types.String{Value: string(out.MultiplexProgramSettings.PreferredChannelPipeline)},
		},
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (m *multiplexProgram) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func preferredChannelPipelineToSlice(p []mltypes.PreferredChannelPipeline) []string {
	s := make([]string, 0)

	for _, v := range p {
		s = append(s, string(v))
	}
	return s
}

type resourceMultiplexProgramData struct {
	MultiplexID              types.String               `tfsdk:"multiplex_id"`
	MultiplexProgramSettings []multiplexProgramSettings `tfsdk:"multiplex_program_settings"`
	ProgramName              types.String               `tfsdk:"program_name"`
	VideoSettings            []videoSettings            `tfsdk:"video_settings"`
}

type multiplexProgramSettings struct {
	ProgramNumber            int32               `tfsdk:"program_number"`
	PreferredChannelPipeline types.String        `tfsdk:"preferred_channel_pipeline"`
	ServiceDescriptor        []serviceDescriptor `tfsdk:"service_descriptor"`
}

type serviceDescriptor struct {
	ProviderName types.String `tfsdk:"provider_name"`
	ServiceName  types.String `tfsdk:"service_name"`
}

type videoSettings struct {
	ConstantBitrate  types.Number       `tfsdk:"constant_bitrate"`
	statemuxSettings []statemuxSettings `tfsdk:"statemux_settings"`
}

type statemuxSettings struct {
	MaximumBitrate types.Number `tfsdk:"maximum_bitrate"`
	MinimimBitrate types.Number `tfsdk:"minimum_bitrate"`
	Priority       types.Number `tfsdk:"priority"`
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
