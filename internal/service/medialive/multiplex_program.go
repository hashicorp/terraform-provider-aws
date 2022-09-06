package medialive

import (
	"context"
	"log"

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
)

func NewResourceMultiplexProgramType(_ context.Context, meta interface{}) provider.ResourceType {
	return &resourceMultiplexProgramType{
		meta: meta,
	}
}

type resourceMultiplexProgramType struct {
	meta interface{}
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
	meta interface{}
}

func (m *multiplexProgram) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	log.Printf("[INFO] meta: %v", m.meta)
	conn := m.meta.(conns.AWSClient).MediaLiveConn

	var plan resourceMultiplexProgramData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &medialive.CreateMultiplexProgramInput{
		MultiplexId: aws.String(plan.MultiplexID.Value),
		RequestId:   aws.String(resourceHelper.UniqueId()),
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
		ID:          types.String{Value: aws.ToString(out.MultiplexProgram.ProgramName)},
		ProgramName: types.String{Value: aws.ToString(out.MultiplexProgram.ProgramName)},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *multiplexProgram) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	ID                       types.String `tfsdk:"id"`
	MultiplexID              types.String `tfsdk:"multiplex_id"`
	MultiplexProgramSettings *multiplexProgramSettings
	ProgramName              types.String `tfsdk:"program_name"`
}

type multiplexProgramSettings struct {
	ProgramNumber            types.Number `tfsdk:"program_number"`
	PreferredChannelPipeline types.String `tfsdk:"preferred_channel_pipeline"`
}
